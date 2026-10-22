# T09-worktree-create-and-remove FIX-01

| Item | Value |
| --- | --- |
| Task | T09-worktree-create-and-remove |
| Item | FIX-01 |
| Proves | AC1, AC2, AC3, AC4 and source findings F01-F06, F08-F10 |
| Base | `ef396c34809e9b5da157928afe92bde7568004e6` |
| Source commit | `1653bf488e61d4f9f910f0a0fe28efdf2c25d066` |
| Date | 2026-10-18 |
| Agent | worker, pi, opencode-go/deepseek-v4-flash:medium; requested thinking medium, effective high |

## Source map

| Path | Change |
| --- | --- |
| `internal/worktree/create.go` | Related-read group, selected-key checks, observed effects before persistence, physical worktree proof, bounded uncertainty, one latest-record group, scoped logs, bounded log copies. |
| `internal/worktree/remove.go` | Related-read group, selected-key checks, permitted task path and branch, registered PRIMARY source check, latest-record ownership checks, scoped logs. |
| `internal/lib/gitx/gitx.go` | `RefExists` separates an absent ref from a repository or command failure. |
| `internal/worktree/remove_test.go` | The target-linked case uses a permitted primary checkout. The assertion stays. |
| `internal/worktree/fix01_test.go` | New assertions for F01-F06, F08-F10, the short read and write groups, and log scope. |
| `internal/lib/gitx/fix01_test.go` | New assertion for F10. |

## Red run

The red tree is the base commit plus the two new test files. The source files are exact base bytes.

```sh
go test -race -count=1 -timeout=120s ./internal/worktree ./internal/lib/gitx -run TestFix -v
```

