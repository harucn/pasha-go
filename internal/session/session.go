// Package session orchestrates a Capture Session: a fixed number of
// Capture Steps that screenshot a Capture Region, append the result to an
// Output Document, click the Advance Click Point, and wait for Step Interval.
//
// Collaborators (Screener, Clicker, PdfWriter, Clock) are injected as
// interfaces so this package can be exercised entirely with fakes.
package session

import (
	"context"
	"errors"
	"image"
	"sync/atomic"
	"time"
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

type Config struct {
	CaptureRegion     image.Rectangle
	AdvanceClickPoint image.Point
	RepeatCount       int
	StepInterval      time.Duration

	Screener  Screener
	Clicker   Clicker
	PdfWriter PdfWriter
	Clock     Clock
}

type CaptureSession struct {
	cfg     Config
	stopped atomic.Bool
}

func New(cfg Config) *CaptureSession {
	return &CaptureSession{cfg: cfg}
}

// Start runs RepeatCount Capture Steps sequentially. It returns the first
// error encountered (from any collaborator or context cancellation) and
// always Closes the PdfWriter before returning.
func (s *CaptureSession) Start(ctx context.Context) (err error) {
	defer func() {
		closeErr := s.cfg.PdfWriter.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for i := 0; i < s.cfg.RepeatCount; i++ {
		if s.stopped.Load() {
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		img, err := s.cfg.Screener.Capture(s.cfg.CaptureRegion)
		if err != nil {
			return err
		}
		if err := s.cfg.PdfWriter.AppendPage(img); err != nil {
			return err
		}
		if err := s.cfg.Clicker.Click(s.cfg.AdvanceClickPoint); err != nil {
			return err
		}
		if err := s.cfg.Clock.Sleep(ctx, s.cfg.StepInterval); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			return err
		}
	}
	return nil
}

// Stop signals the session to finish after the current Capture Step.
func (s *CaptureSession) Stop() {
	s.stopped.Store(true)
}
