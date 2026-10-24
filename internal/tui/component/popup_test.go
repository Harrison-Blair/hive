package component

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestPopupGolden renders the popup of AgentDetails.dc.html lines 54-64 on a
// screen of the size of the box, so the box covers the screen. The TARGET
// rows of the board move RECENT up one row and end each dim run at its text.
func TestPopupGolden(t *testing.T) {
	dim := lipgloss.NewStyle().Foreground(Dim)
	accent := lipgloss.NewStyle().Foreground(Accent)
	key := lipgloss.NewStyle().Foreground(Accent).Bold(true)
	lines := []string{""}
	for _, f := range [][2]string{{"role", "manager"}, {"project", "hive"}, {"status", accent.Render("◐") + " live · idle"}, {"nudged", "4m ago by supervisor"}, {"harness", "claude"}, {"model", "claude-opus-5-5"}, {"pane", "w58:p6"}, {"tab", "manager-hive"}, {"goal", "g1-tui-restart · done"}, {"parent", "coordinator-g2"}, {"replaces", "manager-hive-g1"}} {
		lines = append(lines, DetailField(f[0], f[1]))
	}
	lines = append(lines, "", DetailSection("ACTIONS"),
		key.Render("a")+"  attach to pane",
		key.Render("p")+"  peek pane",
		dim.Bold(true).Render("t")+dim.Render("  open task · none"),
		"", DetailSection("RECENT"),
		fitCell("14:08 inbox 28 sent to coordinator", 54))
	for _, s := range []string{"14:07 land 4-agents-equal-columns · ede94f3e", "13:58 land 3-tui-restart · 3ccd93fb", "13:44 land 2-stale-coordinator-startup · 0e079272"} {
		lines = append(lines, dim.Render(s))
	}
	p := Popup{Details: Details{Title: "manager-hive · esc closes", Lines: lines}, Width: 60, Height: 27}
	screen := strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", 60)+"\n", 27), "\n")
	checkGolden(t, p.View(screen, 60, 27), "AgentDetails", rect{7, 30, 60, 27})
}

// TestPopupDimsScreen draws a 7×3 popup over a styled 11×5 screen. Each cell
// behind the popup is border and not bold, and the box is centered.
func TestPopupDimsScreen(t *testing.T) {
	bold := lipgloss.NewStyle().Bold(true).Foreground(Accent)
	screen := bold.Render("hive header") + "\n" + strings.Repeat("x", 11) + "\n" + "abcdefghijk\n" + strings.Repeat("y", 11) + "\n" + lipgloss.NewStyle().Foreground(Info).Render("z")
	p := Popup{Details: Details{Title: "p", Lines: []string{"ok"}}, Width: 7, Height: 3}
	want := "hive-grid 1 11 5\n" +
		"t|hive header\n" + "k|rrrrrrrrrrr\n" + "b|00000000000\n" +
		"t|xx╭─ p ─╮xx\n" + "k|rraaaaaaarr\n" + "b|00001110000\n" +
		"t|ab│  ok │jk\n" + "k|rrafffffarr\n" + "b|00000000000\n" +
		"t|yy╰─────╯yy\n" + "k|rraaaaaaarr\n" + "b|00000000000\n" +
		"t|z          \n" + "k|rrrrrrrrrrr\n" + "b|00000000000\n"
	checkHandGrid(t, p.View(screen, 11, 5), want)
}

// TestDimScreen checks that each cell of styled rows is border and not
// bold, and that short rows are padded to the width.
func TestDimScreen(t *testing.T) {
	screen := titleStyle.Render("  1 Agents  ") + "\n" + borderStyle.Render("───────────") + accentStyle.Render("━") +
		"\n" + lipgloss.NewStyle().Bold(true).Render("ab")
	want := "hive-grid 1 12 3\n" +
		"t|  1 Agents  \n" + "k|rrrrrrrrrrrr\n" + "b|000000000000\n" +
		"t|───────────━\n" + "k|rrrrrrrrrrrr\n" + "b|000000000000\n" +
		"t|ab          \n" + "k|rrrrrrrrrrrr\n" + "b|000000000000\n"
	checkHandGrid(t, DimScreen(screen, 12, 3), want)
}
