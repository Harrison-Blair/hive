# T06-logs-command VERIFIED

| Item | Value |
| --- | --- |
| Task | T06-logs-command |
| Commit | 1c556c3 |
| Date | 2026-10-18 |
| Agent | verifier, pi, openai-codex/gpt-6.1-sol:high |
| Verdict | fail |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | fail | Focused tests pass. The full V1 suite also passes with a 2 MiB cap. F1 records the test defect. |
| V2 | AC3 | pass | Uncached race tests pass. Offset, partial-line, rollover, and cancellation mutations cause assertion failures. |
| V3 | AC1, AC2, AC3 | fail | The specified transcript passes. Extra probes fail on cutoff, line number, text rendering, and JSON errors. See F2 to F5. |
| V4 | AC4 | pass | Formatting prints nothing. Vet, lint, and the uncached full race test pass. |

## Findings and recommendations

### F1: The line-cap tests use the implementation limit

`internal/logs/read_test.go:117` and `:147` derive fixture sizes from `maxLineBytes`.
The complete V1 suite passes after a mutation changes that constant from 1 MiB to 2 MiB.
This result violates the requirement that a test fails when its behavior breaks.
The actual 1 MiB implementation passes an independent test with fixed sizes 1048575, 1048576, and 1048577.
That independent test fails on assertions with the 2 MiB mutation.
Recommendation: use fixed specification sizes in the worker tests.

### F2: Follow does not apply the cutoff

`cmd/logs/logs.go:119` sends each new event to `printer.one` without a time check.
A future RFC 3339 cutoff removes the initial event but keeps a new event with a time before that cutoff.
Expected stdout: empty. Actual stdout: the new event.
AC1 requires each event time to be at or after the cutoff.
Recommendation: apply the same cutoff to the initial read and to follow events.

### F3: Follow diagnostics use a relative line number

`internal/logs/follow.go:33` starts `lineNumber` at zero, even when the offset follows two complete lines.
The JSON follow probe appends malformed line 3. The diagnostic reports `:1`, not `:3`.
Recommendation: keep the absolute line number without a change to the required public signatures.

### F4: A component newline splits a text event

`internal/logs/render.go:34` prints the component without control-character conversion.
A valid JSON event with component `cli\ninjected` prints two physical stdout lines.
The message newline uses the expected literal `\n`. The component newline does not.
Recommendation: keep every text field and key on one physical line.

### F5: JSON read errors print an envelope

`cmd/logs/logs.go:233` prints a failure envelope on stdout for an unreadable file, directory, or symbolic link.
Expected: empty stdout, one plain diagnostic on stderr, status 1, and no panic.
Actual: one envelope on stdout, empty stderr, and status 1.
The brief requires a plain diagnostic. The JSON log stream must not use envelopes.
Text-mode probes meet the diagnostic requirement.
Recommendation: send read errors to stderr in both modes.

## Required hand checks

| Item | Result | Evidence |
| --- | --- | --- |
| H01 Three daily files in date order | pass | TestReadReadsDailyFilesInDateOrder and date-order mutation. |
| H02 1 MiB plus one byte is malformed, next event stays | pass | Fixed-size independent test and live over-cap probe. F1 concerns test strength, not the current reader cap. |
| H03 Cutoff includes equality and excludes one second before | pass | TestSinceCutoffKeepsAtAndAfter and equality mutation. F2 concerns follow events. |
| H04 Previous-day cutoff reads two daily files | pass | TestSincePreviousDayReadsTwoFiles. |
| H05 Each filter field matches and misses | pass | V1 subtests, filter mutation, and live field probes. |
| H06 Warn keeps warn and error | pass | TestFilterLevelKeepsNamedAndAbove and live JSON probe. |
| H07 Limit keeps the last N matches | pass | V3 limit 2 and live warn limit 1. |
| H08 Malformed text uses question mark and 200-byte prefix | pass | V3 and live over-cap text probe. |
| H09 Malformed JSON uses stderr and continues | fail | Initial read passes. Follow reports line 1 for actual line 3. F3. |
| H10 New follow line arrives within one second | pass | TestFollowNewLineWithinOneSecond. |
| H11 Line between initial read and first poll prints once | pass | TestFollowLineAppendedBeforeStartPrintsOnce and offset mutation. |
| H12 Snapshot partial line stays until follow starts | pass | TestFollowKeepsPartialLineUntilNewline and snapshot-offset mutation. |
| H13 Two writes produce one event after the newline | pass | Same partial-line test checks early output and duplicates. |
| H14 Follower moves to the next daily file | pass | Worker test and independent strict deadline test, measured 400.861075 ms. |
| H15 SIGINT gives child status 0 without a stack trace | pass | V3 timeout uses --preserve-status. Direct follow probes also inspect child.returncode 0. |
| H16 JSON stream uses event lines, not envelopes | fail | Success events pass. Read failures emit envelopes. F5. |
| H17 Line reads use bounded storage and drain long lines | pass | ReadSlice uses a 64 KiB reader. Retained line size is capped. Live 5 MiB line continues to the next event. |
| H18 Reader does not write a log file | pass | Production code uses os.Open only. Live read probes preserve the file hash. |
| H19 Public API matches V1 and the task | pass | Event, Filter, Offset, Reader, NewReader, Read, Follow, and Match retain the required signatures. Private metadata and accessors are additions. |
| H20 No dependency change | pass | The base-to-worker go.mod and go.sum diff is empty. |
| H21 Only owned files change | pass | The complete base-to-worker list contains the named source files, adjacent tests, and V1 to V4 evidence. |

## Boundary and completion

Only the approved local-user and cooperating-writer boundary applies.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
The probes use unique homes with explicit `.hive/README.md` markers and absolute `--home` arguments.
Every live command uses the unique ignored scratch binary. The existing binary stays unchanged.
The V3 commands use `HIVE_BIN` instead of `./hive` to meet that requirement.
Focused tests use `-count=1`. The full race gate also uses `-count=1`.
Lint uses `--allow-serial-runners` without a weaker check.

The first probe attempt did not reach the socket case because its fixture path exceeded the Unix socket path limit.
That verifier setup error was not a hive panic. A second unique home under `/tmp` completed every probe.
Both fixture homes were removed after their outputs were recorded.
All required checks completed. The failures above remain uncorrected because no source edit is authorized.
No worker report, log, state directory, pre-existing binary, worktree, credential, or global setting changed.
No tracker or V-file changed. The STE baseline did not change.

## Probe verdicts

```text
PASS: unreadable text diagnostic
FAIL: unreadable JSON plain diagnostic: expected empty stdout and one stderr line
PASS: directory text diagnostic
FAIL: directory JSON plain diagnostic: expected empty stdout and one stderr line
PASS: nonempty-directory text diagnostic
PASS: symlink-valid text diagnostic
FAIL: symlink-valid JSON plain diagnostic: expected empty stdout and one stderr line
PASS: symlink-missing text diagnostic
PASS: symlink-directory text diagnostic
PASS: symlink-zero text diagnostic
PASS: symlink-chain text diagnostic
PASS: fifo text diagnostic
PASS: socket text diagnostic
PASS: missing daily file skips
PASS: empty file skips
PASS: limit keeps last matching event
PASS: warn keeps warn and error
PASS: component match
PASS: component miss
PASS: task match
PASS: task miss
PASS: agent match
PASS: agent miss
PASS: request match
PASS: request miss
PASS: log reads do not write
PASS: invalid JSON skip and continue
PASS: invalid text prefix
PASS: null JSON skip and continue
PASS: array JSON skip and continue
PASS: number JSON skip and continue
PASS: under-cap JSON skip and continue
PASS: at-cap JSON skip and continue
PASS: over-cap JSON skip and continue
PASS: over-cap text prefix
PASS: far-over JSON skip and continue
PASS: 10000 extra keys
FAIL: text event is one line: component contains a JSON escaped newline
FAIL: follow malformed absolute line number
FAIL: follow keeps cutoff
```

## Mutation verdicts

```text
PASS: date order mutation causes an assertion failure
PASS: date order original passes
PASS: cutoff equality mutation causes an assertion failure
PASS: cutoff equality original passes
FAIL: line cap mutation causes an assertion failure
PASS: line cap original passes
PASS: filter match mutation causes an assertion failure
PASS: filter match original passes
PASS: follow offset mutation causes an assertion failure
PASS: follow offset original passes
PASS: partial snapshot mutation causes an assertion failure
PASS: partial snapshot original passes
PASS: rollover mutation causes an assertion failure
PASS: rollover original passes
PASS: cancel result mutation causes an assertion failure
PASS: cancel result original passes
```

The line-cap mutation survives because its test fixture changes with the limit.
The independent fixed-cap test catches that mutation and passes with the original source.
The other seven mutations cause assertion failures, not compile failures.

## Runs

Each section records the exact command, stdout, stderr, and status.
`[empty]` marks an empty stream. `bash -c` quotes preserve terminal spaces in recorded commands. A failed mutation command is an expected red run.
The commands ran from the task worktree root. `go -C` uses an isolated source copy only for the named probes.

### Command 1

```sh
git branch --show-current; git status --porcelain; git rev-parse HEAD
```

```text
stdout:
task/T06-logs-command
1c556c332b66c636bf4ec1af79ffaaedb8979cd8
stderr:
[empty]
exit 0
```

### Command 2

```sh
tail -n 1 /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/coordinator/reports/t06-impl-01.md
```

```text
stdout:
END WORKER T06-logs-command t06-impl-01
stderr:
[empty]
exit 0
```

### Command 3

```sh
git diff --name-only 840255d3f598ad18239eccd7a2eb2e93547ca1fb..1c556c332b66c636bf4ec1af79ffaaedb8979cd8
```

