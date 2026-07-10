package session_test

import (
	"context"
	"errors"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pasha-go/internal/session"
)

// StepIntervalSeconds 0.01 == the 10ms the fake Clock asserts on.
func validPlan(dir string) session.Plan {
	return session.Plan{
		RepeatCount:         3,
		StepIntervalSeconds: 0.01,
		CaptureRegion:       image.Rect(0, 0, 100, 100),
		AdvanceClickPoint:   image.Pt(50, 50),
		OutputDir:           dir,
		OutputFileName:      "test",
	}
}

// The returned pointer captures the path the PdfWriterFactory was asked for.
func runnerWith(scr session.Screener, clk session.Clicker, pdf session.PdfWriter, clock session.Clock) (*session.Runner, *string) {
	var lastPath string
	newPdf := func(path string) (session.PdfWriter, error) {
		lastPath = path
		return pdf, nil
	}
	return session.NewRunner(scr, clk, clock, newPdf), &lastPath
}

func run(t *testing.T, p session.Plan, scr *fakeScreener, clk *fakeClicker, pdf *fakePdfWriter, clock *fakeClock) (string, error) {
	t.Helper()
	r, _ := runnerWith(scr, clk, pdf, clock)
	return r.Run(context.Background(), p, nil)
}

func TestRun_HappyPath_RunsRepeatCountSteps(t *testing.T) {
	scr, clk, pdf, clock := &fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{}

	p := validPlan(t.TempDir())
	p.RepeatCount = 5
	if _, err := run(t, p, scr, clk, pdf, clock); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if scr.calls != 5 {
		t.Errorf("Screener.Capture calls = %d, want 5", scr.calls)
	}
	if pdf.appendCalls != 5 {
		t.Errorf("PdfWriter.AppendPage calls = %d, want 5", pdf.appendCalls)
	}
	if clk.calls != 5 {
		t.Errorf("Clicker.Click calls = %d, want 5", clk.calls)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("PdfWriter.Close calls = %d, want 1", pdf.closeCalls)
	}
}

func TestRun_StepOrder(t *testing.T) {
	log := &callLog{}
	scr := &fakeScreener{log: log}
	clk := &fakeClicker{log: log}
	pdf := &fakePdfWriter{log: log}
	clock := &fakeClock{log: log}

	p := validPlan(t.TempDir())
	p.RepeatCount = 1
	if _, err := run(t, p, scr, clk, pdf, clock); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := []string{"Clicker", "Sleep", "Screener", "AppendPage", "Close"}
	got := log.snapshot()
	if len(got) != len(want) {
		t.Fatalf("call log = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("call[%d] = %q, want %q (full %v)", i, got[i], want[i], got)
		}
	}
}

func TestRun_SleepsForStepInterval(t *testing.T) {
	clock := &fakeClock{}

	if _, err := run(t, validPlan(t.TempDir()), &fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, clock); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(clock.sleeps) != 3 {
		t.Fatalf("Sleep calls = %d, want 3", len(clock.sleeps))
	}
	for i, d := range clock.sleeps {
		if d != 10*time.Millisecond {
			t.Errorf("Sleep[%d] = %v, want 10ms", i, d)
		}
	}
}

