package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"pasha-go/internal/appwindow"
	"pasha-go/internal/clicker"
	"pasha-go/internal/clock"
	"pasha-go/internal/screener"
	"pasha-go/internal/session"
)

// sessionRunner is the seam through which App delegates the Capture Session.
type sessionRunner interface {
	Run(ctx context.Context, p session.Plan, onProgress func(current, total int)) (string, error)
	Stop()
}

// App is the Wails-bound adapter. Every method delegates behind a seam, so the
// adapter itself holds no state beyond its collaborators.
type App struct {
	ctx    context.Context
	runner sessionRunner
	events sessionEvents
}

// NewApp wires the Runner to the real desktop collaborators.
func NewApp() *App {
	return newAppWithRunner(noopEvents{}, session.NewRunner(
		screener.New(),
		clicker.New(),
		clock.New(),
		session.DefaultPdfWriterFactory,
	))
}

// newAppWithRunner is package-private so tests can substitute fakes without
// exposing the seams. startup swaps events for the live Wails adapter.
func newAppWithRunner(events sessionEvents, r sessionRunner) *App {
	return &App{runner: r, events: events}
}

// startup runs once the Wails runtime exists.
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

// RegionSelection is what the user picked in the selection window: the
// Capture Region and, inside it, the Advance Click Point. Both in Screen Space.
type RegionSelection struct {
	Region     CaptureRegionInput `json:"region"`
	ClickPoint ClickPointInput    `json:"clickPoint"`
}

// GetSelection converts the marker's offset inside the selection window, in
// CSS pixels from its top-left corner, into the Screen Space selection.
//
// Screen Space never crosses into the frontend as arithmetic (ADR-0003):
// Wails' own WindowGetPosition returns screen-local coordinates on macOS, so
// deriving any of this in JS breaks on multi-display setups.
func (a *App) GetSelection(offsetX, offsetY float64) (RegionSelection, error) {
	rect, point, err := appwindow.GetSelection(offsetX, offsetY)
	if err != nil {
		return RegionSelection{}, err
	}
	return RegionSelection{
		Region: CaptureRegionInput{
			X:      rect.Min.X,
			Y:      rect.Min.Y,
			Width:  rect.Dx(),
			Height: rect.Dy(),
		},
		ClickPoint: ClickPointInput{X: point.X, Y: point.Y},
	}, nil
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

// RunCaptureSession translates frontend inputs into a Plan. Capture Region and
// Advance Click Point arrive in Screen Space (ADR-0003).
//
// The returned path is collision-resolved and may differ from
// params.OutputFileName, so the frontend renders it instead of re-assembling
// its own. On error it is empty and Wails rejects the promise anyway.
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
	}, a.events.Progress)

	if err != nil {
		a.events.Failed(humanErrorMessage(err))
		return "", err
	}
	a.events.Completed()
	return outPath, nil
}

// humanErrorMessage keys off the origin sentinel. The underlying technical
// error is intentionally hidden from the user (#11).
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

// StopSession lets the current Capture Step finish, then saves the Output
// Document (issue #09 / PRD Q3). No-op when no session is running.
func (a *App) StopSession() {
	a.runner.Stop()
}
