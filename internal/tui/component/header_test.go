package component

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/charmbracelet/x/ansi"
)

var sampleHeader = Header{Home: "~/source/hive", Coordinator: "working", Supervisor: "live", NeedsYou: 0, Clock: "14:22"}

// TestHeaderGolden renders the Header section, Components.dc.html lines 36-38,
// for the long form, and the short form of CompactAgents.dc.html row 0.
func TestHeaderGolden(t *testing.T) {
	t.Run("long", func(t *testing.T) { checkGolden(t, sampleHeader.View(120), sheet, rect{7, 0, 120, 1}) })
	t.Run("short", func(t *testing.T) { checkGolden(t, sampleHeader.View(80), "CompactAgents", rect{0, 0, 80, 1}) })
}

func TestHeaderForms(t *testing.T) {
	cases := []struct {
		width int
		want  string
	}{
		{119, " hive" + strings.Repeat(" ", 70) + "coord ● · sup ● · ? 0 · needs you 0 · 14:22 "},
		{130, " hive  ~/source/hive" + strings.Repeat(" ", 35) + "coordinator ● working   supervisor ● live   ? 0 asks   needs you 0   14:22 "},
	}
	for _, tc := range cases {
		if got := ansi.Strip(sampleHeader.View(tc.width)); got != tc.want {
			t.Errorf("View(%d) = %q, want %q", tc.width, got, tc.want)
		}
	}
}

func TestHeaderStateDots(t *testing.T) {
	h := Header{Home: "~", Coordinator: "idle", Supervisor: "stopped", NeedsYou: 2, Clock: "09:05"}
	want := "coordinator ○ idle   supervisor ○ stopped   ? 0 asks   needs you 2   09:05 "
	if got := ansi.Strip(h.View(120)); !strings.HasSuffix(got, want) {
		t.Errorf("View(120) = %q, want suffix %q", got, want)
	}
	if got := ansi.Strip(h.View(80)); !strings.HasSuffix(got, "coord ○ · sup ○ · ? 0 · needs you 2 · 09:05 ") {
		t.Errorf("View(80) = %q", got)
	}
}

func TestHeaderBlockedStateUsesWarningGlyph(t *testing.T) {
	h := Header{Home: "~", Coordinator: "blocked", Supervisor: "live", Clock: "09:05"}
	if got := ansi.Strip(h.View(80)); !strings.Contains(got, "coord ▲ · sup ●") {
		t.Fatalf("compact header = %q; want blocked glyph ▲", got)
	}
	if got := ansi.Strip(h.View(120)); !strings.Contains(got, "coordinator ▲ blocked") {
		t.Fatalf("full header = %q; want blocked glyph and word", got)
	}
}

func TestHeaderCutsLongHome(t *testing.T) {
	h := sampleHeader
	h.Home = "~/" + strings.Repeat("x", 100)
	got := ansi.Strip(h.View(120))
	want := " hive  ~/" + strings.Repeat("x", 34) + "… coordinator"
	if !strings.HasPrefix(got, want) {
		t.Errorf("View(120) = %q, want prefix %q", got, want)
	}
}

// TestHeaderNeedsYouCountIsAccent renders row 0 of AgentsMany.dc.html, where
// one task needs the user and the count is accent and bold.
func TestHeaderNeedsYouCountIsAccent(t *testing.T) {
	h := sampleHeader
	h.NeedsYou = 1
	checkGolden(t, h.View(120), "AgentsMany", rect{0, 0, 120, 1})
}

// The header shows the open asks before needs you, also at 0. The count is
// accent when it is above 0. One ask uses the singular word.
func TestHeaderShowsTheAskCount(t *testing.T) {
	cases := []struct {
		asks, width int
		want        string
	}{
		{0, 120, "supervisor ● live   ? 0 asks   needs you 0   14:22 "},
		{1, 120, "supervisor ● live   ? 1 ask   needs you 0   14:22 "},
		{0, 80, "coord ● · sup ● · ? 0 · needs you 0 · 14:22 "},
		{1, 80, "coord ● · sup ● · ? 1 · needs you 0 · 14:22 "},
	}
	for _, tc := range cases {
		h := sampleHeader
		h.Asks = tc.asks
		got := h.View(tc.width)
		if plain := ansi.Strip(got); !strings.HasSuffix(plain, tc.want) || lipgloss.Width(plain) != tc.width {
			t.Errorf("asks %d at %d = %q, want suffix %q", tc.asks, tc.width, plain, tc.want)
			continue
		}
		grid, err := cellgrid.Convert(got, tc.width, 1)
		if err != nil {
			t.Fatal(err)
		}
		plain := ansi.Strip(got)
		col := lipgloss.Width(plain[:strings.Index(plain, "? ")])
		wantToken := "fg"
		if tc.asks > 0 {
			wantToken = "accent"
		}
		for c := col; c < col+3; c++ {
			cell := grid.At(0, c)
			if cell.Token.String() != wantToken || !cell.Bold {
				t.Errorf("asks %d at %d: cell %d = %q %s bold %v, want %s bold", tc.asks, tc.width, c, cell.Rune, cell.Token, cell.Bold, wantToken)
			}
		}
	}
}
