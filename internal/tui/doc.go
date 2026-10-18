// Package tui is the hive terminal interface that bare `hive` opens.
//
// tui.go defines the model, its lifecycle, and the entries that run it.
// header.go reads the store and renders the header.
// logview.go renders the filtered log lines and handles the key modes.
// tick.go schedules the 500-millisecond refresh, drains the follower, and carries the follower error.
package tui
