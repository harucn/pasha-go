package main

import (
	"context"
	"fmt"
	"image"
	"sync"
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
	Run(ctx context.Context, p capturerunner.Plan, onProgress func(current, total int), onStart func(stop func())) error
}

// App is the Wails-bound adapter. It holds the runtime context and a
// pre-wired Runner; every method delegates its work behind the seam so
// the adapter itself stays thin.
type App struct {
	ctx    context.Context
	runner sessionRunner

	// mu guards stop, the cooperative stop handle for the currently
	// running Capture Session. It is nil whenever no session is in flight.
	mu   sync.Mutex
	stop func()
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

// ClickPointInput carries the user-selected Advance Click Point in
// Screen Space (see docs/adr/0003-canonical-screen-coordinate-space.md).
type ClickPointInput struct {
	X int `json:"x"`
	Y int `json:"y"`
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
	AdvanceClickPoint   ClickPointInput    `json:"advanceClickPoint"`
}

// RunTestSession translates frontend inputs into a Capture Session Plan
// and delegates to the Runner. Both Capture Region and Advance Click
// Point are supplied by the user via #05 / #06 UI flows and arrive in
// Screen Space (primary top-left global points).
func (a *App) RunTestSession(params TestSessionParams) error {
	region := image.Rect(
		params.CaptureRegion.X,
		params.CaptureRegion.Y,
		params.CaptureRegion.X+params.CaptureRegion.Width,
		params.CaptureRegion.Y+params.CaptureRegion.Height,
	)
	clickPoint := image.Pt(params.AdvanceClickPoint.X, params.AdvanceClickPoint.Y)

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	err := a.runner.Run(ctx, capturerunner.Plan{
		RepeatCount:         params.RepeatCount,
		StepIntervalSeconds: params.StepIntervalSeconds,
		CaptureRegion:       region,
		AdvanceClickPoint:   clickPoint,
		OutputDir:           params.OutputDir,
		OutputFileName:      params.OutputFileName,
	}, a.emitProgress, a.registerStop)

	a.clearStop()
	if err == nil {
		a.emitCompleted()
	}
	return err
}

// registerStop stores the stop handle for the in-flight Capture Session so
// StopSession can end it later. Called by the Runner once the session exists.
func (a *App) registerStop(stop func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stop = stop
}

// clearStop drops the stop handle once the session has finished.
func (a *App) clearStop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stop = nil
}

// StopSession cooperatively stops the running Capture Session: the current
// Capture Step finishes, then the loop ends and the Output Document is saved
// (per issue #09 / PRD Q3). It is a no-op when no session is running.
func (a *App) StopSession() {
	a.mu.Lock()
	stop := a.stop
	a.mu.Unlock()
	if stop != nil {
		stop()
	}
}

// SessionProgress is the payload emitted on the "session:progress" event
// after each Capture Step completes. Wails generates a matching TypeScript
// type so the frontend can render "Current / Total".
type SessionProgress struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

// emitProgress forwards a Capture Session progress tick to the frontend via
// the Wails runtime. It is a no-op when the runtime context is absent (e.g.
// in unit tests where startup was never called).
func (a *App) emitProgress(current, total int) {
	if a.ctx == nil {
		return
	}
	wailsRuntime.EventsEmit(a.ctx, "session:progress", SessionProgress{
		Current: current,
		Total:   total,
	})
}

// emitCompleted notifies the frontend that the Capture Session has ended
// (whether it ran to completion or was stopped), so the UI can transition to
// the finished state. Same event channel shape as session:progress (#08).
// No-op when the runtime context is absent (e.g. in unit tests).
func (a *App) emitCompleted() {
	if a.ctx == nil {
		return
	}
	wailsRuntime.EventsEmit(a.ctx, "session:completed")
}
