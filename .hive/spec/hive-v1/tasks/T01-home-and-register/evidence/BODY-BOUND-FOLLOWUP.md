# T01-home-and-register body bound follow-up

| Item | Value |
| --- | --- |
| Task | T01-home-and-register |
| Pass | t01-body-bound-fix-01 |
| Proves | AC5 for the shared log body bound |
| Base | ebfad2a7658f5f66d40a6cecb3fd6516b2e9b7c1 |
| Branch | `task/T01-log-body-bound` |
| Test commit | The commit that adds `internal/lib/log/body_bound_test.go` |
| Source commit | The commit that changes `internal/lib/log/log.go` |
| Evidence commit | The commit that adds this file |
| Date | 2026-10-18 |
| Agent | worker, pi, opencode-go/deepseek-v4-flash, medium requested, high effective |

## Scope

This pass fixes one defect in the shared log writer. The base
`addAttr` cut the raw Go string at byte 4096. Then `encoding/json` wrote one
U+FFFD for each invalid byte. Each U+FFFD has three bytes. Thus the decoded
body could pass 4096 bytes. The cut could also split a complete UTF-8
character.

The reproduction used the exact base source. A 9000-byte input of 3000 euro
characters gave a decoded body of 4098 bytes and `truncated: true`. The cut
was inside the 1366th euro character. The decoder replaced that partial
character with U+FFFD. A 5000-byte ASCII input gave 4096 bytes. An aligned
9001-byte input gave 4096 bytes. A 4096-byte ASCII input stayed equal and had
no flag.

The same base behavior gave 6000 decoded bytes for 2000 invalid bytes. The
input was short, so the old length test did not set `truncated`. This
observation was not an old policy acceptance. The body contract below settles
the narrow compatibility policy.

The approved boundary is local user and cooperating writer. A hostile
replacement between `Lstat` and `Open` stays deferred for T01. This pass does
not change the path guard.

## Body contract

- The body string value, after the normal JSON decode, has at most 4096 bytes
  and complete UTF-8 characters.
- Valid ASCII and valid Unicode stay exact when they fit.
- Each invalid Go string byte stays one U+FFFD, as `encoding/json` already
  gives. This pass does not add a new rejection or encoding scheme.
- When the decoded sequence is longer than 4096 bytes, the body keeps the
  largest complete-character prefix within 4096 bytes. The event gets
  `truncated: true`.
- A cut can retain 4095 bytes for repeated euro characters.
- The cap keeps work and memory near the retained budget. It does not build
  the full decoded input before the cut.
- This pass does not change escapes, field names, timestamps, the debug level,
  the component field, the command field or the ID fields.

## Changed files

| Path | Change |
| --- | --- |
| `internal/lib/log/log.go` | `addAttr` uses one private cap path. `capBody` keeps the largest complete-character prefix of the standard decoded sequence. |
| `internal/lib/log/body_bound_test.go` | New assertion tests. They use the real `NewWriter`, `slog`, `Body` and daily JSONL files. |
| `spec/hive-v1/evidence/T01-home-and-register/BODY-BOUND-FOLLOWUP.md` | This evidence file. |

No public logging API changed. No dependency, baseline, task file, tracker or
allowlist byte changed. No other production path changed. The herdr package is
the only body producer. The `addAttr` function is the only cap site. No caller
needs a workaround.

## Red run

The new test file was present and the base `log.go` was active. The package
compiled, so each failure came from an assertion.

```sh
git hash-object internal/lib/log/log.go internal/lib/log/body_bound_test.go
```

```text
fbcc1259f8b0633c555a0ae2ce160d8b4cac99aa
e7375b1c3292a1b70c7b3397b2d79c53386eff0f
```

The base `log.go` SHA-256 was
`cedbe1e03c5ba8c4df2f919e13f750edf2ab0c00baae6187f38a076613ceb90e`.
The failing test SHA-256 was
`0f0b9237b88344eab9a70dc38210439c69ec63c2531c2f0cc7e461e0687af3d8`.

```sh
TMPDIR=/tmp/hbt01 T01_BODY_BOUND_CAPTURE=<captures>/red \
  go test ./internal/lib/log/ -run 'TestBodyBound' -count=1 -v
```

```text
--- FAIL: TestBodyBoundRepeatedEuro/unaligned (0.00s)
    decoded body has 4098 bytes; the literal bound is 4096
    an unaligned euro body kept 4098 bytes; want 1365 euro characters (4095 bytes)
--- FAIL: TestBodyBoundFourByteRunes/unaligned (0.00s)
    decoded body has 4102 bytes; the literal bound is 4096
--- FAIL: TestBodyBoundInvalidBytes/short_heavy (0.00s)
    decoded body has 6000 bytes; the literal bound is 4096
    truncated=<nil> present=false; want true
--- FAIL: TestBodyBoundInvalidBytes/boundary (0.00s)
    decoded body has 4098 bytes; the literal bound is 4096
--- FAIL: TestBodyBoundAttributePaths/ordinary (0.00s)
--- FAIL: TestBodyBoundAttributePaths/inherited (0.00s)
--- FAIL: TestBodyBoundAttributePaths/grouped (0.00s)
--- FAIL: TestBodyBoundAttributePaths/nested_grouped_invalid (0.00s)
--- FAIL: TestBodyBoundAttributePaths/inherited_group (0.00s)
FAIL
```