func TestRun_StopFromAnotherGoroutine(t *testing.T) {
	scr, clk, pdf, clock := &fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{}
	r, _ := runnerWith(scr, clk, pdf, clock)

	// The fake Clock stands in for the user pressing Stop mid-session.
	clock.hook = func(context.Context) { r.Stop() }

	p := validPlan(t.TempDir())
	p.RepeatCount = 100
	if _, err := r.Run(context.Background(), p, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if scr.calls != 1 {
		t.Errorf("Screener.Capture = %d, want 1 (stop after first step)", scr.calls)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("Close calls = %d, want 1", pdf.closeCalls)
	}
}

// A stopped session still saved its pages, so it reports the path it wrote.
func TestRun_StopStillReturnsOutputDocumentPath(t *testing.T) {
	dir := t.TempDir()
	clock := &fakeClock{}
	r, _ := runnerWith(&fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, clock)

	clock.hook = func(context.Context) { r.Stop() }

	p := validPlan(dir)
	p.RepeatCount = 100
	got, err := r.Run(context.Background(), p, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if want := filepath.Join(dir, "test.pdf"); got != want {
		t.Errorf("Run() path = %q, want %q", got, want)
	}
}

// Stop must not arm a pending stop that kills the next session.
func TestStop_WithNoSessionRunning_IsNoOpAndDoesNotArm(t *testing.T) {
	scr := &fakeScreener{}
	r, _ := runnerWith(scr, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{})

	r.Stop() // nothing is in flight

	p := validPlan(t.TempDir())
	p.RepeatCount = 3
	if _, err := r.Run(context.Background(), p, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if scr.calls != 3 {
		t.Errorf("Screener.Capture = %d, want 3 (an earlier Stop must not affect this session)", scr.calls)
	}
}

// Does not pin down that Run released the session: stopping a finished session
// is unobservable either way.
func TestStop_AfterRunReturns_IsHarmless(t *testing.T) {
	scr := &fakeScreener{}
	r, _ := runnerWith(scr, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{})

	if _, err := r.Run(context.Background(), validPlan(t.TempDir()), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	r.Stop() // must not panic
	if scr.calls != 3 {
		t.Errorf("Screener.Capture = %d, want 3", scr.calls)
	}
}

func TestRun_CollaboratorErrorAborts(t *testing.T) {
	boom := errSentinel("boom")

	tests := []struct {
		name                      string
		scr                       *fakeScreener
		clk                       *fakeClicker
		pdf                       *fakePdfWriter
		wantCaptures, wantAppends int
		wantClicks                int
		wantSentinel              error
	}{
		{
			name: "screener", scr: &fakeScreener{err: boom}, clk: &fakeClicker{}, pdf: &fakePdfWriter{},
			wantCaptures: 1, wantAppends: 0, wantClicks: 1, wantSentinel: session.ErrCapture,
		},
		{
			name: "append", scr: &fakeScreener{}, clk: &fakeClicker{}, pdf: &fakePdfWriter{appendErr: boom},
			wantCaptures: 1, wantAppends: 1, wantClicks: 1, wantSentinel: session.ErrPdfWrite,
		},
		{
			name: "clicker", scr: &fakeScreener{}, clk: &fakeClicker{err: boom}, pdf: &fakePdfWriter{},
			wantCaptures: 0, wantAppends: 0, wantClicks: 1, wantSentinel: session.ErrClick,
		},
		{
			name: "close", scr: &fakeScreener{}, clk: &fakeClicker{}, pdf: &fakePdfWriter{closeErr: boom},
			wantCaptures: 3, wantAppends: 3, wantClicks: 3, wantSentinel: session.ErrPdfWrite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := run(t, validPlan(t.TempDir()), tt.scr, tt.clk, tt.pdf, &fakeClock{})
			if err == nil {
				t.Fatal("Run: expected error, got nil")
			}
			if !errors.Is(err, tt.wantSentinel) {
				t.Errorf("error %v is not %v", err, tt.wantSentinel)
			}
			if got != "" {
				t.Errorf("Run() path = %q on error, want empty", got)
			}
			if tt.scr.calls != tt.wantCaptures || tt.pdf.appendCalls != tt.wantAppends || tt.clk.calls != tt.wantClicks {
				t.Errorf("calls: capture=%d append=%d click=%d; want %d/%d/%d",
					tt.scr.calls, tt.pdf.appendCalls, tt.clk.calls,
					tt.wantCaptures, tt.wantAppends, tt.wantClicks)
			}
			// Always closed: a crash still leaves a valid partial PDF (ADR-0001).
			if tt.pdf.closeCalls != 1 {
				t.Errorf("Close calls = %d, want 1", tt.pdf.closeCalls)
			}
		})
	}
}

func TestRun_ContextCancelAborts(t *testing.T) {
	scr, clk, pdf, clock := &fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{}
	r, _ := runnerWith(scr, clk, pdf, clock)

	ctx, cancel := context.WithCancel(context.Background())
	clock.hook = func(context.Context) { cancel() }

	p := validPlan(t.TempDir())
	p.RepeatCount = 100
	got, err := r.Run(ctx, p, nil)
	if err == nil {
		t.Fatal("Run: expected ctx error")
	}
	if got != "" {
		t.Errorf("Run() path = %q on cancel, want empty", got)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("Close calls = %d, want 1", pdf.closeCalls)
	}
	// Cancellation lands in the first Sleep, which precedes the first Capture.
	if scr.calls != 0 {
		t.Errorf("Screener.Capture = %d, want 0 (cancelled during the first Sleep)", scr.calls)
	}
	if clk.calls != 1 {
		t.Errorf("Clicker.Click = %d, want 1", clk.calls)
	}
}

func TestRun_ProgressCalledPerCompletedStep(t *testing.T) {
	r, _ := runnerWith(&fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{})

	var got [][2]int
	onProgress := func(current, total int) { got = append(got, [2]int{current, total}) }

	if _, err := r.Run(context.Background(), validPlan(t.TempDir()), onProgress); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := [][2]int{{1, 3}, {2, 3}, {3, 3}}
	if len(got) != len(want) {
		t.Fatalf("Progress calls = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Progress[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestRun_ProgressNotCalledOnAbortedStep(t *testing.T) {
	r, _ := runnerWith(&fakeScreener{err: errSentinel("boom")}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{})

	var calls int
	onProgress := func(int, int) { calls++ }

	if _, err := r.Run(context.Background(), validPlan(t.TempDir()), onProgress); err == nil {
		t.Fatal("Run: expected error")
	}
	if calls != 0 {
		t.Errorf("Progress calls = %d, want 0 (step never completed)", calls)
	}
}

func TestRun_NilCallbacksAreAllowed(t *testing.T) {
	if _, err := run(t, validPlan(t.TempDir()), &fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{}); err != nil {
		t.Fatalf("Run with nil callbacks: %v", err)
	}
}

func TestRun_RejectsInvalidPlan(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*session.Plan)
	}{
		{"zero repeat count", func(p *session.Plan) { p.RepeatCount = 0 }},
		{"negative repeat count", func(p *session.Plan) { p.RepeatCount = -1 }},
		{"zero step interval", func(p *session.Plan) { p.StepIntervalSeconds = 0 }},
		{"negative step interval", func(p *session.Plan) { p.StepIntervalSeconds = -1 }},
		{"NaN step interval", func(p *session.Plan) { p.StepIntervalSeconds = math.NaN() }},
		{"Inf step interval", func(p *session.Plan) { p.StepIntervalSeconds = math.Inf(1) }},
		{"empty output dir", func(p *session.Plan) { p.OutputDir = "" }},
		{"empty output file name", func(p *session.Plan) { p.OutputFileName = "" }},
		{"whitespace output file name", func(p *session.Plan) { p.OutputFileName = "   " }},
		{"empty capture region", func(p *session.Plan) { p.CaptureRegion = image.Rect(0, 0, 0, 0) }},
		{"zero width region", func(p *session.Plan) { p.CaptureRegion = image.Rect(10, 10, 10, 20) }},
		{"zero height region", func(p *session.Plan) { p.CaptureRegion = image.Rect(10, 10, 20, 10) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scr, clk, pdf := &fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}
			p := validPlan(t.TempDir())
			tt.mutate(&p)

			got, err := run(t, p, scr, clk, pdf, &fakeClock{})
			if err == nil {
				t.Fatal("Run: expected error, got nil")
			}
			if got != "" {
				t.Errorf("Run() path = %q on invalid plan, want empty", got)
			}
			// Rejected before anything touches the screen.
			if clk.calls != 0 || scr.calls != 0 || pdf.appendCalls != 0 {
				t.Errorf("collaborators ran on an invalid Plan: click=%d capture=%d append=%d",
					clk.calls, scr.calls, pdf.appendCalls)
			}
		})
	}
}

// The Output Document must never overwrite an existing PDF.
func TestRun_ResolvesCollisionByAppendingSuffix(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.pdf"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	r, lastPath := runnerWith(&fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{})
	got, err := r.Run(context.Background(), validPlan(dir), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := filepath.Join(dir, "test-2.pdf")
	if got != want {
		t.Errorf("Run() path = %q, want %q", got, want)
	}
	if *lastPath != want {
		t.Errorf("PdfWriterFactory path = %q, want %q", *lastPath, want)
	}
}

func TestRun_UsesDesiredPathWhenNoCollision(t *testing.T) {
	dir := t.TempDir()

	r, lastPath := runnerWith(&fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{})
	got, err := r.Run(context.Background(), validPlan(dir), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := filepath.Join(dir, "test.pdf")
	if got != want {
		t.Errorf("Run() path = %q, want %q", got, want)
	}
	if *lastPath != want {
		t.Errorf("PdfWriterFactory path = %q, want %q", *lastPath, want)
	}
}

type errSentinel string

func (e errSentinel) Error() string { return string(e) }
