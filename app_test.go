package main

import (
	"math"
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
		if err := app.RunTestSession(repeat, 1.0); err == nil {
			t.Errorf("RunTestSession(%d, 1.0): expected error, got nil", repeat)
		}
	}
}

func TestRunTestSession_RejectsInvalidStepInterval(t *testing.T) {
	app := NewApp()

	invalid := []float64{0, -0.1, -100, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, sec := range invalid {
		if err := app.RunTestSession(1, sec); err == nil {
			t.Errorf("RunTestSession(1, %v): expected error, got nil", sec)
		}
	}
}
