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

type Clicker struct{}

func New() *Clicker { return &Clicker{} }

// Click moves the mouse to p (screen coordinates) and emits a left click.
// It fails rather than clicking at the wrong place if the cursor never
// arrives at p.
func (c *Clicker) Click(p image.Point) error {
	robotgo.Move(p.X, p.Y)
	if err := waitForCursor(p); err != nil {
		return err
	}
	return robotgo.Click("left", false)
}

func waitForCursor(want image.Point) error {
	deadline := time.Now().Add(moveSettleTimeout)
	var x, y int
	for {
		x, y = robotgo.Location()
		if abs(x-want.X) <= moveTolerance && abs(y-want.Y) <= moveTolerance {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("cursor did not reach (%d,%d), stalled at (%d,%d)", want.X, want.Y, x, y)
		}
		time.Sleep(movePollInterval)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
