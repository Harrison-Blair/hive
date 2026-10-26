// Package worktree holds the shared worktree rules that the create and remove
// operations and land use.
//
// spec.go owns the spec copy-in record, its read and write, and the file hash
// that land compares. result.go owns the envelopes, the one-line text, and the
// record time. paths.go compares canonical paths. log.go writes the worktree
// log events. render.go renders the effect lines. Each worktree command has its
// operation in a child package.
package worktree
