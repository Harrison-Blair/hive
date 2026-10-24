package boardexport

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// Sizes is the board table of the tui-v2 spec: the width and height of each
// board, by board name.
var Sizes = map[string][2]int{
	"Main":              {120, 40},
	"AgentsMany":        {120, 40},
	"AgentDetails":      {120, 40},
	"CompactAgents":     {80, 24},
	"Chat":              {120, 40},
	"CompactChat":       {80, 24},
	"ChatStatus":        {120, 40},
	"ChatEvents":        {120, 40},
	"CompactChatStatus": {80, 24},
	"WorkGoals":         {120, 40},
	"WorkQueue":         {120, 40},
	"Work":              {120, 40},
	"Compact":           {80, 24},
	"Friction":          {120, 40},
	"FrictionResolved":  {120, 40},
	"CompactFriction":   {80, 24},
	"Health":            {120, 40},
	"HealthAgents":      {120, 40},
	"CompactHealth":     {80, 24},
	"Logs":              {120, 40},
	"CompactLogs":       {80, 24},
	"Help":              {120, 40},
	"TooSmall":          {60, 16},
	"Components":        {120, 67},
	// The general agent boards of task 46-agents-home-graph.
	"AgentsHome":             {120, 40},
	"AgentsNested":           {120, 40},
	"AgentsHiddenParent":     {120, 40},
	"AgentsCoordinatorChild": {120, 40},
	"CompactAgentsHome":      {80, 24},
	"AgentDetailsGeneral":    {120, 40},
	"HealthAgentsGeneral":    {120, 40},
	// The live agent boards of task 124-live-agent-activity.
	"LiveOverview":        {120, 40},
	"LivePod":             {120, 40},
	"CompactLiveOverview": {80, 24},
	"CompactLivePod":      {80, 24},
	// The tree boards of task 125-pod-tree-scroll.
	"TreeWide":          {120, 40},
	"TreeNested":        {120, 40},
	"CompactTreeWide":   {80, 24},
	"CompactTreeNested": {80, 24},
	// The chat mode boards of task 126-chat-modes-mouse.
	"ChatNormal":        {120, 40},
	"ChatInput":         {120, 40},
	"CompactChatNormal": {80, 24},
	// The ask boards of task 131-hive-ask.
	"ChatAsk":        {120, 40},
	"ChatAskMulti":   {120, 40},
	"ChatAskPreview": {120, 40},
	"ChatAskStopped": {120, 40},
	"CompactChatAsk": {80, 24},
	// The ask review boards of task 135-ask-confirm.
	"ChatAskReview":        {120, 40},
	"ChatAskReviewMulti":   {120, 40},
	"CompactChatAskReview": {80, 24},
}

// GfSizes is the board table of golden-frames decision 23: the width and
// height of each Gf board, by board name.
var GfSizes = func() map[string][2]int {
	m := map[string][2]int{"GfComponents": {120, 67}}
	for _, board := range []string{
		"GfChat", "GfChatAsk", "GfWork", "GfFriction", "GfHealth", "GfLogs", "GfAgents", "GfHelp",
		"GfChatEmpty", "GfChatNoCoordinator", "GfChatEvents", "GfChatTall", "GfChatMarkdown", "GfChatNotice",
		"GfChatAskInput", "GfChatAskMulti", "GfChatAskPreview", "GfChatAskStopped", "GfChatAskReview", "GfChatAskMany",
		"GfWorkEmpty", "GfWorkNoGoal", "GfWorkScroll", "GfWorkStops", "GfWorkCancelled", "GfWorkTask", "GfWorkDeps",
		"GfFrictionResolved", "GfFrictionEmpty", "GfFrictionFilter", "GfHealthGreen", "GfHealthRestart", "GfHealthChecking",
		"GfLogsEmpty", "GfLogsRaw", "GfLogsError", "GfAgentsPod", "GfAgentsDetails", "GfAgentsNone",
	} {
		m[board] = [2]int{120, 40}
	}
	return m
}()

// symbols is the symbol list of tui-v2 decision 13. A cell can also hold a
// printable ASCII character.
const symbols = "─│╭╮╰╯├┤┼┴┬└━┃┏┓┗┛┯┷┡┨●◐○▶✓•·–—…↑↓←→›█×▸▾▲┰┸┺┹╃┻◀▼┟┠┥╅"

