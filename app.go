package main

import (
	"context"
	"fmt"
	"image"
	"time"

	"github.com/kbinani/screenshot"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"pasha-go/internal/capturerunner"
	"pasha-go/internal/clicker"
	"pasha-go/internal/clock"
	"pasha-go/internal/screener"
)

// App is the Wails-bound adapter. It holds the runtime context and a
// pre-wired Runner; every method delegates its work behind the seam so
// the adapter itself stays thin.
type App struct {
	ctx    context.Context
	runner *capturerunner.Runner
}

// NewApp constructs the App together with a Runner wired to the real
// desktop collaborators.
func NewApp() *App {
	runner := capturerunner.New(
		screener.New(),
		clicker.New(),
		clock.New(),
		capturerunner.DefaultPdfWriterFactory,
	)
	return &App{runner: runner}
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

// RunTestSession builds a Capture Session Plan from frontend inputs
// (using primary-display defaults for Capture Region and Advance Click
// Point) and delegates to the Runner. All validation, path resolution
// and session construction live behind the Runner interface.
func (a *App) RunTestSession(params TestSessionParams) error {
	bounds := screenshot.GetDisplayBounds(0)
	center := image.Pt(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2)

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	return a.runner.Run(ctx, capturerunner.Plan{
		RepeatCount:         params.RepeatCount,
		StepIntervalSeconds: params.StepIntervalSeconds,
		CaptureRegion:       bounds,
		AdvanceClickPoint:   center,
		OutputDir:           params.OutputDir,
		OutputFileName:      params.OutputFileName,
	})
}
