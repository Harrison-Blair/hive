package component

import (
	"strings"
	"testing"

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
		{119, " hive" + strings.Repeat(" ", 76) + "coord ● · sup ● · needs you 0 · 14:22 "},
		{130, " hive  ~/source/hive" + strings.Repeat(" ", 46) + "coordinator ● working   supervisor ● live   needs you 0   14:22 "},
	}
	for _, tc := range cases {
		if got := ansi.Strip(sampleHeader.View(tc.width)); got != tc.want {
			t.Errorf("View(%d) = %q, want %q", tc.width, got, tc.want)
		}
	}
}

func TestHeaderStateDots(t *testing.T) {
	h := Header{Home: "~", Coordinator: "idle", Supervisor: "stopped", NeedsYou: 2, Clock: "09:05"}
	want := "coordinator ◐ idle   supervisor ○ stopped   needs you 2   09:05 "
	if got := ansi.Strip(h.View(120)); !strings.HasSuffix(got, want) {
		t.Errorf("View(120) = %q, want suffix %q", got, want)
	}
	if got := ansi.Strip(h.View(80)); !strings.HasSuffix(got, "coord ◐ · sup ○ · needs you 2 · 09:05 ") {
		t.Errorf("View(80) = %q", got)
	}
}

func TestHeaderCutsLongHome(t *testing.T) {
	h := sampleHeader
	h.Home = "~/" + strings.Repeat("x", 100)
	got := ansi.Strip(h.View(120))
	want := " hive  ~/" + strings.Repeat("x", 45) + "… coordinator"
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
