// Package tui is the hive terminal interface that bare `hive` opens.
//
// tui.go defines the root model, its lifecycle, and the program runner.
// shell.go defines the screen interface, frame, palette, and placeholder screens.
// header.go loads and renders project, agent, supervisor, task, and clock data.
// logview.go renders and filters log lines and handles log keys.
// tick.go schedules refreshes and carries background load results.
package tui
