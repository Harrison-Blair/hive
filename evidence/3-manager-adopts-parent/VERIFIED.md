# 3-manager-adopts-parent VERIFIED

| Item | Value |
| --- | --- |
| Task | 3-manager-adopts-parent |
| Commit | 88004bf0 |
| Date | 2026-10-19 |
| Agent | t3-verifier-3, verifier, claude, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1, AC2, AC3, AC4 | pass | Green: 8 resolution tests and `TestWorktreeListRequestAndReply` pass. Red: I applied 8 mutations of my own. Each mutation failed the test of its criterion with a request-list or record assertion. |
| V2 | AC5 | pass | Green: 2 tests pass. Red: when I removed the `promptManagerInstructions` call, both tests failed with `agent.prompt requests = 0; want 1`. |
| V3 | AC6, AC6b | pass | Green: 4 tests pass. Red: 4 mutations of my own. Each mutation failed the test of its case. |
| V4 | AC7 | pass | The search gives exactly the AC7 list. The resolver, `WorktreeList`, the wait, and `AgentStart` each have one caller. |
| V5 | AC8 | pass | `wBC` is the parent. The new tab `wBC:t7` has the label `t3-check-manager` and holds pane `wBC:p8`. The pi session got one user message, and that message is the generated instructions. After stop, `wBC` keeps `t1` and `t2`. |
| V6 | AC9 | pass | The four gate commands pass. `gofmt -l` prints nothing. The STE baseline for `AGENTS.md` and the template is 0, and this branch does not change it. The spec is not in `steFiles`. |
| V7 | AC10, AC11 | pass | Green: 4 tests pass. Red: 2 mutations of my own. The fallback retry was not added because V5 passed, and the brief permits that. |

The manager asked for a check of commit `eafd7239`. That commit changes one line in `cmd/agent/spawn/cli_test.go`. It replaces the default `pane.process_info` reply of the fake, `foreground_processes: []`, with one entry: the shell, pid 1, `bash`. The old reply is not ready under the 2026-10-19 rule, and the new reply is ready. No assertion changed and no test was removed. Thus the change does not weaken a test.

## Runs

All runs are from the worktree root at commit `88004bf0`.

### V1, V2, V3, V7 green

```sh
go test ./internal/agent/ -count=1 -v -run '<the 18 test names of V1, V2, V3, V7>'
go test ./internal/lib/herdr/ -run TestWorktreeListRequestAndReply -count=1
```

```text
--- PASS: TestManagerAdoptsParentWorkspaceAndCreatesNamedTab (0.00s)
--- PASS: TestManagerCreatesWorkspaceWhenNoParentExists (0.00s)
--- PASS: TestManagerReusesRecordedFreeShellTab (0.00s)
--- PASS: TestManagerCreatesNewTabWhenRecordedTabIsGone (0.00s)
--- PASS: TestManagerClearsOldLayoutRecordWithoutClosingWorkspace (0.00s)
--- PASS: TestManagerAdoptFlagNeedsParentWorkspaceButKeepsItsLabel (0.00s)
--- PASS: TestManagerAdoptFlagRejectsDifferentWorkspace (0.00s)
--- PASS: TestManagerNotRepositoryReportsWorktreeListError (0.00s)
--- PASS: TestManagerPromptSendsGeneratedInstructionsAfterReady (0.00s)
--- PASS: TestManagerPromptLogCapsLargeBody (0.00s)
--- PASS: TestStopAdoptedManagerDropsClosedTabRecord (0.00s)
--- PASS: TestStopConfirmationRequiredKeepsStoppedRecordAndWorkspace (0.00s)
--- PASS: TestManagerStartFailureCleansOnlyCreatedAdoptedTab (0.00s)
--- PASS: TestManagerSpawnWaitsForShellSecondRead (0.20s)
--- PASS: TestManagerSpawnWaitsForShellStartupProcessesToClear (0.20s)
--- PASS: TestWorkerSpawnWaitsForShellSecondRead (0.21s)
--- PASS: TestManagerSpawnStopsWhenShellDeadlineExpires (0.03s)
--- PASS: TestStopWorkspaceClosed (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/agent	0.669s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	0.002s
```

