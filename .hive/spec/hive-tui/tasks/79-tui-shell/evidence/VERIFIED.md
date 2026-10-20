# VERIFIED — 79-tui-shell

Verifier: `t79-verifier-2`. Check after Repair 1. Task commits: `da97aa89`,
`f164c3c3` (repair). Base: `bf39fc51` (`dev` is still at `bf39fc51`).

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | New tests fail on the base. Each load guard test fails with its guard removed and passes with the guard restored. |
| V2 | pass | All live steps matched the expected result in workspace `wKS`. The workspace is closed. |
| V3 | pass | Only source-table files, the evidence, and `cmd/root_test.go` (Scope change 1, `TestBareHiveRunsTUIUntilQuit` only) changed. |
| V4 | pass | `gofmt -l` printed nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race` exited 0. |

## V1

Red run on the base: I copied the task test files (`internal/tui/*_test.go`,
`internal/lib/herdr/agent_test.go`, `cmd/root_test.go`) into a detached
worktree at `bf39fc51`.

- `internal/lib/herdr` failed to build: `undefined: AgentFocus`.
- `internal/tui` failed to build: `startCurrentScreenLoad`, `headerLoader`,
  `logLoader`, `startHeaderLoad`, `screenIndex`, `headerLoading` undefined.
- `TestBareHiveRunsTUIUntilQuit` failed: the old output has no `⬢ hive`
  header.

Repair 1 guard runs, in a detached scratch worktree at `f164c3c3`, with
`go test ./internal/tui -run 'TestSlow(Header|Screen)LoadDoesNotBlockKeysOrStartTwice' -count=1 -v`:

| Code | `TestSlowHeaderLoad…` | `TestSlowScreenLoad…` |
| --- | --- | --- |
| Task code | PASS | PASS |
| `if m.headerLoading {` changed to `if false {` (`tui.go:257`) | **FAIL** `shell_test.go:92: refresh started a second header load before the first settled` | PASS |
| `m.screenLoading \|\|` removed from the guard (`tui.go:269`) | PASS | **FAIL** `shell_test.go:138: refresh started a second screen load before the first settled` |
| Both guards restored | PASS | PASS |

The restored tests also passed 20 times with `-race -count=20`.

The repaired header test now runs the data commands of the second tick batch
(`runRefreshLoads`) while the first load is blocked. The new screen test does
the same for the Logs screen loader. The repair commit changed only
`internal/tui/shell_test.go`, the brief copy, and the evidence.

Green run on the task branch: all tests pass (see V4).

## V2

herdr 0.9.5. I built the task binary at `f164c3c3` into a scratch folder. The
temp `HIVE_HOME` had `.hive/README.md` and two agent records with status `live`
(`hive agent list` showed `worker-a1` and `worker-a2`). `XDG_CONFIG_HOME` was a
separate temp folder. I made workspace `wKS` (label `79-verify-live-2`) with
`herdr workspace create --no-focus` and ran the binary in pane `wKS:p1`. I read
the pane with `herdr pane read --source visible` after each step.

- Start: header `⬢ hive · 0 projects · 2 live agents · working unknown / idle unknown · supervisor stale · needs_user 0 · 01:29`. Sidebar `▶ 1 Chat`, panel `Chat · next wave`.
- Keys `2`–`7`: the sidebar marker moved to `▶ 2 Agents` … `▶ 7 Health`; the panel showed `Agents · next wave` … `Health · next wave`.
- Key `8`: `▶ 8 Logs`, the panel showed log lines, the footer showed `1–8 screens · / filter · l level · f follow · g/G scroll · ctrl+c×2 quit`.
- Key `q`: Logs stayed selected, the header kept updating, the TUI kept running.
- `ctrl+c` once (`send-keys C-c`): the footer showed `press ctrl+c again to quit`.
- After 4 seconds: the notice was gone, the TUI still ran.
- `ctrl+c` two times: the first showed the notice, the second exited to the shell prompt `[penguin@iceberg home]$`.
- Narrow: I started the binary again, split the pane right with ratio 0.45, and pressed `8`. The widest line of the read was 91 columns. The header wrapped onto two lines, the sidebar showed numbers and names in 24 columns, the panel wrapped log lines inside its border, and the footer showed the hints. No line went past the panel border.

Note: `herdr pane send-keys PANE ctrl+c` sent no `ctrl+c` to the TUI; `C-c`
worked. This is a herdr key-name detail and not a task fault.

Cleanup: I quit the TUI with two `ctrl+c` and ran `herdr workspace close wKS`.
`herdr workspace list` then showed no `79-verify-live` workspace. I removed
the temp home, config, binary, and both scratch worktrees.

## V3

`git diff dev --stat`: `.hive/spec/hive-tui/tasks/79-tui-shell/brief.md`, the
evidence files, `cmd/root_test.go`, `internal/lib/herdr/agent.go`,
`internal/lib/herdr/agent_test.go`, `internal/tui/*` (doc, header, logview,
shell, tick, tui, and tests), `testdata/ste-baseline.json`. The
`cmd/root_test.go` diff changes only `TestBareHiveRunsTUIUntilQuit`: input
`q` to `\x03\x03`, and the home assertion to the `⬢ hive` header assertion.

## V4

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')   # no output
go vet ./...                                             # exit 0
golangci-lint run ./...                                  # 0 issues., exit 0
go test -race ./...                                      # all ok, exit 0
```
