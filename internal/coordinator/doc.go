// Package coordinator runs bare hive: it picks the harness and the model,
// writes the coordinator record, and replaces the process with the harness.
//
// The picker has two steps in one bordered box. Step 1 lists the harnesses
// claude, codex, and pi. Enter lists the models of the chosen harness behind a
// spinner. Step 2 lists the models with a type-to-filter line and an Other…
// row for a typed model name. When the model list fails, step 2 asks for a
// typed name. ▸ marks the cursor, and "✓ last" marks the last choice, where
// the cursor starts. Enter on a model ends the picker with the choice. Esc
// goes back one step, or clears the filter first, and quits on step 1. Ctrl+C
// quits from any step.
//
// discover.go lists the models of each harness from pi and opencode.
// choice.go reads and writes the last choice at <home>/coordinator-choice.json.
// picker.go is the Bubble Tea picker.
// launch.go saves the coordinator record, labels its tab, and starts the harness
// in this pane.
package coordinator
