# 81-agents-screen verification

| Item | Value |
| --- | --- |
| Task | 81-agents-screen |
| Commit verified | db05d499 (code at f6e7e00b) |
| Base | `dev` = `origin/dev` = edfb4c6f |
| Date | 2026-10-20 |
| Agent | t81-verifier-1, claude, claude-opus-5-5 |
| Verdict | fail (V2) |

## Results

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Red run on a temp worktree at `dev` (edfb4c6f) with the task `agents_test.go` and `shell_test.go`: each of the 9 new `TestAgents*` tests and `TestScreenKeysSelectAllEightScreens` fails on an assertion (not on a build error). `TestSlowHeaderLoadDoesNotBlockKeysOrStartTwice` is an existing test that the task adapted; it passes on both. Green run on the task branch: all 11 pass. |
| V2 | **fail** | The Agents screen omits each agent record whose key is longer than 18 characters. On the real home, the live `manager-stenographer-g7` (and `-g1`..`-g6`) is not in any tab, while the header counts it in "9 live agents". Consequences: in the stenographer group `t55-verifier-1` shows as a root with the raw parent key `manager-stenographer-g7` in the table; `t58-verifier-2` and `t72-worker-2` show `parent inferred` with no manager node to connect to. See "V2 defect" below. Other steps worked: `2`, `tab` x2, `h` (shows 16 dim `×` stopped rows, `h` again hides them), arrows and `j`/`k`, `enter` (peek of live `wF8:p1` shows its last lines; peek of a stale pane shows `pane read: ... exit status 1`), `esc` (back to `[ TREE ]`). |
| V3 | pass | Temp `HIVE_HOME` with one record `scratch-agent` -> `wMD:p1`, pi started in `wMD:p1`. Before `a`: `wMD:p1 focused False`. After `a`: `wMD:p1 focused True pi`. Note: with a plain shell (no herdr agent) in the pane, `a` shows `error: agent_not_found: agent target wMD:p1 not found`; this is the existing shell attach, not new to this task. |
| V4 | pass | `git diff dev --stat`: brief copy, evidence V1-V5, `internal/tui/{agents.go,agents_test.go,doc.go,shell.go,shell_test.go,tick.go,tui.go}`, `testdata/ste-baseline.json` (+1 line, brief copy at 0). All inside the source table. |
| V5 | pass | `gofmt -l` prints nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.` exit 0. `go test -race ./...` exit 0, every package `ok`. |
| Extra: merge with `origin/dev` | pass | Temp worktree in the scratchpad at HEAD, `git fetch origin dev`, `git merge --no-edit origin/dev`: `Already up to date.` (the branch already contains edfb4c6f, the Chat screen merge). `go build .` ok. `go test -race -count=1 ./internal/tui/ ./cmd/...` exit 0. Worktree removed. |

## V2 defect

Cause: `loadAgents` in `internal/tui/agents.go:146` reads agents with
`store.List`, and `store.List` (`internal/lib/store/store.go:130`) skips each
file whose ID fails `records.ValidID` (max 18 characters). Agent keys are not
project IDs. The header uses `agent.List` and counts these agents.

Reproduction, against the real home, with a scratch test that calls
`loadAgents("/home/penguin/source/hive", time.Now())`:

```text
GROUP 0 home 19
GROUP 1 hive 116
GROUP 2 qmeter 22
GROUP 3 stenographer 50
manager-hive-g9 live parent= coordinator-g6 true group= 1
manager-qmeter-g3 live parent= coordinator-g6 true group= 2
t72-worker-2 live parent=  true group= 3
```

207 nodes from 214 files in `.hive/state/agents/`. The 7 missing keys are
exactly the keys longer than 18 characters: `manager-stenographer-g1` ..
`manager-stenographer-g7`. No `manager-stenographer` node exists, so the
inferred parent of `t72-worker-2` is empty.

Table tab excerpt (live):

```text
t55-verifier-1 · verifier · stenographer · 55-asr-bench · claude · claude-opus-5-5 · idle · 6m · manager-stenographer-g7
t58-verifier-2 · verifier · stenographer · 58-diarization · claude · claude-opus-5-5 · idle · ? · inferred
▶ t72-worker-2 · worker · stenographer · 72-tui-input-first · pi · openai-codex/gpt-6-luna:xhigh · idle · ? · inferred
```

## Secondary observation (not a verdict item)

When the graph content fills the panel, the Agents panel is 3 rows taller
than the sidebar: the sidebar border `╰` is on row 49 and the panel border on
row 52 of the pane, and the last cross-group edge lines render below the
sidebar. `agentsScreen.View` returns `height` lines, but `renderPanel` adds a
title row and a border, so the content must fit in `height - 3` rows (as
`tui.go` does for Logs with `contentHeight := bodyHeight - 3`).

## Cleanup

herdr workspaces `wMC` (V2) and `wMD` (V3) closed; `herdr workspace list`
shows no `verify-81` workspace. Temp `HIVE_HOME`, temp `XDG_CONFIG_HOME`, the
pi session file and its new directory removed. Temp worktrees `base`,
`merge`, `dbg` removed; `git worktree list` shows none of them.
