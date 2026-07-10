package main

import (
	"context"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// sessionEvents keeps channel names and JSON payloads out of App: the
// interface speaks Go, wailsEvents speaks the wire.
type sessionEvents interface {
	Progress(current, total int)
	// Completed fires whether the session ran to its Repeat Count or was
	// stopped early.
	Completed()
	// Failed takes a user-facing message; see humanErrorMessage.
	Failed(message string)
}

// Shared with frontend/src/App.tsx.
const (
	eventProgress  = "session:progress"
	eventCompleted = "session:completed"
	eventError     = "session:error"
)

type sessionProgress struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

type sessionError struct {
	Message string `json:"message"`
}

// wailsEvents is constructed in startup, once the runtime context exists.
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

// noopEvents is the App's state before startup runs: no runtime, no frontend,
// nobody to tell. Not a testing affordance — tests inject a recording adapter.
type noopEvents struct{}

func (noopEvents) Progress(int, int) {}
func (noopEvents) Completed()        {}
func (noopEvents) Failed(string)     {}
