# VERIFIED — 79-tui-shell

Verifier: `t79-verifier-1`. Task commit: `da97aa89`. Base: `bf39fc51`.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | fail | The R3 test for "does not start twice" does not detect the fault. See below. |
| V2 | pass | All live steps matched the expected result. |
| V3 | pass | Only source-table files, the evidence, and `cmd/root_test.go` (Scope change 1) changed. |
| V4 | pass | `gofmt -l` printed nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race` exited 0. |

## V1

Red run: I copied the task test files (`internal/tui/*_test.go`,
`internal/lib/herdr/agent_test.go`, `cmd/root_test.go`) into a detached
worktree at `bf39fc51`. `internal/tui` and `internal/lib/herdr` failed to
build (`AgentFocus`, `startHeaderLoad`, `screenIndex`, `attachCmd`,
`quitNoticeExpiredMsg` undefined). `TestBareHiveRunsTUIUntilQuit` failed
because the old output has no `⬢ hive` header.

Green run: on the task branch, all tests pass (see V4).

Because the base fails to build, I also broke the task code in a scratch copy
of `da97aa89` and ran `go test ./internal/tui ./internal/lib/herdr`:

| Mutation | Tests that failed |
| --- | --- |
| One `ctrl+c` quits | `TestQuitNeedsTwoCtrlCWithinThreeSeconds`, `TestQuitNoticeExpiresAndOtherKeyClearsIt`, `TestKeysFilterLevelFollowAndQuit`, `TestQuitKeysRequireTwoCtrlCPresses` |
| `q` quits in Logs | `TestQDoesNotQuitOnAnyScreen`, `TestKeysFilterLevelFollowAndQuit`, `TestQuitKeysRequireTwoCtrlCPresses` |
| `AgentFocus` calls `agent.get` | `TestAgentFocusRequestAndReply` |
| Other key does not clear the notice | `TestQuitNoticeExpiresAndOtherKeyClearsIt` |
| Wide sidebar at all widths | four log-view tests |
| Key `8` selects screen 1 | `TestScreenKeysSelectAllEightScreens` |
| Attach error is dropped | `TestAttachUsesBackgroundFocusAndReportsError` |
| **`if m.headerLoading {` changed to `if false {`** | **none** — `TestSlowHeaderLoadDoesNotBlockKeysOrStartTwice` passes |
| **`m.screenLoading` guard removed** | **none** |

Cause: `TestSlowHeaderLoadDoesNotBlockKeysOrStartTwice` (`internal/tui/shell_test.go:54`)
sends `tickMsg` but never runs the returned command. The second load can
therefore never call the loader, and the check on `started` passes with or
without the guard. No test covers the screen-load guard. R6 requires a test
that a slow load does not start twice. AGENTS.md says a test counts only when
it fails when the behavior breaks.

Repair: make the test fail when the guard is absent. For example, assert that
`startHeaderLoad()` returns nil while a load is in flight, or run the tick
command (including its batch) and confirm that the loader starts once. Add the
same check for the screen load.

## V2

herdr 0.9.5. I built the task binary into a scratch folder. The temp
`HIVE_HOME` had `.hive/README.md` and two agent records with status `live`
(`hive agent list` showed `worker-a1` and `worker-a2`). `XDG_CONFIG_HOME` was a
separate temp folder. I made workspace `wKP` (label `79-verify-live`) with
`herdr workspace create --no-focus` and ran the binary in pane `wKP:p1`
(207x57). I read the pane with `herdr pane read` after each step.

- Start: header `⬢ hive · 0 projects · 2 live agents · working unknown / idle unknown · supervisor stale · needs_user 0 · 01:21`. Sidebar `▶ 1 Chat`, panel `Chat · next wave`.
- Keys `2`–`7`: sidebar marker and panel title moved to Agents, Goals, Queue, Tasks, Friction, Health, each with `next wave`.
- Key `8`: `▶ 8 Logs`, panel showed log lines, footer showed the log key hints.
- Key `q`: Logs stayed selected, TUI still running.
- `ctrl+c` once: footer showed `press ctrl+c again to quit`.
- After 4 seconds: notice gone, TUI still running.
- `ctrl+c` twice: first showed the notice, second exited to the shell prompt (pane process `bash` only).
- Narrow: I started the binary again, split the pane with ratio 0.45 (pane `wKP:p1` 93x57), pressed `8`. The header wrapped onto two lines, the sidebar showed numbers and names in 24 columns, the panel wrapped log lines inside its border, and the footer showed the hints. No line exceeded the pane.

Cleanup: quit the TUI, `herdr workspace close wKP`. `herdr workspace list`
then showed no `79-verify-live` workspace. I removed the temp home, config,
binary, and both scratch worktrees.

## Notes (not verdict items)

- The old test `TestHeaderPersistentErrorDoesNotLogEachTick` was replaced by
  `TestHeaderErrorDoesNotStopShell`. With the new 2-second refresh, a
  persistent header error now writes one `tui error` log event per refresh.
- The Logs screen now reads all log files from the start on each 2-second
  refresh and de-duplicates by path and line, in place of the follower.
