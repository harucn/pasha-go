package session_test

import (
	"context"
	"errors"
	"image"
	"math"
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

// openArgs records what the Runner asked the Output Document to be opened as.
type openArgs struct {
	dir      string
	fileName string
}

func runnerWith(scr session.Screener, clk session.Clicker, doc session.OutputDocument, clock session.Clock) (*session.Runner, *openArgs) {
	var asked openArgs
	openDoc := func(dir, fileName string) (session.OutputDocument, error) {
		asked = openArgs{dir: dir, fileName: fileName}
		return doc, nil
	}
	return session.NewRunner(scr, clk, clock, openDoc), &asked
}

func run(t *testing.T, p session.Plan, scr *fakeScreener, clk *fakeClicker, doc *fakeDocument, clock *fakeClock) (string, error) {
	t.Helper()
	r, _ := runnerWith(scr, clk, doc, clock)
	return r.Run(context.Background(), p, nil)
}

func TestRun_HappyPath_RunsRepeatCountSteps(t *testing.T) {
	scr, clk, pdf, clock := &fakeScreener{}, &fakeClicker{}, &fakeDocument{}, &fakeClock{}

	p := validPlan(t.TempDir())
	p.RepeatCount = 5
	if _, err := run(t, p, scr, clk, pdf, clock); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if scr.calls != 5 {
		t.Errorf("Screener.Capture calls = %d, want 5", scr.calls)
	}
	if pdf.appendCalls != 5 {
		t.Errorf("OutputDocument.AppendPage calls = %d, want 5", pdf.appendCalls)
	}
	if clk.calls != 5 {
		t.Errorf("Clicker.Click calls = %d, want 5", clk.calls)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("OutputDocument.Close calls = %d, want 1", pdf.closeCalls)
	}
}

func TestRun_StepOrder(t *testing.T) {
	log := &callLog{}
	scr := &fakeScreener{log: log}
	clk := &fakeClicker{log: log}
	pdf := &fakeDocument{log: log}
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

	if _, err := run(t, validPlan(t.TempDir()), &fakeScreener{}, &fakeClicker{}, &fakeDocument{}, clock); err != nil {
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
	scr, clk, pdf, clock := &fakeScreener{}, &fakeClicker{}, &fakeDocument{}, &fakeClock{}
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
	landed := filepath.Join(dir, "test.pdf")
	clock := &fakeClock{}
	r, _ := runnerWith(&fakeScreener{}, &fakeClicker{}, &fakeDocument{path: landed}, clock)

	clock.hook = func(context.Context) { r.Stop() }

	p := validPlan(dir)
	p.RepeatCount = 100
	got, err := r.Run(context.Background(), p, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got != landed {
		t.Errorf("Run() path = %q, want %q", got, landed)
	}
}

// Stop must not arm a pending stop that kills the next session.
func TestStop_WithNoSessionRunning_IsNoOpAndDoesNotArm(t *testing.T) {
	scr := &fakeScreener{}
	r, _ := runnerWith(scr, &fakeClicker{}, &fakeDocument{}, &fakeClock{})

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
	r, _ := runnerWith(scr, &fakeClicker{}, &fakeDocument{}, &fakeClock{})

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
		pdf                       *fakeDocument
		wantCaptures, wantAppends int
		wantClicks                int
		wantSentinel              error
	}{
		{
			name: "screener", scr: &fakeScreener{err: boom}, clk: &fakeClicker{}, pdf: &fakeDocument{},
			wantCaptures: 1, wantAppends: 0, wantClicks: 1, wantSentinel: session.ErrCapture,
		},
		{
			name: "append", scr: &fakeScreener{}, clk: &fakeClicker{}, pdf: &fakeDocument{appendErr: boom},
			wantCaptures: 1, wantAppends: 1, wantClicks: 1, wantSentinel: session.ErrPdfWrite,
		},
		{
			name: "clicker", scr: &fakeScreener{}, clk: &fakeClicker{err: boom}, pdf: &fakeDocument{},
			wantCaptures: 0, wantAppends: 0, wantClicks: 1, wantSentinel: session.ErrClick,
		},
		{
			name: "close", scr: &fakeScreener{}, clk: &fakeClicker{}, pdf: &fakeDocument{closeErr: boom},
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
	scr, clk, pdf, clock := &fakeScreener{}, &fakeClicker{}, &fakeDocument{}, &fakeClock{}
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
	r, _ := runnerWith(&fakeScreener{}, &fakeClicker{}, &fakeDocument{}, &fakeClock{})

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
	r, _ := runnerWith(&fakeScreener{err: errSentinel("boom")}, &fakeClicker{}, &fakeDocument{}, &fakeClock{})

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
	if _, err := run(t, validPlan(t.TempDir()), &fakeScreener{}, &fakeClicker{}, &fakeDocument{}, &fakeClock{}); err != nil {
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
			scr, clk, pdf := &fakeScreener{}, &fakeClicker{}, &fakeDocument{}
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

// The Runner hands over the Plan's dir and file name verbatim. Extension and
// collision suffix belong to the Output Document, not here.
func TestRun_OpensOutputDocumentWithPlanDirAndFileName(t *testing.T) {
	dir := t.TempDir()

	r, asked := runnerWith(&fakeScreener{}, &fakeClicker{}, &fakeDocument{}, &fakeClock{})
	if _, err := r.Run(context.Background(), validPlan(dir), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if asked.dir != dir {
		t.Errorf("opened dir = %q, want %q", asked.dir, dir)
	}
	if want := "test"; asked.fileName != want {
		t.Errorf("opened fileName = %q, want %q (Run must not add the extension)", asked.fileName, want)
	}
}

// The Output Document may land on a collision-suffixed path. Run reports where
// it actually landed, never the path the Plan asked for.
func TestRun_ReturnsThePathTheOutputDocumentLandedOn(t *testing.T) {
	dir := t.TempDir()
	landed := filepath.Join(dir, "test-2.pdf")

	r, _ := runnerWith(&fakeScreener{}, &fakeClicker{}, &fakeDocument{path: landed}, &fakeClock{})
	got, err := r.Run(context.Background(), validPlan(dir), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got != landed {
		t.Errorf("Run() path = %q, want %q", got, landed)
	}
}

// An Output Document that cannot be opened aborts before any Capture Step.
func TestRun_OpenOutputDocumentFails_AbortsWithoutCapturing(t *testing.T) {
	scr, clk := &fakeScreener{}, &fakeClicker{}
	sentinel := errors.New("disk full")
	openDoc := func(string, string) (session.OutputDocument, error) { return nil, sentinel }
	r := session.NewRunner(scr, clk, &fakeClock{}, openDoc)

	got, err := r.Run(context.Background(), validPlan(t.TempDir()), nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("Run() error = %v, want %v", err, sentinel)
	}
	if got != "" {
		t.Errorf("Run() path = %q, want empty", got)
	}
	if clk.calls != 0 || scr.calls != 0 {
		t.Errorf("collaborators ran: click=%d capture=%d", clk.calls, scr.calls)
	}
}

type errSentinel string

func (e errSentinel) Error() string { return string(e) }
