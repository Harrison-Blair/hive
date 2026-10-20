# Agents board layout

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, repair after wave C |
| Tier | strong |
| Depends on | 91-agents-live-generation. Base: `dev` after T3 lands |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The Agents screen does not match the board that the user accepted:
`/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`. The live screen shows the sidebar of 8 screens and a
flat list. Make the screen match the board. The feature README is
`.hive/spec/hive-tui/README.md`. Read decisions 1, 7, 10, and 14, and G4,
G5, and G12. The decisions win over the board. When the code does not agree
with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## The board

Read the board file. It shows:

- A header: `⬢ hive`, the count of projects and agents, the working and
  idle counts, the supervisor state, the needs-you count, and the time.
- On the left, an explorer tree: `home` with the coordinator and the
  supervisor, then each project with its working count, its manager, and
  the workers under the manager. Below the tree, the screen list in a
  compact form: `1 CHAT · 3 GOALS · 4 QUEUE` and so on.
- In the main panel, the tabs `TREE`, `[ GRAPH ]`, `TABLE` and the edge
  legend: `─ delegates · ┄ nudges · ⋯ replaced`.
- GRAPH: one pod box for each group (`home` and each project with its goal
  title). In each pod, node boxes. Each node box has the status dot, the
  name, the harness or the pid, and the task or the activity. Edges join the
  coordinator to each manager and each manager to its workers. Dashed edges
  show nudges. A ghost node shows a stopped worker of an active task.
- A detail strip for the selected node: parent, model, task with status,
  pane, and next task.
- A key bar: arrows move on the graph, `tab` changes the view, `enter`
  peeks at the pane, `a` attaches, `h` toggles the history (T3), `t`
  opens the task.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/` and its tests | R1, R2, R3. |
| `testdata/ste-baseline.json` | Only to lower a count, and to add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy, when `hive task create` did not write it. |

Do not edit other files.

## Requirements

### R1 Layout

- The Agents screen has the explorer tree on the left and the main panel on
  the right, as on the board. It does not show the full sidebar of 8
  screens. It shows the compact screen list below the tree. Each other
  screen keeps its layout.
- GRAPH draws pod boxes, node boxes, and edges with box-drawing characters,
  as on the board. Use the T3 default set and the T3 history toggle.
- The detail strip shows the fields of the board for the selected node. Show
  `?` for an unknown value.
- TREE and TABLE keep their data and use the main panel.
- The layout fills the terminal at each size of the T1 size test. When the
  graph is wider than the panel, the view scrolls with the selection. No
  line goes past the edge.

### R2 Keys

The keys of the board work. `t` opens the Tasks screen at the task of the
selected node. Each key that the board shows and G12 forbids shows the hint
`ask in Chat`.

### R3 Tests

- Write each test first and see it fail for the correct reason.
- Add golden frame tests for GRAPH, TREE, and TABLE at 120 x 40 and at
  200 x 57 with fixture agents (home, two projects, one ghost). Write each
  golden file by hand from the board, or check each line of the first output
  against the board before you save it. Record how in the evidence.
- Add tests for the selection move, the detail strip, and `t`.

## Not in scope

Other screens. Agent data rules (T3). `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Run each R3 test against the base code, then against the task code | Each new test fails on the base and passes on the task branch. |
| V2 | Build the task binary. Run it on this home in a herdr pane at the size of the user pane (about 200 x 57) and at 120 x 40. Capture GRAPH, TREE, and TABLE with `herdr pane read` | Each capture matches the parts of the board in "The board". Put the captures in the evidence with a list of each difference and its cause. |
| V3 | `git diff origin/dev --stat` | Only the files of the source table changed. |
| V4 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other commands exit 0. |

Close each herdr tab and workspace that you open. Do not close a pane that
you did not open.

## Finish

Commit on the task branch with a subject such as `feat(tui): match the
Agents board`. Push the task branch. Write the evidence. Run
`hive task complete`. Send the commit SHA and the V1-V4 results to the
manager with `hive inbox send --to manager-hive`.

