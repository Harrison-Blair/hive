package boardexport

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// board wraps a script in the board page shape of the design folder.
func board(scripts ...string) string {
	var b strings.Builder
	b.WriteString("<!doctype html>\n<html><head><script src=\"./support.js\"></script></head><body>\n<x-dc><div>{{rows}}</div></x-dc>\n")
	for _, s := range scripts {
		b.WriteString("<script type=\"text/x-dc\" data-dc-script data-props='{\"$preview\":{\"width\":516}}'>" + s + "</script>\n")
	}
	b.WriteString("</body></html>\n")
	return b.String()
}

// tooSmall is a 60×16 board script. Its rows() output is empty, so only the
// cells of T can give the grid.
const tooSmall = `
function T(w, h) { var g = []; for (var y = 0; y < h; y++) { var r = []; for (var x = 0; x < w; x++) r.push({c:' ', k:'fg', b:0}); g.push(r); } return {w:w, h:h, g:g}; }
function put(t, x, y, s, k, b) { s = String(s); for (var i = 0; i < s.length; i++) t.g[y][x + i] = {c:s[i], k:k || 'fg', b:b ? 1 : 0}; }
class Component extends DCLogic {
  renderVals() {
    var t = T(60, 16);
    put(t, 0, 0, ' hive', 'fg', 1);
    put(t, 2, 1, '●', 'working');
    put(t, 59, 15, 'SYM', 'KEY');
    return { rows: [] };
  }
}
`

func TestScriptTakesTheBoardScript(t *testing.T) {
	got, err := Script(board("var a = 1;"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "var a = 1;" {
		t.Errorf("Script = %q, want %q", got, "var a = 1;")
	}
}

func TestScriptFailsWithoutOneBoardScript(t *testing.T) {
	if _, err := Script(board()); err == nil {
		t.Error("Script of a page with no board script: error = nil")
	}
	if _, err := Script(board("var a;", "var b;")); err == nil {
		t.Error("Script of a page with two board scripts: error = nil")
	}
}

func TestExportDirFailsWithoutNode(t *testing.T) {
	dir := t.TempDir()
	writeBoard(t, dir, "TooSmall", board(strings.Replace(tooSmall, "'SYM', 'KEY'", "'x', 'fg'", 1)))
	t.Setenv("PATH", t.TempDir())
	err := ExportDir(dir, dir)
	assertErr(t, err, "Node.js")
	if _, statErr := os.Stat(filepath.Join(dir, "TooSmall.grid")); statErr == nil {
		t.Error("ExportDir wrote a grid without Node.js")
	}
}

func TestExportDirWritesTheCellsOfT(t *testing.T) {
	needNode(t)
	dir, out := t.TempDir(), t.TempDir()
	writeBoard(t, dir, "TooSmall", board(strings.Replace(tooSmall, "'SYM', 'KEY'", "'x', 'dim', 1", 1)))
	if err := os.WriteFile(filepath.Join(dir, "notes.html"), []byte("not a board"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ExportDir(dir, out); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(out, "TooSmall.grid"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	g, err := cellgrid.ReadGrid(f)
	if err != nil {
		t.Fatal(err)
	}
	if g.Width != 60 || g.Height != 16 {
		t.Fatalf("grid %d×%d, want 60×16", g.Width, g.Height)
	}
	for _, c := range []struct {
		row, col int
		want     cellgrid.Cell
	}{
		{0, 0, cellgrid.Cell{Rune: ' ', Token: cellgrid.FG, Bold: true}},
		{0, 1, cellgrid.Cell{Rune: 'h', Token: cellgrid.FG, Bold: true}},
		{0, 5, cellgrid.Cell{Rune: ' ', Token: cellgrid.FG}},
		{1, 2, cellgrid.Cell{Rune: '●', Token: cellgrid.Working}},
		{15, 59, cellgrid.Cell{Rune: 'x', Token: cellgrid.Dim, Bold: true}},
	} {
		if got := g.At(c.row, c.col); got != c.want {
			t.Errorf("cell row %d col %d = %+v, want %+v", c.row, c.col, got, c.want)
		}
	}
	if entries, _ := os.ReadDir(out); len(entries) != 1 {
		t.Errorf("out has %d files, want 1 grid", len(entries))
	}
}

func TestExportDirFailsABadBoard(t *testing.T) {
	needNode(t)
	cases := []struct {
		name, script string
		parts        []string
	}{
		{"symbol", strings.Replace(tooSmall, "'SYM', 'KEY'", "'★', 'fg'", 1), []string{"TooSmall", "row 15 col 59", `"★"`}},
		{"token", strings.Replace(tooSmall, "'SYM', 'KEY'", "'x', 'red'", 1), []string{"TooSmall", "row 15 col 59", `"red"`}},
		{"size", strings.Replace(strings.Replace(tooSmall, "T(60, 16)", "T(61, 16)", 1), "'SYM', 'KEY'", "'x', 'fg'", 1), []string{"TooSmall", "61×16", "60×16"}},
		{"two grids", strings.Replace(strings.Replace(tooSmall, "'SYM', 'KEY'", "'x', 'fg'", 1), "return { rows: [] };", "T(60, 16); return { rows: [] };", 1), []string{"TooSmall", "2 grids"}},
		{"script error", strings.Replace(tooSmall, "var t = T(60, 16);", "throw new Error('boom');", 1), []string{"TooSmall", "boom"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			writeBoard(t, dir, "TooSmall", board(c.script))
			assertErr(t, ExportDir(dir, dir), c.parts...)
			if _, err := os.Stat(filepath.Join(dir, "TooSmall.grid")); err == nil {
				t.Error("ExportDir wrote the grid of a bad board")
			}
		})
	}
}

func writeBoard(t *testing.T, dir, name, page string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".dc.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
}

// needNode skips a test that runs a board script when Node.js is missing. The
// gate does not need Node.js.
func needNode(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node.js is not on PATH")
	}
}