// gfSymbols is the symbol list of golden-frames decision 7.
const gfSymbols = symbols + "▮▁▂▃▄▅▆▇"

// spec is the board table, the symbol list, and the last token of one
// feature.
type spec struct {
	sizes   map[string][2]int
	symbols string
	last    cellgrid.Token
	rule    string
}

var specs = []spec{
	{Sizes, symbols, cellgrid.Code, "tui-v2 decision 13"},
	{GfSizes, gfSymbols, cellgrid.Error, "golden-frames decision 7"},
}

// specOf gives the spec of the feature whose table has the board.
func specOf(board string) (spec, error) {
	for _, s := range specs {
		if _, ok := s.sizes[board]; ok {
			return s, nil
		}
	}
	return spec{}, fmt.Errorf("board %s: not in the spec table", board)
}

// tokens maps each token name of the palette to its token.
var tokens = func() map[string]cellgrid.Token {
	m := map[string]cellgrid.Token{}
	for t := cellgrid.FG; t <= cellgrid.Error; t++ {
		m[t.String()] = t
	}
	return m
}()

func checkSize(board string, w, h int) error {
	s, err := specOf(board)
	if err != nil {
		return err
	}
	size := s.sizes[board]
	if w != size[0] || h != size[1] {
		return fmt.Errorf("board %s: size %d×%d, want %d×%d", board, w, h, size[0], size[1])
	}
	return nil
}

func checkSymbol(s spec, board string, row, col int, r rune) error {
	if (r >= ' ' && r <= '~') || strings.ContainsRune(s.symbols, r) {
		return nil
	}
	return fmt.Errorf("board %s: row %d col %d: symbol %q is not in the list of %s", board, row, col, string(r), s.rule)
}

// CheckGrid checks the size and each symbol of a grid of a board.
func CheckGrid(board string, g cellgrid.Grid) error {
	if err := checkSize(board, g.Width, g.Height); err != nil {
		return err
	}
	s, _ := specOf(board)
	for row := range g.Height {
		for col := range g.Width {
			if err := checkSymbol(s, board, row, col, g.At(row, col).Rune); err != nil {
				return err
			}
		}
	}
	return nil
}

// RawCell is one cell of the board script: a string, a token name, and a
// bold flag of 0 or 1.
type RawCell struct {
	C string `json:"c"`
	K string `json:"k"`
	B int    `json:"b"`
}

// Raw is the grid of T in the board script, before the checks.
type Raw struct {
	W     int         `json:"w"`
	H     int         `json:"h"`
	Cells [][]RawCell `json:"g"`
}

// Check checks the size, each symbol, and each token of the grid of a board.
// The error names the board, the cell, and the value.
func Check(board string, r Raw) error {
	if err := checkSize(board, r.W, len(r.Cells)); err != nil {
		return err
	}
	if err := checkSize(board, r.W, r.H); err != nil {
		return err
	}
	s, _ := specOf(board)
	for row, cells := range r.Cells {
		if len(cells) != r.W {
			return fmt.Errorf("board %s: row %d has %d cells, want %d", board, row, len(cells), r.W)
		}
		for col, c := range cells {
			if utf8.RuneCountInString(c.C) != 1 {
				return fmt.Errorf("board %s: row %d col %d: cell %q is not one symbol", board, row, col, c.C)
			}
			r, _ := utf8.DecodeRuneInString(c.C)
			if err := checkSymbol(s, board, row, col, r); err != nil {
				return err
			}
			if tok, ok := tokens[c.K]; !ok || tok > s.last {
				return fmt.Errorf("board %s: row %d col %d: token %q is not in the palette of the board", board, row, col, c.K)
			}
		}
	}
	return nil
}

// ToGrid converts the grid of a board script to a cell grid. Run Check first.
func ToGrid(r Raw) (cellgrid.Grid, error) {
	g := cellgrid.Grid{Width: r.W, Height: r.H, Cells: make([]cellgrid.Cell, 0, r.W*r.H)}
	for row, cells := range r.Cells {
		for col, c := range cells {
			tok, ok := tokens[c.K]
			if !ok {
				return cellgrid.Grid{}, fmt.Errorf("row %d col %d: token %q is not in the palette", row, col, c.K)
			}
			ch, _ := utf8.DecodeRuneInString(c.C)
			g.Cells = append(g.Cells, cellgrid.Cell{Rune: ch, Token: tok, Bold: c.B != 0})
		}
	}
	return g, nil
}
