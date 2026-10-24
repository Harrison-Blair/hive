package tui

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// frictionText gives the text of an entry file with its front matter.
func frictionText(title, status, opened, decision, body string) string {
	return fmt.Sprintf("---\ntitle: %q\nstatus: %q\nopened: %s\ndecision: %q\n---\n\n%s", title, status, opened, decision, body)
}

// frictionHandEntries are two open and two resolved entries. The clock of
// the tests is 2026-10-23.
var frictionHandEntries = map[string]string{
	"entries/alpha-open.md": frictionText("Alpha issue", "Open", "2026-10-20",
		"keep the alpha decision text long enough to be cut at the end of the row because it does not fit",
		"**Summary:**\nAlpha summary.\n\n**Reproduction steps:**\n1. Do alpha.\n2. See alpha.\n\n**Occurrences:**\n2026-10-20  one\n2026-10-21  two\n2026-10-21  three\n2026-10-22  four\n"),
	"entries/beta-open.md": frictionText("Beta issue", "Open", "2026-10-21", "",
		"**Summary:**\nBeta summary.\n\n**Occurrences:**\n2026-10-21  one\n2026-10-21  two\n"),
	"entries/resolved/gamma-done.md": frictionText("Gamma issue", "Resolved 2026-10-22", "2026-10-18", "gamma fixed",
		"**Summary:**\nGamma summary.\n"),
	"entries/resolved/delta-done.md": frictionText("Delta Agent issue", "Resolved 2026-10-19", "2026-10-10", "delta fixed",
		"**Summary:**\nDelta summary.\n"),
}

// newFrictionModel gives a model at 120×40 on the Friction screen of a
// scratch home with the entry files, by path below the friction folder.
func newFrictionModel(t *testing.T, entries map[string]string) (*Model, string) {
	t.Helper()
	m, dir := newFrameModel(t, 120, 40)
	root := filepath.Join(dir, "reference", "dogfood", "friction")
	writeText(t, filepath.Join(root, "FRICTION.md"), "# Friction\n")
	for path, text := range entries {
		writeText(t, filepath.Join(root, filepath.FromSlash(path)), text)
	}
	m, cmd := raise(t, m, ctrl('f'))
	m = deliverAgentCommands(t, m, cmd)
	return m, dir
}

// press gives each key to m and runs the commands.
func press(t *testing.T, m *Model, keys ...tea.KeyPressMsg) *Model {
	t.Helper()
	for _, k := range keys {
		var cmd tea.Cmd
		m, cmd = raise(t, m, k)
		m = deliverAgentCommands(t, m, cmd)
	}
	return m
}

var (
	downKey = tea.KeyPressMsg{Code: tea.KeyDown}
	upKey   = tea.KeyPressMsg{Code: tea.KeyUp}
)

// typeText gives each rune of text to m as a key.
func typeText(t *testing.T, m *Model, text string) *Model {
	t.Helper()
	for _, r := range text {
		m = press(t, m, key(r, string(r)))
	}
	return m
}

// requireRow fails when row index of the plain view is not want.
func requireRow(t *testing.T, m *Model, index int, want string) {
	t.Helper()
	if got := viewRows(m)[index]; got != want {
		t.Fatalf("row %d =\n%q\nwant\n%q\nview:\n%s", index, got, want, strings.Join(viewRows(m), "\n"))
	}
}

// requireRowPrefix fails when row index of the plain view does not start
// with want.
func requireRowPrefix(t *testing.T, m *Model, index int, want string) {
	t.Helper()
	if got := viewRows(m)[index]; !strings.HasPrefix(got, want) {
		t.Fatalf("row %d = %q; want the prefix %q\nview:\n%s", index, got, want, strings.Join(viewRows(m), "\n"))
	}
}