```text
base ef396c34809e9b5da157928afe92bde7568004e6
tests: worktree/fix01_test.go + lib/gitx/fix01_test.go from source commit 1653bf488e61d4f9f910f0a0fe28efdf2c25d066
$ go test -race -count=1 -timeout=120s ./internal/worktree ./internal/lib/gitx -run TestFix -v
=== RUN   TestFixCreateTaskWriteFaultRetainsKnownRepositoryEffect
    fix01_test.go:116: effects = [{Resource:workspace Path:w2 Action:created} {Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixCreateTaskWriteFaultRetainsKnownRepositoryEffect1247009767/002/projects/repo-live/worktrees/1-live-demo Action:created} {Resource:branch Path:task/1-live-demo Action:created}], want the observed w7 repository workspace
--- FAIL: TestFixCreateTaskWriteFaultRetainsKnownRepositoryEffect (0.01s)
=== RUN   TestFixCreateUnknownPlainFilePathIsNotWorktree
    fix01_test.go:158: effects = [{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixCreateUnknownPlainFilePathIsNotWorktree2295577532/002/projects/repo-live/worktrees/1-live-demo Action:created}], want no created-worktree claim for a plain file
--- FAIL: TestFixCreateUnknownPlainFilePathIsNotWorktree (0.01s)
=== RUN   TestFixCreateUnknownRepositoryOnlyPersistsProject
    fix01_test.go:203: repo_workspace_id = "", want w7 without a linked path
--- FAIL: TestFixCreateUnknownRepositoryOnlyPersistsProject (0.01s)
=== RUN   TestFixCreateUnknownDiagnosticsStayExplicit
=== RUN   TestFixCreateUnknownDiagnosticsStayExplicit/branch_inspection
    fix01_test.go:237: Create() = {Outcome:partial Phase:request Effects:[] Error:bad reply: empty reply Result:{Task:{ID: Project: Title: ControllingRole: ActingAgent: Status: DependsOn:[] Verification:<nil> Worktree:<nil> Delivery:<nil> CreatedAt: UpdatedAt: CancelReason:} WorkspaceID: TabID: RootPane: Path: Branch: RepoWorkspaceID:}}, want a bounded branch inspection diagnostic
=== RUN   TestFixCreateUnknownDiagnosticsStayExplicit/path_inspection
    fix01_test.go:263: Create() = {Outcome:partial Phase:request Effects:[] Error:bad reply: empty reply Result:{Task:{ID: Project: Title: ControllingRole: ActingAgent: Status: DependsOn:[] Verification:<nil> Worktree:<nil> Delivery:<nil> CreatedAt: UpdatedAt: CancelReason:} WorkspaceID: TabID: RootPane: Path: Branch: RepoWorkspaceID:}}, want a non-absence path diagnostic
=== RUN   TestFixCreateUnknownDiagnosticsStayExplicit/repository_ambiguity
    fix01_test.go:289: Create() = {Outcome:partial Phase:request Effects:[] Error:bad reply: empty reply Result:{Task:{ID: Project: Title: ControllingRole: ActingAgent: Status: DependsOn:[] Verification:<nil> Worktree:<nil> Delivery:<nil> CreatedAt: UpdatedAt: CancelReason:} WorkspaceID: TabID: RootPane: Path: Branch: RepoWorkspaceID:}}, want an explicit ambiguity diagnostic
--- FAIL: TestFixCreateUnknownDiagnosticsStayExplicit (0.02s)
    --- FAIL: TestFixCreateUnknownDiagnosticsStayExplicit/branch_inspection (0.01s)
    --- FAIL: TestFixCreateUnknownDiagnosticsStayExplicit/path_inspection (0.01s)
    --- FAIL: TestFixCreateUnknownDiagnosticsStayExplicit/repository_ambiguity (0.01s)
=== RUN   TestFixCreateSelectedKeyMustMatchLoadedTask
=== RUN   TestFixCreateSelectedKeyMustMatchLoadedTask/invalid_record_id
    fix01_test.go:339: methods = [workspace.list], want no request
=== RUN   TestFixCreateSelectedKeyMustMatchLoadedTask/different_valid_record_id
    fix01_test.go:339: methods = [workspace.list], want no request
--- FAIL: TestFixCreateSelectedKeyMustMatchLoadedTask (0.01s)
    --- FAIL: TestFixCreateSelectedKeyMustMatchLoadedTask/invalid_record_id (0.01s)
    --- FAIL: TestFixCreateSelectedKeyMustMatchLoadedTask/different_valid_record_id (0.01s)
=== RUN   TestFixCreateSelectedKeyMustMatchLoadedProject
    fix01_test.go:367: methods = [workspace.list], want no request
--- FAIL: TestFixCreateSelectedKeyMustMatchLoadedProject (0.01s)
=== RUN   TestFixRemoveSelectedKeyMustMatchLoadedTask
    fix01_test.go:394: Remove() = {Outcome:ok Phase:complete Effects:[{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixRemoveSelectedKeyMustMatchLoadedTask1328450023/002/projects/repo-live/worktrees/1-live-demo Action:removed}] Error:<nil> Result:{Task:{ID:2-other Project:repo-live Title:live demo ControllingRole:manager-repo-live ActingAgent: Status:verified DependsOn:[] Verification:<nil> Worktree:0xc0001302a0 Delivery:<nil> CreatedAt:2026-10-18T00:00:00Z UpdatedAt:2026-10-18T11:46:45.422922558Z CancelReason:} Path:/tmp/t9f01r.7B1ony/TestFixRemoveSelectedKeyMustMatchLoadedTask1328450023/002/projects/repo-live/worktrees/1-live-demo Branch:task/1-live-demo WorkspaceClosed:false}}, want failed
--- FAIL: TestFixRemoveSelectedKeyMustMatchLoadedTask (0.01s)
=== RUN   TestFixRemoveSelectedKeyMustMatchLoadedProject
    fix01_test.go:425: Remove() = {Outcome:ok Phase:complete Effects:[{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixRemoveSelectedKeyMustMatchLoadedProject1093297783/002/projects/repo-live/worktrees/1-live-demo Action:removed}] Error:<nil> Result:{Task:{ID:1-live-demo Project:repo-live Title:live demo ControllingRole:manager-repo-live ActingAgent: Status:verified DependsOn:[] Verification:<nil> Worktree:0xc0001d0690 Delivery:<nil> CreatedAt:2026-10-18T00:00:00Z UpdatedAt:2026-10-18T11:46:45.435105702Z CancelReason:} Path:/tmp/t9f01r.7B1ony/TestFixRemoveSelectedKeyMustMatchLoadedProject1093297783/002/projects/repo-live/worktrees/1-live-demo Branch:task/1-live-demo WorkspaceClosed:false}}, want failed
--- FAIL: TestFixRemoveSelectedKeyMustMatchLoadedProject (0.01s)
=== RUN   TestFixRemoveRejectsLinkedRegisteredSource
=== RUN   TestFixRemoveRejectsLinkedRegisteredSource/false
    fix01_test.go:457: Remove() = {Outcome:ok Phase:complete Effects:[{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsLinkedRegisteredSourcefalse451680699/002/projects/repo-live/worktrees/1-live-demo Action:removed}] Error:<nil> Result:{Task:{ID:1-live-demo Project:repo-live Title:live demo ControllingRole:manager-repo-live ActingAgent: Status:verified DependsOn:[] Verification:<nil> Worktree:0xc0001d8930 Delivery:<nil> CreatedAt:2026-10-18T00:00:00Z UpdatedAt:2026-10-18T11:46:45.44876452Z CancelReason:} Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsLinkedRegisteredSourcefalse451680699/002/projects/repo-live/worktrees/1-live-demo Branch:task/1-live-demo WorkspaceClosed:false}}, want the registered-primary refusal
=== RUN   TestFixRemoveRejectsLinkedRegisteredSource/true
    fix01_test.go:457: Remove() = {Outcome:ok Phase:complete Effects:[{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsLinkedRegisteredSourcetrue2727957647/002/projects/repo-live/worktrees/1-live-demo Action:removed}] Error:<nil> Result:{Task:{ID:1-live-demo Project:repo-live Title:live demo ControllingRole:manager-repo-live ActingAgent: Status:verified DependsOn:[] Verification:<nil> Worktree:0xc000482620 Delivery:<nil> CreatedAt:2026-10-18T00:00:00Z UpdatedAt:2026-10-18T11:46:45.460225271Z CancelReason:} Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsLinkedRegisteredSourcetrue2727957647/002/projects/repo-live/worktrees/1-live-demo Branch:task/1-live-demo WorkspaceClosed:false}}, want the registered-primary refusal
--- FAIL: TestFixRemoveRejectsLinkedRegisteredSource (0.03s)
    --- FAIL: TestFixRemoveRejectsLinkedRegisteredSource/false (0.01s)
    --- FAIL: TestFixRemoveRejectsLinkedRegisteredSource/true (0.01s)
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata/path/false
    fix01_test.go:487: Remove() = {Outcome:ok Phase:complete Effects:[{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsPathAndBranchMetadatapathfalse695080619/003/unrelated-checkout Action:removed}] Error:<nil> Result:{Task:{ID:1-live-demo Project:repo-live Title:live demo ControllingRole:manager-repo-live ActingAgent: Status:verified DependsOn:[] Verification:<nil> Worktree:0xc0004820e0 Delivery:<nil> CreatedAt:2026-10-18T00:00:00Z UpdatedAt:2026-10-18T11:46:45.472258978Z CancelReason:} Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsPathAndBranchMetadatapathfalse695080619/003/unrelated-checkout Branch:user/precious WorkspaceClosed:false}}, want the permitted-path refusal
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata/branch/false
    fix01_test.go:513: Remove() = {Outcome:ok Phase:complete Effects:[{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsPathAndBranchMetadatabranchfalse3951325323/002/projects/repo-live/worktrees/1-live-demo Action:removed}] Error:<nil> Result:{Task:{ID:1-live-demo Project:repo-live Title:live demo ControllingRole:manager-repo-live ActingAgent: Status:verified DependsOn:[] Verification:<nil> Worktree:0xc0000ad880 Delivery:<nil> CreatedAt:2026-10-18T00:00:00Z UpdatedAt:2026-10-18T11:46:45.483797544Z CancelReason:} Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsPathAndBranchMetadatabranchfalse3951325323/002/projects/repo-live/worktrees/1-live-demo Branch:user/precious WorkspaceClosed:false}}, want the task-branch refusal
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata/path/true
    fix01_test.go:487: Remove() = {Outcome:ok Phase:complete Effects:[{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsPathAndBranchMetadatapathtrue946300176/003/unrelated-checkout Action:removed}] Error:<nil> Result:{Task:{ID:1-live-demo Project:repo-live Title:live demo ControllingRole:manager-repo-live ActingAgent: Status:verified DependsOn:[] Verification:<nil> Worktree:0xc0001d83f0 Delivery:<nil> CreatedAt:2026-10-18T00:00:00Z UpdatedAt:2026-10-18T11:46:45.493782172Z CancelReason:} Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsPathAndBranchMetadatapathtrue946300176/003/unrelated-checkout Branch:user/precious WorkspaceClosed:false}}, want the permitted-path refusal
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata/branch/true
    fix01_test.go:513: Remove() = {Outcome:ok Phase:complete Effects:[{Resource:worktree Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsPathAndBranchMetadatabranchtrue2381386379/002/projects/repo-live/worktrees/1-live-demo Action:removed}] Error:<nil> Result:{Task:{ID:1-live-demo Project:repo-live Title:live demo ControllingRole:manager-repo-live ActingAgent: Status:verified DependsOn:[] Verification:<nil> Worktree:0xc0001d8e70 Delivery:<nil> CreatedAt:2026-10-18T00:00:00Z UpdatedAt:2026-10-18T11:46:45.503353201Z CancelReason:} Path:/tmp/t9f01r.7B1ony/TestFixRemoveRejectsPathAndBranchMetadatabranchtrue2381386379/002/projects/repo-live/worktrees/1-live-demo Branch:user/precious WorkspaceClosed:false}}, want the task-branch refusal
--- FAIL: TestFixRemoveRejectsPathAndBranchMetadata (0.04s)
    --- FAIL: TestFixRemoveRejectsPathAndBranchMetadata/path/false (0.01s)
    --- FAIL: TestFixRemoveRejectsPathAndBranchMetadata/branch/false (0.01s)
    --- FAIL: TestFixRemoveRejectsPathAndBranchMetadata/path/true (0.01s)
    --- FAIL: TestFixRemoveRejectsPathAndBranchMetadata/branch/true (0.01s)
=== RUN   TestFixExternalLogsCarryApplicableIDs
    fix01_test.go:558: external event = map[command:git rev-parse --git-dir component:worktree level:DEBUG msg:process start time:2026-10-18T07:46:45.50766219-04:00], want project/task scope
--- FAIL: TestFixExternalLogsCarryApplicableIDs (0.01s)
=== RUN   TestFixRemoveExternalLogsCarryApplicableIDs
    fix01_test.go:592: external event = map[command:git rev-parse --git-dir component:worktree level:DEBUG msg:process start time:2026-10-18T07:46:45.520653216-04:00], want project/task scope
--- FAIL: TestFixRemoveExternalLogsCarryApplicableIDs (0.01s)
=== RUN   TestFixLargeDiagnosticLogIsBounded
    fix01_test.go:625: logged error bytes = 9000, want at most 4096
--- FAIL: TestFixLargeDiagnosticLogIsBounded (0.01s)
=== RUN   TestFixCreateTaskFaultLogsKnownEffect
    fix01_test.go:678: known w7 effect missing from logs after a persistence failure
--- FAIL: TestFixCreateTaskFaultLogsKnownEffect (0.01s)
=== RUN   TestFixHeldCreateAndRemoveReleaseStoreLock
=== RUN   TestFixHeldCreateAndRemoveReleaseStoreLock/create
=== RUN   TestFixHeldCreateAndRemoveReleaseStoreLock/remove
--- PASS: TestFixHeldCreateAndRemoveReleaseStoreLock (0.02s)
    --- PASS: TestFixHeldCreateAndRemoveReleaseStoreLock/create (0.01s)
    --- PASS: TestFixHeldCreateAndRemoveReleaseStoreLock/remove (0.01s)
=== RUN   TestFixPrecheckFailureLogCarriesIDs
    fix01_test.go:809: precheck error event = map[command:worktree create component:worktree error:base ref not found: missing level:ERROR msg:worktree error phase:precheck task_id:1-live-demo time:2026-10-18T07:46:45.571752161-04:00], want project/task scope
--- FAIL: TestFixPrecheckFailureLogCarriesIDs (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/worktree	0.238s
=== RUN   TestFixRefExistsDistinguishesAbsenceFromFailure
    fix01_test.go:13: RefExists(non-repository) = false, <nil>, want an error
--- FAIL: TestFixRefExistsDistinguishesAbsenceFromFailure (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/gitx	0.004s
FAIL
exit 1
```

