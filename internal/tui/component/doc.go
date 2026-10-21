// Package component holds the reusable components of the TUI v2 screens. Each
// component is a plain value with a View method that takes the size in cells
// and returns ANSI text of exactly that size. A component does no I/O. It
// measures text with lipgloss.Width and uses only the palette of decision 26.
//
// palette.go holds the palette tokens and their styles.
// header.go holds Header, the first row with the home and the agent states.
// tabbar.go holds TabBar, the screen labels and the rule below them.
// subtabs.go holds SubTabs, the views of a screen.
// panel.go holds Panel, a region with a rounded border and a title.
// scrollmark.go holds ScrollMark and Thumb, the scroll marks of a region.
// viewport.go holds Viewport, a Panel whose body scrolls.
// keyhints.go holds KeyHints, the footer with the region and its keys.
package component
