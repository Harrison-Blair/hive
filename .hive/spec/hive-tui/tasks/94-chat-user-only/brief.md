# Chat send for the user only

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | strong |
| Depends on | none. Base: `dev`. Each other task of wave D lands after this task. |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Hive tags each chat message with `⬢ user:`. The role gate gives this tag the
authority of the user (decision 27, G3). On 2026-10-20 a worker ran bare
`hive` on the home, opened the Chat input, and sent `278` to the
coordinator. Any agent that runs the TUI or `hive chat send` can speak as the
user. Close this gap. The friction entry is
`reference/dogfood/friction/entries/agent-can-send-user-chat.md`. Do not edit
it. When the code does not agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `chat.Send` in `internal/chat/chat.go` calls `checkUserCaller`. It refuses
  a caller with `--agent`, and a caller whose `PaneID` resolves to a live
  agent. `cmd/chat/chat.go` sets `PaneID` from `HERDR_PANE_ID`.
- The TUI calls `chat.Send` from `sendChat` in `internal/tui/chat_screen.go`
  with no `PaneID` and no working directory. So the TUI has no check.
- The worker ran the TUI in a new herdr tab. The pane of that tab is not the
  pane of an agent record. A check that only compares the pane is not
  sufficient.
- An agent record has `pane_id` and `workspace_id`. The user's TUI tab is in
  the hive session workspace with the coordinator and the supervisor.
- A worker and a verifier run with a working directory below
  `projects/<project-id>/worktrees/`. A manager runs below
  `projects/<project-id>/`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/chat/` and its tests | R1, R3. |
| `cmd/chat/chat.go` | R1. Only to pass the working directory. |
| `internal/tui/chat_screen.go`, other `internal/tui/` files for the Chat send path, and their tests | R2, R3. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Refusal rules

`chat.Send` refuses to send, writes no chat entry, and prompts no pane, when
one of these is true:

1. The caller gives `--agent`, or its pane is the pane of a live agent
   record (as today).
2. The working directory of the caller is below `projects/` in the home.
   Resolve symlinks before you compare.
3. The pane of the caller is in the workspace of a live agent record of role
   manager, worker, or verifier.

The error text starts with `chat send is for the user only`. It names the
rule that matched. A pane in the hive session workspace with no agent record
can send, so that the user's TUI tab works. A caller with no `HERDR_PANE_ID`
and a working directory outside `projects/` can send.

### R2 TUI

- The TUI passes its pane (`HERDR_PANE_ID` at start) and its working
  directory to `chat.Send`.
- When `chat.Send` refuses, the Chat screen shows
  `chat send is for the user only` and keeps the input text. It does not
  show the text as sent.
- At start, when rule 2 or rule 3 matches, the Chat screen shows the same
  notice and `i` does not open the input.

### R3 Tests

- Write each test first and see it fail for the correct reason.
- Test each rule for `chat.Send` with fixture agent records and fake herdr.
  Check that Hive writes no chat entry and sends no prompt.
- Test that the user's case sends: a pane in the coordinator workspace with
  no record, and a working directory at the home root.
- Test the TUI path at the model. A refused send shows the notice and keeps
  the text. At start with rule 2, `i` does not open the input.
- Test the CLI path with `ExecuteWithArgs` from a working directory below
  `projects/`.

## Not in scope

`hive chat reply`. Other screens. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R3 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Use a scratch home: a temp copy of a marked home with fixture records, set with `HIVE_HOME`, and a temp `XDG_CONFIG_HOME`. From the worktree, run `./hive chat send hello` | It fails with `chat send is for the user only` and names rule 2. No entry in the chat history of the scratch home. |
| V3 | With the same scratch home, run the TUI from the worktree in a herdr tab that you open. Open Chat and press `i` | The notice shows and the input does not open. Do not press Enter. |
| V4 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0. |

Never run a TUI check on the real home with the Chat input open. Use a
scratch home for each chat check. Close each herdr tab and workspace that you
open. Do not close a pane that you did not open.

## Finish

Commit on the task branch with a subject such as `fix(chat): refuse chat send
from an agent`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V5 results to the
manager with `hive inbox send`. Read only your own inbox.
