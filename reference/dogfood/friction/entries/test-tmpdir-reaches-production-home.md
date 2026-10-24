---
title: "A test TMPDIR below the home permits production state access"
status: "Open"
opened: 2026-10-24
decision: "Require reviewed physical state, socket, PID, and descriptor isolation before branch execution."
---

**Issue:** An environment-only test setup let home-discovery tests open the real state store and migrate it.

**Summary:**
Task 155 verifier 2 first set `TMPDIR` below its worktree in the real home.
The command suite included tests that clear `HIVE_HOME` and change to
`t.TempDir()`. Home discovery intentionally selects the outermost marker.
The real home was therefore an ancestor of the temporary test directories.
Scratch `HOME`, config, and `HIVE_HOME` values did not isolate these tests.

The archived first gate ran `env -u TERM TZ=UTC go test -race ./...`.
Its command file has mtime `2026-10-24T10:28:01.803407108Z`.
`TestRootMissingHomeUsesOneErrorLine` took 1.49 seconds, returned success,
and printed the real hive, qmeter, and stenographer project paths.
Preceding init tests also selected `/home/penguin/source/hive`.
The real project-list request `r-106f99ed` took 1,488 ms and enclosed the
schema 1-to-2 migration at 10:28:04.191 through 10:28:05.554 UTC.
The migration logged 554 agent changes.

This is strong recorded cause evidence. The logs lack a PID/request-to-test
link, so the exact calling PID remains unproved. No unsafe reproduction is
needed to establish that the test printed production data.

The verifier later changed `TMPDIR` to `/tmp` and repeated the checks.
That green run did not undo the earlier production access. The initial
failure was recorded as a setup correction; its state effects were missed
until the marker check during task delivery.

**Impact correction:** The first incident audit counted empty `kind` values
as missing. D1 permits an empty `kind` for spawned agents. The corrected
frozen snapshots have seven records missing all four new keys and 509 with
present-but-empty `kind`. The earlier 516 figure is superseded.
Seven is a snapshot count, not a fixed repair list. Continued legacy-CLI
coordination can create more old-shaped records. Any future recovery needs
a fresh exact key-presence audit and must preserve current and unknown fields.

**Reproduction steps:**
1. Read task 155 `evidence/verifier-2/setup-attempt-1/gate-worktree-race.log`, lines 2–18.
2. Read the original `common.py` tool write in its preserved session history: `TMPDIR=str(d/'tmp')` below the worktree.
3. Read `cmd/project_root_test.go`, `cmd/common/common.go`, `internal/lib/home/home.go`, and `internal/project/list.go`.
4. Compare the recorded test output and timing with real log request `r-106f99ed` and the migration events.
5. Do not repeat the command where production paths, sockets, processes, or descriptors are reachable.

A future executable regression must use a reviewed physical sandbox with
synthetic nested markers and non-mutating negative probes first.

**Evidence:** Task 155 preserves the first failed gate, corrected runs, setup
notes, and session tool history. The incident backup is
`~/.local/share/hive-backup/incident-schema2-20261024T105000Z/`.
The corrected audit is `migration-impact-corrected.json`; the initial audit
remains as superseded history. Project incident files also preserve the
command window and static cause analysis.

**Workaround:** Use the verified task 152 namespace/mount/descriptor pattern,
review its adaptation, and pass negative production-reachability probes
before branch execution. Environment variables alone are insufficient.
No production record repair, marker change, installation, or runtime restart
is authorized by this entry. No new feature scope is requested.