The held-request lock control passes on the base tree. It is a control, not a finding.

## Green run

```sh
go test -race -count=1 -timeout=120s ./internal/worktree ./internal/lib/gitx -run TestFix -v
```

```text
commit 1653bf48
TMPDIR=/tmp/t9f01v.R9SG2J
$ go test -race -count=1 -timeout=120s ./internal/worktree ./internal/lib/gitx -run TestFix -v
=== RUN   TestFixCreateTaskWriteFaultRetainsKnownRepositoryEffect
--- PASS: TestFixCreateTaskWriteFaultRetainsKnownRepositoryEffect (0.01s)
=== RUN   TestFixCreateUnknownPlainFilePathIsNotWorktree
--- PASS: TestFixCreateUnknownPlainFilePathIsNotWorktree (0.01s)
=== RUN   TestFixCreateUnknownRepositoryOnlyPersistsProject
--- PASS: TestFixCreateUnknownRepositoryOnlyPersistsProject (0.01s)
=== RUN   TestFixCreateUnknownDiagnosticsStayExplicit
=== RUN   TestFixCreateUnknownDiagnosticsStayExplicit/branch_inspection
=== RUN   TestFixCreateUnknownDiagnosticsStayExplicit/path_inspection
=== RUN   TestFixCreateUnknownDiagnosticsStayExplicit/repository_ambiguity
--- PASS: TestFixCreateUnknownDiagnosticsStayExplicit (0.02s)
    --- PASS: TestFixCreateUnknownDiagnosticsStayExplicit/branch_inspection (0.01s)
    --- PASS: TestFixCreateUnknownDiagnosticsStayExplicit/path_inspection (0.01s)
    --- PASS: TestFixCreateUnknownDiagnosticsStayExplicit/repository_ambiguity (0.01s)
=== RUN   TestFixCreateSelectedKeyMustMatchLoadedTask
=== RUN   TestFixCreateSelectedKeyMustMatchLoadedTask/invalid_record_id
=== RUN   TestFixCreateSelectedKeyMustMatchLoadedTask/different_valid_record_id
--- PASS: TestFixCreateSelectedKeyMustMatchLoadedTask (0.01s)
    --- PASS: TestFixCreateSelectedKeyMustMatchLoadedTask/invalid_record_id (0.00s)
    --- PASS: TestFixCreateSelectedKeyMustMatchLoadedTask/different_valid_record_id (0.00s)
=== RUN   TestFixCreateSelectedKeyMustMatchLoadedProject
--- PASS: TestFixCreateSelectedKeyMustMatchLoadedProject (0.00s)
=== RUN   TestFixRemoveSelectedKeyMustMatchLoadedTask
--- PASS: TestFixRemoveSelectedKeyMustMatchLoadedTask (0.01s)
=== RUN   TestFixRemoveSelectedKeyMustMatchLoadedProject
--- PASS: TestFixRemoveSelectedKeyMustMatchLoadedProject (0.01s)
=== RUN   TestFixRemoveRejectsLinkedRegisteredSource
=== RUN   TestFixRemoveRejectsLinkedRegisteredSource/false
=== RUN   TestFixRemoveRejectsLinkedRegisteredSource/true
--- PASS: TestFixRemoveRejectsLinkedRegisteredSource (0.02s)
    --- PASS: TestFixRemoveRejectsLinkedRegisteredSource/false (0.01s)
    --- PASS: TestFixRemoveRejectsLinkedRegisteredSource/true (0.01s)
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata/path/false
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata/branch/false
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata/path/true
=== RUN   TestFixRemoveRejectsPathAndBranchMetadata/branch/true
--- PASS: TestFixRemoveRejectsPathAndBranchMetadata (0.02s)
    --- PASS: TestFixRemoveRejectsPathAndBranchMetadata/path/false (0.01s)
    --- PASS: TestFixRemoveRejectsPathAndBranchMetadata/branch/false (0.01s)
    --- PASS: TestFixRemoveRejectsPathAndBranchMetadata/path/true (0.01s)
    --- PASS: TestFixRemoveRejectsPathAndBranchMetadata/branch/true (0.01s)
=== RUN   TestFixExternalLogsCarryApplicableIDs
--- PASS: TestFixExternalLogsCarryApplicableIDs (0.01s)
=== RUN   TestFixRemoveExternalLogsCarryApplicableIDs
--- PASS: TestFixRemoveExternalLogsCarryApplicableIDs (0.01s)
=== RUN   TestFixLargeDiagnosticLogIsBounded
--- PASS: TestFixLargeDiagnosticLogIsBounded (0.01s)
=== RUN   TestFixCreateTaskFaultLogsKnownEffect
--- PASS: TestFixCreateTaskFaultLogsKnownEffect (0.01s)
=== RUN   TestFixHeldCreateAndRemoveReleaseStoreLock
=== RUN   TestFixHeldCreateAndRemoveReleaseStoreLock/create
=== RUN   TestFixHeldCreateAndRemoveReleaseStoreLock/remove
--- PASS: TestFixHeldCreateAndRemoveReleaseStoreLock (0.02s)
    --- PASS: TestFixHeldCreateAndRemoveReleaseStoreLock/create (0.01s)
    --- PASS: TestFixHeldCreateAndRemoveReleaseStoreLock/remove (0.01s)
=== RUN   TestFixPrecheckFailureLogCarriesIDs
--- PASS: TestFixPrecheckFailureLogCarriesIDs (0.01s)
ok  	github.com/Harrison-Blair/hive/internal/worktree	1.201s
=== RUN   TestFixRefExistsDistinguishesAbsenceFromFailure
--- PASS: TestFixRefExistsDistinguishesAbsenceFromFailure (0.01s)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.009s
PASS
exit 0
```

