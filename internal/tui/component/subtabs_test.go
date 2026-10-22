package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func spaces(n int) string { return strings.Repeat(" ", n) }

// TestSubTabsGolden renders the SubTabs section, Components.dc.html line 44,
// and the views of the Work and CompactHealth boards.
func TestSubTabsGolden(t *testing.T) {
	work := []string{"Goals", "Queue", "Tasks"}
	t.Run("sheet", func(t *testing.T) {
		checkGolden(t, SubTabs{Views: work, Active: 2}.View(120), sheet, rect{14, 0, 120, 1})
	})
	t.Run("work", func(t *testing.T) {
		checkGolden(t, SubTabs{Views: work, Active: 2}.View(120), "Work", rect{3, 0, 120, 1})
	})
	t.Run("compact", func(t *testing.T) {
		checkGolden(t, SubTabs{Views: []string{"Home", "Agents"}, Active: 1}.View(80), "CompactHealth", rect{3, 0, 80, 1})
	})
}

func TestSubTabsText(t *testing.T) {
	got := ansi.Strip(SubTabs{Views: []string{"Open", "Resolved"}, Active: 0}.View(90))
	want := " [Open]   Resolved " + spaces(53) + "[ ] switch view   "
	if got != want {
		t.Errorf("View(90) = %q, want %q", got, want)
	}
}

// TestSubTabsHint checks the hint of the Compact board, row 3: while the
// details overlay is open, "list › details" in dim ends 2 cells before the
// right edge.
func TestSubTabsHint(t *testing.T) {
	got := ansi.Strip(SubTabs{Views: []string{"Goals", "Queue", "Tasks"}, Active: 2, Hint: "list › details"}.View(80))
	want := "  Goals    Queue   [Tasks]" + spaces(38) + "list › details  "
	if got != want {
		t.Errorf("View(80) = %q, want %q", got, want)
	}
}
