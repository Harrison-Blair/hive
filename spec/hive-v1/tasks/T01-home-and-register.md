# T01 home and register

| Item | Value |
| --- | --- |
| Task ID | T01-home-and-register |
| Wave | 1 |
| Tier | strong |
| Depends on | none |
| Blocks | CP1, and through CP1 each task of wave 2 |
| Branch | `task/T01-home-and-register` |
| Worktree | `projects/hive/worktrees/T01-home-and-register/` |
| Evidence | `evidence/T01-home-and-register/` |

## Overview

This task is the tracer bullet at the root of the plan. It builds the thin path
from a command line to a record in the state store and a log event. It gives
the home resolution, the state store, the log writer, the JSON envelope, and
the three `project` commands. Each later task builds on these five packages.
The shape of these packages matters more than their size, so the tier is
strong. The task adds no herdr call and no external process.

## User flows

1. The user clones the repository and runs `hive project register . --base dev` from the root. Hive resolves the home from the marker, writes the project record, makes the project directory, and prints one line for each effect.
2. A manager runs `hive --home /home/user/hive project show hive --json` from a project directory. It gets one JSON object with the record in `result`.
3. A worker runs a command from a worktree without `--home`. Hive resolves the outermost marker above the worktree, which is the real home, and acts there.
4. A script runs `hive --home /tmp/nothing project list`. Hive prints `not a hive home: /tmp/nothing` and exits with status 1. It creates nothing.
5. A script runs `hive project show` without an ID. Hive prints one usage error line and exits with status 2.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `main.go` | Set the process exit status from the value that `cmd.Execute` returns: 0, 1, or 2. |
| `cmd/root.go` | Bind the persistent flags `--home` and `--json`. Register the `project` command. Return the exit status of the command to `main.go`. |
| `cmd/common/common.go` | New. The persistent flags, the resolved context (home, json, writers), and the helper that runs an internal function and renders its envelope. |
| `cmd/project/project.go` | New. The `project` parent command. |
| `cmd/project/register/register.go` | New. `hive project register PATH [--id ID] [--base BRANCH]`. |
| `cmd/project/list/list.go` | New. `hive project list`. |
| `cmd/project/show/show.go` | New. `hive project show ID`. |
| `internal/project/doc.go` | New. Package summary and file map. |
| `internal/project/register.go` | New. Options, checks, record write, directory creation, effects. |
| `internal/project/list.go` | New. Read each project record. |
| `internal/project/show.go` | New. Read one project record. |
| `internal/project/render.go` | New. Text rendering of the three results. |
| `internal/lib/home/home.go` | New. Home resolution. |
| `internal/lib/store/doc.go` | New. Package summary and file map. |
| `internal/lib/store/store.go` | New. Open, lock, bounded read, write, update, list, and counters over `.hive/state/`. |
| `internal/lib/store/atomic.go` | New. Atomic file write by temporary file and rename. |
| `internal/lib/records/doc.go` | New. Package summary and file map. One file for each record kind. |
| `internal/lib/records/project.go` | New. The project record type. |
| `internal/lib/log/log.go` | New. The log writer with the daily file change. |
| `internal/lib/output/output.go` | New. The envelope, the effect type, text and JSON rendering, and the exit status. |

## Requirements

### AC1 Home resolution

- `home.Resolve` takes the `--home` value, the `HIVE_HOME` value, and the working directory. It returns one absolute path without symbolic links, or an error.
- The order: the `--home` value wins. Then `HIVE_HOME`. Then the outermost ancestor of the working directory, the directory itself included, that holds a regular file `.hive/README.md`.
- A `--home` or `HIVE_HOME` path without the marker fails with `not a hive home: <path>`.
- When `HIVE_HOME` names a home and the outermost marker above the working directory is in a different home, the command fails. The message is `home conflict: HIVE_HOME is <a>, working directory is in <b>`.
- A `--home` value skips the conflict check. A worker in a worktree passes `--home` to act on the worktree as its own home during checks.
- With no `--home`, no `HIVE_HOME`, and no marker above the working directory, the command fails with `not a hive home: no marker above <cwd>`.
- Hive never creates a marker or a home.

