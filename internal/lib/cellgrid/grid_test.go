package cellgrid

import (
	"bytes"
	"strings"
	"testing"
)

const boxFile = "hive-grid 1 10 2\n" +
	"t|╭─ hive ─╮\n" +
	"k|rrraaaarrr\n" +
	"b|0001111000\n" +
	"t|╰──────╯ x\n" +
	"k|rrrrrrrrfc\n" +
	"b|0000000001\n"

func TestReadGridFile(t *testing.T) {
	got, err := ReadGrid(strings.NewReader(boxFile))
	if err != nil {
		t.Fatal(err)
	}
	want := handGrid(t,
		"╭─ hive ─╮|rrraaaarrr|0001111000",
		"╰──────╯ x|rrrrrrrrfc|0000000001",
	)
	assertGrid(t, got, want)
}

func TestWriteGridFile(t *testing.T) {
	g := handGrid(t,
		"╭─ hive ─╮|rrraaaarrr|0001111000",
		"╰──────╯ x|rrrrrrrrfc|0000000001",
	)
	var buf bytes.Buffer
	if err := WriteGrid(&buf, g); err != nil {
		t.Fatal(err)
	}
	if buf.String() != boxFile {
		t.Errorf("WriteGrid =\n%s\nwant\n%s", buf.String(), boxFile)
	}
}

func TestGridRoundTripEachToken(t *testing.T) {
	g := handGrid(t, "abcdefghi|frdawoiuc|101010101", "·•…—┃━×█●|cuiowadrf|010101010")
	var buf bytes.Buffer
	if err := WriteGrid(&buf, g); err != nil {
		t.Fatal(err)
	}
	got, err := ReadGrid(&buf)
	if err != nil {
		t.Fatal(err)
	}
	assertGrid(t, got, g)
}

func TestReadGridErrorsNameTheLine(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"bad magic", "hive-grid 2 1 1\nt|a\nk|f\nb|0\n", "line 1"},
		{"bad size", "hive-grid 1 x 1\nt|a\nk|f\nb|0\n", "line 1"},
		{"short row", "hive-grid 1 2 1\nt|a\nk|ff\nb|00\n", "line 2"},
		{"long tokens", "hive-grid 1 1 1\nt|a\nk|ff\nb|0\n", "line 3"},
		{"unknown token", "hive-grid 1 1 1\nt|a\nk|z\nb|0\n", "line 3"},
		{"bad bold", "hive-grid 1 1 1\nt|a\nk|f\nb|2\n", "line 4"},
		{"wide rune", "hive-grid 1 2 1\nt|漢\nk|ff\nb|00\n", "line 2"},
		{"bad prefix", "hive-grid 1 1 1\nx|a\nk|f\nb|0\n", "line 2"},
		{"missing row", "hive-grid 1 1 2\nt|a\nk|f\nb|0\n", "line 5"},
		{"extra line", "hive-grid 1 1 1\nt|a\nk|f\nb|0\nt|b\n", "line 5"},
	}
	for _, c := range cases {
		_, err := ReadGrid(strings.NewReader(c.in))
		if err == nil {
			t.Errorf("%s: no error", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %q does not name %q", c.name, err, c.want)
		}
	}
}

func TestWriteGridRejectsWideRune(t *testing.T) {
	g := Grid{Width: 1, Height: 1, Cells: []Cell{{Rune: '漢'}}}
	if err := WriteGrid(&bytes.Buffer{}, g); err == nil {
		t.Error("WriteGrid accepted a wide rune")
	}
}
