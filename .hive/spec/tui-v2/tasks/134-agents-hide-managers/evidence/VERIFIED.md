# Verification of task 134 after rework

Verdict: **pass**. V1 to V4 pass, including the coordinator regression and the additional zero-pod keys.

- Date: 2026-10-23.
- Verifier: `t134-verifier-2`.
- Worker commit: `fd5401c080103688c527c3c23a3181b753f27c14`.
- Verified merge: `7a264df8416d315d638a65cddde1cd0a462bd183`.
- Current `dev` and `origin/dev`: `b34ee38d1cbb91eedc9d52ef1db3a9497fa10764`.
- The verifier merged current `origin/dev` before the final gate. This merge adds friction documentation only.
- Approval: ask `a5`, coordinator record 536, package `219eddba`, final approved grids `21495e2e`.
- First verdict: **fail**, by `t134-verifier-1`, commit `b257f299`.
  Its exact text remains in `verifier/VERIFIED-first-fail.md`.
  Its captures and reproduction remain unchanged in `verifier/`.
- This verification changes no product source, board, golden grid, worker evidence, or tracker.

All evidence paths below are relative to this folder.

## Verification items

| Item | Command or check | Result | Evidence |
| --- | --- | --- | --- |
| V1 | Base red runs, guard mutations, focused tests with `-count=50` and `-race -count=50` | pass | `verifier-2/tests-summary.log`, `red-*.log`, `guard-*.log`, `green-*.log`, `extra-*.log` |
| V2 | Export each board, compare with `cellgrid.Compare`, run each golden check, check approval scope | pass | `verifier-2/export.log`, `grid-compare.log`, `golden.log`, `approval-and-scope.log` |
| V3 | Build the task binary, capture both views at both sizes, check keys and states | pass | `verifier-2/live/`, `live-summary.log`, `extra-live-summary.log`, `live-grid-compare.log`, `extra-live-grid-compare.log` |
| V4 | Merge current `dev`, check diff scope, run the five gate commands | pass | `verifier-2/diff-stat.log`, `gate.log` |

## V1: red and green runs

The verifier used an isolated source copy for every red run and mutation.
The task worktree source stayed unchanged.
`verifier-2/tests.py.txt` records the exact commands and mutations.
Each log records its command, output, and exit status.

- The 12 original focused tests ran with base `6a586f1a` source. Nine failed for the expected manager, pod, hint, or selection difference.
- The three rework tests ran with pre-fix `b257f299` source.
  Coordinator selection failed for up and `k` at both widths.
  Details also failed when the verifier selected the coordinator before hiding the last pod.
- The no-registry guard passed before the fix.
  A mutation removed the registry check from `coordinatorCard`.
  The guard then failed at both widths because details opened over the notice.
- Each other base-passing guard failed with its guarded behavior broken:
  an unwanted zero-count hint, a hidden-worker child outside home, and reversed pod order.
- A notice-position mutation failed its literal row check: row 6 instead of row 5.
- The verifier restored the scratch source after every mutation.

Both focused commands ran from the task worktree:

```sh
go test ./internal/tui/ -run "$R" -count=50 -v
go test -race ./internal/tui/ -run "$R" -count=50 -v
```

`$R` selects each test in `agents_hide_test.go` and `TestAgentsPodsFollowProjectRecordOrder`.
Both commands exited 0. Each produced 1150 PASS lines and no FAIL line.

The first verifier reproduction ran again, with an additional independent zero-pod test.
`verifier-2/zero-pods-regression.go.txt` is its durable source.
It checks left, right, down, enter, up, `k`, deselection, and a remembered pod-view state.
It also rejects a phantom pod view after enter with no selected coordinator.

Both tests failed on `b257f299` at both widths for the expected reason.
Both passed on the task source with `-race -count=50`: 300 PASS lines.
See `verifier-2/extra-red.log` and `extra-green.log`.

## V2: approved grids

Commands:

```sh
mkdir -p .hive/tmp/verify134-second/export
go run ./internal/boardexport/cmd/boardexport -out .hive/tmp/verify134-second/export .hive/spec/tui-v2/design
go test ./internal/tui/ ./internal/boardexport/ -run 'Board|Golden|Grid' -count=1 -v
```

The first exporter attempt lacked its scratch output directory.
After the verifier created the directory, the exact exporter command exited 0.

- All 47 exports match their committed design grids.
- All nine changed exports match their approved grids in `new-grids/`.
- All 56 comparisons use `cellgrid.ReadGrid` and `cellgrid.Compare`, with no mask.
  They compare each rune, colour token, and bold flag.
- The golden test command exited 0: 201 PASS lines, no FAIL line.
- Only the nine approved board/grid pairs differ from current `dev`.
  Their changes remain on title row 3. No board fixture or mask changes.
- `git diff 21495e2e HEAD -- .hive/spec/tui-v2/design/` is empty.
  The rework preserves every approved design file.

## V3: independent live checks

The verifier built `.hive/tmp/verify134-second/hive` with `go build`.
The initial scratch tabs exposed a terminal-size difference: layout 120x40, terminal 119x40.
Its cause remains **unknown**.
`verifier-2/geometry.md` records the exact commands, observations, and isolated-server workaround.

The successful checks used an isolated herdr 0.9.5 server, protocol 22, at 140x48.
Each demo used a scratch home, `HIVE_NO_STARTUP=1`, `TERM=vt100`, and `COLORTERM=truecolor`.
No test TUI ran on the real home.

