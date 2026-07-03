package main

import (
	"context"
	"fmt"
	"image"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"pasha-go/internal/clicker"
	"pasha-go/internal/clock"
	"pasha-go/internal/outputpath"
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

// DefaultOutputFileName returns a timestamp-based default file name (without
// extension) for the Output Document, matching the PRD format:
// "pasha-YYYY-MM-DD_HH-MM".
func (a *App) DefaultOutputFileName() string {
	return time.Now().Format("pasha-2006-01-02_15-04")
}

// ChooseOutputDirectory opens the native folder-selection dialog and
// returns the absolute path the user picked. Returns an empty string with
// no error if the user cancels. This method is a thin Wails-runtime
// wrapper and is verified via manual QA (see PRD Testing Decisions).
func (a *App) ChooseOutputDirectory() (string, error) {
	ctx := a.ctx
	if ctx == nil {
		return "", fmt.Errorf("app context not initialized")
	}
	return wailsRuntime.OpenDirectoryDialog(ctx, wailsRuntime.OpenDialogOptions{
		Title: "Choose output folder",
	})
}

// TestSessionParams bundles the frontend-supplied Capture Session inputs.
// Wails auto-generates a matching TypeScript class so the frontend can
// construct this object directly.
type TestSessionParams struct {
	RepeatCount         int     `json:"repeatCount"`
	StepIntervalSeconds float64 `json:"stepIntervalSeconds"`
	OutputDir           string  `json:"outputDir"`
	OutputFileName      string  `json:"outputFileName"`
}

// RunTestSession runs a Capture Session as a tracer-bullet end-to-end test.
// The Capture Region and Advance Click Point are still hardcoded (full
// primary display, screen center). The Output Document is written to
// OutputDir/OutputFileName.pdf, with numeric suffixes (-2, -3, ...) added
// on collision so existing files are never overwritten.
// Returns an error immediately if any input is invalid.
func (a *App) RunTestSession(params TestSessionParams) error {
	if params.RepeatCount < 1 {
		return fmt.Errorf("repeat count must be >= 1, got %d", params.RepeatCount)
	}
	if math.IsNaN(params.StepIntervalSeconds) || math.IsInf(params.StepIntervalSeconds, 0) || params.StepIntervalSeconds <= 0 {
		return fmt.Errorf("step interval must be a positive finite number of seconds, got %v", params.StepIntervalSeconds)
	}
	if strings.TrimSpace(params.OutputDir) == "" {
		return fmt.Errorf("output directory must not be empty")
	}
	if strings.TrimSpace(params.OutputFileName) == "" {
		return fmt.Errorf("output file name must not be empty")
	}

	desired := filepath.Join(params.OutputDir, params.OutputFileName+".pdf")
	outPath, err := outputpath.Resolve(desired)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	bounds := screenshot.GetDisplayBounds(0)
	center := image.Pt(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2)

	pdf, err := pdfwriter.New(outPath)
	if err != nil {
		return fmt.Errorf("pdfwriter: %w", err)
	}

	cs := session.New(session.Config{
		CaptureRegion:     bounds,
		AdvanceClickPoint: center,
		RepeatCount:       params.RepeatCount,
		StepInterval:      time.Duration(params.StepIntervalSeconds * float64(time.Second)),
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
