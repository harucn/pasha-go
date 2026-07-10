package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"pasha-go/internal/appwindow"
	"pasha-go/internal/clicker"
	"pasha-go/internal/clock"
	"pasha-go/internal/screener"
	"pasha-go/internal/session"
)

// sessionRunner is the seam through which App delegates the Capture Session.
// In production this is *session.Runner; tests can substitute a fake.
type sessionRunner interface {
	Run(ctx context.Context, p session.Plan, onProgress func(current, total int), onStart func(stop func())) (string, error)
}

// App is the Wails-bound adapter. It holds the runtime context and a
// pre-wired Runner; every method delegates its work behind the seam so
// the adapter itself stays thin.
type App struct {
	ctx    context.Context
	runner sessionRunner
	events sessionEvents

	// mu guards stop, the cooperative stop handle for the currently
	// running Capture Session. It is nil whenever no session is in flight.
	mu   sync.Mutex
	stop func()
}

// NewApp constructs the App together with a Runner wired to the real
// desktop collaborators.
func NewApp() *App {
	return newAppWithRunner(noopEvents{}, session.NewRunner(
		screener.New(),
		clicker.New(),
		clock.New(),
		session.DefaultPdfWriterFactory,
	))
}

// newAppWithRunner injects the Runner and the sessionEvents adapter. Kept
// package-private so tests can substitute fakes without exposing the seams
// publicly. startup swaps events for the live Wails adapter.
func newAppWithRunner(events sessionEvents, r sessionRunner) *App {
	return &App{runner: r, events: events}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.events = wailsEvents{ctx: ctx}

	// Best effort: the app works fine on AppKit's default material.
	_ = appwindow.SetTranslucencyMaterial(appwindow.MaterialUnderWindowBackground)
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

// CaptureSessionParams bundles the frontend-supplied Capture Session inputs.
// Wails auto-generates a matching TypeScript class so the frontend can
// construct this object directly.
type CaptureSessionParams struct {
	RepeatCount         int                `json:"repeatCount"`
	StepIntervalSeconds float64            `json:"stepIntervalSeconds"`
	OutputDir           string             `json:"outputDir"`
	OutputFileName      string             `json:"outputFileName"`
	CaptureRegion       CaptureRegionInput `json:"captureRegion"`
	AdvanceClickPoint   ClickPointInput    `json:"advanceClickPoint"`
}

// RunCaptureSession translates frontend inputs into a Capture Session Plan
// and delegates to the Runner. Both Capture Region and Advance Click
// Point are supplied by the user via #05 / #06 UI flows and arrive in
// Screen Space (primary top-left global points).
//
// It returns the path of the Output Document that was written. The Runner
// resolves name collisions by appending "-2", "-3", ..., so this may differ
// from params.OutputFileName; the frontend renders this value instead of
// re-assembling the path. On error the path is empty and Wails rejects the
// promise, so the frontend never sees it.
func (a *App) RunCaptureSession(params CaptureSessionParams) (string, error) {
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

	outPath, err := a.runner.Run(ctx, session.Plan{
		RepeatCount:         params.RepeatCount,
		StepIntervalSeconds: params.StepIntervalSeconds,
		CaptureRegion:       region,
		AdvanceClickPoint:   clickPoint,
		OutputDir:           params.OutputDir,
		OutputFileName:      params.OutputFileName,
	}, a.events.Progress, a.registerStop)

	a.clearStop()
	if err != nil {
		a.events.Failed(humanErrorMessage(err))
		return "", err
	}
	a.events.Completed()
	return outPath, nil
}

// humanErrorMessage translates a Capture Session error into a message the user
// can act on, keyed by the origin sentinel the session package wraps in. The
// underlying technical error is intentionally hidden from the user (#11).
func humanErrorMessage(err error) string {
	switch {
	case errors.Is(err, session.ErrCapture):
		return "Screen capture failed. Screen Recording permission may be disabled."
	case errors.Is(err, session.ErrPdfWrite):
		return "Could not write the PDF. Check the disk space and the permissions of the output folder."
	case errors.Is(err, session.ErrClick):
		return "Auto-click failed. Accessibility permission may be disabled."
	default:
		return "Something went wrong during the session. Please try again."
	}
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
