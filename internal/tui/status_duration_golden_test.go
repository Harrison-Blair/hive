package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// durationSpan is one literal approved difference between an old board and
// the task 183 duration text. Row -1 applies to each row that holds old.
type durationSpan struct {
	row      int
	old, new string
}

// durationDrops gives the status-row segment that has no room after the
// longer duration. The row has room for the kept segments only.
var durationDrops = map[string]string{
	"GfAgents":         " · h stopped",
	"GfChat":           " · ^g agents  ^t work  ^f friction  ^y health  ^l logs",
	"GfChatEmpty":      " · ^g agents  ^t work  ^f friction  ^y health  ^l logs",
	"GfChatEvents":     " · ^g agents  ^t work  ^f friction  ^y health  ^l logs",
	"GfChatNotice":     " · ^g agents  ^t work  ^f friction  ^y health  ^l logs",
	"GfChatTall":       " · ^g agents  ^t work  ^f friction  ^y health  ^l logs",
	"GfHelp":           " · ^g agents  ^t work  ^f friction  ^y health  ^l logs",
	"GfChatAsk":        " · esc input",
	"GfChatAskMany":    " · ^r input",
	"GfChatAskPreview": " · esc input",
	"GfChatAskStopped": " · ^r input",
}

// durationSpans records the approved task 183 duration differences. The
// board bytes stay unchanged. Each value is the literal old text and the
// literal new text of the decision, not a value computed from the renderer.
var durationSpans = map[string][]durationSpan{
	"GfAgents": {
		{5, "claude · working · …", "working · 00:02:00"},
		{12, "claude · working · …", "working · 00:01:00"},
		{12, "codex · working · 2m", "working · 00:02:00"},
		{12, "pi · working · 2m", "working · 00:02:00"},
		{17, "pi · working · 3m", "working · 00:03:00"},
		{22, "pi · working · 1m", "working · 00:01:00"},
		{22, "pi · idle · 6m", "pi · idle · 00:06:00"},
		{39, "working 2m", "working 00:02:00"},
	},
	"GfAgentsPod": {
		{5, "claude · working · …", "working · 00:02:00"},
		{12, "claude · working · …", "working · 00:01:00"},
		{19, "pi · working · 3m", "working · 00:03:00"},
		{19, "pi · working · 1m", "working · 00:01:00"},
		{39, "working 2m", "working 00:02:00"},
	},
	"GfAgentsDetails": {
		{5, "claude · working · …", "working · 00:02:02"},
		{12, "claude · working · …", "working · 00:01:02"},
		// The plan-review card is behind the popup. Only its visible suffix
		// changes when the card elides the harness.
		{12, "working · 2m", "ng · 00:02:02"},
		{14, "working · 1m", "working · 00:01:00"},
		{17, "pi · working · 3m", "working · 00:03:02"},
		{22, "pi · working · 1m", "working · 00:01:00"},
		{39, "working 2m", "working 00:02:02"},
	},
	"GfAgentsNone": {
		{5, "claude · working · …", "working · 00:02:00"},
		{39, "working 2m", "working 00:02:00"},
	},
	"GfChat":             {{-1, "working 2m", "working 00:02:00"}},
	"GfChatEmpty":        {{-1, "working 2m", "working 00:02:00"}},
	"GfChatEvents":       {{-1, "working 2m", "working 00:02:00"}},
	"GfChatNotice":       {{-1, "working 2m", "working 00:02:00"}},
	"GfChatTall":         {{-1, "working 2m", "working 00:02:00"}},
	"GfChatAsk":          {{-1, "working 1m", "working 00:01:00"}},
	"GfChatAskInput":     {{-1, "working 1m", "working 00:01:00"}},
	"GfChatAskMany":      {{-1, "working 1m", "working 00:01:00"}},
	"GfChatAskMulti":     {{-1, "working 1m", "working 00:01:00"}},
	"GfChatAskPreview":   {{-1, "working 1m", "working 00:01:00"}},
	"GfChatAskReview":    {{-1, "working 1m", "working 00:01:00"}},
	"GfChatAskStopped":   {{-1, "working 1m", "working 00:01:00"}},
	"GfFriction":         {{-1, "working 2m", "working 00:02:00"}},
	"GfFrictionEmpty":    {{-1, "working 2m", "working 00:02:00"}},
	"GfFrictionFilter":   {{-1, "working 2m", "working 00:02:00"}},
	"GfFrictionResolved": {{-1, "working 2m", "working 00:02:00"}},
	"GfHelp":             {{-1, "working 2m", "working 00:02:00"}},
	"GfLogs":             {{-1, "working 2m", "working 00:02:00"}},
	"GfLogsEmpty":        {{-1, "working 2m", "working 00:02:00"}},
	"GfLogsError":        {{-1, "working 2m", "working 00:02:00"}},
	"GfLogsRaw":          {{-1, "working 2m", "working 00:02:00"}},
	"GfWork": {
		{17, "t12-worker-1 · 12m  ", "t12-work… · 00:12:00"},
		{39, "working 2m", "working 00:02:00"},
	},
	"GfWorkCancelled": {
		{17, "t12-worker-1 · 12m  ", "t12-work… · 00:12:00"},
		{39, "working 2m", "working 00:02:00"},
	},
	"GfWorkDeps":   {{-1, "working 2m", "working 00:02:00"}},
	"GfWorkEmpty":  {{-1, "working 2m", "working 00:02:00"}},
	"GfWorkNoGoal": {{-1, "working 2m", "working 00:02:00"}},
	"GfWorkScroll": {
		{22, "t34-worker-1 · 8m   ", "t34-work… · 00:08:00"},
		{27, "t35-worker-1 · 2m   ", "t35-work… · 00:02:00"},
		{39, "working 2m", "working 00:02:00"},
	},
	"GfWorkStops": {{-1, "working 2m", "working 00:02:00"}},
	"GfWorkTask": {
		{8, "t12-worker-1 · working · 12m", "t12-worker-1 · working · 00:12:00"},
		{39, "working 2m", "working 00:02:00"},
	},
}