### AC2 Project register

- `hive project register PATH [--id ID] [--base BRANCH]` makes PATH absolute and resolves its symbolic links. PATH must hold `.git` as a directory or a file. Otherwise the command fails with `not a git repository: <path>`.
- The project ID is the `--id` value, else the basename of the resolved path in lowercase. It must match `^[a-z0-9]+(-[a-z0-9]+)*$` and have at most 18 characters. Otherwise the command fails with `invalid project id: <id>` and prints the rule. The bound keeps each herdr agent name of T04 within 32 characters.
- The base branch is the `--base` value, else `main`. This task does not check that the branch exists. T09 adds the check.
- A second register with the same ID fails with `project exists: <id>`. A register of a path that another ID holds fails with `path registered as <id>`.
- On success hive writes the record with `id`, `path`, `base`, and `registered_at`. Then it makes `projects/<id>/`, `projects/<id>/worktrees/`, and `projects/<id>/overrides/`, in this order. A directory that exists stays.
- The effects list one entry for the record and one for each directory. The `action` is `created` or `exists`.
- When a directory step fails after the record write, the outcome is `partial`. The effects list the record, `phase` is `dirs`, and the error names the directory.

### AC3 Project list and show

- `hive project list` prints one line for each project: ID, path, base, in ID order. With no project it prints `no projects`.
- `hive project show ID` prints one line for each field of the record as `<field>: <value>`. An unknown ID fails with `project not found: <id>`.
- With `--json`, `result` holds the array of records for `list` and the record for `show`. With no project the array is `[]`, not null.

### AC4 State store

- `store.Open(home)` uses `.hive/state/`. It makes the directory on first use.
- A mutation runs inside one `flock` on `.hive/state/lock`. The lock covers the read, the change, and the write of the whole mutation. `store.Update(kind, id, fn)` runs the three steps under the lock. A reader of one record needs no lock.
- A record lives at `.hive/state/<kind>/<id>.json` as one indented JSON object. A write goes to a temporary file in the same directory, then renames it over the target. No reader sees a partial file.
- Each read has a size bound. A record file above 1 MiB fails with `record too large: <path>`. A file that does not parse as one JSON object fails with `record invalid: <path>`. `List` returns the first such error and no records.
- `store.Next(name)` returns the next value of the named counter under the lock. The first value is 1. Counters live at `.hive/state/counters/<name>.json`.
- Two processes that mutate the same record in parallel lose no update.

### AC5 Log events

- `log.Open(home, component)` returns a `*slog.Logger` that writes JSON lines to `.hive/logs/YYYY-MM-DD.jsonl` with the local date. It opens the file with `O_APPEND` and mode 0644, and makes `.hive/logs/` on first use.
- The writer takes the local date of each event. When that date differs from the date of the open file, it closes the file and opens the file of the new date. A clock function is an explicit dependency of the writer, so a test can move the date.
- The level is debug. Each event is one `write` call, so events from several processes do not interleave.
- Each event carries `time`, `level`, `msg`, and `component`. A CLI command uses the component `cli` and adds `command`. The event that ends a command adds `duration_ms` and `outcome`.
- `log.Body(text)` returns the attribute for a body. A body above 4096 bytes ends at 4096 bytes, and the event gets `truncated: true`.
- Each `project` command logs one `command start` event and one `command end` event.

### AC6 Command output

- Without `--json`, a command prints text for people to stdout. An error goes to stderr as `hive: <message>`.
- With `--json`, a command prints exactly one JSON object to stdout with `outcome`, `phase`, `effects`, `error`, and `result`. Nothing else goes to stdout.
- `outcome` is `ok`, `failed`, or `partial`. `phase` names the step where the command stopped, such as `home`, `store`, `dirs`, or `render`. `effects` lists what changed, also on failure. `error` is a string or null.
- The process exits with 0 for `ok` and with 1 for `failed` and `partial`. A usage error, such as a missing argument or an unknown flag, prints one line to stderr and exits with 2.

