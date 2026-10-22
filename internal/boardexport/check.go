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
	"Main":             {120, 40},
	"AgentsMany":       {120, 40},
	"AgentDetails":     {120, 40},
	"CompactAgents":    {80, 24},
	"Chat":             {120, 40},
	"CompactChat":      {80, 24},
	"WorkGoals":        {120, 40},
	"WorkQueue":        {120, 40},
	"Work":             {120, 40},
	"Compact":          {80, 24},
	"Friction":         {120, 40},
	"FrictionResolved": {120, 40},
	"CompactFriction":  {80, 24},
	"Health":           {120, 40},
	"HealthAgents":     {120, 40},
	"CompactHealth":    {80, 24},
	"Logs":             {120, 40},
	"CompactLogs":      {80, 24},
	"Help":             {120, 40},
	"TooSmall":         {60, 16},
	"Components":       {120, 67},
}

// symbols is the symbol list of decision 13. A cell can also hold a
// printable ASCII character.
const symbols = "─│╭╮╰╯├┤┼┴┬└━┃┏┓┗┛┯┷┡┨●◐○▶✓•·–—…↑↓←→›█×"

// tokens maps each token name of decision 26 to its token.
var tokens = func() map[string]cellgrid.Token {
	m := map[string]cellgrid.Token{}
	for t := cellgrid.FG; t <= cellgrid.Code; t++ {
		m[t.String()] = t
	}
	return m
}()

func checkSize(board string, w, h int) error {
	size, ok := Sizes[board]
	if !ok {
		return fmt.Errorf("board %s: not in the spec table", board)
	}
	if w != size[0] || h != size[1] {
		return fmt.Errorf("board %s: size %d×%d, want %d×%d", board, w, h, size[0], size[1])
	}
	return nil
}

func checkSymbol(board string, row, col int, r rune) error {
	if (r >= ' ' && r <= '~') || strings.ContainsRune(symbols, r) {
		return nil
	}
	return fmt.Errorf("board %s: row %d col %d: symbol %q is not in the list of decision 13", board, row, col, string(r))
}

// CheckGrid checks the size and each symbol of a grid of a board.
func CheckGrid(board string, g cellgrid.Grid) error {
	if err := checkSize(board, g.Width, g.Height); err != nil {
		return err
	}
	for row := range g.Height {
		for col := range g.Width {
			if err := checkSymbol(board, row, col, g.At(row, col).Rune); err != nil {
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
	for row, cells := range r.Cells {
		if len(cells) != r.W {
			return fmt.Errorf("board %s: row %d has %d cells, want %d", board, row, len(cells), r.W)
		}
		for col, c := range cells {
			if utf8.RuneCountInString(c.C) != 1 {
				return fmt.Errorf("board %s: row %d col %d: cell %q is not one symbol", board, row, col, c.C)
			}
			r, _ := utf8.DecodeRuneInString(c.C)
			if err := checkSymbol(board, row, col, r); err != nil {
				return err
			}
			if _, ok := tokens[c.K]; !ok {
				return fmt.Errorf("board %s: row %d col %d: token %q is not in decision 26", board, row, col, c.K)
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
				return cellgrid.Grid{}, fmt.Errorf("row %d col %d: token %q is not in decision 26", row, col, c.K)
			}
			ch, _ := utf8.DecodeRuneInString(c.C)
			g.Cells = append(g.Cells, cellgrid.Cell{Rune: ch, Token: tok, Bold: c.B != 0})
		}
	}
	return g, nil
}
