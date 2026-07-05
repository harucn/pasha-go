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
//
// The Cocoa helper only fetches raw values (NSWindow.frame in NSScreen
// space + primary display height). The coordinate-space conversion is a
// pure Go function in transform.go so it can be unit-tested without
// requiring AppKit or physical displays.
package appwindow

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

typedef struct {
    int nsX;
    int nsY;
    int nsW;
    int nsH;
    int primaryH;
    int ok;
} RawFrame;

// GetRawWindowFrame fetches [NSApp mainWindow].frame in NSScreen space
// (bottom-left global, points) together with the primary display height
// in points. All coordinate math (y-axis flip, primary-relative origin)
// happens in Go — see nsScreenToKbinani in transform.go.
static RawFrame GetRawWindowFrame(void) {
    __block RawFrame r = {0, 0, 0, 0, 0, 0};

    void (^work)(void) = ^{
        NSWindow* win = [NSApp mainWindow];
        if (!win) {
            // Some Wails startup states leave keyWindow == nil.
            // Fall back to the first visible window.
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

        r.nsX = (int)frame.origin.x;
        r.nsY = (int)frame.origin.y;
        r.nsW = (int)frame.size.width;
        r.nsH = (int)frame.size.height;
        r.primaryH = (int)primary.size.height;
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
	r := C.GetRawWindowFrame()
	if r.ok == 0 {
		return image.Rectangle{}, errors.New("appwindow: no main window available")
	}
	return nsScreenToKbinani(
		int(r.nsX), int(r.nsY), int(r.nsW), int(r.nsH),
		int(r.primaryH),
	), nil
}
