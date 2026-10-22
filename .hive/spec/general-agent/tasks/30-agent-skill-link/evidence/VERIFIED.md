# Verification: 30-agent-skill-link

Verdict: pass.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Link | pass | Git tracks the relative link. It resolves to the canonical skill. No test lists the skill links; V1 permits this exception. |
| V-gate Diff scope and gate | pass | Only permitted source files and evidence changed. All five gate commands exited 0. |

## V1 Link

Independent checks:

```text
$ git ls-files .claude/skills
.claude/skills/hive-agent
.claude/skills/hive-coordinator
.claude/skills/hive-manager
.claude/skills/hive-verifier
.claude/skills/hive-worker
$ git ls-files -s .claude/skills/hive-agent
120000 ab2ac36b5594c7607663bde389abd599676cf0e3 0 .claude/skills/hive-agent
$ readlink .claude/skills/hive-agent
../../.agents/skills/hive-agent
```

Read the full `.claude/skills/hive-agent/SKILL.md` through the link with the
file reader. It names `hive-agent` and gives the general agent procedure.
No skill text changed. No UI changed.

Searched the Go tests for `.claude/skills`, `Readlink`, and `EvalSymlinks`.
No test lists the skill links. The skill tests read the canonical paths.
Removed the link and ran the root tests. They passed, as the exception in V1
permits. Restored the link. Removed its allow rule and ran the ignore test.
That test failed for the expected path. Restored the exact original file and
ran the root tests again. They passed. The next section gives each output.

## V-gate Scope

```text
$ git merge --no-edit dev
Already up to date.
$ git diff dev...HEAD --stat
 .claude/skills/hive-agent                          |   1 +
 .gitignore                                         |   1 +
 .../tasks/30-agent-skill-link/brief.md             |  83 +++++++++++++++
 .../tasks/30-agent-skill-link/evidence/V-gate.md   | 115 +++++++++++++++++++++
 .../tasks/30-agent-skill-link/evidence/V1.md       |  68 ++++++++++++
 testdata/ste-baseline.json                         |   1 +
 6 files changed, 269 insertions(+)
```

Each path is in the source table or the evidence folder. The allow rule is
`!.claude/skills/hive-agent`. The new brief has a baseline of 0.
Before this report, `git status --short` and `git diff --check` printed nothing.
The temporary changes from the negative checks are absent.

## V1 Negative and restored runs

```text
$ go test -count=1 .
ok  	github.com/Harrison-Blair/hive	1.704s
exit 0

$ go test -count=1 -run TestTrackedFilesAreNotIgnored .
--- FAIL: TestTrackedFilesAreNotIgnored (0.00s)
    gitignore_test.go:30: .gitignore ignores tracked files; add the narrowest allow rule for each:
        .claude/skills/hive-agent
FAIL
FAIL	github.com/Harrison-Blair/hive	0.007s
FAIL
exit 1

$ go test -count=1 .
ok  	github.com/Harrison-Blair/hive	1.608s
exit 0
```

## V-gate Full gate output

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
exit 0
$ go mod tidy -diff
exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	27.256s
ok  	github.com/Harrison-Blair/hive/cmd	7.636s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.197s
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	1.357s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.560s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.450s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.149s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.372s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.454s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.393s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.110s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.185s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.047s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	16.011s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.727s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	5.741s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	5.100s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.403s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.380s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.034s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.356s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	11.472s
ok  	github.com/Harrison-Blair/hive/internal/tui	77.203s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	7.419s
exit 0
```
