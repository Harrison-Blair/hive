# Verification: 38-screen-friction

Verifier: `t38-verifier-1`. Date: 2026-10-22.
Code: merge `8180202c`, with `origin/dev` at `fbca94d5f9bd9f180806a9abdd938bc0e75c3836`.
Red-run base: `dev`, `13210c3f64f03f3228a5140688ac9f4c2c0106a9`.

## Verdict

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests, AC1–AC4 | pass | `verifier/base.txt`, `base-hint.txt`, `mutations.txt`, `restored-green.txt`, `green-latest.txt` |
| V2 Live check, AC1, AC2, AC4 | pass | Three board captures have 0 differences under the space-cell rule, also after the latest merge. See `verifier/latest-*-diff.txt`, `keys.json`, and the capture files. |
| V-gate Scope and gate | pass | `verifier/merge-latest.txt`, `scope-latest.txt`, `gate-latest.txt`. Each gate command exits 0. Formatting and tidy print nothing. |

## V1

Ran these commands from the worktree root:

```sh
go test ./internal/tui ./internal/tui/component -run 'Test(Friction|ListHint|ListRowCells|DetailsShowsMoreLinesOnOverflow|FrameRowsMatchEachScreenBoard)' -count=1 -v
go test ./internal/tui -run 'TestFriction(MatchesItsBoards|Keys)$' -count=1
```

Copied the new Friction tests to a detached scratch worktree at `dev`.
Changed only the two calls to the new `name()` method to
`strings.TrimSuffix(entry.file, ".md")` in that scratch test copy.
Ran `go test ./internal/tui -run '^TestFriction' -count=1 -v` there.
The three golden tests, responsive tests at supported widths, empty/long-data
test, keys test, normal-mode refresh test, and file-scroll test fail on the
base. The 79-column cases pass because the base already has the too-small
notice. Existing loader tests also pass on the base.

Copied the new `TestListHint` to the base and ran
`go test ./internal/tui/component -run '^TestListHint$' -count=1 -v`.
It fails to build because the base List has no Hint field.
The List cell rules and Details overflow rule are already on `dev` from task
42. Their tests pass in the task run.

The task run passes. Then made two deliberate changes in `friction.go`:

| Change | Observed failure |
| --- | --- |
| Occurrences token: dim to fg | Friction golden test: 139 cells differ; row 26 column 54 expects dim and gets fg. |
| List end key: G to Z | Keys test: G keeps scratch-hive-starts-real-agents selected; expected worker-gate-adds-golangci-lint. |

Restored both changes and ran the full focused command again. It passes.
No source change remains. Normal-mode tests use scratch homes and no live
coordinator or supervisor.

## V2

Built the task binary with `go build -o SCRATCH/hive .`.
Ran each board in an owned scratch herdr tab, with a scratch home marker:

```sh
HIVE_HOME=SCRATCH/home COLORTERM=truecolor SCRATCH/hive --demo BOARD.fixture.json --size WxH
```

The pane has 57 rows and sufficient columns. Read with `pane.read`, source
`visible`, format `ansi`, until two texts 500 ms apart agree. Each recorded
read has `truncated: false`; each pane has scroll offset 0. Cropped the
upper-left W×H cells with the recorded comparison tool from `live/compare.go.txt`.

| Board | Size | Differences under the live rule |
| --- | --- | --- |
| Friction | 120×40 | 0 |
| FrictionResolved | 120×40 | 0 |
| CompactFriction | 80×24 | 0 |

