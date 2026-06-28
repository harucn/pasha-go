package pdfwriter_test

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"pasha-go/internal/pdfwriter"
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

func TestWriter_AppendPagesAndClose_ProducesPdf(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")

	w, err := pdfwriter.New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

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

func TestWriter_PartialPdfExistsBeforeClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.pdf")

	w, err := pdfwriter.New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
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
