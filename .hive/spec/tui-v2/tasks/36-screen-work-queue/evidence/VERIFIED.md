# Independent verification: 36-screen-work-queue

Verifier: `t36-verifier-1`. Result: **pass**.

Task code: `e40c9588`. Final gate: `333cae55`, after the merge of dev
`ce0c20b5`. The additional dev commit changes a friction document only.
All paths below are relative to this evidence folder.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests, AC1–AC4 | pass | `verifier/base-red.txt`, `green.txt`, `token-red.txt`, `key-red.txt`, `restored-green.txt` |
| V2 Live check, AC1, AC2, AC4 | pass | `verifier/live-diff.txt`: 0 cells differ; fresh ANSI, metadata, and PNG files in `verifier/` |
| V-gate Diff scope and gate | pass | `verifier/scope.txt`, `verifier/gate.txt`: five commands exit 0 |

## V1: independent test runs

Copied the task test files `work_queue_test.go`, `selection_marker_test.go`,
and `shell_test.go` into a detached scratch worktree at dev `866bde2a`.
Ran this command there and in the task worktree:

```sh
go test -count=1 -v ./internal/tui/ -run 'TestWorkQueue|TestQueue|TestScreenSelectionsUseHoneyMarkerAndDefaultForeground|TestScreenKeysSelectAllSixScreens'
```

Each new top-level test fails on the base and passes on the task code.
The Queue selection and shell test changes also fail on the base.
The inherited 79×24 too-small subtest passes on the base, as expected.
The golden test fails on cell differences, not a build error.
Normal-mode tests use scratch record homes. They cover loaders, refresh,
selection, each project, grant groups, stops, landed tasks, and empty data.
Responsive tests cover 79×24, 80×24, 80×25, 119×40, 119×41, 120×40,
and 120×41, with long text, scrolling, whole borders, and visible selection.
Focus tests cover the compact overlay and its task action.

In a second detached scratch worktree at the task HEAD:

1. Changed `queueBorderStyle` from `component.Border` to `component.Dim`.
   `go test -count=1 ./internal/tui/ -run TestWorkQueueMatchesItsBoard`
   fails on four tree glyph cells.
2. Restored the token. Changed the stops-only key from `s` to `x`.
   `go test -count=1 ./internal/tui/ -run 'TestQueueStopsOnly|TestQueueEmptyData'`
   fails both tests.
3. Restored the key. `go test -count=1 ./internal/tui/ -run 'TestWorkQueue|TestQueue'`
   passes.

No component fix is in the diff. No source mutation remains.

## V2: fresh live run

Built the task binary with `go build -o .hive/tmp/t36-verify/hive .`.
Created a scratch home with `.hive/README.md` only. Created tab `w9D:t2`,
pane `w9D:p2`, with `HIVE_HOME` set to that scratch home and
`COLORTERM=truecolor`. The pane had 57 rows.
Ran:

```sh
COLORTERM=truecolor .hive/tmp/t36-verify/hive --demo .hive/spec/tui-v2/design/WorkQueue.fixture.json --size 120x40
```

For each capture, two `pane.read` calls 500 ms apart returned equal text.
Used `source: visible`, `format: ansi`. Each read has `truncated: false`.
Each `pane.get` gives scroll offset 0. The `*-metadata.json` files retain
these responses. CLI visible ANSI reads also agreed before the key run.

A scratch Go script used `cellgrid.Convert` on the top-left 120×40 cells,
then `cellgrid.Compare` against `WorkQueue.grid`. It ignored token and bold
only when both cells were spaces, as the Manager decisions section permits.
Result: **0 cells differ**.

Captured the approved board again with Firefox, a new scratch profile,
`--headless --no-remote --window-size=1220,780 --screenshot`, and
`file:///home/penguin/source/hive/.hive/spec/tui-v2/design/WorkQueue.dc.html`.
Converted the fresh ANSI captures to HTML preformatted text and captured
those with the same Firefox options and a separate scratch profile.
Compared `verifier/board.png` with `verifier/start.png`. Their terminal cells
agree. The HTML wrappers have different outer padding and baseline positions;
these raster offsets are outside the cell comparison.

| Keys, in sequence | Capture | Observation |
| --- | --- | --- |
| start | `start.ansi`, `start.png` | CP5 task selected; list footer agrees with the board |
| `k` | `k.ansi`, `k.png` | selects 4-agents-equal-columns; details show landed |
| `s` | `s.ansi`, `s.png` | only stopped CP5 task; action says show all tasks |
| `s` | `s2.ansi`, `s2.png` | all four tasks return |
| `enter` | `enter.ansi`, `enter.png` | details border has focus; footer says details |
| `esc` | `esc.ansi`, `esc.png` | list focus returns |
| `t` | `t.ansi`, `t.png` | Tasks opens with CP5 selected |

All capture paths in this table are in `verifier/`.

### Board review and state without colour

Stripped ANSI SGR codes and checked the text signals in each capture.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and service state | coordinator working, supervisor live, ● | matches | Board header |
| Tab bar | 3 Work, ━ | matches | Active tab has a word and underline |
| Sub-tabs | [Queue] | matches | Brackets identify the view |
| List title | queue · hive · 4 tasks · 1 stop | matches | Counts are words and numbers |
| Grant headings | GRANT, NO GRANT | matches | Plain-text group labels |
| Dependency tree | └▶ | matches | Tree glyphs show dependencies |
| Task states | landed, assigned | matches | Each state has a word |
| Selection | ▶ 1-cp5-end-to-end | matches | Selection has a glyph |
| Stop row | ! stop: worker unresolved | matches | Stop has a word and reason |
| Details title and fields | task ID; title, status, goal, grant, worker, depends | matches | Board fields and values |
| Stop explanation | STOP, The worker record is live, but its pane is gone. | matches | Board explanation |
| Next step | NEXT STEP, Ask the coordinator in 2 Chat. | matches | Board instruction |
| Action line | t open task in Tasks, s show stops only | matches | Board actions |
| Focus and footer | [list], [details], key words | matches | Focus remains clear without colour |
| Panel borders | ╭ ╮ ╰ ╯ │ | matches | Whole borders; cell comparison passes |

## V-gate

`git merge dev` completed. The final diff against dev contains only the
brief source-table files and task evidence. The README edit is confined to
Manager decisions. `verifier/scope.txt` holds the exact diff stat.

Ran the five AGENTS.md gate commands again after the final merge:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Each command exits 0. The first two print nothing. Lint reports 0 issues.
The race suite passes, including `internal/tui`. No health flake occurred.
`verifier/gate.txt` holds the full output.

## Cleanup

Quit the scratch TUI with two Ctrl+C keys. Checked that its foreground
process was the shell, then closed only tab `w9D:t2`. The tab list contains
only the verifier tab. Removed both detached scratch worktrees and the
literal folder `.hive/tmp/t36-verify`. Checked that the folder is absent.
`verifier/cleanup.txt` records the result. Final evidence is preserved.

The user sign-off in decision 47 remains a later step.
