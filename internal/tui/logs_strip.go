package tui

import (
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// logStripMargin is the cells of a strip row that are not strip cells: the
// two times, their blanks, and the cells after the end time. At 120 columns
// the strip has 104 cells (logs.md rule 1).
const logStripMargin = 16

// logDensity gives the density strip of logs.md rules 1 and 2 for the kept
// rows at cells cells. The range starts at the first row and ends at the
// last row. Each cell counts the rows of its part of the range. The mark is
// the cell of the row at selected. With no row, the strip is empty.
func logDensity(rows []logRow, cells, selected int, zone *time.Location) component.DensityStrip {
	strip := component.DensityStrip{Start: "--:--", End: "--:--", Cells: make([]component.DensityCell, max(cells, 0)), Mark: -1}
	times := make([]time.Time, len(rows))
	var first, last time.Time
	for index, row := range rows {
		when, err := time.Parse(time.RFC3339Nano, logField(row.event, "time"))
		if err != nil {
			continue
		}
		times[index] = when
		if first.IsZero() || when.Before(first) {
			first = when
		}
		if when.After(last) {
			last = when
		}
	}
	if first.IsZero() || cells < 1 {
		return strip
	}
	strip.Start, strip.End = first.In(zone).Format("15:04"), last.In(zone).Format("15:04")
	span := last.Sub(first)
	bucket := func(when time.Time) int {
		if span <= 0 {
			return 0
		}
		return min(int(int64(when.Sub(first))*int64(cells)/int64(span)), cells-1)
	}
	for index, row := range rows {
		if times[index].IsZero() {
			continue
		}
		cell := &strip.Cells[bucket(times[index])]
		cell.Count++
		switch level := logLevel(row.event); {
		case level == "ERROR", level == "WARN" && cell.Level != "ERROR":
			cell.Level = level
		}
	}
	if selected >= 0 && selected < len(rows) && !times[selected].IsZero() {
		row := rows[selected]
		strip.Mark = bucket(times[selected])
		strip.MarkLevel = logLevel(row.event)
		strip.MarkText = times[selected].In(zone).Format("15:04") + " " + strings.ToLower(strip.MarkLevel) + " · " + sanitizeText(logField(row.event, "msg"))
	}
	return strip
}
