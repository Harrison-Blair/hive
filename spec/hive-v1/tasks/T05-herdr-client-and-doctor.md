# T05 herdr client and doctor

| Item | Value |
| --- | --- |
| Task ID | T05-herdr-client-and-doctor |
| Wave | 2 |
| Tier | strong |
| Depends on | T01, T22, CP1 |
| Blocks | T09, T10 |
| Branch | `task/T05-herdr-client-and-doctor` |
| Worktree | `projects/hive/worktrees/T05-herdr-client-and-doctor/` |
| Evidence | `evidence/T05-herdr-client-and-doctor/` |

## Overview

This task gives hive its typed herdr client, the bounded process runner, and
`hive doctor`. The client sends one request for each connection, pins protocol
22, checks the protocol before any other method, and bounds each request and
each reply. The runner bounds each external process. `doctor` is the first
command that uses both. Each later herdr method lives in its own file of the
client package, which the task that needs it adds. `reference/herdr/` is the
API authority.

## User flows

1. The user runs `hive doctor` in a pane of the hive session. Each check prints one line with `ok`, `warn`, or `fail`. The exit status is 0 when each required check passes.
2. The user runs `HERDR_SOCKET_PATH=/tmp/none.sock hive doctor`. The herdr check fails with `herdr socket: /tmp/none.sock: no such file`, and the exit status is 1.
3. A later command calls `client.Call(ctx, "agent.wait", params, &result, herdr.WithDeadline(125*time.Second))`. The call waits for the server answer up to that deadline, and the log holds the request and the reply.
4. The user runs `hive version --json` and gets the envelope with the version data in `result`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `cmd/root.go` | Register the `doctor` command. |
| `cmd/doctor/doctor.go` | New. `hive doctor`. |
| `cmd/version/version.go` | Add `--json` through the envelope. |
| `internal/doctor/doc.go` | New. Package summary and file map. |
| `internal/doctor/checks.go` | New. One check function for each line. One line for each check in the list. |
| `internal/doctor/run.go` | New. Run the checks in order and collect the results. |
| `internal/doctor/render.go` | New. Text rendering. |
| `internal/lib/herdr/doc.go` | New. Package summary and file map. |
| `internal/lib/herdr/client.go` | New. Socket path, one request for each connection, the deadline and its option, the request cap, the reply cap, the reply checks, request IDs, logging. |
| `internal/lib/herdr/protocol.go` | New. `const Protocol = 22`, the ping handshake, and the protocol gate before the first other method. |
| `internal/lib/herdr/errors.go` | New. The error envelope as a Go error with `Code` and `Message`, and the code constants that hive uses. |
| `internal/lib/herdr/types.go` | New. `AgentInfo`, `PaneInfo`, `TabInfo`, `WorkspaceInfo`, and `WorktreeInfo` with the fields that hive reads. |
| `internal/lib/run/run.go` | New. The bounded process runner. |

## Store use

Use `store.WalkRecords(ctx, fn)` for the state-store check. The function receives
each record. It returns the first read error. Do not walk `.hive/state` with
`os.ReadDir`. Do not open a record file directly.

Method matrix: `Store.Read` and `Store.ReadIn` take no lock. A standalone list
takes no lock. A related read uses one `WithLock` callback. Each standalone
mutation uses bounded acquisition. A `LockedStore` method never acquires another
lock. `WalkRecords` uses one `WithLock` callback.

## Requirements

### AC1 Client

