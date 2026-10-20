# VERIFIED: 99-chat-noise-levels

Verifier: Claude Opus 5.5 (`claude-opus-5-5`), independent of the worker.
Task commit: `17469a98` (code commit `0bff20f2`). Base: `ef2dde2e` (merge base with `origin/dev`).
Board: `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Chat.dc.html`.

Verdict: **fail**, on V2. The level names match the board, but the event set of each level is not sensible and does not agree with the board.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Copied the task `chat_screen_test.go` into a detached worktree at the base. `TestChatNoiseLevelsFilterInboxKinds` failed (`view does not contain "noise: landings and decisions"`). The two changed tests also failed on the base (`"2 hive events"`, `"noise: decisions only"`). All three pass on the task branch. |
| V2 | fail | The three level names and the `n` cycle match the board. The frame is exact at 120 x 40 and 200 x 57. The level rules hide `blocked` at the default level and at `decisions only`, and hide `quiet` at the default level. See below. |
| V3 | pass | `git diff origin/dev...HEAD --stat` shows only the source table files and the evidence. See the note. |
| V4 | pass | `gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race ./...` exit 0, with no FAIL. |

## V2: level rules

The evidence (V1.md) gives these rules:

| Level | Inbox kinds shown |
| --- | --- |
| `all` | all seven kinds |
| `landings and decisions` (default) | `message`, `completed`, `verified` |
| `decisions only` | `message` |

Problems:

1. `blocked` is hidden at both filtered levels. The supervisor writes `blocked` when herdr reports agent status `blocked`. `reference/herdr/errors.md` (`agent_blocked`) defines this status as an agent that waits at an approval or question dialog. That is the clearest decision event in the inbox, but `decisions only` hides it and shows only `message`.
2. The board shows the default level `● landings and decisions` with this feed: `✓ qmeter 52 verified`, `▶ 53 worker started`, `· steno quiet`, `▶ bench download 41%`, `✓ qmeter 51 landed`, `✓ steno 54 landed`. The board thus shows `quiet` and progress events at the default level. The task hides `quiet`, `exited`, `blocked`, and `unresolved` at the default level.
3. `unresolved` (the supervisor lost the pane of an agent) is hidden at the default level. It usually needs an action from the coordinator.

A possible repair, for the manager to decide: `decisions only` = `message`, `blocked`, `unresolved`; `landings and decisions` = those plus `completed` and `verified` (and `quiet` if the board feed is the rule); `all` = all kinds.

## V2: live run

- Built the task binary to a scratch directory. `HIVE_HOME` was a scratch home with a copy of `.hive/README.md` and seven coordinator inbox records, one for each kind. `XDG_CONFIG_HOME` was an empty scratch directory.
- Ran the binary with `HERDR_ENV` and `HERDR_PANE_ID` unset, so it started no coordinator and no supervisor. Sent only `n` and `ctrl+c`. Never opened the Chat input. No `chat send` event is in the scratch log.
- herdr workspace `wQG` (created by the verifier, cwd in scratch), pane 206 x 57. Read with `herdr pane read wQG:p1 --source visible --lines 57`. Closed `wQG` after the run.

Default level (`landings and decisions`), events column:

```text
││ hive events                      │
││ Oct 20 11:41 · verifier-c:       │
││ verified V                       │
││ Oct 20 11:40 · worker-b: landing │
││ C                                │
││ Oct 20 11:39 · worker-a:         │
││ decision D                       │
││                                  │
││ noise: landings and decisions    │
```

After one `n` (`decisions only`):

```text
││ hive events                      │
││ Oct 20 11:39 · worker-a:         │
││ decision D                       │
││                                  │
││ noise: decisions only            │
```

After a second `n` (`all`): all seven records, `unresolved U`, `quiet Q`, `exited E`, `blocked B`, `verified V`, `landing C`, `decision D`, then `noise: all`. A third `n` returns to `landings and decisions` (log: `chat noise` actions in order `decisions only`, `all`, `landings and decisions`).

Display: the board shows a radio row `○ all  ● landings and decisions  ○ decisions only`. The task shows one line `noise: <level>`. At 120 x 40 that line wraps to `noise: landings and` / `decisions`. GAPS.md C1 names only the levels, so this is a note and not a fail reason.

## Frame size

- `TestFramesMatchTerminalSizesOnEveryScreen` passes for Chat at 120x40 and 200x57 (empty data).
- herdr does not resize the pty of a pane in an unfocused workspace (`stty size` stayed 57 206 after a split). Thus the exact sizes used a private tmux server (`tmux -L verify99`, cwd in scratch) at `-x 120 -y 40` and `-x 200 -y 57` with the fixture home, at all three levels. After tab expansion with the last-column clamp, each capture has exactly 40 or 57 rows, and each box row is exactly 120 or 200 cells and ends on a border. The tmux server was killed after the run.

## V3 note

The literal `git diff origin/dev --stat` now also lists files of tasks 93 and 98 and changes in `health_screen.go`, `logs_screen_test.go`, and `doctor/health_test.go`. These come from `origin/dev`, which moved after the branch point. They are not changes of this task. The three-dot diff from the merge base lists only: the brief, evidence V1-V4, `internal/tui/chat_screen.go`, `internal/tui/chat_screen_test.go`, and `testdata/ste-baseline.json`.

`git merge-tree --write-tree origin/dev HEAD` reports a content conflict in `testdata/ste-baseline.json` (both sides add a line). The repair must merge or rebase on `origin/dev`.

## Cleanup

Removed the base worktree, the tmux server, herdr workspace `wQG`, and the scratch home and binary.
