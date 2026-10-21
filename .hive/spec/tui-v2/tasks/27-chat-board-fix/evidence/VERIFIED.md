# Independent verification: 27-chat-board-fix

Verifier: `t27-verifier-1`. Result: **pass**.
Worker commit checked: `876e30fa`.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Boards | pass | Both `cmp` commands exit 0. Fresh Firefox PNG captures of each approved board and each spec copy are byte-identical. No final capture shows `{{`. |
| V2 Grid diff | pass | `cellgrid.ReadGrid` and `cellgrid.Compare`, without masks, find 314 changed Chat cells and 170 changed CompactChat cells. Each change is in the allowed scope. Each report cell agrees in rune, token, and bold flag. All other grids equal `dev`. |
| V-gate Diff scope and gate | pass | `git merge dev` reports `Already up to date.` The diff contains only allowed source files and task evidence. All five gate commands exit 0. Format and tidy commands print nothing. |

## V1: board capture and visual check

For each of `Chat` and `CompactChat`, ran:

```sh
cmp /home/penguin/source/hive/plans/tui-v2/design/BOARD.dc.html .hive/spec/tui-v2/design/BOARD.dc.html
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

Each capture has a separate new profile in `.hive/tmp/verify-27/`.
The first attempt needed explicit profile directory creation. The first
approved Chat capture was incomplete. A second capture with that profile
rendered the full board. Checked all four final PNG files visually and with
`cmp`. Both final pairs are byte-identical.

| Board | SHA-256 of both final PNG captures |
| --- | --- |
| Chat | `d7724cd39b9214c5101398db1d7714da429fd7960b3f356131683505e95020fe` |
| CompactChat | `2fcf70145a397703aa89106940666e32ba1ba513ba35bd4236b9688988ae3043` |

Read the `t|` lines of both exported grids without token or bold data. State
labels and glyphs remain present. The tables remain readable without colour.
The following rows apply to both boards unless a board is named.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and path | `hive`, `~/source/hive` in Chat | matches | Same text and position as approved. |
| Coordinator status | `● working`; compact `coord ●` | matches | State glyph remains without colour. |
| Supervisor status | `● live`; compact `sup ●` | matches | State glyph remains without colour. |
| Attention and clock | `needs you 0`, `14:22` | matches | Same text and position. |
| Navigation | `2 Chat`, heavy underline `━` | matches | Active tab has a non-colour signal. |
| Chat panel | `chat`; Chat also `noise normal` | matches | Same border and title. |
| Earlier feed, Chat | `↑ 42 earlier` | matches | Same count and arrow. |
| Earlier user message, Chat | `14:02`, `you`, restart request | matches | Same text and wrapping. |
| Event, Chat | `14:19`, goal and worker status `live` | matches | Same event text. |
| Coordinator message | `14:21 coordinator`, `Interrogate, round 2` | matches | Same message and wrapping. |
| Question and inline code | `spawned=false`, `agent stop` | matches | Same readable text. |
| Table header | `Choice`, `Answer`, `Pros`, `Cons` | matches | Normal weight; same separators. |
| Table rule | `─`, `┼` | matches | Same column widths and positions. |
| Table option A | `A (REC)`, full Answer | matches | Recommendation has a text label. |
| Table option B | `B`, Answer, Pros, Cons | matches | Same text and spacing. |
| Table option C, Chat | `C`, Answer, Pros, Cons | matches | Same text and spacing. |
| Recommendation | `• Recommended: A` | matches | Recommendation is explicit without colour. |
| Reply instructions, Chat | `Reply like 1A, 2A, 3A, or simplify N.` | matches | Same text. |
| Command block, Chat | `hive agent stop`, `hive agent spawn` | matches | Same text and indentation. |
| User reply | `14:24 you`, `1A, 2A, 3B` | matches | Compact reply is on row 18, column 2. |
| Scroll marker | `┃` | matches | Glyph remains without colour. |
| Composer | `message`, `>`, `█` | matches | Same content and cursor. |
| Focus and shortcuts | `[input]`, `enter send`, other key labels | matches | Focus is explicit without colour. |

## V2: independent cell checks

A scratch Go program loads the `dev` and task grids with `cellgrid.ReadGrid`
and compares them with `cellgrid.Compare`. It prints every changed cell,
including old and new values. Full output: [verifier-grid.txt](verifier-grid.txt).
Coordinates are zero-based.

Expected table rows come directly from sections 1 and 3 of the external
Glamour report, parsed independently of the board source. Check each rune,
`fg` or `border` token, and normal weight, including trailing spaces through
the full wrap width. Report comparison:

| Board | Red run: dev mismatches | Green run: task mismatches |
| --- | --- | --- |
| Chat | 313 | 0 |
| CompactChat | 150 | 0 |

Chat table: rows 15–19, columns 23–116. Rule column widths: 19, 35, 18, 19.
The only other change is row 21, column 37: a normal-weight space in `fg`.
CompactChat table: rows 10–13, columns 2–77. Rule column widths: 9, 34, 17, 13.
Its other changes are in rows 17 and 18. Row 17 keeps `14:24 you`; row 18
starts `1A, 2A, 3B` at column 2. Row A contains the full Answer text.

Ran `go run ./internal/boardexport/cmd/boardexport .hive/spec/tui-v2/design`
again. Exit 0, no output, and no worktree changes. `git diff dev --name-only
-- '*.grid'` lists only Chat.grid and CompactChat.grid.

## V-gate: scope and full output

Ran `git merge dev` before the gate: `Already up to date.`
`git diff dev...HEAD --stat` showed nine files: the four named design files,
the task brief, V1.md, V2.md, V-gate.md, and testdata/ste-baseline.json.
No exporter, other board, or mask changed. The existing STE registration
already includes this brief; the baseline adds its zero entry.

Full fresh gate output: [verifier-gate.txt](verifier-gate.txt).

| Command | Exit | Output |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | 0 | empty |
| `go mod tidy -diff` | 0 | empty |
| `go vet ./...` | 0 | empty |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `go test -race ./...` | 0 | All tested packages pass. |

## Cleanup

Removed the scratch program, report data, four captures, and Firefox profiles
after recording the results. Preserved only this verdict and the two evidence
logs. No application screen changed; the visual check covers the board copies.