## First-review probe replay

The exact retained `create_verify_test.go` and `remove_audit_test.go` were copied into the package, run, and removed. Only the external socket peer is replaced. The source is the fixed source.

```sh
T09_CAPTURE_DIR=... go test -race -count=1 -timeout=120s ./internal/worktree -run TestAuditCreate -v
T09_AUDIT_CAPTURES=... go test -race -count=1 -timeout=180s ./internal/worktree -run TestAudit -v
```

```text
create: ok github.com/Harrison-Blair/hive/internal/worktree 1.445s; exit 0
remove: ok github.com/Harrison-Blair/hive/internal/worktree 1.298s; exit 0
```

The create replay passes all 19 top-level audit tests. The remove replay passes all 8 top-level audit tests. The prior F01-F06 and F08-F10 assertions now pass.

## V1 and V2 repeat

The exact V1 and V2 commands ran on the fixed source. The outputs are in `v1-repeat.txt` and `v2-repeat.txt`.

```sh
go test -race -count=1 -timeout=120s ./internal/worktree -run TestCreate -v
go test -race -count=1 -timeout=120s ./internal/worktree ./internal/lib/gitx -run 'TestRemove|TestGitx' -v
```

```text
V1: ok github.com/Harrison-Blair/hive/internal/worktree 1.103s; exit 0
V2: ok github.com/Harrison-Blair/hive/internal/worktree 1.195s; exit 0
V2: ok github.com/Harrison-Blair/hive/internal/lib/gitx 1.030s; exit 0
```

