package cellgrid

import (
	"strings"
	"testing"
)

// testTokens maps the grid file letters to tokens by hand, so that a test
// builds an expected grid without the code under test.
var testTokens = map[byte]Token{
	'f': FG, 'r': Border, 'd': Dim, 'a': Accent, 'w': Working,
	'o': Done, 'i': Info, 'u': User, 'c': Code,
}

// handGrid builds a grid from rows of "text|tokens|bold" strings.
func handGrid(t *testing.T, rows ...string) Grid {
	t.Helper()
	var g Grid
	for _, row := range rows {
		parts := strings.Split(row, "|")
		text, tokens, bold := []rune(parts[0]), parts[1], parts[2]
		if len(text) != len(tokens) || len(text) != len(bold) {
			t.Fatalf("bad hand row %q", row)
		}
		g.Width = len(text)
		g.Height++
		for i, r := range text {
			tok, ok := testTokens[tokens[i]]
			if !ok {
				t.Fatalf("bad hand token %q", tokens[i])
			}
			g.Cells = append(g.Cells, Cell{Rune: r, Token: tok, Bold: bold[i] == '1'})
		}
	}
	return g
}

func assertGrid(t *testing.T, got, want Grid) {
	t.Helper()
	if got.Width != want.Width || got.Height != want.Height {
		t.Fatalf("size = %dx%d, want %dx%d", got.Width, got.Height, want.Width, want.Height)
	}
	for i := range want.Cells {
		if got.Cells[i] != want.Cells[i] {
			t.Errorf("cell row %d col %d = %+v, want %+v", i/want.Width, i%want.Width, got.Cells[i], want.Cells[i])
		}
	}
}

const (
	sgrBorder = "\x1b[38;2;61;58;51m"
	sgrAccent = "\x1b[38;2;245;185;66m"
)

func TestConvertBoxWithAccentTitle(t *testing.T) {
	in := sgrBorder + "╭─ " + "\x1b[1m" + sgrAccent + "hive" + "\x1b[0m" + sgrBorder + " ─╮" + "\x1b[0m\n" +
		sgrBorder + "╰────────╯" + "\x1b[0m"
	got, err := Convert(in, 10, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := handGrid(t,
		"╭─ hive ─╮|rrraaaarrr|0001111000",
		"╰────────╯|rrrrrrrrrr|0000000000",
	)
	assertGrid(t, got, want)
}

func TestConvertDefaultForegroundBoldOffResetAndPadding(t *testing.T) {
	// "ab" default fg; "c" bold done; 22 ends bold; 39 back to fg; reset
	// after a bold accent cell. Row 1 is short and row 2 is absent.
	in := "ab\x1b[1;38;2;124;207;106mc\x1b[22md\x1b[39me\x1b[1;38;2;245;185;66mf\x1b[mg\r\n" +
		"\x1b[38;2;111;183;255mx\n"
	got, err := Convert(in, 8, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := handGrid(t,
		"abcdefg |ffoofaff|00100100",
		"x       |ifffffff|00000000",
		"        |ffffffff|00000000",
	)
	assertGrid(t, got, want)
}

func TestConvertMapsEachPaletteHex(t *testing.T) {
	cases := map[string]Token{
		"61;58;51": Border, "163;157;144": Dim, "245;185;66": Accent,
		"110;230;154": Working, "124;207;106": Done, "111;183;255": Info,
		"233;207;155": User, "159;208;255": Code,
	}
	for rgb, want := range cases {
		got, err := Convert("\x1b[38;2;"+rgb+"mX", 1, 1)
		if err != nil {
			t.Fatalf("%s: %v", rgb, err)
		}
		if got.Cells[0].Token != want {
			t.Errorf("%s: token %v, want %v", rgb, got.Cells[0].Token, want)
		}
	}
}

func TestConvertFailures(t *testing.T) {
	cases := []struct {
		name, in string
		want     []string
	}{
		{"indexed", "ab\nc\x1b[38;5;214mX", []string{"row 1 col 1", `"38;5;214"`}},
		{"16-color", "\x1b[31mX", []string{"row 0 col 0", `"31"`}},
		{"bright 16-color", "a\x1b[93mX", []string{"row 0 col 1", `"93"`}},
		{"background", "\x1b[48;2;61;58;51mX", []string{"row 0 col 0", `"48;2;61;58;51"`}},
		{"16-color background", "\x1b[41mX", []string{"row 0 col 0", `"41"`}},
		{"off-palette hex", "\x1b[38;2;1;2;3mX", []string{"row 0 col 0", `"38;2;1;2;3"`, "#010203"}},
		{"fg hex is not default fg", "\x1b[38;2;230;225;214mX", []string{"row 0 col 0", "#e6e1d6"}},
		{"wide rune", "a漢", []string{"row 0 col 1", "width 2"}},
		{"too wide", "abcd", []string{"row 0 col 3"}},
		{"too tall", "a\nb\nc", []string{"3 rows"}},
		{"not SGR", "\x1b[2Ja", []string{"row 0 col 0", "not SGR"}},
	}
	for _, c := range cases {
		_, err := Convert(c.in, 3, 2)
		if err == nil {
			t.Errorf("%s: no error", c.name)
			continue
		}
		for _, w := range c.want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("%s: error %q does not contain %q", c.name, err, w)
			}
		}
	}
}
