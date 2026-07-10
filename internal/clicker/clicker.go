// Package clicker implements the Clicker interface for macOS using robotgo
// to send a synthetic left mouse click at the given Advance Click Point.
package clicker

import (
	"fmt"
	"image"
	"time"

	"github.com/go-vgo/robotgo"
)

// robotgo.Move posts a kCGEventMouseMoved through the HID event tap and
// returns immediately, so the cursor has not necessarily arrived when it
// does. robotgo.Click carries no coordinates — it fires wherever the cursor
// currently is. Clicking without waiting therefore hits the old position,
// which for this app is its own always-on-top bar.
const (
	moveSettleTimeout = 500 * time.Millisecond
	movePollInterval  = 10 * time.Millisecond
	// The cursor can land a pixel off on scaled displays.
	moveTolerance = 1
)

// mouse is the seam over robotgo's package-level functions. The settle loop
// below is the only non-trivial logic in this package and misbehaving there
// makes the app click its own bar, so it must be reachable from a test
// without a real cursor.
type mouse interface {
	Move(p image.Point)
	Location() image.Point
	Click() error
}

type robotgoMouse struct{}

func (robotgoMouse) Move(p image.Point) { robotgo.Move(p.X, p.Y) }

func (robotgoMouse) Location() image.Point {
	x, y := robotgo.Location()
	return image.Pt(x, y)
}

func (robotgoMouse) Click() error { return robotgo.Click("left", false) }

type Clicker struct {
	mouse         mouse
	settleTimeout time.Duration
	pollInterval  time.Duration
	tolerance     int
}

func New() *Clicker {
	return &Clicker{
		mouse:         robotgoMouse{},
		settleTimeout: moveSettleTimeout,
		pollInterval:  movePollInterval,
		tolerance:     moveTolerance,
	}
}

// Click moves the mouse to p (Screen Space) and emits a left click.
// It fails rather than clicking at the wrong place if the cursor never
// arrives at p.
func (c *Clicker) Click(p image.Point) error {
	c.mouse.Move(p)
	if err := c.waitForCursor(p); err != nil {
		return err
	}
	return c.mouse.Click()
}

func (c *Clicker) waitForCursor(want image.Point) error {
	deadline := time.Now().Add(c.settleTimeout)
	for {
		got := c.mouse.Location()
		if abs(got.X-want.X) <= c.tolerance && abs(got.Y-want.Y) <= c.tolerance {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("cursor did not reach (%d,%d), stalled at (%d,%d)", want.X, want.Y, got.X, got.Y)
		}
		time.Sleep(c.pollInterval)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