Both items pass. The worker V1 create cases, the worker V2 removal cases, and both Git helper tests pass.

## CLI replay

The exact `cli-probe.py` from the create verification ran with a new short `TMPDIR` and the fixed worktree as the source. The script SHA256 is `a78e75a6a6f98272b04e996a8788009db276d13c73ed457ced0b1543387437e3`.

```text
text status 0 wire methods ['ping', 'workspace.list', 'worktree.create', 'workspace.list']
json status 0 wire methods ['ping', 'workspace.list', 'worktree.create', 'workspace.list']
task-fault status 1 wire methods ['ping', 'workspace.list', 'worktree.create', 'workspace.list']
unknown status 1 wire methods ['ping', 'workspace.list', 'worktree.create', 'workspace.list']
concurrent status 0 wire methods ['ping', 'workspace.list', 'worktree.create', 'workspace.list']
```

The task-fault envelope is `partial/task` with effects `w2`, the worktree, the branch, and `w7`, plus `repo_workspace_id: w7`. The unknown envelope is `partial/request` with the real worktree, branch, `w7`, and `repo_workspace_id: w7`. The branch and worktree stay after both. The concurrent cancel stays in the latest record.

## Native V3

The native driver uses a unique marked fixture home, one fixture repository, and one absolute binary from the final source tree. It starts no agent or model child.