// Friction layout and rules 1 and 2: the title row counts the entries. An
// entry row shows the name, the age since opened in days, one ▮ for each
// occurrence, the last-seen date, and the decision cut with … at the row
// end. The first selection is the first open entry, the newest opened.
func TestFrictionRowsShowEachColumn(t *testing.T) {
	m, _ := newFrictionModel(t, frictionHandEntries)
	requireRowPrefix(t, m, 0, " friction · 2 open · 2 resolved ")
	requireRow(t, m, 2, "   ENTRY                                AGE   SEEN       LAST   DECISION                                                ")
	requireRow(t, m, 3, " ▶ beta-open                            2d    ▮▮         10-21                                                          ")
	requireRow(t, m, 4, "   alpha-open                           3d    ▮▮▮▮       10-22  keep the alpha decision text long enough to be cut at…  ")
	requireRow(t, m, 5, strings.Repeat(" ", 120))
	requireRowPrefix(t, m, 6, " ▸ resolved · 2  enter opens ")
	grid := renderGrid(t, m, 120, 40)
	if c := grid.At(3, 1); c.Token != cellgrid.Accent || !c.Bold {
		t.Errorf("▶ is %v bold %v; want accent bold", c.Token, c.Bold)
	}
	if c := grid.At(3, 3); c.Token != cellgrid.Accent || !c.Bold {
		t.Errorf("selected name is %v bold %v; want accent bold", c.Token, c.Bold)
	}
	if a, b := grid.At(4, 46), grid.At(3, 46); a.Token != cellgrid.Accent || b.Token != cellgrid.Dim {
		t.Errorf("seen bars are %v for 4 and %v for 2; want accent from 3, else dim", a.Token, b.Token)
	}
}

// Friction rule 3: enter on the fold line expands the resolved entries,
// newest resolved first, with the resolved date in LAST. ↑ and ↓ walk into
// them, and enter on the fold line folds them again.
func TestFrictionFoldLineExpandsResolvedEntries(t *testing.T) {
	m, _ := newFrictionModel(t, frictionHandEntries)
	m = press(t, m, enterKey)
	requireRowPrefix(t, m, 6, " ▸ resolved · 2  enter opens ")
	m = press(t, m, downKey, downKey, enterKey)
	requireRowPrefix(t, m, 6, " ▶ resolved · 2  enter folds ")
	requireRowPrefix(t, m, 7, "   gamma-done                           5d    ▮          10-22  gamma fixed ")
	requireRowPrefix(t, m, 8, "   delta-done                           13d   ▮          10-19  delta fixed ")
	m = press(t, m, downKey, downKey)
	requireRowPrefix(t, m, 8, " ▶ delta-done ")
	m = press(t, m, upKey, upKey, enterKey)
	requireRowPrefix(t, m, 6, " ▶ resolved · 2  enter opens ")
	if row := viewRows(m)[7]; strings.TrimSpace(row) != "" {
		t.Fatalf("row 7 = %q after the fold; want a blank row", row)
	}
}

// Friction rule 3: the list scrolls with the ┃ thumb on column 119 when its
// rows do not fit, and the selected row stays visible.
func TestFrictionListScrollsWithThumb(t *testing.T) {
	entries := map[string]string{}
	for i := range 30 {
		entries[fmt.Sprintf("entries/item-%02d.md", i)] = frictionText(fmt.Sprintf("Item %d", i), "Open", "2026-10-20", "", "**Summary:**\nText.\n")
	}
	m, _ := newFrictionModel(t, entries)
	rows := viewRows(m)
	if r := []rune(rows[3]); r[119] != '┃' {
		t.Fatalf("row 3 = %q; want the thumb on column 119", rows[3])
	}
	if r := []rune(rows[19]); r[119] == '┃' {
		t.Fatalf("row 19 = %q; want no thumb at the end of the track", rows[19])
	}
	m = press(t, m, key('G', "G"))
	requireRowPrefix(t, m, 19, " ▶ resolved · 0  enter opens ")
	if r := []rune(viewRows(m)[19]); r[119] != '┃' {
		t.Fatalf("row 19 = %q after G; want the thumb at the end", viewRows(m)[19])
	}
	m = press(t, m, key('g', "g"))
	requireRowPrefix(t, m, 3, " ▶ item-00 ")
}

