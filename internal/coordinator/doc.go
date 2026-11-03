// Package coordinator runs bare hive: it picks the harness and the model,
// writes the coordinator record, and replaces the process with the harness.
//
// The picker has two steps in one rounded box of 62 × 15 cells with a key
// line under it, in the palette of internal/lib/style that the agent graph
// uses. Each step has the same size. In a smaller pane the box takes the pane
// width and loses rows until the box and the key line fit; the harness and
// model lists scroll so the cursor row stays in the box. A list row that is
// too narrow drops the last model column of step 1 first, then "✓ last", and
// then cuts the name with …, keeping its first cell; ▸ always stays. The
// filter row drops its count first, then its label, and then cuts the start
// of the typed text; ▏ always stays. A last guard cuts the view to the pane
// size.
// The title in the top edge names the step. Step 1 lists the harnesses
// claude, codex, and pi, each in its harness colour, with the last model of
// each harness, or – when it has none. Enter lists the models of the chosen
// harness behind a spinner. Step 2 lists the models under a type-to-filter
// row, with an Other… row for a typed model name. The cursor starts on the
// last model of that harness. When the model list fails, step 2 shows the
// cause and asks for a typed name. ▸ and bold mark the cursor. "✓ last"
// marks the last-started harness on step 1 and the last model of the open
// harness on step 2. Enter on a model ends the picker with the choice.
// Esc goes back one step, or clears the filter first, and quits on step 1.
// Ctrl+C quits from any step.
//
// discover.go lists the models of each harness from pi and opencode.
// choice.go reads and writes the last choice and the last model of each
// harness at <home>/coordinator-choice.json.
// picker.go is the Bubble Tea picker.
// name.go resolves the coordinator name from the start directory and guards
// the record of that name.
// launch.go saves the coordinator record, labels its tab, and starts the harness
// in this pane.
package coordinator
