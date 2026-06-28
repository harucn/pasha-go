// Package clock provides a context-aware Sleep used by CaptureSession.
package clock

import (
	"context"
	"time"
)

type RealClock struct{}

func New() RealClock { return RealClock{} }

func (RealClock) Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
