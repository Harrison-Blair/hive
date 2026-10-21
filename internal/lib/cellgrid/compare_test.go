package cellgrid

import (
	"reflect"
	"strings"
	"testing"
)

func TestCompareReportsEachDifferingCell(t *testing.T) {
	want := handGrid(t,
		"╭─ hive ─╮|rrraaaarrr|0001111000",
		"│ 12:04  │|rddddddffr|0000000000",
	)
	got := handGrid(t,
		"╭─ hivX ─╮|rrraaaarrr|0001110000",
		"│ 12:04  │|rddddddfar|0000000000",
	)
	diffs, err := Compare(want, got, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantDiffs := []Diff{
		{Row: 0, Col: 6, Want: Cell{'e', Accent, true}, Got: Cell{'X', Accent, false}},
		{Row: 1, Col: 8, Want: Cell{' ', FG, false}, Got: Cell{' ', Accent, false}},
	}
	if !reflect.DeepEqual(diffs, wantDiffs) {
		t.Fatalf("diffs = %+v, want %+v", diffs, wantDiffs)
	}
	text := FormatDiffs(diffs)
	for _, line := range []string{
		`row 0 col 6: want 'e' accent bold, got 'X' accent normal`,
		`row 1 col 8: want ' ' fg normal, got ' ' accent normal`,
	} {
		if !strings.Contains(text, line+"\n") {
			t.Errorf("FormatDiffs output\n%s\nhas no line %q", text, line)
		}
	}
}

func TestCompareMaskIgnoresRuneOnly(t *testing.T) {
	want := handGrid(t,
		"│ 12:04  │|rddddddffr|0000000000",
	)
	clockMask := []Mask{{Row: 0, Col: 2, Width: 5}}

	gotRune := handGrid(t, "│ 09:59  │|rddddddffr|0000000000")
	diffs, err := Compare(want, gotRune, clockMask)
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Errorf("masked rune change gave diffs %+v", diffs)
	}

	gotStyle := handGrid(t, "│ 09:59  │|rddadddffr|0000100000")
	diffs, err = Compare(want, gotStyle, clockMask)
	if err != nil {
		t.Fatal(err)
	}
	wantDiffs := []Diff{
		{Row: 0, Col: 3, Want: Cell{'2', Dim, false}, Got: Cell{'9', Accent, false}},
		{Row: 0, Col: 4, Want: Cell{':', Dim, false}, Got: Cell{':', Dim, true}},
	}
	if !reflect.DeepEqual(diffs, wantDiffs) {
		t.Errorf("diffs = %+v, want %+v", diffs, wantDiffs)
	}

	// The same rune change outside the mask is a diff.
	diffs, err = Compare(want, gotRune, []Mask{{Row: 0, Col: 2, Width: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 2 {
		t.Errorf("narrow mask gave %d diffs, want 2: %+v", len(diffs), diffs)
	}
}

func TestCompareRejectsBadMask(t *testing.T) {
	want := handGrid(t,
		"╭─ hive ─╮|rrraaaarrr|0001111000",
		"│ 12:04  │|rddddddffr|0000000000",
	)
	cases := []struct {
		name string
		mask Mask
		msg  string
	}{
		{"border rune", Mask{Row: 1, Col: 0, Width: 3}, "border"},
		{"right border rune", Mask{Row: 1, Col: 7, Width: 3}, "border"},
		{"accent title", Mask{Row: 0, Col: 3, Width: 2}, "accent"},
		{"outside grid", Mask{Row: 2, Col: 0, Width: 1}, "outside"},
		{"past right edge", Mask{Row: 1, Col: 8, Width: 3}, "outside"},
		{"zero width", Mask{Row: 1, Col: 2, Width: 0}, "width"},
	}
	for _, c := range cases {
		_, err := Compare(want, want, []Mask{c.mask})
		if err == nil {
			t.Errorf("%s: no error", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.msg) {
			t.Errorf("%s: error %q does not contain %q", c.name, err, c.msg)
		}
	}
}

func TestCompareRejectsSizeMismatch(t *testing.T) {
	a := handGrid(t, "ab|ff|00")
	b := handGrid(t, "abc|fff|000")
	if _, err := Compare(a, b, nil); err == nil {
		t.Error("no error for a size mismatch")
	}
}

func TestReadMask(t *testing.T) {
	in := "# masks for the clock\n\n0 108 5 # clock\n  3 2 10\n"
	got, err := ReadMask(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := []Mask{{Row: 0, Col: 108, Width: 5, Reason: "clock"}, {Row: 3, Col: 2, Width: 10}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadMask = %+v, want %+v", got, want)
	}
}

func TestReadMaskErrorsNameTheLine(t *testing.T) {
	for _, in := range []string{
		"# ok\n0 1\n",
		"# ok\n0 1 x\n",
		"# ok\n0 1 2 3\n",
		"# ok\n-1 1 2\n",
	} {
		_, err := ReadMask(strings.NewReader(in))
		if err == nil || !strings.Contains(err.Error(), "line 2") {
			t.Errorf("ReadMask(%q) error = %v, want one that names line 2", in, err)
		}
	}
}
