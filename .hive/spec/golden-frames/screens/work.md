# Work

Work shows one goal as a wave board. Each wave is a column. Each task is a card. The three views Goals, Queue, and Tasks of `tui-v2` decision 16 go away. Their data shows on the board, in the task popup, and in the dependency popup.

## Layout at 120×40

| Rows | Region |
| --- | --- |
| 0 | title row: `work · GOAL` in `accent` bold, `· TITLE` in `dim`, right text `[ ] goal I of N` |
| 1 | `wave W of M · L/T landed · S needs you · grant GRANT` in `dim` |
| 3 to 36 | the wave columns, 28 cells wide, 2 cells between them, 1 cell margin |
| 38 | the legend `● landed  ◐ verified  ○ assigned or created  ! needs you  ─▶ depends on` |
| the last rows | the status rows |

## Rules

1. `[` and `]` change the goal. The cycle holds each goal of the project in creation order. The pseudo-goal `no goal` is last in the cycle when a task has no goal. The pseudo-goal has one column named `no goal`. When every task has a goal, the pseudo-goal is not in the cycle.
2. Row 1 shows `· N stops elsewhere` in `accent` when a stop exists on a goal that is not on the screen. A stop is a task with `needs_user`, or a task whose worker is `unresolved`.
3. A wave column has a rounded border and the title `W · STATE · L/T`. STATE is `done` when each task landed. It is `running` when a task has the status assigned or verified. Else it is `planned`. The column of the selected card has the heavy `accent` border. The foot of the column shows `done when DONE_WHEN` in `dim`, and `✓ done` in `done` for a done wave.
4. A task card is 24×4 cells. Line 1 is the dot and the task ID. Line 2 gives the state. The forms are `landed · COMMIT`, `verified · VN pass`, `WORKER · AGE`, `verify fail · VN` in `accent`, `waits on IDS`, `needs you · TEXT`, and `stop · unresolved`. The dot is `●` in `done` for landed, `◐` in `working` for verified, `○` in `dim` for assigned or created, and `!` in `accent` for a stop. The selected card has the heavy `accent` border. A card of a working worker has the `working` border, as `tui-v2` decision 60 gives.
5. A dependency between two cards in neighbouring columns draws a rail. The rail goes from the right border of the first column to the left border of the second column. The rail goes to the row of the name line of the second card. Each other dependency shows only as `waits on IDS` on line 2 and in the dependency popup.
6. Columns scroll sideways when more columns exist than fit. The board moves only as far as necessary to keep the selected card on the screen. The last body row shows `←→ at the edge scrolls · waves A–D of M` in `dim` on the left and the `━` thumb on the right.
7. A column scrolls vertically when its cards need more rows than it has. The column moves only as far as necessary to keep the selected card on the screen. The `┃` thumb is on the right border of the column.
8. `←` and `→` select the card at the same row in the column before or after. When that column has fewer cards, they select its last card. `↑` and `↓` select the card above or below. `g` and `G` select the first and the last card of the column.
9. `enter` opens the task popup. The popup is the details box of `tui-v2` decision 16. It shows the title, the status, the goal and wave, the worker, the verifier, the branch, the merge, and the dependencies. Then it shows the acceptance list and the evidence list. Inside the popup, `b` opens the brief, `v` opens `VERIFIED.md`, `d` opens the dependency popup, `↑` `↓` `pgup` `pgdn` scroll, and `esc` closes.
10. `d` opens the dependency popup from the board or from the task popup. It lists each dependency of the selected task with its state, and each task that depends on it.
11. `s` shows only the stop cards in each column. The title row then shows `· stops only` in `accent`, and the hint reads `s all`. `s` again shows all cards.
12. Hive hides a cancelled task by default. `h` shows the cancelled cards. A cancelled card has `×` in `dim` as its dot and `cancelled · REASON` on line 2. The foot of a column with hidden cards shows `h shows N cancelled` in `dim`.
13. With no goal and no task, the board shows the notice `no goals · hive goal create starts one` in `dim` at the center. With a goal that has no waves, the board shows one empty column `no waves`. A wave with no tasks shows its column with the foot text only.

## States

| State | Trigger | What shows | Board |
| --- | --- | --- | --- |
| goal board | a goal with waves and tasks | rules 3 to 5 | `GfWork` |
| no goals | no goal and no task | rule 13 | `GfWorkEmpty` |
| no goal pseudo-goal | a task with no goal, with a stop | rule 1, a `!` card | `GfWorkNoGoal` |
| no waves | a goal with no waves | rule 13 | no board |
| wide and tall | 6 waves, a column with 8 cards | rules 6 and 7 | `GfWorkScroll` |
| stops only | `s` | rule 11 | `GfWorkStops` |
| cancelled shown | `h` | rule 12 | `GfWorkCancelled` |
| task popup | `enter` | rule 9 | `GfWorkTask` |
| dependency popup | `d` | rule 10 | `GfWorkDeps` |
| verify fail | a task with a failed verification | the card form of rule 4 | `GfWork` |
| done goal | each wave has each task landed | each column `done`, row 1 `T/T landed` | no board |

## Keys

| Key | Action |
| --- | --- |
| `←`, `→` | select a card in the column before or after |
| `↑`, `↓`, `g`, `G` | select a card in the column |
| `[`, `]` | previous or next goal |
| `enter` | task popup |
| `b`, `v` | brief, `VERIFIED.md`, inside the task popup |
| `d` | dependency popup |
| `s` | stops only, or all |
| `h` | show or hide cancelled |
| `esc` | close a popup, else Chat |
