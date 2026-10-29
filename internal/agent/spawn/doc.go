// Package spawn starts one agent in a herdr pane and writes its record.
//
// spawn.go checks the options, starts the harness, waits for it, and sends the
// first prompt; place.go finds or makes the pane: a tab in the caller workspace,
// a tab in the checkout workspace of a project, or the root pane of a worktree.
package spawn
