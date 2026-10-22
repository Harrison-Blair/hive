# 7-pr-dev-main VERIFIED

| Item | Value |
| --- | --- |
| Task | 7-pr-dev-main |
| Commit | 8d50fb27 |
| Date | 2026-10-19 |
| Agent | verifier, pi, opencode-go/deepseek-v4.1-flash |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | Base `main`, head `dev`, state `OPEN`, title `Hive v1: replace fledge with hive`, and the six items of R1 in the body. One open pull request has base `main` and head `dev`. |
| V2 | AC2 | pass | `origin/main` is `928cfa95e0906315fcb4a65a5cdbdef1470a0503`. `dev..task/7-pr-dev-main` holds two commits, and their diff adds only evidence files. No source file changed. |

## Runs

All commands ran from the root of the worktree at
`/home/penguin/source/hive/projects/hive/worktrees/7-pr-dev-main`.

### V1

```sh
gh pr view 47 --json url,baseRefName,headRefName,title,body,state
echo "exit $?"
```

```text
{"baseRefName":"main","body":"## What is Hive\n\nHive is a Go CLI and TUI that coordinates coding agents in herdr from one home. Managers own projects, and workers carry out one task each in separate worktrees. This pull request replaces the old fledge implementation on `main` with Hive's reauthored implementation.\n\n## Specification\n\n[Hive v1 specification](https://github.com/Harrison-Blair/hive/blob/dev/spec/hive-v1/README.md)\n\n## Landed tasks T01–T21\n\n- T01 — home and register (`3ab084c9`)\n- T02 — version lib (`d8d6fc8b`)\n- T03 — task lifecycle (`47da8abb`)\n- T04 — roles and inbox (`19e41ee6`)\n- T05 — herdr client and doctor (`b8cb5f71`)\n- T06 — logs command (`84da9819`)\n- T07 — memory and handoff (`87da8c9b`)\n- T08 — TUI log view (`cd90451e`)\n- T09 — worktree create and remove (`a5cfd700`)\n- T10 — agent spawn core (`37652687`)\n- T11 — task identity and finish (`c4a09584`)\n- T12 — land (`941f1441`)\n- T13 — manager spawn (`82f2e307`)\n- T14 — worker spawn (`9a1aa92a`)\n- T15 — supervise (`660529ca`)\n- T16 — codex profile (`3f607ba9`)\n- T17 — opencode profile (`cb95422b`)\n- T18 — pi profile (`80787426`)\n- T19 — cursor profile (`976d2736`)\n- T20 — recovery (`7046bfe5`)\n- T21 — role playbooks (`175af5eb`)\n\n## Sidebar layout wave\n\n- `3-manager-adopts-parent`: managers adopt their parent project workspace (`61dcbb0a`).\n- `4-shell-wait-friction`: agent spawn waits until panes are ready shells (`0643bf3f`, `af29c592`).\n- `5-tab-labels`: agent tabs receive Hive labels (`7121ac04`).\n\n## Gate on dev at `18111ae1`\n\n- `gofmt -l $(git ls-files -co --exclude-standard '*.go')` — no output.\n- `go vet ./...` — passed (exit 0).\n- `golangci-lint run ./...` — passed (`0 issues`).\n- `go test -race ./...` — passed for all packages.\n\nOpened under `grants.md 2026-10-19 pull request from dev into main`. The owner merges.\n","headRefName":"dev","state":"OPEN","title":"Hive v1: replace fledge with hive","url":"https://github.com/Harrison-Blair/hive/pull/47"}
exit 0
```

The list of pull requests with base `main` and head `dev`:

```sh
gh pr list --base main --head dev --json number,url,state --jq '.[] | "\(.number) \(.state) \(.url)"'
echo "exit $?"
```

```text
47 OPEN https://github.com/Harrison-Blair/hive/pull/47
exit 0
```

The body, in readable form:

```sh
gh pr view 47 --json body --jq .body
echo "exit $?"
```

