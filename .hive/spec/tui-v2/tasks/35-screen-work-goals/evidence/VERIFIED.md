# Verification: 35-screen-work-goals

Verdict: **fail**. Verifier: `t35-verifier-1`. Date: 2026-10-22.
Code: `e6c27df4`. `git fetch origin` and `git merge origin/dev` reported that the branch was current.

## Items

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests, AC1 and AC4 | fail | The existing Goals tests pass. An independent test shows that `s` does not remove a goal with no stops. See `verifier/green.txt` and `verifier/stops-only.txt`. |
| V2 Live check | fail | Not run. The brief requires a stop when the code does not agree with the brief. The missing `s` behavior prevents acceptance of AC1 and AC4. No live or visual match is claimed. |
| V-gate Diff scope and gate | pass | `verifier/scope.txt` lists only allowed source files and evidence. `verifier/gate.txt` holds the full gate output. All five commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing. |

## Blocking finding

AC1 states that `s` shows stops only. `goalsScreen.Update` in
`internal/tui/work_goals.go` has no `s` case. The frame gives the key to this
screen. The new tests do not cover this action.

An independent scratch-home test used the existing `goalsQueueFixture`, whose
`g1-demo` goal has a task that needs a user decision. It added a second active
goal, `g2-no-stops`, with no waves or stops. Both goals were visible before the
key. After `s`, `g2-no-stops` was still visible. The test failed on that row.

Command: `go test ./internal/tui/ -run '^TestVerifier35StopsOnly$' -count=1 -v`.
Exit status: 1. Full output: `verifier/stops-only.txt`.

The temporary test was:

```go
func TestVerifier35StopsOnly(t *testing.T) {
    home := goalsQueueFixture(t)
    writeRecord(t, home, []string{"goals"}, "g2-no-stops", records.Goal{
        ID: "g2-no-stops", Project: "demo", Title: "No stops", Status: "active",
    })
    m := goalsModel(t, home, 120, 40)
    requireView(t, m, "g2-no-stops", "g1-demo")
    m, _ = raise(t, m, key('s', "s"))
    requireView(t, m, "g1-demo")
    rejectView(t, m, "g2-no-stops")
}
```

The test file was removed after the run. No source change remains.

## Limits and cleanup

The manager was notified with inbox record 238. Verification stopped under the
brief rule for a code conflict. The base-code runs, deliberate token and key
mutations, live pane capture, PNG capture, and visual table were not run.
Those checks remain required after repair. No scratch pane, tab, workspace,
coordinator, or supervisor was started. The temporary test and scratch folder
were removed. Only verification evidence remains.
