package capturerunner_test

import (
	"context"
	"image"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"pasha-go/internal/capturerunner"
	"pasha-go/internal/session"
)

type fakeScreener struct {
	mu    sync.Mutex
	calls int
}

func (f *fakeScreener) Capture(image.Rectangle) (image.Image, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
}

type fakeClicker struct {
	mu    sync.Mutex
	calls int
}

func (f *fakeClicker) Click(image.Point) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return nil
}

type fakePdfWriter struct {
	mu      sync.Mutex
	appends int
	closes  int
}

func (f *fakePdfWriter) AppendPage(image.Image) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.appends++
	return nil
}

func (f *fakePdfWriter) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closes++
	return nil
}

type fakeClock struct{}

func (fakeClock) Sleep(context.Context, time.Duration) error { return nil }

func newRunnerWithFakes(t *testing.T) (*capturerunner.Runner, *fakeScreener, *fakeClicker, *fakePdfWriter) {
	t.Helper()
	r, scr, clk, pdf, _ := newRunnerWithFakesCapturingPath(t)
	return r, scr, clk, pdf
}

func newRunnerWithFakesCapturingPath(t *testing.T) (*capturerunner.Runner, *fakeScreener, *fakeClicker, *fakePdfWriter, *string) {
	t.Helper()
	scr := &fakeScreener{}
	clk := &fakeClicker{}
	pdf := &fakePdfWriter{}
	var lastPath string
	newPdf := func(path string) (session.PdfWriter, error) {
		lastPath = path
		return pdf, nil
	}
	r := capturerunner.New(scr, clk, fakeClock{}, newPdf)
	return r, scr, clk, pdf, &lastPath
}

func validPlan(dir string) capturerunner.Plan {
	return capturerunner.Plan{
		RepeatCount:         3,
		StepIntervalSeconds: 0.01,
		CaptureRegion:       image.Rect(0, 0, 100, 100),
		AdvanceClickPoint:   image.Pt(50, 50),
		OutputDir:           dir,
		OutputFileName:      "test",
	}
}

func TestRunner_Run_HappyPath_WiresAllCollaborators(t *testing.T) {
	r, scr, clk, pdf := newRunnerWithFakes(t)

	if err := r.Run(context.Background(), validPlan(t.TempDir()), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if scr.calls != 3 {
		t.Errorf("Screener.Capture calls = %d, want 3", scr.calls)
	}
	if clk.calls != 3 {
		t.Errorf("Clicker.Click calls = %d, want 3", clk.calls)
	}
	if pdf.appends != 3 {
		t.Errorf("PdfWriter.AppendPage calls = %d, want 3", pdf.appends)
	}
	if pdf.closes != 1 {
		t.Errorf("PdfWriter.Close calls = %d, want 1", pdf.closes)
	}
}

func TestRunner_Run_ForwardsProgressPerStep(t *testing.T) {
	r, _, _, _ := newRunnerWithFakes(t)

	var got [][2]int
	onProgress := func(current, total int) {
		got = append(got, [2]int{current, total})
	}

	if err := r.Run(context.Background(), validPlan(t.TempDir()), onProgress); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := [][2]int{{1, 3}, {2, 3}, {3, 3}}
	if len(got) != len(want) {
		t.Fatalf("progress calls = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("progress[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestRunner_Run_NilProgressIsAllowed(t *testing.T) {
	r, _, _, _ := newRunnerWithFakes(t)
	if err := r.Run(context.Background(), validPlan(t.TempDir()), nil); err != nil {
		t.Fatalf("Run with nil progress: %v", err)
	}
}

func TestRunner_Run_RejectsInvalidRepeatCount(t *testing.T) {
	for _, repeat := range []int{0, -1, -100} {
		r, _, _, _ := newRunnerWithFakes(t)
		p := validPlan(t.TempDir())
		p.RepeatCount = repeat
		if err := r.Run(context.Background(), p, nil); err == nil {
			t.Errorf("Run(RepeatCount=%d): expected error, got nil", repeat)
		}
	}
}

func TestRunner_Run_RejectsInvalidStepInterval(t *testing.T) {
	invalid := []float64{0, -0.1, -100, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, sec := range invalid {
		r, _, _, _ := newRunnerWithFakes(t)
		p := validPlan(t.TempDir())
		p.StepIntervalSeconds = sec
		if err := r.Run(context.Background(), p, nil); err == nil {
			t.Errorf("Run(StepIntervalSeconds=%v): expected error, got nil", sec)
		}
	}
}

func TestRunner_Run_RejectsEmptyOutputDir(t *testing.T) {
	r, _, _, _ := newRunnerWithFakes(t)
	p := validPlan(t.TempDir())
	p.OutputDir = ""
	if err := r.Run(context.Background(), p, nil); err == nil {
		t.Error("Run(empty OutputDir): expected error, got nil")
	}
}

func TestRunner_Run_RejectsEmptyOutputFileName(t *testing.T) {
	r, _, _, _ := newRunnerWithFakes(t)
	p := validPlan(t.TempDir())
	p.OutputFileName = ""
	if err := r.Run(context.Background(), p, nil); err == nil {
		t.Error("Run(empty OutputFileName): expected error, got nil")
	}
}

func TestRunner_Run_RejectsWhitespaceOnlyOutputFileName(t *testing.T) {
	r, _, _, _ := newRunnerWithFakes(t)
	p := validPlan(t.TempDir())
	p.OutputFileName = "   "
	if err := r.Run(context.Background(), p, nil); err == nil {
		t.Error("Run(whitespace-only OutputFileName): expected error, got nil")
	}
}

func TestRunner_Run_RejectsEmptyCaptureRegion(t *testing.T) {
	empties := []image.Rectangle{
		{},
		image.Rect(0, 0, 0, 0),
		image.Rect(10, 10, 10, 20), // zero width
		image.Rect(10, 10, 20, 10), // zero height
	}
	for _, region := range empties {
		r, _, _, _ := newRunnerWithFakes(t)
		p := validPlan(t.TempDir())
		p.CaptureRegion = region
		if err := r.Run(context.Background(), p, nil); err == nil {
			t.Errorf("Run(CaptureRegion=%v): expected error, got nil", region)
		}
	}
}

func TestRunner_Run_ResolvesCollisionByAppendingSuffix(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "test.pdf")
	if err := os.WriteFile(existing, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	r, _, _, _, lastPath := newRunnerWithFakesCapturingPath(t)
	if err := r.Run(context.Background(), validPlan(dir), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := filepath.Join(dir, "test-2.pdf")
	if *lastPath != want {
		t.Errorf("newPdfWriter path = %q, want %q", *lastPath, want)
	}
}

func TestRunner_Run_UsesDesiredPathWhenNoCollision(t *testing.T) {
	dir := t.TempDir()

	r, _, _, _, lastPath := newRunnerWithFakesCapturingPath(t)
	if err := r.Run(context.Background(), validPlan(dir), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := filepath.Join(dir, "test.pdf")
	if *lastPath != want {
		t.Errorf("newPdfWriter path = %q, want %q", *lastPath, want)
	}
}
