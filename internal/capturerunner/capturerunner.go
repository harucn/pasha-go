// Package capturerunner assembles and runs a Capture Session from a
// Capture Session Plan. It owns input validation, StepInterval unit
// conversion, Output Document path collision resolution, and PdfWriter
// construction — everything between "user pressed Start" and
// session.CaptureSession.Start(ctx).
//
// The four collaborators (Screener, Clicker, Clock, PdfWriter factory)
// are injected so the whole assembly path is exercisable through the
// Runner interface with fakes.
package capturerunner

import (
	"context"
	"fmt"
	"image"
	"math"
	"path/filepath"
	"strings"
	"time"

	"pasha-go/internal/outputpath"
	"pasha-go/internal/pdfwriter"
	"pasha-go/internal/session"
)

// Plan bundles all inputs needed to start a Capture Session. See
// CONTEXT.md "Capture Session Plan" for the domain definition.
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
	return nil
}

// PdfWriterFactory constructs a PdfWriter that writes to the given
// resolved output path.
type PdfWriterFactory func(path string) (session.PdfWriter, error)

// DefaultPdfWriterFactory adapts pdfwriter.New to the interface expected
// by the Runner. Wire this in NewApp / production callers.
func DefaultPdfWriterFactory(path string) (session.PdfWriter, error) {
	return pdfwriter.New(path)
}

// Runner assembles Capture Sessions on demand. Collaborators are held
// once at construction; Plan values arrive per Run call.
type Runner struct {
	screener     session.Screener
	clicker      session.Clicker
	clock        session.Clock
	newPdfWriter PdfWriterFactory
}

// New wires a Runner with its collaborators. All four are required.
func New(scr session.Screener, clk session.Clicker, clock session.Clock, newPdf PdfWriterFactory) *Runner {
	return &Runner{
		screener:     scr,
		clicker:      clk,
		clock:        clock,
		newPdfWriter: newPdf,
	}
}

// Run resolves the Output Document path, builds a PdfWriter, and starts
// a Capture Session. It returns the first error encountered.
func (r *Runner) Run(ctx context.Context, p Plan) error {
	if err := p.validate(); err != nil {
		return err
	}

	desired := filepath.Join(p.OutputDir, p.OutputFileName+".pdf")
	outPath, err := outputpath.Resolve(desired)
	if err != nil {
		return err
	}

	pdf, err := r.newPdfWriter(outPath)
	if err != nil {
		return err
	}

	cs := session.New(session.Config{
		CaptureRegion:     p.CaptureRegion,
		AdvanceClickPoint: p.AdvanceClickPoint,
		RepeatCount:       p.RepeatCount,
		StepInterval:      time.Duration(p.StepIntervalSeconds * float64(time.Second)),
		Screener:          r.screener,
		Clicker:           r.clicker,
		PdfWriter:         pdf,
		Clock:             r.clock,
	})
	return cs.Start(ctx)
}