- The socket path is `HERDR_SOCKET_PATH`. Without the variable, each call fails with `herdr socket: HERDR_SOCKET_PATH is not set`. There is no other source.
- `Call(ctx, method, params, result, opts...)` opens one connection, writes one request line with a string `id`, the `method`, and `params`, reads one response line, and closes. `params` is never null. An empty object goes as `{}`.
- Each call has a deadline. The default is 10 seconds. `herdr.WithDeadline(d)` sets the deadline of one call. A wait method passes its server `timeout_ms` plus 5 seconds. An earlier deadline of the context wins.
- A request ID is `<prefix>-<n>`. The prefix is 8 random hex characters that the process draws once. `n` comes from one counter that each caller in the process shares. Two in-process callers share the prefix and get unique complete IDs. Two child processes get distinct prefixes. Thus IDs from parallel hive processes do not collide in the shared log.
- The encoded request line, with its newline, has at most 1,048,576 bytes. A longer line fails with `request too large` before any write.
- A reply above 1 MiB fails with `reply too large`. A reply whose `id` differs from the request `id`, or that has both `result` and `error`, or neither, fails with `bad reply: <reason>`. `result` stays unchanged on a failure.
- A response with `error` returns a `*herdr.Error` with `Code` and `Message`. `errors.As` finds it.
- Each call logs a `herdr request` event and a `herdr reply` event with `request_id`, `method`, `duration_ms`, and the bodies through `log.Body`.
- `Ping(ctx)` sends `ping` and compares the protocol number of the result with `Protocol`, as `reference/herdr/protocol.md` gives. A different number fails with `herdr protocol <n>, hive needs 22`.
- Before the first method other than `ping` on a socket path, the client runs `Ping` once. The client keeps a pass for the life of the process. A failure fails each later call on that path with the same error, and no other request reaches the server.

### AC2 Process runner

- `run.Command(ctx, dir, name, args...)` runs one process and returns `Stdout`, `Stderr`, `ExitCode`, and `Truncated`. The process gets `/dev/null` as stdin.
- The process runs in its own process group. The default deadline is 30 seconds. On the deadline the whole process group gets `SIGKILL`, and the call fails with `run.ErrTimeout` within two seconds. A child process that holds the output pipes does not delay the return.
- Each stream keeps at most 1 MiB. Above it, the result has `Truncated` true and the rest is gone.
- Each run logs `process start` and `process end` with `command`, `exit_code`, and `duration_ms`.
- A failure of the process is a result with a non-zero `ExitCode`, not a Go error. A failure to start is a Go error.

### AC3 Doctor

| Order | Check | Required | Pass rule |
| --- | --- | --- | --- |
| 1 | herdr socket | yes | `HERDR_SOCKET_PATH` has a value, the socket exists, `ping` answers, and the protocol is 22. The detail gives the path and the protocol. |
| 2 | herdr pane | no | `HERDR_ENV` is `1` and `HERDR_PANE_ID` has a value. The detail gives the pane, or `outside a herdr pane`. |
| 3 | home | yes | The marker exists, and `.hive/` is writable. |
| 4 | state store | yes | The lock is free within 2 seconds, and each record file parses. |
| 5 | git | yes | `git --version` exits with 0. |
| 6 | supervisor | no | A `supervisor` claim exists. T15 refines the rule. |
| 7 | harness `claude`, `codex`, `opencode`, `pi`, `cursor-agent` | no | The binary is on `PATH`. One line for each. |
| 8 | gh | no | `gh --version` exits with 0, and `gh auth status` exits with 0. |

- Each line prints `ok`, `warn`, or `fail`, the check name, and a detail. A required check that fails prints `fail`. A check that is not required and fails prints `warn`.
- The exit status is 1 only when a required check fails. With `--json`, `result` holds the list of checks with `name`, `status`, `required`, and `detail`.
- Each check uses the runner or the herdr package. No check runs longer than its bound.
- The state-store check uses one `store.WalkRecords` call. A held lock gives `store lock timeout`. A bad record gives its path. A linked collection or record gives its path and the walk does not descend it. A static symbol link at the state root, an ancestor, the lock, or a leaf fails before the read. The check makes no write.

### AC4 Version envelope

- `hive version --json` prints the envelope with `outcome` `ok` and `result` with `version`, `commit`, `modified`, and `go`.

### AC5 Gate

- The four gate commands pass from the root of the worktree.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The client sends one request for each connection, checks the protocol first, maps errors, bounds requests and replies, checks each reply, and logs. | V1, V2 |
| AC2 | The runner bounds time and output for the whole process group and reports exit codes. | V3 |
| AC3 | `doctor` prints each check with the correct status, the state-store walk rejects a linked path, and the exit status follows the required checks. | V4, V5 |
| AC4 | `hive version --json` prints the envelope. | V5 |
| AC5 | The gate passes. | V6 |

## Verification items

### V1 Client tests

