# 5-tab-labels VERIFIED

| Item | Value |
| --- | --- |
| Task | 5-tab-labels |
| Commit | 7121ac04 (branch head before this file: ea2f7c55) |
| Date | 2026-10-19 |
| Agent | t5-verifier, verifier, pi, opencode-go/deepseek-v4.1-flash |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | Green: 2 tests, 5 subtests pass. Red: a disabled rename failed worker, verifier, and manager-root; a forced rename failed manager-hive-tab and explicit-pane. Each mutation restored. |
| V2 | AC2, AC3 | pass | Green: 3 tests pass. Red: calling `Register` from the command failed both register tests; removing the failed-effect action failed both failure tests. Each mutation restored. |
| V3 | AC4 | pass | One `TabRename` definition and one caller. The only `"tab.rename"` request is in that function. |
| V4 | AC5 | pass | Scratch tab `wBC:tA` became `supervisor`. The worktree home holds the claim and record; the parent home did not change. Scratch pane `wBC:pB` closed; `wBC` kept `t1` and `t2`. |
| V5 | AC6 | pass | `gofmt -l` printed nothing. `go vet`, `golangci-lint`, and `go test -race ./...` passed. |

The wrapper check requested by the manager: `cmd/agent/register/register.go:44` calls
`agentlib.RegisterAndRename`. `internal/agent/spawn.go:122` calls `Register`, the
store-only function. The only non-test caller of `RegisterAndRename` is the
register command. The spawn green tests record exactly one `tab.rename`
request. A forced second rename in `Spawn` produced two `tab ... renamed`
effects, so the one-rename test detects it.

## Runs

All runs are from `/home/penguin/source/hive/projects/hive/worktrees/5-tab-labels`.
Quoted test output keeps the failing lines and drops repeated PASS lines.

### V1 green

```sh
go test ./internal/agent/ -run '^TestSpawnRenames(WorkerAndVerifierTabs|HiveCreatedManagerRootAndSkipsHiveTabAndPane)$' -v -count=1
```

```text
--- PASS: TestSpawnRenamesWorkerAndVerifierTabs (0.02s)
    --- PASS: TestSpawnRenamesWorkerAndVerifierTabs/worker (0.01s)
    --- PASS: TestSpawnRenamesWorkerAndVerifierTabs/verifier (0.01s)
--- PASS: TestSpawnRenamesHiveCreatedManagerRootAndSkipsHiveTabAndPane (0.00s)
    --- PASS: TestSpawnRenamesHiveCreatedManagerRootAndSkipsHiveTabAndPane/manager_root_workspace (0.00s)
    --- PASS: TestSpawnRenamesHiveCreatedManagerRootAndSkipsHiveTabAndPane/manager_hive_tab (0.00s)
    --- PASS: TestSpawnRenamesHiveCreatedManagerRootAndSkipsHiveTabAndPane/explicit_pane (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	0.023s
exit 0
```

### V1 red

Mutation 1: `internal/agent/spawn.go:189`, `if renameTab {` became `if renameTab && false {`.

```text
--- FAIL: TestSpawnRenamesWorkerAndVerifierTabs (0.01s)
    tab_labels_test.go:77: tab.rename requests = 0; want 1
    --- FAIL: TestSpawnRenamesWorkerAndVerifierTabs/worker (0.01s)
    --- FAIL: TestSpawnRenamesWorkerAndVerifierTabs/verifier (0.01s)
--- FAIL: TestSpawnRenamesHiveCreatedManagerRootAndSkipsHiveTabAndPane (0.00s)
    tab_labels_test.go:97: tab.rename requests = 0; want 1
    --- FAIL: .../manager_root_workspace (0.00s)
    --- PASS: .../manager_hive_tab (0.00s)
    --- PASS: .../explicit_pane (0.00s)
FAIL
exit 1
```

Mutation 2: the condition became `if renameTab || true {`.

