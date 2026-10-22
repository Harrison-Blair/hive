# Screen board audit

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | decent |
| Depends on | none. Base: `dev` |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user accepted one board for each TUI screen. Compare each screen of the
live TUI with its board and write a list of the gaps. This task changes no
code. The manager makes fix tasks from the list. The feature README is
`.hive/spec/hive-tui/README.md`. Its decisions and G1-G14 win over a board.
When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- The boards are in `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/`: `Chat.dc.html`, `Goals.dc.html`,
  `Queue.dc.html`, `Tasks.dc.html`, `Friction.dc.html`,
  `Health.dc.html`, `Logs.dc.html`. `Main.dc.html` is the Agents board
  and gives the shared header. `NeedsYou.dc.html` and `Timeline.dc.html`
  are not screens (decision 7). Do not audit them.
- Other tasks fix these faults. Do not list them: the terminal size and typed
  keys (T1), the idle CPU (T2), the Agents data and layout (T3, T4).
- G12 removes each action except chat, attach, peek, and quit. A board
  control that G12 removes is not a gap when the screen shows `ask in Chat`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/hive-tui/tasks/<task-id>/evidence/` | R1. The report and the captures. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |
| `testdata/ste-baseline.json` | Only to add the brief copy with 0. |

Do not edit other files.

## Requirements

### R1 Report

- Build the `dev` binary. Run it on this home in a herdr pane that is about
  200 x 57. For each of the 7 screens, capture the screen with
  `herdr pane read`. Put each capture in the evidence.
- Read each board file. Write `GAPS.md` in the evidence. For each screen,
  list each gap as one row: the board part, what the screen shows, the
  decision or G item that applies, and a size (S, M, or L). Also list each
  board part that matches, in one line for each screen.
- Group the gaps into fix tasks that you propose: one task for each group of
  gaps that touch the same files. For each proposed task, give a title, the
  gaps, the files, and the order.
- Do not fix a gap.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Each of the 7 screens has a capture and a section in `GAPS.md` | 7 captures, 7 sections. |
| V2 | Pick 3 gaps at random. Check each against the board and the capture | Each gap is real and the decision column is correct. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | `go test ./...` from the worktree root | Exit 0 (the brief copy passes `ste_test.go`). |

Close each herdr tab and workspace that you open. Do not close a pane that
you did not open.

## Finish

Commit on the task branch with a subject such as `docs(tui): audit the
screens against the boards`. Push the task branch. Run
`hive task complete`. Send the commit SHA and the proposed fix tasks to the
manager with `hive inbox send`. Read only your own inbox.

## Task names

This brief uses these names: T1 is 88-tui-terminal-io, T2 is 90-tui-idle-cpu,
T3 is 91-agents-live-generation, T4 is 92-agents-board-layout, and T5 is
89-screen-board-audit.

## Scope change: chat safety

A TUI check uses a scratch home (a temp copy of a marked home, set with
`HIVE_HOME`, and a temp `XDG_CONFIG_HOME`), or it does not open the Chat
input. On the real home, never press `i` on the Chat screen and never press
Enter there. A chat message from an agent has the authority of the user.

## Repair 1

`hive land` failed: the branch has a merge conflict with `origin/dev` in
`testdata/ste-baseline.json`. Task 94-chat-user-only also added a row for its
brief copy. Do only this repair:

1. Merge `origin/dev` into the task branch.
2. In `testdata/ste-baseline.json`, keep both rows: the row of 94 and the row
   of this task. Keep the order and format of the file.
3. Run `go test ./` from the worktree root. It must pass.
4. Commit with subject `merge(dev): resolve ste baseline conflict`. Push the
   task branch.
5. Make sure that `git merge-tree --write-tree origin/dev HEAD` exits 0.
6. Run `hive task complete`. Send the commit SHA to the manager with
   `hive inbox send`.

Do not change other files. Do not run the TUI.
