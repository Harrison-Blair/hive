package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// TestTitleRowNoteAndLine checks the two title row forms of the Work board
// (work.md rule 11 and the layout): the note in accent two cells after the
// facts, as GfWorkStops row 0 shows, and a line in place of the rule.
func TestTitleRowNoteAndLine(t *testing.T) {
	row := TitleRow{Name: "work · g2", Facts: "· The goal", Note: "· stops only", Right: "[ ] goal 1 of 1", Line: dimStyle.Render("wave A of 1")}
	got := row.View(60)
	want := " work · g2 · The goal  · stops only         [ ] goal 1 of 1 \n wave A of 1                                                "
	if plain := ansi.Strip(got); plain != want {
		t.Fatalf("View:\n%q\nwant:\n%q", plain, want)
	}
	grid, err := cellgrid.Convert(got, 60, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		row, col int
		token    cellgrid.Token
	}{{0, 23, cellgrid.Accent}, {0, 21, cellgrid.Dim}, {1, 1, cellgrid.Dim}} {
		if cell := grid.At(check.row, check.col); cell.Token != check.token {
			t.Errorf("row %d column %d token %v, want %v", check.row, check.col, cell.Token, check.token)
		}
	}
}