```text
--- PASS: TestSpawnRenamesWorkerAndVerifierTabs (0.01s)
--- FAIL: TestSpawnRenamesHiveCreatedManagerRootAndSkipsHiveTabAndPane (0.00s)
    tab_labels_test.go:116: tab.rename requests = 1; want 0
    tab_labels_test.go:128: tab.rename requests = 1; want 0
    --- PASS: .../manager_root_workspace (0.00s)
    --- FAIL: .../manager_hive_tab (0.00s)
    --- FAIL: .../explicit_pane (0.00s)
FAIL
exit 1
```

`git checkout -- internal/agent/spawn.go` restored the file. `git status
--porcelain` was empty. The green run above passed again after each restore.

### V2 green

```sh
go test ./internal/agent/ ./cmd/agent/register/ -run '^(TestSpawnRenameFailureIsNonFatalAndLogged|TestRegisterCommandRenamesRoleAndProjectLabels|TestRegisterCommandRenameFailureIsAnEffectAndKeepsSuccess)$' -v -count=1
```

```text
--- PASS: TestSpawnRenameFailureIsNonFatalAndLogged (0.00s)
--- PASS: TestRegisterCommandRenamesRoleAndProjectLabels (0.00s)
    --- PASS: .../coordinator_uses_explicit_pane_tab (0.00s)
    --- PASS: .../supervisor_uses_role_label (0.00s)
    --- PASS: .../project_agent_uses_agent_name (0.00s)
--- PASS: TestRegisterCommandRenameFailureIsAnEffectAndKeepsSuccess (0.00s)
PASS
exit 0
```

### V2 red

Mutation 1: `cmd/agent/register/register.go:44` called `Register` in place of
`RegisterAndRename`.

```text
--- PASS: TestSpawnRenameFailureIsNonFatalAndLogged (0.00s)
--- FAIL: TestRegisterCommandRenamesRoleAndProjectLabels (0.00s)
    register_test.go:219: pane.get = []register_test.registerRequest(nil); want pane w2:p9
    register_test.go:219: pane.get = []register_test.registerRequest(nil); want pane w3:p3
    register_test.go:219: pane.get = []register_test.registerRequest(nil); want pane w4:p4
--- FAIL: TestRegisterCommandRenameFailureIsAnEffectAndKeepsSuccess (0.00s)
    register_test.go:259: effects = [...]; want failed tab rename
FAIL
exit 1
```

Mutation 2: `renameAgentTab` in `internal/agent/register.go` no longer set the
action to `rename failed`.

```text
--- FAIL: TestSpawnRenameFailureIsNonFatalAndLogged (0.00s)
    tab_labels_test.go:147: effects = [..., {Resource:"tab", Path:"wBD:t1", Action:"renamed"}, ...]; want failed tab rename
--- PASS: TestRegisterCommandRenamesRoleAndProjectLabels (0.00s)
--- FAIL: TestRegisterCommandRenameFailureIsAnEffectAndKeepsSuccess (0.00s)
    register_test.go:259: effects = [..., {Resource:"tab", Path:"w2:t9", Action:"renamed"}]; want failed tab rename
FAIL
exit 1
```

Mutation 3, the one-rename check: `Spawn` called `RegisterAndRename` in place of
`Register`.

```text
--- FAIL: TestSpawnRenamesWorkerAndVerifierTabs/worker (0.01s)
    fakeherdr_test.go:147: unexpected method "pane.get"
    ... Effects:[..., {Resource:"tab", Path:"w5:t1", Action:"renamed"}, ...,
                     {Resource:"tab", Path:"w5:t1", Action:"renamed"}, ...] ...
--- FAIL: TestSpawnRenamesWorkerAndVerifierTabs/verifier (0.01s)
    ... Effects:[..., {Resource:"tab", Path:"w5:t2", Action:"renamed"}, ...,
                     {Resource:"tab", Path:"w5:t2", Action:"renamed"}, ...] ...
FAIL
exit 1
```

The effects held two `renamed` entries for each spawn, so a second rename is
detectable. The failure also reports a consumed `pane.get` reply. `git checkout`
restored each file, and the V2 green run passed again.

### V3

```sh
grep -rn "TabRename" internal/ --include='*.go' --exclude='*_test.go'
```

```text
internal/agent/register.go:139:	err := herdr.TabRename(ctx, tabID, label, herdr.WithLogger(logger))
internal/lib/herdr/tab.go:11:func TabRename(ctx context.Context, tabID, label string, opts ...Option) error {
exit 0
```

