package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// statusDuration gives the duration from timestamp to now as HH:MM:SS with
// whole seconds and total hours. It reports false when the timestamp is
// missing or invalid. A future timestamp gives 00:00:00.
func statusDuration(timestamp string, now time.Time) (string, bool) {
	if timestamp == "" {
		return "", false
	}
	changedAt, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return "", false
	}
	elapsed := now.Sub(changedAt)
	if elapsed < 0 {
		elapsed = 0
	}
	total := int64(elapsed / time.Second)
	hours := total / 3600
	minutes := total % 3600 / 60
	seconds := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds), true
}

// fitStatusAge gives a status word and its age in width cells. The age stays
// complete. A tight row uses one space before it, and a row that is still too
// tight elides the status word.
func fitStatusAge(status, age string, width int) string {
	if age == "" {
		return ansi.Truncate(status, width, "…")
	}
	if ansi.StringWidth(status)+3+ansi.StringWidth(age) <= width {
		return status + " · " + age
	}
	if ansi.StringWidth(status)+1+ansi.StringWidth(age) <= width {
		return status + " " + age
	}
	room := width - 3 - ansi.StringWidth(age)
	if room <= 0 {
		return age
	}
	return ansi.Truncate(status, room, "…") + " · " + age
}

// fitLabelDuration gives a label and its duration in width cells. The
// duration stays complete and the label is elided to make room.
func fitLabelDuration(label, duration string, width int) string {
	room := width - 3 - ansi.StringWidth(duration)
	if room <= 0 {
		return duration
	}
	return ansi.Truncate(label, room, "…") + " · " + duration
}