### AC7 Gate

- The four gate commands pass from the root of the worktree.
- `internal/` does not import Cobra. `cmd/common` does not import `cmd`.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | `home.Resolve` follows the order and the error rules of the requirements. | V1, V2 |
| AC2 | `project register` writes the record, makes the directories, refuses duplicates, and reports a partial outcome. | V3 |
| AC3 | `project list` and `project show` print text and JSON as the requirements give. | V4 |
| AC4 | Parallel mutations lose no update, no reader sees a partial record, and each read has a size bound. | V5 |
| AC5 | Log events are JSON lines with the fields, at debug level, with the body cap and the daily file change. | V6 |
| AC6 | The envelope has the five fields, and the exit status follows the outcome. | V2, V3, V4 |
| AC7 | The gate passes. | V7 |

## Verification items

The command items V2, V3, and V4 run in this order from the worktree root. The
setup of V2 removes the state of an earlier run, so the sequence repeats with
the same output. The cleanup of V4 removes each fixture.

### V1 Home resolution tests

- Proves: AC1.
- Command: `go test ./internal/lib/home/ -run TestResolve -v`
- Red run: the test fails on its assertions after the worker adds `Resolve` with an empty body.
- Expected green run: each case passes. The cases: the flag wins, the env wins over the ancestor, and the outermost marker wins over a nested marker. The flag also wins without a conflict check when the working directory is in another home. A named path without a marker fails, a conflict fails, and no marker fails.

### V2 Home resolution at the command line

- Proves: AC1, AC6.
- Setup: `go build . && rm -rf .hive/state projects .hive/tmp/repo-one .hive/tmp/repo-two`
- Command:

```sh
./hive --home /tmp/not-a-home project list; echo "exit $?"
./hive --home /tmp/not-a-home --json project list; echo "exit $?"
HIVE_HOME="$PWD" ./hive project list; echo "exit $?"
./hive --home "$PWD" project list; echo "exit $?"
./hive --home "$PWD" --json project list; echo "exit $?"
./hive --home "$PWD" project show; echo "exit $?"
```

- Expected: line one prints `hive: not a hive home: /tmp/not-a-home` and exit 1. Line two prints one JSON object with `outcome` `failed`, `phase` `home`, and the same error, then exit 1. Line three fails with `home conflict`, because the outermost marker above the worktree is the real home. Line four prints `no projects` and exit 0. Line five prints the envelope with `result` `[]` and exit 0. Line six prints one usage line on stderr and exit 2.

### V3 Register tests and transcript

- Proves: AC2, AC6.
- Command: `go test ./internal/project/ -run TestRegister -v`, red and green. The cases cover the ID rule, with `a-b-c`, `ab_c`, `-ab`, and a 19-character ID.
- Then:

```sh
git init -q .hive/tmp/repo-one
./hive --home "$PWD" project register .hive/tmp/repo-one --base main; echo "exit $?"
ls projects/repo-one
./hive --home "$PWD" project register .hive/tmp/repo-one; echo "exit $?"
./hive --home "$PWD" project register .hive/tmp/repo-one --id other; echo "exit $?"
./hive --home "$PWD" project register .hive/tmp/repo-one --id Repo_One; echo "exit $?"
git init -q .hive/tmp/repo-two
touch projects/repo-two
./hive --home "$PWD" --json project register .hive/tmp/repo-two; echo "exit $?"
```

- Expected: four effect lines, one for the record and one for each directory, and exit 0. `ls` prints `overrides` and `worktrees`. The second register fails with `project exists: repo-one` and exit 1. The third fails with `path registered as repo-one` and exit 1. The fourth fails with `invalid project id: Repo_One` and exit 1. The last prints an envelope with `outcome` `partial`, `phase` `dirs`, one effect for the record `repo-two`, an error that names `projects/repo-two`, and exit 1.

