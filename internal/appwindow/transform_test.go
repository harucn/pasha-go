package appwindow

import (
	"image"
	"testing"
)

func TestNsScreenToKbinani(t *testing.T) {
	// Baseline sanity: MacBook Pro 14" Retina, primary display 1512x982 pt.
	// (Physical pixels are 3024x1964, but NSScreen and CGDisplayBounds
	// both work in points.)
	const primaryH = 982

	cases := []struct {
		name    string
		nsX     int
		nsY     int
		nsW     int
		nsH     int
		wantMin image.Point
		wantMax image.Point
	}{
		{
			// A window flush against the top-left of the primary display.
			// NSScreen origin is bottom-left, so a window whose top edge
			// touches the top of the display sits at nsY = primaryH - h.
			name:    "primary top-left corner",
			nsX:     0,
			nsY:     primaryH - 400,
			nsW:     500,
			nsH:     400,
			wantMin: image.Pt(0, 0),
			wantMax: image.Pt(500, 400),
		},
		{
			// A window flush against the bottom-left of the primary display
			// (nsY == 0). In kbinani space, its top should be at
			// primaryH - h.
			name:    "primary bottom-left corner",
			nsX:     0,
			nsY:     0,
			nsW:     500,
			nsH:     400,
			wantMin: image.Pt(0, primaryH-400),
			wantMax: image.Pt(500, primaryH),
		},
		{
			// Secondary display arranged to the RIGHT of primary. Its
			// NSScreen origin.x is primaryW; the y coordinate lives at
			// the same vertical range as primary (0 to primaryH).
			name:    "secondary display to the right",
			nsX:     1512,
			nsY:     primaryH - 400,
			nsW:     500,
			nsH:     400,
			wantMin: image.Pt(1512, 0),
			wantMax: image.Pt(2012, 400),
		},
		{
			// Secondary display arranged ABOVE primary. Its NSScreen
			// origin.y is primaryH (above primary in NSScreen's y-up
			// space). A window sitting at the BOTTOM of that secondary
			// has nsY == primaryH. In kbinani's y-down space, this
			// should yield a NEGATIVE y — the whole secondary sits
			// "above 0" of primary top-left.
			name:    "secondary display above primary yields negative y",
			nsX:     0,
			nsY:     primaryH, // window bottom-left flush with bottom of secondary (which sits atop primary)
			nsW:     500,
			nsH:     400,
			wantMin: image.Pt(0, -400),
			wantMax: image.Pt(500, 0),
		},
		{
			// Secondary display arranged BELOW primary. Its NSScreen
			// origin.y is negative. In kbinani's y-down space, the y
			// coordinate is > primaryH.
			name:    "secondary display below primary yields y > primaryH",
			nsX:     0,
			nsY:     -400, // secondary top edge at nsY = 0 - secondaryH; window at bottom-left of secondary
			nsW:     500,
			nsH:     400,
			wantMin: image.Pt(0, primaryH),
			wantMax: image.Pt(500, primaryH+400),
		},
		{
			// Regression guard for the historical bug: if a caller ever
			// passes screen-local coords (as Wails' WindowGetPosition
			// used to return), the transform still runs but the caller
			// gets wrong output. This test locks the CORRECT semantic:
			// input MUST be global NSScreen space, and the output
			// matches kbinani space exactly.
			name:    "mid-primary window",
			nsX:     300,
			nsY:     200,
			nsW:     640,
			nsH:     320,
			wantMin: image.Pt(300, primaryH-200-320),
			wantMax: image.Pt(940, primaryH-200),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := nsScreenToKbinani(tc.nsX, tc.nsY, tc.nsW, tc.nsH, primaryH)
			want := image.Rectangle{Min: tc.wantMin, Max: tc.wantMax}
			if got != want {
				t.Errorf("nsScreenToKbinani(%d,%d,%d,%d,%d) = %v, want %v",
					tc.nsX, tc.nsY, tc.nsW, tc.nsH, primaryH, got, want)
			}
			if got.Dx() != tc.nsW || got.Dy() != tc.nsH {
				t.Errorf("size not preserved: got %dx%d, want %dx%d",
					got.Dx(), got.Dy(), tc.nsW, tc.nsH)
			}
		})
	}
}