// healthDurationSpans gives the table state column 3 more cells for the
// complete time. The old field has 14 cells and the new one has 17, so PANE,
// STARTED, and RESTART move right by 3 and the gap before the right border
// loses 3 cells.
var healthDurationSpans = []durationSpan{
	{-1, "HARNESS   STATE          PANE", "HARNESS   STATE             PANE"},
	{-1, "working 2m" + strings.Repeat(" ", 5), "working 00:02:05  "},
	{-1, "live 62m" + strings.Repeat(" ", 7), "live 01:02:05" + strings.Repeat(" ", 5)},
	{-1, "working 3m" + strings.Repeat(" ", 5), "working 00:03:05  "},
	{-1, "working 1m" + strings.Repeat(" ", 5), "working 00:01:05  "},
	{-1, "idle 6m" + strings.Repeat(" ", 8), "idle 00:06:05" + strings.Repeat(" ", 5)},
	{-1, "unresolved     ", "unresolved        "},
	{-1, "stopped        ", "stopped           "},
	{-1, "working 2m", "working 00:02:05"},
}

// optionalDurationSpans names the health rows that some Health boards do
// not show.
var optionalDurationSpans = map[string]bool{
	"unresolved     ": true,
	"stopped        ": true,
}

func init() {
	for _, board := range []string{"GfHealth", "GfHealthGreen", "GfHealthRestart", "GfHealthChecking"} {
		durationSpans[board] = healthDurationSpans
	}
}

// durationGrid gives the expected grid with the approved task 183 duration
// text. It keeps each board byte unchanged and uses literal spans.
func durationGrid(t *testing.T, board string, grid cellgrid.Grid) cellgrid.Grid {
	t.Helper()
	for _, span := range durationSpans[board] {
		if span.row >= 0 {
			grid = replaceGridSpan(t, board, grid, span.row, span, durationDrops[board])
			continue
		}
		found := false
		for row := 0; row < grid.Height; row++ {
			if strings.Contains(gridText(grid, row), span.old) {
				grid = replaceGridSpan(t, board, grid, row, span, durationDrops[board])
				found = true
			}
		}
		if !found {
			if optionalDurationSpans[span.old] {
				continue
			}
			t.Fatalf("%s: no row holds %q", board, span.old)
		}
	}
	return grid
}