## Task names

This brief uses these names: T1 is 88-tui-terminal-io, T2 is 90-tui-idle-cpu,
T3 is 91-agents-live-generation, T4 is 92-agents-board-layout, and T5 is
89-screen-board-audit.

## Scope change: chat safety

A TUI check uses a scratch home (a temp copy of a marked home, set with
`HIVE_HOME`, and a temp `XDG_CONFIG_HOME`), or it does not open the Chat
input. On the real home, never press `i` on the Chat screen and never press
Enter there. A chat message from an agent has the authority of the user.

## Scope change: style rules

The user gave these style rules. Paint no background. A selected item,
also the selected graph node, has honey bold text and a `▶` marker, with no
fill block. Normal text uses the default foreground of the terminal. Task
102-selection-marker applies the rules on each screen after this task.

## Scope change: no startup in a scratch home

Bare `hive` starts a coordinator tab and a supervisor process when it runs in
herdr. In a live check on a scratch home, start the binary with the herdr
variables removed, so that it runs only the TUI:
`env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH ./hive`.
After each check, make sure that you started no `hive supervise` process and
no coordinator tab. Stop each process and close each tab that you started.

## Scope change: ownership

Never send a signal to a process and never close a tab, pane, or workspace
that you did not start. Record the pid and the pane ID of each thing that you
start. Before you stop or close it, compare its pid or ID with the record.
Never work in workspace `wBC`: it holds the user's TUI, the coordinator, and
the supervisor.

## Scope change: ghost rule and supervisor row

The coordinator checked 91-agents-live-generation on the real home and found
two gaps. Fix them in this task.

1. Ghosts show for tasks that are not active, for example the cancelled tasks
   20, 21, 25, 26, and 67, and the verified tasks `10-pr47-final` and
   `45-ci-fix` that never landed. Use this rule. A stopped worker or verifier
   is a ghost only when its task is open:
   - the task status is `created`, `assigned`, or `completed`, or
   - the task status is `verified`, it has no delivery with `pushed` true
     (G8), and a goal with status `active` names it in a wave.
   A task that is `cancelled`, landed, or verified with no active goal does
   not give a ghost. With history on (`h`), each stopped agent shows, as 91
   made it.
2. The supervisor row shows `? · unknown · ?`, but `hive agent list` shows
   the supervisor generation 16 as live in pane `wBC:p19`. Find the cause and
   show the live supervisor with its status, pid, and age, as on the board.

Write each test first with fixture records and literal names. See it fail.
Record the cause of gap 2 in the evidence.

## Repair 1

The verifier failed V2 and the R3 golden record. See `VERIFIED.md`.

1. Draw the edges in the graph as on the board. Draw a delegation bus above
   the pods that joins the coordinator to each manager box. Draw the nudge
   edge as a dashed line with the label `nudges` between the two boxes.
   Remove the text list of edges below the pods. Join the sibling workers of
   a manager with a stem from the manager box.
2. For each other difference from the board, make the output agree with the
   board, or record the decision or the T3 rule that causes the difference.
   The verifier found these differences: the pod title `┌─ home ─` against
   `≡ home`, the harness `claude` against `cc`, the detail strip with no
   border, the header `supervisor fresh` and `needs_user 0` against
   `supervisor ●` and `needs you 0`, and the key bar `h history: off` against
   `h hide stopped`. Compare each region of the board again and find each
   other difference.
3. Write the golden files again. Check each line against the board before you
   save it. In the evidence, record the method and each difference that
   stays, with its cause.
4. Merge `origin/dev` into the task branch first. Task 90 changes
   `internal/tui/shell.go`. Keep both tasks' behavior. In
   `testdata/ste-baseline.json`, keep each row of both sides.
5. This edit copies the manager brief over your brief copy again. Apply your
   STE rewording to the copy again, and keep its meaning.
