package cellgrid

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Mask is a run of Width cells from a zero-based row and column. Compare
// ignores the rune of a masked cell but still compares its token and bold.
type Mask struct {
	Row    int
	Col    int
	Width  int
	Reason string
}

// Diff is one cell that differs between the expected and the actual grid.
type Diff struct {
	Row  int
	Col  int
	Want Cell
	Got  Cell
}

// Compare gives each cell of got that differs from want. A mask that is
// outside the grid, or that touches a box-drawing rune or an accent cell of
// want, fails the compare: a border, a title, or the region tag must never
// hide behind a mask.
func Compare(want, got Grid, masks []Mask) ([]Diff, error) {
	if want.Width != got.Width || want.Height != got.Height {
		return nil, fmt.Errorf("size %dx%d differs from expected %dx%d", got.Width, got.Height, want.Width, want.Height)
	}
	masked := make([]bool, len(want.Cells))
	for _, m := range masks {
		if m.Width < 1 {
			return nil, fmt.Errorf("mask %d %d %d: width is not positive", m.Row, m.Col, m.Width)
		}
		if m.Row < 0 || m.Row >= want.Height || m.Col < 0 || m.Col+m.Width > want.Width {
			return nil, fmt.Errorf("mask %d %d %d: outside the %dx%d grid", m.Row, m.Col, m.Width, want.Width, want.Height)
		}
		for col := m.Col; col < m.Col+m.Width; col++ {
			c := want.At(m.Row, col)
			if isBoxDrawing(c.Rune) {
				return nil, fmt.Errorf("mask %d %d %d: col %d has border rune %q", m.Row, m.Col, m.Width, col, c.Rune)
			}
			if c.Token == Accent {
				return nil, fmt.Errorf("mask %d %d %d: col %d has token accent", m.Row, m.Col, m.Width, col)
			}
			masked[m.Row*want.Width+col] = true
		}
	}
	var diffs []Diff
	for i, w := range want.Cells {
		g := got.Cells[i]
		if masked[i] {
			g.Rune = w.Rune
		}
		if g != w {
			diffs = append(diffs, Diff{Row: i / want.Width, Col: i % want.Width, Want: w, Got: got.Cells[i]})
		}
	}
	return diffs, nil
}

// isBoxDrawing reports a rune of the Unicode box-drawing block, which holds
// each border and scroll mark symbol of decision 13.
func isBoxDrawing(r rune) bool { return r >= 0x2500 && r <= 0x257f }

// FormatDiffs prints one line for each diff, for example
// "row 0 col 6: want 'e' accent bold, got 'X' accent normal".
func FormatDiffs(diffs []Diff) string {
	var b strings.Builder
	for _, d := range diffs {
		fmt.Fprintf(&b, "row %d col %d: want %s, got %s\n", d.Row, d.Col, formatCell(d.Want), formatCell(d.Got))
	}
	return b.String()
}

func formatCell(c Cell) string {
	weight := "normal"
	if c.Bold {
		weight = "bold"
	}
	return fmt.Sprintf("%q %s %s", c.Rune, c.Token, weight)
}

// ReadMask reads a mask file. Each line is "ROW COL WIDTH", zero-based, with
// an optional "# reason". A blank line or a "#" line is a comment. Each error
// names the line number.
func ReadMask(r io.Reader) ([]Mask, error) {
	var masks []Mask
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line, reason, _ := strings.Cut(sc.Text(), "#")
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 3 {
			return nil, fmt.Errorf("line %d: want ROW COL WIDTH, got %q", n, sc.Text())
		}
		var v [3]int
		for i, f := range fields {
			x, err := strconv.Atoi(f)
			if err != nil || x < 0 {
				return nil, fmt.Errorf("line %d: %q is not a number of 0 or more", n, f)
			}
			v[i] = x
		}
		masks = append(masks, Mask{Row: v[0], Col: v[1], Width: v[2], Reason: strings.TrimSpace(reason)})
	}
	return masks, sc.Err()
}
