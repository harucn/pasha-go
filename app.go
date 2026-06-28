package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/kbinani/screenshot"

	"pasha-go/internal/clicker"
	"pasha-go/internal/clock"
	"pasha-go/internal/pdfwriter"
	"pasha-go/internal/screener"
	"pasha-go/internal/session"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// RunTestSession runs a hardcoded Capture Session as a tracer-bullet
// end-to-end test: 3 Capture Steps, full primary display as Capture Region,
// screen center as Advance Click Point, 1s Step Interval, output to
// ~/Desktop/pasha-tracer.pdf.
func (a *App) RunTestSession() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	outPath := filepath.Join(home, "Desktop", "pasha-tracer.pdf")

	bounds := screenshot.GetDisplayBounds(0)
	center := image.Pt(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2)

	pdf, err := pdfwriter.New(outPath)
	if err != nil {
		return fmt.Errorf("pdfwriter: %w", err)
	}

	cs := session.New(session.Config{
		CaptureRegion:     bounds,
		AdvanceClickPoint: center,
		RepeatCount:       3,
		StepInterval:      1 * time.Second,
		Screener:          screener.New(),
		Clicker:           clicker.New(),
		PdfWriter:         pdf,
		Clock:             clock.New(),
	})

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return cs.Start(ctx)
}
