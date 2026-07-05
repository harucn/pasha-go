//go:build darwin

// Package appwindow returns the running Wails main window's rectangle in
// the coordinate space that kbinani/screenshot.Capture expects:
//
//	origin = upper-left corner of the primary display
//	unit   = logical points (not physical pixels)
//	axes   = x rightward, y downward
//
// This is deliberately platform-specific because macOS's NSScreen uses a
// bottom-left origin and Wails' own WindowGetPosition returns coordinates
// relative to the *current* screen (not the global desktop), which breaks
// on multi-display setups. See internal Wails source
// internal/frontend/desktop/darwin/Application.m: GetPosition.
package appwindow

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

typedef struct {
    int x;
    int y;
    int w;
    int h;
    int ok;
} WindowRect;

// GetMainWindowRect fetches [NSApp mainWindow].frame, converts it from
// NSScreen space (bottom-left global, points) to kbinani/screenshot's
// Windows-coordinate (top-left of primary display, points), and returns
// it. Must be safe to call from any goroutine, so dispatches to main.
static WindowRect GetMainWindowRect(void) {
    __block WindowRect r = {0, 0, 0, 0, 0};

    void (^work)(void) = ^{
        NSWindow* win = [NSApp mainWindow];
        if (!win) {
            // Fallback: some Wails startup states leave keyWindow == nil.
            // Pick the first visible window on screen.
            for (NSWindow* w in [NSApp windows]) {
                if ([w isVisible]) {
                    win = w;
                    break;
                }
            }
        }
        if (!win) return;

        NSRect frame = [win frame];
        CGRect primary = CGDisplayBounds(CGMainDisplayID());

        // Flip y from NSScreen (bottom-left, y-up) to kbinani (top-left, y-down).
        // The origin (0,0) of NSScreen space is the bottom-left of the primary
        // display, so primary.size.height gives us the flip pivot in points.
        r.x = (int)frame.origin.x;
        r.y = (int)(primary.size.height - (frame.origin.y + frame.size.height));
        r.w = (int)frame.size.width;
        r.h = (int)frame.size.height;
        r.ok = 1;
    };

    if ([NSThread isMainThread]) {
        work();
    } else {
        dispatch_sync(dispatch_get_main_queue(), work);
    }

    return r;
}
*/
import "C"

import (
	"errors"
	"image"
)

// GetMainWindowRect returns the current app's main window rectangle in the
// coordinate space that kbinani/screenshot.Capture expects. Returns an
// error only if no window is available yet (e.g., called before startup).
func GetMainWindowRect() (image.Rectangle, error) {
	r := C.GetMainWindowRect()
	if r.ok == 0 {
		return image.Rectangle{}, errors.New("appwindow: no main window available")
	}
	x, y, w, h := int(r.x), int(r.y), int(r.w), int(r.h)
	return image.Rect(x, y, x+w, y+h), nil
}
