# VERIFIED — 92-agents-board-layout

Verifier: `t92-verifier-1` (Claude Opus 5.5). Task head: `e1c2ed53`. Base for red runs: `598f2732`.
Board: `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`.

Verdict: **fail**. Two items fail: V2 and the R3 golden record. Each other item passes.

| ID | Check | Result |
| --- | --- | --- |
| V1 | R3 tests red on base, green on task | pass |
| V2 | Live captures match the board; each difference listed with its cause | **fail** |
| R3 golden record | Each golden file written from the board or checked line by line against it, and the method recorded | **fail** |
| V3 | `git diff origin/dev --stat` | pass |
| V4 | Gate | pass |
| Scope: ghost rule | Ghosts only for open tasks | pass |
| Scope: supervisor row | Live supervisor shows status, pid, age; cause recorded | pass |
| Scope: style rules | No background; selected item honey bold with `▶` | pass |
| Scope: nudge edge rule | Supervisor to coordinator or manager with an unread nudged record | pass |
| Frame size | Exactly w x h at 120 x 40 and 200 x 57 | pass |

## V1

I made a detached worktree at `598f2732` in my scratch folder. I copied the task tests and the goldens into it. The tests did not compile on the base (`agentNode.pid`, `explorerLines`, `graphOffset` are new). I removed only these three references, as the worker did. Red run on the base:

- `TestAgentsBoardGoldenFrames`: all six frames show the `SCREENS` sidebar.
- `TestAgentsBoardDetailShowsTaskStatusAndNextTask`: omits `task 10-alpha · assigned`, `next 11-next`, no `?`.
- `TestAgentsGhostsRequireAnOpenTask`: `ghost-cancelled` and `ghost-verified-unassigned` visible without history.
- `TestAgentsSupervisorRowUsesLifecycleClaimPIDAndStartAge`: `unknown/?`, want `working/50m`.
- `TestAgentsNudgeEdgesRequireUnreadNudgedRoleInbox`: edge count 0, want 1 (manager and coordinator cases).
- `TestAgentsGraphArrowSelectionMovesBetweenPodsAndScrolls`: right selected `supervisor`, want `manager-alpha`.
- `TestAgentsBoardTaskShortcutOpensSelectedTask`: screen index 1, want 4.
- `TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` passes on the base (negative test). Mutation on the task code: I removed the coordinator/manager role filter. The test failed: `graph has a nudge edge for a worker or verifier`.
- Ghost rule mutation: I set each `verified` task to open. The test failed: `closed task ghost "ghost-verified-unassigned" is visible`.

Green on the task branch: all eight tests `PASS`, `ok github.com/Harrison-Blair/hive/internal/tui`.

## V2 — live check

Setup: binary built from `e1c2ed53`. Scratch home: a copy of `.hive/README.md`, `.hive/state`, and `.hive/spec` of the real home, set with `HIVE_HOME`. A temp `XDG_CONFIG_HOME`. Each start used `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH`. The pane is 207 x 57. I set the tty size with `stty cols W rows H` before each start. I never opened the Chat input and never pressed Enter on Chat.

Observations (`herdr pane read --source visible --format text`):

- 200 x 57 and 120 x 40: each of TREE, GRAPH, TABLE is exactly 57 / 40 lines, and no line inside the frame is wider than 200 / 120. Line 3 has a stray `│` past the frame edge (column 206 / 129). This is old Chat-screen paint outside the `stty` area of the wider pane, not TUI output.
- Footer: `not in herdr: no coordinator or supervisor started`. No new `hive supervise` process. No coordinator tab.
- Header, explorer tree with working counts, compact screen list, `TREE [ GRAPH ] TABLE` tabs, and edge legend are present.
- Pods and node boxes are drawn with box-drawing characters. Ghost nodes have `┄` borders.
- Right x3, Down at 120 x 40: selection `▶ ● t62-worker-1` in the stenographer pod. The graph scrolls left to keep it in view. Detail strip: `parent manager-stenographer · model openai-codex/gpt-6-luna:xhigh · task 62-hub-files · assigned · pane wRD:p1 · next 66-hub-docs-acceptance`.
- `t` opens Tasks with `> 62-hub-files` selected and its detail.

