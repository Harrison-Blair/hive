package tui

import (
	"strings"
	"testing"
	"time"
)

// AC1 and AC3: the coordinator marker advances through the Pi frames at each
// marker tick, and its duration comes from status_changed_at only while its
// state is working.
func TestCoordinatorWorkingMarkerAnimates(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	m.header.coordinator = "working"
	m.header.coordinatorChangedAt = gfDay.Add(-2 * time.Minute).Format(time.RFC3339)
	m.invalidateView()
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ● working 00:02:00 · ") {
		t.Fatalf("still status row = %q, want the rest glyph and its age", got)
	}
	for frame, glyph := range []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏", "⠋"} {
		m, cmd := raise(t, m, markerTickMsg{})
		if cmd == nil {
			t.Fatalf("frame %d scheduled no next tick", frame+1)
		}
		want := " " + glyph + " working 00:02:00 · "
		if got := viewRow(m, 39); !strings.HasPrefix(got, want) {
			t.Fatalf("status row at frame %d = %q, want prefix %q", frame+1, got, want)
		}
	}
}

// AC3: the working duration counts from status_changed_at and the display
// follows the TUI clock. An idle state shows no duration.
func TestCoordinatorWorkingDurationUsesStatusChange(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	m.header.coordinator = "working"
	m.header.coordinatorChangedAt = gfDay.Add(-2 * time.Minute).Format(time.RFC3339)
	m.invalidateView()
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ● working 00:02:00 · ") {
		t.Fatalf("status row = %q, want a 00:02:00 working duration", got)
	}
	m.now = func() time.Time { return gfDay.Add(3 * time.Minute) }
	m.invalidateView()
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ● working 00:05:00 · ") {
		t.Fatalf("status row after 3m = %q, want a 5m working duration", got)
	}
	m.header.coordinator = "idle"
	m.invalidateView()
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ idle · ") {
		t.Fatalf("idle status row = %q, want no duration", got)
	}
}

// AC2 and AC3: an idle or blocked coordinator has no duration, and its
// marker stays still. The timer stops when no working marker remains.
func TestNonWorkingCoordinatorStaysStillAndStopsTheTimer(t *testing.T) {
	for _, tc := range []struct {
		state string
		want  string
	}{
		{"idle", " ○ idle · "},
		{"blocked", " ▲ blocked · "},
		{"starting", " ● starting · "},
		{"", " ○ none · "},
	} {
		t.Run(tc.state, func(t *testing.T) {
			m, _ := newFrameModel(t, 120, 40)
			m.header.coordinator = tc.state
			m.header.coordinatorChangedAt = gfDay.Add(-5 * time.Minute).Format(time.RFC3339)
			m.invalidateView()
			if got := viewRow(m, 39); !strings.HasPrefix(got, tc.want) {
				t.Fatalf("status row = %q, want prefix %q", got, tc.want)
			}
			if m.workingMarkerNeeded() {
				t.Fatal("a non-working coordinator asked for marker ticks")
			}
			m, cmd := raise(t, m, markerTickMsg{})
			if cmd != nil || m.markerFrame != 0 {
				t.Fatalf("a marker tick advanced a still screen: frame %d, next %v", m.markerFrame, cmd != nil)
			}
			if got := viewRow(m, 39); !strings.HasPrefix(got, tc.want) {
				t.Fatalf("status row after a tick = %q, want prefix %q", got, tc.want)
			}
		})
	}
}

// AC1: a working marker starts its own timer on the first tick and stops it
// on the first tick after the marker becomes idle.
func TestWorkingMarkerTimerStartsAndStops(t *testing.T) {
	m, _ := newFrameModel(t, 120, 40)
	m.header.coordinator = "working"
	m.invalidateView()
	m = startMarkerTicks(t, m)
	if m.markerFrame != 0 {
		t.Fatalf("frame = %d before a tick, want 0", m.markerFrame)
	}
	m, cmd := raise(t, m, markerTickMsg{})
	if cmd == nil || m.markerFrame != 1 {
		t.Fatalf("first tick: frame %d, next %v; want frame 1 and a next tick", m.markerFrame, cmd != nil)
	}
	m.header.coordinator = "idle"
	m, cmd = raise(t, m, markerTickMsg{})
	if cmd != nil || m.markerTickRunning || m.markerFrame != 1 {
		t.Fatalf("stop tick: frame %d, running %v, next %v; want a still frame and no timer", m.markerFrame, m.markerTickRunning, cmd != nil)
	}
}

// AC1: every agent working marker uses the same frames, and an idle marker
// stays still.
func TestAgentWorkingMarkerAnimates(t *testing.T) {
	m := newGfDemoModel(t, "GfAgents")
	m.header.coordinator = "idle"
	m.header.coordinatorChangedAt = gfDay.Add(-5 * time.Minute).Format(time.RFC3339)
	m.invalidateView()
	if !m.workingMarkerNeeded() {
		t.Fatal("the working agents asked for no marker ticks")
	}
	requireView(t, m, "● t5-worker-1", "● manager-hive", "◐ t3-verifier-1")
	m, cmd := raise(t, m, markerTickMsg{})
	if cmd == nil {
		t.Fatal("the working agent scheduled no next tick")
	}
	requireView(t, m, "⠋ t5-worker-1", "⠋ manager-hive", "◐ t3-verifier-1")
}

// AC1: a working agent row on Health uses the same frames, and an idle row
// stays still.
func TestHealthAgentWorkingMarkerAnimates(t *testing.T) {
	m := newGfDemoModel(t, "GfHealth")
	m.header.coordinator = "idle"
	m.header.coordinatorChangedAt = gfDay.Add(-5 * time.Minute).Format(time.RFC3339)
	m.invalidateView()
	if !m.workingMarkerNeeded() {
		t.Fatal("the working Health agents asked for no marker ticks")
	}
	requireView(t, m, "● coordinator", "● manager-hive", "○ t3-verifier-1")
	m, cmd := raise(t, m, markerTickMsg{})
	if cmd == nil {
		t.Fatal("the working Health agent scheduled no next tick")
	}
	requireView(t, m, "⠋ coordinator", "⠋ manager-hive", "○ t3-verifier-1")
}

// AC1 and AC2: task working markers animate, and each other task marker
// stays still. A board with no working marker starts no timer.
func TestTaskWorkingMarkerAnimatesAndStops(t *testing.T) {
	m := newGfDemoModel(t, "GfWork")
	m.header.coordinator = "idle"
	m.header.coordinatorChangedAt = gfDay.Add(-5 * time.Minute).Format(time.RFC3339)
	m.invalidateView()
	if !m.workingMarkerNeeded() {
		t.Fatal("the verified task asked for no marker ticks")
	}
	requireView(t, m, "◐ 11-demo-mode", "● 10-board-export")
	m, cmd := raise(t, m, markerTickMsg{})
	if cmd == nil {
		t.Fatal("the verified task scheduled no next tick")
	}
	requireView(t, m, "⠋ 11-demo-mode", "● 10-board-export")

	m.tasks.data.items = nil
	m.invalidateView()
	m, cmd = raise(t, m, markerTickMsg{})
	if cmd != nil || m.markerTickRunning {
		t.Fatalf("an empty board kept the timer: running %v, next %v", m.markerTickRunning, cmd != nil)
	}
}

// helper: startMarkerTicks records the first timer without running it.
func startMarkerTicks(t *testing.T, m *Model) *Model {
	t.Helper()
	cmd := m.startMarkerTick()
	if cmd == nil || !m.markerTickRunning {
		t.Fatal("a working marker did not start its timer")
	}
	return m
}