| Item | Value |
| --- | --- |
| Binary SHA256 | `a395c977d447474a6a75692e7ff09c1786b1c31723ba7c677ea2079b5460f52b` |
| Fixture home | `/tmp/t9f01n.fh7gjqh_` |
| Owned linked workspace | `wAV` |
| Owned PRIMARY repository workspace | `wAT` |
| Before set | `w9X`, `wAD`, `wAE`, `wAF` |
| After create | `w9X`, `wAD`, `wAE`, `wAF`, `wAT`, `wAV` |
| After remove | `w9X`, `wAD`, `wAE`, `wAF`, `wAT` |
| Final set | `w9X`, `wAD`, `wAE`, `wAF` |

```text
create exit 0
remove-created exit 1
remove-cancelled exit 0
branch-kept exit 0
workspace closed: wAV
worktree removed: /tmp/t9f01n.fh7gjqh_/projects/repo-t09-fix/worktrees/1-live-demo
branch kept: task/1-live-demo
cleanup-close-primary ok
final owned absent true
```

The product leaves the PRIMARY repository workspace `wAT` open. Cleanup gets `wAT`, checks its owned PRIMARY metadata, closes it with `close_group=false`, and parses the final list before the home is removed. The task keeps `cancel_reason`, `created_at`, path, and branch. The live topology is empty. The project keeps `repo_workspace_id: wAT`. No source or user workspace is closed.

