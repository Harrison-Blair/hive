# 133-dev-main-pr VERIFIED

| Item | Value |
| --- | --- |
| Task | 133-dev-main-pr |
| Commit | `b43d62c1bcc96391c3d2dd62f8053dd1167b9981` |
| Date | 2026-10-23 |
| Agent | verifier, Pi, `gpt-6.1-sol`, `t133-verifier-1` |
| Pull request | https://github.com/Harrison-Blair/hive/pull/49 (#49) |
| PR head | `1caad89519c8754c8fa33eb17fa617b24f8ac6d5` |
| PR base | `9569ba99281aec1105393ba75685268fa7e0aaa9` |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1, AC2 | pass | One open PR has the required title, branches, and body sections. |
| V2 | AC3 | pass | Both `gate` checks report `SUCCESS` on the PR head. |
| V3 | AC4 | pass | Task documents and the explicitly permitted STE baseline entry only. The worktree is clean. |
| V4 | AC5 | pass | `main` did not move. PR #49 is open and unmerged, from `dev` into `main`. |
| V5 | AC6 | pass | All five gate commands pass in the task worktree. They also pass at the exact PR head. |

## Scope

This task opens a pull request. It changes no product source, test, or UI screen.
No test item needs a red run. No visual check applies to this task.
The verifier did not change the PR, the worker evidence, or the tracker.

## Runs

All commands ran from the worktree root. `git fetch origin` completed with no output and exit 0 before the checks.

### V1 Pull request

```sh
gh pr list --base main --head dev --state open --json number,url,title,headRefName,baseRefName,state
```

```text
[{"baseRefName":"main","headRefName":"dev","number":49,"state":"OPEN","title":"Hive TUI dashboard, agent screens, and hive ask","url":"https://github.com/Harrison-Blair/hive/pull/49"}]
```

```sh
gh pr view 49 --json body --jq .body
```

```text
## What hive is

Hive is a Go CLI and TUI that coordinates coding agents in herdr from one home. The user talks to one coordinator. A manager owns each project, and workers do one task each in their own worktree, with independent verifiers before a task lands. Hive builds hive: this repository is both the home and the first registered project.

Design authority: [`.hive/spec/hive-v1/README.md` on `dev`](https://github.com/Harrison-Blair/hive/blob/dev/.hive/spec/hive-v1/README.md)

Context: g1-hive-tui waves A-D (tasks 73 to 120) are already on `main` through #47 (merged 2026-10-22, merge `9569ba99`). They are not part of this pull request.

## What changed since #47

`origin/main..dev`: 86 commits (72 non-merge), head `1caad895`.

**Task 56-doctor-home-version** (releases)
- 56 Doctor warns when the binary version differs from the home tag — `feat(doctor): warn when the binary version differs from the home tag`

**g3-chat-status**
- 122 Chat working status, growing input, event toggle — `feat(tui): show the coordinator status in the chat input`, `fix(tui): keep the wrapped input cursor on the insertion point`

**g4-agents-live-default**
- 123 Agents: stopped agents hidden by default, working cards marked — `feat(tui): hide stopped agents at start and mark working cards`

**g5-live-agent-activity**
- 124 Agents: live overview, pod view, lit message lines — `feat(agent): log the sender of agent send and add the live agent boards`, `feat(tui): add the live pods overview, the one-pod view, and lit message lines`, `fix(tui): group lit messages by card so the newest message gives the arrow`
- 125 Agents: one-pod view always a tree — `feat(tui): draw the one-pod view as a scrolled tree`, `feat(tui): draw the coordinator line to the top cards of home`

**g6-hive-ask**
- 126 Mouse on all tabs and chat modes — `feat(tui): add mouse support, chat modes, and the ask count`, `feat(tui): install the approved chat mode grids and the ctrl+o input hint`, `fix(tui): let the mouse work while the Logs filter is open`
- 129 Agents cards from the live record — `fix(tui): draw each agent card from the live record`
- 131 `hive ask` and the question card — `feat(ask): add hive ask, the ask record, and the answer delivery`, `feat(tui): add the ask card to Chat`, `fix(tui): refuse ask answers from a blocked caller`, `fix(tui): pick the current option on enter from a note line`

**Task 132-tui-clock-tests**
- 132 TUI header clock tests no longer depend on the local zone — `fix(test): make the header clock tests zone-independent`

**Docs commits outside a task (dogfood friction log)**
- `e253857b docs(dogfood): add shutdown command entry`
- `6f741a6b docs(dogfood): record up and down offset in the wrapped chat input`
- `fa3ebf34 docs(dogfood): record no repair of a completed task before verify`
- `bdd2ee46 docs(dogfood): log the stale record on Agents cards`
- `dcdaca94 docs(dogfood): log a worker spawn with --pane that skips assign and prompt`
- `25620a65 docs(dogfood): record worker rm glob prompts`
- `34a7d85d docs(dogfood): merge rm glob prompts into the existing entry`
- `7d8f42db docs(dogfood): record silent stale binary on restart`
- `1caad895 docs(dogfood): record the CI time-zone gate gap`

Each task also carries its brief, evidence, decisions, approved boards, and verification commits (`docs(...)`, `test(tui): ...`).

## Gate on `dev` at `1caad895`

Run after `git fetch origin` on a clean checkout of `1caad895`:

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | no output, exit 0 |
| `go mod tidy -diff` | no output, exit 0 |
| `go vet ./...` | exit 0 |
| `golangci-lint run ./...` | exit 0 |
| `go test -race ./...` | all packages ok, exit 0 |

## CI on the head commit

`gh run list --branch dev --limit 1`: run [39906002053](https://github.com/Harrison-Blair/hive/actions/runs/39906002053) (CI, push, `1caad895`) — conclusion `success`.

Opened under `grants.md 2026-10-19 pull request from dev into main`. The owner merges.
```

