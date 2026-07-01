package main

import "testing"

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
		if err := app.RunTestSession(repeat); err == nil {
			t.Errorf("RunTestSession(%d): expected error, got nil", repeat)
		}
	}
}
