package cellgrid

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
)

// Cell is one terminal cell: a rune of width 1, a foreground token, and bold.
type Cell struct {
	Rune  rune
	Token Token
	Bold  bool
}

// Grid is a Width×Height block of cells in row order.
type Grid struct {
	Width  int
	Height int
	Cells  []Cell
}

// At returns the cell at a zero-based row and column.
func (g Grid) At(row, col int) Cell { return g.Cells[row*g.Width+col] }

// gridMagic starts line 1 of a grid file: "hive-grid 1 W H".
const gridMagic = "hive-grid 1"

// tokenOf gives the token of a grid file letter.
func tokenOf(letter byte) (Token, bool) {
	for t := range palette {
		if palette[t].letter == letter {
			return Token(t), true
		}
	}
	return 0, false
}

// WriteGrid writes g in the grid file format. Line 1 is "hive-grid 1 W H".
// Each row then has three lines: "t|" and its runes, "k|" and its token
// letters, and "b|" and a 0 or 1 for each bold flag.
func WriteGrid(w io.Writer, g Grid) error {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %d %d\n", gridMagic, g.Width, g.Height)
	for row := 0; row < g.Height; row++ {
		cells := g.Cells[row*g.Width : (row+1)*g.Width]
		var text, toks, bold strings.Builder
		for col, c := range cells {
			if lipgloss.Width(string(c.Rune)) != 1 || c.Rune == '\n' {
				return fmt.Errorf("row %d col %d: rune %q is not one cell wide", row, col, c.Rune)
			}
			text.WriteRune(c.Rune)
			toks.WriteByte(palette[c.Token].letter)
			if c.Bold {
				bold.WriteByte('1')
			} else {
				bold.WriteByte('0')
			}
		}
		fmt.Fprintf(&b, "t|%s\nk|%s\nb|%s\n", text.String(), toks.String(), bold.String())
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// ReadGrid reads a grid file. Each error names the line number.
func ReadGrid(r io.Reader) (Grid, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	n := 0
	next := func() (string, bool) {
		if !sc.Scan() {
			return "", false
		}
		n++
		return sc.Text(), true
	}
	head, ok := next()
	if !ok {
		return Grid{}, errors.Join(fmt.Errorf("line 1: missing header"), sc.Err())
	}
	var g Grid
	rest, found := strings.CutPrefix(head, gridMagic+" ")
	if !found {
		return Grid{}, fmt.Errorf("line 1: header %q does not start with %q", head, gridMagic)
	}
	if _, err := fmt.Sscanf(rest, "%d %d", &g.Width, &g.Height); err != nil || g.Width < 1 || g.Height < 1 ||
		fmt.Sprintf("%d %d", g.Width, g.Height) != rest {
		return Grid{}, fmt.Errorf("line 1: header %q needs a positive width and height", head)
	}
	g.Cells = make([]Cell, 0, g.Width*g.Height)
	for row := 0; row < g.Height; row++ {
		var parts [3][]rune
		for i, prefix := range []string{"t|", "k|", "b|"} {
			line, ok := next()
			if !ok {
				return Grid{}, errors.Join(fmt.Errorf("line %d: missing %q line of row %d", n+1, prefix, row), sc.Err())
			}
			body, found := strings.CutPrefix(line, prefix)
			if !found {
				return Grid{}, fmt.Errorf("line %d: row %d line does not start with %q", n, row, prefix)
			}
			parts[i] = []rune(body)
			if len(parts[i]) != g.Width {
				return Grid{}, fmt.Errorf("line %d: row %d has %d cells, want %d", n, row, len(parts[i]), g.Width)
			}
			for col, c := range parts[i] {
				switch i {
				case 0:
					if lipgloss.Width(string(c)) != 1 {
						return Grid{}, fmt.Errorf("line %d: row %d col %d: rune %q is not one cell wide", n, row, col, c)
					}
				case 1:
					if _, ok := tokenOf(byte(c)); !ok || c > 0x7f {
						return Grid{}, fmt.Errorf("line %d: row %d col %d: unknown token letter %q", n, row, col, c)
					}
				case 2:
					if c != '0' && c != '1' {
						return Grid{}, fmt.Errorf("line %d: row %d col %d: bold %q is not 0 or 1", n, row, col, c)
					}
				}
			}
		}
		for col := range g.Width {
			tok, _ := tokenOf(byte(parts[1][col]))
			g.Cells = append(g.Cells, Cell{Rune: parts[0][col], Token: tok, Bold: parts[2][col] == '1'})
		}
	}
	if _, ok := next(); ok {
		return Grid{}, fmt.Errorf("line %d: extra line after %d rows", n, g.Height)
	}
	if err := sc.Err(); err != nil {
		return Grid{}, fmt.Errorf("line %d: %w", n+1, err)
	}
	return g, nil
}
