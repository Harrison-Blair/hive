# 81-agents-screen verification

| Item | Value |
| --- | --- |
| Task | 81-agents-screen |
| Check | After Repair 1 |
| Commit verified | f183bd5b (repair code at dcea4944) |
| Base | `origin/dev` = cfc825d6, contained in the task branch |
| Date | 2026-10-20 |
| Agent | t81-verifier-2, claude, claude-opus-5-5 |
| Verdict | pass |

## Results

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Repair red run: temp worktree at de2826f5 (the code before Repair 1) with the task `agents_test.go`. `TestAgentsLoadsLongAgentKeysAsNodesAndParents` fails: `Agents view does not contain "long-manager · manager · demo"`. `TestAgentsPanelFitsTerminalFrameAtStandardAndLargeSizes` fails: `80x24 ... panel has 19 rendered lines; sidebar has 14`, `120x40 ... 35 ... 31`. Base red run: temp worktree at `origin/dev` (cfc825d6) with the task `agents_test.go` and `shell_test.go`: the 10 `TestAgents*` tests except the panel test, and `TestScreenKeysSelectAllEightScreens`, fail. (The panel test does not fail on `dev`, because `dev` has no Agents screen; its red run is the de2826f5 run.) Green run on f183bd5b: all 12 tests `PASS`. |
| V2 | pass | Real home, read only, task binary, temp `XDG_CONFIG_HOME`, herdr workspace `wMH` (`verify-81-r1`). All 7 `manager-stenographer-g1`..`-g7` records show in each tab (name `manager-stenographer`). Live `-g7` shows `○ ... idle · 3m` (agent status `done` shows as idle); `-g1`..`-g6` (stale) show `?`, each with a `┄┄` inferred edge from the coordinator. Tree: live `t58-verifier-2` and `t72-worker-2` (no `parent`) are indented below `-g7` with `┄┄`; after `h`, stopped `t55-verifier-1` (explicit `parent` `manager-stenographer-g7`) is below `-g7` with a solid `├─`. Graph: `-g7` node `▶ ┄┄─ ○ manager-stenographer · claude · 4m ← coordinator`, children below it, `t55-verifier-1` with `├──`. Table: `t55-verifier-1 ... · manager-stenographer` (resolved name, no raw key); `t58-verifier-2`, `t72-worker-2` `· inferred`. Detail of `-g7`: `parent inferred · ... · pane wF7:p1 · age 4m`. `h` shows the dim `×` stopped rows. Arrows: selection moved 40 rows down to `t72-worker-2` and 50 up to `-g7`. `enter`: `PEEK · wF7:p1` with the pane lines. `esc`: back to `[ GRAPH ]`. Frame fit: at the pane size (57 rows) the sidebar and panel bottom borders are on the same row (49, `╯╰`) in tree, graph, table, and peek. At `stty rows 24 cols 80`: both borders on row 16 in tree, table, and peek, and the detail line wraps to 2 rows inside the panel. |
| V3 | pass | Temp `HIVE_HOME` (scratchpad) with one record `scratch-agent` -> `wMJ:p1`, pi started in `wMJ:p1`, task binary in `wMJ:p2`. Before `a`: `wMJ:p1 false pi idle`. After `a`: `wMJ:p1 true pi idle`, `wMJ:p2 false`. |
| V4 | pass | `git diff origin/dev --stat`: brief copy, evidence V1-V5 and VERIFIED.md, `internal/tui/{agents.go,agents_test.go,doc.go,shell.go,shell_test.go,tick.go,tui.go}`, `testdata/ste-baseline.json` (+1). All in the source table. |
| V5 | pass | `gofmt -l` prints nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race ./...` exit 0. |

## Notes

- Repair scope: dcea4944 changes only `agents.go`, `agents_test.go`, and the
  brief. The agent load uses `agent.List(home, agent.ListOptions{All: true})`.
  `View` gives the content `height - 3` rows, and wraps the detail line at the
  panel width. The wrap is part of the fit fix.
- The panel test builds the frame rows with its own copy of the shell layout
  arithmetic. The live check above confirms the fit independently.
- During V2, one `ctrl+c` did not quit the TUI (the shell asks for a second
  `ctrl+c`). The next `herdr pane run` text went into the TUI as one paste,
  and its `enter` opened a peek of `wF7:p1`. No attach occurred: the log has
  no `agent attach` event, and `wF7:p1` stayed `focused: false`. `a` was not
  pressed on a real agent.

## Cleanup

herdr workspaces `wMH` and `wMJ` closed; `herdr workspace list` shows no
`verify-81` workspace. The pi session folder has no new file. Temp worktrees
`red` and `base` removed. By the manager's instruction, the scratchpad files
(temp home, temp `XDG_CONFIG_HOME`, binary, pane captures) are left in place.
They are in the session scratchpad, outside the checkout and the real home.
