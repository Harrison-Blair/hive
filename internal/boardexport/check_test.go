package boardexport

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// raw builds a hand-written board grid. Each row is "text|tokens", with one
// token letter for each cell: f fg, r border, a accent, e error, x a token
// outside the palette.
func raw(rows ...string) Raw {
	names := map[rune]string{'f': "fg", 'r': "border", 'a': "accent", 'e': "error", 'x': "red"}
	var g Raw
	for _, row := range rows {
		text, toks, _ := strings.Cut(row, "|")
		var cells []RawCell
		tk := []rune(toks)
		for i, c := range []rune(text) {
			cells = append(cells, RawCell{C: string(c), K: names[tk[i]]})
		}
		g.Cells = append(g.Cells, cells)
	}
	g.H = len(g.Cells)
	if g.H > 0 {
		g.W = len(g.Cells[0])
	}
	return g
}

// fill gives a TooSmall-size (60×16) grid of spaces in fg, with the first row
// replaced by top.
func fill(top string, toks string) Raw {
	rows := []string{top + strings.Repeat(" ", 60-len([]rune(top))) + "|" + toks + strings.Repeat("f", 60-len(toks))}
	for range 15 {
		rows = append(rows, strings.Repeat(" ", 60)+"|"+strings.Repeat("f", 60))
	}
	return raw(rows...)
}

func TestCheckAcceptsAGoodBoard(t *testing.T) {
	g := fill("╭─ hive ─╮ ●◐○ ×…", "rrraaaarrrfaaafff")
	if err := Check("TooSmall", g); err != nil {
		t.Fatal(err)
	}
}

func TestCheckAcceptsChatStatusSymbols(t *testing.T) {
	g := fill("▸▾▲", "fff")
	if err := Check("TooSmall", g); err != nil {
		t.Fatal(err)
	}
}

