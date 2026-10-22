# Agents: live overview, pod view, and lit message lines

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g5-live-agent-activity`, wave A |
| Tier | strong |
| Depends on | none. Base: `dev` at 26ff49cd or later |
| Grant | goal `g5-live-agent-activity`: the user approved the design on 2026-10-22. Build and land on `dev`. |
| Evidence | `.hive/spec/tui-v2/tasks/124-live-agent-activity/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved four new Agents boards on 2026-10-22 (canvas version 68,
id 1792704888-ced6). This task copies the boards into the spec, exports their
golden grids, adds the decisions to the tui-v2 spec, and builds the Agents
screen to match the boards:

1. An overview where you select a full pod, and a one-pod view where you
   select single agents.
2. A coordinator card above the pods, with a line to each pod.
3. A message between two agents lights the line between them for 5 seconds.
4. `hive agent send` logs its sender.

The TUI must match each board cell for cell: the glyph, the colour token, and
the bold flag. When a board and this brief do not agree, stop and report it
with `hive inbox send --to manager-hive TEXT --agent NAME`. Do not choose for
the user.

## Design

The approved boards, copied from the canvas "Hive TUI Redesign"
(https://claude.ai/artifact/VrhcA98g9G2G9cFQao5Z9h, version
1792704888-ced6), in `/home/penguin/source/hive/.hive/tmp/g5-live/boards/`:

- `LiveOverview.dc.html` (120×40): coordinator card, `hive` pod selected,
  the line from the coordinator to `manager-hive` lit.
- `LivePod.dc.html` (120×40): one-pod view of `hive` (enter on `hive`),
  tree, the line from `manager-hive` to `t6-worker-1` lit.
- `CompactLiveOverview.dc.html` (80×24): the overview.
- `CompactLivePod.dc.html` (80×24): the one-pod view.

The boards include the g4 decisions 58 to 61. The staged folder has no
`support.js`. Use the copy in `.hive/spec/tui-v2/design/`.