// Friction rule 4: the details box shows the selected entry. Its title is
// NAME · STATUS · opened DATE · seen N · PATH, and the body has the sections
// in accent bold. A resolved entry shows Resolved DATE as the status.
func TestFrictionDetailsShowTheSelectedEntry(t *testing.T) {
	m, _ := newFrictionModel(t, frictionHandEntries)
	m = press(t, m, downKey)
	requireRowPrefix(t, m, 21, "╭─ alpha-open · Open · opened 2026-10-20 · seen 4 · entries/alpha-open.md ─")
	for index, want := range []string{
		"│ SUMMARY ", "│ Alpha summary. ", "│  ", "│ REPRODUCTION ", "│ 1. Do alpha. ", "│ 2. See alpha. ", "│  ",
		"│ OCCURRENCES ", "│ 2026-10-20  one     2026-10-21  two     2026-10-21  three     2026-10-22  four ", "│  ",
		"│ DECISION keep the alpha decision text long enough to be cut at the end of the row because it does not fit ",
	} {
		requireRowPrefix(t, m, 22+index, want)
	}
	grid := renderGrid(t, m, 120, 40)
	if c := grid.At(22, 2); c.Token != cellgrid.Accent || !c.Bold {
		t.Errorf("SUMMARY is %v bold %v; want accent bold", c.Token, c.Bold)
	}
	m = press(t, m, downKey, enterKey, downKey)
	requireRowPrefix(t, m, 21, "╭─ gamma-done · Resolved 2026-10-22 · opened 2026-10-18 · seen 1 · entries/resolved/gamma-done.md ─")
}