func TestCheckAcceptsNewChatBoardSizes(t *testing.T) {
	for _, board := range []struct {
		name string
		w, h int
	}{
		{"ChatStatus", 120, 40},
		{"ChatEvents", 120, 40},
		{"CompactChatStatus", 80, 24},
	} {
		t.Run(board.name, func(t *testing.T) {
			if err := checkSize(board.name, board.w, board.h); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckAcceptsTheHeavyBorder(t *testing.T) {
	g := fill("┏━┯━┓┃┗┷┛┡┨", "aaaaaaaaaaa")
	if err := Check("TooSmall", g); err != nil {
		t.Fatal(err)
	}
}

func TestCheckAcceptsTheLiveBoards(t *testing.T) {
	g := fill("┰┸┺┹╃┻◀▼", "aaaaaaaa")
	if err := Check("TooSmall", g); err != nil {
		t.Fatal(err)
	}
	for _, board := range []struct {
		name string
		w, h int
	}{
		{"LiveOverview", 120, 40},
		{"LivePod", 120, 40},
		{"CompactLiveOverview", 80, 24},
		{"CompactLivePod", 80, 24},
	} {
		if err := checkSize(board.name, board.w, board.h); err != nil {
			t.Error(err)
		}
	}
}

func TestCheckAcceptsTheTreeBoards(t *testing.T) {
	for _, board := range []struct {
		name string
		w, h int
	}{
		{"TreeWide", 120, 40},
		{"TreeNested", 120, 40},
		{"CompactTreeWide", 80, 24},
		{"CompactTreeNested", 80, 24},
	} {
		if err := checkSize(board.name, board.w, board.h); err != nil {
			t.Error(err)
		}
	}
}

func TestCheckAcceptsTheChatModeBoards(t *testing.T) {
	for _, board := range []struct {
		name string
		w, h int
	}{
		{"ChatNormal", 120, 40},
		{"ChatInput", 120, 40},
		{"CompactChatNormal", 80, 24},
	} {
		if err := checkSize(board.name, board.w, board.h); err != nil {
			t.Error(err)
		}
	}
}

func TestCheckFailsABadSymbol(t *testing.T) {
	g := fill("ab★", "fff")
	err := Check("TooSmall", g)
	assertErr(t, err, "TooSmall", "row 0 col 2", `"★"`)
}

func TestCheckFailsATokenOutsideThePalette(t *testing.T) {
	g := fill("abc", "ffx")
	err := Check("TooSmall", g)
	assertErr(t, err, "TooSmall", "row 0 col 2", `"red"`)
}

func TestCheckFailsACellOfTwoRunes(t *testing.T) {
	g := fill("abc", "fff")
	g.Cells[3][5].C = "ab"
	err := Check("TooSmall", g)
	assertErr(t, err, "TooSmall", "row 3 col 5", `"ab"`)
}

func TestCheckFailsAWrongSize(t *testing.T) {
	g := raw("ab|ff", "cd|ff")
	err := Check("TooSmall", g)
	assertErr(t, err, "TooSmall", "2×2", "60×16")
}

func TestCheckFailsARaggedRow(t *testing.T) {
	g := fill("abc", "fff")
	g.Cells[7] = g.Cells[7][:59]
	err := Check("TooSmall", g)
	assertErr(t, err, "TooSmall", "row 7", "59")
}

func TestCheckFailsABoardOutsideTheTable(t *testing.T) {
	g := fill("abc", "fff")
	err := Check("Extra", g)
	assertErr(t, err, "Extra", "spec table")
}

func TestCheckGridFailsABadSymbol(t *testing.T) {
	g, err := ToGrid(fill("ab", "ff"))
	if err != nil {
		t.Fatal(err)
	}
	g.Cells[60*2+4] = cellgrid.Cell{Rune: '★'}
	assertErr(t, CheckGrid("TooSmall", g), "TooSmall", "row 2 col 4", `"★"`)
}

func TestCheckGridFailsAWrongSize(t *testing.T) {
	g, err := ToGrid(raw("ab|ff"))
	if err != nil {
		t.Fatal(err)
	}
	assertErr(t, CheckGrid("TooSmall", g), "TooSmall", "2×1", "60×16")
}

func TestToGridKeepsEachCell(t *testing.T) {
	r := raw("╭a|ra", "x │|fff")
	r.Cells[1] = r.Cells[1][:2]
	r.Cells[0][1].B = 1
	r.Cells[1][1].K = "dim"
	r = Raw{W: 2, H: 2, Cells: r.Cells}
	g, err := ToGrid(r)
	if err != nil {
		t.Fatal(err)
	}
	want := []cellgrid.Cell{
		{Rune: '╭', Token: cellgrid.Border},
		{Rune: 'a', Token: cellgrid.Accent, Bold: true},
		{Rune: 'x', Token: cellgrid.FG},
		{Rune: ' ', Token: cellgrid.Dim},
	}
	if g.Width != 2 || g.Height != 2 || len(g.Cells) != 4 {
		t.Fatalf("grid %d×%d with %d cells, want 2×2 with 4", g.Width, g.Height, len(g.Cells))
	}
	for i, c := range want {
		if g.Cells[i] != c {
			t.Errorf("cell %d = %+v, want %+v", i, g.Cells[i], c)
		}
	}
}

func assertErr(t *testing.T, err error, parts ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want one with %q", parts)
	}
	for _, p := range parts {
		if !strings.Contains(err.Error(), p) {
			t.Errorf("error %q does not name %q", err, p)
		}
	}
}

func TestCheckAcceptsTheAskBoards(t *testing.T) {
	for _, board := range []struct {
		name string
		w, h int
	}{
		{"ChatAsk", 120, 40},
		{"ChatAskMulti", 120, 40},
		{"ChatAskPreview", 120, 40},
		{"ChatAskStopped", 120, 40},
		{"CompactChatAsk", 80, 24},
		{"ChatAskReview", 120, 40},
		{"ChatAskReviewMulti", 120, 40},
		{"CompactChatAskReview", 80, 24},
	} {
		if err := checkSize(board.name, board.w, board.h); err != nil {
			t.Error(err)
		}
	}
}

// gfFill gives a 120×40 Gf board grid of spaces in fg, with the first row
// replaced by top.
func gfFill(top string, toks string) Raw {
	rows := []string{top + strings.Repeat(" ", 120-len([]rune(top))) + "|" + toks + strings.Repeat("f", 120-len(toks))}
	for range 39 {
		rows = append(rows, strings.Repeat(" ", 120)+"|"+strings.Repeat("f", 120))
	}
	return raw(rows...)
}

// Golden-frames decision 23: each board is 120×40 except GfComponents at
// 120×67.
func TestCheckAcceptsTheGoldenFramesBoardSizes(t *testing.T) {
	boards := []string{
		"GfChat", "GfChatAsk", "GfWork", "GfFriction", "GfHealth", "GfLogs", "GfAgents", "GfHelp",
		"GfChatEmpty", "GfChatNoCoordinator", "GfChatEvents", "GfChatTall", "GfChatMarkdown", "GfChatNotice",
		"GfChatAskInput", "GfChatAskMulti", "GfChatAskPreview", "GfChatAskStopped", "GfChatAskReview", "GfChatAskMany",
		"GfWorkEmpty", "GfWorkNoGoal", "GfWorkScroll", "GfWorkStops", "GfWorkCancelled", "GfWorkTask", "GfWorkDeps",
		"GfFrictionResolved", "GfFrictionEmpty", "GfFrictionFilter", "GfHealthGreen", "GfHealthRestart", "GfHealthChecking",
		"GfLogsEmpty", "GfLogsRaw", "GfLogsError", "GfAgentsPod", "GfAgentsDetails", "GfAgentsNone",
	}
	for _, board := range boards {
		if err := checkSize(board, 120, 40); err != nil {
			t.Error(err)
		}
	}
	if err := checkSize("GfComponents", 120, 67); err != nil {
		t.Error(err)
	}
	assertErr(t, checkSize("GfChat", 120, 67), "GfChat", "120×67", "120×40")
}

// Golden-frames decisions 7 and 21: a Gf board can use the bar symbols and
// the error token.
func TestCheckAcceptsTheGoldenFramesSymbolsAndError(t *testing.T) {
	g := gfFill("▮▁▂▃▄▅▆▇ error: ▸▾", "aaaaaaaafeeeeeefff")
	if err := Check("GfLogs", g); err != nil {
		t.Fatal(err)
	}
}

// A tui-v2 board keeps the symbol list of tui-v2 decision 13 and the eight
// tokens of tui-v2 decision 26.
func TestCheckKeepsTheTuiV2ListsForATuiV2Board(t *testing.T) {
	assertErr(t, Check("TooSmall", fill("ab▮", "fff")), "TooSmall", "row 0 col 2", `"▮"`)
	assertErr(t, Check("TooSmall", fill("abc", "ffe")), "TooSmall", "row 0 col 2", `"error"`)
}

func TestCheckFailsAGoldenFramesBoardOutsideItsLists(t *testing.T) {
	assertErr(t, Check("GfChat", gfFill("ab★", "fff")), "GfChat", "row 0 col 2", `"★"`)
	assertErr(t, Check("GfChat", gfFill("abc", "ffx")), "GfChat", "row 0 col 2", `"red"`)
	assertErr(t, Check("GfChat", fill("abc", "fff")), "GfChat", "60×16", "120×40")
}
