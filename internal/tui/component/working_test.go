package component

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
)

// The Pi 1.4.0 working indicator uses ten braille frames at 80 ms each.
// Frame 0 keeps the static state glyph of the approved boards.
func TestWorkingGlyphUsesThePiFrames(t *testing.T) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if WorkingFrameInterval != 80*time.Millisecond {
		t.Fatalf("WorkingFrameInterval = %v, want 80ms", WorkingFrameInterval)
	}
	for frame, want := range append([]string{"●"}, frames...) {
		if got := WorkingGlyph(frame, "●"); got != want {
			t.Errorf("WorkingGlyph(%d, ●) = %q, want %q", frame, got, want)
		}
	}
	for _, frame := range frames {
		if w := lipgloss.Width(frame); w != 1 {
			t.Errorf("the Pi frame %q takes %d cells, want 1", frame, w)
		}
	}
	if got := WorkingGlyph(len(frames)+1, "●"); got != "⠋" {
		t.Errorf("WorkingGlyph after one cycle = %q, want ⠋", got)
	}
	if got := WorkingGlyph(1, "◐"); got != "⠋" {
		t.Errorf("WorkingGlyph(1, ◐) = %q, want ⠋", got)
	}
}
