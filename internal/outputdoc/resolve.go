package outputdoc

// Collision resolution for the Output Document's path.
//
// The Capture Session must never overwrite an existing PDF: if the user picks
// an output path that already exists, we append "-2", "-3", ... before the
// extension until we land on a free slot. This keeps recording
// non-interactive (no confirmation dialogs mid-session) and matches the PRD's
// "no overwrite" guarantee.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// resolve returns path if it does not exist. Otherwise it returns the
// first "<base>-N<ext>" variant (N=2,3,...) that does not exist, where
// <base> and <ext> are the extension-split parts of path.
func resolve(path string) (string, error) {
	if !exists(path) {
		return path, nil
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for n := 2; n < 1_000_000; n++ {
		candidate := fmt.Sprintf("%s-%d%s", base, n, ext)
		if !exists(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("outputdoc: exhausted numeric suffixes for %q", path)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true
}
