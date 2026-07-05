package appwindow

import "image"

// nsScreenToKbinani converts a window rectangle from macOS NSScreen
// coordinate space (origin at the bottom-left of the primary display,
// y-axis upward, unit: logical points) to the coordinate space that
// kbinani/screenshot.Capture expects (origin at the top-left of the
// primary display, y-axis downward, unit: logical points).
//
// primaryH is the height of the primary display in points, used as
// the flip pivot for the y-axis. On multi-display setups this correctly
// yields:
//
//   - primary display window: 0 <= y <= primaryH - h
//   - secondary display to the right of primary: x >= primaryW
//   - secondary display above primary (in Displays arrangement): y < 0
//   - secondary display below primary: y > primaryH
//
// The function is a pure numerical conversion — it does not touch AppKit
// or CoreGraphics — so it is unit-testable on any platform.
func nsScreenToKbinani(nsX, nsY, nsW, nsH, primaryH int) image.Rectangle {
	x := nsX
	y := primaryH - (nsY + nsH)
	return image.Rect(x, y, x+nsW, y+nsH)
}
