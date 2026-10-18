# T09-worktree-create-and-remove FIX-02

| Item | Value |
| --- | --- |
| Task | T09-worktree-create-and-remove |
| Item | FIX-02 |
| Proves | AC1, AC2, AC4 and source findings R01 and F06 |
| Base | `bd865cb33b295759a78b03d2fa3fcfc6c10fbe45` |
| Source and test commit | `1cde6665cf559ea568bee7e200e74c65a179d6b5` |
| Date | 2026-10-18 |
| Agent | worker, pi, `opencode-go/deepseek-v4-flash:medium`. Requested thinking medium, effective high. |
| Session | `01a14eef-4aa4-7526-9d93-b6333c29ad28` |

## Source map

| Path | Change |
| --- | --- |
| `internal/worktree/create.go` | Check the captured project owner before the repository workspace write. Keep valid UTF-8 in the bounded log copy. |
| `internal/worktree/remove.go` | Check the captured workspace, tab, and root pane before the clear. |
| `internal/worktree/fix02_test.go` | New assertions for R01, F06, and the boundary controls. |

The source commit changes only those three paths. It changes no dependency, baseline, task, tracker, `AGENTS.md`, glossary, root state, or handoff. The old V1-V4, FIX-01, and VERIFIED bytes stay exact.

## R01 current resource metadata

### Create with a changed repository workspace

A cooperating writer stores `w99` on the project during the external wait. Create now reads the current project under the one short latest-record handle. The current `repo_workspace_id` differs from the captured value and from the observed `w7`. The product keeps `w99` and writes no project change. The task write stays. The outcome is `partial` with phase `project`. The effects list the actual `w2`, worktree, branch, and `w7`. The result keeps the observed `w7`.

### Create with a changed repository owner

A cooperating writer changes the registered path to another real repository during the external wait. Create compares the current physical path with the captured approved path. A different physical path is a conflict. The product keeps the changed owner and its unrelated fields. It attaches no old `w7` to that owner. The outcome is `partial` with the known effects. A physical alias of the captured path stays the same owner, because the check uses resolved paths.

### Remove with changed current topology

A cooperating writer stores `w3`, `w3:t2`, and `w3:p4` during the close reply. Remove closes `w2` and removes the real Git path first. The latest-record check then compares the current topology with the captured `w2`, `w2:t1`, and `w2:p1`. The current topology differs and is not empty. The product keeps the current record and returns `partial` with phase `task`. The effects keep the `w2` close and the Git removal. No rollback occurs.

### R01 controls

The new tests cover four create controls and four remove controls. The create controls keep unchanged metadata, the same desired repository workspace, a physical repository alias, and an unrelated concurrent field. The remove controls keep unchanged metadata, an already cleared topology, unrelated concurrent task fields, and a physical worktree alias. Each control keeps normal success. The tests check the outcome, the effects, the current record values, and the stored record bytes.

## F06 multibyte diagnostic bound

`boundLogText` now walks back from byte 4096 to the start of a UTF-8 character. The returned text stays valid UTF-8 and has at most 4096 bytes. The `truncated` field stays true when the copy is shorter. The create and remove error logs use the same helper. The output error stays separate and keeps the full diagnostic.

The new test sends a 9000-byte euro diagnostic through the real accepted socket helper. Only the external peer is fake. Git, the state store, and the product are real. Both commands produce a decoded log error of at most 4096 bytes with `truncated: true`. The test also checks literal ASCII and multibyte boundaries at 4096 and 4097 bytes. The check uses literal expected lengths, not values from the production constant.

## Red and green

The red tree was an owned copy of the exact base commit `bd865cb3`. It adds only `internal/worktree/fix02_test.go` with SHA256 `93eb1724d00f7b8589e7e3f67b3e8c31febfd3618ba00340be098ddc9b60c362`. The create and remove source bytes stay at the base.

```sh
TMPDIR=/tmp/t09f02ro.bwO18y go test -race -count=1 -timeout=180s ./internal/worktree -run 'TestFix02' -v
```

The three R01 tests fail on the outcome assertions. The two Unicode cases fail on the decoded length assertion. The ASCII and multibyte boundary test fails on the valid UTF-8 assertion. The control groups pass. No failure is a compile error or a timeout. The red output SHA256 is `7db943eedd0f8d6ee8ffd96198f8a2cce778811b781a84c4782dd3b19432cd94`.

```sh
TMPDIR=/tmp/t09f02r.iHwVvK go test -race -count=1 -timeout=180s ./internal/worktree -run 'TestFix02' -v
```

The green run passes every FIX-02 test at source commit `1cde6665`. Its output SHA256 is `1bdd8292368015a69e55ca7ac680b15f80831b1858b00d8c7810030e7b4fa90d`.

## First-review replay

The exact retained first-review probes ran in owned copies of the final source. Only the external socket peer is fake. Git, the state store, and the product are real.

| Replay | Result |
| --- | --- |
| V1 create command | PASS. 8 top-level tests. |
| V2 remove and gitx command | PASS. 12 remove tests and 2 gitx tests. |
| Create audit checklist | PASS. 19 top-level tests. |
| Remove audit checklist | PASS. 8 top-level tests. |
| Repeat controls | PASS. 9 top-level tests. |
| Held store coordination, 10 runs | PASS. 10 top-level runs. |

The create audit replay keeps F01-F05 and F06 controls. The remove audit replay keeps F08-F10 and the identity, force, path, nested repository, and unknown-close controls. The repeat replay keeps the task ID, project, path, and branch refusals. The held-store replay proves that the store lock is free during the external wait.

