// Package session runs a Capture Session: a fixed number of Capture Steps
// that click the Advance Click Point, wait for the Step Interval, screenshot
// the Capture Region, and append the result to an Output Document.
//
// Everything between "the user pressed Start" and the finished Output
// Document lives behind one entry point, Runner.Run: input validation,
// Step Interval unit conversion, Output Document path collision resolution,
// PdfWriter construction, and the step loop itself.
//
// Clicking first means the page visible when the session starts is never
// captured: RepeatCount pages are collected starting one advance ahead.
//
// The four collaborators (Screener, Clicker, Clock, PdfWriter factory) are
// injected as interfaces, so the whole path is exercisable with fakes
// through Run — the only seam callers and tests need.
package session

import (
	"context"
	"errors"
	"fmt"
	"image"
	"math"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"pasha-go/internal/outputpath"
	"pasha-go/internal/pdfwriter"
)

// Origin sentinels tag a Capture Step failure by the collaborator that
// produced it, so callers can render a cause-specific message (issue #11).
// Errors are wrapped with %w, so use errors.Is to test them.
var (
	ErrCapture  = errors.New("screen capture failed")
	ErrPdfWrite = errors.New("pdf write failed")
	ErrClick    = errors.New("advance click failed")
)

type Screener interface {
	Capture(region image.Rectangle) (image.Image, error)
}

type Clicker interface {
	Click(p image.Point) error
}

type PdfWriter interface {
	AppendPage(img image.Image) error
	Close() error
}

type Clock interface {
	Sleep(ctx context.Context, d time.Duration) error
}

// Plan bundles all inputs needed to start a Capture Session. See CONTEXT.md
// "Capture Session Plan" for the domain definition: it is the snapshot taken
// the moment the user presses Start, and does not change during the session.
type Plan struct {
	RepeatCount         int
	StepIntervalSeconds float64
	CaptureRegion       image.Rectangle
	AdvanceClickPoint   image.Point
	OutputDir           string
	OutputFileName      string
}

func (p Plan) validate() error {
	if p.RepeatCount < 1 {
		return fmt.Errorf("repeat count must be >= 1, got %d", p.RepeatCount)
	}
	if math.IsNaN(p.StepIntervalSeconds) || math.IsInf(p.StepIntervalSeconds, 0) || p.StepIntervalSeconds <= 0 {
		return fmt.Errorf("step interval must be a positive finite number of seconds, got %v", p.StepIntervalSeconds)
	}
	if strings.TrimSpace(p.OutputDir) == "" {
		return fmt.Errorf("output directory must not be empty")
	}
	if strings.TrimSpace(p.OutputFileName) == "" {
		return fmt.Errorf("output file name must not be empty")
	}
	if p.CaptureRegion.Dx() <= 0 || p.CaptureRegion.Dy() <= 0 {
		return fmt.Errorf("capture region must be non-empty, got %v", p.CaptureRegion)
	}
	return nil
}

// PdfWriterFactory constructs a PdfWriter that writes to the given resolved
// output path.
type PdfWriterFactory func(path string) (PdfWriter, error)

// DefaultPdfWriterFactory adapts pdfwriter.New to the interface expected by
// the Runner. Wire this in NewApp / production callers.
func DefaultPdfWriterFactory(path string) (PdfWriter, error) {
	return pdfwriter.New(path)
}

// Runner runs Capture Sessions on demand. Collaborators are held once at
// construction; Plan values arrive per Run call.
type Runner struct {
	screener     Screener
	clicker      Clicker
	clock        Clock
	newPdfWriter PdfWriterFactory
}

// NewRunner wires a Runner with its collaborators. All four are required.
func NewRunner(scr Screener, clk Clicker, clock Clock, newPdf PdfWriterFactory) *Runner {
	return &Runner{
		screener:     scr,
		clicker:      clk,
		clock:        clock,
		newPdfWriter: newPdf,
	}
}

// Run validates the Plan, resolves the Output Document path, builds a
// PdfWriter, and runs RepeatCount Capture Steps sequentially. onProgress, if
// non-nil, is invoked after each Capture Step completes with (completedSteps,
// totalSteps). onStart, if non-nil, is invoked once with a stop function that
// cooperatively ends the session after the current Capture Step; callers hold
// onto it to implement a stop button.
//
// It returns the path of the Output Document that was written, together with
// the first error encountered. The path is the collision-resolved one, which
// may differ from the caller's requested file name — callers must render this
// value rather than re-assembling the path themselves.
//
// On error the path is empty. A Capture Session that fails before its first
// Capture Step never creates a file at all, so there is no path a caller could
// safely show.
func (r *Runner) Run(ctx context.Context, p Plan, onProgress func(current, total int), onStart func(stop func())) (string, error) {
	if err := p.validate(); err != nil {
		return "", err
	}

	desired := filepath.Join(p.OutputDir, p.OutputFileName+".pdf")
	outPath, err := outputpath.Resolve(desired)
	if err != nil {
		return "", err
	}

	pdf, err := r.newPdfWriter(outPath)
	if err != nil {
		return "", err
	}

	cs := &captureSession{
		cfg: config{
			CaptureRegion:     p.CaptureRegion,
			AdvanceClickPoint: p.AdvanceClickPoint,
			RepeatCount:       p.RepeatCount,
			StepInterval:      time.Duration(p.StepIntervalSeconds * float64(time.Second)),
			Screener:          r.screener,
			Clicker:           r.clicker,
			PdfWriter:         pdf,
			Clock:             r.clock,
			Progress:          onProgress,
		},
	}
	if onStart != nil {
		onStart(cs.stop)
	}
	if err := cs.start(ctx); err != nil {
		return "", err
	}
	return outPath, nil
}

// config is the step loop's own view of a Plan: units converted, collaborators
// resolved. Unexported — Run is the only way in.
type config struct {
	CaptureRegion     image.Rectangle
	AdvanceClickPoint image.Point
	RepeatCount       int
	StepInterval      time.Duration

	Screener  Screener
	Clicker   Clicker
	PdfWriter PdfWriter
	Clock     Clock

	// Progress, if non-nil, is called after each Capture Step completes with
	// the number of completed steps and the total (RepeatCount). It is not
	// called for a step that aborts partway (error or cancellation).
	Progress func(current, total int)
}

type captureSession struct {
	cfg     config
	stopped atomic.Bool
}

// start runs RepeatCount Capture Steps sequentially. It returns the first
// error encountered (from any collaborator or context cancellation) and
// always Closes the PdfWriter before returning.
func (s *captureSession) start(ctx context.Context) (err error) {
	defer func() {
		closeErr := s.cfg.PdfWriter.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("%w: %v", ErrPdfWrite, closeErr)
		}
	}()

	for i := 0; i < s.cfg.RepeatCount; i++ {
		if s.stopped.Load() {
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if err := s.cfg.Clicker.Click(s.cfg.AdvanceClickPoint); err != nil {
			return fmt.Errorf("%w: %v", ErrClick, err)
		}
		if err := s.cfg.Clock.Sleep(ctx, s.cfg.StepInterval); err != nil {
			return err
		}
		img, err := s.cfg.Screener.Capture(s.cfg.CaptureRegion)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCapture, err)
		}
		if err := s.cfg.PdfWriter.AppendPage(img); err != nil {
			return fmt.Errorf("%w: %v", ErrPdfWrite, err)
		}

		if s.cfg.Progress != nil {
			s.cfg.Progress(i+1, s.cfg.RepeatCount)
		}
	}
	return nil
}

// stop signals the session to finish after the current Capture Step.
func (s *captureSession) stop() {
	s.stopped.Store(true)
}
