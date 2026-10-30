// Package harness defines the native profile of one harness kind.
//
// harness.go owns the Profile interface, the bounds check, and the lookup.
// registry.go owns the profile registry. claude.go owns the Claude Code profile.
// codex.go owns the Codex profile. pi.go owns the Pi profile. message.go owns
// the shared parts of message delivery: the harness process of a pane and the
// socket write. startline.go owns the length limit of the start command line.
package harness