### Red runs

For each mutation, I changed one source file, ran the tests of the item, and then ran `git checkout` on the file. After each run, `git status --short` was empty.

| Mutation | Change | Tests that failed, and the assertion |
| --- | --- | --- |
| v1-noparent | `managerParentWorkspace` returns `""` | Adopts: shell timeout on `wBD:p1`, and no tab in `wBC`. ReusesRecorded and CreatesNewTab: `unexpected method "workspace.create"`. ClearsOldLayout: no adoption. AdoptFlagNeeds: `does not hold the parent checkout` |
| v1-nocwd | remove `Cwd` from `tab.create` | Adopts and CreatesNewTab: `tab.create = {...no cwd...}; want one named tab` |
| v1-noreuse | `reuseManagerTab` never reuses | ReusesRecorded: `unexpected method "tab.create"` |
| v1-keepcreated | flip `CreatedByHive` before the new tab | CreatesNewTab: `want old workspace ownership and new tab ownership` |
| v1-label | workspace label without `hive:` | CreatesWorkspace: `workspace.create = {...label "demo"...}; want one repo workspace` |
| v1-oldkeep | keep the old-layout record | ClearsOldLayout: `tab create: returned topology does not match the request` |
| v1-oldclose | send `workspace.close` for the old-layout workspace | ClearsOldLayout: `unexpected method "workspace.close"`, `old-layout workspace was closed` |
| v2-noprompt | remove the `promptManagerInstructions` call | both V2 tests: `agent.prompt requests = 0; want 1` |
| v3-noclear | `clearProjectTab` keeps the tab fields | StopAdopted and StartFailure: tab fields remain |
| v3-noconfirm | ignore `confirmation_required` | StopConfirmationRequired: `agent status = "live"; want stopped` |
| v3-cleanupclose | tab cleanup also sends `workspace.close` | StartFailure: `start failure closed an adopted workspace` |
| v3-closedkeep | "closed" topology keeps the record | StopWorkspaceClosed: `project workspace = &{ID:"w1"...}; want nil` |
| v7-nowait | `waitForShell` returns at once | all 4 V7 tests (for example, `pane.get requests before agent.start = 0; want 2`, and `want pane and deadline`) |
| v7-nocount | remove the one-foreground-process rule | StartupProcessesToClear: `pane.process_info requests before agent.start = 1; want 2` |

### V4

```text
commit 88004bf0
internal/agent/spawn_worker.go:145:		created, createErr := herdr.TabCreate(ctx, herdr.TabCreateParams{
internal/agent/workspace.go:171:	created, err := herdr.TabCreate(ctx, herdr.TabCreateParams{
internal/agent/workspace.go:204:	created, err := herdr.WorkspaceCreate(ctx, herdr.WorkspaceCreateParams{
internal/lib/herdr/tab_create.go:24:func TabCreate(ctx context.Context, params TabCreateParams, opts ...Option) (TabCreated, error) {
internal/lib/herdr/workspace.go:44:func WorkspaceCreate(ctx context.Context, params WorkspaceCreateParams, opts ...Option) (WorkspaceCreated, error) {
-- WorktreeList
internal/lib/herdr/worktree.go:30:func WorktreeList(ctx context.Context, params WorktreeListParams, opts ...Option) (WorktreeListResult, error) {
internal/agent/workspace.go:115:	listing, err := herdr.WorktreeList(ctx, herdr.WorktreeListParams{Cwd: repositoryPath}, herdr.WithLogger(logger))
-- resolver
internal/agent/spawn.go:339:		choice, message := resolveManagerWorkspace(ctx, home, options, logger)
internal/agent/workspace.go:28:func resolveManagerWorkspace(ctx context.Context, home string, options SpawnOptions, logger *slog.Logger) (managerChoice, string) {
internal/agent/workspace.go:53:	parentID, err := managerParentWorkspace(ctx, project.Path, logger)
internal/agent/workspace.go:114:func managerParentWorkspace(ctx context.Context, repositoryPath string, logger *slog.Logger) (string, error) {
-- wait/AgentStart
internal/agent/shell_ready.go:17:func waitForShell(ctx context.Context, paneID string, timeout time.Duration, logger *slog.Logger) error {
internal/lib/herdr/agent.go:84:func AgentStart(ctx context.Context, params AgentStartParams, opts ...Option) (AgentStarted, error) {
internal/agent/spawn.go:152:	if waitErr := waitForShell(ctx, paneID, readyTimeout, logger); waitErr != nil {
internal/agent/spawn.go:165:	started, startErr := herdr.AgentStart(ctx, herdr.AgentStartParams{
```

