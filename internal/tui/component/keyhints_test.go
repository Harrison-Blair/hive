package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

var graphHints = []string{"←↑↓→ move", "enter details", "a attach", "p peek", "t task"}

// TestKeyHintsGolden renders the KeyHints section, Components.dc.html line
// 79, at the width of its text (72 cells), and the footers of the Main and
// CompactAgents boards, which pad the hints in dim to the full width.
func TestKeyHintsGolden(t *testing.T) {
	t.Run("sheet", func(t *testing.T) {
		k := KeyHints{Region: "graph", Hints: append(graphHints[:5:5], "? help")}
		checkGolden(t, k.View(72), sheet, rect{64, 0, 72, 1})
	})
	t.Run("main", func(t *testing.T) {
		k := KeyHints{Region: "graph", Hints: append(graphHints[:5:5], "h hide stopped", "? help")}
		checkGolden(t, k.View(120), "Main", rect{39, 0, 120, 1})
	})
	t.Run("compact", func(t *testing.T) {
		k := KeyHints{Region: "graph", Hints: append(graphHints[:4:4], "?")}
		checkGolden(t, k.View(80), "CompactAgents", rect{23, 0, 80, 1})
	})
}

func TestKeyHintsFit(t *testing.T) {
	k := KeyHints{Region: "list", Hints: []string{"↑↓ move", "enter open"}}
	cases := []struct {
		width int
		want  string
	}{
		{30, " [list] ↑↓ move · enter open  "},
		{28, " [list] ↑↓ move · enter open"},
		{27, " [list] ↑↓ move · enter op…"},
		{16, " [list] ↑↓ move…"},
	}
	for _, tc := range cases {
		if got := ansi.Strip(k.View(tc.width)); got != tc.want {
			t.Errorf("View(%d) = %q, want %q", tc.width, got, tc.want)
		}
	}
}
