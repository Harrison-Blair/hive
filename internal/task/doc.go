// Package task holds the shared task rules that several task operations use.
// transition.go owns the status machine and the shared transition and finish
// runners; list.go owns the task list that the TUI and the list command read;
// brief.go reads and writes briefs; paths.go names the feature paths; spec.go
// copies the evidence out; render.go formats shared results; log.go writes
// task log events. Each task command has its operation in a child package.
package task
