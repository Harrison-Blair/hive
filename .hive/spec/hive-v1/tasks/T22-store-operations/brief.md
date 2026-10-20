# T22 store operations

| Item | Value |
| --- | --- |
| Task ID | T22-store-operations |
| Wave | 2 |
| Tier | strong |
| Depends on | T01, CP1 |
| Blocks | T03, T04, T05, T07 |
| Branch | `task/T22-store-operations` |
| Worktree | `projects/hive/worktrees/T22-store-operations/` |
| Evidence | `evidence/T22-store-operations/` |

## Overview

This task makes the state store safe for more than one record operation. It gives
a public bounded lock, a locked handle, nested collections, a generic list, and a
bounded walk of each record. T03, T04, T05, and T07 use these operations. The
task keeps the JSON files and the one `flock` of T01. It adds no journal, no
rollback, and no dependency.

## Known facts

- T01 merged at `3ab084c9`. The tracker landed T01 at `30e99a95`. The current store is
  `internal/lib/store/store.go`.
- `Store.List` uses `records.ValidID` for each kind. That rule limits a project ID
  to 18 characters. A task ID can be longer.
- `withLock` is private and covers one operation. A caller cannot group two
  operations. T04 and T07 need grouped operations.
- The inbox layout is `inbox/<role>/<seq>.json`. The old kind check refuses a
  slash. The store needs structured path components.
- T05 doctor needs one bounded lock check and a bounded read of each record file.
- The approved trust boundary is a local user and a cooperating writer. A hostile
  file replacement between `Lstat` and `Open` stays deferred.
- Lock acquisition defaults to two seconds. The wait is cancellable. A caller can
  give a longer context deadline. An acquisition failure changes no record.
- A caller must not hold the lock for a model wait, a Git wait, a network wait, or
  a stdin wait.

## User flows

1. T03 writes a task with the ID `12-spawn-worker`. `ListIn` returns that record.
2. T04 writes an inbox record at `inbox/coordinator/1.json`. A register changes
   the agent records and the claim under one lock.
3. T05 doctor gets the lock and walks each record file with a size bound.
4. T07 reads the memory files, writes one memory file, and writes the index under
   one lock.
