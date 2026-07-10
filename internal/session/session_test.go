package session_test

import (
	"context"
	"errors"
	"image"
	"testing"
	"time"

	"pasha-go/internal/session"
)

func newConfig(s session.Screener, c session.Clicker, w session.PdfWriter, clk session.Clock, repeat int) session.Config {
	return session.Config{
		CaptureRegion:     image.Rect(0, 0, 100, 100),
		AdvanceClickPoint: image.Pt(50, 50),
		RepeatCount:       repeat,
		StepInterval:      10 * time.Millisecond,
		Screener:          s,
		Clicker:           c,
		PdfWriter:         w,
		Clock:             clk,
	}
}

func TestCaptureSession_HappyPath_RunsRepeatCountSteps(t *testing.T) {
	scr := &fakeScreener{}
	clk := &fakeClicker{}
	pdf := &fakePdfWriter{}
	clock := &fakeClock{}

	cs := session.New(newConfig(scr, clk, pdf, clock, 5))

	if err := cs.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	if scr.calls != 5 {
		t.Errorf("Screener.Capture calls = %d, want 5", scr.calls)
	}
	if pdf.appendCalls != 5 {
		t.Errorf("PdfWriter.AppendPage calls = %d, want 5", pdf.appendCalls)
	}
	if clk.calls != 5 {
		t.Errorf("Clicker.Click calls = %d, want 5", clk.calls)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("PdfWriter.Close calls = %d, want 1", pdf.closeCalls)
	}
}

func TestCaptureSession_StepOrder(t *testing.T) {
	log := &callLog{}
	scr := &fakeScreener{log: log}
	clk := &fakeClicker{log: log}
	pdf := &fakePdfWriter{log: log}
	clock := &fakeClock{log: log}

	cs := session.New(newConfig(scr, clk, pdf, clock, 1))
	if err := cs.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	want := []string{"Clicker", "Sleep", "Screener", "AppendPage", "Close"}
	got := log.snapshot()
	if len(got) != len(want) {
		t.Fatalf("call log = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("call[%d] = %q, want %q (full %v)", i, got[i], want[i], got)
		}
	}
}

func TestCaptureSession_SleepsForStepInterval(t *testing.T) {
	clock := &fakeClock{}
	cs := session.New(newConfig(&fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, clock, 3))

	if err := cs.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if len(clock.sleeps) != 3 {
		t.Fatalf("Sleep calls = %d, want 3", len(clock.sleeps))
	}
	for i, d := range clock.sleeps {
		if d != 10*time.Millisecond {
			t.Errorf("Sleep[%d] = %v, want 10ms", i, d)
		}
	}
}

func TestCaptureSession_StopFromAnotherGoroutine(t *testing.T) {
	scr := &fakeScreener{}
	clk := &fakeClicker{}
	pdf := &fakePdfWriter{}
	clock := &fakeClock{}

	cs := session.New(newConfig(scr, clk, pdf, clock, 100))

	clock.hook = func(ctx context.Context) {
		cs.Stop()
	}

	if err := cs.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if scr.calls != 1 {
		t.Errorf("Screener.Capture = %d, want 1 (stop after first step)", scr.calls)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("Close calls = %d, want 1", pdf.closeCalls)
	}
}

func TestCaptureSession_ScreenerErrorAborts(t *testing.T) {
	scr := &fakeScreener{err: errSentinel("screener boom")}
	clk := &fakeClicker{}
	pdf := &fakePdfWriter{}
	clock := &fakeClock{}

	cs := session.New(newConfig(scr, clk, pdf, clock, 5))
	err := cs.Start(context.Background())
	if err == nil {
		t.Fatalf("Start: expected error")
	}
	if scr.calls != 1 || pdf.appendCalls != 0 || clk.calls != 1 {
		t.Errorf("expected abort after first Capture: scr=%d append=%d clk=%d",
			scr.calls, pdf.appendCalls, clk.calls)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("Close calls = %d, want 1", pdf.closeCalls)
	}
}

func TestCaptureSession_AppendPageErrorAborts(t *testing.T) {
	scr := &fakeScreener{}
	clk := &fakeClicker{}
	pdf := &fakePdfWriter{appendErr: errSentinel("pdf boom")}
	clock := &fakeClock{}

	cs := session.New(newConfig(scr, clk, pdf, clock, 5))
	err := cs.Start(context.Background())
	if err == nil {
		t.Fatalf("Start: expected error")
	}
	if scr.calls != 1 || pdf.appendCalls != 1 || clk.calls != 1 {
		t.Errorf("expected abort after AppendPage: scr=%d append=%d clk=%d",
			scr.calls, pdf.appendCalls, clk.calls)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("Close calls = %d, want 1", pdf.closeCalls)
	}
}

