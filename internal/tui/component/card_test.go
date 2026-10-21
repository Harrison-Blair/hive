package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestCardGolden renders the three cards of the Card section,
// Components.dc.html lines 71-73.
func TestCardGolden(t *testing.T) {
	cases := []struct {
		card Card
		r    rect
	}{
		{Card{Name: "manager-hive", Sub: "claude · working", State: "working", Selected: true}, rect{51, 0, 24, 4}},
		{Card{Name: "t3-verifier-1", Sub: "pi · idle · 6m", State: "idle"}, rect{51, 26, 24, 4}},
		{Card{Name: "t4-worker-1", Sub: "pi · stopped", State: "stopped"}, rect{51, 52, 24, 4}},
	}
	for _, tc := range cases {
		t.Run(tc.card.State, func(t *testing.T) { checkGolden(t, tc.card.View(), sheet, tc.r) })
	}
}

// TestCardCutsLongText checks that the name and the sub line are cut with …
// inside the 24×4 card, as Main.dc.html row 9 shows.
func TestCardCutsLongText(t *testing.T) {
	c := Card{Name: "a-very-long-agent-name-here", Sub: "claude · idle · nudged 4m", State: "idle"}
	want := "╭──────────────────────╮\n│ ◐ a-very-long-agent… │\n│ claude · idle · nud… │\n╰──────────────────────╯"
	if got := ansi.Strip(c.View()); got != want {
		t.Errorf("View:\n%s\nwant:\n%s", got, want)
	}
}
