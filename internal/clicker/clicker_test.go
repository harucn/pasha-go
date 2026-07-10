package clicker

import (
	"errors"
	"image"
	"strings"
	"testing"
	"time"
)

// fakeMouse reports a scripted sequence of cursor locations, one per poll.
// Once the script runs out it keeps reporting the last entry, standing in for
// a cursor that has settled (or stalled) wherever it is.
type fakeMouse struct {
	locations []image.Point
	polls     int

	moved    []image.Point
	clicks   int
	clickErr error
}

func (m *fakeMouse) Move(p image.Point) { m.moved = append(m.moved, p) }

func (m *fakeMouse) Location() image.Point {
	i := m.polls
	m.polls++
	if i >= len(m.locations) {
		i = len(m.locations) - 1
	}
	return m.locations[i]
}

func (m *fakeMouse) Click() error {
	m.clicks++
	return m.clickErr
}

// testClicker polls fast and gives up fast, so the settle loop runs in
// microseconds rather than the half-second the real one allows.
func testClicker(m *fakeMouse) *Clicker {
	return &Clicker{
		mouse:         m,
		settleTimeout: 20 * time.Millisecond,
		pollInterval:  time.Millisecond,
		tolerance:     moveTolerance,
	}
}

func TestClick_CursorAlreadyAtTarget_ClicksOnce(t *testing.T) {
	target := image.Pt(100, 200)
	m := &fakeMouse{locations: []image.Point{target}}

	if err := testClicker(m).Click(target); err != nil {
		t.Fatalf("Click: %v", err)
	}

	if len(m.moved) != 1 || m.moved[0] != target {
		t.Errorf("Move calls = %v, want exactly [%v]", m.moved, target)
	}
	if m.clicks != 1 {
		t.Errorf("Click calls = %d, want 1", m.clicks)
	}
}

func TestClick_WaitsForCursorToArriveBeforeClicking(t *testing.T) {
	target := image.Pt(100, 200)
	// The cursor lags at the old position for two polls, then lands.
	m := &fakeMouse{locations: []image.Point{
		image.Pt(0, 0),
		image.Pt(50, 100),
		target,
	}}

	if err := testClicker(m).Click(target); err != nil {
		t.Fatalf("Click: %v", err)
	}

	if m.polls < 3 {
		t.Errorf("polled %d times, want at least 3 (it clicked before the cursor landed)", m.polls)
	}
	if m.clicks != 1 {
		t.Errorf("Click calls = %d, want 1", m.clicks)
	}
}

// The whole point of the settle loop: a click that fires at the stale cursor
// position lands on the app's own always-on-top bar.
func TestClick_CursorNeverArrives_ErrsWithoutClicking(t *testing.T) {
	target := image.Pt(100, 200)
	stalled := image.Pt(7, 9)
	m := &fakeMouse{locations: []image.Point{stalled}}

	err := testClicker(m).Click(target)
	if err == nil {
		t.Fatal("Click: expected an error when the cursor never arrives, got nil")
	}
	if m.clicks != 0 {
		t.Errorf("Click calls = %d, want 0 — it clicked at the stale position", m.clicks)
	}
	for _, want := range []string{"100,200", "7,9"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Click() error = %q, want it to name %q", err, want)
		}
	}
}

// Scaled displays land the cursor a pixel off; that must not fail the session.
func TestClick_AcceptsCursorWithinTolerance(t *testing.T) {
	target := image.Pt(100, 200)
	for _, got := range []image.Point{
		{X: 101, Y: 200},
		{X: 99, Y: 200},
		{X: 100, Y: 201},
		{X: 101, Y: 201},
	} {
		m := &fakeMouse{locations: []image.Point{got}}
		if err := testClicker(m).Click(target); err != nil {
			t.Errorf("Click with cursor at %v: %v", got, err)
		}
		if m.clicks != 1 {
			t.Errorf("cursor at %v: Click calls = %d, want 1", got, m.clicks)
		}
	}
}

func TestClick_CursorJustOutsideTolerance_Errs(t *testing.T) {
	m := &fakeMouse{locations: []image.Point{{X: 102, Y: 200}}}

	if err := testClicker(m).Click(image.Pt(100, 200)); err == nil {
		t.Fatal("Click: expected an error 2px off target, got nil")
	}
	if m.clicks != 0 {
		t.Errorf("Click calls = %d, want 0", m.clicks)
	}
}

func TestClick_PropagatesClickFailure(t *testing.T) {
	target := image.Pt(100, 200)
	sentinel := errors.New("event tap refused")
	m := &fakeMouse{locations: []image.Point{target}, clickErr: sentinel}

	if err := testClicker(m).Click(target); !errors.Is(err, sentinel) {
		t.Errorf("Click() error = %v, want %v", err, sentinel)
	}
}
