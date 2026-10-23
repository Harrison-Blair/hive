# Health

Health shows the checks of `hive doctor` as a grid of tiles, the details of the selected tile, and the agents table. The two views Home and Agents of `tui-v2` decision 18 go away. The restart rules of `tui-v2` decisions 19 and 34 of `general-agent` stay.

## Layout at 120×40

| Rows | Region |
| --- | --- |
| 0 | title row: `health` in `accent` bold, `· N checks · K ok · W warn · checked HH:MM:SS · R checks again` in `dim` |
| 2 to 6 | the check grid, one row for each area |
| 8 to 16 | the details box of the selected tile |
| 18 to 29 | the agents table |
| the last rows | the status rows |

## Rules

1. The grid has one row for each area, in this order: `herdr`, `home`, `agents`, `supervisor`, `harness`. The area name is in `dim` bold in the first 11 cells. Each tile is a mark and a text, with 4 cells between tiles. The mark is `✓` in `done` for a passed check and `!` in `accent` for a warn or a fail. The selected tile has `▶` in `accent` bold before its mark and its text in `accent` bold.
2. The first selection is the first tile that did not pass. When each tile passed, it is the first tile.
3. The details box shows the selected tile. Its title is the mark and the check name. Its body shows the result rows, then `WHAT IT MEANS` and `NEXT STEP` in `accent` bold with their texts, then `source hive doctor · checked HH:MM:SS` in `dim`. For a passed check, the body shows the result rows and the source line only.
4. The agents table lists each agent record, including the coordinator and the supervisor. The columns are `AGENT`, `ROLE`, `HARNESS`, `STATE`, `PANE`, `STARTED`, and `RESTART`. The state dot follows `tui-v2` decision 55. The `RESTART` column shows `r restart` on the coordinator row and the supervisor row. The row under the table shows `restart: coordinator and supervisor only · hive asks to confirm only when the agent is live · other rows show a notice` in `dim`.
5. The grid and the agents table are the two boxes of README decision 6. `ctrl+]` moves the focus. In the grid, `↑` and `↓` move between rows, and `←` and `→` move between tiles. In the table, `↑` and `↓` select a row. The table scrolls with the `┃` thumb when its rows do not fit.
6. `r` on the coordinator row or the supervisor row opens the restart prompt on the notice line: `restart the coordinator? y restarts · any other key cancels`, in `accent` bold. `y` restarts. Each other key cancels. Hive asks only when the agent is live. For an agent that is not live, `r` restarts at once. `r` on each other row shows the notice `restart: coordinator and supervisor only`.
7. `R` runs `hive doctor` again. While it runs, the title row ends with `· checking…` and each tile is `dim`. When it ends, the tiles and the details box show the new result. When `hive doctor` fails to run, the notice line shows `error: hive doctor: TEXT` and the tiles keep their last result.
8. `a` attaches to the pane of the selected agent row. `p` peeks at it. In demo mode, both show the demo notice.

## States

| State | Trigger | What shows | Board |
| --- | --- | --- | --- |
| one warn | one check did not pass | rules 1 to 4 | `GfHealth` |
| all green | each check passed | rule 2, the short details box of rule 3 | `GfHealthGreen` |
| restart prompt | `r` on the supervisor row | rule 6 | `GfHealthRestart` |
| restart notice | `r` on a manager row | the notice of rule 6 | no board |
| checking | `R` | rule 7 | `GfHealthChecking` |
| doctor failed | `hive doctor` did not run | the error of rule 7 | no board |
| table focus | `ctrl+]` | the table has the `accent` border | no board |
| many agents | more agent rows than fit | the `┃` thumb on the table | no board |
| stopped coordinator | the coordinator record has the status stopped, unresolved, or stale | the state column, `r restart` without a prompt | no board |
| old supervisor | the supervisor runs an old binary | the `supervisor binary` tile shows `!` and `old` | no board |

## Keys

| Focus | Key | Action |
| --- | --- | --- |
| grid | `↑`, `↓`, `←`, `→` | select a tile |
| table | `↑`, `↓`, `g`, `G` | select a row |
| grid, table | `ctrl+]` | move the focus to the other box |
| table | `r` | restart, on the coordinator or supervisor row |
| table | `a`, `p` | attach, peek |
| any | `R` | run `hive doctor` again |
| any | `esc` | Chat |
