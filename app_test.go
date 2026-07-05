package main

import (
	"context"
	"image"
	"regexp"
	"testing"

	"pasha-go/internal/capturerunner"
)

type fakeRunner struct {
	called   bool
	lastPlan capturerunner.Plan
	err      error
}

func (f *fakeRunner) Run(_ context.Context, p capturerunner.Plan) error {
	f.called = true
	f.lastPlan = p
	return f.err
}

func validTestSessionParams() TestSessionParams {
	return TestSessionParams{
		RepeatCount:         1,
		StepIntervalSeconds: 0.1,
		OutputDir:           "/tmp",
		OutputFileName:      "test",
		CaptureRegion:       CaptureRegionInput{X: 10, Y: 20, Width: 100, Height: 50},
	}
}

func TestGreet(t *testing.T) {
	app := NewApp()

	got := app.Greet("World")
	want := "Hello World, It's show time!"

	if got != want {
		t.Errorf("Greet() = %q, want %q", got, want)
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

func TestRunTestSession_PropagatesCaptureRegionToPlan(t *testing.T) {
	r := &fakeRunner{}
	app := newAppWithRunner(r)

	if err := app.RunTestSession(validTestSessionParams()); err != nil {
		t.Fatalf("RunTestSession: %v", err)
	}

	if !r.called {
		t.Fatal("Runner.Run was not called")
	}
	want := image.Rect(10, 20, 110, 70)
	if got := r.lastPlan.CaptureRegion; got != want {
		t.Errorf("Plan.CaptureRegion = %v, want %v", got, want)
	}
}

func TestRunTestSession_SetsAdvanceClickPointToRegionCenter(t *testing.T) {
	r := &fakeRunner{}
	app := newAppWithRunner(r)

	if err := app.RunTestSession(validTestSessionParams()); err != nil {
		t.Fatalf("RunTestSession: %v", err)
	}

	want := image.Pt(60, 45)
	if got := r.lastPlan.AdvanceClickPoint; got != want {
		t.Errorf("Plan.AdvanceClickPoint = %v, want %v", got, want)
	}
}
