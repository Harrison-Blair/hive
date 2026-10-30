---
title: "A test TMPDIR below the home lets tests reach the real home"
description: "Home discovery picks the outermost marker, so a t.TempDir() below the home resolves to the real home"
status: "Resolved 2026-10-30"
opened: 2026-10-24
decision: "PR #64 moved the home to $HIVE_HOME or ~/.hive, which removed marker discovery and the TMPDIR path, and TestMain in cmd/root_test.go now sets HOME to a temporary directory and unsets HIVE_HOME"
---

**Issue:** Home discovery picks the outermost `.hive/README.md` marker above the
working directory (`internal/lib/home/home.go:54-68`). Some `cmd` tests clear
`HIVE_HOME` and change to `t.TempDir()` (`cmd/project_root_test.go:36-38`,
`cmd/homeinit_root_test.go:13-24`). TestMain isolates only the config directory
(`cmd/root_test.go:17-23`). When `TMPDIR` is below the real home, for example in
`.hive/tmp/`, these tests resolve the real home, read its records, and write its
logs.

**Summary:** On 2026-10-24, a verifier of task 155 set `TMPDIR` below its
worktree in the real home and ran the gate tests.
`TestRootMissingHomeUsesOneErrorLine` printed the real hive, qmeter, and
stenographer project paths. A request on the real home also ran the old schema
migration. That migration is gone (61c2b5bf). Home discovery and the tests are
unchanged. A rerun with `TMPDIR=/tmp` was clean.

**Reproduction steps:**
1. Do not run this against a home with data you need. Use a copy of the checkout as the home.
2. In that copy, set `TMPDIR` to a directory below it and run `env -u TERM TZ=UTC go test -race ./cmd/`.
3. Expected from the code, not yet run: the tests resolve the copy as the home, and its `.hive/logs/` gets new events.

**Update 2026-10-30 (branch `home-dot-hive`, PR #64):** The home moves to
`$HIVE_HOME` or `~/.hive`, so marker discovery and the `TMPDIR` path are gone.
The same class of issue moved to `HOME`: a test that clears `HIVE_HOME` now
resolves `~/.hive` of the real user. In the first test run of that branch,
before the tests were changed, the old "no home" tests ran with the real
`HOME` and created `~/.hive` with `logs/` and `schema.json` (15 events from
`guide` and `project list`; the builder removed it). PR #64 sets `HOME` to a
temporary directory in `TestMain` of `cmd` and unsets `HIVE_HOME` there. A test
package outside `cmd` that runs a command with `HIVE_HOME` empty is still not
guarded.
