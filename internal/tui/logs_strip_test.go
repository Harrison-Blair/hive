package tui

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// stripRow gives a list row of an event at the time and level.
func stripRow(t *testing.T, at, level, msg string) logRow {
	t.Helper()
	fields := map[string]any{"time": at, "level": level, "msg": msg}
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return logRow{event: logs.Event{Raw: raw, Fields: fields}, text: msg}
}

// logs.md rule 1: the range is the first to the last kept event, cut into
// the strip cells. Each cell counts its events and takes the highest of WARN
// and ERROR. Rule 2: the mark is the cell of the selected event, with the
// text HH:MM level · msg.
func TestLogDensityBuckets(t *testing.T) {
	rows := []logRow{
		stripRow(t, "2026-10-21T13:00:00Z", "DEBUG", "a"),
		stripRow(t, "2026-10-21T13:00:30Z", "INFO", "b"),
		stripRow(t, "2026-10-21T13:01:00Z", "WARN", "c"),
		stripRow(t, "2026-10-21T13:03:59Z", "ERROR", "d"),
		stripRow(t, "2026-10-21T13:03:59Z", "WARN", "e"),
		stripRow(t, "2026-10-21T13:04:00Z", "INFO", "verify fail"),
	}
	// 4 minutes in 4 cells: 13:00, 13:01, 13:02, 13:03. The last event at
	// the end of the range is in the last cell. A WARN after an ERROR keeps
	// the ERROR of the cell.
	got := logDensity(rows, 4, 5, time.UTC)
	want := []component.DensityCell{{Count: 2}, {Count: 1, Level: "WARN"}, {}, {Count: 3, Level: "ERROR"}}
	if len(got.Cells) != len(want) {
		t.Fatalf("cells = %+v, want %+v", got.Cells, want)
	}
	for i := range want {
		if got.Cells[i] != want[i] {
			t.Errorf("cell %d = %+v, want %+v", i, got.Cells[i], want[i])
		}
	}
	if got.Start != "13:00" || got.End != "13:04" {
		t.Errorf("start %q end %q, want 13:00 and 13:04", got.Start, got.End)
	}
	if got.Mark != 3 || got.MarkText != "13:04 info · verify fail" || got.MarkLevel != "INFO" {
		t.Errorf("mark %d %q %q, want 3, 13:04 info · verify fail, INFO", got.Mark, got.MarkText, got.MarkLevel)
	}
}

// logs.md rules 7 and 8: with no kept event, the strip is empty: blank
// cells, --:-- at the ends, and no mark.
func TestLogDensityEmpty(t *testing.T) {
	got := logDensity(nil, 5, 0, time.UTC)
	if got.Start != "--:--" || got.End != "--:--" || got.Mark >= 0 || len(got.Cells) != 5 {
		t.Fatalf("empty strip = %+v, want 5 blank cells, --:--, no mark", got)
	}
	for i, cell := range got.Cells {
		if cell.Count != 0 {
			t.Errorf("cell %d count %d, want 0", i, cell.Count)
		}
	}
}
