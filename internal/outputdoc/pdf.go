package outputdoc

// The PDF encoding half of the Output Document. Implements ADR-0001: each
// AppendPage flushes the current document to disk so a crash leaves a valid
// partial PDF.
//
// gopdf.WritePdf mutates internal state (it appends page IDs to the PagesObj
// on every call), so it cannot be invoked multiple times on the same GoPdf
// instance. We instead retain the encoded page bytes and rebuild a fresh
// GoPdf from scratch on every flush.

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/signintech/gopdf"
)

// Document is an open Output Document. Obtain one from Create.
type Document struct {
	path  string
	pages []pageData
}

type pageData struct {
	pngBytes []byte
	width    int
	height   int
}

// AppendPage encodes the image as PNG, retains it, and flushes the
// document so a crash leaves a valid partial PDF.
func (d *Document) AppendPage(img image.Image) error {
	if img == nil {
		return fmt.Errorf("outputdoc: nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return fmt.Errorf("outputdoc: empty image bounds")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("outputdoc: encode png: %w", err)
	}

	d.pages = append(d.pages, pageData{
		pngBytes: buf.Bytes(),
		width:    bounds.Dx(),
		height:   bounds.Dy(),
	})

	return d.flush()
}

// Close performs one final flush so the file reflects all accumulated
// pages. Safe to call even if AppendPage was never invoked (no-op).
func (d *Document) Close() error {
	if len(d.pages) == 0 {
		return nil
	}
	return d.flush()
}

func (d *Document) flush() error {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	pageW := gopdf.PageSizeA4.W
	pageH := gopdf.PageSizeA4.H

	for i, p := range d.pages {
		imgW := float64(p.width)
		imgH := float64(p.height)
		scale := pageW / imgW
		if pageH/imgH < scale {
			scale = pageH / imgH
		}
		drawW := imgW * scale
		drawH := imgH * scale

		pdf.AddPageWithOption(gopdf.PageOption{PageSize: &gopdf.Rect{W: drawW, H: drawH}})

		holder, err := gopdf.ImageHolderByBytes(p.pngBytes)
		if err != nil {
			return fmt.Errorf("outputdoc: image holder page %d: %w", i+1, err)
		}
		if err := pdf.ImageByHolder(holder, 0, 0, &gopdf.Rect{W: drawW, H: drawH}); err != nil {
			return fmt.Errorf("outputdoc: draw image page %d: %w", i+1, err)
		}
	}

	if err := pdf.WritePdf(d.path); err != nil {
		return fmt.Errorf("outputdoc: write: %w", err)
	}
	return nil
}
