package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"regexp"
	"strings"
	"sync"
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

	// started, if non-nil, is closed once Run begins; release, if non-nil,
	// blocks Run until closed. Together they let a test observe an in-flight
	// session.
	started    chan struct{}
	release    chan struct{}
	stopCalled atomic.Bool
}

func (f *fakeRunner) Run(_ context.Context, p session.Plan, onProgress func(current, total int)) (string, error) {
	f.called = true
	f.lastPlan = p
	f.onProgress = onProgress
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

func (f *fakeRunner) Stop() { f.stopCalled.Store(true) }

// recordingEvents captures what the user would have been shown.
type recordingEvents struct {
	mu        sync.Mutex
	progress  [][2]int
	completed int
	failures  []string
}

func (e *recordingEvents) Progress(current, total int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.progress = append(e.progress, [2]int{current, total})
}

func (e *recordingEvents) Completed() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.completed++
}

func (e *recordingEvents) Failed(message string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.failures = append(e.failures, message)
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
	app := newAppWithRunner(&recordingEvents{}, r)

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

// The path may carry a "-2" collision suffix the caller never asked for.
func TestRunCaptureSession_ReturnsResolvedOutputDocumentPath(t *testing.T) {
	r := &fakeRunner{outPath: "/tmp/test-2.pdf"}
	app := newAppWithRunner(&recordingEvents{}, r)

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
	app := newAppWithRunner(&recordingEvents{}, r)

	got, err := app.RunCaptureSession(validCaptureSessionParams())
	if err == nil {
		t.Fatal("RunCaptureSession: expected error, got nil")
	}
	if got != "" {
		t.Errorf("RunCaptureSession() path = %q on error, want empty", got)
	}
}

func TestRunCaptureSession_EmitsHumanReadableFailure(t *testing.T) {
	events := &recordingEvents{}
	r := &fakeRunner{err: fmt.Errorf("%w: exit status 1", session.ErrCapture)}
	app := newAppWithRunner(events, r)

	if _, err := app.RunCaptureSession(validCaptureSessionParams()); err == nil {
		t.Fatal("RunCaptureSession: expected error, got nil")
	}

	if len(events.failures) != 1 {
		t.Fatalf("Failed events = %v, want exactly one", events.failures)
	}
	if want := "Screen Recording permission"; !strings.Contains(events.failures[0], want) {
		t.Errorf("Failed(%q), want it to contain %q", events.failures[0], want)
	}
	// The technical error must not leak to the user (#11).
	if strings.Contains(events.failures[0], "exit status 1") {
		t.Errorf("Failed(%q) leaked the technical error", events.failures[0])
	}
	if events.completed != 0 {
		t.Errorf("Completed fired %d times on a failed session, want 0", events.completed)
	}
}

func TestRunCaptureSession_EmitsCompletedOnSuccess(t *testing.T) {
	events := &recordingEvents{}
	app := newAppWithRunner(events, &fakeRunner{outPath: "/tmp/test.pdf"})

	if _, err := app.RunCaptureSession(validCaptureSessionParams()); err != nil {
		t.Fatalf("RunCaptureSession: %v", err)
	}

	if events.completed != 1 {
		t.Errorf("Completed fired %d times, want 1", events.completed)
	}
	if len(events.failures) != 0 {
		t.Errorf("Failed fired on a successful session: %v", events.failures)
	}
}

func TestRunCaptureSession_ForwardsProgressToEvents(t *testing.T) {
	events := &recordingEvents{}
	r := &fakeRunner{}
	app := newAppWithRunner(events, r)

	if _, err := app.RunCaptureSession(validCaptureSessionParams()); err != nil {
		t.Fatalf("RunCaptureSession: %v", err)
	}

	r.onProgress(1, 10)
	r.onProgress(2, 10)

	want := [][2]int{{1, 10}, {2, 10}}
	if len(events.progress) != len(want) {
		t.Fatalf("Progress events = %v, want %v", events.progress, want)
	}
	for i := range want {
		if events.progress[i] != want[i] {
			t.Errorf("Progress[%d] = %v, want %v", i, events.progress[i], want[i])
		}
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
	app := newAppWithRunner(&recordingEvents{}, r)

	go func() { _, _ = app.RunCaptureSession(validCaptureSessionParams()) }()
	<-started // wait until the session has registered its stop handle

	app.StopSession()
	if !r.stopCalled.Load() {
		t.Fatal("StopSession did not stop the active session")
	}
	close(release)
}

func TestStopSession_NoActiveSession_IsNoOp(t *testing.T) {
	app := newAppWithRunner(&recordingEvents{}, &fakeRunner{})

	// Must not panic when nothing is running.
	app.StopSession()
}

func TestRunCaptureSession_PropagatesAdvanceClickPointFromParams(t *testing.T) {
	r := &fakeRunner{}
	app := newAppWithRunner(&recordingEvents{}, r)

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
