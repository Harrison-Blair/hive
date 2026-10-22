# Independent verification: 34-screen-chat

## Land repair verification: t34-verifier-3

Date: 2026-10-22. Verdict: **fail**.

The manager limited this run to the repair after `5bbe5781` and permitted
V1 to carry forward. Evidence for this run is in `repair-verifier/`.

| Item | Result | Evidence and observation |
| --- | --- | --- |
| Repair scope | pass | Reviewed both repair merges and their combined diffs. The README keeps the Chat, Logs, and Friction decisions. `tui.go` keeps the Chat feed and Logs fields and removes the unused viewports and layout method. The duplicate `checkBoard` helper was removed; the Friction helper keeps the same grid comparison. Other source changes come from dev. The two non-merge commits change evidence only. |
| V1 Tests | pass | Prior V1 carries forward under the manager instruction. `green.txt` runs all Chat tests again. `token-red.txt` records a deliberate change from `dimStyle` to `titleStyle` for the earlier-count note: the Chat golden fails. After restoration, both golden checks pass in `golden-restored.txt`. No mutation remains. |
| V2 Live check | pass | Built commit `5682e84c` and ran Chat at 120x40 in scratch pane `wAB:p2`, tab `wAB:t2`, with a scratch home. Two reads 500 ms apart agree. The pane has scroll offset 0 and 57 visible rows; the CLI supplies no truncated field. `compare.txt` reports zero differences under the live-space rule. `keys.txt` checks esc to feed, printable input, and demo send refusal. ANSI captures, pane metadata, and fresh board/live PNGs are included. |
| V-gate Diff scope and gate | fail | The full gate passed on `5682e84c`, after merging dev `5a4b29e6`. Format and tidy printed nothing; vet, lint, and race tests exited 0. No flake retry was needed. At the final refresh, dev advanced to `7ea7497d`. Its merge conflicts in `internal/tui/shell_test.go` and `testdata/ste-baseline.json`. Thus the required gate on current dev cannot pass in this run. See `merge.txt`, `scope.txt`, and `gate.txt`. |

The failed final merge was aborted. The earlier clean dev merge remains on
this branch. A worker must repair the two conflicts. A fresh verifier must
then run the checks against the merged code.

### Visual review

Compared the fresh Chat board and live PNGs. The HTML wrappers have different
margins and line heights. The cell comparison checks the terminal content.
The plain capture is `repair-verifier/Chat-plain.txt`.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | coordinator working; supervisor live; needs you 0; 14:22 | matches | Fixed fixture |
| Tab bar | 2 Chat; underline | matches | Decision 3 |
| Feed title and earlier count | chat · noise normal; ↑ 42 earlier | matches | Chat decisions |
| Time and name gutter | you; coordinator; times | matches | Decision 30 |
| User messages | you; message text | matches | Decision 15 |
| Folded event | · manager-hi…; event text | matches | Decision 15 |
| Markdown heading and strong text | Interrogate, round 2; Recommended | matches | Decisions 28–31 |
| Inline code and code block | spawned=false; hive agent stop manager-hive | matches | Decisions 29, 31 |
| Table | Choice; Answer; Pros; Cons; rule glyphs | matches | Decision 32 |
| List | • | matches | Decision 29 |
| Feed outline and scroll mark | border glyphs; ┃ | matches | Decisions 5, 27 |
| Input panel | message; >; █ | matches | Decision 15 |
| Focus and footer | [input]; [feed]; focus key hints | matches | Decisions 4, 21 |
| Demo refusal | demo mode: no action | matches | Decision 35 |
| Blank-cell styles | space | overridden | Manager live-space rule |

Closed the scratch tab. Removed the scratch binary, home, HTML, helper, and
browser profiles. `tabs-after-cleanup.json` shows only the original agent tab.
Final evidence remains. The previous verification record follows.

---

Verifier: `t34-verifier-2`. Date: 2026-10-22.
Verdict: **pass**.