## Mutation sensitivity

Each mutation ran in an owned source copy. The test failed on an assertion. The copy was restored and its SHA256 matched the source copy before the next mutation.

```text
F01-drop-repository-effect exit 1
F02-claim-path-only exit 1
F03-project-write-needs-worktree exit 1
F04-drop-branch-diagnostic exit 1
F04-drop-ambiguity-diagnostic exit 1
F05-create-ignore-task-id exit 1
F05-create-ignore-project-id exit 1
F05-remove-ignore-task-id exit 1
F06-drop-log-scope exit 1
F06-drop-log-bound exit 1
F06-drop-effect-log exit 1
F08-accept-linked-source exit 1
F09-ignore-permitted-path exit 1
F09-ignore-task-branch exit 1
F10-refexists-old-absence exit 1
restored green exit 0
```

## F07 disposition

F07 is a design observation, not a separate source finding. The implementation applies the approved grouping.

- One short held handle reads the related task and project records and checks the selected keys. It releases before Git or socket work.
- One short latest-record handle writes the owned task and project fields after the external waits.
- The task update keeps a concurrent status, title, acting agent, verification, dependencies, reason, and unrelated project fields.
- The task update rejects a changed record ID, project, or existing worktree. The remove update rejects a changed ID, project, path, branch, or a second removal.
- A failed prefix stays. There is no rollback, journal, or commit marker.
- `TestFixHeldCreateAndRemoveReleaseStoreLock` holds a real socket request and performs a real store write with a 500 ms deadline. The write succeeds during the external wait.
- The first-review `TestAuditCreateLatestProjectAndTask` and `TestAuditRemoveConcurrentLatestFields` pass on the fixed source.

