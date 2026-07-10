package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"pasha-go/internal/session"
)

type fakeRunner struct {
	called     bool
	lastPlan   session.Plan
	onProgress func(current, total int)
	outPath    string
	err        error

	// started, if non-nil, is closed once Run has registered its stop
	// function via onStart; release, if non-nil, blocks Run until closed.
	// Together they let a test observe an in-flight session.
	started    chan struct{}
	release    chan struct{}
	stopCalled atomic.Bool
}

func (f *fakeRunner) Run(_ context.Context, p session.Plan, onProgress func(current, total int), onStart func(stop func())) (string, error) {
	f.called = true
	f.lastPlan = p
	f.onProgress = onProgress
	if onStart != nil {
		onStart(func() { f.stopCalled.Store(true) })
	}
	if f.started != nil {
		close(f.started)
	}
	if f.release != nil {
		<-f.release
	}
	if f.err != nil {
		return "", f.err
	}
	return f.outPath, nil
}

func validCaptureSessionParams() CaptureSessionParams {
	return CaptureSessionParams{
		RepeatCount:         1,
		StepIntervalSeconds: 0.1,
		OutputDir:           "/tmp",
		OutputFileName:      "test",
		CaptureRegion:       CaptureRegionInput{X: 10, Y: 20, Width: 100, Height: 50},
		AdvanceClickPoint:   ClickPointInput{X: 60, Y: 45},
	}
}

func TestDefaultOutputFileName_MatchesTimestampFormat(t *testing.T) {
	app := NewApp()

	got := app.DefaultOutputFileName()
	re := regexp.MustCompile(`^pasha-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}$`)
	if !re.MatchString(got) {
		t.Errorf("DefaultOutputFileName() = %q, want match %s", got, re)
	}
}

func TestRunCaptureSession_PropagatesCaptureRegionToPlan(t *testing.T) {
	r := &fakeRunner{}
	app := newAppWithRunner(r)

	if _, err := app.RunCaptureSession(validCaptureSessionParams()); err != nil {
		t.Fatalf("RunCaptureSession: %v", err)
	}

	if !r.called {
		t.Fatal("Runner.Run was not called")
	}
	want := image.Rect(10, 20, 110, 70)
	if got := r.lastPlan.CaptureRegion; got != want {
		t.Errorf("Plan.CaptureRegion = %v, want %v", got, want)
	}
}

func TestRunCaptureSession_SuppliesProgressCallbackToRunner(t *testing.T) {
	r := &fakeRunner{}
	app := newAppWithRunner(r)

	if _, err := app.RunCaptureSession(validCaptureSessionParams()); err != nil {
		t.Fatalf("RunCaptureSession: %v", err)
	}

	if r.onProgress == nil {
		t.Fatal("Runner.Run received nil onProgress; expected a progress callback")
	}
	// The callback must not panic when the app has no runtime context
	// (as in unit tests, where startup was never called).
	r.onProgress(1, 10)
}

// The Output Document path the Runner resolved must reach the frontend
// verbatim: it may carry a "-2" collision suffix the caller never asked for.
func TestRunCaptureSession_ReturnsResolvedOutputDocumentPath(t *testing.T) {
	r := &fakeRunner{outPath: "/tmp/test-2.pdf"}
	app := newAppWithRunner(r)

	got, err := app.RunCaptureSession(validCaptureSessionParams())
	if err != nil {
		t.Fatalf("RunCaptureSession: %v", err)
	}
	if want := "/tmp/test-2.pdf"; got != want {
		t.Errorf("RunCaptureSession() path = %q, want %q", got, want)
	}
}

func TestRunCaptureSession_ReturnsEmptyPathOnError(t *testing.T) {
	r := &fakeRunner{outPath: "/tmp/test.pdf", err: errors.New("boom")}
	app := newAppWithRunner(r)

	got, err := app.RunCaptureSession(validCaptureSessionParams())
	if err == nil {
		t.Fatal("RunCaptureSession: expected error, got nil")
	}
	if got != "" {
		t.Errorf("RunCaptureSession() path = %q on error, want empty", got)
	}
}

func TestHumanErrorMessage_ClassifiesByOrigin(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		wantSub string
	}{
		{"capture", fmt.Errorf("%w: x", session.ErrCapture), "Screen capture failed"},
		{"pdf", fmt.Errorf("%w: x", session.ErrPdfWrite), "PDF"},
		{"click", fmt.Errorf("%w: x", session.ErrClick), "Auto-click failed"},
		{"unknown", errors.New("mystery"), "Something went wrong"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := humanErrorMessage(tc.err)
			if !strings.Contains(got, tc.wantSub) {
				t.Errorf("humanErrorMessage(%v) = %q, want to contain %q", tc.err, got, tc.wantSub)
			}
		})
	}
}

func TestStopSession_StopsActiveSession(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	r := &fakeRunner{started: started, release: release}
	app := newAppWithRunner(r)

	go func() { _, _ = app.RunCaptureSession(validCaptureSessionParams()) }()
	<-started // wait until the session has registered its stop handle

	app.StopSession()
	if !r.stopCalled.Load() {
		t.Fatal("StopSession did not stop the active session")
	}
	close(release)
}

func TestStopSession_NoActiveSession_IsNoOp(t *testing.T) {
	app := newAppWithRunner(&fakeRunner{})

	// Must not panic when nothing is running.
	app.StopSession()
}

func TestRunCaptureSession_PropagatesAdvanceClickPointFromParams(t *testing.T) {
	r := &fakeRunner{}
	app := newAppWithRunner(r)

	params := validCaptureSessionParams()
	params.AdvanceClickPoint = ClickPointInput{X: 200, Y: 300}

	if _, err := app.RunCaptureSession(params); err != nil {
		t.Fatalf("RunCaptureSession: %v", err)
	}

	want := image.Pt(200, 300)
	if got := r.lastPlan.AdvanceClickPoint; got != want {
		t.Errorf("Plan.AdvanceClickPoint = %v, want %v", got, want)
	}
}