The four introductory sentences agree with `AGENTS.md` and the v1 README.
The README link names `dev`. The change list names all nine landed tasks and all nine standalone docs commits.
Task-owned docs and evidence commits are grouped with their tasks.
The verifier compared `git log origin/main..dev` with the body and checked the available task records and landed briefs.
The range contains 86 commits, including 72 non-merge commits.
PR #47 has merge commit `9569ba99281aec1105393ba75685268fa7e0aaa9` and merge date 2026-10-22.
The body separates that prior work from the new work.
The gate table has all five commands. V5 also checks the exact PR head.
The CI run and grant line are present. Both CI runs still pass at this head.

```sh
gh run list --branch dev --limit 1
```

```text
completed	success	Hive TUI dashboard, agent screens, and hive ask	CI	dev	pull_request	39906533590	1m58s	2026-10-23T05:34:45Z
```

The latest run now includes the PR event. The push run recorded in the body remains successful on the same head.

### V2 CI

```sh
gh pr checks 49 --json name,state,link
```

```text
[{"link":"https://github.com/Harrison-Blair/hive/actions/runs/39906533590/job/120259680742","name":"gate","state":"SUCCESS"},{"link":"https://github.com/Harrison-Blair/hive/actions/runs/39906002053/job/120258005395","name":"gate","state":"SUCCESS"}]
```

```sh
gh run view 39906533590
```

```text

✓ dev CI Harrison-Blair/hive#49 · 39906533590
Triggered via pull_request about 4 minutes ago

JOBS
✓ gate in 1m54s (ID 120259680742)

ANNOTATIONS
- "The ubuntu-latest label will migrate to Ubuntu 26 beginning October 19, 2026. For more information, see https://github.com/actions/runner-images/issues/14748"
gate: .github#1


For more information about the job, try: gh run view --job=120259680742
View this run on GitHub: https://github.com/Harrison-Blair/hive/actions/runs/39906533590
```

```sh
gh run view 39906002053
```

```text

✓ dev CI Harrison-Blair/hive#49 · 39906002053
Triggered via push about 10 minutes ago

JOBS
✓ gate in 1m43s (ID 120258005395)

ANNOTATIONS
- "The ubuntu-latest label will migrate to Ubuntu 26 beginning October 19, 2026. For more information, see https://github.com/actions/runner-images/issues/14748"
gate: .github#1


For more information about the job, try: gh run view --job=120258005395
View this run on GitHub: https://github.com/Harrison-Blair/hive/actions/runs/39906002053
```

```sh
gh run view 39906533590 --json databaseId,headSha,headBranch,event,status,conclusion,url
gh run view 39906002053 --json databaseId,headSha,headBranch,event,status,conclusion,url
```

```text
{"conclusion":"success","databaseId":39906533590,"event":"pull_request","headBranch":"dev","headSha":"1caad89519c8754c8fa33eb17fa617b24f8ac6d5","status":"completed","url":"https://github.com/Harrison-Blair/hive/actions/runs/39906533590"}
{"conclusion":"success","databaseId":39906002053,"event":"push","headBranch":"dev","headSha":"1caad89519c8754c8fa33eb17fa617b24f8ac6d5","status":"completed","url":"https://github.com/Harrison-Blair/hive/actions/runs/39906002053"}
```

```sh
gh pr checks 49 --watch --interval 10
printf 'watch exit=%s\n' "$?"
```

The output below omits trailing tabs.

```text
gate	pass	1m43s	https://github.com/Harrison-Blair/hive/actions/runs/39906002053/job/120258005395
gate	pass	1m54s	https://github.com/Harrison-Blair/hive/actions/runs/39906533590/job/120259680742
gate	pass	1m43s	https://github.com/Harrison-Blair/hive/actions/runs/39906002053/job/120258005395
gate	pass	1m54s	https://github.com/Harrison-Blair/hive/actions/runs/39906533590/job/120259680742
watch exit=0
```

Both runs cover the PR head. The runner annotation does not report a failed check.

### V3 Diff

```sh
git diff dev..HEAD --stat
git status --short
```