`verifier/live.py` records the live procedure. It uses terminal sequences
ESC [ 6 ~ and ESC [ 5 ~ for page down and page up because herdr rejects the
logical names `pgdown` and `pagedown`.

Checked j/k, arrows, g/G, page up/down, view brackets, tab, details scrolling,
esc, enter, help, and ignored action/filter keys. `keys.json` names each
capture and its footer. At 80 columns enter opens details; 1 does not switch
the screen while the overlay is open; esc returns to the list. Brackets
switch between Open and Resolved. Selection and scrolling match the tests.

The first key sequence opens an entry with no fixture file and correctly
shows `demo mode: no file`. A separate fresh run opens the fixture entry
scratch-hive-starts-real-agents. Assertions confirm file open, j scrolls
one line, G reaches the end, and esc restores the entry details at the top.
See `verifier/file-{enter,j,G,esc}.{ansi,json}`.

Captured fresh PNGs of all three boards and live screens. Captured the
compact overlay and open entry file too. Used Firefox with a new scratch
profile for each capture, `--headless --no-remote --window-size=1220,780
--screenshot OUT.png`. Board sources are the three absolute paths in the
brief. The live PNGs use the existing `live/ansi2html.py` converter.
PNG margins, line height, and background outside the cell grid differ
because of the capture wrappers. Cell comparisons are authoritative.

### Visual and no-colour review

Removed ANSI colour from each screen and checked the state signals.
Reviewed the six fresh board/live PNGs and the overlay/file PNGs.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | hive, home, needs you, clock | matches | Fixed fixture values. |
| Coordinator and supervisor | ●, working/live; compact coord/sup with ● | matches | Glyphs and words remain without colour. |
| Active screen | 4 Friction and underline | matches | Board frame. |
| Open view | [Open] | matches | Brackets identify the active view. |
| Resolved view | [Resolved] | matches | Brackets identify the active view. |
| List panel title | open · 15 / resolved · 20 | matches | Counts and view names. |
| List columns | ENTRY, OPENED / RESOLVED | matches | Wide dates use MM-DD. |
| Selected entry | ▶ | matches | Selection remains without colour. |
| Long entry names | … | matches | Task 43 corrected the Resolved board. |
| Compact columns | full date and SEEN count | matches | Same fixture records. |
| Compact bottom hint | enter opens details | matches | New List Hint row. |
| Details title and fields | entry name, title, status, opened, file | matches | Same fields and positions. |
| Open state | Open | matches | State is explicit text. |
| Resolved state | Resolved 2026-10-21 | matches | State is explicit text. |
| Seen / decision field | 3 times · last date / decision text | matches | Open and Resolved differ as the boards show. |
| Section headings | SUMMARY, REPRODUCTION, OCCURRENCES, MITIGATION, RESOLVED BY | matches | Upper-case section names. |
| Details body | fixture text and wrapped lines | matches | Whole-grid compare. |
| Overflow and scroll | ↓ 5 more lines, ┃ | matches | Shared Details rule from task 42. |
| Focus and key hints | [list] / [details] plus key words | matches | Footer identifies focus without colour. |
| Compact details overlay | entry name · esc back | matches | Spec decision 9; no separate approved overlay board. |
| Open file | .md title · esc closes, file text | matches | Retained file reader behavior; no separate approved file board. |

## V-gate

The first merge was already up to date. At the manager's request, fetched
`origin/dev` again and merged `fbca94d5`. There was one merge-only conflict
in README Manager decisions. Kept both Friction and Logs decision blocks,
as the manager instructed. No Go file conflicted. Committed merge `8180202c`.
Ran the focused tests, all three live comparisons, and the full gate again.
They pass after the merge.

Ran `git diff origin/dev...HEAD --stat` after the latest merge. Only allowed
source paths, the brief, Manager decisions, baseline, and task evidence differ. The source edits in
shared files are confined to Friction registration, its file map, its tests,
and List Hint. Details and selected-cell rules already came from task 42.

The final post-merge gate output is in `verifier/gate-latest.txt`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

All five exit 0. No known-flake exception was necessary.

## Cleanup

Closed all owned scratch tabs (`wA1:t2`, `wA1:t3`, `wA1:t4`). The workspace
tab list shows only the verifier tab. Removed the detached base worktree, binary,
comparison executable, scratch home, HTML files, and browser profiles.
Confirmed `.hive/tmp/t38-verifier` no longer exists. Evidence remains in
this task folder. No source edit remains from the deliberate-break checks.
