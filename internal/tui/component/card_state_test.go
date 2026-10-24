package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// tui-v2 decision 55 and agents.md: each agent state has its glyph. Working
// is ●, idle is ◐, blocked is ▲, and each other state is ○.
func TestCardStateGlyphs(t *testing.T) {
	for _, tc := range []struct {
		state string
		glyph string
	}{
		{"working", "●"},
		{"idle", "◐"},
		{"blocked", "▲"},
		{"starting", "○"},
		{"unresolved", "○"},
		{"stale", "○"},
		{"stopped", "○"},
	} {
		t.Run(tc.state, func(t *testing.T) {
			rows := strings.Split(ansi.Strip((Card{Name: "probe", State: tc.state}).View()), "\n")
			if !strings.Contains(rows[1], "│ "+tc.glyph+" probe") {
				t.Fatalf("card of %s lacks the glyph %q:\n%s", tc.state, tc.glyph, strings.Join(rows, "\n"))
			}
		})
	}
}
