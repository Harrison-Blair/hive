# Verification: 39-screen-health

Verifier: `t39-verifier-3`. Date: 2026-10-22.
Task commit: `6746efb0`. Verdict: **fail** (current dev merge).

## Items

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | pass | Base red run, task green run, token mutation, key mutation, and kept-behavior mutations ran again. |
| V2 Live check | pass | Each of the three boards has 0 differing cells under the space-cell rule. Keys and restart routes passed in a scratch home. |
| V-gate Diff scope and gate | fail | All five gate commands pass on the task commit. The required merge of current dev conflicts in shell_test.go and ste-baseline.json. The integrated result cannot be tested. |

New evidence is in `verifier3/`. Earlier evidence is retained.

## V1 Tests

Commands and outputs:

- `go test -race -count=1 -v ./internal/tui -run 'Health|Restart'`:
  `green.txt`, pass, including the three golden tests.
- Copied `health_test.go` and `restart_test.go` to a detached scratch worktree
  at base `b145a41f`. Renamed the copied string-reader test to avoid its old
  copy. The first run failed to compile because the new Health view constants
  and `acceptAgents` do not exist (`base-compile.txt`).
- Added a test-only compatibility shim: `healthHome = 0`, `healthAgents = 1`,
  and a no-op `(*healthState).acceptAgents(agentsData)`. This shim adds no
  behavior. Ran `go test ./internal/tui -run
  'Health|AgentsRestart|HeaderMarksOldSupervisor' -count=1 -v`.
  `base-red.txt` shows all nine new Health tests fail. Seven moved restart
  tests also fail. The board checks fail on cell content or palette tokens.
- In a second detached scratch worktree at the task commit, removed
  `dimStyle.Render` from the agent role column. Ran `go test ./internal/tui
  -run 'HealthAgentsMatchesItsBoard|CompactHealthMatchesItsBoard' -count=1`.
  Both golden checks fail (`token-red.txt`).
- Restored the token. Changed the Health `h` key to `x`. Ran `go test
  ./internal/tui -run 'HealthKeysAndFooterHints|HealthAgentsHTogglesStoppedAgents'
  -count=1`. Both tests fail (`key-red.txt`).
- The unchanged restart-result and old-supervisor-header guards pass on the
  base. Changed the result prefix from `restarted` to `broken`, and the
  header state word from `old` to `broken`. Both guards fail with `go test
  ./internal/tui -run
  'AgentsRestartResultIsRenderedAndLogged|HeaderMarksOldSupervisorBinaryAndShowsRestartHint'
  -count=1 -v` (`kept-red.txt`). These are kept-behavior guards.
- Restored all mutations. `go test ./internal/tui -run
  'Health|AgentsRestart|HeaderMarksOldSupervisor' -count=1` passes
  (`restored-green.txt`). No flake rerun was needed.

Only scratch worktrees received test copies and mutations. No task source
file changed.

## V2 Live check

Built the task binary with `go build`. Started owned scratch tab `wA4:t2`
and pane `wA4:p2`. The pane had 57 rows and more than 120 columns.
Each run used a scratch `HIVE_HOME`. Each demo command was:

`COLORTERM=truecolor hive --demo FILE --size WxH`

The script `verifier3/live.py` records the exact paths and commands. It checks
protocol 22, waits for two equal reads 500 ms apart, rejects a truncated read,
and checks scroll offset 0. The converter crops the top-left board size and
compares all cells. It excludes only the token and weight of matching space
cells, as the Manager decision requires. `compare.go.txt` records this code.

| Board | Size | Result |
| --- | --- | --- |
| Health | 120×40 | 0 cells differ |
| HealthAgents | 120×40 | 0 cells differ |
| CompactHealth | 80×24 | 0 cells differ |

Each board has a new ANSI capture, full read response, pane metadata, diff,
screen PNG, and board PNG in `verifier3/`. Captured all six PNGs again with
Firefox headless, a separate new profile for each image, and a 1220×780
window. Screen HTML uses an ANSI-cell-derived `pre`. Reviewed the six PNGs.
Page padding, background outside the grid, and browser font metrics differ;
the terminal cells match.

### Keys and normal mode

`verifier3/keys.py` records the live steps. Plain-text captures record results.

