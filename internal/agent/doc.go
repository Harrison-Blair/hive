// Package agent registers, lists, and shows agent sessions.
// register.go writes agent records and singleton claims; list.go and show.go read them; render.go formats results.
// spawn.go starts one agent and records its identity; start_coordinator.go starts and prompts the coordinator; send.go prompts one agent; stop.go closes one spawned pane; inspect.go classifies a missing name.
// workspace.go finds, adopts, or makes the manager workspace and chooses its root pane; instructions.go writes the project instructions.
// spawn_worker.go resolves the worktree pane of a worker or verifier spawn and sends its prompt; brief.go builds the prompt texts.
// supervisor.go starts the singleton supervisor under one held store lock.
// resolve.go settles an unresolved record by hand and checks the live identity.
// restart.go restarts the coordinator or supervisor and updates the old record.
package agent
