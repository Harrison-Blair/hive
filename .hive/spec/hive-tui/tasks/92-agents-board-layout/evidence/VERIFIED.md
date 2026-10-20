# VERIFIED — 92-agents-board-layout

Verifier: `t92-verifier-1` (Claude Opus 5.5). Task head: `e1c2ed53`. Base for red runs: `598f2732`.
Board: `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`.

Verdict of the first verify (`e1c2ed53`): **fail**. Two items failed: V2 and the R3 golden record. See [Repair 1](#repair-1) and [Repair 2](#repair-2) for the verify of each repair.

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

## Repair 1

Verifier: `t92-verifier-2` (Claude Opus 5.5). Task head: `45770aec`. `origin/dev`: `8ee27847`.

Verdict: **fail**. Two items fail: the nudge-edge negative test no longer detects its fault, and the list of remaining board differences is not complete.

| ID | Check | Result |
| --- | --- | --- |
| R1.1 bus | Delegation bus above the pods joins the coordinator to each manager box | pass |
| R1.1 nudges | Dashed `nudges` line between the two boxes | pass (same pod); see note for cross-pod |
| R1.1 edge list | No text list of edges below the pods | pass |
| R1.1 stem | Stem from the manager box joins sibling workers | pass, with one defect (dangling stem) |
| R1.2 / V2 differences | Each remaining difference listed with its cause | **fail** |
| R1.3 golden method | Method recorded; goldens checked against the board | pass (method recorded in V2.md) |
| R1.4 merge of 90 | Task 90 tests pass; no refresh change in `shell.go` lost | pass |
| R1.6 live V2 | Scratch home, own tab, herdr variables removed | pass |
| R1.7 merge-tree | `git merge-tree --write-tree origin/dev HEAD` exit 0 | pass (`c658238e`) |
| V1 repair tests | Red on pre-repair `ee35be40`, green on task | pass |
| First-verify passes still hold | Ghost rule, supervisor row, style, nudge rule, frame size | **fail** (nudge-rule test) |
| V3 | `git diff origin/dev --stat` | pass |
| V4 | Gate | pass |
| Frame size | Exactly 120 x 40 and 200 x 57 | pass |

### Fail 1: the nudge-edge negative test is vacuous

`TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` (`internal/tui/agents_test.go`) checks that the graph does not contain `supervisor ┄┄▶ worker-alpha · nudge`. Repair 1 removed that text form, so the test cannot fail. Mutation: I removed `|| (record.Role != "coordinator" && record.Role != "manager")` from the inbox loop in `agents.go`. The test passed. A temporary render of the same fixture then showed a dashed nudge route and the label `nudges` from the supervisor to `worker-alpha`. In the first verify, the same mutation failed this test. The test must look for the new dashed route or the `nudges` label (as `TestAgentsNudgeEdgesRequireUnreadNudgedRoleInbox` does). I restored `agents.go` after the run.

### Fail 2: board differences that are not listed

I compared each region of the board with the live captures and the goldens. V2.md does not list these differences or a cause for them:

- Key bar: the board has `←↑↓→ move on graph`, `enter peek pane`, `t open task`, and the order `… a attach · h hide stopped · t open task`. The task has `↑/↓←/→ move`, `enter peek`, `t task`, and `h hide stopped` before `enter`. (`ask in Chat` is from R2 / G12.)
- Detail strip: the task starts with the label `DETAIL`. The board has no label.
- Pod title: the board puts `≡ home` on the top border of the pod. The task puts it on its own row inside the pod.
- Node second line: the board has `cc you · now` and `pi 47-brief-copy`. The task has `cc · you · now` and `pi · 10-alpha`.
- Explorer: the board indents the rows with no connectors. The task draws `├─`, `└─`, and `┄┄` for an inferred manager.
- Selected node: the board uses a 2 px honey border and a fill. The task uses `▶` and honey bold text. The cause is the style-rule scope change, but V2.md does not say so.
- Dangling stem: in the live GRAPH, `manager-qmeter` has a `┴` on its bottom border with no visible child below it. `graphPod` adds `┴` when `s.children(key)` is not empty, and `children` also counts stopped workers that history-off hides.
- Bus ends: the bus starts and ends with `┬`. The board line ends at the first and last stem.
- TREE view: the last child in a project uses `├─`, for example `├─ × t92-worker-2` and `├─ ● t107-worker-1`. The explorer uses `└─` correctly for the same rows.

Note (not a fail reason): a nudge to a manager in a different pod uses a route below the pods. Its vertical parts are solid `│`, which the legend uses for delegation. It also writes over pod border cells (`├┄│`, `┄┤`).

### V1 for the repair tests

In a detached worktree at `ee35be40`, I copied the task `agents_test.go` and the six goldens. `ee35be40` does not build alone (`shell.go` calls `activeHeaderLine`). I added a temporary adapter `activeHeaderLine() { return m.headerLine() }`, as the worker did. Red: `TestAgentsScreenRendersTopologyTabsAndDetail`, `TestAgentsHeaderUsesBoardLabels`, all six `TestAgentsBoardGoldenFrames` frames, `TestAgentsBoardDetailShowsTaskStatusAndNextTask`, `TestAgentsNudgeEdgesRequireUnreadNudgedRoleInbox` (manager and coordinator cases), and `TestAgentsGraphConnectsSiblingWorkersToManagerStem` failed. Green: each passes on `45770aec`.

### First-verify passes

- Ghost rule: mutation (each `verified` task open) fails `TestAgentsGhostsRequireAnOpenTask` with `closed task ghost "ghost-verified-unassigned" is visible without history`. Holds.
- Supervisor row: live GRAPH shows `● supervisor` / `pid 2107907 · 27m`; PID 2107907 is the running `hive supervise`. Holds.
- Style rules: the selected node shows `▶`; no `Background(` was added. Holds.
- Nudge edge rule: the code filter is present, but its negative test does not detect its removal (Fail 1).
- Frame size: each golden has exactly 40 or 57 lines, and each line is exactly 120 or 200 cells. The live frames are exactly 40 / 57 lines with the bottom border on line 37 / 54 and the key bar on the last line. No frame line is wider than 120 / 200. Line 3 has a stray `│` past the frame (column 129 / 206). This is old paint in the 207-column pane outside the `stty` area.

### Merge of task 90

`git diff origin/dev -- internal/tui/shell.go` keeps the cache check `screenViewCacheValid` and its width, height, and index keys. The task adds only the Agents explorer, the panel with no title, and `activeHeaderLine`. `refresh.go` adds `state/inbox` to the Agents refresh paths. `TestIdleTickDoesNotReloadUnchangedData`, `TestIdleTickReloadsChangedScreenOnce`, `TestHeaderRefreshReusesUnchangedScreenRender`, and `TestMinuteRefreshUpdatesOnlyTheHeader` pass. Note: `renderAgentsExplorer` is not cached, so each view of the Agents screen builds the explorer again. `agentsHeaderLine` does not show `headerErr`.

### Live check

Binary built from `45770aec`. Scratch home: a copy of `.hive/README.md`, `.hive/state`, and `.hive/spec` of the real home, set with `HIVE_HOME`, and a temp `XDG_CONFIG_HOME`. Each start used `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH ./hive` after `stty cols W rows H`. I never opened the Chat input and never pressed Enter on Chat.

- 200 x 57: TREE, GRAPH, TABLE captured. The GRAPH has the bus with stems to `manager-hive`, `manager-qmeter`, and `manager-stenographer`, the sibling stem through `t104-verifier-1` (ghost), `t104-worker-2`, and `t107-worker-1`, and no text edge list. No unread nudged record was in the copy, so no nudge line showed live; the same-pod `┄┄ nudges ┄┄` line shows in the worker capture `live-207x57-graph.txt`.
- Right x3, Down: `▶ × t104-verifier-1`, detail `parent manager-stenographer · model claude-opus-5-5 · task 104-relabel-memory · assigned · pane wRR:p1 · next ?`. `t` opens Tasks at `> 104-relabel-m…`.
- 120 x 40: the graph clips at the panel edge. The detail strip wraps to two rows inside its border.
- Footer: `not in herdr: no coordinator or supervisor started`. The only `hive supervise` process before and after was PID 2107907, which I did not start. No coordinator tab was made.

### Gate

`gofmt -l` printed nothing. `go mod tidy -diff` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race -count=1 ./...` exit 0.

### Resources

| Resource | ID | Result |
| --- | --- | --- |
| Tab in my workspace `wRW` | `wRW:t2` (`t92-v2-live`), root pane `wRW:p2`, shell PID 2242398 | Closed after a label and ID check; shell gone |
| TUI process 200 x 57 | PID 2243941 in `wRW:p2` | Stopped with Ctrl+C after a PID check |
| TUI process 120 x 40 | PID 2247276 in `wRW:p2` | Stopped with Ctrl+C after a PID check |

I did not use workspace `wBC`. I removed the scratch home, the binary, the pre-repair worktree, and the temporary test file. I restored `agents.go` after each mutation; `git status` was clean.

## Repair 2

Verifier: `t92-verifier-3` (Claude Opus 5.5). Task head: `3e6d1b54`. `origin/dev`: `8ee27847` (ancestor of the head; it did not move). Repair 1 tip for red runs: `45770aec`.
Board: `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`.

Verdict: **fail**. One item fails: item 3. V2 does not list five board differences, and it gives no cause for them.

| ID | Check | Result |
| --- | --- | --- |
| R2.1 nudge test | Filter removed: `TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` fails; filter restored: passes | pass |
| R2.2 drawing faults | Each of the four faults has a literal-text test that fails on `45770aec` and passes on `3e6d1b54` | pass |
| R2.3 / V2 differences | Each other board difference fixed or listed in V2 with a cause | **fail** |
| R2.4 goldens | Goldens written again; method recorded in V1.md and V2.md | pass |
| R2.5 brief copy | STE rewording applied again (headings `Fix 1` / `Fix 2` for `Repair 1` / `Repair 2`); meaning kept | pass |
| R2.6 live V2 | Scratch home, own tab, herdr variables removed | pass |
| R2.7 merge-tree | `git merge-tree --write-tree origin/dev HEAD` exit 0 (`e7ecbb11`) | pass |
| First-verify passes | Ghost rule, supervisor row, style, nudge rule, task 90 idle CPU | pass |
| Frame size | Exactly 120 x 40 and 200 x 57 | pass |
| V3 | `git diff origin/dev --stat` | pass |
| V4 | Gate | pass |

### R2.1 nudge test

I ran the mutation in a detached scratch worktree at `3e6d1b54`. I removed `|| (record.Role != "coordinator" && record.Role != "manager")` from `agents.go:144`. `go test ./internal/tui -run TestAgentsNudgeEdges -count=1`: `TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` FAIL, `agents_test.go:725: drawn graph has a nudge route for a worker or verifier`, and the graph has two `nudges` labels. `TestAgentsNudgeEdgesRequireUnreadNudgedRoleInbox` PASS. I restored the file (`git status` clean). The same run then passed.

### R2.2 drawing faults

I copied the task `agents_test.go` and the six goldens into a detached worktree at `45770aec`. `go test ./internal/tui -count=1`:

- `TestAgentsGraphHasNoStemForManagerWithOnlyHiddenChildren` FAIL: `manager-beta draws a stem with no visible child: └───────────────┴───────────────┘`.
- `TestAgentsGraphDelegationBusUsesOpenEndGlyphs` FAIL: `delegation bus at coordinator = '┬', want '┌'` and `at manager-beta = '┬', want '┐'`.
- `TestAgentsTreeLastVisibleChildUsesBottomConnector` FAIL: `├─ ● worker-alpha · working · 4m`.
- `TestAgentsGraphNodeSubtitlesHaveNoExtraSeparator` FAIL: omits `cc idle · 3m`, `pi 10-alpha`; has `cc · idle`, `pi · 10-alpha`.
- Also red: `TestAgentsScreenRendersTopologyTabsAndDetail`, `TestAgentsBoardGoldenFrames` (all six), `TestAgentsBoardKeyHintsMatchTheBoard`, `TestAgentsGraphPodTitleSharesTopBorder`, `TestAgentsExplorerUsesIndentationWithoutConnectors`.

Each test passes on `3e6d1b54`. The bus ends agree with the board: the board line runs from `left: 95px` to `left: 776px`, with stems at 95, 315, 540, and 775, so the ends are corners.

Note (not a fail reason): no test fails when the `DETAIL` label comes back. `TestAgentsBoardDetailShowsTaskStatusAndNextTask` passes on `45770aec`, because `worker-alpha · parent manager-alpha` is a substring of `DETAIL worker-alpha · parent …`. The goldens catch the label. V1.md says that the detail test found the label. That is not correct.

### R2.3 board differences that are not listed

I compared each board region with the goldens and with my live captures (`verify-repair2-200x57-graph.txt`, `verify-repair2-120x40-graph.txt`). V2.md "Fix 2 live check" lists the key bar, pod titles, subtitles, detail label, explorer, supervisor, G4 `inferred`, the T3 nudge rule, clipping, and the selection style with their causes. It does not list these differences, and it gives no cause for them:

1. **Stem breaks at the pod border.** On the board, each delegation stem runs without a break from the bus (`top: 8px`) through the pod border (`top: 34px`) to the node box (`top: 60px`). The task draws `─` where the stem crosses the pod top border. Example, home pod: line 5 has `│` at panel column 13, line 6 `┌─ ≡ home ─────` has `─` at column 13, line 8 has `│` again. The same break shows in each pod in the goldens and the live captures.
2. **Pod corners.** The board pods have `border-radius: 3px` (rounded, as the panels and the detail strip that the task draws with `╭╮╰╯`). The task draws pods with `┌┐└┘`.
3. **Node box padding.** The board node boxes have `padding: 4px 8px`. The task puts the text against the left border, and in long rows against the right border too: `│● t107-worker-2`, `│cc 107-hub-no-background│`.
4. **Header layout.** The board header has groups with gaps and the time at the right edge (a `flex: 1` spacer before `02:46`). The task joins the groups with `·` and puts the time inline: `… · needs you 0 · 14:28`.
5. **Detail strip separators.** The board separates the detail fields with gaps (`gap: 24px`). The task joins them with `·`.

Each one must agree with the board, or V2 must give its cause (a decision, a T3 rule, a brief rule, or a terminal limit).

### First-verify passes and task 90

All in the scratch worktree at `3e6d1b54`, each file restored after the run:

- Ghost rule: `open = activeGoalTasks[task.ID]` to `open = true` fails `TestAgentsGhostsRequireAnOpenTask` (`closed task ghost "ghost-verified-unassigned" is visible without history`, and in the explorer). Holds.
- Task 90, render cache: `if true` for both `View` cache checks (`tui.go:420`, `shell.go:156`) fails `TestHeaderRefreshReusesUnchangedScreenRender` (`header refresh rendered screen 2 times`) and `TestMinuteRefreshUpdatesOnlyTheHeader` (`minute refresh rendered screen 2 times`). A mutation of `shell.go:156` alone does not fail them, because the `View` cache in `tui.go` also holds the screen.
- Task 90, idle reload: `headerChanged := true` (`tui.go:242`) fails `TestIdleTickDoesNotReloadUnchangedData` (`1 header loads`) and `TestIdleTickReloadsChangedScreenOnce` (`1 header loads and 1 screen loads; want 0 and 1`).
- Unchanged code: the four task 90 tests and the ghost, supervisor, golden, arrow-selection, and `t` tests pass.
- Supervisor row: live GRAPH shows `● supervisor` / `pid 2107907 · 50m`. PID 2107907 is the running `hive supervise`. Holds.
- Style: the selected node shows `▶`. `git diff 45770aec HEAD` adds no `Background(`. Holds.

### Live check

Binary built from `3e6d1b54` in my scratch folder. Scratch home: a copy of `.hive/README.md`, `.hive/state`, and `.hive/spec` of the real home in my scratch folder, set with `HIVE_HOME`. A temp `XDG_CONFIG_HOME`. Each start: `stty cols W rows H`, then `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH ./hive`. I never opened the Chat input and never pressed Enter.

- 200 x 57: TREE, GRAPH, and TABLE have 57 lines. The top border is 200 cells. The key bar is the last line. The bus has `┌ … ┬ … ┬ … ┐`. `manager-qmeter` has no stem. TREE uses `└─` for each last child. Subtitles: `cc you · now`, `cc idle · 4m`, `pi 92-agents-board-layout`. The detail strip has no `DETAIL` label.
- Selection: Down in GRAPH selects `▶ ● t92-verifier-3`. Detail: `t92-verifier-3 · parent manager-hive · model claude-opus-5-5 · task 92-agents-board-layout · completed · pane wS1:p1 · next 102-selection-marker`. `t` opens Tasks at `92-agents-board-layout`.
- 120 x 40: each view has 40 lines with a 120-cell top border. Right x3 scrolls the graph to `▶ ○ manager-stenographer`.
- Text reads show stray `│` cells past the frame (column 206 at 200 x 57; column 129 on lines 3, 4, 9, 13, 14 at 120 x 40, after a Chat to Agents change). An ANSI read of the same lines ends at the frame border. The pane is 207 columns wide. I think that these cells are terminal leftovers outside the `stty` area. Each frame line is exactly 120 or 200 cells.
- Footer: `not in herdr: no coordinator or supervisor started`. The only `hive supervise` process before and after was PID 2107907, which I did not start. No coordinator tab was made.

### V3 and gate

`git diff origin/dev --stat`: `internal/tui/` (`agents.go`, `agents_test.go`, `fix_test.go`, `header.go`, `refresh.go`, `shell.go`, `shell_test.go`, `tui.go`, six goldens), `testdata/ste-baseline.json` (+1), and the brief copy and evidence of this task. Pass.

Gate: `gofmt -l` printed nothing. `go mod tidy -diff` printed nothing, exit 0. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race -count=1 ./...` exit 0.

### Resources

| Resource | ID | Result |
| --- | --- | --- |
| Tab in my workspace `wS1` | `wS1:t2` (`t92-v3-live`), root pane `wS1:p2`, shell PID 2356204 | Closed after a label and ID check; shell gone |
| TUI process 200 x 57 | PID 2357068 in `wS1:p2` | Stopped with Ctrl+C after a PID check |
| TUI process 120 x 40 | PID 2361315 in `wS1:p2` | Stopped with Ctrl+C after a PID check |

I did not use workspace `wBC`. Tab `wS1:t1` stays open.

## Repair 3

Verifier: `t92-verifier-4`. Task head: `72d51057`. Repair 2 tip: `3e6d1b54`. Board: `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`.

Result: **fail**. R3.1 (unbroken stem) fails, and the V2 region table lists the pod title as "same" when the task moved it off the border.

| Item | Result |
| --- | --- |
| R3.1 stem is one line from the bus to the manager node, junction at the pod border | fail |
| R3.2 rounded pod corners | pass |
| R3.3 one cell of node padding | pass |
| R3.4 header groups with spaces, time at the right edge | pass |
| R3.5 detail fields with spaces | pass |
| R3 tests fail on `3e6d1b54` and pass on `72d51057` | pass |
| V2 region table | fail (pods row) |
| Earlier passes (nudge mutation, ghost rule, task 90 idle CPU) | pass |
| Exact frame at 120 x 40 and 200 x 57 | pass |
| V3 diff scope | pass |
| V4 gate | pass |
| `git merge-tree --write-tree origin/dev HEAD` | exit 0 |

### R3 red and green

I copied `internal/tui/agents_test.go` of `72d51057` into a scratch worktree at `3e6d1b54` and ran:

```sh
go test ./internal/tui -run 'TestAgentsGraphDelegationStemJoinsPodBorder|TestAgentsGraphPodsHaveRoundedCorners|TestAgentsGraphNodeHasOneCellOfPadding|TestAgentsHeaderUsesSpacesAndRightAlignsTime|TestAgentsDetailFieldsUseSpaces|TestAgentsGraphPodTitleIsInsideBorder' -count=1
```

Each of the six tests failed for its point: `delegation stem at the pod border = 'o', want '┬'`, `pod corner … = '┌', want '╭'` (4 corners), `coordinator node has no left padding: │▶ ● coordinator`, `header groups are not separated by spaces: ⬢ hive · 2 projects · …`, `detail = "worker-alpha · parent …"`, and the title test. On `72d51057`, `go test ./internal/tui -count=1` passes.

### R3.1 stem: fail

The brief: "The delegation stem is one line from the bus to the manager node." The board draws each stem as one div from `top: 8px` (bus) to `top: 60px` (node top), through the pod border at `top: 34px`.

The task draws `┬` on the pod top border, then breaks the stem on the next two rows. The pod title row is under the stem column, and in a project pod the `┄ inferred` label replaces the stem on the row after it. Golden `graph-120x40.golden`, home pod, panel column 13:

```text
│             ┌────────────────…   bus
│             │                    stem
│ ╭───────────┬──────────╮         pod border: ┬ (good)
│ │ ≡ home               │         title row: blank at column 13 (break)
│ │           │          │         stem
│ │ ┌─────────┬────────┐ │         node
```

The alpha pod has `≡ alpha · Alpha goal for long graph` across the stem column. Live (`verify-repair3-200x57-graph.txt`, hive pod): `╭───────────────┬───────────────╮`, then `│ ≡ hive · Hive TUI dashboard   │` (text at the stem column), then `│          ┄ inferred           │` (label at the stem column), then the node `┌─────────────┬─────────────┐`. The stem has a gap of two rows in each project pod.

`TestAgentsGraphDelegationStemJoinsPodBorder` checks only the border glyph (`lines[2]`). It does not check the rows between the border and the node, so it passes with the break.

### V2 region table: pods row fails

The region table says Pods: "Same … Titles sit at the left inside each pod." That is not the board. The board places each title on the pod top border: the title div has `top: 24px`, line height 20px, and an opaque background, and the pod border is at `top: 34px`, so the title is centered on the border line (a legend). The Repair 1 verification recorded this ("the board puts `≡ home` on the top border of the pod"), and Repair 2 put the title on the border with `TestAgentsGraphPodTitleSharesTopBorder`. Repair 3 moved the title to its own row inside the pod and changed that test to `TestAgentsGraphPodTitleIsInsideBorder`, which now requires the title to be off the border. V1.md says "The board places the title at the left inside the pod." The board does not.

This difference is not listed with a cause, so it is a fault under the Repair 3 rule. It is also the cause of the R3.1 break: the title row sits between the border junction and the node. Note for the repair: on the board the stems of `hive`, `qmeter`, and `stenographer` cross their title text (the stem divs paint after the title divs). A terminal cannot draw both in one cell. Either place the title on the border clear of the stem column, or record the conflict as a terminal limit, and draw the stem without a gap. The `┄ inferred` label also needs a place that does not replace the stem cell.

Each other row of the table agrees with the board or gives a cause I accept: header (spaces, time at the right edge; live counts), explorer, screen list, tabs and legend, nodes (dot, name, harness or PID, task, one-cell padding; selection style from the user rule), edges apart from the stem gap (bus ends, G4 `inferred`, T3 nudge rule, clipping at 120), detail strip (rounded border, `?`, spaces), and key bar.

### Earlier passes

In a scratch worktree at `72d51057`, each file restored with `git checkout` after each run:

- Nudge filter: I removed `|| (record.Role != "coordinator" && record.Role != "manager")` from the inbox loop in `loadAgents` (`agents.go:144`). `TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` failed: `drawn graph has a nudge route for a worker or verifier`. Holds.
- Ghost rule: `open = activeGoalTasks[task.ID]` to `open = true` (`agents.go:127`). `TestAgentsGhostsRequireAnOpenTask` failed: `closed task ghost "ghost-verified-unassigned" is visible without history` and in the explorer. Holds.
- Task 90 render cache: both `View` cache checks to `if true` (`tui.go` and `shell.go:156`). `TestHeaderRefreshReusesUnchangedScreenRender` (`rendered screen 2 times`) and `TestMinuteRefreshUpdatesOnlyTheHeader` (`rendered screen 2 times`) failed. Holds.
- Task 90 idle reload: `headerChanged := true` (`tui.go:242`). `TestIdleTickDoesNotReloadUnchangedData` (`1 header loads and 0 screen loads`) and `TestIdleTickReloadsChangedScreenOnce` (`1 header loads and 1 screen loads; want 0 and 1`) failed. Holds.
- Unchanged code: these six tests pass.

### Live check

Binary built from `72d51057` in my scratch folder. Scratch home: a copy of `.hive/README.md`, `.hive/state`, and `.hive/spec` of the real home, set with `HIVE_HOME`. A temp `XDG_CONFIG_HOME`. I never opened the Chat input and never pressed Enter.

1. In my tab `wS3:t2` (pane `wS3:p2`), with `stty` set to each size and `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH`: PID 2504214 (200 x 57) and PID 2506637 (120 x 40). `/proc/2504214/environ` had no `HERDR_ENV`, `HERDR_PANE_ID`, `HERDR_WORKSPACE_ID`, or `HERDR_SOCKET_PATH`. These reads show the same stem gap. They are not exact frames: the pane is 207 x 57, and after a Tab key the TUI got a resize event and drew itself at 207 x 57. The stray `│` at column 129 in earlier text reads comes from this size mismatch.
2. For exact frames I started a private tmux server from the same pane: `tmux -L t92v4 -f /dev/null new-session -d -s v120 -x 120 -y 40 'env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH -u HERDR_TAB_ID HIVE_HOME=… XDG_CONFIG_HOME=… …/hive'`, and the same with `-s v200 -x 200 -y 57`. Server PID 2513043, TUI PIDs 2513044 (120 x 40) and 2513310 (200 x 57). I sent `2`, then Tab, Tab with `tmux send-keys` and read each view with `tmux capture-pane -p`.

| Size | View | Capture | Lines | Line widths |
| --- | --- | --- | --- | --- |
| 120 x 40 | TREE | [`verify-repair3-120x40-tree.txt`](verify-repair3-120x40-tree.txt) | 40 | 120 max; no line past the edge |
| 120 x 40 | GRAPH | [`verify-repair3-120x40-graph.txt`](verify-repair3-120x40-graph.txt) | 40 | 120 max |
| 120 x 40 | TABLE | [`verify-repair3-120x40-table.txt`](verify-repair3-120x40-table.txt) | 40 | 120 max |
| 200 x 57 | TREE | [`verify-repair3-200x57-tree.txt`](verify-repair3-200x57-tree.txt) | 57 | 200 max |
| 200 x 57 | GRAPH | [`verify-repair3-200x57-graph.txt`](verify-repair3-200x57-graph.txt) | 57 | 200 max |
| 200 x 57 | TABLE | [`verify-repair3-200x57-table.txt`](verify-repair3-200x57-table.txt) | 57 | 200 max |

Each frame line is 120 or 200 cells (the footer and key bar lines are shorter text). The header ends with the time at the right edge: `… supervisor ●  needs you 0 … 14:58`. Pods use `╭╮╰╯`. Nodes have one space of padding: `│ ▶ ● coordinator  │`. Detail: `coordinator  parent ?  model claude-opus-5-5  task ?  ?  pane wBC:pV  next ?`. Key bar: `←↑↓→ move on graph · tab tree/graph/table · enter peek pane · a attach · h hide stopped · t open task · ask in Chat`.

Footer: `not in herdr: no coordinator or supervisor started`. The only `hive supervise` process before and after was PID 2107907, which I did not start. No coordinator tab was made.

### V3, merge, and gate

`git diff origin/dev --stat`: `internal/tui/` (`agents.go`, `agents_test.go`, `fix_test.go`, `header.go`, `refresh.go`, `shell.go`, `shell_test.go`, `tui.go`, six goldens), `testdata/ste-baseline.json` (+1), and the brief copy and evidence of this task. Pass. `git merge-tree --write-tree origin/dev HEAD` exit 0 (`origin/dev` at `8ee27847`).

Gate: `gofmt -l` printed nothing. `go mod tidy -diff` printed nothing, exit 0. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race -count=1 ./...` exit 0.

### Resources

| Resource | ID | Result |
| --- | --- | --- |
| Tab in my workspace `wS3` | `wS3:t2` (`t92-v4-live`), root pane `wS3:p2` | Closed after a label and ID check |
| TUI 200 x 57 in `wS3:p2` | PID 2504214 | Stopped with Ctrl+C after a PID check |
| TUI 120 x 40 in `wS3:p2` | PID 2506637 | Stopped with Ctrl+C after a PID check |
| tmux server `-L t92v4` | PID 2513043, TUI PIDs 2513044 and 2513310 | Stopped with `tmux -L t92v4 kill-server`; all three PIDs gone |

I did not use workspace `wBC`. Tab `wS3:t1` stays open.

## Repair 4

Verifier: `t92-verifier-5`. Task head: `07a5585e` (code at `626cb9f0`). Repair 3 tip: `72d51057`. Board: `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`.

Result: **fail**. The stem repair passes. The pod title repair fails: a title is cut off in the committed golden, and live the `≡` glyph is split away from the project name. The Design element table omits the line and dot colors of the board.

| Item | Result |
| --- | --- |
| R4.1 pod title on the top border with the board glyph | fail |
| R4.2 stem unbroken in each cell from the bus to the manager node | pass |
| R4 tests fail on `72d51057` and pass on the task head | pass |
| R4.3 golden files checked against the board | fail (beta title cut off) |
| Design visual check: worker element table | fail (pods row, colors not listed) |
| Earlier passes (nudge mutation, ghost rule, task 90 idle CPU) | pass |
| Exact frame at 120 x 40 and 200 x 57 | pass |
| V3 diff scope | pass |
| V4 gate | pass |
| `git merge-tree --write-tree origin/dev HEAD` | exit 0 |

### R4 red and green

I made a scratch worktree at `72d51057` and copied `internal/tui/agents_test.go` of the task head into it:

```sh
go test ./internal/tui -run 'TestAgentsGraphPodTitleSharesTopBorder|TestAgentsGraphDelegationStemJoinsPodBorder|TestAgentsGraphNodeCardsStayInsetInLongTitlePods|TestAgentsScreenRendersTopologyTabsAndDetail' -count=1
```

Red on `72d51057`: `home pod title is not on the top border` (the title is on row 2, inside the pod), `delegation stem at row 3 = 'a', want '│'` (the title text is in the stem column), and `alpha node cards have only 4 cells of total pod margin, want at least 6`. Green: `go test ./internal/tui -count=1` passes on the task head.

The stem test now checks each cell from the bus (row 0) to the node top (`manager.row - 1`): `┬` on the bus, the pod border, and the node top, `│` on each other row. Two mutations on the task head, each restored after the run:

- `graphPodLabelLineWithStem`: `cells[center] = '│'` to `' '` (stem removed on the `┄ inferred` row). The test failed: `delegation stem at row 4 = ' ', want '│': … │ ┄ inferred    …`.
- `graphPod`: `lines[1] = graphPodLine(width, width/2, '│')` to `' '` (stem removed on the row below the border). The test failed: `delegation stem at row 3 = ' ', want '│'`.

Live, the stem is unbroken in each pod: bus `┬`, `│`, border `┬`, `│`, `┄ inferred  │`, node top `┬`. R4.2 passes.

### R4.1 and R4.3 pod title: fail

The brief asks for the title on the top border with the board glyph, for example `╭─ ≡ home ───╮`. The task splits the title text around the stem column. The worker records this as a terminal limit and says "The full title remains visible." I accept the split around the stem as a terminal limit. Two outputs do not agree with that cause:

1. **Cut-off title in the committed golden.** `internal/tui/testdata/agents-board/graph-200x57.golden` line 6: `╭─ ≡ beta · Beta──── ┬ goal for long grap╮`. The fixture title is `Beta goal for long graph` (`agents_test.go:1039`). The last `h` is lost and the text touches the corner. The alpha title also touches the corner: `≡ alpha · Alpha── ┬ goal for long graph╮`. `TestAgentsGraphPodTitleSharesTopBorder` checks only the alpha pod, so it does not see the beta loss. A golden with a cut-off title did not get a check of each line against the board (R4.3).
2. **Glyph split from the name, live.** The scratch home at 207 x 57 (`verify-repair4-207x57-graph.txt` line 6, and the worker's own `repair4-live-207x57-graph.txt` and `repair4-agents-graph-207x57.png`): `╭─ ≡─────────── ┬ stenographer ╮`. The `≡` stands alone with border dashes after it, and the name is on the other side of the stem. The cause is in `graphTextAroundStem`: when the text does not fit left of the stem, the split loop finds the only space after `≡` and puts the full name on the right. On the board the title reads `≡ stenographer · hub` as one unit. The terminal limit does not cause this. The pod can be wider, or the title can move clear of the stem.

Other titles: `╭─ ≡ home ─ ┬ ─────────╮` and `╭─ ≡ hive · Hive ┬ TUI dashboard ─╮` agree with the board under the terminal limit. `╭─ ≡ qmeter  ┬ ──────────╮` has two spaces before the stem; this is minor.

### Design visual check

Board PNG: `firefox --headless --no-remote --profile .hive/tmp/t92v5/ff --window-size=1220,780 --screenshot .hive/tmp/t92v5/board.png file://…/Main.dc.html`. It stays at `/home/penguin/source/hive/projects/hive/worktrees/92-agents-board-layout/.hive/tmp/t92v5/board.png`, not in Git.

Screen PNGs: in my pane `wSC:p2` I ran the task binary on a scratch home, read TREE, GRAPH, and TABLE with `herdr pane read wSC:p2 --source visible --format ansi`, changed each to an HTML `<pre>` with a short Python script, and ran the same Firefox command at 1500 x 900. They stay at `.hive/tmp/t92v5/screen-{tree,graph,table}.png`. My render and the worker's PNGs both show a width drift on rows with `≡`, `●`, and `▶` (the browser font draws these glyphs wider than one cell). The text captures have exact cells, so I judged cell positions from the text.

Judgment of the worker's element table (V2, "Fix 4 visual check"):

| Board element | Worker | My judgment |
| --- | --- | --- |
| Header | matches | agree |
| Explorer | matches | agree |
| Screen list | matches | agree |
| Tabs | matches | agree |
| Legend | matches | agree |
| Pods | differs: title split at a word boundary, full title visible | **disagree**: the beta title is cut off in the golden, and the stenographer `≡` is split from its name live (see R4.1). |
| Nodes | overridden by the selection rule | agree for the selection. The table does not list the colors (below). |
| Edges | differs: no nudge record, clipping at 120 | agree for the shape. The table does not list the colors (below). |
| Detail strip | matches | agree. The board border is honey; the task border is the muted line color. |
| Key bar | overridden by G12 | agree |
| **Colors (no row)** | not listed | **difference with no cause.** On the board, pod borders are blue, delegation edges and live node borders are green, the selected node border is honey, the status dots are green, and the harness `cc` is honey and `pi` is blue. The GRAPH ANSI capture has only the muted line color `38;2;61;58;51`, default text, grey for stopped nodes, honey for the selection, and one green dot (header supervisor). The brief style rules forbid a background and set normal text to the default foreground. They do not name line or dot colors. The worker must match the board colors or record the decision that overrides them. |

### Earlier passes

In the scratch worktree at the task head, each file restored with `git checkout` after each run:

- Nudge filter removed (`agents.go:144`): `TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` failed: `drawn graph has a nudge route for a worker or verifier`. Holds.
- Ghost rule `open = true` (`agents.go:127`): `TestAgentsGhostsRequireAnOpenTask` failed: `closed task ghost "ghost-verified-unassigned" is visible without history` and in the explorer. Holds.
- Task 90 render cache, `tui.go:420` and `shell.go:156` to `if true`: `header refresh rendered screen 2 times` and `minute refresh rendered screen 2 times`. Holds.
- Task 90 idle reload, `headerChanged := true` (`tui.go:242`): `unchanged tick started 1 header loads and 0 screen loads; want none` and `changed tick started 1 header loads and 1 screen loads; want 0 and 1`. Holds.
- Unchanged code: the six tests pass.

### Live check

Binary built from `07a5585e`. Scratch home: a copy of `.hive/README.md`, `.hive/state`, and `.hive/spec` of the real home, set with `HIVE_HOME`, and a temp `XDG_CONFIG_HOME`. Each start removed `HERDR_ENV`, `HERDR_PANE_ID`, `HERDR_WORKSPACE_ID`, `HERDR_SOCKET_PATH`, and `HERDR_TAB_ID`; `/proc/2688473/environ` had no `HERDR_ENV`, `HERDR_PANE_ID`, `HERDR_WORKSPACE_ID`, or `HERDR_SOCKET_PATH`. I never opened the Chat input and never pressed Enter in the TUI.

1. Tab `wSC:t2` (`t92-v5-live`), pane `wSC:p2`, 207 x 57: TUI PID 2688473. Capture [`verify-repair4-207x57-graph.txt`](verify-repair4-207x57-graph.txt).
2. Exact frames: from `wSC:p2`, `tmux -L t92v5 -f /dev/null new-session -d -s v120 -x 120 -y 40 'env -u … ./hive'` and `-s v200 -x 200 -y 57`. Server PID 2696132, TUI PIDs 2696136 and 2696142. I sent `2` and Tab with `tmux send-keys` and read each view with `tmux capture-pane -p`.

| Size | View | Capture | Lines | Widest line |
| --- | --- | --- | --- | --- |
| 120 x 40 | TREE | [`verify-repair4-120x40-tree.txt`](verify-repair4-120x40-tree.txt) | 40 | 120 |
| 120 x 40 | GRAPH | [`verify-repair4-120x40-graph.txt`](verify-repair4-120x40-graph.txt) | 40 | 120 |
| 120 x 40 | TABLE | [`verify-repair4-120x40-table.txt`](verify-repair4-120x40-table.txt) | 40 | 120 |
| 200 x 57 | TREE | [`verify-repair4-200x57-tree.txt`](verify-repair4-200x57-tree.txt) | 57 | 200 |
| 200 x 57 | GRAPH | [`verify-repair4-200x57-graph.txt`](verify-repair4-200x57-graph.txt) | 57 | 200 |
| 200 x 57 | TABLE | [`verify-repair4-200x57-table.txt`](verify-repair4-200x57-table.txt) | 57 | 200 |

No line goes past the edge. Footer: `not in herdr: no coordinator or supervisor started`. The only `hive supervise` process before and after was PID 2107907, which I did not start. No coordinator tab was made.

### V3, merge, and gate

`git diff origin/dev --stat`: `internal/tui/` (`agents.go`, `agents_test.go`, `fix_test.go`, `header.go`, `refresh.go`, `shell.go`, `shell_test.go`, `tui.go`, six goldens), `testdata/ste-baseline.json` (+1), and the brief copy and evidence of this task. Pass. `git merge-tree --write-tree origin/dev HEAD` exit 0 (`origin/dev` at `21d4b8c0`).

Gate from the worktree root: `gofmt -l` printed nothing. `go mod tidy -diff` printed nothing, exit 0. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race -count=1 ./...` exit 0.

### Resources

| Resource | ID | Result |
| --- | --- | --- |
| Tab in my workspace `wSC` | `wSC:t2` (`t92-v5-live`), root pane `wSC:p2` | Closed after a label and ID check |
| TUI 207 x 57 in `wSC:p2` | PID 2688473 | Stopped with Ctrl+C after a PID check |
| tmux server `-L t92v5` | PID 2696132, TUI PIDs 2696136 and 2696142 | Stopped with `tmux -L t92v5 kill-server`; all three PIDs gone |
| Scratch worktree | `.hive/tmp/t92v5/wt` at the task head | Removed with `git worktree remove` |

I did not use workspace `wBC`. Tab `wSC:t1` stays open.

## Repair 5

Verifier: `t92-verifier-6`. Task head: `9c51e950`. Repair 4 tip: `07a5585e`. Board: `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`.

Result: **fail**. The structure and the pod title repair pass. The Repair 5 colors fail on one row of each GRAPH frame: on the row that holds the selected node's label, each pod border, each live node border, and each edge after the label shows the default foreground, not `colorInfo` or `colorWorking`.

| Item | Result |
| --- | --- |
| R5.1 title before the stem, `…` truncation, glyph kept with the name | pass |
| R5.1 test red on `07a5585e`, green on the head | pass |
| R5.2 colors from the `shell.go` palette | **fail** (selected row) |
| R5.2 SGR test red on `07a5585e`, green on the head | pass |
| R5.2 no background | pass |
| R5.2 selection style untouched | pass |
| Structure (explorer tree, pods with titles, node boxes, edges with the unbroken stem, detail strip) against the board | pass |
| Earlier passes (nudge mutation, ghost rule, task 90 idle CPU tests) | pass |
| Exact frames at 120 x 40 and 200 x 57 | pass |
| V3 diff scope | pass |
| V4 gate | pass |
| `git merge-tree --write-tree origin/dev HEAD` | exit 0 |

### R5 red and green

Scratch worktree at `07a5585e` with `internal/tui/agents_test.go` of the head:

```sh
go test ./internal/tui -run '^(TestAgentsGraphLongPodTitleKeepsNameTogetherBeforeStem|TestAgentsGraphColorsMatchBoard|TestAgentsGraphPodTitleSharesTopBorder)$' -count=1
```

Red on `07a5585e`:

- `long pod title = "╭─ ≡ stenographer · a── ┬ goal with a long title╮", want prefix "╭─ ≡ stenographer · a… ─┬"`
- `alpha pod title = "╭─ ≡ alpha · Alpha── ┬ goal for long graph╮", want prefix "╭─ ≡ alpha · Alpha… ─┬"`
- `TestAgentsGraphColorsMatchBoard`: each of the six checks (pod border, delegation edge, live node border, live dot, `cc`, `pi`) got foreground `""`.

Green: `go test ./internal/tui -count=1` passes on the head. The SGR values in the test agree with `shell.go`: `colorInfo #6fb7ff` = `38;2;111;183;255`, `colorWorking #6ee69a` = `38;2;110;230;154`, `colorAccent #f5b942` = `38;2;245;185;66`.

### R5.1 pod titles: pass

Live at 207 x 57 and in both exact frames: `╭─ ≡ home ──┬──────────╮`, `╭─ ≡ hive · Hi… ─┬────────────────╮`, `╭─ ≡ qmeter ──┬────────────╮`, `╭─ ≡ stenographer ──┬──────────────────╮`. Each title is at the left of the top border, `≡` stays with the name, the junction comes after the title, and a long title ends in `…`. Goldens: `╭─ ≡ alpha · Alpha… ─┬…╮` and `╭─ ≡ beta · Beta g… ─┬…╮`; no title touches a corner.

### R5.2 colors: fail on the selected row

I rendered the GRAPH view of the fixture model (`agentsBoardModel`, 176 x 48) on `07a5585e` and on the head with a scratch test that was not committed, and I read the live ANSI of `herdr pane read wSJ:p2 --source visible --format ansi`. Foreground per cell (B = `colorInfo`, G = `colorWorking`, H = honey, `_` = default), live at 207 x 57:

```text
row 9   │ │ ▶ ○ coordinator  │ │   │  │ ○ manager-hive            │  │   │  │ ○ manager-qmeter
        B _ H H HHHHHHHHHHH  _ _   _  _ _ ____________            _  _   _  _ _ ______________
row 10  │ │ cc you · now     │ │   │  │ cc idle · 1m              │  │   │  │ cc idle · 4h
        B _ HH ___ _ ___     _ B   B  G HH ____ _ __              G  B   B  G HH ____ _ __
```

On row 10 the pod borders are B and the live node borders are G, as the brief asks. On row 9, the row of the selected label, each border after the label is default. The raw bytes for each such glyph are `\e[38;2;111;183;255m\e[1;38;2;245;185;66m\e[m│`: the color opens, the selected style opens again, a reset follows, and then the glyph prints. Cause: `colorGraphLine` (`agents.go`) cuts a line that already holds the selected label's SGR; `ansi.Cut` carries that style into each later segment, and its reset lands before the glyph. The same row is default in the 120 x 40 and 200 x 57 exact frames (tmux ANSI) and is visible in the screen PNG. `TestAgentsGraphColorsMatchBoard` checks no cell on the selected row, so it does not see this.

The selected node itself is untouched: its label is `\e[1;38;2;245;185;66m▶ ○ coordinator` on both `07a5585e` and the head, and its box border stays default. Task 102 owns that style. No frame has a `48;` background code (`grep -c` gives 0 for the fixture render and each pane capture).

### Design visual check

Board PNG: `firefox --headless --no-remote --profile .hive/tmp/t92v6/ff --window-size=1220,780 --screenshot .hive/tmp/t92v6/board.png file://…/Main.dc.html`. It stays at `/home/penguin/source/hive/projects/hive/worktrees/92-agents-board-layout/.hive/tmp/t92v6/board.png`, not in Git. Screen PNGs: the three `herdr pane read wSJ:p2 --format ansi` captures, changed to an HTML `<pre>` with a short Python script, then the same Firefox command at 1700 x 960. They stay at `.hive/tmp/t92v6/screen-{tree,graph,table}.png`. The browser font draws box glyphs wider than one cell, so I judged cell positions from the text captures and colors from the ANSI.

Worker element table (V2, "Repair 5 visual check"):

| Board element | Worker | My judgment |
| --- | --- | --- |
| Header, Explorer, Screen list, Tabs, Legend | matches | agree for structure; colors in "Gaps for 102" |
| Pods | overridden (title before stem, `…`) | agree |
| Nodes | overridden (selection rule) | agree |
| Edges | differs (no nudge record; 120 clip) | agree |
| Detail strip | matches | agree for fields; colors in "Gaps for 102" |
| Key bar | overridden (G12) | agree |
| Pod border color | matches | **disagree**: default on the selected row |
| Delegation and graph edge colors | matches | **disagree**: default on the selected row |
| Live node border color | matches | **disagree**: default on the selected row |
| Live working dot color | matches | agree |
| `cc` / `pi` harness colors | matches | agree |
| Selection and fill | overridden | agree |

### Gaps for 102

These differ from the board and do not fail this task.

| Element | Board | Output |
| --- | --- | --- |
| Header working count | `6 working` in green | default text |
| Header needs-you count | `0` in honey | default text |
| Tabs | `[ GRAPH ]` honey, `TREE` and `TABLE` dim | the whole tab line honey bold |
| Legend | right-aligned, dim | left, after the tabs, honey bold |
| Explorer status dots | live `●` green | default text |
| Explorer working count | `· 1 working` dim | default text |
| Explorer selected row | fill block behind the row | honey bold with `▶`, no fill (user rule) |
| Pod title text | full goal title, e.g. `≡ hive · H2 spec layout` | `≡ hive · Hi…`; the pod is narrow, so the title is cut before the stem |
| Selected node border | honey | default |
| Detail strip border | honey | muted line color |
| Detail strip name | honey bold | default text |
| Detail strip labels and values | labels dim, values bright | all default text |
| Key bar keys | key chips with a light fill | plain text (no-background rule) |

### Earlier passes

In the scratch worktree at the head, each file restored with `git checkout` after each run:

- Nudge filter removed (`agents.go:144`): `TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` failed. Holds.
- Ghost rule `open := true` (`agents.go:121`): `TestAgentsGhostsRequireAnOpenTask` failed: `closed task ghost "ghost-cancelled" is visible without history`, and for `ghost-landed`, in the graph and the explorer. Holds.
- Task 90 render cache, `tui.go:420` and `shell.go:156` to `if true`: `header refresh rendered screen 2 times` and `minute refresh rendered screen 2 times`. Holds.
- Task 90 idle reload, `headerChanged := true` (`tui.go:242`): `unchanged tick started 1 header loads and 0 screen loads; want none` and `changed tick started 1 header loads and 1 screen loads; want 0 and 1`. Holds.
- Unchanged code: the tests pass.

### Live check

Binary built from `9c51e950` at `.hive/tmp/t92v6/hive`. Scratch home `.hive/tmp/t92v6/home` (copy of `.hive/README.md`, `.hive/state`, `.hive/spec`) in `HIVE_HOME`, temp `XDG_CONFIG_HOME` `.hive/tmp/t92v6/config`. Each start removed `HERDR_ENV`, `HERDR_PANE_ID`, `HERDR_WORKSPACE_ID`, `HERDR_SOCKET_PATH`, and `HERDR_TAB_ID`; `/proc/2853346/environ` had no `HERDR_ENV`, `HERDR_PANE_ID`, `HERDR_WORKSPACE_ID`, or `HERDR_SOCKET_PATH`. I never opened the Chat input and never pressed Enter in the TUI.

1. Tab `wSJ:t2` (`t92-v6-live`), pane `wSJ:p2`, 207 x 57, shell PID 2852512: TUI PID 2853346. ANSI captures of TREE, GRAPH, and TABLE for the PNGs.
2. Exact frames: from `wSJ:p2`, `tmux -L t92v6 -f /dev/null new-session -d -s v120 -x 120 -y 40 'env -u … ./hive'` and `-s v200 -x 200 -y 57`. Server PID 2856178, TUI PIDs 2856179 and 2856182.

| Size | View | Capture | Lines | Widest line |
| --- | --- | --- | --- | --- |
| 120 x 40 | TREE | [`verify-repair5-120x40-tree.txt`](verify-repair5-120x40-tree.txt) | 40 | 120 |
| 120 x 40 | GRAPH | [`verify-repair5-120x40-graph.txt`](verify-repair5-120x40-graph.txt) | 40 | 120 |
| 120 x 40 | TABLE | [`verify-repair5-120x40-table.txt`](verify-repair5-120x40-table.txt) | 40 | 120 |
| 200 x 57 | TREE | [`verify-repair5-200x57-tree.txt`](verify-repair5-200x57-tree.txt) | 57 | 200 |
| 200 x 57 | GRAPH | [`verify-repair5-200x57-graph.txt`](verify-repair5-200x57-graph.txt) | 57 | 200 |
| 200 x 57 | TABLE | [`verify-repair5-200x57-table.txt`](verify-repair5-200x57-table.txt) | 57 | 200 |

No line goes past the edge. Footer: `not in herdr: no coordinator or supervisor started`. The only `hive supervise` process before and after was PID 2107907, which I did not start. No coordinator tab was made.

### V3, merge, and gate

`git diff origin/dev --name-only`: no file outside `internal/tui/`, `testdata/ste-baseline.json`, and this task's brief and evidence. Pass. `git merge-tree --write-tree origin/dev HEAD` exit 0 (`origin/dev` at `21d4b8c0`).

Gate from the worktree root: `gofmt -l` printed nothing. `go mod tidy -diff` printed nothing, exit 0. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race -count=1 ./...` exit 0.

### Resources

| Resource | ID | Result |
| --- | --- | --- |
| Tab in my workspace `wSJ` | `wSJ:t2` (`t92-v6-live`), root pane `wSJ:p2`, shell PID 2852512 | Closed after a PID and ID check |
| TUI 207 x 57 in `wSJ:p2` | PID 2853346 | Stopped with Ctrl+C after a cmdline check |
| tmux server `-L t92v6` | PID 2856178, TUI PIDs 2856179 and 2856182 | Stopped with `tmux -L t92v6 kill-server`; all three PIDs gone |
| Scratch worktree | `.hive/tmp/t92v6/wt` | Removed with `git worktree remove` |

I did not use workspace `wBC`. Tab `wSJ:t1` stays open.
