# Chat noise levels

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | 88-tui-terminal-io, 94-chat-user-only (both change Chat). Base: `dev` after both land |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Task 89-screen-board-audit compared each screen with the board that the user
accepted. This task fixes the gaps below. Read the gap list in
`.hive/spec/hive-tui/tasks/89-screen-board-audit/evidence/GAPS.md` (on
`dev` after 89 lands, else on the branch `origin/task/89-screen-board-audit`)
and the board `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Chat.dc.html`. The feature README is
`.hive/spec/hive-tui/README.md`. Its decisions and G1-G14 win over the
board. G12: the TUI has no action except chat, attach, peek, and quit, and
read-only views. When the code does not agree with this brief, stop and report
it with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Gaps to fix

- C1: The noise levels are `all`, `landings and decisions`, and `decisions only`, as on the board. Write in the evidence which events each level shows.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/chat_screen.go` and their tests | The gaps. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files. When a gap needs a change in another file, stop and
report it.

## Requirements

- Write each test first and see it fail for the correct reason. Test at the
  model with `Update` and `View`. Write each expected text as a literal.
- Each frame stays exactly the size of the terminal at 120 x 40 and 200 x 57.
  Use the frame size test of 88-tui-terminal-io.
- Keep each test that passes today. Do not weaken an assertion.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Copy a marked home to a temp directory. Set `HIVE_HOME` to that directory. Set `XDG_CONFIG_HOME` to a temp directory. Start the binary in a Herdr tab at about 200 x 57. Use `herdr pane read` to capture the screen for each gap. | The screen matches the board for each gap. Save the captures in the evidence. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never open the Chat input on the real home. Close each herdr tab and
workspace that you open. Do not close a pane that you did not open.

## Finish

Commit on the task branch with a subject such as `fix(tui): match the Chat noise levels`. Push the task
branch. Write the evidence. Run `hive task complete`. Send the commit SHA and
the V1-V4 results to the manager with `hive inbox send`. Read only your own
inbox.

## Repair 1

The verifier failed V2: the level rules hide `blocked` and `unresolved`, and
the default level hides `quiet`, which the board shows. See `VERIFIED.md`.
Use these rules for the inbox kinds:

| Level | Kinds shown |
| --- | --- |
| `decisions only` | `message`, `blocked`, `unresolved` |
| `landings and decisions` (default) | the kinds of `decisions only`, and `completed`, `verified`, `quiet` |
| `all` | each kind |

Do these steps:

1. Merge `origin/dev` into the task branch. In `testdata/ste-baseline.json`,
   keep each row of both sides.
2. Change the level rules to the table above. Change the tests first: write
   each expected kind set as a literal, and see each changed test fail on the
   current code.
3. Update the level rules in the evidence. Run V2 again on a scratch home
   with one inbox record of each kind. Run the gate.
4. Commit and push. Make sure that `git merge-tree --write-tree origin/dev
   HEAD` exits 0. Run `hive task complete`. Send the commit SHA to the
   manager with `hive inbox send`.

Do not change other files.
