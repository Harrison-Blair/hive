# VERIFIED: 99-chat-noise-levels

Verifier: Claude Opus 5.5 (`claude-opus-5-5`), independent of the worker.
Task commit: `17469a98` (code commit `0bff20f2`). Base: `ef2dde2e` (merge base with `origin/dev`).
Board: `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Chat.dc.html`.

## Repair 1

Verifier: Claude Opus 5.5 (`claude-opus-5-5`), independent of the worker.
Task commit: `0a4f5bb9` (code commit `7af1aeec`, test commit `fcf71744`). Base: `origin/dev` `b4f1d71e`, which is an ancestor of the task branch.

Verdict: **pass**. Each item passes.

| ID | Result | Observation |
| --- | --- | --- |
| Level table | pass | `visibleChatEvents` at `7af1aeec`: `decisions only` = `message`, `blocked`, `unresolved`; `landings and decisions` (default) = those and `completed`, `verified`, `quiet`; `all` = each kind. This agrees with the Repair 1 table. |
| Literal kind sets | pass | `TestChatNoiseLevelsFilterInboxKinds` writes each expected kind set as a literal (for example `[]string{"message", "blocked", "unresolved"}`) and also asserts literal view text per level. |
| Red on pre-repair code | pass | Detached worktree at `HEAD` with `internal/tui/chat_screen.go` from `7af1aeec^`. `TestChatNoiseAndEventFoldingDoNotMarkRecordsRead` failed (`view does not contain "2 hive events"`). `TestChatNoiseLevelsFilterInboxKinds` failed in `landings_and_decisions` (`visible inbox kinds = [completed message verified]; want [blocked completed message quiet unresolved verified]`), `decisions_only` (`= [message]; want [blocked message unresolved]`), and `returns_to_landings_and_decisions`. The `all` subtest passed, as expected, because that rule did not change. All pass on the task branch. |
| Merge of `origin/dev` | pass | `git merge-base --is-ancestor origin/dev HEAD` exits 0. `git merge-tree --write-tree origin/dev HEAD` exits 0. A script compared `testdata/ste-baseline.json` at `HEAD` with `origin/dev` and with the task side `990e2eed`: no row is missing, no count differs, and no other row was added. |
| V1 | pass | Detached worktree at `HEAD` with `internal/tui/chat_screen.go` from `origin/dev`. `TestChatLoadsRecentConversationAndCoordinatorState` and `TestChatNoiseAndEventFoldingDoNotMarkRecordsRead` failed (`"2 hive events"`). `TestChatNoiseLevelsFilterInboxKinds` failed in three subtests (for example `decisions_only`: `= [blocked completed exited message quiet unresolved verified]; want [blocked message unresolved]`). All pass on the task branch. |
| V2 | pass | Live run below. Each level shows exactly the kinds of the Repair 1 table. |
| V3 | pass | `git diff origin/dev --stat` lists only the brief, evidence `V1.md`-`V4.md` and `VERIFIED.md`, `internal/tui/chat_screen.go`, `internal/tui/chat_screen_test.go`, and `testdata/ste-baseline.json` (1 line added). |
| V4 | pass | `gofmt -l` and `go mod tidy -diff` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...`: `0 issues.`, exit 0. `go test -race ./...` exit 0, each package `ok` (for example `internal/tui 16.436s`). |

### Repair 1: V2 live run

- Built the task binary to the verifier scratch directory. `HIVE_HOME` was a scratch home with a copy of `.hive/README.md` and seven coordinator inbox records, one of each kind: `message`, `completed`, `verified`, `blocked`, `exited`, `quiet`, `unresolved` (text `Fixture <kind> <letter>`). `XDG_CONFIG_HOME` was an empty scratch directory.
- Before the run: one `hive supervise` process (PID `1288721`, not started by the verifier). Recorded the tab list.
- Created herdr workspace `wQW` (tab `wQW:t1`, pane `wQW:p1`, 57 rows, cwd in scratch). Not in `wBC`.
- Ran `HIVE_HOME=… XDG_CONFIG_HOME=… env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH ./hive` in `wQW:p1`. `herdr pane process-info` gave TUI PID `1624562`. The status line showed `0 live agents` and the Chat header `not live · no coordinator`. Sent only `n` and `ctrl+c`. Never opened the Chat input and sent no chat.
- Read each level with `herdr pane read wQW:p1 --source visible --lines 57`.

Default level, events column:

```text
││ hive events                      │
││ Oct 20 06:07 · fixture-          │
││ unresolved: Fixture unresolved U │
││ Oct 20 06:06 · fixture-quiet:    │
││ Fixture quiet Q                  │
││ Oct 20 06:04 · fixture-blocked:  │
││ Fixture blocked B                │
││ Oct 20 06:03 · fixture-verified: │
││ Fixture verified V               │
││ Oct 20 06:02 · fixture-          │
││ completed: Fixture completed C   │
││ Oct 20 06:01 · fixture-message:  │
││ Fixture message M                │
││                                  │
││ noise: landings and decisions    │
```

Chat column: `── 6 hive events ──`. `exited` is hidden.

After one `n`:

```text
││ hive events                      │
││ Oct 20 06:07 · fixture-          │
││ unresolved: Fixture unresolved U │
││ Oct 20 06:04 · fixture-blocked:  │
││ Fixture blocked B                │
││ Oct 20 06:01 · fixture-message:  │
││ Fixture message M                │
││                                  │
││ noise: decisions only            │
```

Chat column: `── 3 hive events ──`.

After a second `n`:

```text
││ hive events                      │
││ Oct 20 06:07 · fixture-          │
││ unresolved: Fixture unresolved U │
││ Oct 20 06:06 · fixture-quiet:    │
││ Fixture quiet Q                  │
││ Oct 20 06:05 · fixture-exited:   │
││ Fixture exited E                 │
││ Oct 20 06:04 · fixture-blocked:  │
││ Fixture blocked B                │
││ Oct 20 06:03 · fixture-verified: │
││ Fixture verified V               │
││ Oct 20 06:02 · fixture-          │
││ completed: Fixture completed C   │
││ Oct 20 06:01 · fixture-message:  │
││ Fixture message M                │
││                                  │
││ noise: all                       │
```

Chat column: `── 7 hive events ──`. A third `n` returned to `noise: landings and decisions` with `── 6 hive events ──`.

### Repair 1: cleanup

- Sent `ctrl+c` twice to `wQW:p1` after `process-info` again gave PID `1624562`. `kill -0 1624562` then failed: the TUI exited. Closed tab `wQW:t1`, the ID that herdr returned at create. `herdr workspace get wQW` then gave `workspace_not_found`.
- After the run: `pgrep -af 'hive supervise'` gave only PID `1288721`, the process that existed before. The tab list had no tab that was not in the list before.
- Note for the manager: tab `wBC:t12` (label `tui-idle-cpu`) was in the list before and was not in the list after. The verifier did not start it, did not send it a key, and did not close it. Another agent closed it during the run.
- Removed the scratch worktree with `git worktree remove`.

## Initial verification

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
