package session_test

import (
	"context"
	"image"
	"sync"
	"time"
)

type fakeScreener struct {
	mu    sync.Mutex
	calls int
	img   image.Image
	err   error
	log   *callLog
}

func (f *fakeScreener) Capture(region image.Rectangle) (image.Image, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.log != nil {
		f.log.record("Screener")
	}
	if f.err != nil {
		return nil, f.err
	}
	if f.img != nil {
		return f.img, nil
	}
	return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
}

type fakeClicker struct {
	mu    sync.Mutex
	calls int
	at    []image.Point
	err   error
	log   *callLog
}

func (f *fakeClicker) Click(p image.Point) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.at = append(f.at, p)
	if f.log != nil {
		f.log.record("Clicker")
	}
	return f.err
}

type fakeDocument struct {
	mu          sync.Mutex
	path        string
	appendCalls int
	closeCalls  int
	appendErr   error
	closeErr    error
	log         *callLog
}

func (f *fakeDocument) Path() string { return f.path }

func (f *fakeDocument) AppendPage(img image.Image) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.appendCalls++
	if f.log != nil {
		f.log.record("AppendPage")
	}
	return f.appendErr
}

func (f *fakeDocument) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closeCalls++
	if f.log != nil {
		f.log.record("Close")
	}
	return f.closeErr
}

type fakeClock struct {
	mu     sync.Mutex
	sleeps []time.Duration
	log    *callLog
	hook   func(ctx context.Context)
}

func (f *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	f.mu.Lock()
	f.sleeps = append(f.sleeps, d)
	if f.log != nil {
		f.log.record("Sleep")
	}
	hook := f.hook
	f.mu.Unlock()
	if hook != nil {
		hook(ctx)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

type callLog struct {
	mu  sync.Mutex
	seq []string
}

func (l *callLog) record(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seq = append(l.seq, name)
}

func (l *callLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, len(l.seq))
	copy(out, l.seq)
	return out
}
