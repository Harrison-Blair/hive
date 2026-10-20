# TUI terminal input and size

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | strong |
| Depends on | none. Base: `dev` |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user ran bare `hive` in a herdr pane. Two faults occurred:

1. Keys that the user typed seemed to have no effect. Keys sent with
   `herdr pane send-keys` had an effect.
2. The TUI did not fill the pane. The right border of the main panel was cut
   off at the pane edge.

The feature README is `.hive/spec/hive-tui/README.md`. Its decisions are
approved. When the code does not agree with this brief, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `Run` in `internal/tui/tui.go` wraps the input in `guardedReader` and the
  output in `guardedOutput`. `guardedReader` has no `Fd` method.
  `guardedOutput` keeps the `term.File` shape for the renderer.
- Hypothesis, not proved: Bubble Tea cannot find the terminal of the input
  through `guardedReader`. Then it cannot set raw mode, and the terminal
  holds typed keys until Enter. Prove or disprove this before you change
  code. Record the result in the evidence.
- `Run` sets `tea.WithWindowSize(defaultWidth, defaultHeight)` (80 x 24).
- `Update` sets `m.width` and `m.height` from `tea.WindowSizeMsg`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` and its tests | R1, R2, R3. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Typed keys

In a real terminal, each key that the user types reaches `Update` at once,
with no Enter. Keep the effect of the stream guards: an input or output error
stops the program, and the end of input ends it.

### R2 Full terminal

- Each frame is exactly the width and the height of the last
  `tea.WindowSizeMsg`: the same count of lines as the height, and each line
  has the display width of the terminal. No border goes past the edge.
- At start, the TUI uses the real size of the terminal, not 80 x 24, when the
  output is a terminal.
- Each resize gives a new frame at the new size.

### R3 Tests

- Write each test first and see it fail for the correct reason.
- Add a model test: send `tea.WindowSizeMsg` for each of these sizes and
  check the frame of each screen: 80 x 24, 120 x 40, 200 x 57, 57 x 20,
  300 x 80. Count lines and the display width of each line with an
  independent measure (for example `ansi.StringWidth` on each line). Do not
  use the layout code under test to compute the expected size.
- Add a test for R1 when you can make one without a real terminal. When you
  cannot, say why in the evidence and rely on V2.

## Not in scope

The Agents layout, the idle CPU, the content of other screens. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R3 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Run it in a new herdr tab. Set `XDG_CONFIG_HOME` and `HIVE_HOME` to temporary paths. Set `HIVE_HOME` to a temporary copy of a fixture home, or use this home in read mode. Send single keys to the pane without Enter. Use a terminal that you control, such as `script` or `tmux`. Repeat the check with the base binary. | Record whether the base binary acts on each single key. Record the raw-mode state with `stty -a < /dev/pts/N` for each binary when possible. The task binary changes the screen on each key. |
| V3 | In the same pane, resize the pane (herdr split or a terminal resize). Capture the frame with `herdr pane read` before and after | Each capture fills the pane. The right border is at the last column. |
| V4 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V5 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other commands exit 0. |

Close each herdr tab and workspace that you open. Do not close a pane that
you did not open. Do not write `~/.config/hive/config.toml` of the user.

## Finish

Commit on the task branch with a subject such as `fix(tui): read typed keys
and fill the terminal`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V5 results to the
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
