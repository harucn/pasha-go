package main

import (
	"context"
	"fmt"
	"image"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"pasha-go/internal/appwindow"
	"pasha-go/internal/capturerunner"
	"pasha-go/internal/clicker"
	"pasha-go/internal/clock"
	"pasha-go/internal/screener"
)

// sessionRunner is the seam through which App delegates Capture Session
// assembly. In production this is *capturerunner.Runner; tests can
// substitute a fake.
type sessionRunner interface {
	Run(ctx context.Context, p capturerunner.Plan) error
}

// App is the Wails-bound adapter. It holds the runtime context and a
// pre-wired Runner; every method delegates its work behind the seam so
// the adapter itself stays thin.
type App struct {
	ctx    context.Context
	runner sessionRunner
}

// NewApp constructs the App together with a Runner wired to the real
// desktop collaborators.
func NewApp() *App {
	return newAppWithRunner(capturerunner.New(
		screener.New(),
		clicker.New(),
		clock.New(),
		capturerunner.DefaultPdfWriterFactory,
	))
}

// newAppWithRunner injects the Runner. Kept package-private so tests
// can substitute a fake without exposing the seam publicly.
func newAppWithRunner(r sessionRunner) *App {
	return &App{runner: r}
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

// CaptureRegionInput carries the user-selected Capture Region in
// screen coordinates. X/Y is the top-left corner, Width/Height is the
// size. Sent from the frontend after the drag-selection overlay
// completes.
type CaptureRegionInput struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// GetSelectedRegion returns the current main-window rectangle in the
// coordinate space that the underlying capture library expects.
//
// This exists because Wails' own WindowGetPosition returns *screen-local*
// coordinates on macOS (relative to the display the window happens to be
// on), which breaks kbinani/screenshot.Capture on multi-display setups.
// The cgo helper in internal/appwindow performs the correct conversion
// via NSApp.mainWindow.frame + CGDisplayBounds(CGMainDisplayID()).
func (a *App) GetSelectedRegion() (CaptureRegionInput, error) {
	rect, err := appwindow.GetMainWindowRect()
	if err != nil {
		return CaptureRegionInput{}, err
	}
	region := CaptureRegionInput{
		X:      rect.Min.X,
		Y:      rect.Min.Y,
		Width:  rect.Dx(),
		Height: rect.Dy(),
	}
	return region, nil
}

// TestSessionParams bundles the frontend-supplied Capture Session inputs.
// Wails auto-generates a matching TypeScript class so the frontend can
// construct this object directly.
type TestSessionParams struct {
	RepeatCount         int                `json:"repeatCount"`
	StepIntervalSeconds float64            `json:"stepIntervalSeconds"`
	OutputDir           string             `json:"outputDir"`
	OutputFileName      string             `json:"outputFileName"`
	CaptureRegion       CaptureRegionInput `json:"captureRegion"`
}

// RunTestSession translates frontend inputs into a Capture Session Plan
// and delegates to the Runner. Advance Click Point defaults to the
// centre of the selected Capture Region until issue #06 supplies a
// user-picked value.
func (a *App) RunTestSession(params TestSessionParams) error {
	region := image.Rect(
		params.CaptureRegion.X,
		params.CaptureRegion.Y,
		params.CaptureRegion.X+params.CaptureRegion.Width,
		params.CaptureRegion.Y+params.CaptureRegion.Height,
	)
	center := image.Pt(
		region.Min.X+region.Dx()/2,
		region.Min.Y+region.Dy()/2,
	)

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	return a.runner.Run(ctx, capturerunner.Plan{
		RepeatCount:         params.RepeatCount,
		StepIntervalSeconds: params.StepIntervalSeconds,
		CaptureRegion:       region,
		AdvanceClickPoint:   center,
		OutputDir:           params.OutputDir,
		OutputFileName:      params.OutputFileName,
	})
}