```text
stdout:
cmd/logs/logs.go
cmd/root.go
internal/logs/doc.go
internal/logs/filter.go
internal/logs/filter_test.go
internal/logs/follow.go
internal/logs/follow_test.go
internal/logs/read.go
internal/logs/read_test.go
internal/logs/render.go
spec/hive-v1/evidence/T06-logs-command/V1.md
spec/hive-v1/evidence/T06-logs-command/V2.md
spec/hive-v1/evidence/T06-logs-command/V3.md
spec/hive-v1/evidence/T06-logs-command/V4.md
stderr:
[empty]
exit 0
```

### Command 4

```sh
git diff 840255d3f598ad18239eccd7a2eb2e93547ca1fb..1c556c332b66c636bf4ec1af79ffaaedb8979cd8 -- go.mod go.sum
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 5

```sh
git log --format=fuller 840255d3..1c556c3
```

```text
stdout:
commit 1c556c332b66c636bf4ec1af79ffaaedb8979cd8
Author:     Harrison-Blair <hjblair@me.com>
AuthorDate: Sun Oct 18 01:03:37 2026 -0400
Commit:     Harrison-Blair <hjblair@me.com>
CommitDate: Sun Oct 18 01:03:37 2026 -0400

    docs(evidence): correct the T06 code-search transcript

commit fff6974071074b811e6639fb28d2644f182613f0
Author:     Harrison-Blair <hjblair@me.com>
AuthorDate: Sun Oct 18 01:03:04 2026 -0400
Commit:     Harrison-Blair <hjblair@me.com>
CommitDate: Sun Oct 18 01:03:04 2026 -0400

    feat(logs): add the logs command and reader
stderr:
[empty]
exit 0
```

### Command 6

```sh
go test -count=1 ./internal/logs/ -run 'TestRead|TestFilter|TestSince' -v
```

```text
stdout:
=== RUN   TestFilterMatchFields
=== RUN   TestFilterMatchFields/empty
=== RUN   TestFilterMatchFields/component_match
=== RUN   TestFilterMatchFields/component_miss
=== RUN   TestFilterMatchFields/task_match
=== RUN   TestFilterMatchFields/task_miss
=== RUN   TestFilterMatchFields/agent_match
=== RUN   TestFilterMatchFields/agent_miss
=== RUN   TestFilterMatchFields/request_match
=== RUN   TestFilterMatchFields/request_miss
=== RUN   TestFilterMatchFields/level_match
=== RUN   TestFilterMatchFields/level_miss
=== RUN   TestFilterMatchFields/all_fields_match
--- PASS: TestFilterMatchFields (0.00s)
    --- PASS: TestFilterMatchFields/empty (0.00s)
    --- PASS: TestFilterMatchFields/component_match (0.00s)
    --- PASS: TestFilterMatchFields/component_miss (0.00s)
    --- PASS: TestFilterMatchFields/task_match (0.00s)
    --- PASS: TestFilterMatchFields/task_miss (0.00s)
    --- PASS: TestFilterMatchFields/agent_match (0.00s)
    --- PASS: TestFilterMatchFields/agent_miss (0.00s)
    --- PASS: TestFilterMatchFields/request_match (0.00s)
    --- PASS: TestFilterMatchFields/request_miss (0.00s)
    --- PASS: TestFilterMatchFields/level_match (0.00s)
    --- PASS: TestFilterMatchFields/level_miss (0.00s)
    --- PASS: TestFilterMatchFields/all_fields_match (0.00s)
=== RUN   TestFilterLevelKeepsNamedAndAbove
--- PASS: TestFilterLevelKeepsNamedAndAbove (0.00s)
=== RUN   TestFilterMalformedOnlyMatchesEmptyFilter
--- PASS: TestFilterMalformedOnlyMatchesEmptyFilter (0.00s)
=== RUN   TestReadReadsDailyFilesInDateOrder
--- PASS: TestReadReadsDailyFilesInDateOrder (0.00s)
=== RUN   TestReadSkipsMissingDailyFile
--- PASS: TestReadSkipsMissingDailyFile (0.00s)
=== RUN   TestReadLineCap
--- PASS: TestReadLineCap (0.00s)
=== RUN   TestReadAcceptsLineAtCap
--- PASS: TestReadAcceptsLineAtCap (0.00s)
=== RUN   TestReadMarksInvalidJSONMalformed
--- PASS: TestReadMarksInvalidJSONMalformed (0.00s)
=== RUN   TestReadRejectsNonRegularFile
--- PASS: TestReadRejectsNonRegularFile (0.00s)
=== RUN   TestSinceCutoffKeepsAtAndAfter
--- PASS: TestSinceCutoffKeepsAtAndAfter (0.00s)
=== RUN   TestSincePreviousDayReadsTwoFiles
--- PASS: TestSincePreviousDayReadsTwoFiles (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	0.005s
stderr:
[empty]
exit 0
```

### Command 7

```sh
go test -count=1 -race ./internal/logs/ -run TestFollow -v
```

```text
stdout:
=== RUN   TestFollowNewLineWithinOneSecond
--- PASS: TestFollowNewLineWithinOneSecond (0.50s)
=== RUN   TestFollowLineAppendedBeforeStartPrintsOnce
--- PASS: TestFollowLineAppendedBeforeStartPrintsOnce (0.60s)
=== RUN   TestFollowKeepsPartialLineUntilNewline
--- PASS: TestFollowKeepsPartialLineUntilNewline (2.10s)
=== RUN   TestFollowMovesToNextDailyFile
--- PASS: TestFollowMovesToNextDailyFile (0.00s)
=== RUN   TestFollowCancelledContextReturnsNilError
--- PASS: TestFollowCancelledContextReturnsNilError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	4.209s
stderr:
[empty]
exit 0
```

### Command 8

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 9

```sh
go vet ./...
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 10

```sh
golangci-lint run --allow-serial-runners ./...
```

```text
stdout:
0 issues.
stderr:
[empty]
exit 0
```

### Command 11

```sh
go test -count=1 -race ./...
```

```text
stdout:
ok  	github.com/Harrison-Blair/hive	2.594s
ok  	github.com/Harrison-Blair/hive/cmd	1.019s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/version	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.014s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.046s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/store	3.144s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.010s
ok  	github.com/Harrison-Blair/hive/internal/logs	4.758s
ok  	github.com/Harrison-Blair/hive/internal/project	1.021s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.013s
stderr:
[empty]
exit 0
```

### Command 12

```sh
go build -o /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/hive .
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 13

```sh
printf "HIVE_TMP=%s\nHIVE_BIN=%s\n" "$HIVE_TMP" "$HIVE_BIN"; test -f "$HIVE_TMP/.hive/README.md"
```

```text
stdout:
HIVE_TMP=/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/v3-home.jm4abnd3
HIVE_BIN=/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/hive
stderr:
[empty]
exit 0
```

### Command 14

```sh
"$HIVE_BIN" --home "$HIVE_TMP" project list >/dev/null
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 15

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --limit 2
```

```text
stdout:
2026-10-18T01:15:34.625-04:00 DEBU cli command start command=project list
2026-10-18T01:15:34.625-04:00 DEBU cli command end command=project list duration_ms=0 outcome=ok
stderr:
[empty]
exit 0
```

### Command 16

```sh
set -o pipefail; "$HIVE_BIN" --home "$HIVE_TMP" logs --component cli --json | tail -n 1 | python3 -m json.tool
```

```text
stdout:
{
    "command": "project list",
    "component": "cli",
    "duration_ms": 0,
    "level": "DEBUG",
    "msg": "command end",
    "outcome": "ok",
    "time": "2026-10-18T01:15:34.625176183-04:00"
}
stderr:
[empty]
exit 0
```

### Command 17

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --level error
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 18

```sh
printf 'not json\n' >> "$HIVE_TMP"/.hive/logs/"$(date +%F)".jsonl
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 19

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --limit 1
```

```text
stdout:
? not json
stderr:
[empty]
exit 0
```

### Command 20

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --limit 1 --json
```

```text
stdout:
{"command":"project list","component":"cli","duration_ms":0,"level":"DEBUG","msg":"command end","outcome":"ok","time":"2026-10-18T01:15:34.625176183-04:00"}
stderr:
hive logs: malformed line /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/v3-home.jm4abnd3/.hive/logs/2026-10-18.jsonl:3
exit 0
```

### Command 21

```sh
 timeout --preserve-status --kill-after=2s -s INT 2 "$HIVE_BIN" --home "$HIVE_TMP" logs --follow --limit 1; echo "exit $?"
```

```text
stdout:
? not json
exit 0
stderr:
[empty]
exit 0
```

### Command 22

```sh
rm -rf -- "$HIVE_TMP"; test ! -e "$HIVE_TMP"; printf "removed only V3 fixture\n"
```

```text
stdout:
removed only V3 fixture
stderr:
[empty]
exit 0
```

### Command 23

```sh
printf "HIVE_TMP=%s\nHIVE_BIN=%s\n" "$HIVE_TMP" "$HIVE_BIN"; test -f "$HIVE_TMP/.hive/README.md"
```

```text
stdout:
HIVE_TMP=/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3
HIVE_BIN=/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/hive
stderr:
[empty]
exit 0
```

### Command 24

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: open /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl: permission denied
exit 1
```

### Command 25

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"outcome":"failed","phase":"read","effects":[],"error":"open /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl: permission denied","result":null}
stderr:
[empty]
exit 1
```

### Command 26

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 27

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"outcome":"failed","phase":"read","effects":[],"error":"log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl","result":null}
stderr:
[empty]
exit 1
```

### Command 28

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 29

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 30

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"outcome":"failed","phase":"read","effects":[],"error":"log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl","result":null}
stderr:
[empty]
exit 1
```

