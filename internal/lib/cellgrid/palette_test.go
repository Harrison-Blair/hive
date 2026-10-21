package cellgrid

import "testing"

// TestPaletteMatchesDecision26 checks the table against the hex values that
// decision 26 of the tui-v2 spec gives, plus the fg value for a board render.
func TestPaletteMatchesDecision26(t *testing.T) {
	want := []struct {
		tok  Token
		name string
		hex  string
	}{
		{FG, "fg", "#e6e1d6"},
		{Border, "border", "#3d3a33"},
		{Dim, "dim", "#a39d90"},
		{Accent, "accent", "#f5b942"},
		{Working, "working", "#6ee69a"},
		{Done, "done", "#7ccf6a"},
		{Info, "info", "#6fb7ff"},
		{User, "user", "#e9cf9b"},
		{Code, "code", "#9fd0ff"},
	}
	if len(palette) != len(want) {
		t.Fatalf("palette has %d tokens, want %d", len(palette), len(want))
	}
	for _, w := range want {
		if w.tok.String() != w.name || w.tok.Hex() != w.hex {
			t.Errorf("token %d = %s %s, want %s %s", w.tok, w.tok, w.tok.Hex(), w.name, w.hex)
		}
	}
}
