package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestCardGolden renders the two unselected cards of the Card section,
// Components.dc.html lines 71-73. TestCardSelectedHeavyBorder checks the
// selected card.
func TestCardGolden(t *testing.T) {
	cases := []struct {
		card Card
		r    rect
	}{
		{Card{Name: "t3-verifier-1", Sub: "pi · idle · 6m", State: "idle"}, rect{51, 26, 24, 4}},
		{Card{Name: "t4-worker-1", Sub: "pi · stopped", State: "stopped"}, rect{51, 52, 24, 4}},
	}
	for _, tc := range cases {
		t.Run(tc.card.State, func(t *testing.T) { checkGolden(t, tc.card.View(), sheet, tc.r) })
	}
}

// TestCardSelectedHeavyBorder checks that the selected card has a heavy
// border in accent, the plain-text signal of selection (user decision
// 2026-10-21, option A). The cells inside the border are those of
// Components.dc.html line 71.
func TestCardSelectedHeavyBorder(t *testing.T) {
	c := Card{Name: "manager-hive", Sub: "claude · working", State: "working", Selected: true}
	want := "hive-grid 1 24 4\n" +
		"t|┏━━━━━━━━━━━━━━━━━━━━━━┓\n" + "k|aaaaaaaaaaaaaaaaaaaaaaaa\n" + "b|000000000000000000000000\n" +
		"t|┃ ● manager-hive       ┃\n" + "k|afwfaaaaaaaaaaaafffffffa\n" + "b|000011111111111100000000\n" +
		"t|┃ claude · working     ┃\n" + "k|afddddddddddddddddfffffa\n" + "b|000000000000000000000000\n" +
		"t|┗━━━━━━━━━━━━━━━━━━━━━━┛\n" + "k|aaaaaaaaaaaaaaaaaaaaaaaa\n" + "b|000000000000000000000000\n"
	checkHandGrid(t, c.View(), want)
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

// TestCardWorkingBorder checks decision 60: the card of a working agent has
// its rounded border in working. Each other state keeps border, and a
// selected working card keeps the heavy accent border
// (TestCardSelectedHeavyBorder).
func TestCardWorkingBorder(t *testing.T) {
	c := Card{Name: "t3-worker-2", Sub: "claude · working", State: "working"}
	want := "hive-grid 1 24 4\n" +
		"t|╭──────────────────────╮\n" + "k|wwwwwwwwwwwwwwwwwwwwwwww\n" + "b|000000000000000000000000\n" +
		"t|│ ● t3-worker-2        │\n" + "k|wfwffffffffffffffffffffw\n" + "b|000000000000000000000000\n" +
		"t|│ claude · working     │\n" + "k|wfddddddddddddddddfffffw\n" + "b|000000000000000000000000\n" +
		"t|╰──────────────────────╯\n" + "k|wwwwwwwwwwwwwwwwwwwwwwww\n" + "b|000000000000000000000000\n"
	checkHandGrid(t, c.View(), want)
	for _, state := range []string{"idle", "stopped", "blocked"} {
		c := Card{Name: "t3-worker-2", Sub: "claude · " + state, State: state}
		top := strings.SplitN(c.View(), "\n", 2)[0]
		if !strings.Contains(top, "38;2;61;58;51m╭") {
			t.Errorf("%s card top border %q, want the token border", state, top)
		}
	}
}
