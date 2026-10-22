# Verification: 29-frame-layout

Verdict: **pass**. Verifier: `t29-verifier-2`. Date: 2026-10-22.

Task code: `b95da586`. `git merge dev` reports `Already up to date`.
The base is `dev` at `02ffae7f`. This report replaces the previous failed
verdict. The evidence of that verdict stays in `verify/`.
No project source file was changed in this worktree.

## Items

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | `verify-r2/green.txt`, `base-compile.txt`, `base-red.txt`, `pre-repair-red.txt`, five `*-mutation.txt` files, and `restored.txt`. See the base-run limit below. |
| V2 Live check | pass | `verify-r2/*.ansi`, `*-pane.json`, `*.png`, `live.txt`, `compare.txt`, and `*-diff.txt`. The repaired compact overlay keeps screen and view keys. |
| V-gate Diff scope and gate | pass | `verify-r2/scope.txt` and `gate.txt`. Only approved paths and evidence differ. All five commands pass. No flake occurred. |

## V1 Tests

The independent green run used:

```sh
go test ./internal/tui ./internal/tui/component -run 'Test(Frame|FooterShowsStartupNoticeBeforeError|DemoUIStateSetsTheFrameState|DemoRestartPromptYesDoesNotRestart|Help|EachScreenKeepsTheFrameAtEachBreakpoint|TabMovesTheAccentBorder|DetailsOverlayBelow120Columns|Overlay|WideDetails|CtrlCTwiceQuitsInEachFocusState|NormalModeUsesTheFrame|HeaderNeedsYouCountIsAccent|DemoActionKeysShowTheNotice)' -count=1 -v
```

The gate also runs each old test, the CLI tests, and the whole-grid TooSmall
test. The model golden checks keep the token and bold check on every cell,
including a space cell. The Help check uses the approved box, rows 6-34 and
columns 18-101, with token-only checks behind it.

A detached scratch worktree at the base received the current frame and
component header test files. The full frame test file fails to compile:
`undefined: tabAgents` and other frame API names. For a behavioral red run,
the scratch copy removed only functions that need the absent API and the
unused `fmt` import. The remaining new tests fail: Main frame (461 cells),
each screen board frame, startup notice, demo restart, Help, normal frame,
and the component needs-you count token.

Limit: the base compile failure proves an absent API. It does not prove each
key assertion independently. The earlier verification used the same limit.
The current task code passes each test. A separate run used the current
tests with `frame.go` from `59e952b6`, before the repair. It reproduces the
compact overlay failures. The action-key and wide-details guard tests pass
before the repair; deliberate breaks prove both guards below.

Each deliberate break used a second detached scratch worktree. After each
run, the original file was restored.

| Deliberate break | Expected failure observed |
| --- | --- |
| Global help route changed from `?` to `BROKEN?` | `TestHelpKeys` |
| Wide details panel focus set to false | `TestTabMovesTheAccentBorder` |
| Compact overlay route disabled | `TestOverlayKeepsScreenAndViewKeys` |
| Overlay action key list empty | `TestOverlayGivesActionKeysToTheScreen` |
| Overlay route applied at each width | `TestWideDetailsKeepsScreenKeys` |

The restored tests pass. The mutation worktree has no source diff.
`V1.md` records the approved changes to old test expectations.

## V2 Live check

Built the task binary with `go build -o .hive/tmp/verify29r2/hive .`.
Created scratch tab `w8T:t2`, pane `w8T:p2`, in this verifier's workspace.
Herdr client and server are 0.9.5, protocol 22. The pane fits 120x40 and has
57 rows. `COLORTERM=truecolor` and a scratch `HIVE_HOME` apply to the pane.
Demo mode implies no startup. No live coordinator or supervisor was started.

The first setup attempt had no scratch home marker and exited with
`not a hive home`. After a marker was copied to the scratch home, all five
fixture runs were repeated. The retained captures are from the valid runs.

Each run used the built `hive --demo ABSOLUTE_FIXTURE --size WxH`:
Main and Help at 120x40, CompactAgents and Compact at 80x24, and TooSmall at
60x16. Each capture used:

```sh
herdr pane read w8T:p2 --source visible --format ansi
```

Two reads 500 ms apart were equal. Each saved pane record has scroll offset
0. The CLI returns ANSI text, with no truncated result. The top-left board
rectangle was compared with `internal/lib/cellgrid`. Main, CompactAgents,
and Compact compare rows 0-2 and the footer. Help compares that frame and
the approved box fully, with token-only checks behind it. TooSmall compares
the whole grid. Only live space-cell attributes use the manager rule.

