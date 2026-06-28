// Package screener implements the Screener interface for macOS using
// kbinani/screenshot. Captures the given Capture Region in screen
// coordinates and returns it as image.Image.
package screener

import (
	"fmt"
	"image"

	"github.com/kbinani/screenshot"
)

type Screener struct{}

func New() *Screener { return &Screener{} }

// Capture grabs the pixels inside region (screen coordinates) and returns
// them as an image.Image. region must be non-empty.
func (s *Screener) Capture(region image.Rectangle) (image.Image, error) {
	if region.Dx() <= 0 || region.Dy() <= 0 {
		return nil, fmt.Errorf("screener: empty capture region %v", region)
	}
	img, err := screenshot.CaptureRect(region)
	if err != nil {
		return nil, fmt.Errorf("screener: capture: %w", err)
	}
	return img, nil
}
