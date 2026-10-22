package component

import "charm.land/lipgloss/v2"

// The palette tokens of decision 26. fg and the background are the terminal
// defaults, so text in fg has no colour.
var (
	Border  = lipgloss.Color("#3d3a33")
	Dim     = lipgloss.Color("#a39d90")
	Accent  = lipgloss.Color("#f5b942")
	Working = lipgloss.Color("#6ee69a")
	Done    = lipgloss.Color("#7ccf6a")
	Info    = lipgloss.Color("#6fb7ff")
	User    = lipgloss.Color("#e9cf9b")
	Code    = lipgloss.Color("#9fd0ff")
)

var (
	borderStyle = lipgloss.NewStyle().Foreground(Border)
	dimStyle    = lipgloss.NewStyle().Foreground(Dim)
	accentStyle = lipgloss.NewStyle().Foreground(Accent)
	titleStyle  = accentStyle.Bold(true)
)
