# T08-tui-log-view ESCAPE-FOLLOWUP

| Item | Value |
| --- | --- |
| Task | T08-tui-log-view |
| Pass | `t08-escape-fix-01` |
| Date | 2026-10-18 |
| Agent | worker, Pi, `opencode-go/deepseek-v4-flash:high` (requested :medium) |
| Branch | `task/T08-escape-clear` |
| Entry source | `b4ab63ca1943c2c0491ae193392d8f7bdf3ea857` |
| Entry tree | `1c09a38e555eb3e6595ea9cc7566e0ca56c9c61a` |
| Entry tree entries | 315 |
| Result | Narrow fix implemented. Independent check is pending. |

## Failure basis

The CP4 run report records one failure. The report is
`.hive/tmp/coordinator/reports/cp4-run-01.md`.

The actual home run used source `b4ab63ca` and the binary with SHA256
`1aa9afa6f010770b95cb3cf39a8ed5034eabc826cc1e88a81d4477d43b913251`.
The key sequence was `/`, `project`, Enter, `l`, `f`, `g`, `G`, Escape, `q`.

The `G` capture and the Escape capture are byte-identical. Their diff is 0
bytes. The footer stays `level: info  text: project  follow: on`.
The TUI child exit status is 0.

The real `project list` event, the header `supervisor: down`, `agents: 0`, and
`inbox: 0 unread` passed. CP4 stays open. This evidence does not accept CP4.

I reproduced the failure at the model before the source edit. The new test
failed on the old source. The old source left the text filter and the input in
place after a viewing Escape.

## Scope

| Path | Change |
| --- | --- |
| `internal/tui/logview.go` | Add the viewing Escape case and one shared clear helper. |
| `internal/tui/escape_clear_test.go` | New model regression test for the CP4 key sequence. |
| `spec/hive-v1/evidence/T08-tui-log-view/ESCAPE-FOLLOWUP.md` | This record. |

No other Go file, old test, old evidence, task file, tracker row, instruction
file, dependency file, or STE baseline changed.

## Behavior

Escape in viewing mode now clears the applied text filter and the input.
The mode stays viewing. The level filter and the follow state stay as they
were. The view rebuilds with the unfiltered text.

Escape in editing mode keeps the old behavior. It clears the input and the
applied filter, returns to viewing mode, and logs `edit cancel`.

The viewing Escape writes one action log event through the existing checked
log path. The action name is `filter clear`.

All other viewing and editing keys keep their behavior. This covers `/`, Enter,
`q`, Ctrl-C, `l`, `f`, `g`, `G`, the follower, the body output, EOF, TTY
cancellation, and the log error path.

## Test-first record

The test drives a real `Model` through `Update` and `View`. It uses a temporary
home, the daily log file, and the fixed clock. The key sequence is `/`,
`project`, Enter, `l`, `f`, `g`, `G`, Escape.

The test checks the literal footer `level: info  text: -  follow: on`.
It also checks that the data filter is empty, the input is empty, the mode is
viewing, the level is `info`, and follow is on. It checks that the view shows
non-project lines and hides the debug line. It checks one `filter clear` log
event.

Old source file hash:
`f03da3d2c724c1b9107a9818fa5aa5042f48970715c0275b1910506dd179c493`.
Final test file hash:
`c6eeb82a2b9c1209122feed5e6b52201e55f26d3b57a45025ea9495cf62e0d57`.

Red command:

```sh
go test ./internal/tui/ -run TestViewingEscapeClearsAppliedTextFilter -v -count=1
```

Red output on the old source:

```text
=== RUN   TestViewingEscapeClearsAppliedTextFilter
    escape_clear_test.go:48: log lines went from 11 to 11 after viewing Escape; want one event
--- FAIL: TestViewingEscapeClearsAppliedTextFilter (0.00s)
FAIL
exit 1
```

The first test revision also failed on the footer before I added the log check.
That red output shows `text: project` after Escape.

The raw red is in
`.hive/tmp/coordinator/reports/t08-escape-fix-01-captures/red-final-go-test.txt`.

## Green and mutation

Fixed source hash:
`6c4571045f30763f8c87be13aaff8dfeb56b137cea922353a103b1947591d362`.
The focused test passed with the fixed source.

The focused old tests also passed. This set covers editing Escape, editing
`g` and `q` as ordinary text, `q` and Ctrl-C as quit, follower order, follow
state, action log refusal, and finite EOF.

I removed the viewing Escape case in the owned copy. The regression test failed
on the missing log event. I restored the file. Its hash equals the fixed hash.

## Gate

The full gate ran once at the final Go source:

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
```

`gofmt` printed nothing. `vet` exited 0. `lint` reported `0 issues.`.
The race suite passed all packages. Raw outputs are in the pass captures.

## Fixture TUI

I built one absolute binary from the fixed worktree.
Binary SHA256:
`8bfd16cfbb6316fd0ada0e1e2346d49079fb25f2bb312ac0798e94e51205cb4e`.
Binary size is `10018666` bytes. Build revision is `b4ab63ca` with
`vcs.modified=true`.

The fixture home holds the marker `.hive/README.md`. It holds a synthetic daily
log file. The synthetic `project list` event is safe and local. The run used a
new own nonmodel tab and PTY in workspace `wBA`.

The raw keys were `/`, `project`, Enter, `l`, `f`, `g`, `G`, Escape, `q`.
I captured the screen after each key.

| Key | Footer after the key |
| --- | --- |
| `/` | `filter:` |
| `project` | `filter: project` |
| Enter | `level: debug  text: project  follow: on` |
| `l` | `level: info  text: project  follow: on` |
| `f` | `level: info  text: project  follow: off` |
| `g` | `level: info  text: project  follow: off` |
| `G` | `level: info  text: project  follow: on` |
| Escape | `level: info  text: -  follow: on` |

After Escape the view shows `project list`, `project alpha`, `other beta`,
`project gamma`, and `delta`. The debug line stays hidden by the `info` level.
The diff from the `G` capture shows the two unfiltered lines and the new footer.
The fixture log holds one `filter clear` event. The child exit status is 0.
The wrapper shell returned to the fixture home.

I closed only the fixture tab `wBA:t2` after its identity check. Parsed lists
show tab `wBA:t2` and pane `wBA:p2` absent. Workspace `wBA` and pane `wBA:p1`
stay. The own source workspace stays open.

## Limits

The independent verifier owns the new `ESCAPE-VERIFIED.md` and source
acceptance. This pass does not write that file.

The actual Root home CP4 replay is pending. It comes after a verified Root
landing. This pass does not claim a CP4 pass.

The local-user and cooperating-writer boundary applies. Hostile replacement
between `Lstat` and `Open` stays deferred for T01. Malicious peers beyond T05
stay outside this pass.

END T08 ESCAPE FOLLOWUP
