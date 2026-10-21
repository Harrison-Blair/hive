# Independent verification: 22-components-board-fix

Agent: `t22-verifier-2` (pi). Verdict: **pass**.

Verified task tip: `b25cc1e7f195e4ae4a904f60356dc5f7a9ba72da`.
Base `dev`: `9d4def4dc72fac2947a79582392683942043a6e7`.
`git merge --no-edit dev` reported `Already up to date.`
No source file was edited by the verifier.

This report replaces the previous failed verdict. The previous failure was
an absent Design section and an absent support script. The brief now has the
Design section. Both design folders now have the support script. The existing
`verifier-approved.png` and `verifier-copy.png` are historical captures from
that failed run. They are not the captures of this run.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Board | pass | `cmp` of the approved Components board and the task copy printed nothing and exited 0. |
| V2 Grid diff | pass | `cellgrid.ReadGrid` and `cellgrid.Compare` found exactly 46 changed cells. Row 14 columns 8, 9, 17, 18: dim to fg. Row 30 columns 65 through 106: accent to dim. No rune or bold flag changed. `verifier-grid.txt` lists each cell. |
| V2 Export and other grids | pass | A fresh export equals every committed grid. Only Components.grid differs from dev. |
| V3 Board render | pass | Fresh Firefox captures of Components, Main, and Work render with no literal `{{`. The approved Components capture and its task copy are byte-identical. The support scripts are byte-identical. |
| V-gate Diff scope | pass | `git diff dev...HEAD --stat`, after the merge check, lists only the two Components files, support.js, the task brief, testdata/ste-baseline.json, and task evidence. |
| V-gate Full gate | pass | All five commands exited 0. gofmt and tidy printed nothing. Lint reported `0 issues.` Full output is in `verifier-gate.txt`. |

## Commands and grid check

All commands ran from the task worktree root.

```sh
cmp /home/penguin/source/hive/plans/tui-v2/design/Components.dc.html .hive/spec/tui-v2/design/Components.dc.html
cmp /home/penguin/source/hive/plans/tui-v2/design/support.js .hive/spec/tui-v2/design/support.js
git show dev:.hive/spec/tui-v2/design/Components.grid > .hive/tmp/verify22/dev.grid
go run .hive/tmp/verify22/check.go .hive/tmp/verify22/dev.grid .hive/spec/tui-v2/design/Components.grid
mkdir .hive/tmp/verify22/export
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verify22/export .hive/spec/tui-v2/design
for f in .hive/spec/tui-v2/design/*.grid; do
  cmp "$f" ".hive/tmp/verify22/export/$(basename "$f")" || exit
done
git diff --name-only dev HEAD -- '.hive/spec/tui-v2/design/*.grid'
git diff dev...HEAD --stat
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

The scratch program read both grids with `cellgrid.ReadGrid`, compared them
with `cellgrid.Compare` without masks, and printed `cellgrid.FormatDiffs`.
It required exactly 46 differences at the positions listed above. Each
position had to have the required old and new tokens, the same rune, and the
same bold flag. Each assertion passed. Coordinates are zero-based.

The first export attempt failed because its output folder did not exist.
After `mkdir`, the export and every `cmp` exited 0. This was a scratch setup
error. There is no new test item in this board-copy task, so no red/green test
run applies. The gate ran the existing tests.

## Visual check

Approved board:
`/home/penguin/source/hive/plans/tui-v2/design/Components.dc.html`.

Copy folder: `.hive/spec/tui-v2/design/` in this worktree.

For each capture, a new profile folder was created below
`.hive/tmp/verify22/`. Each Firefox command exited 0:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

Captured and opened the approved Components board, the task Components board,
Main, and Work. No capture showed a literal `{{`. Main showed the project
agent graph. Work showed its task list and details.

The 780-pixel capture clips the bottom of Components. Thus both Components
boards were also captured at `--window-size=1220,1400` with new profiles.
These full captures were byte-identical. The full task capture was opened
and inspected through its final Notice line.

| Capture | SHA-256 |
| --- | --- |
| approved and copy, 1220×780 | f57528eabd56aaf3e0ba0489cf15cd0bc351dfe28990c63ba2951da44a12b29e |
| Main, 1220×780 | 029536facb26a33f8b006594290c463bfade6b2cdde5ab6b90f361155456a078 |
| Work, 1220×780 | 75a85c51b90156049d7e4df466143ec9c6c3bd241270aece7df0a69971c99c91 |
| approved and copy, 1220×1400 | b372ca50e5ffa1040e3b8a216ed676bb3dc8af1e1395e8236ae70c5b81bcf4c6 |

The table compares the Components copy with its approved board. For the
colour-independent check, only the `t|` text rows of Components.grid were
printed. Token and bold rows were excluded. State words and glyphs remain.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Title | HIVE TUI COMPONENTS | matches | Identical full capture. |
| Palette | fg, border, dim, accent, working, done, info, user, code | matches | Each swatch has a text label. |
| Header | working, live, needs you 0 | matches | State words remain. |
| TabBar | 1 Agents and underline ━ | matches | Active tab retains its underline. |
| SubTabs | Goals, Queue, [Tasks] | matches | Brackets mark the active tab. Four gap tokens are fg; V2 checks them. |
| Panel | unfocused, focused | matches | Both states have words. |
| Viewport + ScrollMark | line 41 of 120; ┃ and ━ | matches | Marks retain distinct glyphs. The horizontal mark uses dim; V2 checks all 42 cells. |
| List | ▶ 3-tui-restart; landed | matches | Selection and state have text signals. |
| Details | status landed; ✓ AC1 | matches | State and acceptance retain words and a glyph. |
| FilterBar | / level:warn agent:manager-hive █ | matches | Text and cursor remain. |
| Input + Chat text | > restart the supervisor please█ | matches | Prompt and cursor remain. |
| Card | ● working, ◐ idle, ○ stopped | matches | Each agent state has a word and a distinct glyph. |
| Popup | manager-hive · esc closes; a attach · p peek · t task | matches | Actions remain in text. |
| KeyHints | [graph] and action words | matches | Region and actions remain in text. |
| Notice | below 80×24; too-small notice | matches | The board contains the explanatory line, not a separate notice screen. |

This task changes no application screen. No live application UI behavior is
claimed. V3 samples the three boards specified by the brief; the other board
screens were not captured individually.

## Cleanup

The fresh captures, Firefox profiles, scratch comparison program, exported
grids, and scratch logs were removed after the results above were recorded.
The final gate output and cell list remain in this evidence folder. No
`VERIFIED.md` was created at the checkout root.