## Gate

The full gate ran once on the final working tree before the source commit. The source commit bytes equal the gated bytes. Format printed nothing. Vet and lint passed. Race count 1 and `TestSTE` passed.

```text
commit ef396c34
TMPDIR=/tmp/t9f01g.x2sZWZ
$ gofmt -l $(git ls-files -co --exclude-standard "*.go")
exit 0
$ go vet ./...
exit 0
$ golangci-lint run --allow-serial-runners ./...
0 issues.
exit 0
$ go test -race -count=1 ./...
ok  	github.com/Harrison-Blair/hive	3.268s
ok  	github.com/Harrison-Blair/hive/cmd	1.085s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.205s
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.208s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.024s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	1.258s
ok  	github.com/Harrison-Blair/hive/internal/doctor	3.044s
ok  	github.com/Harrison-Blair/hive/internal/handoff	1.139s
ok  	github.com/Harrison-Blair/hive/internal/inbox	1.019s
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.070s
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.494s
ok  	github.com/Harrison-Blair/hive/internal/lib/home	1.023s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	1.225s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.068s
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/run	3.472s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	1.142s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.263s
ok  	github.com/Harrison-Blair/hive/internal/lib/version	1.004s
ok  	github.com/Harrison-Blair/hive/internal/logs	6.024s
ok  	github.com/Harrison-Blair/hive/internal/memory	2.153s
ok  	github.com/Harrison-Blair/hive/internal/project	1.011s
ok  	github.com/Harrison-Blair/hive/internal/task	1.164s
ok  	github.com/Harrison-Blair/hive/internal/tui	1.005s
ok  	github.com/Harrison-Blair/hive/internal/worktree	1.497s
exit 0
$ go test -run TestSTE -count=1 .
ok  	github.com/Harrison-Blair/hive	0.151s
exit 0
```

`go.mod` and `go.sum` have no diff. The new text did not raise the STE baseline.

## Hashes

| Path | SHA256 at source commit |
| --- | --- |
| `internal/lib/gitx/gitx.go` | `0d4ebd15812e10c518520b72aed0ed56c55a6bca23897c509a9a61715cf422a1` |
| `internal/lib/gitx/fix01_test.go` | `d807289cc6191d182a15461c7b554236e132992d9364dc6a604175a4dee6c881` |
| `internal/worktree/create.go` | `2d26277c825c1c1a31e63b798c85db459b088945ed091913f03fa882c143a25e` |
| `internal/worktree/remove.go` | `5cdf536735a72de798b631f6a84e817bcb8acce7dbfd0ba1d960774041bd3aec` |
| `internal/worktree/remove_test.go` | `741e57871b4f8c7ae7b3cad33c91fa5915c57bb8c28a6c697d53ce65da14f418` |
| `internal/worktree/fix01_test.go` | `240e012958ac46932ab97c06db231fc59db8b3588af79151d12ebeec7139d8d8` |

The raw red, green, audit replay, CLI, native, mutation, gate, and hash files are in the ignored capture directory `.hive/tmp/coordinator/reports/t09-fix-01-captures/`.

## Limits

- This pass corrects source only. Source acceptance remains pending fresh independent verification.
- The approved boundary is a local user and a cooperating writer. Hostile `Lstat` and `Open` replacement stays deferred for T01.
- The exact historical worker test-first stub setup was not retained. This pass does not fabricate that chronology or change the old V1-V4 evidence.
- The native check uses native workspaces and their default shells. It starts no agent or model child.
- Other agents can change their own resources. The native check accounts only for its owned IDs. It does not require the global workspace set to stay frozen.
- The short-writer observation from the first review stays outside the T09 acceptance conditions.

END T09 FIX 01
