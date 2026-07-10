package outputdoc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_ReturnsPathWhenNotExisting(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "pasha-2026-06-28_15-30.pdf")

	got, err := resolve(target)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != target {
		t.Errorf("Resolve(%q) = %q, want %q (unchanged)", target, got, target)
	}
}

func TestResolve_AppendsDash2WhenPathExists(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "pasha.pdf")
	mustCreate(t, target)

	got, err := resolve(target)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := filepath.Join(dir, "pasha-2.pdf")
	if got != want {
		t.Errorf("Resolve(%q) = %q, want %q", target, got, want)
	}
}

func TestResolve_AppendsDash3WhenDash2AlsoExists(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "pasha.pdf")
	mustCreate(t, target)
	mustCreate(t, filepath.Join(dir, "pasha-2.pdf"))

	got, err := resolve(target)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := filepath.Join(dir, "pasha-3.pdf")
	if got != want {
		t.Errorf("Resolve(%q) = %q, want %q", target, got, want)
	}
}

func TestResolve_PreservesExtensionCorrectly(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "no-ext")
	mustCreate(t, target)

	got, err := resolve(target)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := filepath.Join(dir, "no-ext-2")
	if got != want {
		t.Errorf("Resolve(%q) = %q, want %q (no extension case)", target, got, want)
	}
}

func mustCreate(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", path, err)
	}
}
