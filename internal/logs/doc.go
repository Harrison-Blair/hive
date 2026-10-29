// Package logs reads, filters, and renders the daily hive log files, and runs hive logs.
//
// read.go finds the daily files, reads each line with a 1 MiB cap, and returns an end offset.
// filter.go matches an event on level, component, agent or target, and request.
// follow.go follows one daily file and the day rollover.
// render.go makes the text line for one event.
// run.go checks the options of hive logs, prints the matches with --limit, and follows.
package logs