5. A writer holds the lock. A second writer with a 100 millisecond context fails.
   The second writer changes no record.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/lib/store/store.go` | Add generic read, write, update, and list operations. Refactor the unlocked helpers. Add the bounded walk. |
| `internal/lib/store/lock.go` | New. Public `WithLock`, `LockedStore`, the lock wait, and the two lock errors. |
| `internal/lib/store/store_test.go` | Add tests for names, collections, legacy behavior, grouped reads, and the walk. |
| `internal/lib/store/lock_test.go` | New. Tests for timeout, cancellation, unlock, nested lock, and grouped writes. |
| `internal/lib/store/doc.go` | Add the map line for `lock.go`. |

## Requirements

### AC1 Lock contract

The store uses one method matrix. A method takes one row.

| Method | Lock rule |
| --- | --- |
| `Store.Read`, `Store.ReadIn` | No lock. The store bounds the read. |
| `Store.List`, `Store.ListIn` | No lock. A standalone list sees one directory snapshot. |
| A read of related records | One `WithLock` callback. The callback reads each record through its handle. |
| `Store.Write`, `Store.Update`, `Store.Next`, `Store.WriteIn`, `Store.UpdateIn` | Bounded acquisition. The method waits at most the lock deadline. |
| Each `LockedStore` method | Uses the held lock. It never acquires a lock. |

- `Store.WithLock(ctx context.Context, fn func(*LockedStore) error) error` opens
  `.hive/state/lock` with `O_CREATE|O_RDWR` and mode `0600`.
- The function tries `syscall.LOCK_EX|LOCK_NB` in a loop. The wait between tries
  is at most 50 milliseconds.
- The default wait is two seconds. A context deadline later than two seconds
  replaces the default. A context deadline earlier than two seconds shortens the
  wait.
- A context that ends before the wait returns the context error. `errors.Is`
  finds `context.Canceled` or `context.DeadlineExceeded`.
- A wait that ends before the context returns `ErrLockTimeout`. The error value is
  `store lock timeout`.
- A failure to acquire the lock runs no callback and changes no record.
- The function releases the lock when the callback returns. It releases the lock
  when the callback panics.
- The function returns the callback error unchanged. A successful callback
  returns nil.
- `LockedStore` has no `WithLock` method. A second `Store.WithLock` from inside a
  callback waits and then fails. The callback must not do this.
- A `LockedStore` method after the callback returns gives `ErrLockReleased`. The
  error value is `store lock released`. A nil handle gives the same error.
- The errors are `ErrLockTimeout` and `ErrLockReleased`. Tests use `errors.Is`.

### AC2 Generic record operations

- The new methods are `ReadIn`, `WriteIn`, `UpdateIn`, and `ListIn`.
- Each method takes `collection []string`. The collection has one to eight
  components. Each component matches `^[a-z0-9][a-z0-9_-]*$` and has at most 200
  bytes.
- The `id` matches the same rule. The ID has no 18-character limit.
- `ReadIn` reads `.hive/state/<component>/.../<id>.json` with `readRecordFile`.
  It takes no lock.
- `WriteIn` writes the JSON object to that path with `writeRecordFile` under
  bounded acquisition.
- `UpdateIn` reads, changes, and writes one record under bounded acquisition. It
  uses the same callback shape as `Update`. An error from the callback changes
  nothing.
- `ListIn` reads the immediate directory of the collection. It takes no lock. It
  reads each file with the name `<id>.json` in name order. The method skips a name that
  fails the component rule.
- Before a read, list, or create, the store checks the state root, each collection
  ancestor, the lock, and the selected record path with `Lstat`. A symbolic link,
  a FIFO, a socket, or another wrong type fails with `store path invalid: <path>`.
  A wrong type at a record leaf fails with `record invalid: <path>`.
- The lock file must be a regular file when it exists. A static symbolic link at
  `.hive/state/lock` fails before the open.
- `ListIn` takes `include func(filename string) bool`. The function receives the
  file name with the `.json` suffix. A nil function accepts each valid name. The
  function selects files. It does not replace the name check.
- `ListIn` returns an empty slice for a missing directory. It returns the first
  read error and no records for an invalid file.
- `WalkRecords(ctx context.Context, fn func(StoredRecord) error) error` gets the
  lock and walks each record below `.hive/state`. `StoredRecord` holds
  `Collection []string`, `Name string`, and `Data json.RawMessage`. `Name` is the
  file name, including the `.json` suffix.
- The walk descends each directory. It reads a file whose name ends in `.json` as
  a record. It does not read another file name as a record. A file name with a
  colon is valid, so `counters/inbox:coordinator.json` remains visible.
- The walk rejects a symbolic-link collection or record with its path. It does not
  descend through a link and does not skip the link in silence.
- The walk skips the `lock` file. It skips only an atomic temporary file: a name
  that starts with `.` and contains `.tmp-`, as `atomicWrite` makes it. It does not
  skip an arbitrary malformed record name. It reads each record with
  `readRecordFile`. It stops at the first callback error or read error.
- `LockedStore` gives the same read, write, update, list, next, and generic
  methods. These methods do not take the lock again.
- The legacy `Read`, `Write`, `Update`, `List`, and `Next` keep their signatures.
  `List` keeps the `records.ValidID` filter. A legacy read or write of a long ID
  still works.
- One function reads a JSON record file. That function is `readRecordFile`. One
  function writes a JSON record file. That function is `writeRecordFile`. The
  writer uses `atomicWrite`. The temporary name that `atomicWrite` makes starts
  with `.` and contains `.tmp-`.
- A collection component or ID that is empty, `.`, `..`, has a slash, has a
  backslash, has a control character, has an uppercase letter, or is too long
  fails with `invalid store <name>: <value>`.
- `writeRecordFile` rejects an encoded record above 1048576 bytes before
  `atomicWrite` creates or replaces a file. The encoded record is the JSON object
  and one final newline. The error is `record too large: <path>`. An encoded
  record of exactly 1048576 bytes is valid. A rejected write keeps the old record
  bytes and makes no new record. The legacy `Write` and `Update` paths, the
  generic `WriteIn` and `UpdateIn` paths, and the `LockedStore` methods use this
  check. A multi-write caller reports its earlier effects and makes no rollback
  claim. The reader and lock contracts do not change.
- A directory, a symbolic link, an empty file, or a JSON array in place of a
  record fails with `record invalid: <path>`. A file above 1 MiB fails with
  `record too large: <path>`. A file of exactly 1 MiB is valid.

### AC3 Grouped use

- A grouped read of related records runs under one `WithLock` call. It sees the
  records of one completed writer or the next completed writer. It sees no
  mixture while successful write groups run.
- A write that succeeds before a callback error stays. After the failed group
  releases the lock, a related reader can see its completed prefix. The store
  makes no transaction and no rollback claim.
- A nested inbox path `["inbox", "coordinator"]` with ID `1` writes
  `.hive/state/inbox/coordinator/1.json`.
- Two OS processes that use `WithLock` and `Next` lose no update.
- The walk finds a record in a nested collection and returns the path of a bad
  record.

### AC4 Legacy behavior

- `Read`, `Write`, `Update`, `List`, and `Next` keep the behavior of T01.
- `List("projects")` still hides a file whose ID has more than 18 characters.
- `Read("tasks", "12-spawn-worker", &record)` reads a long task ID.
- `Next("inbox:coordinator")` keeps the colon in the counter name.
- The old unsafe-name tests still pass. The new checks can refuse more names.

### AC5 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The lock has the default wait, the cancellation, the release, and the handle rules. | V1 |
| AC2 | Generic record operations use safe names, bounded reads, atomic writes, and nested paths. | V2, V4 |
| AC3 | Grouped reads are consistent, partial writes stay, and the walk finds each record while it rejects a linked collection or record. | V3 |
| AC4 | Legacy behavior stays compatible. | V3 |
| AC5 | The gate passes. | V5 |

## Verification items

### V1 Lock tests

- Proves: AC1.
- Command: `go test -race ./internal/lib/store/ -run 'TestWithLock' -v`, red and green.
- Expected: a held lock makes a default call return `ErrLockTimeout` after about
  two seconds. A 100 millisecond context returns `context.DeadlineExceeded`. A
  canceled context returns `context.Canceled`. A three second context waits
  longer than two seconds and then gets the lock.
- Expected: a failed acquisition changes no record. A callback error releases the
  lock and a later call gets it. A callback panic releases the lock. A held handle
  gives `ErrLockReleased` after the callback. A nested `WithLock` from the
  callback fails and the outer lock still works.
- Expected: a held lock does not block `Store.Read` or `Store.ReadIn`. A legacy
  `Write`, `Update`, or `Next` against a held lock returns `ErrLockTimeout` after
  the bounded wait.

### V2 Name and path tests

- Proves: AC2.
- Command: `go test -race ./internal/lib/store/ -run 'TestStoreNames|TestReadIn|TestWriteIn|TestListIn|TestWriteLimit' -v`, red and green.
- Expected: the ID `12-spawn-worker` works. A 40-character slug works in a task
  ID. The nested inbox path writes `inbox/coordinator/1.json`. A missing
  collection lists as an empty slice. The `include` function selects one file and
  reads no other file.
- Expected: these names fail on `ReadIn` and `WriteIn`: empty, `.`, `..`, `../x`,
  `a/b`, `a\b`, a control character, an uppercase letter, and a 201-byte
  component. `ListIn` skips a file with such a name.
- Expected: a directory, a symbolic link, an empty file, and a JSON array give
  `record invalid`. A file of 1 MiB plus one byte gives `record too large`. A file
  of exactly 1 MiB works. A bad selected file makes `ListIn` return no records.
- Expected: `TestWriteLimit` uses the fixed literal `1048576` and not the
  implementation constant. It writes an encoded record of exactly 1048576 bytes,
  including the final newline, and the write succeeds. It writes one byte more
  and gets `record too large: <path>`.
- Expected: a rejected write through each of `Write`, `Update`, `WriteIn`,
  `UpdateIn`, and a `LockedStore` write or update keeps the old record bytes
  equal. A rejected create makes no record. A multi-write caller reports its
  earlier effect.
- Expected: a symbolic link at the state root, a collection ancestor, a record
  leaf, or the lock gives `store path invalid` or `record invalid`. A dangling
  link, a directory, a FIFO, and a socket fail with no write.

### V3 Grouped, legacy, and walk tests

- Proves: AC3, AC4.
- Command: `go test -race ./internal/lib/store/ -run 'TestGrouped|TestLegacy|TestWalk' -v`, red and green.
- Expected: one writer writes two records with one version under the lock. A
  reader under the lock sees equal versions in 200 rounds. A reader without the
  lock can see a mixture. This shows the lock is necessary.
- Expected: a callback writes record A, then returns an error. A later read shows
  A. A callback writes A and B, then errors. A later read shows A and B.
- Expected: four OS processes use `WithLock` and `LockedStore.Next` 50 times. The
  final counter is 200. A grouped test never calls `Store.Next` inside its own
  `WithLock` callback.
- Expected: `List("projects")` hides a 19-character ID and `ListIn` returns it.
  `Read` reads a long task ID. `Next` accepts a colon. The old unsafe-name tests
  pass.
- Expected: `WalkRecords` finds `projects`, `tasks`, `inbox/coordinator`, and
  `counters/inbox:coordinator.json` records. It skips `lock` and only a
  dot-prefixed atomic temporary file. A linked collection or a linked record
  returns its path and the walk does not descend it. A bad nested record returns
  its path.

### V4 Reader and writer check

- Proves: AC2.
- Command:

```sh
grep -nE 'os\.(Open|OpenFile|ReadFile|WriteFile|Create|CreateTemp)\(' internal/lib/store/*.go
grep -n 'syscall.Flock' internal/lib/store/*.go
grep -n 'io.LimitReader' internal/lib/store/*.go
```

- Expected: the only JSON record reader is `readRecordFile`. The only JSON record
  writer is `writeRecordFile`. `atomicWrite` is the only temporary file maker.
  The lock open is in `lock.go`. `syscall.Flock` is in `lock.go`. The size limit
  is in `readRecordFile`.

### V5 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Trust boundary

- The approved boundary is a local user and a cooperating writer. The store
  protects against parallel hive processes, a static symbolic link, and a short
  write. It does not protect against a hostile process that replaces a file
  between `Lstat` and `Open`. That replacement race stays deferred.
- A caller passes a safe collection and ID. The store checks them again.
- The store does not log credentials. The store reads no environment variable.

## Return contract

- The coordinator gives the pass ID, the report path, and the report marker in
  the dispatch. The worker writes one report.
- The worker commits on `task/T22-store-operations` with a subject of the form
  `type(scope): summary`. The worker pushes the branch. The worker adds no
  co-author and no trailer.
- The worker writes each `V<n>.md` in `spec/hive-v1/evidence/T22-store-operations/`.
  The worker runs the gate from the worktree root.
- The worker does not edit the tracker, `AGENTS.md`, or a task file. The worker
  does not remove a worktree. The worker does not start another agent.
- On a blocked step, the worker writes the report with a plain `BLOCKED` line and
  continues with each step that does not need the blocked step.

## Out of scope

- A journal, a transaction, or an automatic rollback.
- A Markdown reader or writer. T07 owns the memory and handoff files.
- A change to the `records.ValidID` rule.
- A change to the record JSON shape.
- A new dependency.
- A hostile file replacement between `Lstat` and `Open`.

## Changes

None.
