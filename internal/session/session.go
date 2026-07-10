// Package session runs a Capture Session. Runner.Run is the only entry point;
// the step loop behind it is an implementation detail, so the four injected
// collaborators are the only thing tests need to substitute.
//
// Clicking precedes capturing, so the page visible when the session starts is
// never captured: RepeatCount pages are collected starting one advance ahead.
package session

import (
	"context"
	"errors"
	"fmt"
	"image"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"pasha-go/internal/outputpath"
	"pasha-go/internal/pdfwriter"
)

// Origin sentinels let callers render a cause-specific message (issue #11).
// Wrapped with %w, so test them with errors.Is.
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

// Plan is the Capture Session Plan of CONTEXT.md: the snapshot taken when the
// user presses Start, unchanged for the rest of the session.
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

type PdfWriterFactory func(path string) (PdfWriter, error)

func DefaultPdfWriterFactory(path string) (PdfWriter, error) {
	return pdfwriter.New(path)
}

// Runner runs one Capture Session at a time — the constraint the UI already
// enforces by hiding Start while a session is in flight. Concurrent Runs share
// one Stop: the most recent Run wins, and Stop ends only that one.
type Runner struct {
	screener     Screener
	clicker      Clicker
	clock        Clock
	newPdfWriter PdfWriterFactory

	// mu guards current, which is nil whenever no session is in flight.
	mu      sync.Mutex
	current *captureSession
}

// NewRunner takes all four collaborators; none may be nil.
func NewRunner(scr Screener, clk Clicker, clock Clock, newPdf PdfWriterFactory) *Runner {
	return &Runner{
		screener:     scr,
		clicker:      clk,
		clock:        clock,
		newPdfWriter: newPdf,
	}
}

// Run blocks until the session ends. onProgress, if non-nil, fires after each
// completed Capture Step.
//
// The returned path is collision-resolved and may differ from the requested
// file name, so callers must render it rather than re-assemble their own.
//
// On error the path is empty: a session that fails before its first Capture
// Step writes no file, so there is nothing a caller could safely show.
func (r *Runner) Run(ctx context.Context, p Plan, onProgress func(current, total int)) (string, error) {
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

	// Releasing the finished session matters for memory, not for Stop: the
	// session holds the PdfWriter, which retains every captured page.
	r.setCurrent(cs)
	defer r.setCurrent(nil)

	if err := cs.start(ctx); err != nil {
		return "", err
	}
	return outPath, nil
}

// Stop ends the running session after its current Capture Step; Run then
// returns normally with the path it wrote.
//
// No-op when nothing is in flight, including the window after Run is called
// but before the session exists. No Capture Step has run by then.
func (r *Runner) Stop() {
	r.mu.Lock()
	cs := r.current
	r.mu.Unlock()
	if cs != nil {
		cs.stop()
	}
}

func (r *Runner) setCurrent(cs *captureSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.current = cs
}

// config is a Plan with units converted and collaborators resolved.
type config struct {
	CaptureRegion     image.Rectangle
	AdvanceClickPoint image.Point
	RepeatCount       int
	StepInterval      time.Duration

	Screener  Screener
	Clicker   Clicker
	PdfWriter PdfWriter
	Clock     Clock

	// Progress is not called for a step that aborts partway.
	Progress func(current, total int)
}

type captureSession struct {
	cfg     config
	stopped atomic.Bool
}

// start always Closes the PdfWriter, so a failed session still leaves a valid
// partial Output Document (ADR-0001).
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

func (s *captureSession) stop() {
	s.stopped.Store(true)
}