6. Do V2 again on a scratch home in a tab of your own workspace, with the
   herdr variables removed. Stop and close only the pid and pane that you
   started. Never work in workspace `wBC`.
7. Run the gate. Commit and push. Make sure that `git merge-tree
   --write-tree origin/dev HEAD` exits 0. Run `hive task complete`. Send the
   commit SHA to `manager-hive` with `hive inbox send --to manager-hive`.

## Repair 2

The verifier failed two items of Repair 1. See `VERIFIED.md`, section
Repair 1.

1. The test `TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` looks for the
   removed text form of the edge. It passes if code removes the role filter
   and the graph draws a `nudges` route to a worker. Change it so that it
   reads the drawn graph. Remove the filter, see the test fail, then put the
   filter back.
2. Fix these drawing faults to agree with the board:
   - `manager-qmeter` has a stem `┴` with only hidden children. Draw no stem
     when a manager shows no child.
   - The ends of the bus are `┬`. Use the end glyphs of the board.
   - In TREE, the last child is `├─`. Use `└─`.
   - The node second line has an extra `·`.
   Write a test for each fault first, with literal expected text.
3. For each other difference, make the output agree with the board, or
   record the decision or T3 rule that causes it, in V2. The verifier found:
   the key bar words and order (move on graph, peek pane, open task), the
   `DETAIL` label, the pod title inside the pod and not on the border, the
   explorer connectors, and the cause of the selection style. The selection
   style comes from the user rule in the brief (honey bold with `▶`).
4. Write the golden files again where the output changes. Check each line
   against the board and record it.
5. This edit copies the manager brief over your brief copy again. Apply your
   STE rewording to the copy again, and keep its meaning.
6. Merge `origin/dev` when it moved. Run V2 again on a scratch home in a tab
   of your own workspace, with the herdr variables removed. Stop and close
   only the pid and pane that you started. Never work in workspace `wBC`.
7. Run the gate. Commit and push. Make sure that `git merge-tree
   --write-tree origin/dev HEAD` exits 0. Run `hive task complete`. Send the
   commit SHA to the manager with `hive inbox send --to manager-hive`.

## Repair 3

The verifier passed each item of Repair 2 except the list of differences.
See `VERIFIED.md`, section Repair 2. Make these five points agree with the
board:

1. The delegation stem is one line from the bus to the manager node. Where
   it crosses the top border of a pod, draw a junction glyph, not `─`.
2. Draw pod corners as `╭`, `╮`, `╰`, and `╯`.
3. Each node box has one space of padding on the left and the right of its
   text.
4. The header separates its groups with spaces, not `·`, and puts the time
   at the right edge.
5. The detail strip separates its fields with spaces, not `·`.

Write a test for each point first, with literal expected text, and see it
fail. Write the golden files again where the output changes.

Then compare each region of the board with the output once more: header,
explorer, screen list, tabs, legend, pods, nodes, edges, detail strip, and
key bar. In V2, write one table with a row for each region. Each row gives
"same", or the difference and its cause. A difference with no cause is a
fault: fix it.

This edit copies the manager brief over your brief copy again. Apply your STE
rewording to the copy again, and keep its meaning. Merge `origin/dev` when it
moved. Run V2 on a scratch home in a tab of your own workspace, with the
herdr variables removed. Stop and close only the pid and pane that you
started. Never work in workspace `wBC`. Run the gate. Commit and push. Make
sure that `git merge-tree --write-tree origin/dev HEAD` exits 0. Run
`hive task complete`. Send the full commit SHA to the manager with
`hive inbox send --to manager-hive`.

## Repair 4

The verifier passed each item of Repair 3 except two. See `VERIFIED.md`,
section Repair 3.

1. The pod title is on the top border of the pod, as on the board, with the
   glyph of the board: for example `╭─ ≡ home ───╮`. Repair 1 named only a
   glyph difference (`┌─ home ─` against `≡ home`). It did not ask you to move
   the title into the pod. Change the Repair 2 test back so that it expects
   the title on the border.