// replaceGridSpan changes one literal span of a row and keeps the row at its
// width. The next border, or the clock, stays in its column: the change takes
// space from the gap before it. A row with neither keeps its left alignment.
// A drop span removes the trailing segment that the status row no longer has
// room for.
func replaceGridSpan(t *testing.T, board string, grid cellgrid.Grid, row int, span durationSpan, drop string) cellgrid.Grid {
	t.Helper()
	text := gridText(grid, row)
	if count := strings.Count(text, span.old); count != 1 {
		t.Fatalf("%s row %d: %q occurs %d times, want 1: %q", board, row, span.old, count, text)
	}
	base := row * grid.Width
	cells := make([]cellgrid.Cell, grid.Width)
	copy(cells, grid.Cells[base:base+grid.Width])
	dropped := false
	if drop != "" {
		if dropByte := strings.Index(text, drop); dropByte >= 0 {
			dropStart := utf8.RuneCountInString(text[:dropByte])
			dropRunes := []rune(drop)
			cells = append(cells[:dropStart], cells[dropStart+len(dropRunes):]...)
			text = gridCellsText(cells)
			dropped = true
		}
	}
	startByte := strings.Index(text, span.old)
	if startByte < 0 {
		t.Fatalf("%s row %d: no %q after the drop: %q", board, row, span.old, text)
	}
	start := utf8.RuneCountInString(text[:startByte])
	oldRunes := []rune(span.old)
	newRunes := []rune(span.new)
	end := start + len(oldRunes)
	anchor := -1
	for col := end; col < len(cells); col++ {
		if gridBorder(cells[col].Rune) {
			anchor = col
			break
		}
	}
	if anchor < 0 {
		if clock := gridClockStart(cells); clock >= end {
			anchor = clock
		}
	}
	token, bold := cells[start].Token, cells[start].Bold
	out := append([]cellgrid.Cell(nil), cells[:start]...)
	for _, r := range newRunes {
		out = append(out, cellgrid.Cell{Rune: r, Token: token, Bold: bold})
	}
	out = append(out, cells[end:]...)
	moved := -1
	if anchor >= 0 {
		moved = anchor + len(newRunes) - len(oldRunes)
		switch delta := len(newRunes) - len(oldRunes); {
		case dropped:
		case delta > 0:
			for range delta {
				if moved <= 0 || out[moved-1].Rune != ' ' {
					t.Fatalf("%s row %d: no gap before column %d for %q", board, row, anchor, span.new)
				}
				out = append(out[:moved-1], out[moved:]...)
				moved--
			}
		case delta < 0:
			spaces := make([]cellgrid.Cell, -delta)
			padToken := cells[anchor].Token
			if anchor > 0 {
				padToken = cells[anchor-1].Token
			}
			for index := range spaces {
				spaces[index] = cellgrid.Cell{Rune: ' ', Token: padToken}
			}
			out = append(out[:moved], append(spaces, out[moved:]...)...)
		}
	}
	if len(out) < grid.Width {
		pad := grid.Width - len(out)
		spaces := make([]cellgrid.Cell, pad)
		padToken := cellgrid.FG
		at := len(out)
		if moved >= 0 {
			at = moved
			if at > 0 {
				padToken = out[at-1].Token
			}
		}
		for index := range spaces {
			spaces[index] = cellgrid.Cell{Rune: ' ', Token: padToken}
		}
		out = append(out[:at], append(spaces, out[at:]...)...)
	}
	if len(out) > grid.Width {
		out = out[:grid.Width]
	}
	copy(grid.Cells[base:base+grid.Width], out)
	return grid
}

// gridCellsText gives the runes of cells as a string.
func gridCellsText(cells []cellgrid.Cell) string {
	var out strings.Builder
	for _, cell := range cells {
		out.WriteRune(cell.Rune)
	}
	return out.String()
}

// gridBorder reports a box border rune that can hold the row width.
func gridBorder(r rune) bool {
	return strings.ContainsRune("│┃╮╯╗┓╰╭", r)
}

// gridClockStart gives the first cell of the row clock, or -1.
func gridClockStart(cells []cellgrid.Cell) int {
	end := len(cells) - 1
	for end >= 0 && cells[end].Rune == ' ' {
		end--
	}
	start := end
	for start > 0 && cells[start-1].Rune != ' ' {
		start--
	}
	if end-start+1 != 5 || cells[start+2].Rune != ':' {
		return -1
	}
	for _, col := range []int{start, start + 1, start + 3, start + 4} {
		if cells[col].Rune < '0' || cells[col].Rune > '9' {
			return -1
		}
	}
	return start
}

func gridText(grid cellgrid.Grid, row int) string {
	var out strings.Builder
	for col := 0; col < grid.Width; col++ {
		out.WriteRune(grid.Cells[row*grid.Width+col].Rune)
	}
	return out.String()
}
