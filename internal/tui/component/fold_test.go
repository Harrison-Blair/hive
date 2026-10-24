package component

import "testing"

// Friction layout and rule 3: GfFriction row 19 is the folded line and
// GfFrictionResolved row 13 is the open line. The glyph and the hint are dim,
// the text is dim bold. Column 119 of row 13 can hold the list thumb, so the
// check reads 119 cells.
func TestFoldGolden(t *testing.T) {
	checkGolden(t, Fold{Text: "resolved · 23", Hint: "enter opens"}.View(120), "GfFriction", rect{19, 0, 120, 1})
	checkGolden(t, Fold{Open: true, Text: "resolved · 23", Hint: "enter folds"}.View(119), "GfFrictionResolved", rect{13, 0, 119, 1})
}

// Friction rule 2 on the fold line: the selected line shows ▶ in accent bold
// in place of its glyph, and its text in accent bold. The hint keeps its
// word, so the state shows without colour.
func TestFoldSelected(t *testing.T) {
	want := "hive-grid 1 30 1\n" +
		"t| ▶ resolved · 2  enter folds  \n" +
		"k|fafaaaaaaaaaaaaffdddddddddddff\n" +
		"b|010111111111111000000000000000\n"
	checkHandGrid(t, Fold{Open: true, Selected: true, Text: "resolved · 2", Hint: "enter folds"}.View(30), want)
}
