# Handoff 36-screen-work-queue (work in progress)

State: the first failing test, `TestWorkQueueMatchesItsBoard` in
`internal/tui/work_queue_test.go`, is written. It fails on the base (708
cells differ). There is no implementation yet.

Plan from the board (`WorkQueue.dc.html`, function `screen`):

- List panel title: `queue · <project> · N tasks · M stop(s)`. N counts all
  the tasks of the project, including the landed tasks.
- Body: a blank row, then groups. Each group has a dim header (`GRANT  <grant>`
  or `NO GRANT`), a blank row, and then the task rows. A blank row separates
  two groups. The grant comes from the goal wave of the task, and then from
  `delivery.grant`.
- Rows form a dependency tree in the group. A root row has 2 spaces with the
  border token. A row at depth d≥1 has 2(d-1)+2 fg spaces, `└▶` in border,
  and a space. The name is at content offset 2 (root) or 2d+3, fit to 30
  cells with `…`. The status is at content offset 38, fit to 10 cells. Its
  token is done for landed and info for other states.
- The selected row has `▶` in accent bold and a space in place of the first
  2 cells. The name is in accent bold and padded to 30 cells.
- A stop row follows a stopped task: `! stop: <reason>` in accent at the name
  offset. A stop is NeedsUser text, or an acting agent with the status
  `unresolved` ("worker unresolved").
- Details lines: a blank row, then the fields title, status (info or done),
  goal (`id · wave X` or none), grant, worker (`name · record status`, in
  accent on a stop), and depends. Each value is padded to width-14 with its
  token. Then a blank row, `STOP` with 2 lines, a blank row, `NEXT STEP` with
  the dim line "Ask the coordinator in 2 Chat.", a blank row, and the hint
  line `t open task in Tasks  s show stops only`. The keys are in accent
  bold.
- Wiring: add the `panels` hook (list and details) to `frameScreen`. Set it
  only in the tabWork entry for "Queue". Use it in `contentView` and in
  `detailsPanel`. Add agents to `goalsQueueData` (store collection
  `agents`, and `data.Agents` in the demo loader).
- Keys: `t` opens the task in Tasks. `s` toggles stops only. `enter` from 120
  columns moves the focus to details. `/` has no filter bar for Work in the
  frame. Report this gap to the manager.
- After that, remove the Queue part of `goals_queue.go` and its tests.