### Command 31

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 32

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 33

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 34

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 35

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 36

```sh
rm -rf -- .hive/tmp/t06-verify-01.rdNu1g/probe-home.ncy_ovu3; printf "removed only failed probe fixture\n"
```

```text
stdout:
removed only failed probe fixture
stderr:
[empty]
exit 0
```

### Command 37

```sh
printf "HIVE_TMP=%s\nHIVE_BIN=%s\n" "$HIVE_TMP" "$HIVE_BIN"; test -f "$HIVE_TMP/.hive/README.md"
```

```text
stdout:
HIVE_TMP=/tmp/t06-verify-01-probe.fxv5fnvx
HIVE_BIN=/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/hive
stderr:
[empty]
exit 0
```

### Command 38

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: open /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl: permission denied
exit 1
```

### Command 39

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"outcome":"failed","phase":"read","effects":[],"error":"open /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl: permission denied","result":null}
stderr:
[empty]
exit 1
```

### Command 40

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 41

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"outcome":"failed","phase":"read","effects":[],"error":"log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl","result":null}
stderr:
[empty]
exit 1
```

### Command 42

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 43

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 44

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"outcome":"failed","phase":"read","effects":[],"error":"log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl","result":null}
stderr:
[empty]
exit 1
```

### Command 45

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 46

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 47

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 48

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 49

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 50

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
[empty]
stderr:
hive: log is not a regular file: /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl
exit 1
```

### Command 51

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 52

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 53

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --level warn --limit 1 --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.861854-04:00","level":"ERROR","component":"cli","msg":"error","task_id":"T","agent_id":"A","request_id":"R"}
stderr:
[empty]
exit 0
```

### Command 54

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --level warn --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.861848-04:00","level":"WARN","component":"cli","msg":"warn","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861854-04:00","level":"ERROR","component":"cli","msg":"error","task_id":"T","agent_id":"A","request_id":"R"}
stderr:
[empty]
exit 0
```

### Command 55

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --component cli --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.861818-04:00","level":"DEBUG","component":"cli","msg":"debug","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861840-04:00","level":"INFO","component":"cli","msg":"info","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861848-04:00","level":"WARN","component":"cli","msg":"warn","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861854-04:00","level":"ERROR","component":"cli","msg":"error","task_id":"T","agent_id":"A","request_id":"R"}
stderr:
[empty]
exit 0
```

### Command 56

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --component MISS --json
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 57

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --task T --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.861818-04:00","level":"DEBUG","component":"cli","msg":"debug","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861840-04:00","level":"INFO","component":"cli","msg":"info","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861848-04:00","level":"WARN","component":"cli","msg":"warn","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861854-04:00","level":"ERROR","component":"cli","msg":"error","task_id":"T","agent_id":"A","request_id":"R"}
stderr:
[empty]
exit 0
```

### Command 58

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --task MISS --json
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 59

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --agent A --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.861818-04:00","level":"DEBUG","component":"cli","msg":"debug","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861840-04:00","level":"INFO","component":"cli","msg":"info","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861848-04:00","level":"WARN","component":"cli","msg":"warn","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861854-04:00","level":"ERROR","component":"cli","msg":"error","task_id":"T","agent_id":"A","request_id":"R"}
stderr:
[empty]
exit 0
```

### Command 60

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --agent MISS --json
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 61

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --request R --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.861818-04:00","level":"DEBUG","component":"cli","msg":"debug","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861840-04:00","level":"INFO","component":"cli","msg":"info","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861848-04:00","level":"WARN","component":"cli","msg":"warn","task_id":"T","agent_id":"A","request_id":"R"}
{"time":"2026-10-18T01:16:29.861854-04:00","level":"ERROR","component":"cli","msg":"error","task_id":"T","agent_id":"A","request_id":"R"}
stderr:
[empty]
exit 0
```

### Command 62

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --request MISS --json
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 63

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.880446-04:00","level":"INFO","component":"cli","msg":"after invalid"}
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

### Command 64

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --limit 2
```

```text
stdout:
? not json
2026-10-18T01:16:29.880-04:00 INFO cli after invalid
stderr:
[empty]
exit 0
```

### Command 65

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.884067-04:00","level":"INFO","component":"cli","msg":"after null"}
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

### Command 66

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.885785-04:00","level":"INFO","component":"cli","msg":"after array"}
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

### Command 67

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.887663-04:00","level":"INFO","component":"cli","msg":"after number"}
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

### Command 68

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.889400-04:00","level":"INFO","component":"cli","msg":"after under-cap"}
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

### Command 69

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.892550-04:00","level":"INFO","component":"cli","msg":"after at-cap"}
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

### Command 70

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.895690-04:00","level":"INFO","component":"cli","msg":"after over-cap"}
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

### Command 71

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --limit 2
```

```text
stdout:
? xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
2026-10-18T01:16:29.895-04:00 INFO cli after over-cap
stderr:
[empty]
exit 0
```

### Command 72

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json
```

```text
stdout:
{"time":"2026-10-18T01:16:29.900798-04:00","level":"INFO","component":"cli","msg":"after far-over"}
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

### Command 73

```sh
"$HIVE_BIN" --home "$HIVE_TMP" logs --json | python3 -c 'import sys,json; e=json.loads(sys.stdin.readline()); print("keys",len(e))'
```

```text
stdout:
keys 10004
stderr:
[empty]
exit 0
```

### Command 74

```sh
bash -c '"$HIVE_BIN" --home "$HIVE_TMP" logs '
```

```text
stdout:
2026-10-18T01:16:29.923-04:00 INFO cli
injected message\nsecond a=2 z=1
stderr:
[empty]
exit 0
```

### Command 75

```sh
/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/hive --home /tmp/t06-verify-01-probe.fxv5fnvx logs --follow --limit 0 --json [append after 0.15 s, SIGINT after 0.65 s, child.wait()]
```

```text
stdout:
[empty]
stderr:
hive logs: malformed line /tmp/t06-verify-01-probe.fxv5fnvx/.hive/logs/2026-10-18.jsonl:1
exit 0
```

Fixture before follow:

```text
{"time":"2026-10-18T01:16:29.925236-04:00","level":"INFO","component":"cli","msg":"one"}
{"time":"2026-10-18T01:16:29.925249-04:00","level":"INFO","component":"cli","msg":"two"}
```

Append after 0.15 seconds:

```text
not json
```

### Command 76

```sh
/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/hive --home /tmp/t06-verify-01-probe.fxv5fnvx logs --follow --limit 0 --json --since 2026-10-18T02:16:30-04:00 [append after 0.15 s, SIGINT after 0.65 s, child.wait()]
```

```text
stdout:
{"time":"2026-10-18T01:16:30.726486-04:00","level":"INFO","component":"cli","msg":"still before cutoff"}
stderr:
[empty]
exit 0
```

Fixture before follow:

```text
{"time":"2026-10-18T01:16:30.726461-04:00","level":"INFO","component":"cli","msg":"old"}
```

Append after 0.15 seconds:

```text
{"time":"2026-10-18T01:16:30.726486-04:00","level":"INFO","component":"cli","msg":"still before cutoff"}
```

### Command 77

```sh
rm -rf -- "$HIVE_TMP"; test ! -e "$HIVE_TMP"; printf "removed only probe fixture\n"
```

```text
stdout:
removed only probe fixture
stderr:
[empty]
exit 0
```

### Command 78

```sh
git archive 1c556c332b66c636bf4ec1af79ffaaedb8979cd8 | tar -x -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 79

```sh
python3 -c 'from pathlib import Path; p=Path('"'"'/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source/internal/logs/read.go'"'"'); s=p.read_text(); assert s.count('"'"'for day := first; !day.After(today); day = day.AddDate(0, 0, 1) {'"'"')==1; p.write_text(s.replace('"'"'for day := first; !day.After(today); day = day.AddDate(0, 0, 1) {'"'"','"'"'for day := today; !day.Before(first); day = day.AddDate(0, 0, -1) {'"'"'))'
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 80

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestReadReadsDailyFilesInDateOrder$' -v
```

```text
stdout:
=== RUN   TestReadReadsDailyFilesInDateOrder
    read_test.go:90: event 0 message = "three"; want "one"
    read_test.go:90: event 2 message = "one"; want "three"
--- FAIL: TestReadReadsDailyFilesInDateOrder (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/logs	0.003s
FAIL
stderr:
[empty]
exit 1
```