**Fail reason.** The brief says: "Edges join the coordinator to each manager". R1 says: GRAPH draws edges "as on the board". The board draws a delegation bus above the pods that joins the coordinator to each manager box, and draws the nudge edge as a dashed line labelled `nudges` between boxes. The task draws these edges as a text list below the pods:

```text
coordinator ──▶ manager-hive · inferred
coordinator ──▶ manager-qmeter · inferred
coordinator ──▶ manager-stenographer · inferred
supervisor ┄┄▶ manager-hive · nudge
```

The V2 evidence of the worker does not list this difference or its cause. V2 requires a list of each difference with its cause. These other differences from the board are also not listed: pod title `┌─ home ─` against `≡ home`; harness `claude` against `cc`; detail strip with no border; header `supervisor fresh` and `needs_user 0` against `supervisor ●` and `needs you 0`; key bar `h history: off` against `h hide stopped`. Some of these can come from decisions or T3; the evidence must say which.

Also seen (not a fail reason): in the hive pod, the connector between sibling workers is a lone `├` with no stem from the manager.

## R3 golden record

The brief says: "Write each golden file by hand from the board, or check each line of the first output against the board before you save it. Record how in the evidence." The evidence has only the words "hand-checked golden frames" in V1.md. It does not say which method was used, or how each line was checked against the board. The golden files contain the same unlisted board differences as the live output (for example the text edge list below the pods). This agrees with a capture of the output, not with a check against the board. I cannot confirm from the evidence that the goldens were checked against the board.

The golden test itself checks the frame size: 40 x 120 and 57 x 200 display cells on each line. I measured each golden file independently: all lines are exactly 120 or 200 cells wide, with 40 or 57 lines.

## Scope changes

- Ghost rule: in the scratch copy, I computed the open-task ghosts independently from the task, goal, and agent records. Only `t90-worker-1` (task 90, `completed`) and `t92-worker-1` (task 92, `completed`) qualify. 256 other stopped workers and verifiers do not qualify. The live TREE and GRAPH show exactly these two ghosts. Code: `created`, `assigned`, `completed` are open; `verified` is open only when it is not landed and an active goal names it in a wave.
- Supervisor row: live GRAPH shows `● supervisor` / `pid 1288721 · 1h`. PID 1288721 is the running `hive supervise`. TREE shows `supervisor · working · 1h`. The cause is recorded in V2.md (lifecycle `live` record has no `agent_status`, so the generic mapper gave `unknown`; no claim PID fallback; age used only `status_changed_at`).
- Style rules: `selectedStyle = lipgloss.NewStyle().Foreground("#f5b942").Bold(true)` with no background in `internal/tui/shell.go:27`. The Agents screen renders the selected explorer row, graph node, tree row, and table row as `selectedStyle.Render("▶ " + …)`. No `Background(` call is in `internal/tui`.
- Nudge edge rule: the real copy has no unread record with `nudged_at` in the coordinator or manager inboxes, and the GRAPH has no nudge edge. I added two scratch records: an unread nudged record for `manager-hive`, and one for `t62-worker-1`. The GRAPH then showed `supervisor ┄┄▶ manager-hive · nudge` and no edge to `t62-worker-1`.

## V3

`git diff origin/dev --stat`: `internal/tui/agents.go`, `agents_test.go`, `fix_test.go`, `shell.go`, `shell_test.go`, six goldens under `internal/tui/testdata/agents-board/`, `testdata/ste-baseline.json` (+1), and the brief copy and evidence of this task. `origin/dev` is an ancestor of the task head. Pass.

## V4

From the worktree root: `gofmt -l` printed nothing. `go mod tidy -diff` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race ./...` exit 0 (`internal/tui` ok).

## Out of scope

At 120 x 40, the Tasks screen opened by `t` with a task detail looks taller than the frame: the main panel continues 3 lines below the bottom of the sidebar. This task does not change `task_screen.go`, and its `shell.go` change applies only to the Agents screen.

## Resources

| Resource | ID | Result |
| --- | --- | --- |
| Tab in my workspace `wRF` | `wRF:t2`, root pane `wRF:p2`, shell PID 2019008 | Closed |
| Split pane | `wRF:p3` | Closed (the split ratio could not reach 200 columns) |
| TUI processes | PIDs 2019608, 2023682 (200 x 57), 2028247 (120 x 40) | Each stopped with Ctrl+C; PID checked first |

I did not use workspace `wBC`. I removed the scratch home, binary, and base worktree after the checks.
