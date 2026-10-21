# VERIFIED

Result: pass.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | Each red test failed on `dev` for the missing behavior and passed on the task branch. |
| V2 Brief copy and ratchet | pass | `go test ./ -run 'STE|Skill'` passed. |
| V3 Live check | pass | Live restarts, pane-gone restart, old-binary header, retained records, history toggle, and readiness checks passed. |
| V4 Diff scope | pass | `dev` was up to date. The diff contained only paths in the brief's source table or evidence folder. |
| V5 Gate | pass | All five gate commands passed. |

## V1 Tests

I ran each test in a detached worktree at `dev`, then in this task worktree.

| Test | `dev` result | Task result |
| --- | --- | --- |
| `TestAgentsRestartConfirmsLiveCoordinatorAndCancelsOnOtherKey` | Failed: no coordinator pane check started. | Passed. |
| `TestStartCoordinatorWaitsForHerdrReadiness` | Failed: zero `agent.get` requests. | Passed. |
| `TestAgentGetPreservesReadinessFields` | Failed: `launch_pending` was missing. | Passed. |

## V2 Brief copy and ratchet

Command: `go test ./ -run 'STE|Skill'`

Result: `ok github.com/Harrison-Blair/hive`.

## V3 Live check

I used the isolated Herdr session `hive-verifier-3-tui-restart`, workspace `w2`, a temporary `HIVE_HOME`, and a temporary `XDG_CONFIG_HOME` with the low-cost coordinator model `opencode-go/deepseek-v4.1-flash`. An initial attempt in the default Herdr session met an existing coordinator name. I moved the check to the isolated session and completed it there.

I made the installed test binary newer than a separate supervisor executable. The Health check reported `supervisor binary is older`. The Agents header showed `supervisor old · r to restart`. The same plain-text mark appeared at 120x40 and 200x57.

For the live supervisor, `r` showed `restart supervisor? y/n`. After `y`, Herdr closed pane `w2:p5` and Hive opened supervisor tab `w2:t6`, pane `w2:p6`. Pane process information showed `hive supervise`. The old `supervisor-g3` record became stale. `supervisor-g4` became live. The old-binary mark cleared.

For the live coordinator, `r` showed `restart coordinator? y/n`. After `y`, Hive closed pane `w2:p2` and opened coordinator pane `w2:p7`. The old coordinator record became stale. I then closed `w2:p7` by hand. Before the next `r`, the current record still had status `live`. That `r` showed `restarting` and did not ask for confirmation. Hive opened coordinator pane `w2:p8`; the old record became stale and the new record became live.

I toggled `h` to show the old generations, then toggled it again to hide them. The final records were:

| Record | Status | Pane | File kept |
| --- | --- | --- | --- |
| `coordinator-g1` | stale | `w2:p2` | yes |
| `coordinator-g2` | stale | `w2:p7` | yes |
| `coordinator-g3` | stale | `w2:p7` | yes |
| `coordinator-g4` | live | `w2:p8` | yes |
| `supervisor-g1` | stale | `w5Q:p3` | yes |
| `supervisor-g2` | stale | `w2:p3` | yes |
| `supervisor-g3` | stale | `w2:p5` | yes |
| `supervisor-g4` | live | `w2:p6` | yes |

The readiness log showed `outcome=ready` before one prompt for each restarted coordinator pane, `w2:p7` and `w2:p8`.

The scratch coordinator also showed a policy error when it tried to send chat from below `projects/`. This error did not block the restart checks. I closed the scratch agents, workspace, and session after capture. I removed the temporary worktree and scratch files.

### Visual check

The brief has no approved board. I stripped ANSI styling and checked each state word in the captured screen text. Color did not carry a state by itself.

| Screen state | State without color | Result | Cause or decision |
| --- | --- | --- | --- |
| Old supervisor, 120x40 | `supervisor old · r to restart` | matches | AC5 |
| Old supervisor, 200x57 | `supervisor old · r to restart` | matches | AC5 |
| Live supervisor confirmation | `restart supervisor? y/n` | matches | AC1, AC3 |
| Supervisor result | `restarted supervisor` | matches | AC3, AC4 |
| Live coordinator confirmation | `restart coordinator? y/n` | matches | AC1, AC2 |
| Pane-gone coordinator restart | `restarting`; no confirmation | matches | AC1 |
| Coordinator result | `restarted coordinator` | matches | AC2, AC4 |
| Agent history | Old rows appeared after `h` and hid after the next `h` | matches | AC4 |

Screenshots captured from the isolated panes:

- `verified-agents-120x40-old.png`
- `verified-agents-200x57-old.png`
- `verified-agents-120x40-supervisor-confirm.png`
- `verified-agents-120x40-supervisor-result.png`
- `verified-agents-120x40-coordinator-confirm.png`
- `verified-agents-120x40-coordinator-result.png`
- `verified-agents-120x40-coordinator-stale-restarting.png`
- `verified-agents-120x40-history-visible.png`

## V4 Diff scope

`git merge dev` reported `Already up to date.` I ran `git diff dev...HEAD --stat`. All changed tracked paths were in the brief's source table or its evidence folder. The new verifier screenshots and this report are also in the evidence folder.

## V5 Gate

I ran the gate from the worktree root:

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

`gofmt -l` and `go mod tidy -diff` printed nothing. `go vet ./...` passed. `golangci-lint run ./...` reported `0 issues`. `go test -race ./...` passed.
