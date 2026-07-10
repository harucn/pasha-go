package appwindow

import (
	"image"
	"testing"
)

func TestAdvanceClickPointAt(t *testing.T) {
	cases := []struct {
		name    string
		region  image.Rectangle
		offsetX float64
		offsetY float64
		want    image.Point
	}{
		{
			name:   "marker at the frame's top-left is the region origin",
			region: image.Rect(100, 200, 600, 600),
			want:   image.Pt(100, 200),
		},
		{
			name:    "offset is added to the region origin",
			region:  image.Rect(100, 200, 600, 600),
			offsetX: 250, offsetY: 200,
			want: image.Pt(350, 400),
		},
		{
			// The bar can sit on a display left of or above the primary one.
			name:    "negative region origin (secondary display)",
			region:  image.Rect(-500, -300, 0, 100),
			offsetX: 250, offsetY: 200,
			want: image.Pt(-250, -100),
		},
		{
			name:    "fractional offset rounds half away from zero",
			region:  image.Rect(0, 0, 500, 400),
			offsetX: 250.5, offsetY: 199.5,
			want: image.Pt(251, 200),
		},
		{
			name:    "fractional offset rounds down below .5",
			region:  image.Rect(0, 0, 500, 400),
			offsetX: 250.49, offsetY: 199.4,
			want: image.Pt(250, 199),
		},
		{
			// A pointer drag can push the marker past the frame's edge; that is
			// the user's business, not ours. Screen Space allows any value.
			name:    "offset outside the region is not clamped",
			region:  image.Rect(100, 100, 600, 500),
			offsetX: -20, offsetY: 900,
			want: image.Pt(80, 1000),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := advanceClickPointAt(tc.region, tc.offsetX, tc.offsetY)
			if got != tc.want {
				t.Errorf("advanceClickPointAt(%v, %v, %v) = %v, want %v",
					tc.region, tc.offsetX, tc.offsetY, got, tc.want)
			}
		})
	}
}