## CLI replay

The exact retained `cli-probe.py` ran with a new short `TMPDIR` and the final source copy. Its SHA256 is `a78e75a6a6f98272b04e996a8788009db276d13c73ed457ced0b1543387437e3`.

```text
text status 0
json status 0
task-fault status 1
unknown status 1
concurrent status 0
```

The task-fault case keeps `w2`, the worktree, the branch, and `w7`. The unknown case keeps the real worktree and `w7`. The concurrent cancel stays in the latest task record. The real Git worktree and branch stay after both cases.

## Native V3

The native driver uses one unique marked fixture home, one fixture repository, and one absolute binary from source commit `1cde6665`. The installed herdr is `0.9.5`, protocol `22`. The run starts native workspaces and their default shells only. It starts no agent and no model child.

| Item | Value |
| --- | --- |
| Binary SHA256 | `0e08b37fabd36a6c18854db78794e34552f83036ac8d01ac5ec21f78f3911f8d` |
| Binary build | `vcs.modified=false`, revision `1cde6665cf559ea568bee7e200e74c65a179d6b5` |
| Fixture home | `/tmp/t09f02n.6495csrl` |
| Owned PRIMARY workspace | `wAY` |
| Owned linked workspace | `wAZ` |
| Before set | `w9X`, `wAE`, `wAF` |
| After create | `w9X`, `wAE`, `wAF`, `wAY`, `wAZ` |
| After remove | `w9X`, `wAE`, `wAF`, `wAY` |
| Final set | `w9X`, `wAE`, `wAF` |

The create exits 0. The created-status remove exits 1 with `task is created: complete, verify, or cancel it first`. The cancel exits 0 and records its reason. The clean no-force remove exits 0 and prints:

```text
workspace closed: wAZ
worktree removed: /tmp/t09f02n.6495csrl/projects/repo-t09-verify/worktrees/1-live-demo
branch kept: task/1-live-demo
```

The path is absent. The branch stays at `54e34733fb81d4445dae97153caf83e06a84574d`. The task keeps its path, branch, creation time, cancellation reason, and removal time. The live topology is empty. The project keeps `repo_workspace_id: wAY`. The product leaves `wAY` open. Cleanup checks the owned PRIMARY metadata and closes only `wAY` with `close_group=false`. The close succeeds. Parsed lists prove both owned workspaces absent before fixture deletion. The before and final sets match. The `w9X` panes `w9X:p4` and `w9X:p5` stay. No other resource changed in this run.

All 42 real external native events carry `project_id` and `task_id`. The one error event is the created-status removal refusal.

## Mutation sensitivity

Each mutation ran in an owned source copy. The run restored the copy and its SHA256 matched before the next mutation. The restored tree passes.

| Mutation | Target | Result |
| --- | --- | --- |
| Drop the repository workspace check | create workspace conflict | FAIL on assertion |
| Drop the owner path check | create owner conflict | FAIL on assertion |
| Use a raw string path check | create physical alias control | FAIL on assertion |
| Drop the topology check | remove topology conflict | FAIL on assertion |
| Invert the UTF-8 boundary loop | Unicode and boundary tests | FAIL on assertion |

The first F06 mutation attempt removed the `unicode/utf8` use and caused a compile failure. The captures retain it as `mutation-F06-old-byte-slice-invalid-build.txt` and it does not count as a valid mutation. The corrected mutation keeps the import and fails on assertions.

## Gate

The full uncached gate ran once in the actual worktree at source commit `1cde6665`, before the evidence commit. The short `TMPDIR` was `/tmp/t09f02g.NCOkHm`.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 .
```

| Command | Result |
| --- | --- |
| Format | Empty output, exit 0 |
| Vet | Exit 0 |
| Lint | `0 issues.`, exit 0 |
| Race count 1 | Every package passes, exit 0 |
| `TestSTE` | PASS, exit 0 |

`go.mod` SHA256 is `17c5d345235b7d9ef496f32abfc4eb3b744335d8b565444fd67da7e0ab937a0b`. `go.sum` SHA256 is `b6c34f49f2d16df2a99bea85cd409c6fbfc909d9e750b3954a3fbb1068eedeee`. The baseline SHA256 is `d4c5882428b5e9ae7e829fc60f395c501f5eeb82b1d5ec2d74654e9e1197cfc9`. These bytes stay exact.

## Hashes

| Path | SHA256 at source commit |
| --- | --- |
| `internal/worktree/create.go` | `709850f50705955cbf184b944f36bf02cd18ebd971854b995015112cf3ef1f1e` |
| `internal/worktree/remove.go` | `b19677ab2324602577c9b828db6dded3722fd48938987980a97712b28657f1c8` |
| `internal/worktree/fix02_test.go` | `93eb1724d00f7b8589e7e3f67b3e8c31febfd3618ba00340be098ddc9b60c362` |

## Limits

- Source acceptance needs a fresh independent verification. This pass makes no landing, CP2, or later claim.
- The approved boundary is a local user, a cooperating writer, and static types. Hostile `Lstat` and `Open` replacement stays deferred for T01.
- A malicious socket peer beyond T05 and the live or unresolved agent removal guard before T20 stay deferred.
- The nonconforming nil-error short writer stays outside T09 acceptance.
- The exact historical original stub setup was not retained. This pass does not invent that chronology.

END T09 FIX 02