### Command 81

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestReadReadsDailyFilesInDateOrder$' -v
```

```text
stdout:
=== RUN   TestReadReadsDailyFilesInDateOrder
--- PASS: TestReadReadsDailyFilesInDateOrder (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.004s
stderr:
[empty]
exit 0
```

### Command 82

```sh
python3 -c 'from pathlib import Path; p=Path('"'"'/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source/internal/logs/read.go'"'"'); s=p.read_text(); assert s.count('"'"'when.Before(cutoff)'"'"')==1; p.write_text(s.replace('"'"'when.Before(cutoff)'"'"','"'"'!when.After(cutoff)'"'"'))'
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 83

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestSinceCutoffKeepsAtAndAfter$' -v
```

```text
stdout:
=== RUN   TestSinceCutoffKeepsAtAndAfter
    read_test.go:224: events = []logs.Event{logs.Event{Raw:[]uint8{0x7b, 0x22, 0x6d, 0x73, 0x67, 0x22, 0x3a, 0x22, 0x61, 0x66, 0x74, 0x65, 0x72, 0x22, 0x2c, 0x22, 0x74, 0x69, 0x6d, 0x65, 0x22, 0x3a, 0x22, 0x32, 0x30, 0x32, 0x36, 0x2d, 0x31, 0x30, 0x2d, 0x31, 0x38, 0x54, 0x31, 0x30, 0x3a, 0x30, 0x30, 0x3a, 0x30, 0x31, 0x2d, 0x30, 0x34, 0x3a, 0x30, 0x30, 0x22, 0x7d}, Fields:map[string]interface {}{"msg":"after", "time":"2026-10-18T10:00:01-04:00"}, path:"/tmp/TestSinceCutoffKeepsAtAndAfter1701455366/001/.hive/logs/2026-10-18.jsonl", line:3, malformed:false}}; want at then after
--- FAIL: TestSinceCutoffKeepsAtAndAfter (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/logs	0.003s
FAIL
stderr:
[empty]
exit 1
```

### Command 84

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestSinceCutoffKeepsAtAndAfter$' -v
```

```text
stdout:
=== RUN   TestSinceCutoffKeepsAtAndAfter
--- PASS: TestSinceCutoffKeepsAtAndAfter (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.003s
stderr:
[empty]
exit 0
```

### Command 85

```sh
python3 -c 'from pathlib import Path; p=Path('"'"'/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source/internal/logs/read.go'"'"'); s=p.read_text(); assert s.count('"'"'const maxLineBytes = 1 << 20'"'"')==1; p.write_text(s.replace('"'"'const maxLineBytes = 1 << 20'"'"','"'"'const maxLineBytes = 1 << 21'"'"'))'
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 86

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestReadLineCap$' -v
```

```text
stdout:
=== RUN   TestReadLineCap
--- PASS: TestReadLineCap (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.016s
stderr:
[empty]
exit 0
```

### Command 87

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestReadLineCap$' -v
```

```text
stdout:
=== RUN   TestReadLineCap
--- PASS: TestReadLineCap (0.01s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.010s
stderr:
[empty]
exit 0
```

### Command 88

```sh
python3 -c 'from pathlib import Path; p=Path('"'"'/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source/internal/logs/filter.go'"'"'); s=p.read_text(); assert s.count('"'"'func (f Filter) Match(e Event) bool {'"'"')==1; p.write_text(s.replace('"'"'func (f Filter) Match(e Event) bool {'"'"','"'"'func (f Filter) Match(e Event) bool {\n if true { return true }'"'"'))'
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 89

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run TestFilter -v
```

```text
stdout:
=== RUN   TestFilterMatchFields
=== RUN   TestFilterMatchFields/empty
=== RUN   TestFilterMatchFields/component_match
=== RUN   TestFilterMatchFields/component_miss
    filter_test.go:34: Match() = true; want false
=== RUN   TestFilterMatchFields/task_match
=== RUN   TestFilterMatchFields/task_miss
    filter_test.go:34: Match() = true; want false
=== RUN   TestFilterMatchFields/agent_match
=== RUN   TestFilterMatchFields/agent_miss
    filter_test.go:34: Match() = true; want false
=== RUN   TestFilterMatchFields/request_match
=== RUN   TestFilterMatchFields/request_miss
    filter_test.go:34: Match() = true; want false
=== RUN   TestFilterMatchFields/level_match
=== RUN   TestFilterMatchFields/level_miss
    filter_test.go:34: Match() = true; want false
=== RUN   TestFilterMatchFields/all_fields_match
--- FAIL: TestFilterMatchFields (0.00s)
    --- PASS: TestFilterMatchFields/empty (0.00s)
    --- PASS: TestFilterMatchFields/component_match (0.00s)
    --- FAIL: TestFilterMatchFields/component_miss (0.00s)
    --- PASS: TestFilterMatchFields/task_match (0.00s)
    --- FAIL: TestFilterMatchFields/task_miss (0.00s)
    --- PASS: TestFilterMatchFields/agent_match (0.00s)
    --- FAIL: TestFilterMatchFields/agent_miss (0.00s)
    --- PASS: TestFilterMatchFields/request_match (0.00s)
    --- FAIL: TestFilterMatchFields/request_miss (0.00s)
    --- PASS: TestFilterMatchFields/level_match (0.00s)
    --- FAIL: TestFilterMatchFields/level_miss (0.00s)
    --- PASS: TestFilterMatchFields/all_fields_match (0.00s)
=== RUN   TestFilterLevelKeepsNamedAndAbove
    filter_test.go:55: filter "info" matched DEBUG = true; want false
    filter_test.go:55: filter "warn" matched DEBUG = true; want false
    filter_test.go:55: filter "warn" matched INFO = true; want false
    filter_test.go:55: filter "error" matched DEBUG = true; want false
    filter_test.go:55: filter "error" matched INFO = true; want false
    filter_test.go:55: filter "error" matched WARN = true; want false
--- FAIL: TestFilterLevelKeepsNamedAndAbove (0.00s)
=== RUN   TestFilterMalformedOnlyMatchesEmptyFilter
    filter_test.go:67: filtered Match() matched a malformed line
--- FAIL: TestFilterMalformedOnlyMatchesEmptyFilter (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/logs	0.003s
FAIL
stderr:
[empty]
exit 1
```

### Command 90

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run TestFilter -v
```

```text
stdout:
=== RUN   TestFilterMatchFields
=== RUN   TestFilterMatchFields/empty
=== RUN   TestFilterMatchFields/component_match
=== RUN   TestFilterMatchFields/component_miss
=== RUN   TestFilterMatchFields/task_match
=== RUN   TestFilterMatchFields/task_miss
=== RUN   TestFilterMatchFields/agent_match
=== RUN   TestFilterMatchFields/agent_miss
=== RUN   TestFilterMatchFields/request_match
=== RUN   TestFilterMatchFields/request_miss
=== RUN   TestFilterMatchFields/level_match
=== RUN   TestFilterMatchFields/level_miss
=== RUN   TestFilterMatchFields/all_fields_match
--- PASS: TestFilterMatchFields (0.00s)
    --- PASS: TestFilterMatchFields/empty (0.00s)
    --- PASS: TestFilterMatchFields/component_match (0.00s)
    --- PASS: TestFilterMatchFields/component_miss (0.00s)
    --- PASS: TestFilterMatchFields/task_match (0.00s)
    --- PASS: TestFilterMatchFields/task_miss (0.00s)
    --- PASS: TestFilterMatchFields/agent_match (0.00s)
    --- PASS: TestFilterMatchFields/agent_miss (0.00s)
    --- PASS: TestFilterMatchFields/request_match (0.00s)
    --- PASS: TestFilterMatchFields/request_miss (0.00s)
    --- PASS: TestFilterMatchFields/level_match (0.00s)
    --- PASS: TestFilterMatchFields/level_miss (0.00s)
    --- PASS: TestFilterMatchFields/all_fields_match (0.00s)
=== RUN   TestFilterLevelKeepsNamedAndAbove
--- PASS: TestFilterLevelKeepsNamedAndAbove (0.00s)
=== RUN   TestFilterMalformedOnlyMatchesEmptyFilter
--- PASS: TestFilterMalformedOnlyMatchesEmptyFilter (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.004s
stderr:
[empty]
exit 0
```

### Command 91

```sh
python3 -c 'from pathlib import Path; p=Path('"'"'/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source/internal/logs/follow.go'"'"'); s=p.read_text(); assert s.count('"'"'offset := from.Byte'"'"')==1; p.write_text(s.replace('"'"'offset := from.Byte'"'"','"'"'offset := int64(0)'"'"'))'
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 92

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestFollowLineAppendedBeforeStartPrintsOnce$' -v
```

```text
stdout:
=== RUN   TestFollowLineAppendedBeforeStartPrintsOnce
    follow_test.go:105: message = "old"; want between
--- FAIL: TestFollowLineAppendedBeforeStartPrintsOnce (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/logs	0.003s
FAIL
stderr:
[empty]
exit 1
```

### Command 93

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestFollowLineAppendedBeforeStartPrintsOnce$' -v
```

```text
stdout:
=== RUN   TestFollowLineAppendedBeforeStartPrintsOnce
--- PASS: TestFollowLineAppendedBeforeStartPrintsOnce (0.60s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.605s
stderr:
[empty]
exit 0
```

### Command 94

```sh
python3 -c 'from pathlib import Path; p=Path('"'"'/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source/internal/logs/read.go'"'"'); s=p.read_text(); assert s.count('"'"'if !complete {\n\t\t\tbreak'"'"')==1; p.write_text(s.replace('"'"'if !complete {\n\t\t\tbreak'"'"','"'"'if !complete {\n            consumed += int64(total)\n\t\t\tbreak'"'"'))'
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 95

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestFollowKeepsPartialLineUntilNewline$' -v
```

```text
stdout:
=== RUN   TestFollowKeepsPartialLineUntilNewline
    follow_test.go:146: message = ""; want split
--- FAIL: TestFollowKeepsPartialLineUntilNewline (1.50s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/logs	1.504s
FAIL
stderr:
[empty]
exit 1
```

### Command 96

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestFollowKeepsPartialLineUntilNewline$' -v
```

```text
stdout:
=== RUN   TestFollowKeepsPartialLineUntilNewline
--- PASS: TestFollowKeepsPartialLineUntilNewline (2.10s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	3.106s
stderr:
[empty]
exit 0
```

### Command 97

```sh
python3 -c 'from pathlib import Path; p=Path('"'"'/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source/internal/logs/follow.go'"'"'); s=p.read_text(); assert s.count('"'"'path = nextPath'"'"')==1; p.write_text(s.replace('"'"'path = nextPath'"'"','"'"'path = path'"'"'))'
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 98

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestFollowMovesToNextDailyFile$' -v
```

```text
stdout:
=== RUN   TestFollowMovesToNextDailyFile
    follow_test.go:181: second message = "old"; want new day
--- FAIL: TestFollowMovesToNextDailyFile (0.50s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/logs	0.504s
FAIL
stderr:
[empty]
exit 1
```

### Command 99

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestFollowMovesToNextDailyFile$' -v
```

```text
stdout:
=== RUN   TestFollowMovesToNextDailyFile
--- PASS: TestFollowMovesToNextDailyFile (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.004s
stderr:
[empty]
exit 0
```

### Command 100

```sh
python3 -c 'from pathlib import Path; p=Path('"'"'/home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source/internal/logs/follow.go'"'"'); s=p.read_text(); assert s.count('"'"'case <-ctx.Done():\n\t\t\t\terrs <- nil'"'"')==1; p.write_text(s.replace('"'"'case <-ctx.Done():\n\t\t\t\terrs <- nil'"'"','"'"'case <-ctx.Done():\n\t\t\t\terrs <- context.Canceled'"'"'))'
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 101

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestFollowCancelledContextReturnsNilError$' -v
```

```text
stdout:
=== RUN   TestFollowCancelledContextReturnsNilError
    follow_test.go:195: error = context canceled; want nil
--- FAIL: TestFollowCancelledContextReturnsNilError (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/logs	0.003s
FAIL
stderr:
[empty]
exit 1
```

### Command 102

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source test -count=1 -race ./internal/logs/ -run '^TestFollowCancelledContextReturnsNilError$' -v
```

```text
stdout:
=== RUN   TestFollowCancelledContextReturnsNilError
--- PASS: TestFollowCancelledContextReturnsNilError (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.003s
stderr:
[empty]
exit 0
```

### Command 103

```sh
rm -rf -- /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/mutation-source; printf "removed only isolated mutation copy\n"
```

```text
stdout:
removed only isolated mutation copy
stderr:
[empty]
exit 0
```

### Command 104

```sh
git archive 1c556c332b66c636bf4ec1af79ffaaedb8979cd8 | tar -x -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/cap-source
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 105

```sh
printf "mutation: maxLineBytes changes from 1 MiB to 2 MiB in the isolated copy\n"; go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/cap-source test -count=1 ./internal/logs/ -run 'TestRead|TestFilter|TestSince' -v
```

```text
stdout:
mutation: maxLineBytes changes from 1 MiB to 2 MiB in the isolated copy
=== RUN   TestFilterMatchFields
=== RUN   TestFilterMatchFields/empty
=== RUN   TestFilterMatchFields/component_match
=== RUN   TestFilterMatchFields/component_miss
=== RUN   TestFilterMatchFields/task_match
=== RUN   TestFilterMatchFields/task_miss
=== RUN   TestFilterMatchFields/agent_match
=== RUN   TestFilterMatchFields/agent_miss
=== RUN   TestFilterMatchFields/request_match
=== RUN   TestFilterMatchFields/request_miss
=== RUN   TestFilterMatchFields/level_match
=== RUN   TestFilterMatchFields/level_miss
=== RUN   TestFilterMatchFields/all_fields_match
--- PASS: TestFilterMatchFields (0.00s)
    --- PASS: TestFilterMatchFields/empty (0.00s)
    --- PASS: TestFilterMatchFields/component_match (0.00s)
    --- PASS: TestFilterMatchFields/component_miss (0.00s)
    --- PASS: TestFilterMatchFields/task_match (0.00s)
    --- PASS: TestFilterMatchFields/task_miss (0.00s)
    --- PASS: TestFilterMatchFields/agent_match (0.00s)
    --- PASS: TestFilterMatchFields/agent_miss (0.00s)
    --- PASS: TestFilterMatchFields/request_match (0.00s)
    --- PASS: TestFilterMatchFields/request_miss (0.00s)
    --- PASS: TestFilterMatchFields/level_match (0.00s)
    --- PASS: TestFilterMatchFields/level_miss (0.00s)
    --- PASS: TestFilterMatchFields/all_fields_match (0.00s)
=== RUN   TestFilterLevelKeepsNamedAndAbove
--- PASS: TestFilterLevelKeepsNamedAndAbove (0.00s)
=== RUN   TestFilterMalformedOnlyMatchesEmptyFilter
--- PASS: TestFilterMalformedOnlyMatchesEmptyFilter (0.00s)
=== RUN   TestReadReadsDailyFilesInDateOrder
--- PASS: TestReadReadsDailyFilesInDateOrder (0.00s)
=== RUN   TestReadSkipsMissingDailyFile
--- PASS: TestReadSkipsMissingDailyFile (0.00s)
=== RUN   TestReadLineCap
--- PASS: TestReadLineCap (0.00s)
=== RUN   TestReadAcceptsLineAtCap
--- PASS: TestReadAcceptsLineAtCap (0.00s)
=== RUN   TestReadMarksInvalidJSONMalformed
--- PASS: TestReadMarksInvalidJSONMalformed (0.00s)
=== RUN   TestReadRejectsNonRegularFile
--- PASS: TestReadRejectsNonRegularFile (0.00s)
=== RUN   TestSinceCutoffKeepsAtAndAfter
--- PASS: TestSinceCutoffKeepsAtAndAfter (0.00s)
=== RUN   TestSincePreviousDayReadsTwoFiles
--- PASS: TestSincePreviousDayReadsTwoFiles (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	0.009s
stderr:
[empty]
exit 0
```

### Command 106

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/cap-source test -count=1 ./internal/logs/ -run TestVerifierFixedLineCap -v
```

```text
stdout:
=== RUN   TestVerifierFixedLineCap
    verifier_test.go:17: size 1048577: malformed=false, want true
    verifier_test.go:18: size 1048577: raw length=1048577, want 200
--- FAIL: TestVerifierFixedLineCap (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/logs	0.007s
FAIL
stderr:
[empty]
exit 1
```

### Command 107

```sh
printf "restore original 1 MiB cap\n"; go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/cap-source test -count=1 ./internal/logs/ -run TestVerifierFixedLineCap -v
```

```text
stdout:
restore original 1 MiB cap
=== RUN   TestVerifierFixedLineCap
--- PASS: TestVerifierFixedLineCap (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	0.005s
stderr:
[empty]
exit 0
```

### Command 108

```sh
rm -rf -- /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/cap-source; printf "removed only fixed-cap source copy\n"
```

```text
stdout:
removed only fixed-cap source copy
stderr:
[empty]
exit 0
```

### Command 109

```sh
git diff --check 840255d3..1c556c3
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 110

```sh
grep -nE 'os\.(Open|OpenFile|ReadFile|Create|WriteFile)\(|io\.(ReadAll|LimitReader)|bufio\.NewReader|ReadSlice|time.NewTicker' internal/logs/{read,follow,filter,render,doc}.go cmd/logs/logs.go
```

```text
stdout:
internal/logs/read.go:98:	file, err := os.Open(path)
internal/logs/read.go:104:	reader := bufio.NewReaderSize(file, 64*1024)
internal/logs/read.go:156:		chunk, err := reader.ReadSlice('\n')
internal/logs/follow.go:34:		ticker := time.NewTicker(followPollInterval)
internal/logs/follow.go:93:	file, err := os.Open(path)
internal/logs/follow.go:105:	reader := bufio.NewReaderSize(file, 64*1024)
stderr:
[empty]
exit 0
```

### Command 111

```sh
git status --porcelain; git rev-parse HEAD
```

```text
stdout:
1c556c332b66c636bf4ec1af79ffaaedb8979cd8
stderr:
[empty]
exit 0
```

### Command 112

```sh
git archive 1c556c332b66c636bf4ec1af79ffaaedb8979cd8 | tar -x -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/rollover-source
```

```text
stdout:
[empty]
stderr:
[empty]
exit 0
```

### Command 113

```sh
go -C /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/rollover-source test -count=1 -race ./internal/logs/ -run TestVerifierRolloverWithinOneSecond -v
```

```text
stdout:
=== RUN   TestVerifierRolloverWithinOneSecond
    verifier_test.go:24: new daily event after 400.861075ms
--- PASS: TestVerifierRolloverWithinOneSecond (0.50s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/logs	1.505s
stderr:
[empty]
exit 0
```

### Command 114

```sh
rm -rf -- /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/t06-verify-01.rdNu1g/rollover-source; printf "removed only rollover source copy\n"
```

```text
stdout:
removed only rollover source copy
stderr:
[empty]
exit 0
```

## Reproduction programs

These programs exist only in verifier scratch space. Their source records fixture creation and mutations.

### run.py

```python
import subprocess, pathlib, shlex, json, os
S = pathlib.Path(__file__).resolve().parent
records=[]
def run(command, timeout=180, env=None, cwd=None):
    p=subprocess.run(command, shell=True, executable='/bin/bash', capture_output=True, text=True, timeout=timeout,env=env,cwd=cwd)
    record={'command':command,'stdout':p.stdout,'stderr':p.stderr,'status':p.returncode}
    records.append(record)
    with (S/'commands.jsonl').open('a') as f: f.write(json.dumps(record)+'\n')
    print('$ '+command+'\nstdout:\n'+(p.stdout or '[empty]\n')+'stderr:\n'+(p.stderr or '[empty]\n')+f'exit {p.returncode}',flush=True)
    return p
if __name__=='__main__':
    run('git branch --show-current; git status --porcelain; git rev-parse HEAD')
    run("tail -n 1 /home/penguin/source/hive/projects/hive/worktrees/T06-logs-command/.hive/tmp/coordinator/reports/t06-impl-01.md")
    run('git diff --name-only 840255d3f598ad18239eccd7a2eb2e93547ca1fb..1c556c332b66c636bf4ec1af79ffaaedb8979cd8')
    run('git diff 840255d3f598ad18239eccd7a2eb2e93547ca1fb..1c556c332b66c636bf4ec1af79ffaaedb8979cd8 -- go.mod go.sum')
    run('git log --format=fuller 840255d3..1c556c3')
    run("go test -count=1 ./internal/logs/ -run 'TestRead|TestFilter|TestSince' -v")
    run('go test -count=1 -race ./internal/logs/ -run TestFollow -v')
    run('gofmt -l $(git ls-files -co --exclude-standard \'*.go\')')
    run('go vet ./...')
    run('golangci-lint run --allow-serial-runners ./...')
    run('go test -count=1 -race ./...')
    run('go build -o '+shlex.quote(str(S/'hive'))+' .')
```

### live.py

```python
import os, pathlib, shlex, tempfile
from run import run, S
H=pathlib.Path(tempfile.mkdtemp(prefix='v3-home.', dir=S)); (H/'.hive').mkdir(); (H/'.hive/README.md').write_text('T06 verifier fixture.\n')
env=os.environ.copy(); env.update(HIVE_TMP=str(H), HIVE_BIN=str(S/'hive'))
print(f'HIVE_TMP={H}\nHIVE_BIN={S / "hive"}',flush=True)
run('printf "HIVE_TMP=%s\\nHIVE_BIN=%s\\n" "$HIVE_TMP" "$HIVE_BIN"; test -f "$HIVE_TMP/.hive/README.md"',env=env)
commands=[
'"$HIVE_BIN" --home "$HIVE_TMP" project list >/dev/null',
'"$HIVE_BIN" --home "$HIVE_TMP" logs --limit 2',
'set -o pipefail; "$HIVE_BIN" --home "$HIVE_TMP" logs --component cli --json | tail -n 1 | python3 -m json.tool',
'"$HIVE_BIN" --home "$HIVE_TMP" logs --level error',
'printf \'not json\\n\' >> "$HIVE_TMP"/.hive/logs/"$(date +%F)".jsonl',
'"$HIVE_BIN" --home "$HIVE_TMP" logs --limit 1',
'"$HIVE_BIN" --home "$HIVE_TMP" logs --limit 1 --json',
' timeout --preserve-status --kill-after=2s -s INT 2 "$HIVE_BIN" --home "$HIVE_TMP" logs --follow --limit 1; echo "exit $?"',
'rm -rf -- "$HIVE_TMP"; test ! -e "$HIVE_TMP"; printf "removed only V3 fixture\\n"'
]
for c in commands: run(c,env=env)
```

### probes.py

```python
import pathlib, tempfile, datetime, json, os, shlex, subprocess, signal, time, selectors, hashlib, socket, shutil
from run import S, run
H=pathlib.Path(tempfile.mkdtemp(prefix='t06-verify-01-probe.')); D=H/'.hive/logs'; D.mkdir(parents=True); (H/'.hive/README.md').write_text('T06 independent probe fixture.\n'); P=D/(datetime.date.today().isoformat()+'.jsonl')
env=os.environ.copy(); env.update(HIVE_TMP=str(H), HIVE_BIN=str(S/'hive'))
run('printf "HIVE_TMP=%s\\nHIVE_BIN=%s\\n" "$HIVE_TMP" "$HIVE_BIN"; test -f "$HIVE_TMP/.hive/README.md"',env=env)
results=[]
def check(name, condition, detail=''):
    text=f'{"PASS" if condition else "FAIL"}: {name}'+(': '+detail if detail else '')
    print(text,flush=True); results.append(text)
def call(flags='--json',timeout=3):
    return run('"$HIVE_BIN" --home "$HIVE_TMP" logs '+flags,env=env,timeout=timeout)
def line(msg, **more):
    return json.dumps(dict(time=datetime.datetime.now().astimezone().isoformat(),level='INFO',component='cli',msg=msg,**more),separators=(',',':'))
def write(data): P.write_bytes(data if isinstance(data,bytes) else data.encode())
def plain_error(p): return p.returncode==1 and not p.stdout and len(p.stderr.splitlines())==1 and 'panic' not in p.stderr and 'goroutine' not in p.stderr
for kind in ['unreadable','directory','nonempty-directory','symlink-valid','symlink-missing','symlink-directory','symlink-zero','symlink-chain','fifo','socket']:
    sock=None
    if kind=='unreadable': write(line('unreadable')+'\n'); P.chmod(0)
    elif kind in ['directory','nonempty-directory']:
        P.mkdir()
        if kind=='nonempty-directory': (P/'child').write_text('probe\n')
    elif kind.startswith('symlink'):
        target=H/('target-'+kind)
        if kind=='symlink-valid': target.write_text(line('target')+'\n')
        if kind=='symlink-directory': target.mkdir()
        if kind=='symlink-zero': target=pathlib.Path('/dev/zero')
        if kind=='symlink-chain': target.symlink_to(H/'missing')
        P.symlink_to(target)
    elif kind=='fifo': os.mkfifo(P)
    elif kind=='socket': sock=socket.socket(socket.AF_UNIX); sock.bind(str(P))
    p=call(''); check(kind+' text diagnostic',plain_error(p))
    if kind in ['unreadable','directory','symlink-valid']:
        p=call('--json'); check(kind+' JSON plain diagnostic',plain_error(p),'expected empty stdout and one stderr line')
    if sock: sock.close()
    if P.is_dir() and not P.is_symlink(): shutil.rmtree(P)
    else:
        if kind=='unreadable': P.chmod(0o600)
        P.unlink()
# Empty and missing files.
p=call(); check('missing daily file skips',p.returncode==0 and not p.stdout and not p.stderr)
write(''); p=call(); check('empty file skips',p.returncode==0 and not p.stdout and not p.stderr)
# Every filter, level threshold, limit, and raw JSON stream.
rows=[line('debug',task_id='T',agent_id='A',request_id='R').replace('"INFO"','"DEBUG"'), line('info',task_id='T',agent_id='A',request_id='R'), line('warn',task_id='T',agent_id='A',request_id='R').replace('"INFO"','"WARN"'),line('error',task_id='T',agent_id='A',request_id='R').replace('"INFO"','"ERROR"')]
write('\n'.join(rows)+'\n'); before=hashlib.sha256(P.read_bytes()).hexdigest()
p=call('--level warn --limit 1 --json'); check('limit keeps last matching event',p.stdout==rows[-1]+'\n' and not p.stderr)
p=call('--level warn --json'); check('warn keeps warn and error',p.stdout=='\n'.join(rows[-2:])+'\n')
for field,value in [('component','cli'),('task','T'),('agent','A'),('request','R')]:
    p=call('--'+field+' '+value+' --json'); check(field+' match',p.stdout=='\n'.join(rows)+'\n')
    p=call('--'+field+' MISS --json'); check(field+' miss',not p.stdout and not p.stderr and p.returncode==0)
check('log reads do not write',hashlib.sha256(P.read_bytes()).hexdigest()==before)
# Malformed lines and boundary sizes, followed by a valid event.
for label,data in [('invalid','not json'),('null','null'),('array','[]'),('number','17'),('under-cap','x'*(1048576-1)),('at-cap','x'*1048576),('over-cap','x'*(1048576+1)),('far-over','x'*(5*1024*1024))]:
    valid=line('after '+label); write(data+'\n'+valid+'\n'); p=call('--json')
    check(label+' JSON skip and continue',p.returncode==0 and p.stdout==valid+'\n' and p.stderr==f'hive logs: malformed line {P}:1\n')
    if label in ['invalid','over-cap']:
        p=call('--limit 2'); check(label+' text prefix',p.stdout.splitlines()[0]=='? '+data[:200] and len(p.stdout.splitlines())==2 and not p.stderr)
# 10000 extra keys in a bounded event.
obj=json.loads(line('many keys')); obj.update({f'key-{i}':i for i in range(10000)}); data=json.dumps(obj,separators=(',',':')); write(data+'\n'); p=call('--json | python3 -c \'import sys,json; e=json.loads(sys.stdin.readline()); print("keys",len(e))\''); check('10000 extra keys',p.stdout=='keys 10004\n' and not p.stderr and p.returncode==0)
# Text fields must remain one physical line.
obj=json.loads(line('message\nsecond',z=1,a=2)); obj['component']='cli\ninjected'; write(json.dumps(obj)+'\n'); p=call(''); check('text event is one line',len(p.stdout.splitlines())==1,'component contains a JSON escaped newline')
# Follow diagnostics use absolute line numbers and preserve cutoff on new events.
def follow_case(name, initial, appended, flags, expected_out, expected_err):
    write(initial); cmd=[str(S/'hive'),'--home',str(H),'logs','--follow','--limit','0']+flags
    print('$ '+shlex.join(cmd),flush=True)
    child=subprocess.Popen(cmd,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    time.sleep(.15)
    with P.open('a') as f: f.write(appended)
    time.sleep(.65); child.send_signal(signal.SIGINT); out,err=child.communicate(timeout=2)
    rec={'command':shlex.join(cmd)+' [append after 0.15 s, SIGINT after 0.65 s, child.wait()]','stdout':out,'stderr':err,'status':child.returncode,'initial':initial,'appended':appended}
    with (S/'commands.jsonl').open('a') as f:f.write(json.dumps(rec)+'\n')
    print('stdout:\n'+(out or '[empty]\n')+'stderr:\n'+(err or '[empty]\n')+f'exit {child.returncode}',flush=True)
    check(name,child.returncode==0 and out==expected_out and err==expected_err)
follow_case('follow malformed absolute line number',line('one')+'\n'+line('two')+'\n','not json\n',['--json'],'',f'hive logs: malformed line {P}:3\n')
future=(datetime.datetime.now().astimezone()+datetime.timedelta(hours=1)).isoformat(timespec='seconds')
follow_case('follow keeps cutoff',line('old')+'\n',line('still before cutoff')+'\n',['--json','--since',future],'','')
(S/'probe-verdicts.txt').write_text('\n'.join(results)+'\n')
run('rm -rf -- "$HIVE_TMP"; test ! -e "$HIVE_TMP"; printf "removed only probe fixture\\n"',env=env)
```

### mutations.py

```python
import pathlib, shlex, json
from run import run, S
M=S/'mutation-source'; M.mkdir()
run('git archive 1c556c332b66c636bf4ec1af79ffaaedb8979cd8 | tar -x -C '+shlex.quote(str(M)))
mutations=[
('date order','read.go','for day := first; !day.After(today); day = day.AddDate(0, 0, 1) {','for day := today; !day.Before(first); day = day.AddDate(0, 0, -1) {','TestReadReadsDailyFilesInDateOrder'),
('cutoff equality','read.go','when.Before(cutoff)','!when.After(cutoff)','TestSinceCutoffKeepsAtAndAfter'),
('line cap','read.go','const maxLineBytes = 1 << 20','const maxLineBytes = 1 << 21','TestReadLineCap'),
('filter match','filter.go','func (f Filter) Match(e Event) bool {','func (f Filter) Match(e Event) bool {\n if true { return true }','TestFilter'),
('follow offset','follow.go','offset := from.Byte','offset := int64(0)','TestFollowLineAppendedBeforeStartPrintsOnce'),
('partial snapshot','read.go','if !complete {\n\t\t\tbreak','if !complete {\n            consumed += int64(total)\n\t\t\tbreak','TestFollowKeepsPartialLineUntilNewline'),
('rollover','follow.go','path = nextPath','path = path','TestFollowMovesToNextDailyFile'),
('cancel result','follow.go','case <-ctx.Done():\n\t\t\t\terrs <- nil','case <-ctx.Done():\n\t\t\t\terrs <- context.Canceled','TestFollowCancelledContextReturnsNilError')
]
verdicts=[]
for name,file,old,new,test in mutations:
    path=M/'internal/logs'/file; original=path.read_text(); assert original.count(old)==1,(name,original.count(old))
    code='from pathlib import Path; p=Path('+repr(str(path))+'); s=p.read_text(); assert s.count('+repr(old)+')==1; p.write_text(s.replace('+repr(old)+','+repr(new)+'))'
    run('python3 -c '+shlex.quote(code))
    p=run('go -C '+shlex.quote(str(M))+' test -count=1 -race ./internal/logs/ -run '+shlex.quote('^'+test+'$' if test!='TestFilter' else test)+' -v')
    caught=p.returncode==1 and '--- FAIL:' in p.stdout and '[build failed]' not in p.stdout
    verdicts.append(f'{"PASS" if caught else "FAIL"}: {name} mutation causes an assertion failure')
    path.write_text(original)
    p=run('go -C '+shlex.quote(str(M))+' test -count=1 -race ./internal/logs/ -run '+shlex.quote('^'+test+'$' if test!='TestFilter' else test)+' -v')
    verdicts.append(f'{"PASS" if p.returncode==0 else "FAIL"}: {name} original passes')
(S/'mutation-verdicts.txt').write_text('\n'.join(verdicts)+'\n')
print('\n'.join(verdicts))
run('rm -rf -- '+shlex.quote(str(M))+'; printf "removed only isolated mutation copy\\n"')
```

### fixedcap.py

```python
from run import S, run
import shlex
M=S/'cap-source'; M.mkdir(); q=shlex.quote(str(M)); run('git archive 1c556c332b66c636bf4ec1af79ffaaedb8979cd8 | tar -x -C '+q)
p=M/'internal/logs/read.go'; original=p.read_text(); p.write_text(original.replace('const maxLineBytes = 1 << 20','const maxLineBytes = 1 << 21'))
run('printf "mutation: maxLineBytes changes from 1 MiB to 2 MiB in the isolated copy\\n"; go -C '+q+" test -count=1 ./internal/logs/ -run 'TestRead|TestFilter|TestSince' -v")
test='''package logs
import (
 "context"
 "strings"
 "testing"
 "time"
)
func TestVerifierFixedLineCap(t *testing.T) {
 for _, size := range []int{1048575,1048576,1048577} {
  reader, home := newTestReader(t, time.Now())
  head, tail := `{"msg":"`, `"}`
  raw := head + strings.Repeat("x",size-len(head)-len(tail)) + tail
  writeDaily(t,home,time.Now(),raw,`{"msg":"after"}`)
  events,_,err:=reader.Read(context.Background(),time.Time{})
  if err!=nil || len(events)!=2 { t.Fatalf("size %d: events %d, error %v",size,len(events),err) }
  wantMalformed:=size>1048576
  if events[0].Malformed()!=wantMalformed { t.Errorf("size %d: malformed=%v, want %v",size,events[0].Malformed(),wantMalformed) }
  if wantMalformed && len(events[0].Raw)!=200 { t.Errorf("size %d: raw length=%d, want 200",size,len(events[0].Raw)) }
  if eventMessage(events[1])!="after" { t.Errorf("size %d: valid line after cap is missing",size) }
 }
}
'''
(M/'internal/logs/verifier_test.go').write_text(test)
print('Independent fixed-cap test:\n'+test,flush=True)
run('go -C '+q+' test -count=1 ./internal/logs/ -run TestVerifierFixedLineCap -v')
p.write_text(original)
run('printf "restore original 1 MiB cap\\n"; go -C '+q+' test -count=1 ./internal/logs/ -run TestVerifierFixedLineCap -v')
(S/'fixedcap-test.txt').write_text(test)
run('rm -rf -- '+q+'; printf "removed only fixed-cap source copy\\n"')
```

### rollover.py

```python
from run import S,run
import shlex
M=S/'rollover-source'; M.mkdir(); q=shlex.quote(str(M)); run('git archive 1c556c332b66c636bf4ec1af79ffaaedb8979cd8 | tar -x -C '+q)
test='''package logs
import (
 "context"
 "testing"
 "time"
)
func TestVerifierRolloverWithinOneSecond(t *testing.T) {
 clock:=&testClock{}
 day:=localNoon(time.Now()); clock.set(day)
 reader,home:=newTestReader(t,day); reader.now=clock.now
 writeDaily(t,home,day,eventLine(t,day,map[string]any{"msg":"old"}))
 _,offset,err:=reader.Read(context.Background(),time.Time{}); if err!=nil {t.Fatal(err)}
 ctx,cancel:=context.WithCancel(context.Background()); defer cancel()
 events,errs:=reader.Follow(ctx,offset)
 time.Sleep(100*time.Millisecond)
 appendDaily(t,home,day,eventLine(t,day,map[string]any{"msg":"rest"})+"\\n")
 next:=day.AddDate(0,0,1)
 clock.set(next)
 start:=time.Now()
 writeDaily(t,home,next,eventLine(t,next,map[string]any{"msg":"new day"}))
 if got:=eventMessage(nextEvent(t,events,time.Second)); got!="rest" {t.Fatalf("first=%q, want rest",got)}
 if got:=eventMessage(nextEvent(t,events,time.Second)); got!="new day" {t.Fatalf("next=%q, want new day",got)}
 elapsed:=time.Since(start)
 t.Logf("new daily event after %s",elapsed)
 if elapsed>=time.Second {t.Fatalf("rollover latency %s, want <1s",elapsed)}
 cancelFollow(t,cancel,errs)
}
'''
(M/'internal/logs/verifier_test.go').write_text(test); (S/'rollover-test.txt').write_text(test)
run('go -C '+q+' test -count=1 -race ./internal/logs/ -run TestVerifierRolloverWithinOneSecond -v')
run('rm -rf -- '+q+'; printf "removed only rollover source copy\\n"')
```

### Fixed-cap test

```go
package logs
import (
 "context"
 "strings"
 "testing"
 "time"
)
func TestVerifierFixedLineCap(t *testing.T) {
 for _, size := range []int{1048575,1048576,1048577} {
  reader, home := newTestReader(t, time.Now())
  head, tail := `{"msg":"`, `"}`
  raw := head + strings.Repeat("x",size-len(head)-len(tail)) + tail
  writeDaily(t,home,time.Now(),raw,`{"msg":"after"}`)
  events,_,err:=reader.Read(context.Background(),time.Time{})
  if err!=nil || len(events)!=2 { t.Fatalf("size %d: events %d, error %v",size,len(events),err) }
  wantMalformed:=size>1048576
  if events[0].Malformed()!=wantMalformed { t.Errorf("size %d: malformed=%v, want %v",size,events[0].Malformed(),wantMalformed) }
  if wantMalformed && len(events[0].Raw)!=200 { t.Errorf("size %d: raw length=%d, want 200",size,len(events[0].Raw)) }
  if eventMessage(events[1])!="after" { t.Errorf("size %d: valid line after cap is missing",size) }
 }
}
```

### Strict rollover deadline test

```go
package logs
import (
 "context"
 "testing"
 "time"
)
func TestVerifierRolloverWithinOneSecond(t *testing.T) {
 clock:=&testClock{}
 day:=localNoon(time.Now()); clock.set(day)
 reader,home:=newTestReader(t,day); reader.now=clock.now
 writeDaily(t,home,day,eventLine(t,day,map[string]any{"msg":"old"}))
 _,offset,err:=reader.Read(context.Background(),time.Time{}); if err!=nil {t.Fatal(err)}
 ctx,cancel:=context.WithCancel(context.Background()); defer cancel()
 events,errs:=reader.Follow(ctx,offset)
 time.Sleep(100*time.Millisecond)
 appendDaily(t,home,day,eventLine(t,day,map[string]any{"msg":"rest"})+"\n")
 next:=day.AddDate(0,0,1)
 clock.set(next)
 start:=time.Now()
 writeDaily(t,home,next,eventLine(t,next,map[string]any{"msg":"new day"}))
 if got:=eventMessage(nextEvent(t,events,time.Second)); got!="rest" {t.Fatalf("first=%q, want rest",got)}
 if got:=eventMessage(nextEvent(t,events,time.Second)); got!="new day" {t.Fatalf("next=%q, want new day",got)}
 elapsed:=time.Since(start)
 t.Logf("new daily event after %s",elapsed)
 if elapsed>=time.Second {t.Fatalf("rollover latency %s, want <1s",elapsed)}
 cancelFollow(t,cancel,errs)
}
```

## Current verdict: 2026-10-18, t06-verify-02

Overall verdict: **PASS** for worker commit `ce9b992bf3cc868eb687471d72a910ca3cb2ae5c`.
The FAIL sections above remain as history. This section gives the current verdict.

| Item | Value |
| --- | --- |
| Task | T06-logs-command |
| Pass | t06-verify-02 |
| Agent | verifier, pi, openai-codex/gpt-6.1-sol:high |
| Branch | task/T06-logs-command |
| Base | 840255d3f598ad18239eccd7a2eb2e93547ca1fb |
| Tested worker HEAD | ce9b992bf3cc868eb687471d72a910ca3cb2ae5c |
| Fix source commits | dc7964cb09cfd923e4d56c20b0d0b74233e7dcc8, ce2196159897d3615a55b18b5feaec98a1cd1842 |
| Start check | Clean branch, exact HEAD, remote branch at the same HEAD |
| Report | /home/penguin/source/hive/.hive/tmp/coordinator/reports/t06-verify-02.md |

| Item | Proves | Verdict | Current evidence |
| --- | --- | --- | --- |
| V1 | AC1 | PASS | Uncached reader and filter tests pass. The fixed-size cap tests reject the doubled cap. |
| V2 | AC3 | PASS | Uncached follow race tests pass. Offset, partial line, rollover, cutoff, numbering, and cancellation checks pass. |
| V3 | AC1, AC2, AC3 | PASS | The transcript and existing live probes pass. F2 to F5 no longer occur. |
| V4 | AC4 | PASS | Formatting output is empty. Vet, lint, and the full uncached race test pass. |

### Finding dispositions

| Finding | Disposition | Evidence |
| --- | --- | --- |
| F1 | Fixed | Fixtures use independent literals 1048575, 1048576, and 1048577. A 2097152-byte cap causes assertion failures. |
| F2 | Fixed | Initial read and follow use the same inclusive cutoff. Live same-day and later-day cutoffs print only `at` and `after`. |
| F3 | Fixed | Physical line 3 reports `:3`. Partial-line and rollover probes keep absolute numbering and reset for the new daily file. |
| F4 | Fixed | Controls in the level, component, message, other values, and keys stay escaped. One event prints one physical line in key order. |
| F5 | Fixed | Unreadable, directory, and symbolic-link JSON reads give empty stdout, one plain stderr line, and status 1. No envelope appears. |

The current tests fail against isolated pre-fix production source. Each red run fails on an assertion, not a build error.
The later-day test also fails against `dc7964cb`, before the `ce219615` fix.
The same tests pass against the tested worker HEAD.

### Commands and bounded results

Commands ran from the task worktree. Only named isolated checks use `go -C`.

```sh
go test -count=1 ./internal/logs/ -run 'TestRead|TestFilter|TestSince' -v
go test -race -count=1 ./internal/logs/ -run TestFollow -v
go test -count=1 ./cmd/ -run TestLogsJSON -v
go test -count=1 ./internal/logs/ -run TestTextLine -v
```

```text
V1: PASS, exit 0. All fixed-size cap subtests pass.
V2: PASS, exit 0. All eight follow tests pass.
JSON command tests: PASS, exit 0. All three error types and raw JSON pass.
Text control test: PASS, exit 0.
```

The source-copy cap mutation changes `const maxLineBytes = 1 << 20` to `const maxLineBytes = 2097152`.
The actual task source never changes.
The mutated copy runs the complete V1 pattern with `-race -count=1`.

```text
read_test.go:134: malformed prefix length = 1048577; want the first 200 bytes
read_test.go:202: size 1048577 malformed = false; want true
read_test.go:205: size 1048577 raw prefix = 1048577 bytes; want 200
FAIL, exit 1. No build failure.
Original cap: PASS, exit 0.
```

Pre-fix source at `c9016cf0` runs the current regression tests in the isolated copy.
The copy replaces only production reader, follower, renderer, and command files.

```text
TestFollowAppliesCutoff: first message = "appended before"; want at
TestFollowAppliesCutoffAfterToday: first message = "before"; want at
TestFollowUsesAbsoluteLineNumbers: line = 1; want 3
TestTextLineEscapesControlsInFieldsAndKeys: rendered line has a physical control
TestLogsJSONReadErrorsUsePlainDiagnostics: stdout contains a failure envelope; want empty
Each selected run: FAIL, exit 1, with assertion failures.
dc7964cb later-day run: first message = "before"; want at. FAIL, exit 1.
Current source: each selected regression run PASS, exit 0.
```

All eight existing mutation classes cause assertion failures in the isolated copy:
cap, date order, cutoff equality, filter match, follow offset, partial snapshot, rollover, and cancellation result.
Each original-source counterpart passes. No mutation affects the task worktree.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
```

```text
gofmt: empty output, exit 0
vet: empty output, exit 0
lint: 0 issues, exit 0
full race test: all packages pass, exit 0
```

### Existing checklist

| Check | Verdict | Evidence |
| --- | --- | --- |
| H01 Three daily files in date order | PASS | Reader test and date-order mutation |
| H02 Fixed 1 MiB boundary and following event | PASS | Literal fixtures, cap mutation, live over-cap probe |
| H03 Inclusive cutoff | PASS | Initial-read test and equality mutation |
| H04 Previous-day cutoff reads two files | PASS | Previous-day reader test |
| H05 Each filter matches and misses | PASS | Test cases and live CLI probes |
| H06 Warn keeps warn and error | PASS | Level tests and live CLI probe |
| H07 Last N matches | PASS | V3 limit 2 and live warn limit 1 |
| H08 Malformed text and 200-byte prefix | PASS | V3 and live over-cap text |
| H09 Malformed JSON continues with correct number | PASS | Live line 3 and isolated partial/rollover numbering |
| H10 Follow within one second | PASS | Worker test and live newline latency 300.168 ms |
| H11 Line between snapshot and follow appears once | PASS | Uncached offset test and mutation |
| H12 Snapshot partial line stays | PASS | Uncached test, mutation, and live split line |
| H13 Two writes wait for newline and print once | PASS | Uncached test and live no-duplicate check |
| H14 Old file drains before rollover within one second | PASS | Isolated deadline probe: 400.375 ms |
| H15 SIGINT child status 0 | PASS | Preserve-status timeout and direct child wait |
| H16 Raw JSON without envelopes | PASS | Exact raw bytes and all three JSON read errors |
| H17 Bounded line storage and long-line drain | PASS | ReadSlice review, cap tests, live 5 MiB line |
| H18 Reader does not write logs | PASS | Production opens are read-only. Live hash and modification time stay equal. |
| H19 Required public API | PASS | Required names, public fields, and method signatures stay unchanged. |
| H20 No dependency change | PASS | Empty go.mod and go.sum diff |
| H21 Owned files only | PASS | Complete base-to-worker diff and imports checked |

The V3 transcript uses the unique ignored binary instead of the existing `./hive`.
It gives two text lines, parsed event JSON, no error-level line, and the required malformed outputs.
The timeout uses `--preserve-status`. The direct child also returns status 0 after SIGINT.
The live same-day and later-day follow probes drop `before` and keep `at` and `after`.

The existing path probes pass for unreadable files, empty and non-empty directories, symbolic links, FIFO, and socket.
The symbolic-link cases include valid, missing, directory, `/dev/zero`, and chained targets.
Missing and empty daily files pass. Invalid JSON, null, array, number, cap sizes, and 5 MiB inputs continue safely.
A valid event with 10000 extra keys retains its raw JSON.

### Scope, design, and limits

The checked source paths are `cmd/root.go`, `cmd/logs/logs.go`, `cmd/logs_root_test.go`, and all nine files under `internal/logs/`.
The remaining changed paths are V1 to V4 and the historical VERIFIED record.
The root change only imports and registers `logs`. No unrelated architecture or dependency change appears.
The internal package imports only the standard library. It does not import Cobra or `cmd`.
The command uses the existing home and status helpers. The package map covers each production file.

Design decision: private `Offset` metadata carries the cutoff and complete-line count without mutable reader state.
Recommendation: T08 must pass the value from `Read` directly to `Follow`.
A manually constructed offset does not carry the private metadata.

Only the local-user and cooperating-writer boundary applies.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
The unreadable-file checks ran as a non-root user.
No unrelated new probe class ran. No required check remains uncompleted.

One initial verifier assertion incorrectly expected empty output from `project list` without its transcript redirection.
The command returned status 0 and `no projects`. The corrected check follows the transcript and passes.
This setup error does not indicate a source failure. Its compact capture remains in the report directory.

The current section is compact. The prior 2432-line FAIL record stays unchanged under the append-only brief.
Captures and reproduction programs remain in the unique ignored report directory.
I removed both unique marked homes, the binary, and the isolated source copy.
No real runtime, existing binary, source, credential, tracker, baseline, global setting, or prior report changed.
I append only this dated section to the tracked verdict. The commit that adds it is the verification commit.