| Capture against board | Strict differences | Space-cell attributes | Live differences |
| --- | ---: | ---: | ---: |
| Main | 67 | 67 | 0 |
| CompactAgents | 59 | 59 | 0 |
| Help | 1384 | 1384 | 0 |
| TooSmall | 120 | 120 | 0 |
| Compact | 37 | 37 | 0 |
| Compact after `1` | 37 | 37 | 0 |
| Compact after `6`, `[`, `]`, `tab` | 37 | 37 | 0 |
| Compact after `esc`, `enter` | 37 | 37 | 0 |
| Help after `2` | 1384 | 1384 | 0 |

Each non-space character, token, and bold flag agrees in the checked area.
The Help content behind the box has the permitted wave 6 scope.

### Keys

- Main: `tab` keeps graph focus. `?` opens Help. Down scrolls it; `esc`
  closes it. `3`, `tab` moves Work focus to details. Another `tab` returns
  to the list. The first `ctrl+c` shows the quit notice; the second exits.
- Help: `2` keeps Help open over Agents. `?` closes it. Another `?` opens
  it again. `esc` closes it.
- CompactAgents: `tab`, `enter` opens the popup. `3` keeps it open over
  Agents. `esc` closes it. `?` opens Help; `esc` closes Help.
- Compact: `1`, `6`, `[`, `]`, and `tab` keep Work Tasks details open.
  `esc` returns to the list. `enter` opens full-width details again.
  After `esc`, `1` selects Agents, as expected for list focus.
- Each fixture exits after two `ctrl+c` keys. TooSmall keeps its fixed quit
  hint after the first key.

### PNG and plain-text review

Each captured UI state has a fresh PNG, except the shell after exit.
Each of the five boards has a fresh `*-board.png`. A scratch script converts
SGR to an HTML `pre`. Firefox captures that HTML and each original board:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://INPUT
```

Each capture uses a new profile folder. The board and ANSI HTML use different
font metrics and padding. Cell comparisons use ANSI, not PNG pixels.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Main long header | `coordinator ● working`, `supervisor ● live`, `needs you 0`, `14:22` | matches | Decision 2; live grid agrees. |
| Compact short header | `coord ● · sup ● · needs you 0 · 14:22` | matches | Decision 2; live grid agrees. |
| Six tab labels | `1 Agents` through `6 Logs` | matches | Decision 3. |
| Active tab | `━` underline | matches | State is visible without colour. |
| Fixed frame and footer | `[graph]`, `[list]`, `[details]` and key hints | matches | Decisions 1, 4, 21. |
| Focus change on Work | `[list]` then `[details]` | matches | `tab` moves the accent border and footer word. |
| Main and CompactAgents content | Current TREE screen | overridden | The brief keeps the current content until wave 6. |
| Help title | `keys · ? or esc closes` | matches | Approved Help box. |
| Help global and action columns | `GLOBAL`, `ACTIONS`, key names | matches | Cell check agrees. No `--demo` hint. |
| Help filter and chat sections | `FILTER FIELDS`, `CHAT` | matches | Cell check agrees. |
| Screen behind Help | Header and tab words remain visible | matches | Non-space cells use `border`; decision 5. |
| Agents content behind Help | Current TREE screen | overridden | AC5 defers full content to wave 6. |
| Help scroll and close | `[help] ↑↓ scroll · ? or esc close` | matches | Down scrolls; both close keys work. |
| Agents popup | Agent name, `[details]`, `esc close` | matches | `enter` opens it; screen keys keep it open. |
| Compact Work details | `[details]`, `esc back`, `3 Work` underline | matches | The repaired route keeps screen and view keys. |
| Compact Work content | Current task details | overridden | Content belongs to wave 6. |
| Too-small notice | `terminal too small`, `this pane is 60×16`, `hive needs at least 80×24` | matches | Whole live grid agrees. |
| Too-small frame | `hive`, `ctrl+c ×2 quit` | matches | No normal screen content. |
| Quit notice | `press ctrl+c again to quit` | matches | The next key exits. |
| Blank-cell token and bold | Space | overridden | Approved live rule. Model checks stay strict. |

The checked states have a word or a glyph without colour. Colour is not the
only state signal.

## V-gate

`scope.txt` records `git diff dev...HEAD --stat`. The diff has only the
approved source paths and task evidence. The README change stays in Manager
decisions. The component change fixes the needs-you count token. The two
CLI test changes have the approved header and demo-key scope.

Full output is in `verify-r2/gate.txt`:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

The first two commands print nothing. Vet prints nothing. Lint reports
`0 issues.` The race test passes on the first run, including the known
Health string-reader test. No retry was needed.

## Cleanup

The demo programs exited. The scratch tab was closed. The tab list for
`w8T` has only verifier tab `w8T:t1`. Both detached worktrees were removed.
The scratch binary, home, profiles, scripts, and HTML were removed after the
evidence was saved. The task worktree and previous evidence remain.
