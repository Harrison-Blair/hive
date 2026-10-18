// Package agent registers, lists, and shows agent sessions.
// register.go writes agent records and singleton claims; list.go and show.go read them; render.go formats results.
// spawn.go starts one agent and records its identity; send.go prompts one agent; stop.go closes one spawned pane; inspect.go classifies a missing name.
// workspace.go finds, adopts, or makes the manager workspace and chooses its root pane; instructions.go writes the project instructions.
package agent
