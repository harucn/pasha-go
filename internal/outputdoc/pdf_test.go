package outputdoc

import (
	"image"
	"image/color"
	"math"
	"os"
	"regexp"
	"strconv"
	"testing"
)

func newSolidImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestDocument_PageMatchesImageAspectRatio(t *testing.T) {
	dir := t.TempDir()
	w, err := Create(dir, "out")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	path := w.Path()
	if err := w.AppendPage(newSolidImage(200, 100, color.RGBA{R: 255, A: 255})); err != nil {
		t.Fatalf("AppendPage: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// The first /MediaBox is the document default on the Pages node; the
	// per-page override we care about comes after it.
	re := regexp.MustCompile(`/MediaBox \[ 0 0 ([0-9.]+) ([0-9.]+) \]`)
	all := re.FindAllSubmatch(raw, -1)
	if len(all) < 2 {
		t.Fatalf("no per-page /MediaBox found in PDF (got %d)", len(all))
	}
	m := all[len(all)-1]
	pw, _ := strconv.ParseFloat(string(m[1]), 64)
	ph, _ := strconv.ParseFloat(string(m[2]), 64)

	if got, want := pw/ph, 2.0; math.Abs(got-want) > 0.01 {
		t.Errorf("page aspect ratio = %v (%vx%v), want %v", got, pw, ph, want)
	}
}

func TestDocument_AppendPagesAndClose_ProducesPdf(t *testing.T) {
	dir := t.TempDir()
	w, err := Create(dir, "out")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	path := w.Path()

	for i := 0; i < 3; i++ {
		if err := w.AppendPage(newSolidImage(200, 100, color.RGBA{R: 255, A: 255})); err != nil {
			t.Fatalf("AppendPage[%d]: %v", i, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("pdf is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data[:4]) != "%PDF" {
		t.Errorf("file does not start with %%PDF header: %q", string(data[:8]))
	}
	if got := countPdfPages(data); got != 3 {
		t.Errorf("page count = %d, want 3", got)
	}
}

// countPdfPages returns the number of `/Type /Page` (not /Pages) objects
// in a PDF byte stream. Sufficient for a smoke check.
func countPdfPages(data []byte) int {
	count := 0
	needle := []byte("/Type /Page")
	for i := 0; i+len(needle) <= len(data); i++ {
		if string(data[i:i+len(needle)]) != string(needle) {
			continue
		}
		// Skip "/Type /Pages"
		if i+len(needle) < len(data) && data[i+len(needle)] == 's' {
			continue
		}
		count++
	}
	return count
}

func TestDocument_PartialPdfExistsBeforeClose(t *testing.T) {
	dir := t.TempDir()
	w, err := Create(dir, "partial")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	path := w.Path()
	t.Cleanup(func() { _ = w.Close() })

	if err := w.AppendPage(newSolidImage(100, 100, color.RGBA{B: 255, A: 255})); err != nil {
		t.Fatalf("AppendPage: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("partial pdf not on disk after AppendPage: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("partial pdf empty")
	}
}
