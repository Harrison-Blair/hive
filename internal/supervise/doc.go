// Package supervise runs the one relay process of a hive home.
//
// supervise.go owns the claim, the start, the main loop, and the exit.
// subscribe.go owns the selector list, the rebuild on a change, and the
// reconcile across the subscription gap. events.go owns the event handlers of
// the AC2 table. watch.go owns the inotify watch of the inbox and agent
// directories. quiet.go owns the quiet timers and the injectable clock.
// nudge.go owns the nudge rule and the prompt line. reconcile.go owns the
// reconcile by pane and claim.
package supervise