### V5

Before the check, `wBC` had the tabs `wBC:t1` and `wBC:t2`. I built the binary in my scratch directory, registered the worktree as the project `t03-verify`, and spawned the manager.

```text
$ hive agent spawn --role manager --project t03-verify --name t3-check-manager --harness pi --model opencode-go/deepseek-v4.1-flash --json
outcome ok, phase complete, last effect "prompted"
agent pane_id wBC:p8, workspace_id wBC, tab_id wBC:t7, cwd /home/penguin/source/hive/projects/hive/worktrees/3-manager-adopts-parent
$ herdr worktree list --cwd "$PWD"
source: {"source_checkout_path":"/home/penguin/source/hive","source_workspace_id":"wBC"}
$ herdr tab list --workspace wBC
wBC:t1 label "1" | wBC:t2 label "2" | wBC:t7 label "t3-check-manager", pane_count 1
$ herdr pane list --workspace wBC
wBC:p1 tab t1 (claude, coordinator) | wBC:p3 tab t2 (shell) | wBC:p8 tab t7 (pi)
$ herdr pane read wBC:p8
the end of the generated instructions (the command list), then one turn: "[skill] hive-manager", "$ cd .../3-manager-adopts-parent && pwd && git status --short ..."
```

One message: the pi session file holds 1 user message. Its first line is `You are the manager of project t03-verify.`, and it has 1593 characters. `projects/t03-verify/AGENTS.md` has 1594 bytes, the same text plus the final newline. In the home log, `agent.prompt` gave one `agent_not_ready` reply, which is the retry before the turn starts, and then one `agent_prompted` reply.

```text
$ hive agent stop --agent t3-check-manager --json
outcome ok; effects: stopped, pane wBC:p8 closed, workspace wBC "kept: 2 other panes"; topology "surviving"
$ herdr tab list --workspace wBC
wBC:t1 label "1" | wBC:t2 label "2"
$ herdr pane get wBC:p8
pane_not_found
```

No step sent `workspace.close`. The only pane that this check closed is `wBC:p8`, which the check made. Cleanup removed the binary, the agent record, the claim, the project record `t03-verify`, `projects/t03-verify/`, and `overrides/presets.md`. After cleanup, `.hive/state` holds only `counters`, `inbox`, and `lock`, the same as before the check. The logs stay.

### V6

```text
== gofmt
exit 0
== vet
exit 0
== lint
0 issues.
exit 0
== test
(all packages ok; no FAIL)
exit 0
== ste
--- PASS: TestSTERatchet (0.18s)
ok  	github.com/Harrison-Blair/hive	0.179s
exit 0
```

`ste_test.go` lists `AGENTS.md` and `internal/agent/templates/project-agents.md`. In `testdata/ste-baseline.json`, both have the baseline 0. `git log origin/dev..HEAD -- testdata/ste-baseline.json ste_test.go` is empty. `03-architecture` does not occur in `ste_test.go`.
