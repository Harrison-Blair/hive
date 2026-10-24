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
	subtabs := SubTabs{Views: []string{"Goals", "Queue", "Tasks", long}, Active: 1}
	status := StatusRow{Segments: []string{DimText(long), DimText("short")}, Tail: []string{StrongText("? 2 asks")}, Clock: "14:22"}
	for width := frameMinWidth; width <= maxFitWidth; width++ {
		checkRows(t, "SubTabs", subtabs.View(width), width, 1)
		checkRows(t, "StatusRow", status.View(width), width, 1)
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
	checkBorderOf(t, name, view, width, height, lipgloss.RoundedBorder())
}

// checkBorderOf checks that view has a whole border of b. A scroll mark can
// replace a cell of the right or the bottom border.
func checkBorderOf(t *testing.T, name, view string, width, height int, b lipgloss.Border) {
	t.Helper()
	tl, h, tr, v := []rune(b.TopLeft)[0], []rune(b.Top)[0], []rune(b.TopRight)[0], []rune(b.Left)[0]
	rows := checkRows(t, name, view, width, height)
	if len(rows) != height {
		return
	}
	r := []rune(rows[0])
	if r[0] != tl || r[1] != h || r[len(r)-1] != tr || r[len(r)-2] != h {
		t.Errorf("%s %dx%d: broken top border %q", name, width, height, rows[0])
	}
	for _, row := range rows[1 : height-1] {
		r := []rune(row)
		if r[0] != v || (r[len(r)-1] != v && r[len(r)-1] != '┃') {
			t.Errorf("%s %dx%d: broken side border %q", name, width, height, row)
		}
	}
	last := rows[height-1]
	if !strings.HasPrefix(last, b.BottomLeft) || !strings.HasSuffix(last, b.BottomRight) || strings.Trim(last, b.BottomLeft+b.BottomRight+b.Bottom+"━") != "" {
		t.Errorf("%s %dx%d: broken bottom border %q", name, width, height, last)
	}
}

// TestFitWave4 checks the components of wave 4 at each width from the
// minimum to 200 columns, with text longer than the width: each row is
// exactly the width and each border is whole. The Chat lines and the
// markdown lines are not padded, so they are checked for no line past
// column width-2 (golden-frames chat.md rule 1).
func TestFitWave4(t *testing.T) {
	long := strings.Repeat("wide·text ", 30)
	list := List{Title: long, Focused: true, Columns: []Column{{Title: long, Width: 12}, {Title: "S", Width: 8}, {Title: long}}, Selected: 4}
	var details Details
	for range 9 {
		list.Rows = append(list.Rows, []Cell{{Text: long}, {Text: long}, {Text: long}})
		details.Lines = append(details.Lines, DetailField(long, long))
	}
	details.Title, details.Offset = long, 3
	for width := PanelMinWidth; width <= maxFitWidth; width++ {
		for height := PanelMinHeight; height <= 8; height++ {
			checkBorder(t, "List", list.View(width, height), width, height)
			checkBorder(t, "Details", details.View(width, height), width, height)
		}
		checkBorder(t, "FilterBar", FilterBar{Text: "level:warn " + long, Editing: width%2 == 0}.View(width), width, 3)
		checkBorder(t, "Composer", Composer{Text: long, Cursor: width%2 == 0, CursorColumn: 5, InputRows: 3}.View(width), width, InputHeight(long, width, 3)+2)
		checkRows(t, "Notice", Notice{Text: long}.View(width), width, 1)
		screen := strings.Repeat(long+"\n", 12)
		for _, h := range []int{3, 12} {
			checkRows(t, "Popup", Popup{Details: details, Width: 60, Height: 27}.View(screen, width, h), width, h)
		}
	}
	checkBorder(t, "Card", Card{Name: long, Sub: long, State: "idle"}.View(), CardWidth, CardHeight)
	checkBorderOf(t, "Selected card", Card{Name: long, Sub: long, State: "idle", Selected: true}.View(), CardWidth, CardHeight, lipgloss.ThickBorder())
	longNames := []ChatMessage{
		{Time: "09:00", Name: long, Kind: ChatEvent, Text: long},
		{Time: "09:00", Name: long, Kind: ChatAgent, Text: "ok"},
		{Time: "09:00", Name: long, Kind: ChatUser, Text: "ok"},
	}
	for width := 30; width <= maxFitWidth; width++ {
		lines, err := ChatLines(append(gfChatSample(), longNames...), width)
		if err != nil {
			t.Fatal(err)
		}
		for i, l := range lines {
			if w := lipgloss.Width(l); w > width-1 {
				t.Errorf("Chat at %d: line %d is %d cells, want at most %d", width, i, w, width-1)
			}
		}
	}
}
