package main

import (
	"context"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// sessionEvents is the seam through which App tells the frontend how a
// Capture Session is going. Two adapters satisfy it in production —
// wailsEvents once the runtime is up, noopEvents before that — and a
// recording fake satisfies it in tests, so every message the user can see
// is assertable.
//
// The interface speaks Go. Channel names and JSON payloads are wire format
// and live in wailsEvents.
type sessionEvents interface {
	// Progress reports that a Capture Step completed, out of total.
	Progress(current, total int)
	// Completed reports that the Capture Session ended, whether it ran to
	// its Repeat Count or was stopped early.
	Completed()
	// Failed reports that the Capture Session aborted. message is
	// user-facing and human-readable; see humanErrorMessage.
	Failed(message string)
}

// Event channel names shared with frontend/src/App.tsx.
const (
	eventProgress  = "session:progress"
	eventCompleted = "session:completed"
	eventError     = "session:error"
)

// sessionProgress is the payload on eventProgress. The frontend renders
// "Current / Total" (issue #08).
type sessionProgress struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

// sessionError is the payload on eventError. Message drives the bar's red
// error display (issue #11).
type sessionError struct {
	Message string `json:"message"`
}

// wailsEvents emits over the Wails runtime bridge. Constructed in startup,
// once the runtime context exists.
type wailsEvents struct {
	ctx context.Context
}

func (e wailsEvents) Progress(current, total int) {
	wailsRuntime.EventsEmit(e.ctx, eventProgress, sessionProgress{
		Current: current,
		Total:   total,
	})
}

func (e wailsEvents) Completed() {
	wailsRuntime.EventsEmit(e.ctx, eventCompleted)
}

func (e wailsEvents) Failed(message string) {
	wailsRuntime.EventsEmit(e.ctx, eventError, sessionError{Message: message})
}

// noopEvents drops everything. It is the App's state before startup runs:
// there is no runtime and no frontend, so there is nobody to tell. This is
// not a testing affordance — tests inject a recording adapter instead.
type noopEvents struct{}

func (noopEvents) Progress(int, int) {}
func (noopEvents) Completed()        {}
func (noopEvents) Failed(string)     {}
