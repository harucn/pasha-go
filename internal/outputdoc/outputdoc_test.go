package outputdoc

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestCreate_AppendsPdfExtension(t *testing.T) {
	dir := t.TempDir()

	d, err := Create(dir, "pasha-2026-06-28_15-30")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	want := filepath.Join(dir, "pasha-2026-06-28_15-30.pdf")
	if d.Path() != want {
		t.Errorf("Path() = %q, want %q", d.Path(), want)
	}
}

// Collision resolution and incremental flushing meet here: the pages must land
// in the suffixed file, and the pre-existing document must be left alone. This
// is the interaction that no test could reach while the two lived in separate
// modules.
func TestCreate_ExistingDocumentIsNeverOverwritten(t *testing.T) {
	dir := t.TempDir()
	original := filepath.Join(dir, "pasha.pdf")
	if err := os.WriteFile(original, []byte("original contents"), 0o600); err != nil {
		t.Fatalf("seed existing document: %v", err)
	}

	d, err := Create(dir, "pasha")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if want := filepath.Join(dir, "pasha-2.pdf"); d.Path() != want {
		t.Fatalf("Path() = %q, want %q", d.Path(), want)
	}

	if err := d.AppendPage(newSolidImage(100, 50, color.RGBA{G: 255, A: 255})); err != nil {
		t.Fatalf("AppendPage: %v", err)
	}
	if err := d.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// The page went to the suffixed document...
	written, err := os.ReadFile(d.Path())
	if err != nil {
		t.Fatalf("read %s: %v", d.Path(), err)
	}
	if got := countPdfPages(written); got != 1 {
		t.Errorf("page count in %s = %d, want 1", d.Path(), got)
	}

	// ...and the original is untouched.
	untouched, err := os.ReadFile(original)
	if err != nil {
		t.Fatalf("read %s: %v", original, err)
	}
	if string(untouched) != "original contents" {
		t.Errorf("%s was overwritten: %q", original, untouched)
	}
}

// Create must claim its path before the first Capture Step, so two sessions
// started against the same file name never write to the same document.
func TestCreate_TwiceInARow_SecondCollidesOnlyAfterFirstFlushes(t *testing.T) {
	dir := t.TempDir()

	first, err := Create(dir, "pasha")
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	if err := first.AppendPage(newSolidImage(10, 10, color.RGBA{A: 255})); err != nil {
		t.Fatalf("AppendPage: %v", err)
	}

	second, err := Create(dir, "pasha")
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}

	if first.Path() == second.Path() {
		t.Fatalf("both documents resolved to %q", first.Path())
	}
	if want := filepath.Join(dir, "pasha-2.pdf"); second.Path() != want {
		t.Errorf("second Path() = %q, want %q", second.Path(), want)
	}
}

func TestClose_WithNoPages_WritesNothing(t *testing.T) {
	dir := t.TempDir()

	d, err := Create(dir, "empty")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := d.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := os.Stat(d.Path()); !os.IsNotExist(err) {
		t.Errorf("Close with no pages created %s, want no file", d.Path())
	}
}
