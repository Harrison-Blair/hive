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
// with the ┃ thumb on the right border. tui-v2 decision 118: one blank row is
// above FILTER FIELDS. The frame keeps the terminal size (tui-v2 decision 7).
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
				if strings.Contains(line, "╭─ keys · ? or esc closes") {
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
			if !thumb {
				t.Errorf("the body is taller than the box but the right border has no ┃ thumb")
			}
			for index := top + 1; index < bottom; index++ {
				if strings.Contains(lines[index], "FILTER FIELDS") {
					above := strings.Trim(lines[index-1], " │┃")
					twoAbove := strings.Trim(lines[index-2], " │┃")
					if above != "" || twoAbove == "" {
						t.Errorf("rows above FILTER FIELDS are %q, %q; want one blank row", lines[index-2], lines[index-1])
					}
				}
			}
		})
	}
}
