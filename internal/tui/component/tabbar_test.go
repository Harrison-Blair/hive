package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestTabBarGolden renders the TabBar section, Components.dc.html lines
// 40-42, and the tab bars of the Work and CompactFriction boards.
func TestTabBarGolden(t *testing.T) {
	t.Run("sheet", func(t *testing.T) { checkGolden(t, TabBar{Active: 0}.View(120), sheet, rect{10, 0, 120, 2}) })
	t.Run("work", func(t *testing.T) { checkGolden(t, TabBar{Active: 2}.View(120), "Work", rect{1, 0, 120, 2}) })
	t.Run("compact", func(t *testing.T) { checkGolden(t, TabBar{Active: 3}.View(80), "CompactFriction", rect{1, 0, 80, 2}) })
}

func TestTabBarSpacing(t *testing.T) {
	cases := []struct {
		width int
		want  string
	}{
		{119, "  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs " + spaces(62)},
		{121, "  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs " + spaces(59)},
	}
	for _, tc := range cases {
		got := ansi.Strip(TabBar{Active: 5}.View(tc.width))
		if want := tc.want + "\n"; got[:len(want)] != want {
			t.Errorf("View(%d) labels = %q, want %q", tc.width, got[:len(want)], want)
		}
	}
}
