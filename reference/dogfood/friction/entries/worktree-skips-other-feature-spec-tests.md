---
title: "A worktree and CI skip the tests that read the spec of another feature"
status: "Open"
opened: 2026-10-24
decision: ""
---

**Summary:**
Since task 140 (g14), `hive worktree create` copies only the feature folder
of the task into the worktree. Many `internal/tui` tests read fixtures and
grids from `.hive/spec/tui-v2/design/`. In a `golden-frames` worktree these
tests skip with `local spec file missing: ../../.hive/spec/tui-v2/design/...:
Git does not track .hive/spec/`. On 2026-10-24, `go test ./internal/tui/ -v`
in worktree 146-gf-logs gave 78 SKIP lines. CI has no `.hive/spec/` at all, so
CI skips the same tests. A task can break such a test, and the gate, the
verifier, and CI all stay green. Worker t143-worker-1 found it and copied
`tui-v2/` from the home by hand to run the tests.

**Partial fix, 2026-10-24:** Task `151-design-copy-in` landed on `dev` at
`b5e0289d`. The new command also copies each other feature's `design/`
folder. Independent verification reduced the top-level skip count from 78 to
2. The count includes 14 skips when subtests count too. Each remaining skip
reads a file outside `design/`:

- `TestEachBoardFixtureLoadsAtItsSpecSize` reads `.hive/spec/tui-v2/README.md`.
- The ask review tests read JSON files under `.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/`.
  They include `TestAskStoppedEditorKeepsTheStateInTheCard` and the subtests
  listed in task 151's `evidence/VERIFIED.md`.

This entry stays open. Task 151 changes no CI behavior. CI still has no spec
tree. A fresh clone also has no spec tree. The installed CLI remains at
`3187a9fb` until the next install.

**Reproduction steps:**
1. Build `dev` at `b5e0289d` or later.
2. Use that binary to create a `golden-frames` worktree in a scratch home with the spec tree.
3. In the worktree, run `env -u TERM TZ=UTC go test ./internal/tui/ -count=1 -v | grep -c '^--- SKIP'`.
4. The count is 2. The skip messages name files outside another feature's `design/` folder.

**Workaround:** copy `.hive/spec/tui-v2/` from the home into each existing
worktree and fresh clone before tests. Report the skip count. Keep this step
until the installed binary has task 151. After that install, new worktrees
get the design files. The tests that read other paths still need those files.

**Fixture version constraint, 2026-10-24:** Task 143's verifier copied 12
newly landed Health artifacts into branch `08326acb`. That branch predates
the Health grid registrations. `TestGfGoldenGridFiles` rejected the four
extra grids. The verifier retained those setup failures and each copied
path/hash. It removed only its new unmatched copies. With matching fixtures,
every declared grid ran and the full gate passed with no missing-fixture skip.

Keep fixture versions matched to the checked code. Preserve the task's own
spec and design copies. A fresh clone must receive those matching copies.
Do not copy later boards into older code without the corresponding source
integration. Do not remove a test or add a skip to hide a mismatch. Check the
newly landed screen after the source merge.


**Clean-checkout duration regression, 2026-10-26:**
Task 183 local checks had all 3,208 design and task 135 inputs. Its delivered
CI failed because two new Health width regressions directly opened the
ignored `GfHealth.fixture.json`. The failure was not a product render failure.
Task `185-duration-ci-inputs` replaces only their input with fixed,
self-contained records and clock values. Every original literal assertion,
size, boundary, and column check remains intact. No skip was added.

Independent verification reproduced all 13 missing-file failures in the
pre-fix clean archive. The fixed clean archive ran all 13 cases successfully,
with zero skips. A deliberate width break failed the assertions and exact
restoration passed. Product source bytes remained unchanged.
The delivered fix is `1b53a389968f77d7d93e326597edb779586ca7b4`.
Both actual CI runs, `40276720423` and `40276713655`, completed successfully
at that exact head. Evidence is in task 185's `evidence/` folder.

This narrow fix leaves the broader issue open. The independent clean-suite
inventory still has 575 historical design-input skips and one memory-helper
skip. The complete local UI, component, and board-export run has 1,572 passes
and zero skips. Keep both input modes and their limits separate.
