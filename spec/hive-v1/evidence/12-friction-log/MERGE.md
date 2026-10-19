# Merge dev into task/12-friction-log

The only conflicted path was `testdata/ste-baseline.json`. The resolution keeps both task entries in key order: `11-ci-workflow.md`, then `12-friction-log.md`. After resolution, `git diff --name-only --diff-filter=U` printed no paths.

## Gate

### Formatting

Command:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
```

Output:

```text
exit 0
```

The command printed no file paths.

### Vet

Command:

```sh
go vet ./...
```

Output:

```text
exit 0
```

### Lint

Command:

```sh
golangci-lint run ./...
```

Output:

```text
0 issues.
exit 0
```

### Race tests

Command:

```sh
go test -race ./...
```

Output:

```text
ok   github.com/Harrison-Blair/hive 3.821s
ok   github.com/Harrison-Blair/hive/cmd 1.411s
?    github.com/Harrison-Blair/hive/cmd/agent [no test files]
?    github.com/Harrison-Blair/hive/cmd/agent/list [no test files]
ok   github.com/Harrison-Blair/hive/cmd/agent/register 1.027s
ok   github.com/Harrison-Blair/hive/cmd/agent/resolve 1.301s
?    github.com/Harrison-Blair/hive/cmd/agent/send [no test files]
?    github.com/Harrison-Blair/hive/cmd/agent/show [no test files]
ok   github.com/Harrison-Blair/hive/cmd/agent/spawn 1.029s
?    github.com/Harrison-Blair/hive/cmd/agent/stop [no test files]
?    github.com/Harrison-Blair/hive/cmd/common [no test files]
?    github.com/Harrison-Blair/hive/cmd/doctor [no test files]
?    github.com/Harrison-Blair/hive/cmd/handoff [no test files]
ok   github.com/Harrison-Blair/hive/cmd/handoff/write 1.198s
?    github.com/Harrison-Blair/hive/cmd/inbox [no test files]
?    github.com/Harrison-Blair/hive/cmd/land [no test files]
?    github.com/Harrison-Blair/hive/cmd/logs [no test files]
?    github.com/Harrison-Blair/hive/cmd/memory [no test files]
ok   github.com/Harrison-Blair/hive/cmd/memory/write 1.202s
?    github.com/Harrison-Blair/hive/cmd/project [no test files]
?    github.com/Harrison-Blair/hive/cmd/project/list [no test files]
?    github.com/Harrison-Blair/hive/cmd/project/register [no test files]
?    github.com/Harrison-Blair/hive/cmd/project/show [no test files]
?    github.com/Harrison-Blair/hive/cmd/supervise [no test files]
?    github.com/Harrison-Blair/hive/cmd/task [no test files]
?    github.com/Harrison-Blair/hive/cmd/task/assign [no test files]
?    github.com/Harrison-Blair/hive/cmd/task/cancel [no test files]
?    github.com/Harrison-Blair/hive/cmd/task/complete [no test files]
?    github.com/Harrison-Blair/hive/cmd/task/create [no test files]
?    github.com/Harrison-Blair/hive/cmd/task/list [no test files]
?    github.com/Harrison-Blair/hive/cmd/task/show [no test files]
?    github.com/Harrison-Blair/hive/cmd/task/verify [no test files]
ok   github.com/Harrison-Blair/hive/cmd/version 1.012s
?    github.com/Harrison-Blair/hive/cmd/worktree [no test files]
?    github.com/Harrison-Blair/hive/cmd/worktree/create [no test files]
?    github.com/Harrison-Blair/hive/cmd/worktree/remove [no test files]
ok   github.com/Harrison-Blair/hive/internal/agent (cached)
ok   github.com/Harrison-Blair/hive/internal/doctor 3.033s
ok   github.com/Harrison-Blair/hive/internal/handoff (cached)
ok   github.com/Harrison-Blair/hive/internal/inbox (cached)
ok   github.com/Harrison-Blair/hive/internal/land (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/gitx (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/harness (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/herdr (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/home (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/identity 1.031s
ok   github.com/Harrison-Blair/hive/internal/lib/inbox (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/log (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/process (cached)
?    github.com/Harrison-Blair/hive/internal/lib/output [no test files]
?    github.com/Harrison-Blair/hive/internal/lib/records [no test files]
ok   github.com/Harrison-Blair/hive/internal/lib/run (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/scope (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/store (cached)
ok   github.com/Harrison-Blair/hive/internal/lib/version (cached)
ok   github.com/Harrison-Blair/hive/internal/logs (cached)
ok   github.com/Harrison-Blair/hive/internal/memory (cached)
ok   github.com/Harrison-Blair/hive/internal/project (cached)
ok   github.com/Harrison-Blair/hive/internal/supervise (cached)
ok   github.com/Harrison-Blair/hive/internal/task (cached)
ok   github.com/Harrison-Blair/hive/internal/tui (cached)
ok   github.com/Harrison-Blair/hive/internal/worktree (cached)
exit 0
```
