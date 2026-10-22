# Verification: 39-screen-health

Verifier: `t39-verifier-2`. Date: 2026-10-22.
Task commit: `1eadc24a`. Verdict: **fail** (integration with current dev).

## Items

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tests | pass | Nine Health tests fail on the base with the test-only compatibility shim. Seven moved restart tests fail on the base. Task tests pass. Separate token and key changes each fail the relevant tests. |
| V2 Live check | pass | Each live board capture has 0 differing cells under the space-cell rule. Normal-mode restart routing and key checks pass in a scratch home. |
| V-gate Diff scope and gate | fail | The full gate passed against dev `866bde2a`. A final required merge of current dev `e7cdbd09` has four conflicts. The current integrated result cannot be tested. |

## V1

Independent runs are in `verifier/`:

- `base-red.txt`: the new test file does not compile on base dev `866bde2a`.
- `base-red-shim.txt`: all nine task tests fail after a test-only shim adds
  `healthHome = 0`, `healthAgents = 1`, and a no-op `acceptAgents` method.
  The copied string-reader test has a different name to avoid its base copy.
  Command: `go test ./internal/tui -run Health -count=1`.
- `base-restart-red.txt`: seven moved restart tests fail. Command:
  `go test ./internal/tui -run 'AgentsRestart|HealthAgentsHTogglesStoppedAgents|HeaderMarksOldSupervisor' -count=1`.
- `green.txt`, `green-final.txt`: `go test ./internal/tui -run 'Health|AgentsRestart' -count=1` passes on the task branch.
- `token-red.txt`: remove `dimStyle.Render` from the role column. Both agent
  golden checks fail. Command: `go test ./internal/tui -run 'HealthAgentsMatchesItsBoard|CompactHealthMatchesItsBoard' -count=1`.
- `key-red.txt`: with the token restored, change `h` to `x`. Both key tests
  fail. Command: `go test ./internal/tui -run 'HealthKeysAndFooterHints|HealthAgentsHTogglesStoppedAgents' -count=1`.

The base and mutation runs used two scratch detached worktrees. Neither run
changed the task source. Both scratch worktrees were removed.

## V2

Built the task binary with `go build`. Created scratch tab `w9G:t2`, pane
`w9G:p2`, in the verifier workspace. The pane was 57 rows and wider than
120 columns. Each run set `HIVE_HOME` to a scratch home and
`COLORTERM=truecolor`. Demo commands used each board fixture with
`--size 120x40` or `--size 80x24`.

`verifier/live.py` records the capture procedure. Each board has two equal
reads 500 ms apart. The saved socket response has `truncated: false`. Each
saved pane metadata response has scroll offset 0. Cropped the top-left board
size with `ansi.Truncate`, then called `cellgrid.Convert` and
`cellgrid.Compare`. Only the token and bold flag of matching space cells were
excluded. The converter script is saved as `verifier/compare.go.txt`.

| Board | Size | Result |
| --- | --- | --- |
| Health | 120×40 | 0 cells differ |
| HealthAgents | 120×40 | 0 cells differ |
| CompactHealth | 80×24 | 0 cells differ |

Captures, full socket responses, metadata, diffs, board PNGs, and screen PNGs
are in `verifier/`. Each PNG was captured again with Firefox headless, a
separate new profile, and a 1220×780 window. The screen PNG uses an HTML
`pre` made from the captured cells. Browser font metrics and page padding
are presentation differences; the cell comparisons above determine accuracy.

### Keys and normal mode

Saved plain-text observations are in `verifier/`.

- Home: `down` moves the selection, `G` selects the last check, and `enter`
  gives focus to details. The details footer changes.
- HealthAgents: `esc` cancels the restart prompt. `j` moves to the manager.
  `h` hides the stopped row. Demo `r` shows `demo mode: no action`.
- CompactHealth: `enter` opens the details overlay; `esc` returns to the list.
- Normal mode: ran the binary with `--no-startup` in the scratch home.
  `R` loaded four scratch agent records. No coordinator or supervisor
  process was live. The coordinator record referred only to the owned scratch
  pane, which held the TUI. Its `r` key opened the prompt. `n` cancelled it.
  In another run, `y` accepted it and reached the restart operation.