// Friction rule 5 and the details focus state: ctrl+] gives the details box
// the focus and the accent border. The box takes the scroll keys, shows the
// ┃ thumb on its right border for a long entry, and ctrl+] gives the focus
// back to the list.
func TestFrictionDetailsFocusScrolls(t *testing.T) {
	long := ""
	for i := range 40 {
		long += fmt.Sprintf("- Line %d of the summary.\n", i)
	}
	m, _ := newFrictionModel(t, map[string]string{"entries/long.md": frictionText("Long", "Open", "2026-10-20", "", "**Summary:**\n"+long)})
	if r := []rune(viewRows(m)[22]); r[119] != '┃' {
		t.Fatalf("row 22 = %q; want the thumb on the right border of the box", viewRows(m)[22])
	}
	m = press(t, m, ctrl(']'))
	requireKeys(t, m, "details", "↑↓ scroll")
	if c := renderGrid(t, m, 120, 40).At(21, 0); c.Token != cellgrid.Accent {
		t.Fatalf("box corner is %v; want the accent border of the focused box", c.Token)
	}
	requireRowPrefix(t, m, 23, "│ - Line 0 of the summary.")
	m = press(t, m, downKey, downKey)
	requireRowPrefix(t, m, 22, "│ - Line 1 of the summary.")
	m = press(t, m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	requireRowPrefix(t, m, 22, "│ - Line 16 of the summary.")
	m = press(t, m, key('G', "G"))
	requireRowPrefix(t, m, 36, "│ - Line 39 of the summary.")
	m = press(t, m, key('g', "g"))
	requireRowPrefix(t, m, 22, "│ SUMMARY")
	requireRowPrefix(t, m, 3, " ▶ long ")
	m = press(t, m, ctrl(']'))
	requireKeys(t, m, "list", "↑↓ move")
}

// Friction rule 6: / opens the filter line on row 2. enter applies the text
// and closes the line, and the title row shows · / TEXT. The filter matches
// the name and the title, case-insensitive. esc closes the line and keeps
// the applied filter. With no match, the list shows no entry matches TEXT.
func TestFrictionFilter(t *testing.T) {
	m, _ := newFrictionModel(t, frictionHandEntries)
	m = press(t, m, key('/', "/"))
	requireKeys(t, m, "filter", "type · enter apply · esc close · backspace")
	m = typeText(t, m, "ALPHA")
	requireRowPrefix(t, m, 2, " / ALPHA █")
	if !strings.Contains(viewRows(m)[2], "enter applies · esc closes and keeps the applied filter") {
		t.Fatalf("row 2 = %q; want the hint of the filter line", viewRows(m)[2])
	}
	requireRowPrefix(t, m, 3, " ▶ beta-open ")
	m = press(t, m, enterKey)
	requireRowPrefix(t, m, 0, " friction · 2 open · 2 resolved · / ALPHA ")
	requireRowPrefix(t, m, 2, "   ENTRY ")
	requireRowPrefix(t, m, 3, " ▶ alpha-open ")
	requireRowPrefix(t, m, 5, " ▸ resolved · 2 · 0 match  enter opens ")
	if c := renderGrid(t, m, 120, 40).At(0, 34); c.Token != cellgrid.Accent {
		t.Fatalf("title tag is %v; want accent", c.Token)
	}
	// The title matches: Delta Agent issue.
	m = press(t, m, key('/', "/"))
	for range len("ALPHA") {
		m = press(t, m, tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	m = typeText(t, m, "agent issue")
	m = press(t, m, enterKey)
	requireRowPrefix(t, m, 3, "   no open entries ")
	requireRowPrefix(t, m, 5, " ▸ resolved · 2 · 1 match  enter opens ")
	m = press(t, m, key('/', "/"))
	m = typeText(t, m, "zzz")
	m = press(t, m, escKey())
	requireRowPrefix(t, m, 0, " friction · 2 open · 2 resolved · / agent issue ")
	if m.activeTab() != tabFriction {
		t.Fatal("esc on the filter line left the screen")
	}
	m = press(t, m, key('/', "/"))
	m = typeText(t, m, "zzz")
	m = press(t, m, enterKey)
	requireRowPrefix(t, m, 3, "   no entry matches agent issuezzz ")
	requireRowPrefix(t, m, 21, "╭─ no entry ─")
}

// Friction rule 7: with no open entry, row 3 shows no open entries. The fold
// line and the details box stay. With no entry at all, the box shows no
// entry in dim.
func TestFrictionNoOpenEntries(t *testing.T) {
	m, _ := newFrictionModel(t, map[string]string{
		"entries/resolved/gamma-done.md": frictionHandEntries["entries/resolved/gamma-done.md"],
	})
	requireRowPrefix(t, m, 0, " friction · 0 open · 1 resolved ")
	requireRowPrefix(t, m, 3, "   no open entries ")
	requireRowPrefix(t, m, 5, " ▸ resolved · 1  enter opens ")
	requireRowPrefix(t, m, 21, "╭─ no entry ─")
	m, _ = newFrictionModel(t, nil)
	requireRowPrefix(t, m, 3, "   no open entries ")
	requireRowPrefix(t, m, 21, "╭─ no entry ─")
	if c := renderGrid(t, m, 120, 40).At(21, 3); c.Token != cellgrid.Dim {
		t.Fatalf("no entry is %v; want dim", c.Token)
	}
}

// Friction rule 8: no key opens the entry file. enter on an entry row does
// nothing.
func TestFrictionEnterOnAnEntryDoesNothing(t *testing.T) {
	m, _ := newFrictionModel(t, frictionHandEntries)
	before := m.View().Content
	m, cmd := raise(t, m, enterKey)
	if cmd != nil || m.View().Content != before {
		t.Fatalf("enter on an entry changed the screen:\n%s", strings.Join(viewRows(m), "\n"))
	}
}

// The key table: esc returns to Chat from each box when no filter line is
// open.
func TestFrictionEscReturnsToChat(t *testing.T) {
	for _, keys := range [][]tea.KeyPressMsg{{escKey()}, {ctrl(']'), escKey()}} {
		m, _ := newFrictionModel(t, frictionHandEntries)
		m = press(t, m, keys...)
		if m.activeTab() != tabChat {
			t.Fatalf("keys %v gave tab %d; want Chat", keys, m.activeTab())
		}
	}
}

// AGENTS.md logging: each user action on the screen writes a tui action
// event.
func TestFrictionLogsEachAction(t *testing.T) {
	m, dir := newFrictionModel(t, frictionHandEntries)
	m = press(t, m, downKey, downKey, enterKey, key('/', "/"))
	m = typeText(t, m, "a")
	press(t, m, enterKey, ctrl(']'), downKey, ctrl(']'), key('/', "/"), escKey())
	var got []string
	for _, event := range tuiActions(t, dir) {
		got = append(got, event["action"].(string))
	}
	for _, want := range []string{"friction selection", "friction resolved expand", "friction filter open", "friction filter apply", "focus region", "friction filter close"} {
		if !slices.Contains(got, want) {
			t.Errorf("actions %v; want %q", got, want)
		}
	}
}

// README decisions 6 and 20: a click on a list row selects it, and a click
// on the details box gives it the focus. The wheel scrolls the box below the
// pointer and keeps the focus.
func TestFrictionMouse(t *testing.T) {
	m, _ := newFrictionModel(t, frictionHandEntries)
	m = press(t, m)
	m, _ = raise(t, m, click(10, 4))
	requireRowPrefix(t, m, 4, " ▶ alpha-open ")
	m, _ = raise(t, m, click(10, 25))
	if m.focusState() != "details" {
		t.Fatalf("a click on the box gave focus %q; want details", m.focusState())
	}
	m, _ = raise(t, m, wheel(10, 3, tea.MouseWheelUp))
	requireRowPrefix(t, m, 3, " ▶ beta-open ")
	if m.focusState() != "details" {
		t.Fatalf("the wheel over the list moved the focus to %q", m.focusState())
	}
	m, _ = raise(t, m, click(10, 6))
	requireRowPrefix(t, m, 6, " ▶ resolved · 2 ")
	if m.focusState() != "list" {
		t.Fatalf("a click on the list gave focus %q; want list", m.focusState())
	}
}

// The status row: the screen segment shows the keys of the list focus.
func TestFrictionListKeys(t *testing.T) {
	m, _ := newFrictionModel(t, frictionHandEntries)
	requireKeys(t, m, "list", "↑↓ move · enter resolved · / filter · ^] details · pgup/pgdn")
}

// R reads the entries again in normal mode and keeps the selection.
func TestFrictionRefreshKeepsSelection(t *testing.T) {
	m, dir := newFrictionModel(t, frictionHandEntries)
	m = press(t, m, downKey)
	writeText(t, filepath.Join(dir, "reference", "dogfood", "friction", "entries", "zeta-open.md"),
		frictionText("Zeta issue", "Open", "2026-10-22", "", "**Summary:**\nZeta.\n"))
	m = press(t, m, key('R', "R"))
	requireRowPrefix(t, m, 0, " friction · 3 open · 2 resolved ")
	requireRowPrefix(t, m, 3, "   zeta-open ")
	requireRowPrefix(t, m, 5, " ▶ alpha-open ")
}

// An entry file with no valid front matter shows ? in its columns and its
// status.
func TestFrictionInvalidEntry(t *testing.T) {
	m, _ := newFrictionModel(t, map[string]string{"entries/invalid.md": "No front matter\n"})
	requireRowPrefix(t, m, 3, " ▶ invalid                              ?                ? ")
	requireRowPrefix(t, m, 21, "╭─ invalid · ? · entries/invalid.md ─")
	requireRowPrefix(t, m, 22, "│ the front matter of this entry is not valid ")
}

// Friction rule 1 and board case F15: a bar stops at 10 ▮, and the SEEN
// column of 11 cells keeps one gap cell before LAST.
func TestFrictionSeenBarFitsItsColumn(t *testing.T) {
	occurrences := ""
	for range 12 {
		occurrences += "2026-10-22  again\n"
	}
	m, _ := newFrictionModel(t, map[string]string{"entries/busy.md": frictionText("Busy", "Open", "2026-10-20", "busy decision", "**Occurrences:**\n"+occurrences)})
	requireRowPrefix(t, m, 3, " ▶ busy                                 3d    ▮▮▮▮▮▮▮▮▮▮ 10-22  busy decision ")
}
