# Visual check rule in the skills

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | 102-selection-marker (order only). Base: `dev` |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved a visual check rule on 2026-10-20. A task that changes a
UI must compare PNG captures of each changed screen with PNG captures of each
approved board. This task writes the rule into the worker skill, the verifier
skill, the brief format, and the evidence format. It adds no script and no
command. When the code does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

## The rule

- Trigger: the brief has a `Design` section that names each approved board by
  absolute path. A verifier also flags a UI change with no `Design` section,
  for example a change in `internal/tui/`, a `*_view.py` file, or hub CSS.
- Worker: capture each changed screen and each board as PNG. Compare them.
  Fix each gap before you finish.
- Verifier: capture the screens and the boards again, independently. Judge
  the result.
- Board to PNG: `firefox --headless --no-remote --profile SCRATCH
  --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html`, with a
  new scratch profile folder.
- Textual app to PNG: in a pilot test, `app.save_screenshot()` gives an SVG.
  Then `rsvg-convert -o OUT.png IN.svg`.
- Terminal pane to PNG: `herdr pane read PANE --source visible --format ansi
  > F.ansi` on a pane that you started. Change the ANSI SGR codes to an HTML
  `<pre>` with a short script in scratch. Then use the firefox command.
- Evidence: commit the screen PNGs in the task evidence. A board PNG stays in
  the home below `.hive/tmp/`. Give its path. Never commit a board.
- Report: one table with one row for each board element: `matches`,
  `differs` (with the cause), or `overridden` (with the decision that
  overrides it).
- Ownership: capture only your own panes or a scratch run.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.agents/skills/hive-worker/SKILL.md` | R1. |
| `.agents/skills/hive-verifier/SKILL.md` | R2. |
| `.hive/spec/BRIEF-FORMAT.md` | R3. |
| `.hive/spec/EVIDENCE-FORMAT.md` | R4. |
| `skills_test.go` and the tests that check these files | Only when a test reads changed text. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

- R1: The worker skill gives the trigger, the worker step, the three capture
  recipes inline, the evidence placement, the report table, and the
  ownership limit.
- R2: The verifier skill gives the trigger (with the flag for a UI change
  with no `Design` section), the verifier step, the same recipes inline, and
  the report table.
- R3: The brief format adds a `Design` section: each approved board by
  absolute path, and the rule that a UI task must have it.
- R4: The evidence format adds the element table and the PNG placement.
- Write all text in ASD-STE100 Strict. Use each term of `STE-GLOSSARY.md`
  only with its meaning. `ste_test.go` must pass. A baseline count cannot
  rise.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Read each changed file against The rule | Each point of the rule is in the file that R1-R4 names. The task adds no script or command. |
| V2 | Run the board recipe on `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html` and the pane recipe on a scratch TUI run in a pane that you started | Each gives a PNG that shows the board or the screen. Do not commit them. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never work in workspace `wBC`. Stop and close only the pid and pane that you
started.

## Finish

Commit on the task branch with a subject such as `docs(skills): add the
visual check rule`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the
manager with `hive inbox send --to manager-hive`.

## Scope change: colour is never the only signal

The user is colourblind (`overrides/colourblind-ui.md` in the home). Add this
point to the rule in each file of R1-R4:

- In each user interface, a glyph or a word shows each state. Colour can add
  to a state. Colour must not be the only signal of a state. The worker
  checks the plain text of each changed screen (with the colour removed) for
  the glyph or word of each state. The element table has a column "state
  without colour". The verifier fails a state that only colour shows.

V1 also checks this point in each file.