- A stopped supervisor record reached the restart operation without a prompt.
  A stale coordinator record with a nonexistent pane also reached it without
  a prompt. Each record used workspace `verify39-nonexistent`. Each attempt
  stopped at `workspace_not_found`, before any agent could start. Successful
  process launch is covered by the existing automated gate, not this live run.
- Manager and worker `r` each showed exactly
  `restart: coordinator and supervisor only`.
- Each run quit with two `ctrl+c` keys. The scratch log is retained as
  `verifier/scratch-log.jsonl`.

### Visual review and state without colour

Reviewed all six new PNGs and the plain text of each screen.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Wide header | `coordinator ● working`, `supervisor ● live`, count and clock | matches | Health and HealthAgents boards |
| Compact header | `coord ●`, `sup ●`, count and clock | matches | CompactHealth board |
| Screen tabs | `5 Health` with `━` underline | matches | Decision 3 |
| View bar | `[Home]` or `[Agents]`, `[ ] switch view` | matches | Decision 18 |
| Home list title | `checks · 17 · 1 warn` | matches | Health board |
| Check columns | `CHECK`, `RESULT`, check names and results | matches | Health board |
| Check states | `✓`, `!`, and warning text | matches | States remain visible without colour |
| Home selection | `▶ ! unresolved agents` | matches | Health board |
| Check detail fields | `result warn · 1 agent`, `source`, `checked` | matches | Health board |
| Unresolved-agent details | `AGENTS`, `ID`, `ROLE`, `STATE`, `PANE`, `unresolved`, `gone` | matches | Health board |
| Check explanation | `WHAT IT MEANS`, `NEXT STEP` | matches | Health board |
| Wide agent-list title | `agents · 5 · h hides 1 stopped` | matches | HealthAgents board |
| Compact agent-list title | `agents · 5` | matches | CompactHealth board |
| Agent columns | `AGENT`, `ROLE`, `STATE`, compact `PANE` | matches | Both agent boards |
| Agent states | `● working`, `● live`, `◐ idle`, `● unresolved`, `○ stopped` | matches | Each state has a word |
| Agent selection | `▶ supervisor-g4` or `▶ coordinator-g2` | matches | Each board fixture |
| Agent detail fields | `role`, `state live`, `pid`, `pane`, `binary current`, `started` | matches | HealthAgents board |
| Actions | `ACTIONS`, `r restart`, `a attach`, `p peek` | matches | HealthAgents board |
| Restart prompt | `restart supervisor-g4?`, `y restart`, `esc cancel` | matches | HealthAgents board |
| Restart explanation | live confirmation and stale/stopped immediate restart text | matches | HealthAgents board |
| Compact notes | restart scope and `h hides 1 stopped agent` | matches | CompactHealth board |
| List borders | continuous rounded border; selected item remains visible | matches | Golden and responsive tests |
| Compact details overlay | agent name, `esc back`, `[details]` | matches | Decision 9, live key run |
| Footer | `[list]`, `[details]`, `[restart]` with focus key hints | matches | Decision 21 |

## V-gate

`verifier/gate.txt` contains the full output of all five commands. `gofmt -l`
and `go mod tidy -diff` print nothing. Vet, lint, and race tests exit 0.
No flake rerun was necessary. `verifier/scope.txt` shows the diff against dev
`866bde2a`; the changed files are within the brief source table and evidence.

At the start, `git merge dev --no-edit` reported `Already up to date`.
During verification, dev advanced to `e7cdbd09`. A final merge failed in:

- `.hive/spec/tui-v2/README.md`
- `internal/tui/frame.go`
- `internal/tui/selection_marker_test.go`
- `testdata/ste-baseline.json`

Ran `git merge --abort`. No conflict resolution or source edit was made.
Reported the block to `manager-hive` in inbox record 257. A repair worker
must merge current dev. A fresh verifier must then test the integrated code.

## Cleanup

Closed the owned scratch tab. `herdr tab list --workspace w9G` shows only
`w9G:t1`. Removed both scratch worktrees, the scratch home, binary, browser
profiles, and remaining scratch files. Verified that `.hive/tmp/verify39`
does not exist. Final evidence is retained here.
