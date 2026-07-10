package clock_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"pasha-go/internal/clock"
)

func TestSleep_WaitsForTheFullDuration(t *testing.T) {
	const d = 50 * time.Millisecond

	start := time.Now()
	err := clock.New().Sleep(context.Background(), d)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Sleep() = %v, want nil", err)
	}
	if elapsed < d {
		t.Errorf("Sleep(%v) returned after %v, want at least %v", d, elapsed, d)
	}
}

func TestSleep_ReturnsEarlyWhenContextIsCancelledMidSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const d = 10 * time.Second
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := clock.New().Sleep(ctx, d)
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Sleep() = %v, want context.Canceled", err)
	}
	if elapsed >= d {
		t.Errorf("Sleep() blocked for %v; it should abandon the timer on cancel", elapsed)
	}
}

func TestSleep_AlreadyCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// A cancelled context wins over a live timer: the Capture Session must
	// not keep waiting out a Step Interval after it has been cancelled.
	if err := clock.New().Sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("Sleep(cancelled, 1h) = %v, want context.Canceled", err)
	}
}

// A non-positive Step Interval is not an error on its own: Sleep reports
// whatever the context says, so a live context yields nil and a cancelled
// one yields its error.
func TestSleep_NonPositiveDurationReportsContextState(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name string
		ctx  context.Context
		d    time.Duration
		want error
	}{
		{"zero duration, live context", context.Background(), 0, nil},
		{"negative duration, live context", context.Background(), -time.Second, nil},
		{"zero duration, cancelled context", cancelled, 0, context.Canceled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clock.New().Sleep(tt.ctx, tt.d)
			if !errors.Is(err, tt.want) {
				t.Errorf("Sleep(_, %v) = %v, want %v", tt.d, err, tt.want)
			}
		})
	}
}

func TestSleep_DeadlineExceededIsReported(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := clock.New().Sleep(ctx, time.Hour); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Sleep() = %v, want context.DeadlineExceeded", err)
	}
}
