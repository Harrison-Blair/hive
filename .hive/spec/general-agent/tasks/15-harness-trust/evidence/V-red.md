# V-red Tests red and green

Proves: AC2. New test: `TestEachProfileGivesItsTrustResult` in
`internal/lib/harness/trust_test.go`. Its expected values come from the V1 live
check, not from the profile code.

## Red on the base code

The test file on the base (`dev` at c591a036) with no profile change:

```text
$ go test -run TestEachProfileGivesItsTrustResult ./internal/lib/harness/
# github.com/Harrison-Blair/hive/internal/lib/harness [github.com/Harrison-Blair/hive/internal/lib/harness.test]
./trust_test.go:33:21: profile.Trust undefined (type Profile has no field or method Trust)
FAIL	github.com/Harrison-Blair/hive/internal/lib/harness [build failed]
FAIL
```

The base profile has no trust result. This is the correct reason.

## Red with the type and empty results

After `Trust` and `Profile.Trust()` were added, each profile returned
`Trust{}`:

```text
--- FAIL: TestEachProfileGivesItsTrustResult (0.00s)
    --- FAIL: TestEachProfileGivesItsTrustResult/claude (0.00s)
        trust_test.go:35: Trust().Version = ""; want "2.1.307"
    --- FAIL: TestEachProfileGivesItsTrustResult/codex (0.00s)
        trust_test.go:35: Trust().Version = ""; want "0.160.1"
    --- FAIL: TestEachProfileGivesItsTrustResult/cursor (0.00s)
        trust_test.go:35: Trust().Version = ""; want "2026.10.19-6bbf2c4"
    --- FAIL: TestEachProfileGivesItsTrustResult/opencode (0.00s)
        trust_test.go:35: Trust().Version = ""; want "1.18.32"
    --- FAIL: TestEachProfileGivesItsTrustResult/pi (0.00s)
        trust_test.go:35: Trust().Version = ""; want "1.3.0"
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/harness	0.002s
FAIL
```

## Green on the task branch

```text
$ go test -count=1 -run TestEachProfileGivesItsTrustResult -v ./internal/lib/harness/
=== RUN   TestEachProfileGivesItsTrustResult
=== RUN   TestEachProfileGivesItsTrustResult/claude
=== RUN   TestEachProfileGivesItsTrustResult/codex
=== RUN   TestEachProfileGivesItsTrustResult/cursor
=== RUN   TestEachProfileGivesItsTrustResult/opencode
=== RUN   TestEachProfileGivesItsTrustResult/pi
--- PASS: TestEachProfileGivesItsTrustResult (0.00s)
    --- PASS: TestEachProfileGivesItsTrustResult/claude (0.00s)
    --- PASS: TestEachProfileGivesItsTrustResult/codex (0.00s)
    --- PASS: TestEachProfileGivesItsTrustResult/cursor (0.00s)
    --- PASS: TestEachProfileGivesItsTrustResult/opencode (0.00s)
    --- PASS: TestEachProfileGivesItsTrustResult/pi (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	0.002s
```

## Broken on purpose

`pi.go` changed to `return Trust{Version: "1.3.0", Args: nil}` (the
`--approve` method removed):

```text
--- FAIL: TestEachProfileGivesItsTrustResult (0.00s)
    --- FAIL: TestEachProfileGivesItsTrustResult/pi (0.00s)
        trust_test.go:38: Trust().Args = []string(nil); want []string{"--approve"}
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/harness	0.002s
FAIL
```

`pi.go` was restored from a copy. `git diff` showed no change from the
committed file, and the test passed again.
