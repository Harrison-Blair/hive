package component

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// KeyHints is the footer row: the focused region and its key hints
// (decision 4).
type KeyHints struct {
	Region string
	Hints  []string
}

// View renders the footer at exactly width cells: " [region] " in accent
// and bold, then the hints in dim, joined with " · ". The hints are cut with
// … when they do not fit, and padded in dim when they are shorter.
func (k KeyHints) View(width int) string {
	tag := titleStyle.Render(" [" + k.Region + "] ")
	room := width - lipgloss.Width(tag)
	if room < 1 {
		return fitRow(tag, width)
	}
	hints := ansi.Truncate(strings.Join(k.Hints, " · "), room, "…")
	hints += strings.Repeat(" ", room-lipgloss.Width(hints))
	return tag + dimStyle.Render(hints)
}
