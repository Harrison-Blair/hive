# VERIFIED 37-screen-work-tasks

Result: pass

Verifier: `t37-verifier-4`. Date: 2026-10-22.
Reviewed task commit: `628c75bcd5e73fa2feebf6d988bec6ee440daa8c`.

## Scope

The manager limited this run to merge repair after `78ed6ac2`.
Earlier V1 red/green runs of 12 tests, deliberate key breaks, and the live
Compact check carry forward from `verify2/`. This report replaces the
previous failed verdict. New evidence is in `verify4/`.

| Item | Result | Evidence |
| --- | --- | --- |
| Repair review | pass | `repair-review.txt`: merge `9bd77745` imports dev `d3e1aa2c`. The README resolution keeps both Tasks and Chat decisions. The only source change to the automatic merge removes unused `panelStyle` from `shell.go`. Commit `628c75bc` changes evidence only. |
| V1 | pass | Earlier base red/green and key-break checks carry forward. New `golden.log` and `restored-green.log` pass Work, Compact, WorkGoals, and WorkQueue. `token-red.log` fails on four wave cells after the deliberate dim-to-info token change. The source was restored. |
| V2 | pass | New Work live run: `live-diff.log` reports 0 cells differ. `keys.log` confirms selection, dependencies, brief, and VERIFIED file. Earlier Compact result carries forward. |
| V-gate scope | pass | One fetch and merge at the start used dev `d3e1aa2c4b2f0a63cef6e417ca86a40bedbff6e4`. `merge.log` says already up to date. `scope.log` records `git diff origin/dev...HEAD --stat`. Changes stay within the source table and task evidence. No later dev movement is a failure condition. |
| V-gate commands | pass | Full output in `gate.log`. Formatting and tidy print nothing. Vet, lint, and race tests exit 0. Lint reports 0 issues. No known flake occurred. |

## Commands

```sh
git fetch origin dev
git merge --no-edit origin/dev
git diff origin/dev...HEAD --stat
go test ./internal/tui -run 'TestWork(TasksMatches(Work|Compact)Board|(Goals|Queue)MatchesItsBoard)$' -count=1 -v
# Change workTasksListModel: dim := dimStyle to dim := infoStyle.
go test ./internal/tui -run '^TestWorkTasksMatchesWorkBoard$' -count=1
# Restore dimStyle; rerun the four golden checks.
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

The deliberate change failed with the expected token differences on C, B,
A, and —. The restored four-test run passed. No source mutation remains.

## Live Work check

Built `go build -o .hive/tmp/verify4/hive .` from the reviewed tree. Created
scratch tab `wAP:t2`, pane `wAP:p2`, in the verifier workspace. Ran:

```sh
HIVE_HOME="$PWD/.hive/tmp/verify4/home" HIVE_NO_STARTUP=1 COLORTERM=truecolor \
  "$PWD/.hive/tmp/verify4/hive" \
  --demo "$PWD/.hive/spec/tui-v2/design/Work.fixture.json" --size 120x40
herdr pane read wAP:p2 --source visible --format ansi
```

Two reads 500 ms apart were identical (`Work-first.ansi`, `Work.ansi`). The
pane had 57 viewport rows and scroll offset 0 (`Work-pane.json`). The CLI
capture contains the complete screen; it exposes no truncation flag.
The comparer crops to 120×40 and applies only the approved space-cell rule.
It uses no mask. Its source is `livecmp.go.txt`.

Pressed `j d`: task 2 selected; dependency graph shows NEEDS and BLOCKS.
Pressed `esc k b`: task 3 brief shows the restart title from the fixture.
Pressed `esc v`: VERIFIED shows task 3 and `Result: pass`. Pressed `esc` to
close the file. Captures and extracted observations are in `verify4/`.

## Visual review

Captured the approved board again from
`/home/penguin/source/hive/.hive/spec/tui-v2/design/Work.dc.html` with Firefox,
a new profile, and `--window-size=1220,780`. Converted the new ANSI capture
to an HTML pre and captured it with a separate new profile. Images:
`Work-board.png`, `Work-live.png`. The screenshot font, margin, and cell
pitch differ between the HTML pre and board renderer. The grid comparison
shows no content, token, or weight difference outside the space-cell rule.
The plain capture with colour removed is `Work-plain.txt`.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | coordinator working, supervisor live, needs you 0, 14:22 | matches | Fixture clock and records |
| Tab bar | 3 Work with underline | matches | Active tab has text and underline |
| View row | [Tasks], Goals, Queue, switch view | matches | Selected view has brackets |
| List panel | tasks · 4 | matches | Full border and title |
| Selected row | ▶ 3-tui-restart | matches | Selection has a glyph |
| Other IDs | long task ID ends with … | matches | Column truncation |
| Status column | landed, assigned | matches | Words remain without colour |
| Wave column | C, B, A, — | matches | Values remain without colour |
| Details panel | 3-tui-restart | matches | Title and full border |
| Detail fields | title, status, goal, worker, verifier, branch, merge, depends | matches | Labelled values; stopped and pass are words |
| Acceptance | ACCEPTANCE and six ✓ marks | matches | Pass has a glyph |
| Evidence actions | b brief.md, v VERIFIED.md, d deps graph | matches | Keys and labels |
| Evidence rows | V1, V3, pass | matches | Result is a word |
| Footer | [list] and key hints | matches | Focus and actions have words |

The earlier Compact visual review and 0-cell live result remain applicable
under the manager's merge-only scope.

## Cleanup

Closed only the scratch tab `wAP:t2` (`cleanup.json`). The workspace then
contained only verifier tab `wAP:t1`. Removed `.hive/tmp/verify4`, including
the scratch home, binary, HTML, and browser profiles; checked that the folder
is absent. Preserved the render and comparison scripts as evidence text.
No source file differs from HEAD. Pre-existing scratch folders were left
unchanged.
