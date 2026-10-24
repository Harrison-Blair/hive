// Package worktree creates and removes the worktree of one task.
//
// create.go owns the eligibility checks, the Git prechecks, the herdr request,
// the effect inspection, and the record write. remove.go owns the identity
// checks, the guards, the workspace close, the Git removal, and the record
// update. spec.go owns the spec copy-in and its hash record. render.go owns
// the text output.
package worktree
