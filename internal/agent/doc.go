// Package agent holds the shared agent rules that several agent operations, the TUI, and the supervisor use.
// register.go writes agent records and singleton claims; list.go and show.go read them; render.go formats shared effects.
// spawn.go starts one agent and records its identity; coordinator_claim.go takes, settles, and releases the coordinator starting claim; start_coordinator.go starts the coordinator harness under the claim, writes its record, and prompts it; stop.go closes one spawned pane; inspect.go classifies a missing name.
// workspace.go finds, adopts, or makes the manager workspace and chooses its root pane; instructions.go writes the project instructions.
// spawn_worker.go resolves the worktree pane of a worker or verifier spawn and sends its prompt; brief.go builds the prompt texts.
// supervisor.go starts the singleton supervisor under one held store lock.
// restart.go restarts the coordinator under a new starting claim, or the supervisor, and updates the old record.
// caller.go finds the caller of one command from the herdr pane (decision 24); route.go gives the inbox key of an agent and the parent inbox of its nudges (decisions 26, 28).
// Each agent command has its operation in a child package; agenttest holds the fixtures of their tests.
package agent