2. The delegation stem is one unbroken vertical line from the bus to the top
   of the manager node. It crosses the pod top border with a junction glyph,
   and it continues on each row inside the pod above the manager node, for
   example the `┄ inferred` row. Change
   `TestAgentsGraphDelegationStemJoinsPodBorder` so that it checks each cell
   of the stem column from the bus to the node. See it fail on the Repair 3
   tip `72d51057`.
3. Write the golden files again where the output changes. Check each line
   against the board. Update the Pods row of the V2 region table.

This edit copies the manager brief over your brief copy again. Apply your STE
rewording to the copy again, and keep its meaning. Merge `origin/dev` when it
moved. Run V2 on a scratch home in a tab of your own workspace, with the
herdr variables removed. Stop and close only the pid and pane that you
started. Never work in workspace `wBC`. Run the gate. Commit and push. Make
sure that `git merge-tree --write-tree origin/dev HEAD` exits 0. Run
`hive task complete`. Send the full commit SHA to the manager with
`hive inbox send --to manager-hive`.

## Design

The approved boards for this task:

- `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html`

Visual check rule (the user approved it on 2026-10-20):

- Worker: capture each changed screen and each board above as PNG. Compare
  them. Fix each gap before you finish.
- Board to PNG: `firefox --headless --no-remote --profile SCRATCH
  --window-size=1220,780 --screenshot OUT.png file://BOARD`, with a new
  scratch profile folder below `.hive/tmp/`.
- Screen to PNG: start the TUI on a scratch home in a pane that you started.
  Run `herdr pane read PANE --source visible --format ansi > F.ansi`. Change
  the ANSI SGR codes to an HTML `<pre>` with a short script in `.hive/tmp/`.
  Then use the firefox command on that HTML file.
- Evidence: commit the screen PNGs in the task evidence. Do not commit a
  board PNG. Keep it below `.hive/tmp/` and give its path.
- Report: in the evidence, one table with one row for each board element:
  `matches`, `differs` (with the cause), or `overridden` (with the
  decision that overrides it).
- Capture only your own panes or a scratch run.

## Repair 5

The verifier passed the stem of Repair 4 and each earlier item. Two items
fail. See `VERIFIED.md`, section Repair 4.

1. Pod titles. The stem cuts the title: line 6 of `graph-200x57.golden` shows
   `goal for long grap╮`, and the live stenographer pod shows
   `╭─ ≡─────────── ┬ stenographer ╮`. Keep the glyph and the name together.
   Put the title at the left of the top border. Put the stem junction after
   the title. When the title and the stem do not fit, shorten the title with
   `…`, never the glyph. Write a test with a long pod name and a stem, with
   literal expected text, and see it fail on `07a5585e`.
2. Colors. Use the colors of the board with the palette in
   `internal/tui/shell.go`: pod borders `colorInfo`, edges and the borders of
   live nodes `colorWorking`, live dots `colorWorking`, harness `cc`
   `colorAccent`, harness `pi` `colorInfo`. Do not change the style of the
   selected node: task 102 owns the selection style (honey bold with `▶`).
   Use no background. Write a test that reads the SGR codes of each item
   first, and see it fail.
3. In the Design element table, add one row for each color item and correct
   the Pods row. Each row is `matches`, `differs` (with the cause), or
   `overridden` (with the decision). Capture the PNGs again.

This edit copies the manager brief over your brief copy again. Apply your STE
rewording to the copy again, and keep its meaning. Merge `origin/dev` when it
moved. Run V2 on a scratch home in a tab of your own workspace, with the
herdr variables removed. Stop and close only the pid and pane that you
started. Never work in workspace `wBC`. Run the gate. Commit and push. Make
sure that `git merge-tree --write-tree origin/dev HEAD` exits 0. Run
`hive task complete`. Send the full commit SHA to the manager with
`hive inbox send --to manager-hive`.
