// Package component holds the reusable components of the TUI v2 screens. Each
// component is a plain value with a View method that takes the size in cells
// and returns ANSI text of exactly that size. A component does no I/O. It
// measures text with lipgloss.Width and uses only the palette of decision 26.
//
// palette.go holds the palette tokens and their styles.
// subtabs.go holds SubTabs, the views of a screen.
// panel.go holds Panel, a region with a rounded border and a title.
// scrollmark.go holds ScrollMark and Thumb, the scroll marks of a region.
// viewport.go holds Viewport, a Panel whose body scrolls.
// list.go holds List, a Panel of rows in fixed columns with a selection.
// details.go holds Details, a Panel of label and value lines that scrolls.
// filterbar.go holds FilterBar, the filter text of the log screen.
// input.go holds Input, the message line of the chat.
// card.go holds Card, the box of one agent.
// boxjoin.go holds BoxArms and BoxRune, the light and heavy box joints.
// notice.go holds Notice, a dim line of one notice.
// popup.go holds Popup and DimScreen, a Details box over a dimmed screen.
// markdown.go holds RenderMarkdown, an agent reply rendered by Glamour.
// markdown_style.go holds the Glamour style of decision 29.
// chat.go holds ChatLines, the lines of the chat feed.
// statusrow.go holds StatusRow, a status row of segments and the clock.
// titlerow.go holds TitleRow, the title row and the rule of a screen.
// seenbar.go holds SeenBar, the bar of occurrences of a friction entry.
// fold.go holds Fold, the fold line of a group of rows.
package component