- Home: `down` moves to git, `G` moves to gh, and `enter` focuses details.
- Agents: `esc` cancels the prompt, `j` moves, `h` hides stopped agents,
  and demo `r` gives `demo mode: no action`.
- Compact Agents: `enter` opens details; `esc` returns to the list.
- Normal mode used `--no-startup`. Four scratch agent records loaded with
  `R`. Coordinator `r` asked for confirmation. `n` cancelled. A separate
  `y` accepted and reached the restart operation.
- Manager and worker `r` each gave exactly
  `restart: coordinator and supervisor only`.
- A stopped supervisor and a stale coordinator each reached the restart
  operation without a prompt. The coordinator live-pane check used only
  the owned scratch TUI pane. Other pane and workspace targets did not exist.
  Each accepted restart stopped at `workspace_not_found`, before an agent
  could start. This live check does not claim successful process launch.
- Each run quit with two `ctrl+c` keys. No real home record changed.

### Board elements and state without colour

Checked plain text with colour removed. Each state retains a glyph or word.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Wide header | coordinator working, supervisor live, needs-you count, clock | matches | Health and HealthAgents |
| Compact header | coord ●, sup ●, needs-you count, clock | matches | CompactHealth |
| Tabs | 5 Health and ━ underline | matches | Decision 3 |
| Views | [Home], [Agents], [ ] switch view | matches | Decision 18 |
| Home list title | checks · 17 · 1 warn | matches | Health |
| Check columns | CHECK, RESULT | matches | Health |
| Check rows and states | ✓, !, check names, result words | matches | Health |
| Home selection | ▶ ! unresolved agents | matches | Health |
| Check fields | result warn · 1 agent, source, checked | matches | Health |
| Unresolved agent table | AGENTS, ID, ROLE, STATE, PANE, unresolved, gone | matches | Health |
| Explanation | WHAT IT MEANS, NEXT STEP and body | matches | Health |
| Wide agent title | agents · 5 · h hides 1 stopped | matches | HealthAgents |
| Compact agent title | agents · 5 | matches | CompactHealth |
| Agent columns | AGENT, ROLE, STATE, compact PANE | matches | Both agent boards |
| Agent states | ● working, ● live, ◐ idle, ● unresolved, ○ stopped | matches | Words remain without colour |
| Agent selection | ▶ supervisor-g4 or ▶ coordinator-g2 | matches | Board fixture |
| Agent fields | role, state live, pid, pane, binary current, started | matches | HealthAgents |
| Actions | ACTIONS, r restart, a attach, p peek | matches | HealthAgents |
| Restart box | restart supervisor-g4?, y restart, esc cancel | matches | HealthAgents |
| Restart explanation | live confirmation and immediate stale/stopped restart | matches | HealthAgents |
| Compact notes | restart scope and h hides 1 stopped agent | matches | CompactHealth |
| Borders | rounded continuous borders | matches | Three grids and responsive tests |
| Compact details | coordinator-g2, esc back, [details] | matches | Decision 9 and live keys |
| Footers | [list], [details], [restart] and their key hints | matches | Decision 21 |

## V-gate and integration finding

`verifier3/gate.txt` contains the full output and exit status of each exact
AGENTS.md gate command. `gofmt -l` and `go mod tidy -diff` print nothing.
Vet, lint, and race tests exit 0. No rerun was needed.

The initial `git merge dev --no-edit` and `git merge origin/dev --no-edit`
reported `Already up to date` at dev `13210c3f`. `verifier3/scope.txt` records
`git diff dev...HEAD --stat`; source changes are within the brief table.

Dev advanced to `fbca94d5` during verification. Repeated the required merge
before the verdict. It conflicts in:

- `internal/tui/shell_test.go`
- `testdata/ste-baseline.json`

`merge-conflicts.txt` records the conflicts and `final-merge.txt` records the
last attempt. Aborted both attempts. The shell test conflict is outside the
manager's README/ste-baseline-only exception. No conflict resolution was
made. A repair worker must merge dev and keep both sides of shared files.
A fresh verifier must test that integrated result.

## Cleanup

Closed owned tab `wA4:t2`. The tab list contains only `wA4:t1`.
Removed both detached scratch worktrees, the scratch home, binary, browser
profiles, and temporary scripts. Checked that `.hive/tmp/verify39v3` does not
exist. Retained final evidence in this task's evidence folder.
