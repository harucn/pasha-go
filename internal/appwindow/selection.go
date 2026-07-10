package appwindow

import "image"

// GetSelection reports the Capture Region and the Advance Click Point the user
// just chose, both in Screen Space.
//
// The Capture Region is the selection window's own rectangle. The Advance
// Click Point sits at offsetX/offsetY inside it, measured in CSS pixels from
// its top-left corner.
//
// The two are returned together because they are chosen together: no caller
// should ever hold one without the other.
func GetSelection(offsetX, offsetY float64) (image.Rectangle, image.Point, error) {
	region, err := GetMainWindowRect()
	if err != nil {
		return image.Rectangle{}, image.Point{}, err
	}
	return region, advanceClickPointAt(region, offsetX, offsetY), nil
}
