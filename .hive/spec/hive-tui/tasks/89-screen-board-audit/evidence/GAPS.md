# Screen board audit

The `dev` binary was built from commit `7824c1c0` and run in Herdr pane `wPG:p1`. The pane size was 207 x 57. The seven captures are `herdr pane read` results. The brief names seven screen captures; I used `Main.dc.html` as the shared-header reference and did not audit Agents layout or data, which T3 and T4 cover. `Chat.txt` includes the sent `278` message. `INCIDENT.md` records that event.

Size: S is a small label or control gap. M is a missing panel or navigation gap. L is a major view gap.

Do not include the known T1, T2, T3, or T4 faults in this report. The brief names them as separate tasks.

## Chat

| ID | Board part | What the screen shows | Decision or G item | Size |
| --- | --- | --- | --- | --- |
| C1 | Noise choices: `all`, `landings and decisions`, and `decisions only` | The screen cycles `all`, `messages`, and `off`. The filter works, but the levels differ. | Decision 4 requires a noise filter. It does not set the levels. The board has no later override. | S |

Matches: coordinator status, chat and folded event timeline, separate inbox feed, and chat input opened by the `i` control. The capture leaves the input closed under the chat-safety scope.

## Goals

| ID | Board part | What the screen shows | Decision or G item | Size |
| --- | --- | --- | --- | --- |
| G1 | `t` opens a wave's tasks; `o` opens its design | The screen shows goal and wave progress. Its hints offer goal selection and `ask in Chat`; it has no wave-task or design view shortcut. | Decision 14 allows read-only views. G12 removes changes and decision controls. These two controls only read data. | M |

Matches: goal and wave cards, progress to the selected done-when point, task IDs, and the unassigned task list.

## Queue

| ID | Board part | What the screen shows | Decision or G item | Size |
| --- | --- | --- | --- | --- |
| Q1 | Header counts for work that runs, stops, and waits behind a stop | The shared header shows project, agent, supervisor, and needs-user counts. It shows no queue counts. | Decision 2 and G8 define the queue and its stops. The board gives the counts. | S |
| Q2 | One panel per project and no view cycle | The screen offers `p project` to cycle projects. It builds panels only for projects with open tasks, so qmeter has no panel in this capture. | Decision 9 requires one panel per project and no view cycle. | M |
| Q3 | `enter` opens the selected task | The screen has no task selection or open-task control. Its hints show the stop filter and project filter. | Decision 14 allows read-only views. G12 does not remove read-only task details. | M |

Matches: task chains grouped by project, grant labels, dependency order, and task state.

## Tasks

| ID | Board part | What the screen shows | Decision or G item | Size |
| --- | --- | --- | --- | --- |
| T1 | Separate task details and evidence panels, with the brief and verification history | The screen puts the task list and selected-task details in one vertical panel. It shows evidence names and the current pass, but no evidence history panel. | Decision 3 says Tasks shows evidence. | L |
| T2 | `v` opens `VERIFIED.md`; `d` opens a dependency graph | The screen offers `enter` for evidence and `b` for the brief. It has no `v` or `d` control. | Decision 14 allows read-only views. G12 removes state changes, not these views. | M |
| T3 | A note defines `landed` and `verified` | The screen uses both status labels but gives no definition. | G8 defines landed. | S |

Matches: task filters, project and status fields, agent and verification fields, dependencies, branch, grant, needs-user text, and evidence names.

## Friction

| ID | Board part | What the screen shows | Decision or G item | Size |
| --- | --- | --- | --- | --- |
| F1 | The list has an issue and a `goal` column | The screen shows a `decision` column in place of `goal`. | Decision 3 gives Friction its own screen. It does not change the board fields. | M |
| F2 | Selected entry details sit in a right-hand panel | The screen stacks entry details below the list in the same panel. | Decision 3 gives Friction its own screen. It does not change the board layout. | M |

Matches: open/resolved filters, newest-first entries, and selected-entry summary and reproduction text.

## Health

| ID | Board part | What the screen shows | Decision or G item | Size |
| --- | --- | --- | --- | --- |
| H1 | Framed home and agent panels, a separate warning panel, and a capacity panel | The screen shows two unframed check columns, one `DETAIL` line, and a plain capacity line. It has no separate warning or capacity panels. | Decision 6 and G14 define the health checks. G12 removes restart and cleanup controls; the screen shows `ask in Chat`. | M |

Matches: home and agent checks, a supervisor-binary warning, stale-worktree status, and a note that capacity is out of scope.

## Logs

Gaps: none.

Matches: date, level and metadata filters, follow state, event rows, selected-event JSON, and request filtering.

## Proposed fix tasks

The order below is a proposed order. Each task covers gaps in the same screen code.

1. **Restore Goals and Queue navigation and queue summaries**
   - Gaps: G1, Q1, Q2, Q3.
   - Files: `internal/tui/goals_queue.go`, `internal/tui/header.go`.
2. **Restore the task evidence view and read-only shortcuts**
   - Gaps: T1, T2, T3.
   - Files: `internal/tui/task_screen.go`.
3. **Match the Friction list and detail layout**
   - Gaps: F1, F2.
   - Files: `internal/tui/friction_screen.go`.
4. **Match the Health warning and capacity panels**
   - Gaps: H1.
   - Files: `internal/tui/health_screen.go`.
5. **Match the Chat noise choices**
   - Gaps: C1.
   - Files: `internal/tui/chat_screen.go`.
