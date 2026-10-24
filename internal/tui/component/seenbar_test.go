package component

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Friction rule 1: GfFriction rows 14, 7, and 3, the seen bars of 5, 3,
// and 2 occurrences. From 3 occurrences the bar is accent, else dim.
func TestSeenBarGolden(t *testing.T) {
	checkGolden(t, SeenBar{Count: 5}.View(), "GfFriction", rect{14, 46, 5, 1})
	checkGolden(t, SeenBar{Count: 3}.View(), "GfFriction", rect{7, 46, 3, 1})
	checkGolden(t, SeenBar{Count: 2}.View(), "GfFriction", rect{3, 46, 2, 1})
}

// Friction rule 1: the bar has one ▮ for each occurrence, up to 10.
func TestSeenBarStopsAtTen(t *testing.T) {
	for _, tc := range []struct {
		count int
		want  string
	}{{0, ""}, {1, "▮"}, {10, "▮▮▮▮▮▮▮▮▮▮"}, {14, "▮▮▮▮▮▮▮▮▮▮"}} {
		if got := ansi.Strip(SeenBar{Count: tc.count}.View()); got != tc.want {
			t.Errorf("SeenBar{%d} = %q, want %q", tc.count, got, tc.want)
		}
	}
}