- Proves: AC1.
- Command: `go test -race ./internal/lib/herdr/ -v`, red and green.
- Expected: the tests start a fake herdr server on a Unix socket in a temporary directory. The server records each method that it receives. Cases: a success envelope, an error envelope with the code, and a reply above 1 MiB. Three bad replies fail: another `id`, both `result` and `error`, and neither. A server that never answers fails at the deadline. A server that answers after 1.5 seconds fails with a 1-second deadline and passes with `WithDeadline(3*time.Second)`. A pong with protocol 21 fails, and the server then receives no `workspace.list`. A request of 1,048,576 bytes with non-ASCII text passes the local check, and one byte more fails with `request too large`. Two in-process callers share one prefix and get unique complete IDs. Two child processes get distinct prefixes. An unset `HERDR_SOCKET_PATH` fails with the message of AC1.

### V2 Live ping

- Proves: AC1.
- Setup: `go build .` and make a unique temporary home. Create its marker `.hive/README.md` explicitly. Export `HIVE_TMP=<absolute fixture path>`. Run from a pane of the hive session. Do not remove the worktree `.hive` directories.
- Command:

```sh
./hive --home "$HIVE_TMP" --json doctor | head -c 2000
grep '"method":"ping"' "$HIVE_TMP"/.hive/logs/"$(date +%F)".jsonl | tail -n 2
```

- Expected: the herdr check has the status `ok` and a detail with `protocol 22` and the socket path. The two log lines are the `herdr request` and `herdr reply` events for `ping`, with the same `request_id`. The live herdr pane requirement stays. The fixture home holds the state and the log.
- Cleanup: remove only `$HIVE_TMP`. Do not remove the worktree `.hive/state`, `.hive/logs`, `.hive/tmp`, a report, or a log.

### V3 Runner tests

- Proves: AC2.
- Command: `go test -race ./internal/lib/run/ -v`, red and green.
- Expected: `sleep 5` with a one-second deadline fails with `run.ErrTimeout` within two seconds. A shell that starts `sleep 30` in the background with the inherited pipes and then runs `sleep 5` fails within two seconds too. Both processes are gone after the call. A process that prints 2 MiB gives `Truncated` true and 1 MiB. `false` gives exit code 1 and no Go error. A missing binary gives a Go error. A process that reads stdin gets end of file at once.

### V4 Doctor tests

- Proves: AC3.
- Command: `go test ./internal/doctor/ -v`, red and green.
- Expected: with a fake socket and a unique temporary home, each check line has the expected status. With `HERDR_SOCKET_PATH` unset, the herdr check is `fail` and the exit status is 1. With a path to no file, the herdr check is `fail` with the path in the detail. Without `HERDR_ENV`, the pane check is `warn` and the exit status is 0. With no `gh` on `PATH`, the gh check is `warn` and the exit status is 0. A held store lock gives a `fail` state-store line with `store lock timeout`. A linked collection or record gives its path.

### V5 Doctor and version transcript

- Proves: AC3, AC4.
- Setup: `go build .` and make a unique temporary home. Create its marker `.hive/README.md` explicitly. Export `HIVE_TMP=<absolute fixture path>`. Do not remove the worktree `.hive` directories.
- Command:

```sh
./hive --home "$HIVE_TMP" doctor; echo "exit $?"
HERDR_SOCKET_PATH=/tmp/none.sock ./hive --home "$HIVE_TMP" doctor; echo "exit $?"
env -u HERDR_SOCKET_PATH ./hive --home "$HIVE_TMP" doctor; echo "exit $?"
./hive --home "$HIVE_TMP" --json version; echo "exit $?"
```

- Expected: the first run prints eight groups of lines and exit 0, or exit 1 with a `fail` line that explains why. The second run prints `fail herdr socket` with `/tmp/none.sock` and exit 1. The third prints `fail herdr socket` with `HERDR_SOCKET_PATH is not set` and exit 1. The fourth prints the envelope and exit 0.
- Cleanup: remove only `$HIVE_TMP`. Do not remove the worktree `.hive/state`, `.hive/logs`, `.hive/tmp`, a report, or a log.

### V6 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The methods `worktree.*`, `agent.*`, `workspace.*`, `pane.*`, and `events.subscribe`. T09, T10, T13, and T15 add one file each.
- A live check of the supervisor process. T15 adds it.
- Stdin for a child process. A command that needs to give input to a process writes a file under `.hive/tmp/` and passes its path.

## Changes

None.
