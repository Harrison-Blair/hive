# T07 memory and handoff

| Item | Value |
| --- | --- |
| Task ID | T07-memory-and-handoff |
| Wave | 2 |
| Tier | small |
| Depends on | T01, T22, CP1 |
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

## Store use

Use one `store.WithLock` call around the existing-file read, the memory write, the
scan, and the index write. Do not hold the lock while stdin reads. Read stdin
before the lock. The store has no Markdown writer, so use one local atomic
Markdown helper.

Method matrix: `Store.Read` and `Store.ReadIn` take no lock. A standalone list
takes no lock. A related read uses one `WithLock` callback. Each standalone
mutation uses bounded acquisition. A `LockedStore` method never acquires another
lock.

## Requirements

### AC1 Scope and input

- A scope is `home` or `project:<id>`. `home` resolves to the home root. `project:<id>` resolves to `projects/<id>/` and needs a registered project. Any other value fails with `bad scope: <value>`.
- The scope directory and each Markdown path stay inside the home. Check the scope root, each ancestor, and the target with `Lstat`. Reject a static symbolic link or a wrong type before a read or a write. A static directory or symbolic link at `MEMORY.md` is a no-write refusal. Hostile replacement between `Lstat` and `Open` stays deferred.
- Both commands read stdin up to 1 MiB and fail above it with `input too large`. The read has a deadline of 30 seconds. On expiry the command fails with `input timeout`. Both failures write nothing. The deadline is an explicit parameter of the reader, so a test can shorten it.

### AC2 Memory write

- `hive memory write --scope SCOPE NAME` takes NAME that matches `^[a-z0-9][a-z0-9-]*$`.
- The content starts with a frontmatter block: a line `---`, then `key: value` lines, then a line `---`. A value is the text after the first `: `, trimmed, on one line. There is no quoting and no multi-line value. The keys `name` and `description` must exist once each with a value. Other keys pass without effect. A standard library parser reads the block, so the task adds no dependency.
- The errors: `frontmatter: no block`, `frontmatter: missing <key>`, `frontmatter: empty <key>`, `frontmatter: duplicate <key>`, and `frontmatter: name <value> differs from <NAME>`.
- Before any write, hive reads each existing `*.md` file in the memory directory except the index, with the 1 MiB bound, and checks its frontmatter. An invalid or oversized file fails with `memory file invalid: <file>: <reason>`, and nothing changes.
- Hive writes `<scope dir>/memory/<name>.md` atomically. It makes the directory on first use. A new file has effect memory `created`. An existing file has effect memory `updated`.
- Then hive regenerates `<scope dir>/memory/MEMORY.md`. One line for each file: `- [<name>](<name>.md): <description>`, in name order. The read of the existing files, the write, the scan, and the index write run under the store lock. Thus two parallel writes give an index with both entries.
- The effects list the memory file and the index.

### AC3 Handoff write

- `hive handoff write --scope SCOPE` writes the input to `<scope dir>/handoff.md` atomically. It replaces the previous content. Empty input fails with `empty input`.
- The effect lists the file.

### Partial state for memory write

The rules in this task file are the sole authority for a worker or verifier. The
safe order for a memory write is:

1. Read stdin with the 1 MiB bound and the 30 second deadline. Check the name and
   the frontmatter before the lock. Check the scope and the Markdown paths. A
   static directory or symbolic link at `MEMORY.md` fails here with no write.
2. Under one lock, read each existing memory file and check it. Write nothing on
   an invalid file.
3. Write `<name>.md` with an atomic write. Effect: memory `created` for a new file
   or memory `updated` for an existing file.
4. Build the index from each memory file, including the new file.
5. Write `MEMORY.md`. Effect: index `updated`.

The memory file is the authority. The index is a derived file. A failure after
the memory file leaves a stale index. For the partial-index test, inject an index
write or rename failure after the successful static checks and the memory write.
Do not use a static directory at `MEMORY.md` for this state. The command
reports `partial` with the actual failed phase and the memory effect. A later
successful memory write regenerates the index. The command does not remove the memory file and does not
restore the old index. It returns a nil Go error for a partial outcome and builds
an `output.Envelope` with outcome `partial`.

A handoff write replaces one file with one atomic write. A failure before the
rename keeps the old content. No partial state exists. Empty input and oversized
input write nothing.

### AC4 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Scopes parse and resolve as the requirements give, and the input reader bounds size and time. | V1 |
| AC2 | A memory write checks the frontmatter and the existing files. It writes the file with a `created` or `updated` effect and regenerates the index under the lock with a reported partial state. | V1, V3 |
| AC3 | A handoff write replaces the file atomically and refuses empty or oversized input. | V2, V3 |
| AC4 | The gate passes. | V4 |

## Verification items

### V1 Memory tests

- Proves: AC1, AC2.
- Command: `go test -race ./internal/memory/ ./internal/lib/scope/ -v`, red and green.
- Expected: cases for each scope value and each frontmatter error. A 1 MiB plus one byte input fails with `input too large`. A pipe that sends no end of file fails with `input timeout` under a 100-millisecond deadline, and the directory stays unchanged. An existing file with a bad block fails with `memory file invalid`, and the previous index stays. After two writes the index has two entries in name order. A new memory file reports `created`. A replacement of the same name reports `updated`. Two OS processes, the test binary again with an environment flag, write two memories at the same time, and the final index holds both.
- Expected: a static directory or symbolic link at `MEMORY.md` fails before the memory write and changes nothing. This is a no-write refusal, not the partial-index case.
- Expected: an injected index write or rename failure after the memory write gives `partial`, phase `index`. A new memory file has effect memory `created`. A replacement has effect memory `updated`. The fact stays. The old index stays. A later successful write regenerates the index.
- Expected: a static symbol link in the scope root, an ancestor, or a Markdown leaf fails before the read or write.

### V2 Handoff tests

- Proves: AC3.
- Command: `go test -race ./internal/handoff/ -v`, red and green.
- Expected: a writer replaces the handoff 200 times while a reader goroutine reads it without pause. Each read gives the complete old or the complete new content, and no temporary file remains. Empty input fails. A 1 MiB plus one byte input fails with `input too large`, and the previous content stays.

### V3 Command transcript

- Proves: AC2, AC3.
- Setup: Run `go build .` and create a unique temporary home. Create its marker `.hive/README.md`. Set `HIVE_TMP` to its path. Run each CLI command as `HIVE_HOME="$HIVE_TMP" ./hive CMD`. Keep real memory and handoff files.
- Command:

```sh
printf -- '---\nname: base-is-dev\ndescription: The base branch of hive is dev.\n---\nThe base branch of hive is dev.\n' | HIVE_HOME="$HIVE_TMP" ./hive memory write --scope home base-is-dev
cat "$HIVE_TMP"/memory/MEMORY.md
printf -- '---\nname: other\n---\nNo description.\n' | HIVE_HOME="$HIVE_TMP" ./hive memory write --scope home other; echo "exit $?"
printf 'Open decisions: none.\n' | HIVE_HOME="$HIVE_TMP" ./hive handoff write --scope home
cat "$HIVE_TMP"/handoff.md
```

- Expected: two effect lines, then an index with one line for `base-is-dev`. The second write fails with `frontmatter: missing description` and exit 1. Then one effect line for the handoff, and the handoff content.
- Cleanup: remove only `$HIVE_TMP`. Do not remove the worktree `.hive/state`, `.hive/logs`, `.hive/tmp`, a report, or a log.

### V4 Gate

- Proves: AC4.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- A read or a search command for memory. The index file is the list.

## Changes

None.
