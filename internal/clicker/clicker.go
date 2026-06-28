// Package clicker implements the Clicker interface for macOS using robotgo
// to send a synthetic left mouse click at the given Advance Click Point.
package clicker

import (
	"image"

	"github.com/go-vgo/robotgo"
)

type Clicker struct{}

func New() *Clicker { return &Clicker{} }

// Click moves the mouse to p (screen coordinates) and emits a left click.
func (c *Clicker) Click(p image.Point) error {
	robotgo.Move(p.X, p.Y)
	if err := robotgo.Click("left", false); err != nil {
		return err
	}
	return nil
}