func TestCaptureSession_ClickerErrorAborts(t *testing.T) {
	scr := &fakeScreener{}
	clk := &fakeClicker{err: errSentinel("click boom")}
	pdf := &fakePdfWriter{}
	clock := &fakeClock{}

	cs := session.New(newConfig(scr, clk, pdf, clock, 5))
	err := cs.Start(context.Background())
	if err == nil {
		t.Fatalf("Start: expected error")
	}
	if scr.calls != 0 || pdf.appendCalls != 0 || clk.calls != 1 {
		t.Errorf("expected abort after Click: scr=%d append=%d clk=%d",
			scr.calls, pdf.appendCalls, clk.calls)
	}
	if pdf.closeCalls != 1 {
		t.Errorf("Close calls = %d, want 1", pdf.closeCalls)
	}
}

func TestCaptureSession_WrapsErrorsWithOriginSentinel(t *testing.T) {
	cases := []struct {
		name string
		cfg  func() session.Config
		want error
	}{
		{
			name: "capture",
			cfg: func() session.Config {
				return newConfig(&fakeScreener{err: errSentinel("boom")}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{}, 3)
			},
			want: session.ErrCapture,
		},
		{
			name: "append",
			cfg: func() session.Config {
				return newConfig(&fakeScreener{}, &fakeClicker{}, &fakePdfWriter{appendErr: errSentinel("boom")}, &fakeClock{}, 3)
			},
			want: session.ErrPdfWrite,
		},
		{
			name: "click",
			cfg: func() session.Config {
				return newConfig(&fakeScreener{}, &fakeClicker{err: errSentinel("boom")}, &fakePdfWriter{}, &fakeClock{}, 3)
			},
			want: session.ErrClick,
		},
		{
			name: "close",
			cfg: func() session.Config {
				return newConfig(&fakeScreener{}, &fakeClicker{}, &fakePdfWriter{closeErr: errSentinel("boom")}, &fakeClock{}, 1)
			},
			want: session.ErrPdfWrite,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cs := session.New(tc.cfg())
			err := cs.Start(context.Background())
			if err == nil {
				t.Fatalf("Start: expected error")
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("error %v is not %v", err, tc.want)
			}
		})
	}
}

func TestCaptureSession_ContextCancelAborts(t *testing.T) {
	scr := &fakeScreener{}
	clk := &fakeClicker{}
	pdf := &fakePdfWriter{}
	clock := &fakeClock{}

	ctx, cancel := context.WithCancel(context.Background())
	clock.hook = func(c context.Context) {
		cancel()
	}

	cs := session.New(newConfig(scr, clk, pdf, clock, 100))
	err := cs.Start(ctx)
	if err == nil {
		t.Fatalf("Start: expected ctx error")
	}
	if pdf.closeCalls != 1 {
		t.Errorf("Close calls = %d, want 1", pdf.closeCalls)
	}
	// Cancellation lands in the first Sleep, which now precedes the first
	// Capture, so nothing is ever captured.
	if scr.calls != 0 {
		t.Errorf("Screener.Capture = %d, want 0 (cancelled during the first Sleep)", scr.calls)
	}
	if clk.calls != 1 {
		t.Errorf("Clicker.Click = %d, want 1", clk.calls)
	}
}

func TestCaptureSession_ProgressCalledPerCompletedStep(t *testing.T) {
	var got [][2]int
	cfg := newConfig(&fakeScreener{}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{}, 3)
	cfg.Progress = func(current, total int) {
		got = append(got, [2]int{current, total})
	}

	cs := session.New(cfg)
	if err := cs.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	want := [][2]int{{1, 3}, {2, 3}, {3, 3}}
	if len(got) != len(want) {
		t.Fatalf("Progress calls = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Progress[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestCaptureSession_ProgressNotCalledOnAbortedStep(t *testing.T) {
	var calls int
	cfg := newConfig(&fakeScreener{err: errSentinel("boom")}, &fakeClicker{}, &fakePdfWriter{}, &fakeClock{}, 3)
	cfg.Progress = func(int, int) { calls++ }

	cs := session.New(cfg)
	if err := cs.Start(context.Background()); err == nil {
		t.Fatal("Start: expected error")
	}
	if calls != 0 {
		t.Errorf("Progress calls = %d, want 0 (step never completed)", calls)
	}
}

type errSentinel string

func (e errSentinel) Error() string { return string(e) }
