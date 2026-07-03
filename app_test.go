package main

import (
	"math"
	"regexp"
	"testing"
)

func TestGreet(t *testing.T) {
	app := NewApp()

	got := app.Greet("World")
	want := "Hello World, It's show time!"

	if got != want {
		t.Errorf("Greet() = %q, want %q", got, want)
	}
}

func TestRunTestSession_RejectsInvalidRepeatCount(t *testing.T) {
	app := NewApp()

	for _, repeat := range []int{0, -1, -100} {
		params := TestSessionParams{RepeatCount: repeat, StepIntervalSeconds: 1.0}
		if err := app.RunTestSession(params); err == nil {
			t.Errorf("RunTestSession(RepeatCount=%d): expected error, got nil", repeat)
		}
	}
}

func TestRunTestSession_RejectsInvalidStepInterval(t *testing.T) {
	app := NewApp()

	invalid := []float64{0, -0.1, -100, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, sec := range invalid {
		params := TestSessionParams{RepeatCount: 1, StepIntervalSeconds: sec}
		if err := app.RunTestSession(params); err == nil {
			t.Errorf("RunTestSession(StepIntervalSeconds=%v): expected error, got nil", sec)
		}
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

func TestRunTestSession_RejectsEmptyOutputDir(t *testing.T) {
	app := NewApp()
	params := TestSessionParams{
		RepeatCount:         1,
		StepIntervalSeconds: 1.0,
		OutputDir:           "",
		OutputFileName:      "pasha-2026-06-28_15-30",
	}
	if err := app.RunTestSession(params); err == nil {
		t.Error("RunTestSession(empty OutputDir): expected error, got nil")
	}
}

func TestRunTestSession_RejectsEmptyOutputFileName(t *testing.T) {
	app := NewApp()
	params := TestSessionParams{
		RepeatCount:         1,
		StepIntervalSeconds: 1.0,
		OutputDir:           t.TempDir(),
		OutputFileName:      "",
	}
	if err := app.RunTestSession(params); err == nil {
		t.Error("RunTestSession(empty OutputFileName): expected error, got nil")
	}
}

func TestRunTestSession_RejectsWhitespaceOnlyOutputFileName(t *testing.T) {
	app := NewApp()
	params := TestSessionParams{
		RepeatCount:         1,
		StepIntervalSeconds: 1.0,
		OutputDir:           t.TempDir(),
		OutputFileName:      "   ",
	}
	if err := app.RunTestSession(params); err == nil {
		t.Error("RunTestSession(whitespace-only OutputFileName): expected error, got nil")
	}
}
