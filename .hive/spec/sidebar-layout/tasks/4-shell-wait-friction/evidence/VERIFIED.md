# 4-shell-wait-friction VERIFIED

| Item | Value |
| --- | --- |
| Task | 4-shell-wait-friction |
| Commit | 223b279b |
| Date | 2026-10-19 |
| Agent | t4-verifier-3, Claude Code, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | Three manager spawns in a row give `ok` on the first attempt. The cwd of each workspace is the scratch path. The log shows the shell wait before each `agent.start` and no `agent_pane_busy`. |
| V2 | AC2 | pass | The diff has the R1 change and one entry for each of the five R2 items, and no other change. |
| V3 | AC3 | pass | `gofmt -l` prints nothing. `go vet`, `golangci-lint`, and `go test -race` pass. |

## V1 Live check

Run from the worktree with `--home "$PWD"` and the binary from `go build .`.
Scratch repository: `.hive/tmp/t4v3-live`, one commit. Project ID: `t4v3`.
Harness `pi`, model `opencode-go/deepseek-v4.1-flash`.

| Spawn | Name | Outcome | Workspace | Pane cwd | Label | Stop |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | `t4v3-mgr-1` | ok, rc 0 | `wC2` | `<worktree>/.hive/tmp/t4v3-live` | `hive:t4v3` | ok, pane and workspace closed |
| 2 | `t4v3-mgr-2` | ok, rc 0 | `wC3` | `<worktree>/.hive/tmp/t4v3-live` | `hive:t4v3` | ok, pane and workspace closed |
| 3 | `t4v3-mgr-3` | ok, rc 0 | `wC4` | `<worktree>/.hive/tmp/t4v3-live` | `hive:t4v3` | ok, pane and workspace closed |

The log `.hive/logs/2026-10-19.jsonl` shows this order of herdr requests for
each spawn (the wait is the `pane.get` and `pane.process_info` poll):

```text
spawn 1: workspace.create 09:01:56.084, 2 x pane.process_info, agent.start 09:01:56.298
spawn 2: workspace.create 09:02:00.661, 3 x pane.process_info, agent.start 09:02:01.098
spawn 3: workspace.create 09:02:04.474, 2 x pane.process_info, agent.start 09:02:04.783
```

`grep -c agent_pane_busy .hive/logs/2026-10-19.jsonl` printed `0`.

Cleanup: `herdr workspace list` after the run shows only `wBC wBP wBX wC1`,
the workspaces from before the run. The run removed the scratch repository,
`.hive/state/` (project, agent, and claim records), `projects/t4v3`,
`overrides/presets.md`, and the `hive` binary. The permission classifier of
the verifier harness blocked the removal of `.hive/logs/2026-10-19.jsonl`.
That ignored log file stays in the worktree.

## V2 Friction review

`git diff origin/dev -- reference/dogfood/friction.md`:

- R1: the status of "Manager spawn fails with `agent_pane_busy` right after
  `workspace.create`" changes from Open to Resolved 2026-10-19. The summary
  adds the cause: the shell runs commands in its own process group during
  rc-file startup, so a group ID equal to the shell PID is not sufficient.
  Commit `aa29b8d9` also requires one foreground process, the shell. The
  commit is in the summary of the entry, not in the Status line.
- R2: five new entries, one for each backlog item except the ready rule:
  no `hive task edit --brief` (Open), exited name blocks reuse (Open),
  verifier `--pane` requires `--project` (Open), shell-ready race in worker
  and verifier spawns (Resolved 2026-10-19, `aa29b8d9` in the summary), and
  verifier `--pane` skips the first prompt (Open). Each entry uses
  Issue / Status / Summary / Reproduction steps.
- No other change in the file. The branch changes only this file and
  `evidence/4-shell-wait-friction/`.

## V3 Gate

```text
== gofmt
rc=0            (no output)
== vet
rc=0
== lint
0 issues.
rc=0
== test
ok for each package with tests, rc=0
```