Verified commit: `72d44ff71d57e85beab8bb57ee2878b75db84441`.
`git merge dev` merged `13210c3f64f03f3228a5140688ac9f4c2c0106a9` without a conflict.
All paths below are relative to this evidence folder.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | `verifier/base-red.txt`, `green.txt`, `key-red.txt`, `token-red.txt`, `kept-red.txt`, `restored-green.txt` |
| V2 Live check | pass | `verifier/compare.txt`, ANSI captures, pane records, key observations, board and screen PNGs |
| V-gate Diff scope and gate | pass | `verifier/merge.txt`, `diff-stat.txt`, `gate.txt` |

## V1

Command on base and task code:

```sh
go test -count=1 -v -run 'TestChat|TestCompactChat|TestDemoChat' ./internal/tui/
```

The base run used an archive of `dev` at `13210c3f`. It replaced the old
`chat_screen_test.go` with the task's `chat_test.go`. The three new noise
constant names were replaced with the corresponding base string labels so
that the tests could compile. No assertion was removed. The first test,
`TestChatMatchesItsBoard`, failed on the base.

| Test | Base | Task |
| --- | --- | --- |
| TestChatMatchesItsBoard | fail | pass |
| TestCompactChatMatchesItsBoard | fail | pass |
| TestChatFixtureScrollPutsTheMessageAtTheTop | fail | pass |
| TestChatLoadsRecentConversationAndInboxEvents | fail | pass |
| TestChatKeysScrollTheFeed | fail | pass |
| TestChatKeysFollowTheFocusRules | fail | pass |
| TestDemoChatSendShowsTheNotice | fail | pass |
| TestChatFitsEachSize | fail | pass |
| TestChatProgramRunsWithStringInput | fail | pass |
| TestChatNoiseLevelsFilterInboxKinds | pass: kept-behavior guard | pass |
| TestChatCompactFooterHints | pass: kept-behavior guard | pass |
| TestChatStartFromProjectsShowsRefusalAndBlocksInput | pass: existing test | pass |
| TestChatRefusedSendShowsNoticeAndKeepsDraft | pass: existing test | pass |
| TestChatSendRunsInBackgroundAndKeepsFailedDraft | pass: existing test | pass |

The manager clarified that a new test of kept behavior can pass on the base
when a deliberate break fails that test. This applies to the noise and
compact-footer guards above.

All deliberate changes were made in a scratch archive of the task code:

| Deliberate change | Command (`go test -count=1 -run NAME ./internal/tui/`) | Result |
| --- | --- | --- |
| Feed end key `G` becomes `H` | NAME=TestChatKeysScrollTheFeed | fail: line-40 is absent |
| Earlier-count token uses titleStyle instead of dimStyle | NAME=TestChatMatchesItsBoard | fail: golden cells differ |
| Normal noise excludes InboxQuiet | NAME=TestChatNoiseLevelsFilterInboxKinds | fail: quiet is absent |
| Compact feed hint becomes `i broken` | NAME=TestChatCompactFooterHints | fail: expected footer is absent |

The last two commands ran together with a test-name alternation. Both tests
failed. After all scratch changes were restored, the full Chat test command
passed again. Production source files were not changed by these checks.
The normal-mode tests use scratch homes and no live coordinator or supervisor.

## V2

Built the verified code with:

```sh
go build -o .hive/tmp/t34-verifier/hive .
```

Created scratch tab `w9Z:t2`, pane `w9Z:p2`, with `COLORTERM=truecolor` and
`HIVE_HOME` set to `.hive/tmp/t34-verifier/home`. This home had its own marker.
For each board, ran the built binary with:

```text
hive --demo WORKTREE/.hive/spec/tui-v2/design/Chat.fixture.json --size 120x40
hive --demo WORKTREE/.hive/spec/tui-v2/design/CompactChat.fixture.json --size 80x24
```

