package component

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// frameMinWidth is the minimum width of the TUI (decision 7). The header, the
// tab bar, the views, and the footer span the full width.
const frameMinWidth = 80

const maxFitWidth = 200

// checkRows checks that view has rows rows of exactly width cells.
func checkRows(t *testing.T, name, view string, width, rows int) []string {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != rows {
		t.Errorf("%s at %d: %d rows, want %d", name, width, len(lines), rows)
	}
	for i, l := range lines {
		if w := lipgloss.Width(l); w != width {
			t.Errorf("%s at %d: row %d is %d cells wide: %q", name, width, i, w, ansi.Strip(l))
		}
	}
	for i := range lines {
		lines[i] = ansi.Strip(lines[i])
	}
	return lines
}

// TestFitFrameRows checks the full-width components at each width from the
// minimum to 200 columns, with text longer than the width.
func TestFitFrameRows(t *testing.T) {
	long := strings.Repeat("long-text-", 30)
	header := Header{Home: "~/" + long, Coordinator: "working", Supervisor: "stopped", NeedsYou: 12345, Clock: "14:22"}
	wordy := Header{Home: "~", Coordinator: long, Supervisor: long, NeedsYou: 1, Clock: long}
	subtabs := SubTabs{Views: []string{"Goals", "Queue", "Tasks", long}, Active: 1}
	hints := KeyHints{Region: "list", Hints: []string{"↑↓ move", long}}
	for width := frameMinWidth; width <= maxFitWidth; width++ {
		checkRows(t, "Header", header.View(width), width, 1)
		checkRows(t, "Header", wordy.View(width), width, 1)
		checkRows(t, "SubTabs", subtabs.View(width), width, 1)
		checkRows(t, "KeyHints", hints.View(width), width, 1)
		for active := range tabNames {
			rows := checkRows(t, "TabBar", TabBar{Active: active}.View(width), width, 2)
			if strings.Trim(rows[len(rows)-1], "─━") != "" {
				t.Errorf("TabBar at %d: rule %q has a cell that is not ─ or ━", width, rows[len(rows)-1])
			}
		}
	}
}

// TestFitPanels checks Panel and Viewport at each width from the minimum to
// 200 columns: each row is exactly the width and the border is whole.
func TestFitPanels(t *testing.T) {
	long := strings.Repeat("wide·text ", 30)
	lines := []string{long, "", "short", long, long, long, long}
	for width := PanelMinWidth; width <= maxFitWidth; width++ {
		for height := PanelMinHeight; height <= 8; height++ {
			p := Panel{Title: long, Focused: height%2 == 0, Lines: lines, VMark: ScrollMark{Start: 1, Size: 3}, HMark: Thumb(width-2, 300, width, 7)}
			checkBorder(t, "Panel", p.View(width, height), width, height)
			v := Viewport{Panel: p, YOffset: 2, XOffset: 3}
			checkBorder(t, "Viewport", v.View(width, height), width, height)
		}
	}
}

func checkBorder(t *testing.T, name, view string, width, height int) {
	t.Helper()
	rows := checkRows(t, name, view, width, height)
	if len(rows) != height {
		return
	}
	r := []rune(rows[0])
	if r[0] != '╭' || r[1] != '─' || r[len(r)-1] != '╮' || r[len(r)-2] != '─' {
		t.Errorf("%s %dx%d: broken top border %q", name, width, height, rows[0])
	}
	for _, row := range rows[1 : height-1] {
		r := []rune(row)
		if r[0] != '│' || (r[len(r)-1] != '│' && r[len(r)-1] != '┃') {
			t.Errorf("%s %dx%d: broken side border %q", name, width, height, row)
		}
	}
	last := rows[height-1]
	if !strings.HasPrefix(last, "╰") || !strings.HasSuffix(last, "╯") || strings.Trim(last, "╰╯─━") != "" {
		t.Errorf("%s %dx%d: broken bottom border %q", name, width, height, last)
	}
}
