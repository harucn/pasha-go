package main

import (
	"image"
	"testing"
)

// Negative coordinates and values past the primary display's width are
// legitimate Screen Space values on multi-display setups (ADR-0003), so the
// conversions must carry them through untouched rather than clamp to zero.
func TestCaptureRegionInput_Rectangle(t *testing.T) {
	cases := []struct {
		name string
		in   CaptureRegionInput
		want image.Rectangle
	}{
		{
			name: "on the primary display",
			in:   CaptureRegionInput{X: 10, Y: 20, Width: 100, Height: 50},
			want: image.Rect(10, 20, 110, 70),
		},
		{
			name: "display above or left of primary: negative origin",
			in:   CaptureRegionInput{X: -800, Y: -200, Width: 300, Height: 100},
			want: image.Rect(-800, -200, -500, -100),
		},
		{
			name: "display right of primary: origin past primary width",
			in:   CaptureRegionInput{X: 2560, Y: 0, Width: 400, Height: 300},
			want: image.Rect(2560, 0, 2960, 300),
		},
		{
			name: "region straddling the primary's left edge",
			in:   CaptureRegionInput{X: -50, Y: 10, Width: 120, Height: 40},
			want: image.Rect(-50, 10, 70, 50),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.rectangle(); got != tc.want {
				t.Errorf("%+v.rectangle() = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestCaptureRegionOf(t *testing.T) {
	cases := []struct {
		name string
		in   image.Rectangle
		want CaptureRegionInput
	}{
		{
			name: "on the primary display",
			in:   image.Rect(10, 20, 110, 70),
			want: CaptureRegionInput{X: 10, Y: 20, Width: 100, Height: 50},
		},
		{
			name: "negative origin",
			in:   image.Rect(-800, -200, -500, -100),
			want: CaptureRegionInput{X: -800, Y: -200, Width: 300, Height: 100},
		},
		{
			name: "origin past primary width",
			in:   image.Rect(2560, 0, 2960, 300),
			want: CaptureRegionInput{X: 2560, Y: 0, Width: 400, Height: 300},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := captureRegionOf(tc.in); got != tc.want {
				t.Errorf("captureRegionOf(%v) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

// The Capture Region crosses the Wails seam twice — out via GetSelection, back
// in via RunCaptureSession — so the two conversions must be exact inverses.
func TestCaptureRegion_RoundTripsThroughTheWireShape(t *testing.T) {
	for _, want := range []image.Rectangle{
		image.Rect(0, 0, 1, 1),
		image.Rect(10, 20, 110, 70),
		image.Rect(-800, -200, -500, -100),
		image.Rect(2560, -1080, 2960, -780),
	} {
		if got := captureRegionOf(want).rectangle(); got != want {
			t.Errorf("round trip of %v = %v", want, got)
		}
	}
}

func TestClickPoint_RoundTripsThroughTheWireShape(t *testing.T) {
	for _, want := range []image.Point{
		image.Pt(0, 0),
		image.Pt(60, 45),
		image.Pt(-120, -30),
		image.Pt(3000, 1400),
	} {
		if got := clickPointOf(want).point(); got != want {
			t.Errorf("round trip of %v = %v", want, got)
		}
	}
}

func TestRegionSelectionOf_PairsRegionAndClickPoint(t *testing.T) {
	region := image.Rect(-50, 10, 70, 50)
	clickPoint := image.Pt(-20, 30)

	got := regionSelectionOf(region, clickPoint)

	if want := (CaptureRegionInput{X: -50, Y: 10, Width: 120, Height: 40}); got.Region != want {
		t.Errorf("Region = %+v, want %+v", got.Region, want)
	}
	if want := (ClickPointInput{X: -20, Y: 30}); got.ClickPoint != want {
		t.Errorf("ClickPoint = %+v, want %+v", got.ClickPoint, want)
	}
}
