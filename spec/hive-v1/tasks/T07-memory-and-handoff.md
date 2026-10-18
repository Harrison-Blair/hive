# T07 memory and handoff

| Item | Value |
| --- | --- |
| Task ID | T07-memory-and-handoff |
| Wave | 2 |
| Tier | small |
| Depends on | T01, CP1 |
| Blocks | T21 |
| Branch | `task/T07-memory-and-handoff` |
| Worktree | `projects/hive/worktrees/T07-memory-and-handoff/` |
| Evidence | `evidence/T07-memory-and-handoff/` |

## Overview

Agents compose memory and handoff text. Hive writes the files atomically at
the correct scope and keeps the memory index current. This task gives the two
write commands. It reads stdin with a size bound and a deadline, checks the
frontmatter of a memory, and regenerates `MEMORY.md` under the store lock.

## User flows

1. The coordinator runs `hive memory write --scope home user-prefers-tables < fact.md`. Hive writes `memory/user-prefers-tables.md` and adds one line to `memory/MEMORY.md`.
2. A manager runs `hive memory write --scope project:hive base-is-dev < fact.md`. Hive writes under `projects/hive/memory/`.
3. The coordinator ends its session with `hive handoff write --scope home < handoff.md`. Hive replaces `handoff.md` atomically.
4. A manager passes `--scope project:nope`. Hive fails with `project not found: nope`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Register the `memory` and `handoff` commands. |
| `cmd/memory/memory.go`, `cmd/memory/write/write.go` | New. |
| `cmd/handoff/handoff.go`, `cmd/handoff/write/write.go` | New. |
| `internal/memory/doc.go`, `write.go`, `index.go`, `frontmatter.go` | New. The memory write, the index generation, and the frontmatter check. |
| `internal/handoff/write.go` | New. The handoff write. |
| `internal/lib/scope/scope.go` | New. Parse `home` or `project:<id>` and resolve its directory. T21 playbooks name the same scopes. |
| `internal/lib/scope/input.go` | New. The bounded stdin reader with a size limit and a deadline. Both commands use it. |

## Requirements

### AC1 Scope and input

- A scope is `home` or `project:<id>`. `home` resolves to the home root. `project:<id>` resolves to `projects/<id>/` and needs a registered project. Any other value fails with `bad scope: <value>`.
- Both commands read stdin up to 1 MiB and fail above it with `input too large`. The read has a deadline of 30 seconds. On expiry the command fails with `input timeout`. Both failures write nothing. The deadline is an explicit parameter of the reader, so a test can shorten it.

### AC2 Memory write

- `hive memory write --scope SCOPE NAME` takes NAME that matches `^[a-z0-9][a-z0-9-]*$`.
- The content starts with a frontmatter block: a line `---`, then `key: value` lines, then a line `---`. A value is the text after the first `: `, trimmed, on one line. There is no quoting and no multi-line value. The keys `name` and `description` must exist once each with a value. Other keys pass without effect. A standard library parser reads the block, so the task adds no dependency.
- The errors: `frontmatter: no block`, `frontmatter: missing <key>`, `frontmatter: empty <key>`, `frontmatter: duplicate <key>`, and `frontmatter: name <value> differs from <NAME>`.
- Before any write, hive reads each existing `*.md` file in the memory directory except the index, with the 1 MiB bound, and checks its frontmatter. An invalid or oversized file fails with `memory file invalid: <file>: <reason>`, and nothing changes.
- Hive writes `<scope dir>/memory/<name>.md` atomically. It makes the directory on first use.
- Then hive regenerates `<scope dir>/memory/MEMORY.md`. One line for each file: `- [<name>](<name>.md): <description>`, in name order. The read of the existing files, the write, the scan, and the index write run under the store lock. Thus two parallel writes give an index with both entries.
- The effects list the memory file and the index.

### AC3 Handoff write

- `hive handoff write --scope SCOPE` writes the input to `<scope dir>/handoff.md` atomically. It replaces the previous content. Empty input fails with `empty input`.
- The effect lists the file.

### AC4 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Scopes parse and resolve as the requirements give, and the input reader bounds size and time. | V1 |
| AC2 | A memory write checks the frontmatter and the existing files, writes the file, and regenerates the index under the lock. | V1, V3 |
| AC3 | A handoff write replaces the file atomically and refuses empty or oversized input. | V2, V3 |
| AC4 | The gate passes. | V4 |

## Verification items

### V1 Memory tests

- Proves: AC1, AC2.
- Command: `go test -race ./internal/memory/ ./internal/lib/scope/ -v`, red and green.
- Expected: cases for each scope value and each frontmatter error. A 1 MiB plus one byte input fails with `input too large`. A pipe that sends no end of file fails with `input timeout` under a 100-millisecond deadline, and the directory stays unchanged. An existing file with a bad block fails with `memory file invalid`, and the previous index stays. After two writes the index has two entries in name order. Two OS processes, the test binary again with an environment flag, write two memories at the same time, and the final index holds both.

### V2 Handoff tests

- Proves: AC3.
- Command: `go test -race ./internal/handoff/ -v`, red and green.
- Expected: a writer replaces the handoff 200 times while a reader goroutine reads it without pause. Each read gives the complete old or the complete new content, and no temporary file remains. Empty input fails. A 1 MiB plus one byte input fails with `input too large`, and the previous content stays.

### V3 Command transcript

- Proves: AC2, AC3.
- Setup: `go build . && rm -rf memory handoff.md`
- Command:

```sh
printf -- '---\nname: base-is-dev\ndescription: The base branch of hive is dev.\n---\nThe base branch of hive is dev.\n' | ./hive --home "$PWD" memory write --scope home base-is-dev
cat memory/MEMORY.md
printf -- '---\nname: other\n---\nNo description.\n' | ./hive --home "$PWD" memory write --scope home other; echo "exit $?"
printf 'Open decisions: none.\n' | ./hive --home "$PWD" handoff write --scope home
cat handoff.md
```

- Expected: two effect lines, then an index with one line for `base-is-dev`. The second write fails with `frontmatter: missing description` and exit 1. Then one effect line for the handoff, and the handoff content.
- Cleanup: `rm -rf memory handoff.md`

### V4 Gate

- Proves: AC4.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- A read or a search command for memory. The index file is the list.

## Changes

None.
