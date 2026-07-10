// Package outputdoc owns the Output Document: the multi-page PDF a Capture
// Session appends to, one page per Capture Step.
//
// Create is the only entry point. Everything the Output Document guarantees
// lives behind it:
//
//   - the file is named "<fileName>.pdf" in dir — callers never spell the
//     extension;
//   - an existing file is never overwritten, so the document may land on
//     "<fileName>-2.pdf"; Path reports where it actually landed;
//   - every AppendPage flushes to disk, so a crash mid-session leaves a valid
//     partial PDF (ADR-0001).
//
// Callers must render Path rather than re-assembling a path of their own.
package outputdoc

import "path/filepath"

const ext = ".pdf"

// Create opens the Output Document for a Capture Session. It resolves a
// collision-free path first, so two sessions started against the same file
// name write to different files.
func Create(dir, fileName string) (*Document, error) {
	path, err := resolve(filepath.Join(dir, fileName+ext))
	if err != nil {
		return nil, err
	}
	return &Document{path: path}, nil
}

// Path is where the document is being written, which may carry a collision
// suffix the caller never asked for.
func (d *Document) Path() string { return d.path }