```text
## What is Hive

Hive is a Go CLI and TUI that coordinates coding agents in herdr from one home. Managers own projects, and workers carry out one task each in separate worktrees. This pull request replaces the old fledge implementation on `main` with Hive's reauthored implementation.

## Specification

[Hive v1 specification](https://github.com/Harrison-Blair/hive/blob/dev/spec/hive-v1/README.md)

## Landed tasks T01–T21

- T01 — home and register (`3ab084c9`)
- T02 — version lib (`d8d6fc8b`)
- T03 — task lifecycle (`47da8abb`)
- T04 — roles and inbox (`19e41ee6`)
- T05 — herdr client and doctor (`b8cb5f71`)
- T06 — logs command (`84da9819`)
- T07 — memory and handoff (`87da8c9b`)
- T08 — TUI log view (`cd90451e`)
- T09 — worktree create and remove (`a5cfd700`)
- T10 — agent spawn core (`37652687`)
- T11 — task identity and finish (`c4a09584`)
- T12 — land (`941f1441`)
- T13 — manager spawn (`82f2e307`)
- T14 — worker spawn (`9a1aa92a`)
- T15 — supervise (`660529ca`)
- T16 — codex profile (`3f607ba9`)
- T17 — opencode profile (`cb95422b`)
- T18 — pi profile (`80787426`)
- T19 — cursor profile (`976d2736`)
- T20 — recovery (`7046bfe5`)
- T21 — role playbooks (`175af5eb`)

## Sidebar layout wave

- `3-manager-adopts-parent`: managers adopt their parent project workspace (`61dcbb0a`).
- `4-shell-wait-friction`: agent spawn waits until panes are ready shells (`0643bf3f`, `af29c592`).
- `5-tab-labels`: agent tabs receive Hive labels (`7121ac04`).

## Gate on dev at `18111ae1`

- `gofmt -l $(git ls-files -co --exclude-standard '*.go')` — no output.
- `go vet ./...` — passed (exit 0).
- `golangci-lint run ./...` — passed (`0 issues`).
- `go test -race ./...` — passed for all packages.

Opened under `grants.md 2026-10-19 pull request from dev into main`. The owner merges.
exit 0
```

The 21 task commits, in the order of the body:

```sh
for s in 3ab084c9 d8d6fc8b 47da8abb 19e41ee6 b8cb5f71 84da9819 87da8c9b cd90451e a5cfd700 37652687 c4a09584 941f1441 82f2e307 9a1aa92a 660529ca 3f607ba9 cb95422b 80787426 976d2736 7046bfe5 175af5eb; do git log -1 --format="$s %s" $s; done
echo "exit $?"
```

```text
3ab084c9 merge(T01-home-and-register): land home and register
d8d6fc8b merge(T02-version-lib): land version lib
47da8abb merge(T03): land the task lifecycle fixes
19e41ee6 merge(T04): land roles and inbox
b8cb5f71 merge(T05): land herdr client and doctor
84da9819 merge(T06): land logs command
87da8c9b merge(T07): land memory and handoff
cd90451e merge(T08): land TUI log view
a5cfd700 merge(T09): land worktree create and remove
37652687 merge(T10): land agent spawn core
c4a09584 merge(T11): land task identity and finish
941f1441 merge(T12): land the land command
82f2e307 merge(T13): land the manager spawn
9a1aa92a merge(T14): land the worker spawn
660529ca merge(T15): land the supervisor
3f607ba9 merge(T16): land the codex profile
cb95422b merge(T17): land the opencode profile
80787426 merge(T18): land the pi profile
976d2736 merge(T19): land the cursor profile
7046bfe5 merge(T20): land the recovery commands
175af5eb merge(T21): land the role playbooks
exit 0
```

The three sidebar layout tasks are in `dev`:

```sh
for c in 61dcbb0a 0643bf3f af29c592 7121ac04; do git log -1 --format="%h %s" $c; git merge-base --is-ancestor $c 18111ae1 && echo "$c in dev"; done
echo "exit $?"
```

```text
61dcbb0a feat(agent): adopt parent workspace for managers
61dcbb0a in dev
0643bf3f fix(agent): wait for panes to become ready shells
0643bf3f in dev
af29c592 fix(agent): detect shell startup processes
af29c592 in dev
7121ac04 feat(agent): label hive agent tabs
7121ac04 in dev
exit 0
```

### Gate at the head of `dev`

The PR head is `18111ae1`. The worktree switched to that commit for the four
gate commands and returned to the task branch.

```sh
git switch --detach 18111ae1
git rev-parse HEAD
```

```text
18111ae1d4b6f42180de64cfd470d178767a24e4
```

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
```

```text
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
ok  	github.com/Harrison-Blair/hive	(cached)
ok  	github.com/Harrison-Blair/hive/cmd	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	(cached)
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	(cached)
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/process	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/records	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/run	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

```sh
git switch task/7-pr-dev-main
git status --short --branch
```

```text
## task/7-pr-dev-main...origin/task/7-pr-dev-main
```

### V2

```sh
git ls-remote origin main
echo "exit $?"
```

```text
928cfa95e0906315fcb4a65a5cdbdef1470a0503	refs/heads/main
exit 0
```

```sh
git log --oneline dev..task/7-pr-dev-main
echo "exit $?"
```

```text
8d50fb27 docs(evidence): record unchanged main
1f2038db docs(evidence): record PR verification
exit 0
```

```sh
git diff --name-status dev..task/7-pr-dev-main
echo "exit $?"
```

```text
A	evidence/7-pr-dev-main/V1.md
A	evidence/7-pr-dev-main/V2.md
exit 0
```
