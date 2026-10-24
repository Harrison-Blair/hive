package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// The Help overlay has a layout-rule test that reads no file under
// .hive/spec/. golden-frames screens/help.md rule 1: the overlay is a rounded
// box at the center of the screen. Rule 4: a body taller than the box scrolls
// with the ┃ thumb on the right border, and a body that fits has no thumb.
// The frame keeps the terminal size (tui-v2 decision 7).
func TestHelpOverlayLayoutRules(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 57}} {
		width, height := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			m := chatModel(t, width, height, 0)
			m, _ = raise(t, m, key('?', "?"))
			lines := strings.Split(ansi.Strip(m.View().Content), "\n")
			if len(lines) != height {
				t.Fatalf("view has %d lines, want %d", len(lines), height)
			}
			top, bottom := -1, -1
			for index, line := range lines {
				if got := ansi.StringWidth(line); got != width {
					t.Fatalf("line %d has width %d, want %d: %q", index+1, got, width, line)
				}
				if strings.Contains(line, "╭─ help · esc or ? closes · ↑↓ scroll") {
					top = index
				}
				if top >= 0 && bottom < 0 && strings.Contains(line, "╰") {
					bottom = index
				}
			}
			if top < 0 || bottom < 0 {
				t.Fatalf("no rounded help box in the view:\n%s", strings.Join(lines, "\n"))
			}
			left := ansi.StringWidth(lines[top][:strings.Index(lines[top], "╭")])
			right := width - ansi.StringWidth(strings.TrimRight(lines[top], " "))
			if diff := left - right; diff < -1 || diff > 1 {
				t.Errorf("box margins left %d, right %d; want the box at the center", left, right)
			}
			if !strings.HasSuffix(strings.TrimRight(lines[bottom], " "), "╯") || !strings.HasSuffix(strings.TrimRight(lines[top], " "), "╮") {
				t.Errorf("box corners are not rounded: %q, %q", lines[top], lines[bottom])
			}
			thumb := false
			for _, line := range lines[top+1 : bottom] {
				if strings.HasSuffix(strings.TrimRight(line, " "), "┃") {
					thumb = true
				}
			}
			if fits := len(m.helpLines())+2 <= height; thumb == fits {
				t.Errorf("thumb %v with a body of %d lines in %d rows; want a thumb only when the body is taller than the box", thumb, len(m.helpLines()), height)
			}
		})
	}
}