### V4 List and show

- Proves: AC3, AC6.
- Command:

```sh
./hive --home "$PWD" project list
./hive --home "$PWD" --json project list
./hive --home "$PWD" project show repo-one
./hive --home "$PWD" --json project show repo-one
./hive --home "$PWD" --json project show nope; echo "exit $?"
```

- Expected: two text lines, `repo-one` with its path and `main`, then `repo-two`. Then an envelope whose `result` is an array of the two records. Then four lines `id: repo-one`, `path: <path>`, `base: main`, and `registered_at: <time>`. Then one JSON object with `outcome` `ok` and the record in `result`. Then one JSON object with `outcome` `failed` and the error `project not found: nope`, and exit 1.
- Cleanup: `rm -rf .hive/state projects .hive/tmp/repo-one .hive/tmp/repo-two`

### V5 Store concurrency and bounds

- Proves: AC4.
- Command: `go test -race ./internal/lib/store/ -run 'TestConcurrent|TestAtomic|TestUpdate|TestRead' -v`, red and green.
- Expected: the concurrency test starts several OS processes, the test binary again with an environment flag, that each increment the same counter 50 times. The final value equals the total. The update test starts several OS processes that each add 1 to a field of one record 50 times with `Update`. The final field equals the total. The atomic test replaces one record 200 times while a reader goroutine reads it without pause. Each read gives a complete object with either the old or the new value, and no temporary file remains. The read test gives `record too large` for a file of 1 MiB plus one byte and `record invalid` for a file with a truncated object.

### V6 Log events

- Proves: AC5.
- Command: `go test ./internal/lib/log/ -v`, red and green. Then:

```sh
./hive --home "$PWD" project list >/dev/null
tail -n 2 .hive/logs/"$(date +%F)".jsonl
```

- Expected: the test shows a 5000-byte body cut at 4096 bytes with `truncated: true`. After parallel writers the line count equals the event count. With a test clock, one event before midnight and one after it land in the two daily files. The tail shows the `command start` and `command end` events with `component` `cli`, `command`, and `duration_ms`.

### V7 Gate

- Proves: AC7.
- Command:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run ./...
go test -race ./...
```

- Expected: the first command prints nothing. The other three pass.

## Out of scope

- Any herdr call and any `git` process. T05 adds the bounded process runner.
- `--json` for `hive version`. T05 adds it.
- Records other than the project record and the counters.

## Changes

- 2026-10-17, untrusted inputs. These rules add to AC1, AC2, AC3, and AC4. V1, V3, and V5 prove them with tests that their `-run` patterns select.
  - An empty `HIVE_HOME` counts as unset. An empty `--home` value or an empty `PATH` is a usage error with exit status 2.
  - A relative `--home`, `HIVE_HOME`, or `PATH` resolves against the working directory. Hive does not trim spaces around a value. The conflict check compares resolved paths.
  - A `--base` value that is empty, starts with `-`, or holds a space or a control character fails with `invalid base branch: <value>`. A `PATH` with a control character fails with `invalid path: <value>` before any file access. Hive prints such a value with `%q`.
  - `project show` applies the ID rule before it reads a file. An ID that fails the rule gives `invalid project id: <id>`.
  - The store refuses a kind, an ID, or a counter name that is empty or holds `/` or `..`. The temporary file of an atomic write starts with `.`. `List` reads only files with the name `<id>.json` where `<id>` passes the ID rule.
  - A directory, a symbolic link, an empty file, or a JSON array in place of a record gives `record invalid: <path>`. The reader does not follow a symbolic link. A record of exactly 1 MiB is valid.
  - One function in the store reads a record file, and one function writes one. The write function can call the atomic helper of `atomic.go`. No other code in the package opens a record file. V5 adds one block with the command `grep -nE 'os\.(Open|OpenFile|ReadFile|WriteFile|Create|CreateTemp)\(' internal/lib/store/*.go` and its output.