```sh
grep -rn '"tab\.rename"' --include='*.go' --exclude='*_test.go' internal/ cmd/
```

```text
internal/lib/herdr/tab.go:20:	if err := Call(ctx, "tab.rename", params, &result, opts...); err != nil {
```

```sh
grep -rn "RegisterAndRename" --include='*.go' --exclude='*_test.go' .
```

```text
./cmd/agent/register/register.go:44:			envelope := agentlib.RegisterAndRename(home, agentlib.RegisterOptions{
./internal/agent/register.go:107:// RegisterAndRename registers one agent and best-effort labels its tab.
./internal/agent/register.go:108:func RegisterAndRename(home string, options RegisterOptions) output.Envelope {
```

### V4

Before the check, `wBC` held `wBC:t1` (label `1`) and `wBC:t2` (label `2`).
I built `/tmp/hive-5-tab-labels-verify` from the worktree.

```sh
herdr tab create --workspace wBC --cwd "$PWD"
```

```text
{"result":{"root_pane":{"pane_id":"wBC:pB",...},"tab":{"label":"3","number":10,"pane_count":1,"tab_id":"wBC:tA","workspace_id":"wBC"},"type":"tab_created"}}
exit 0
```

The shell was ready after one `pane process-info` try: foreground `/usr/bin/bash`,
pid 351139. Then:

```sh
herdr pane run wBC:pB /tmp/hive-5-tab-labels-verify --home /home/penguin/source/hive/projects/hive/worktrees/5-tab-labels agent register --role supervisor --pane wBC:pB
herdr pane read wBC:pB --source recent --lines 20
```

```text
stale agent .hive/state/agents/supervisor-g2.json
created agent .hive/state/agents/supervisor-g3.json
updated claim .hive/state/claims/supervisor.json
renamed tab wBC:tA
exit 0
```

```sh
herdr tab list --workspace wBC
```

```text
wBC:t1 label "1" | wBC:t2 label "2" | wBC:tA label "supervisor"
```

Worktree home:

```text
worktree agent: key=supervisor-g3 name=supervisor role=supervisor status=live pane_id=wBC:pB
worktree claim: key=supervisor agent=supervisor-g3 generation=3
worktree supervisor agents: supervisor-g1.json, supervisor-g2.json, supervisor-g3.json
```

Parent home:

```text
parent claim: agent=supervisor-g2 generation=2
parent supervisor agents: supervisor-g1.json, supervisor-g2.json
parent claim mtime unchanged: 2026-10-18 22:59:33 -0400
```

Cleanup:

```sh
herdr pane close wBC:pB
herdr tab list --workspace wBC
herdr workspace list
rm -f /tmp/hive-5-tab-labels-verify /tmp/hive-5-tab-labels
```

```text
{"result":{"type":"ok"}}
wBC:t1 label "1" | wBC:t2 label "2"
workspace wBC label "hive" still exists
exit 0
```

No step sent `workspace.close`. The only pane this check closed is `wBC:pB`,
which the check made. The register left `supervisor-g3` and claim generation 3
in the worktree home; `supervisor-g2` became stale. The worktree home is
gitignored runtime state, and the worker V4 already left `g2` and the claim.

### V5

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
```

```text
(no output)
exit 0
```

```sh
go vet ./...
```

```text
exit 0
```

```sh
golangci-lint run ./...
```

```text
0 issues.
exit 0
```

```sh
go test -race ./...
```

```text
(all packages ok; no FAIL, panic, or DATA RACE)
exit 0
```

A fresh run of the changed packages passed as well:

```text
ok  	github.com/Harrison-Blair/hive/internal/agent	2.767s
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.016s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.567s
```

## Notes

- `internal/lib/herdr/tab.go:10` has the comment `// TabListParams names an
  optional workspace for tab.list.` above `TabRename`. That line belongs to
  `TabListParams`. It is a copied comment and does not change behavior. No
  verification item covers it.
- The register failure test checks the failed effect, the exit status 0, and the
  live record. The shared `renameAgentTab` log is asserted by
  `TestSpawnRenameFailureIsNonFatalAndLogged`.
