package main

import "image"

// The wire types for Screen Space values, and the only place that converts
// between their origin+size shape and the image.Rectangle / image.Point shape
// the rest of the Go side uses.
//
// Both shapes are Screen Space (ADR-0003) — these conversions change the
// representation, not the coordinate space. Wails generates a TypeScript class
// per type here, so the frontend constructs them directly and never does
// arithmetic of its own.

// CaptureRegionInput carries the Capture Region as its top-left corner plus a
// size, which is what a drag-selection overlay naturally produces.
type CaptureRegionInput struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// rectangle is the min/max form. Width and Height are trusted as-is: a
// negative size would make image.Rect silently swap the corners, and the Plan
// rejects an empty Capture Region anyway.
func (r CaptureRegionInput) rectangle() image.Rectangle {
	return image.Rect(r.X, r.Y, r.X+r.Width, r.Y+r.Height)
}

func captureRegionOf(r image.Rectangle) CaptureRegionInput {
	return CaptureRegionInput{
		X:      r.Min.X,
		Y:      r.Min.Y,
		Width:  r.Dx(),
		Height: r.Dy(),
	}
}

// ClickPointInput carries the Advance Click Point.
type ClickPointInput struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p ClickPointInput) point() image.Point { return image.Pt(p.X, p.Y) }

func clickPointOf(p image.Point) ClickPointInput {
	return ClickPointInput{X: p.X, Y: p.Y}
}

// RegionSelection is what the user picked in the selection window: the Capture
// Region and, inside it, the Advance Click Point.
type RegionSelection struct {
	Region     CaptureRegionInput `json:"region"`
	ClickPoint ClickPointInput    `json:"clickPoint"`
}

// regionSelectionOf pairs the two, which are chosen together: no caller should
// ever hold one without the other.
func regionSelectionOf(region image.Rectangle, clickPoint image.Point) RegionSelection {
	return RegionSelection{
		Region:     captureRegionOf(region),
		ClickPoint: clickPointOf(clickPoint),
	}
}
