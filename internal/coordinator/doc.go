// Package coordinator runs bare hive: it picks the harness and the model,
// writes the coordinator record, and replaces the process with the harness.
//
// discover.go lists the models of each harness from pi and opencode.
// choice.go reads and writes the last choice at .hive/coordinator-choice.json.
// launch.go saves the coordinator record and starts the harness in this pane.
package coordinator