```text
 .hive/spec/tui-v2/tasks/133-dev-main-pr/brief.md   | 122 ++++++++++++++++++
 .../tui-v2/tasks/133-dev-main-pr/evidence/V1.md    | 100 +++++++++++++++
 .../tui-v2/tasks/133-dev-main-pr/evidence/V2.md    |  83 ++++++++++++
 .../tui-v2/tasks/133-dev-main-pr/evidence/V3.md    |  59 +++++++++
 .../tui-v2/tasks/133-dev-main-pr/evidence/V4.md    |  37 ++++++
 .../tui-v2/tasks/133-dev-main-pr/evidence/V5.md    | 142 +++++++++++++++++++++
 testdata/ste-baseline.json                         |   1 +
 7 files changed, 544 insertions(+)
```

`git status --short` printed nothing before this report was created.
The only change outside the task folder is the following entry:

```sh
git diff dev..HEAD -- testdata/ste-baseline.json
```

```diff
diff --git a/testdata/ste-baseline.json b/testdata/ste-baseline.json
index a2f33836..13a2d9ad 100644
--- a/testdata/ste-baseline.json
+++ b/testdata/ste-baseline.json
@@ -132,6 +132,7 @@
   ".hive/spec/tui-v2/tasks/126-chat-modes-mouse/brief.md": 0,
   ".hive/spec/tui-v2/tasks/131-hive-ask/brief.md": 0,
   ".hive/spec/tui-v2/tasks/132-tui-clock-tests/brief.md": 0,
+  ".hive/spec/tui-v2/tasks/133-dev-main-pr/brief.md": 0,
   ".hive/spec/tui-v2/tasks/129-agents-stale-card/brief.md": 0,
   ".hive/spec/tui-v2/tasks/21-components-1/brief.md": 0,
   ".hive/spec/tui-v2/tasks/22-components-board-fix/brief.md": 0,
```

The brief's Source files table explicitly permits this entry for AC6.
This specific permission supplies the exception to the narrower V3 expected-path text.
No Go source, test code, module file, or CI workflow changed.

```sh
git diff --exit-code dev HEAD -- '*.go' go.mod go.sum .github/workflows
```

```text
```

### V4 Main and pull request state

```sh
git ls-remote origin main
gh pr view 49 --json state,mergedAt,baseRefName,headRefName
```

```text
9569ba99281aec1105393ba75685268fa7e0aaa9	refs/heads/main
{"baseRefName":"main","headRefName":"dev","mergedAt":null,"state":"OPEN"}
```

The verifier repeated the remote and PR checks at 05:41 UTC. Neither changed.

```sh
git ls-remote origin main
gh pr view 49 --json state,mergedAt,baseRefName,headRefName,headRefOid
```

```text
9569ba99281aec1105393ba75685268fa7e0aaa9	refs/heads/main
{"baseRefName":"main","headRefName":"dev","headRefOid":"1caad89519c8754c8fa33eb17fa617b24f8ac6d5","mergedAt":null,"state":"OPEN"}
```

### V5 Gate

Task worktree at `b43d62c1bcc96391c3d2dd62f8053dd1167b9981`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
printf 'gofmt exit=%s\n' "$?"
```

```text
gofmt exit=0
```

```sh
go mod tidy -diff
printf 'tidy exit=%s\n' "$?"
```

```text
tidy exit=0
```

```sh
go vet ./...
printf 'vet exit=%s\n' "$?"
```

```text
vet exit=0
```

```sh
golangci-lint run ./...
printf 'lint exit=%s\n' "$?"
```

```text
0 issues.
lint exit=0
```

```sh
go test -race ./...
printf 'race exit=%s\n' "$?"
```

```text
ok  	github.com/Harrison-Blair/hive	17.590s
ok  	github.com/Harrison-Blair/hive/cmd	3.475s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.115s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.141s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.227s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.154s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/ask	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.100s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	1.485s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	1.502s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.160s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.119s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.104s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.063s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	4.399s
ok  	github.com/Harrison-Blair/hive/internal/ask	(cached)
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.206s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	4.619s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	2.051s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.062s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.026s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.006s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.160s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	1.738s
ok  	github.com/Harrison-Blair/hive/internal/tui	27.622s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	1.959s
race exit=0
```

#### Exact PR head

The verifier also ran the five commands on detached commit `1caad89519c8754c8fa33eb17fa617b24f8ac6d5` in this worktree.
A shell trap restored `task/133-dev-main-pr` after the run.

| Command | Output | Exit |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | none | 0 |
| `go mod tidy -diff` | none | 0 |
| `go vet ./...` | none | 0 |
| `golangci-lint run ./...` | `0 issues.` | 0 |
| `go test -race ./...` | all packages pass | 0 |

The root package took 17.184s. The remaining package results were cached.
`git status --short` printed nothing at the PR head and after the branch was restored.
No scratch pane, profile, script, or product change was created by the verifier.
The pre-existing `.hive/tmp/pr-body.md` remains unchanged.
The verifier removed its temporary tool log after recording the result.