These controls passed on the same base source:

- ASCII below 4096 bytes, exactly 4096 bytes and above 4096 bytes.
- Euro at an exact 4096-byte end and at an aligned cut.
- Four-byte runes at an exact end.
- Invalid bytes below the budget and at an exact 4096-byte decoded end.
- JSON escape characters below and above the budget.

The red run kept 19 real daily JSONL lines and 19 observations in the worker
captures.

## Green run

The new source SHA-256 was
`2deb992bfa25f199ef4fdcc079398102bd6f4e347d3c5d84af342b3476d90d85`.
The test SHA-256 stayed
`0f0b9237b88344eab9a70dc38210439c69ec63c2531c2f0cc7e461e0687af3d8`.

```sh
TMPDIR=/tmp/hbt01 T01_BODY_BOUND_CAPTURE=<captures>/green \
  go test ./internal/lib/log/ -run 'TestBodyBound' -count=1 -v
```

All 19 subcases passed. The euro case kept 4095 bytes and set the flag. The
invalid-byte case kept 1365 U+FFFD characters (4095 bytes) and set the flag.
The exact 4096-byte cases stayed equal and had no flag.

The expected value of each test comes from `encoding/json` and the literal
4096. The tests do not read `maxBodyBytes` or copy `capBody`.

## Mutations

This pass used copies of the package under `/tmp/hbt01/mut` with a
standard-library-only module. Each copy held the final tests.

| Mutation | Change | Result |
| --- | --- | --- |
| Old slice | Cut the raw string at 4096 and set the flag. | 13 failures. The euro case kept 4098 bytes. |
| Larger cap | Move the decoded bound to 4099 bytes. | 18 failures. ASCII kept 4099 bytes. |
| Broken flag | Keep the cap and drop `truncated`. | 18 failures. Each cut case lacked the flag. |
| Restored | Copy the final source back. | Pass. The SHA-256 matched the final source. |

Each mutation failed on assertions. No run counted a compile error, a panic or
a timeout. The captures keep each patch and each raw output.

## Unchanged checks

```sh
TMPDIR=/tmp/hbt01 go test -race -count=1 ./internal/lib/log/ -v
TMPDIR=/tmp/hbt01 go test -race -count=1 ./internal/lib/herdr/ ./cmd/
```

```text
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.062s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.486s
ok  	github.com/Harrison-Blair/hive/cmd	1.095s
```

The full log package held 21 passing tests. They cover the old body tests, the
path guard, rotation, the concurrent JSON lines and the new grouped paths. The
herdr package holds the only body producer. The `cmd` package holds the CLI
log consumers.

The actual CLI ran two `project list` commands against a synthetic home. The
four daily events were valid JSON lines. None had a body field. The CLI text
and JSON output did not change.

## Gate

The gate ran once from the worktree with a short `TMPDIR` outside every home.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 -v .
```

The raw outputs stay in the worker captures. `gofmt` printed nothing. `vet`,
`golangci-lint`, the race suite and `TestSTE` passed. The new prose gave zero
hard findings from the base glossary and the project glossary.

Dependencies and the baseline stayed exact. The `go.mod`, `go.sum` and
`testdata/ste-baseline.json` hashes equal the base hashes.

## Custody

The old reproduction root `/tmp/t01br01.6drZsZOU` had 1552 no-follow entries.
This pass copied 47 compact entries to the worker captures. The copy held the
frozen source, the probe test, the setup record, the command statuses, all raw
outputs, all five real fixtures and the full no-follow inventory. A SHA-256
manifest proved each copy equal to the source. The copied `log.go` hash was
`cedbe1e03c5ba8c4df2f919e13f750edf2ab0c00baae6187f38a076613ceb90e`.
The copied probe hash was
`06797a1bde5cd6cb0723edede89872969b27256f7dcd6a1f4bbbedbcad74a8d6`.
The raw unicode output hash was
`6b73817d2950ddca6d8e0be59a26ba0bcd6ccb76fbc5ff2b71feffaeb415d822`.

The old build caches and the temporary build paths were not necessary for a
replay. This pass wrote a fresh no-follow manifest before it removed the root.
The separate harness log `/tmp/pi-bash-cfefc4e403523a9d.log` stayed untouched.

## Cleanup

This pass removed the transferred reproduction root and its own build copies
and `TMPDIR`. It removed no worktree runtime root, log, state or temporary path
that it did not own. It wrote no root file. It changed no other task, worktree
or agent data.

## Limits

Independent verification stays pending. This file does not write
`BODY-BOUND-VERIFIED.md`. This pass does not prove a total JSON line bound. It
does not prove a hostile local replacement between `Lstat` and `Open`. That
case stays deferred for T01. This pass makes no Task10, checkpoint, root
integration or land claim.