Each capture uses `source=visible` and `format=ansi`.
Two equal reads 500 ms apart establish stability.
Each metadata file records `truncated=false` and scroll offset 0.
The comparisons crop the top-left board size and use `cellgrid.Convert` and `cellgrid.Compare`, with no mask.

All **34 live comparisons** gave zero cell differences:

- Nine changed boards, six approved scenarios, and five unchanged Agents references.
- Four overview/one-pod scenarios before and after `h` twice.
- Stopped, exited, and stale manager variants with a live worker and a live general child, at both sizes.

The 148 ANSI captures and plain text captures remain in `verifier-2/live/`.
That folder also contains metadata, comparison output, and 56 terminal PNGs.
`live.py.txt`, `extra-live.py.txt`, `gridtool.go.txt`, and `screenshots.py.txt` record the capture procedures.

### Key results at 120x40 and 80x24

| Scenario | Observed result |
| --- | --- |
| Only the coordinator remains | Up and `k` select its heavy border. Enter opens its details. Esc closes details and keeps selection. |
| Empty overview keys | Left, right, down, and enter are safe. Enter opens no phantom pod or details. |
| Selected coordinator with zero pods | Left and right keep selection. Down removes selection. Enter then does nothing. Up or `k` selects it again. |
| `h` twice with coordinator selected | Pods return, then hide. The coordinator keeps selection. Enter opens details again. |
| Coordinator selected before the last pod hides | Selection survives hiding. Enter opens details. Esc closes them. |
| Hidden manager with a live general child | The child stays in the project pod. Right selects it. Enter opens its details. |
| Selected project pod disappears | Selection moves to home. The next `h` restores the remembered project selection. |
| All project pods hide, with live home cards | Notice, coordinator, home, and title hint remain. Enter opens home. Esc returns. |
| No visible card | Only the notice remains in the body. The additional zero-pod keys remain safe. |
| No registered project | Only the notice remains in the body. Up, `k`, and enter open no coordinator details. Decision 10 stays unchanged. |

### Visual table

The verifier captured each of the 14 referenced or changed boards again with Firefox and fresh scratch profiles.
`boards.log` records the approved task-board captures.
`reference-boards.log` records the home references before the approved g9 title change.
The board PNGs stayed in `/home/penguin/source/hive/.hive/tmp/verify134-second-boards/` until cleanup.
The comparison uses the approved code grids. Browser padding and fonts do not set terminal cells.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and tabs | Coordinator/supervisor words, `●`, tab underline | matches | Same frame as the approved grids |
| Legacy title without the hint | `project agents` | overridden | Decisions 109 and 112 approve the new title row |
| Persistent title hint | `h shows N stopped`, including 80 columns | matches | Decision 109 |
| No hidden cards | No `h shows` title text | matches | First `h` shows every stopped card |
| Hidden-manager pod border and title | Box glyphs, project name, goal words | matches | Decision 107 |
| Overview top children | Complete cards with no rail joint | matches | Approved home-tree layout |
| One-pod top connectors | `┼`, `│`, bus, `┷`, `┴` | matches | Rule A and decision 107 |
| General child of hidden manager | `◐ api-review-1` in `example-api` | matches | Ask `a5` |
| Child of hidden worker | General child in `home` | matches | Existing decision 39 |
| Nested tree and scroll marks | `┬`, `┴`, `┃`, `━`, arrow words | matches | Existing tree grids remain equal |
| Pod stopped counts | `h shows N stopped`, `h hides N stopped` | matches | Manager counts with the other stopped cards |
| Card states | `● working`, `◐ idle`, `○ stopped` | matches | Glyphs and words remain after colour removal |
| Message path | Heavy line, arrow, legend words | matches | Existing Live grids remain equal |
| Selected pod or card | Heavy box border | matches | Approved selection signal |
| All-hidden project state | `no project agents`, coordinator, `home` | matches | Decision 110 |
| Notice-only state | `no project agents` alone in the body | matches | Decisions 10 and 110 |
| Coordinator selected with zero pods | Heavy border after up or `k` | matches | AC2 and decision 65. Rework fixes the first failure |
| Coordinator details with zero pods | Agent title, `esc closes`, `[details]` | matches | Decision 65. Works before and after the `h` cycle |
| Legend and key hints | Plain text key names | matches | Same approved frame and action hints |

## V4: scope and full gate

`git diff dev...HEAD --stat` names only the allowed product, spec, design, baseline, and task evidence files.
The merge of `b34ee38d` adds only documentation already on current `dev`.
The rework changes only `agents_screen.go` and `agents_hide_test.go` outside task evidence and the updated brief.
The existing record-order test keeps its order check. Its fixture adds live managers because empty pods now hide.

The final gate ran from the task worktree after the merge:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Each command exited 0. The first two commands printed nothing.
The linter printed `0 issues.`. Every race test passed.
`verifier-2/gate.log` contains the full output.

## Cleanup

`verifier-2/cleanup.log` records the cleanup and absence checks.

- Each live script closed only its own scratch tab and pane.
- The verifier closed its diagnostic tabs in the task workspace.
- The verifier closed the isolated workspace and stopped its server.
- Server PID 3874404 is absent. Its session directory is absent.
- `.hive/tmp/verify134-second/` is absent, including source copies, binaries, scripts, profiles, exports, and the scratch home.
- `/home/penguin/source/hive/.hive/tmp/verify134-second-boards/` is absent.
- Final evidence, captures, reproduction text, and the first verdict remain.
- Only the verifier agent tab remains in the task workspace.
