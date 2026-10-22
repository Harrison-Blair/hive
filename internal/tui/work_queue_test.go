package tui

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// AC2: the demo view of the WorkQueue fixture matches its golden grid cell
// for cell.
func TestWorkQueueMatchesItsBoard(t *testing.T) {
	want := readBoardGrid(t, "WorkQueue")
	m := newTestDemoModel(t, "WorkQueue.fixture.json", demoSize{})
	got := renderGrid(t, m, want.Width, want.Height)
	diffs, err := cellgrid.Compare(want, got, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diffs) != 0 {
		t.Fatalf("%d cells differ from WorkQueue.grid:\n%s", len(diffs), cellgrid.FormatDiffs(diffs))
	}
}
