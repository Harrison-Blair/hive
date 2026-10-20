# No painted backgrounds

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | none. Base: `dev` at `ef2dde2e` or later |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user wants the terminal background to show everywhere in the TUI. Remove
each painted background. Text colors, borders, and the honey accent
`#f5b942` stay. When the code does not agree with this brief, stop and report
it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `internal/tui/shell.go` defines `colorBackground = "#0e0f13"` and
  `colorPanel = "#12141a"`. `headerStyle` and `footerStyle` use
  `colorBackground` as a background. `panelStyle` uses `colorPanel`.
- `selectedStyle` is honey text with bold and no background. It marks the
  selected sidebar item and the selected rows.
- The boards also show the fills `#241f14` and `#2a2015`. The code on `dev`
  has no such fill. Find each other background in `internal/tui/` (for
  example a `Background` call, a `BackgroundColor` field of `tea.View`, or
  `Reverse`).
- Tasks 90, 91, 96, 97, 98, and 99 change other parts of `internal/tui/` at
  the same time. Keep the change small: change only the style definitions and
  the places that paint a background. Do not move code or rename a style.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/shell.go`, each other `internal/tui/` file that paints a background, and their tests | R1, R2. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Styles

- No style in `internal/tui/` sets a background, except the selected
  highlight when it has one.
- The selected sidebar item and the selected rows keep the look that they
  have now (honey text, bold).
- Remove each color constant that becomes unused.

### R2 Test

- Write the test first and see it fail on the base for the correct reason.
- Render each of the 8 screens at 120 x 40 with fixture data. Check the raw
  output, with the ANSI codes. No SGR sequence sets a background: no `48;`
  parameter and no parameter from 40 to 47 or 100 to 107. Parse each SGR
  sequence in the test. Do not use the style code under test to find the
  codes.
- Check that the honey foreground `38;2;245;185;66` is still in the output
  of the sidebar.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run the R2 test against the base code, then against the task code | It fails on the base and passes on the task branch. |
| V2 | Build the task binary. Run it on a scratch home (`HIVE_HOME` and a temp `XDG_CONFIG_HOME`) in a herdr tab. Capture each screen with `herdr pane read --format ansi` or a similar raw capture | No background code in each capture. The honey selection shows. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never open the Chat input on the real home. Close each herdr tab and
workspace that you open. Do not close a pane that you did not open.

## Finish

Commit on the task branch with a subject such as `fix(tui): remove painted
backgrounds`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the
manager with `hive inbox send`. Read only your own inbox.
