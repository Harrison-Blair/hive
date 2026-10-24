// Package task creates, queries, and moves hive task records.
// create.go owns task creation; edit.go replaces briefs; transition.go owns
// the status machine and the shared command runner; assign.go, complete.go,
// verify.go, cancel.go, and reopen.go are the transition commands; list.go and
// show.go read records; spec.go checks the feature README in the home and
// copies the evidence out; render.go formats results; log.go writes task log events.
package task
