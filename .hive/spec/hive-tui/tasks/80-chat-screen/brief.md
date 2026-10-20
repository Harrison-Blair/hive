# Chat screen

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, wave B |
| Tier | strong |
| Depends on | 79-tui-shell. Base: `dev` after it lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Fill screen 1, Chat: the user talks to the coordinator in the TUI. The feature
README is `.hive/spec/hive-tui/README.md`. Read it first, mainly goal
decisions 4, 11, 12, 15, 16, 27 and G2, G3, G12, G14. Its decisions are
approved. Do not change a decision. When the code does not agree with this
brief or the README, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Sources

- The design in the home: `.hive/tmp/coordinator/designs/hive-tui/Chat.dc.html`.
  The decisions win over the design. Do not make the question card, the
  `1-3 answer` keys, the typing guard, or `esc stop the reply` (G12, goal
  decision 15).
- The shell: `internal/tui/shell.go` (the `screen` interface) and
  `internal/tui/tick.go` (background load messages).
- The chat transport: `internal/chat/` and `hive chat send/reply/log`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` (a new file for the screen, and the shell wiring for it) and its tests | R1-R5. |
| `internal/chat/` and its tests | Only to export a function that the screen calls, when the package has none. No change of behavior. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### R1 Conversation

- The main panel has the title `coordinator`, with the state of the
  coordinator (live or not, its harness status), its model, and its pane ID.
- It shows the chat entries of the last 7 days from `.hive/state/chat/`, in
  order: the sender (`you` or `coordinator`), the time, and the text. A user
  message shows its delivery state when it is not `sent`. A reply with
  `reply_to` shows a short mark that names the ID.
- The newest entry is at the bottom. `pgup` and `pgdown` scroll.

### R2 Input

- `i` starts the input. `enter` sends. `shift+enter` (and `alt+enter`, for a
  terminal that does not report `shift+enter`) adds a new line. `esc` leaves
  the input and keeps the draft.
- The send runs in a `tea.Cmd` with the same function as `hive chat send`.
  While a send runs, a second `enter` does not send again.
- A send that fails, or that returns `unknown`, shows the reason in the
  footer and keeps the text in the input. It never sends again on its own.
- In the input, `ctrl+c` keeps its shell meaning. The keys `1` to `8` go
  into the text and do not change the screen.

### R3 Hive events

- A side panel `hive events` shows the inbox records of the coordinator,
  newest first: the time, the sender, and the first line of the text.
- `n` changes the noise level: `all` (each record), `messages` (only records
  of kind `message`), `off` (the panel is empty and shows `off`). The default
  is `messages`.
- In the conversation, the records that came between two chat entries fold
  into one line, `── N hive events ──`. `e` expands and folds them again.
- The TUI does not mark a record as read.

### R4 Loading

All reads run in a `tea.Cmd` through the shell `DataRequest`. A load reads
only the chat files of the last 7 days and the coordinator inbox.

### R5 Tests

- Test at the model with a temp home that holds chat files and inbox
  records. Cover each R1-R3 rule and the keys of the footer.
- A test with a fake send that blocks: a key still works, and a second
  `enter` does not send.
- A program test on a string reader.

## Live check of G14

The README asks for one live check of a chat message and a nudge to the same
pane at the same time in wave B. V2 does it.

## Rules

- Each Go change is test-first. See each new test fail for the correct
  reason.
- When `ste_test.go` fails on the brief copy, change the words of each
  sentence in the copy that causes a violation. Keep its meaning. Record each
  change in the evidence.
- Live checks use a temp `HIVE_HOME` and a temp `XDG_CONFIG_HOME`, and clean
  up each herdr workspace that they make. Do not run `rm -rf` on an unquoted
  variable.


## Not in scope

A question card. A typing guard. A change to `hive chat` behavior. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each new test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Live, in a new herdr workspace, with a temp `HIVE_HOME`: register a coordinator record in a pane that runs a cheap agent. Run the task binary in another pane. Type a two-line message in Chat and send it. At the same time, run `hive agent send` to the coordinator pane. Then run `hive chat reply --to <id>` as the coordinator. Read both panes with `herdr pane read`. | The coordinator pane shows `⬢ user:` with the two lines, separate from the other prompt. The Chat screen shows the message as sent and the reply under it. Clean up the workspace. |
| V3 | `git diff dev --stat` | Only the files of the source table and the evidence changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `feat(tui): add the Chat
screen`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V4 results to the manager with
`hive inbox send`.
