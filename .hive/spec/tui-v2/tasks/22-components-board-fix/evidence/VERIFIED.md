# Independent verification: 22-components-board-fix

Agent: `t22-verifier-1` (pi). Verdict: **fail**.

Worker tip: `8b4a2397`. Verified tip after `git merge --no-edit dev`:
`06f6f2fe3ba0afb60de48f8a58c495800bf2020b`.
The merge had no conflict. No source file was edited by the verifier.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Board | pass | `cmp /home/penguin/source/hive/plans/tui-v2/design/Components.dc.html .hive/spec/tui-v2/design/Components.dc.html` printed nothing and exited 0. |
| V2 Grid diff | pass | `cellgrid.ReadGrid` and `cellgrid.Compare` found exactly 46 changed cells. Row 14 columns 8, 9, 17, 18: dim to fg. Row 30 columns 65 through 106: accent to dim. No rune or bold flag changed. |
| V2 Export and other grids | pass | `go run ./internal/boardexport/cmd/boardexport -out SCRATCH/export .hive/spec/tui-v2/design` exited 0. Each exported grid equals its committed grid. `git diff --name-only dev HEAD -- '.hive/spec/tui-v2/design/*.grid'` lists only Components.grid. |
| V-gate Diff scope | pass | After the merge, `git diff dev...HEAD --stat` lists only the board, grid, task brief, STE baseline, and task evidence. |
| V-gate Full gate | pass | All five commands exited 0. gofmt and tidy printed nothing. Lint reported 0 issues. Full output is in `verifier-gate.txt`. |
| Required visual check | fail | Both fresh Firefox captures show the literal template `{{run.text}}`, with no rendered components. Both board folders lack support.js. The brief has no Design section. |

## Grid evidence

`verifier-grid.txt` lists each changed cell. The comparison used the dev grid
from `git show dev:.hive/spec/tui-v2/design/Components.grid` and the task grid.
The scratch Go program imported `internal/lib/cellgrid`; it did not implement
its own cell comparison. A fresh export was compared with every committed grid
by `cmp`. There are no new Go tests in this board-copy task, so there is no
red/green test item.

## Visual evidence and limit

Approved board path from AC1:
`/home/penguin/source/hive/plans/tui-v2/design/Components.dc.html`.

Task board path:
`/home/penguin/source/hive/projects/hive/worktrees/22-components-board-fix/.hive/spec/tui-v2/design/Components.dc.html`.

For each path, ran:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

Each profile folder was created fresh. Captures are `verifier-approved.png`
and `verifier-copy.png`. Each shows the same unrendered template. Each HTML
file loads `./support.js`; that file is absent in both folders. Thus identical
captures do not prove visual agreement of the components. No application screen
was changed by this task. No application UI implementation was tested.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| SubTabs gaps | Source has Goals, Queue, and [Tasks]; gap cells are spaces | differs | Capture has no rendered SubTabs. Grid comparison proves the four token changes only. |
| Sideways scroll mark | Source and grid retain the glyph ━ | differs | Capture has no rendered scroll mark. Grid comparison proves the 42 token changes only. |

The textual grid retains the active-state brackets and scroll-mark glyph when
colour tokens are ignored. The required rendered-screen check remains blocked.
Supply a working board preview and a Design section, or obtain an explicit
exception for this board-copy task before a pass verdict.

## Cleanup

Scratch comparison code, exported grids, Firefox profiles, and temporary logs
were removed after the evidence was copied. Only the final evidence files remain.
