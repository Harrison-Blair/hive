# Verification: 29-frame-layout

Verdict: **fail**. Verifier: `t29-verifier-1`. Date: 2026-10-21.

Task code: `d6f3b98b`. This run merged `dev` at `02ffae7f` into the task branch. The merge commit is `82e98ff6`. No source file was changed by the verifier.

## Items

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | `verify/green.txt`, `base-compile.txt`, `base-red.txt`, `key-mutation.txt`, `token-mutation.txt`, `restored.txt`. The tests pass on the task code. The base run and both deliberate breaks fail. See the limits below. |
| V2 Live check | fail | `verify/*.ansi`, `*-live.png`, `*-board.png`, `live.txt`, and `*-diff.txt`. A screen key leaves the compact details overlay. Strict grid comparison also finds attributes that differ on blank cells. |
| V-gate Diff scope and gate | pass | `verify/scope.txt` and `verify/gate.txt`. Only the allowed source paths and evidence differ. All five gate commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing. The race test passes on the first run. |

## V1: independent rerun

The green run used `go test ./internal/tui ./internal/tui/component -run 'Test(Frame|FooterShowsStartupNoticeBeforeError|DemoUIStateSetsTheFrameState|DemoRestartPromptYesDoesNotRestart|Help|EachScreenKeepsTheFrameAtEachBreakpoint|TabMovesTheAccentBorder|DetailsOverlayBelow120Columns|CtrlCTwiceQuitsInEachFocusState|NormalModeUsesTheFrame|HeaderNeedsYouCountIsAccent|DemoActionKeysShowTheNotice)' -count=1 -v`.

A detached scratch worktree used `dev` at `02ffae7f`. The new frame test file and the component header test file were copied to it. The full frame test file fails to compile because the base has no frame API (`undefined: tabAgents`, `tabChat`, and others). This confirms an API dependency, but does not prove each key assertion independently.

For a behavioral red run, these API-dependent functions were removed from the scratch copy only: `TestDemoUIStateSetsTheFrameState`, `frameScreens`, `requireWholeBorders`, `TestEachScreenKeepsTheFrameAtEachBreakpoint`, `focusTokens`, `TestTabMovesTheAccentBorder`, `TestHelpKeys`, `TestDetailsOverlayBelow120Columns`, and `TestCtrlCTwiceQuitsInEachFocusState`. The unused `fmt` import was removed. The remaining new tests fail on the base, including the first test, `TestFrameRowsMatchMain` (461 different cells). The component test fails on the needs-you count token.

In a second detached scratch worktree at the task merge commit:

1. The global help key route was changed from `?` to `BROKEN?`. `go test ./internal/tui -run '^TestHelpKeys$' -count=1 -v` fails.
2. The wide details panel focus argument was changed to `false`. `go test ./internal/tui -run '^TestTabMovesTheAccentBorder$' -count=1 -v` fails.
3. The source was restored. Both tests pass with `-count=1`.

The full gate checks the old tests as well. Both scratch worktrees were removed. No deliberate break remains.

## V2: live run

The binary was built with `go build -o .hive/tmp/verify29/hive .`. A new scratch tab `w8R:t2`, pane `w8R:p2`, used a scratch `HIVE_HOME` and `COLORTERM=truecolor`. Herdr is 0.9.5. The pane has 41 rows and fits the requested boards. Demo mode implies no startup. No live agent was started.

Each run used `hive --demo ABSOLUTE_FIXTURE_PATH --size WxH`. The fixtures were `Main` and `Help` at 120x40, `CompactAgents` and `Compact` at 80x24, and `TooSmall` at 60x16. Each capture used `herdr pane read w8R:p2 --source visible --format ansi`. Two reads 500 ms apart were equal. The pane metadata had scroll offset 0. No capture was marked as truncated. The CLI returns ANSI text directly.

The ANSI captures were rendered as an HTML `pre`. Firefox captured that HTML and each of the five boards with `--headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://INPUT`. Each capture used a new scratch profile. The PNGs use different browser padding and font metrics; the cell comparison used the ANSI text, not PNG pixels.

Observed keys:

- Main: `tab` keeps graph focus. `?` opens Help. `down`, `esc` closes Help. `3`, `tab` moves the Work focus to details. A second `tab` returns to the list. The first `ctrl+c` shows the quit notice; the second returns to the shell.
- Help: `esc` closes the overlay.
- CompactAgents: `tab`, `enter` opens the details popup. `esc` closes it. `?` opens Help, and `esc` closes Help.
- Compact: `esc` closes the full-width details overlay. `enter` opens it again.
- Each run quits with `ctrl+c` twice.

### Blocking key route

Reproduction:

```sh
hive --demo .hive/spec/tui-v2/design/Compact.fixture.json --size 80x24
# The fixture starts in the Work details overlay. Press 1.
```

Expected: the details overlay keeps focus. Decision 21 gives an overlay only its scroll keys, its action keys, `ctrl+c`, and `esc` to close it.

Observed: the tab changes to Agents and the footer becomes `[graph]`. See `verify/overlay-key-1.ansi` and `.png`. In `internal/tui/frame.go`, the `details` branch restricts keys only for Agents. The other details regions reach `screenKey` before their overlay route. The manager received inbox record 153. The verifier stopped further source investigation after this report.

### Strict live cell comparison

The top-left board rectangle of each live ANSI capture was converted with `internal/lib/cellgrid`. Main and CompactAgents compare rows 0, 1, 2, and the footer. Help compares the same frame and the approved help box, with token-only checks behind it. TooSmall compares the whole grid.

| Board | Different cells | Observation |
| --- | ---: | --- |
| Main | 67 | Only blank-cell token or bold attributes differ. |
| CompactAgents | 59 | Only blank-cell token or bold attributes differ. |
| Help | 1384 | Only blank-cell token or bold attributes differ. |
| TooSmall | 120 | Only blank-cell token or bold attributes differ. |

No non-space character differs in these comparisons. The non-space tokens and bold attributes also match. The model-level golden tests pass. The live path can lose or extend styling on blank cells; this run does not identify which layer causes that difference. No mask or normalization was added. The strict live comparison is not a pass.

### Visual and plain-text review

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Main long header | `coordinator ● working`, `supervisor ● live`, `needs you 0`, `14:22` | differs | Visible text agrees. Blank-cell attributes differ in the live capture. |
| CompactAgents short header | `coord ● · sup ● · needs you 0 · 14:22` | differs | Visible text agrees. Blank-cell attributes differ. |
| Six tab labels and active underline | `1 Agents` to `6 Logs`, `━` | matches | Main and CompactAgents show the approved labels and underline. |
| Fixed frame and footer hints | `[graph]`, `[list]`, `[details]` | differs | Visible text agrees. Trailing blank-cell attributes differ. |
| Focus on Work | `[list]` then `[details]` | matches | `tab` moves the accent border and the footer word. |
| Main and CompactAgents content | Current TREE view | overridden | The brief keeps the current screen until wave 6. |
| Help title and key map | `keys · ? or esc closes`, `GLOBAL`, `ACTIONS`, `FILTER FIELDS`, `CHAT` | matches | The approved box text and non-space styling agree. No `--demo` hint. |
| Help screen behind the box | Dimmed header and current Agents content | differs | The non-space tokens agree; blank-cell tokens differ. The current content is permitted until wave 6. |
| Help scroll and close | `[help] ↑↓ scroll · ? or esc close` | matches | Down scrolls. `esc` closes. Model tests also check `?` close. |
| CompactAgents popup | Agent name, `[details]`, `esc close` | matches | `enter` opens and `esc` closes it. |
| Compact Work overlay | `[details]`, `esc back` | differs | Open and close work, but `1` leaves the overlay. Decision 21 does not permit this route. |
| Too-small title and size | `terminal too small`, `this pane is 60×16`, `hive needs at least 80×24` | differs | Visible content agrees. Blank-cell attributes differ. |
| Too-small frame | `hive`, `ctrl+c ×2 quit` | matches | No normal screen content is visible. |
| Quit notice | `press ctrl+c again to quit` | matches | The second key exits. |

The checked states have words or glyphs without colour. Colour is not their only signal.

## Cleanup

The scratch program was stopped. Tab `w8R:t2` was closed. `herdr tab list --workspace w8R` shows only verifier tab `w8R:t1`. Both detached scratch worktrees were removed. The scratch binary, home, profiles, scripts, and HTML files were removed after the evidence was saved. The task worktree remains for repair.
