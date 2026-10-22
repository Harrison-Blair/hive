package cellgrid

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
)

// Convert turns ANSI text into a width×height grid. Each line of the text is
// one row. A cell with no character is a space with token fg. The converter
// reads only SGR sequences with reset, bold on and off, a 24-bit foreground of
// a palette token, and the default foreground. Each other code, each other
// escape sequence, a rune that is not one cell wide, and text outside the size
// fail with an error that names the row and the column.
func Convert(text string, width, height int) (Grid, error) {
	g := Grid{Width: width, Height: height, Cells: make([]Cell, width*height)}
	for i := range g.Cells {
		g.Cells[i].Rune = ' '
	}
	lines := strings.Split(text, "\n")
	if len(lines) > 1 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > height {
		return Grid{}, fmt.Errorf("row %d: text has %d rows, grid height is %d", height, len(lines), height)
	}
	var tok Token
	var bold bool
	for row, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		col := 0
		rs := []rune(line)
		for i := 0; i < len(rs); i++ {
			r := rs[i]
			if r == '\x1b' {
				end, err := readSGR(rs, i, row, col, &tok, &bold)
				if err != nil {
					return Grid{}, err
				}
				i = end
				continue
			}
			if w := lipgloss.Width(string(r)); w != 1 {
				return Grid{}, fmt.Errorf("row %d col %d: rune %q has width %d, want 1", row, col, r, w)
			}
			if col >= width {
				return Grid{}, fmt.Errorf("row %d col %d: text is wider than grid width %d", row, col, width)
			}
			g.Cells[row*width+col] = Cell{Rune: r, Token: tok, Bold: bold}
			col++
		}
	}
	return g, nil
}

// readSGR reads the escape sequence at rs[start], applies it to tok and
// bold, and returns the index of its last rune.
func readSGR(rs []rune, start, row, col int, tok *Token, bold *bool) (int, error) {
	if start+1 >= len(rs) || rs[start+1] != '[' {
		return 0, fmt.Errorf("row %d col %d: escape sequence is not CSI", row, col)
	}
	end := start + 2
	for end < len(rs) && (rs[end] >= '0' && rs[end] <= '9' || rs[end] == ';') {
		end++
	}
	if end >= len(rs) || rs[end] != 'm' {
		return 0, fmt.Errorf("row %d col %d: escape sequence %q is not SGR", row, col, string(rs[start+1:min(end+1, len(rs))]))
	}
	body := string(rs[start+2 : end])
	params := strings.Split(body, ";")
	for p := 0; p < len(params); p++ {
		switch params[p] {
		case "", "0":
			*tok, *bold = FG, false
		case "1":
			*bold = true
		case "22":
			*bold = false
		case "39":
			*tok = FG
		case "38":
			if p+1 < len(params) && params[p+1] == "2" && p+4 < len(params) {
				t, err := hexToken(params[p+2 : p+5])
				if err != nil {
					return 0, fmt.Errorf("row %d col %d: code %q: %w", row, col, body, err)
				}
				*tok = t
				p += 4
				continue
			}
			return 0, fmt.Errorf("row %d col %d: code %q: foreground is not 24-bit", row, col, body)
		default:
			return 0, fmt.Errorf("row %d col %d: code %q: unsupported SGR parameter %s", row, col, body, params[p])
		}
	}
	return end, nil
}

// hexToken gives the token of an r, g, b triple. The fg hex never matches:
// only the default foreground is fg.
func hexToken(rgb []string) (Token, error) {
	var v [3]int
	for i, s := range rgb {
		n, err := strconv.Atoi(s)
		if err != nil || n > 255 {
			return 0, fmt.Errorf("bad color component %q", s)
		}
		v[i] = n
	}
	hex := fmt.Sprintf("#%02x%02x%02x", v[0], v[1], v[2])
	for t := Border; int(t) < len(palette); t++ {
		if palette[t].hex == hex {
			return t, nil
		}
	}
	return 0, fmt.Errorf("hex %s is not a palette token", hex)
}