Follow the visual check rule in the worker skill. The user is colourblind.
Each state has a glyph and a word. Colour only adds to a state: a lit line
also has heavy glyphs and an arrow.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/<Board>.dc.html`, `<Board>.grid`, `<Board>.fixture.json` for the 4 boards | AC1. New copies, grids, and fixtures. |
| `internal/boardexport/check.go` and its tests | AC1. Add the 4 boards to `Sizes`. Add the new symbols to the list. |
| `internal/tui/component/symbols_test.go` | AC1. Width test for each new symbol. |
| `.hive/spec/tui-v2/README.md` | AC2. Decisions 62 to 72. Mark decision 14 as replaced. |
| `internal/tui/agents_screen.go`, `internal/tui/agents.go`, `internal/tui/frame.go`, `internal/tui/demo.go`, `internal/tui/component/*.go`, and their tests | AC3 to AC9. |
| `internal/tui/tui.go`, `internal/tui/refresh.go`, `internal/tui/tick.go` | Only when AC3 to AC9 need a change there. Record the cause in the evidence. |
| `internal/agent/send.go` and its tests | AC9. The sender field. |
| `.hive/spec/tui-v2/design/<Board>.fixture.json`, `.dc.html`, `.grid` for the 10 Agents boards that exist | AC11 only. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |

Do not edit other files. Do not add a mask. Do not change the exporter
except `check.go`. The 10 Agents boards that exist change only as AC11 says.
The other 21 boards that exist do not change. Each of their grids and golden
checks must stay the same.

## Requirements

### AC1 Boards, grids, symbols

- Copy the 4 boards into `.hive/spec/tui-v2/design/`. `cmp` against the
  staged copies gives no difference.
- Add each board to `Sizes` in `internal/boardexport/check.go`: 120×40 for
  `LiveOverview` and `LivePod`, 80×24 for the two `Compact` boards.
- Add the new symbols to decision 13 and to the symbol list: `┰ ┸ ┺ ┹ ╃ ┻
  ◀ ▼`, and each mixed joint that AC6 needs. Each takes one cell.
- Run the exporter. Commit the 4 new grids. Each other grid stays the same.
- Write one demo fixture for each board, in the format of the fixtures that
  exist. The fixture holds the agents, the selection, the view, and the lit
  lines with a fixed clock.

### AC2 Spec

Add decisions 62 to 72 to the tui-v2 README in ASD-STE100 Strict. Keep the
meaning of this brief. Mark decision 14 as replaced by decision 65.

### AC3 Decision 62: two levels of selection

- The overview shows all pods with the pod layout of today: tree or rail by
  the rules that exist, and the pod width rule that exists.
- In the overview, you select a full pod. The home pod is a pod as a project
  is. A selected pod has a heavy `accent` border (`┏ ━ ┓ ┃ ┗ ┛`). Its title
  is `accent` bold. The title of a pod that is not selected stays `dim` bold.
- In the overview, a card is not selectable. A card has no heavy border.
- `←` and `→` select the pod to the left and to the right. At the edge of a
  page, `←` and `→` scroll the pages as today.
- `enter` on a pod opens the one-pod view of that pod.
- In the one-pod view, you select single agents with the keys of today.
  `esc` goes back to the overview, with the same pod selected.

### AC4 Decision 63: one-pod view

- The one-pod view shows one pod, with the pod width rule of today for V=1:
  at most 64 cells from 120 columns, and the full body width below 120. The
  pod is at the center of the body.
- The pod uses the tree or the rail by the rule of today. The user has a
  decision open about a pod with too many children for the tree. Do not add a
  rule for it.
- The box title is `project agents · POD`, for example
  `project agents · hive`.

### AC5 Decision 64: one project

With one project and no home pod, the Agents screen opens the one-pod view
at once. `esc` does nothing in this case. The footer does not show
`esc pods`.

### AC6 Decision 65: coordinator card

- The coordinator card is at the top center of the body, in the overview and
  in the one-pod view. It is a card as the others are, with the colour rules
  of decision 60.
- The coordinator joint is always at the card x+12.
- In the overview, one bus row is below the coordinator card. A line goes
  from the coordinator joint to the bus row, and from the bus row to the top
  border of each pod on the page. Each pod drop is at the center of its pod,
  with `┴` on the pod top border. A drop that is 1 column or less from the
  coordinator joint moves to the column of the joint, and the two merge
  (`╃`, `┼`, and so on). The bus goes only to the pods on the visible page.
- In the one-pod view, the line goes from the coordinator joint down through
  the pod top border (`┼`) to the top joint of the manager card.
- The line is in the token `border` when it is not lit.
- In the overview, `↑` selects the coordinator card. `↓` selects the last
  selected pod again. `esc` does nothing there. `enter` on the coordinator
  card opens its details popup.
- In the one-pod view, the coordinator card is not selectable.
- The box title does not change: `project agents · N projects` in the
  overview.
- The supervisor has no card. A nudge stays as `nudged Nm`.

### AC7 Decision 66 to 68: lit lines

- A message lights the line from its sender to its receiver. The sources are:
  - Each inbox record: the sender is the field `agent`, the receiver is the
    field `role`, the time is the field `created_at`.
  - Each `agent prompt` log event of `hive agent send`: the sender is the new
    field of AC9, the receiver is `agent_id`.
- A message with a sender or a receiver that has no card (the user, the
  supervisor, or an unknown name) lights nothing.
- A lit line uses heavy glyphs (`━ ┃` and heavy or mixed joints, for example
  `┏ ┓ ┰ ┸ ┺ ┹ ╃ ┻`) in the token `info`. The `info` token wins over the
  card colour of decision 60 on a joint of the card border. A joint that also
  has arms that are not lit uses the mixed glyph with heavy arms only on the
  lit side, as the boards show.
- The line has one arrow toward the receiver (`◀ ▶ ▲ ▼`) in `info`. The
  arrow is on the straight cell nearest to the receiver. On a vertical run it
  is on the cell next to the receiver joint. On a horizontal run one `━`
  stays between the corner and the arrow. The boards show both cases.
- When two agents have no direct line, the full path between them lights.
  In the overview, a line from or to a card in a pod lights up to the pod
  border, as `LiveOverview` shows.
- No board shows a lit rail. On a rail, use the mixed box glyph that has the
  lit arms heavy (for example `┠ ┡ ┢ ┣ ┞ ┟`). Add each glyph that you use to
  decision 13 and to the symbol list. Test each case at the model.
- A line stays lit for 5 seconds after the time of the message. A new message
  on the same line starts the 5 seconds again.
- Lines light in the overview and in the one-pod view.

### AC8 Decision 69 to 71: legend, footers, Help

- The legend at 120 columns adds `━ ▶ message, lit 5s`. `━ ▶` is in `info`.
  The rest is `dim`, as the boards show.
- Footers:
  - Overview, 120: `←→ select pod · ↑ coordinator · enter open pod · h stopped · ? help`
  - Overview, 80: `←→ select pod · ↑ coordinator · enter open · ?`
  - One-pod view, 120: `←↑↓→ move · enter details · a attach · p peek · t task · esc pods · h stopped · ? help`
  - One-pod view, 80: `←↑↓→ move · enter details · esc pods · ?`
- Update the Help overlay for the new keys. Keep the Help board grid the
  same. When the Help board and the new keys do not agree, stop and report.

### AC9 Decision 72: sender of `hive agent send`

- The `agent prompt` log event of `hive agent send` gets the field `from`
  with the name of the acting agent.
- The acting agent is the agent record with a `pane_id` equal to the
  `HERDR_PANE_ID` of the caller. When no record agrees, the event has no
  `from` field, and the message lights nothing.

### AC11 The 10 Agents boards that exist

The coordinator gave this decision on 2026-10-22. The 4 Live boards and
decisions 62 to 72 are the design authority. The new code makes the new
grids of the 10 Agents boards that exist: Main, AgentsMany, CompactAgents,
AgentDetails, AgentsHome, AgentsNested, AgentsHiddenParent,
AgentsCoordinatorChild, CompactAgentsHome, AgentDetailsGeneral. The user
approves them.

1. Update the fixtures of the 10 boards. Render each of the 10 cases from the
   new code at its board size. Each grid follows decisions 62 to 72 and
   matches the 4 Live boards for each element that they show: the
   coordinator card, the bus, the pod selection, the legend, the footers,
   and the lit glyphs.
   - Main and CompactAgents become one-pod views (AC5).
   - AgentsMany, AgentsHome, AgentsHiddenParent, and CompactAgentsHome show
     the overview with the pod of the old selected card selected.
   - AgentDetails, AgentDetailsGeneral, AgentsNested, and
     AgentsCoordinatorChild show the one-pod view of the pod of the old
     selected card, with the same card selected (and the popup open on the
     two details boards).
2. Write the 10 new grids to
   `.hive/spec/tui-v2/tasks/124-live-agent-activity/evidence/new-grids/` in
   the grid format of the exporter. Commit and push them. Send the path to
   the manager. Do not change the golden files of the 10 boards yet.
3. Hold until the manager sends the approval. Then the manager stages the
   10 approved boards from the canvas. Copy them over the `.dc.html` files.
   Run the exporter. Each exported grid must equal its grid from step 2.
   Commit the grids. Then the golden checks of the 10 boards use them.
4. In the spec, record that the new grids replace the old grids of the 10
   boards.

### AC10 Tests

- Write each test first and see it fail for the correct reason.
- The golden check passes for each of the 4 new boards, for the 10 boards
  of AC11, and for each other board that passes on `dev`.
- Model tests with a fixed clock: a line is lit at t, lit at t+4s, and not
  lit at t+5s. A new message at t+3s keeps it lit at t+7s. Use literal
  expected glyphs and tokens.
- Model tests also cover: pod selection with `←` `→`, `enter` and `esc`;
  `↑` and `↓` with the coordinator card; the one-project case; the drop
  merge; a lit path with no direct line; a lit rail; a message with no card
  for the sender; the footers at both widths.
- A test of `internal/agent` checks the `from` field with and without a
  record for the pane.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Board copies, grids, symbols, fixtures | V1, V2 |
| AC2 | The spec has decisions 62 to 72 and decision 14 is replaced | V-gate |
| AC3 to AC8 | The Agents screen matches the boards and the rules | V3, V4, V5 |
| AC9 | `hive agent send` logs `from` | V3, V4 |
| AC10 | The tests fail on the base and pass on the task branch | V3 |
| AC11 | The 10 boards that exist have approved new grids | V2, V4, V5 |

## Verification items

### V1 Boards

- Proves: AC1.
- Command: `cmp` of each staged board against its copy. Capture both with
  the Firefox command of the verifier skill. Put the profile folder and the
  captures in `.hive/tmp/` of the worktree. Remove them after the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grids

- Proves: AC1.
- Command: run the exporter on the design folder. Compare each grid with the
  committed grid, cell by cell, with `internal/lib/cellgrid`.
- Expected: the 4 new grids and the 10 approved grids of AC11 are
  committed. No other grid changes. Each of the 10 grids follows decisions
  62 to 72 and the 4 Live boards for each element that they show.

### V3 Tests red and green

- Proves: AC3 to AC10.
- Command: run the new tests on the base code, then on the task branch.
- Expected: each fails on the base and passes on the task branch.

### V4 Live check

- Proves: AC3 to AC9.
- Command: the live check of decision 43 for each of the 4 boards and the 10 boards of
  AC11 at its
  board size, with `COLORTERM=truecolor hive --demo FILE --size WxH`. Then:
  1. In the demo of `LiveOverview`, press `→`, `↑`, `↓`, `enter`, and
     `esc`. Read the pane after each key.
  2. In a scratch home (`HIVE_HOME` in `.hive/tmp/` of the worktree), run
     the real TUI with `HIVE_NO_STARTUP=1`. Write one inbox record between
     two agents of the scratch home. Read the pane at once and again after
     6 seconds.
  3. Run `hive agent send` from a pane of an agent record of the scratch
     home. Read the log event.
- Expected: each capture equals its grid. After each key the selection moves
  as AC3 and AC6 say. In step 2 the line is lit, then not lit. In step 3 the
  event has `from`, and the line lights.

### V5 Visual alignment

- Proves: AC3 to AC8.
- Command: run the real TUI in a herdr pane for the overview and the one-pod
  view, each at 120×40 and at 80×24, with a lit line in each. Compare the
  capture at the board size with the board grid, cell by cell. Look at each
  capture and check each point:
  - Each panel border, pod border, card border, and corner lines up.
  - The coordinator line meets each pod border and the manager card with no
    gap, and each joint has the correct glyph.
  - Each lit path is heavy and `info` from the sender to the receiver, with
    no gap, and one arrow toward the receiver.
  - The selected pod or card has the heavy `accent` border.
  - The legend and the footer show the text of the boards.
- Expected: each capture at the board size equals its grid. Each point
  passes at both sizes. Put each capture, each diff, and a table of the
  points with pass or fail in the evidence. A mismatch is a fail.
- The worker and the verifier each do this item.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`.
- Never open the Chat input on the real home. Never send chat.
- Do not write inbox records or send prompts in the real home for a test.
  Use the scratch home of V4.
- Close only the panes and the processes that you started. A demo quits
  with two `ctrl+c` presses.

## Out of scope

The rule for a pod with too many children for the tree. Health › Agents.
Each other screen. The canvas.

## Finish

Commit on the task branch with a subject such as `feat(tui): add the live
agent overview and lit message lines`. Push the task branch. Write the
evidence. Run `hive task complete`. Send the commit SHA and the V1 to V5
results to the manager with `hive inbox send --to manager-hive`.
