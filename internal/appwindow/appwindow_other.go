//go:build !darwin

package appwindow

import (
	"errors"
	"image"
)

// GetMainWindowRect is a stub on non-darwin builds. Multi-display coord
// handling on Windows/Linux is a future concern (see issues #01-08 MVP).
func GetMainWindowRect() (image.Rectangle, error) {
	return image.Rectangle{}, errors.New("appwindow: not implemented on this platform")
}