Each capture used `herdr pane read w9Z:p2 --source visible --format ansi`.
Two reads 500 ms apart were equal. Pane records show scroll offset 0 and 57
visible rows. The CLI exposes no truncated field; both captures contained all
board rows. A scratch Go program cropped the top-left board rectangle, used
`internal/lib/cellgrid.Convert`, and compared it with the committed grid.
It ignored token and bold differences only when both cells were spaces.

| Board | Strict differences | Differences under live rule |
| --- | --- | --- |
| Chat | 2644 | 0 |
| CompactChat | 675 | 0 |

`verifier/keys.txt` records successful live assertions. Chat checks covered
send refusal, esc to feed, printable input, tab, help open/close, screen
switches, top/end, page up/down, line movement, and the quit notice.
CompactChat checks covered typing, send refusal, esc, page up, end, input
focus, tab, and double ctrl+c. Page sequences were sent as ESC [5~ and
ESC [6~ through `send-text`. Demo send kept the draft and showed
`demo mode: no action`.

### Visual and colour-free review

Captured both approved absolute board paths from the brief again. A scratch
script converted the ANSI captures to HTML preformatted text. Each board
and screen was captured with Firefox using a new scratch profile:

```text
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://SOURCE
```

PNG files: `verifier/Chat-board.png`, `Chat-screen.png`,
`CompactChat-board.png`, `CompactChat-screen.png`. Browser margins and font
scale differ between the board wrapper and the scratch HTML. The cell
comparison verifies the terminal content independently of those wrappers.
ANSI-stripped captures were also checked for each state word or glyph.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Wide header | coordinator ● working; supervisor ● live; needs you 0; 14:22 | matches | Decision 2 |
| Compact header | coord ●; sup ●; needs you 0; 14:22 | matches | Decision 2 |
| Tab bar | 2 Chat and ━ underline | matches | Decision 3 |
| Feed title | chat · noise normal; compact chat | matches | Task manager decisions |
| Earlier-message count | ↑ 42 earlier | matches | Task manager decisions |
| Wide time/name gutter | 14:02 you; 14:21 coordinator | matches | Decision 30 |
| Compact message labels | time and name on their own line | matches | Decision 30 |
| User messages | you and message text | matches | Decisions 15, 30 |
| Folded event | · manager-hi… and one-line event | matches | Decision 15 |
| Markdown heading and strong text | Interrogate, round 2; Recommended | matches | Decisions 28–31 |
| Inline code and code block | spawned=false; hive agent stop manager-hive | matches | Decisions 29, 31 |
| Markdown table | Choice, Answer, Pros, Cons; │ ┼ ─ rules | matches | Decision 32 |
| Markdown list | • items | matches | Decision 29 |
| Feed border and scroll mark | panel outline; ┃ | matches | Decisions 5, 27 |
| Input panel | message; >; cursor █ | matches | Decision 15 |
| Focus | [input] or [feed]; cursor in input | matches | Decisions 4, 6 |
| Footer hints | enter send; esc leave input; i type; page hints | matches | Decision 21 |
| Demo send state | demo mode: no action | matches | Decision 35 |
| Space-cell token or bold | blank cell | overridden | Manager live-space rule |

The states remain identifiable without colour. The compact header uses the
approved board's plain-text role labels and dots.

## V-gate

Ran from the worktree root after the merge:

```sh
git diff dev...HEAD --stat
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

The diff contains only allowed source paths, the brief copy, manager decisions,
STE baseline, and task evidence. No component change was needed.
`gofmt -l` and `go mod tidy -diff` printed nothing. Each command exited 0.
Lint reported `0 issues.` The complete output is in `verifier/gate.txt`.
No known-flake retry was needed.

## Cleanup

Closed scratch tab `w9Z:t2`. `verifier/cleanup.json` shows only the verifier's
original tab. Removed the scratch binary, homes, source archives, browser
profiles, and helper scripts after recording the results. Final evidence is
kept in this task folder. No real-home record was used by the live demo.
