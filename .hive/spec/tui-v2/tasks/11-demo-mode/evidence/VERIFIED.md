# Verification: 11-demo-mode

Verdict: fail. Reviewed task commit `e433d7f9`.

The brief requires the verifier to stop when the code or the spec does not agree with the brief. The review found an AC3 conflict before the test and live runs. No worker evidence counts as an independent run.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | fail | Not run. Review found that demo mode does not isolate all screen data from the live filesystem. See AC3 below. Red and green runs remain unverified. |
| V2 Live check | fail | Not run after the required stop. Pane size, fixed clock, file changes, and process effects remain unverified. |
| V3 Schema review | fail | The schema record states that task brief and evidence bodies are absent and that demo mode reads these files from the home. This conflicts with AC3 and spec decision 35. |
| V-gate Diff scope and gate | fail | `git diff dev...HEAD --stat` showed 28 files, all in the allowed paths. No gate command was run after the required stop. The full gate remains unverified. |

## AC3: live file reads

AC3 requires all screen data to come from the fixture and forbids live record reads. Spec decision 35 has the same requirement.

- `internal/tui/demo.go`: `demoData` has evidence names but no task brief or evidence file bodies. `useDemoLoaders` replaces record loaders but does not replace task file reads.
- `internal/tui/task_screen.go:322-324`: the `b` action builds the brief path from the selected project and task, then calls `startTaskFileRead`.
- `internal/tui/task_screen.go:731-739`: `startTaskFileRead` calls `os.ReadFile(path)` without a demo guard or fixture source. The result becomes screen content in `acceptTaskFile`.
- Worker evidence `V3.md`, Decisions, explicitly states that opening a task brief or evidence file in demo mode reads the file from the home.

This finding follows the actual action path. It does not depend on a failing test or a visual comparison. Repair must supply fixture-backed content or obtain a decision that changes the requirement. Do not read live files as a fallback in demo mode.

## Design check

The brief changes the UI but has no `Design` section with approved board absolute paths. The verifier skill requires this section. No new PNG capture or board comparison was made after the required stop. Visual accuracy and colour-independent state checks remain unverified.

## Cleanup

No scratch files, panes, workspaces, or processes were created. No source file or worker evidence file was changed.
