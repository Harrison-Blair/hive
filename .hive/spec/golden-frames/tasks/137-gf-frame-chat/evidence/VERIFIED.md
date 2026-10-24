# Task 137 verification

- Task: `137-gf-frame-chat`
- Verifier: `t137-verifier-1`
- Date: 2026-10-24
- Code checked: `70de1640`
- Verdict: **fail**

V2 found two failures of the spec. The existing tests and the gate pass.
The 14 live fixtures match their revised golden grids with no mask.

## Verification items

| Item | Result | Command and observation | Evidence |
| --- | --- | --- | --- |
| V1 Board copies and export | pass | `sha256sum -c -`: 16 initial copies at `6ce218d0` are OK. The current unchanged `canvas.json` is OK. Export of the 15 current boards gives identical grids. `git status --short .hive/spec/golden-frames/design/` prints nothing. AC10 accounts for the 15 changed boards. | `verifier/board-copy-sha.txt`, `verifier/export.txt` |
| V2 Behavior, red and green | fail | `go test ./internal/... -count=1 -v -run NAMES` passes for 383 test names from the new or changed test files. The pre-change production overlay fails to compile against the new APIs. Thirteen deliberate rule breaks fail their named tests. Each test passes without its break. The additional spec checks fail for F1 and F2 below. | `verifier/behavior-test-names.txt`, `verifier/behavior-all-green.txt`, `verifier/behavior-prechange-red.txt`, `verifier/mutations/`, `verifier/probes.txt` |
| V3 Golden checks | pass | The golden and breakpoint commands below pass. All 14 screen grids and the in-scope component regions pass. The list has 48 stopped checks and 2 kept checks. The four breakpoint checks pass for the fixtures. F1 shows a valid case that those fixtures miss. | `verifier/golden-ste.txt` |
| V4 Live check | pass | The task binary ran all 14 fixtures with `COLORTERM=truecolor`, `--demo FILE`, and `--size 120x40` in herdr. Each capture has zero differing cells. The required key checks pass. The verifier captured the boards and screens as PNGs again. The verifier deleted the scratch session. | `verifier/live/`, `verifier/live-compare.txt`, `verifier/keys-check.txt`, `verifier/visual.md`, `verifier/png/` |
| V6 Board changes | pass | `git diff --no-index` compares each current board with its pinned source. All 15 changed-row lists and old/new hashes agree with the evidence. The differences follow the recorded cases. `go test ./ -run STE -count=1` passes. The existing side-by-side board review remains for user review before landing. | `verifier/board-diffs/`, `verifier/board-changes-check.txt`, `verifier/golden-ste.txt`, `board-review.html` |
| V5 Gate | pass | All five gate commands exit 0. `gofmt -l` and `go mod tidy -diff` print nothing. The linter reports `0 issues.` The race tests pass. | `verifier/gate.txt` |

The V3 commands were:

```sh
go test ./internal/boardexport/ ./internal/tui/ ./internal/tui/component/ -count=1 -v -run 'Golden|Gf|Stopped'
go test ./internal/tui/ -count=1 -v -run 'Retired|EachScreenKeepsTheFrame'
```

## F1: a resize breaks the ask border

- Requirement: README decision 5 and AC3. A resize never breaks a border.
- Code: `internal/tui/component/composer.go:72-82`, `ComposerCard.top`.
- Input: an asker named `manager-project-with-a-long-name`, with four questions.
  Each header is 12 characters: `Question0000` through `Question0003`.
  Each question has two options. `ask.Parse` accepts this request.
- Steps: render the card at 120×40, then send `tea.WindowSizeMsg{Width: 80, Height: 24}`.
- Expected: the resized top border ends with `╮` within the 80 columns.
- Actual: the top row ends inside `Question0002`. The top-right corner is absent.
  `ComposerCard.top` removes the right text when space is short.
  It then truncates the title and the final corner together.

`TestVerifierAskResizeKeepsTheTopRightCorner` fails on the task code.
The same card has its corner at 120×40.
`verifier/probe-captures/ask-resize-before.png` and `ask-resize-after.png` show both states.
These PNGs come from model output, without colour.
Decision 27 defers the 80×24 styling. The border property still applies at that size.

## F2: the config refuses a valid Bubble Tea key name

- Requirement: README decision 18 and AC6. A rebound key follows the Bubble Tea key names.
- Code: `internal/lib/home/config.go:126-135`, `validKeyName`.
- Input:

```toml
home = "/tmp/hive"
[tui.keys]
help = "alt+up"
```

- Expected: the config accepts `alt+up`. It has no duplicate or reserved binding.
- Actual: `home.ReadSettings()` refuses it with `tui.keys.help: "alt+up" is not a key name`.
  The parser requires one rune after the modifier prefixes.
  This excludes named keys such as `up`.

`TestVerifierKeysAcceptBubbleTeaNamedKey` fails on the task code.
It first checks that Bubble Tea names the key `alt+up`.
It then reads the TOML through the real config loader.

## Deliberate rule breaks

Each break used a Go overlay of one production file.
The production files in the worktree did not change.
Each row has its patch and both command outputs in `verifier/mutations/`.
The red run exits 1 for the named assertion. The green run exits 0.

| Break file stem | AC | Rule changed | Test that fails with the break |
| --- | --- | --- | --- |
| `error` | AC2 | Change the error hex from `#e8705f` | `TestConvertErrorNoticePrefix` |
| `chord` | AC3, AC6 | Route the Agents chord to Chat | `TestFrameChordsOpenScreens` |
| `title` | AC3 | Draw the title rule with heavy glyphs | `TestTitleRowGolden` |
| `status-lines` | AC4 | Show all status rows below 30 rows | `TestVerifierStatusRowsHeight` |
| `status-join` | AC4 | Join segments with ` / ` | `TestStatusRowGolden` |
| `throttle` | AC4 | Remove the 300 ms throttle | `TestStatusCommandRunsAtMostOnceEach300ms` |
| `demo-command` | AC4 | Permit a status process in demo mode | `TestStatusCommandRunsOnlyWhenItApplies` |
| `notice` | AC5 | Remove the startup notice before an error | `TestFooterShowsStartupNoticeBeforeError` |
| `box` | AC6 | Ignore the box chord on two-box screens | `TestFrameBoxFocusKeys` |
| `placeholder` | AC7 | Show the normal placeholder for a blocked coordinator | `TestVerifierChatStatePlaceholders` |
| `send` | AC7 | Remove the no-coordinator check before send | `TestChatSendWithNoCoordinatorKeepsTheText` |
| `earlier` | AC7 | Put the earlier count at column 50 | `TestChatTallTranscriptFollowsItsEnd` |
| `messages` | AC7 | Remove the blank row between messages | `TestChatBlankRows` |

The pre-change run restores production files from `55df4ace` and keeps the task tests.
It fails on absent APIs such as `Error`, `Composer`, and `Settings.TUI`.
That compile failure does not prove each behavior. The deliberate breaks give behavioral red runs for the rules above.
This report does not claim a separate red run for every assertion of all 383 tests.

## Reproduce the additional checks

The test source is `verifier/probes_test.go.txt`.
Use a Go overlay to add it without a source edit:

```sh
mkdir -p .hive/tmp/t137-recheck
python3 - <<'PY'
import json, pathlib
r = pathlib.Path.cwd()
e = r / '.hive/spec/golden-frames/tasks/137-gf-frame-chat/evidence/verifier'
p = r / '.hive/tmp/t137-recheck/probes.json'
p.write_text(json.dumps({'Replace': {
    str(r / 'internal/tui/t137_verifier_test.go'): str(e / 'probes_test.go.txt')
}}))
PY
go test -overlay .hive/tmp/t137-recheck/probes.json ./internal/tui/ -count=1 -v -run '^TestVerifier'
```

Expected on `70de1640`: F1 and F2 fail. The placeholder and status-height checks pass.
Both passing checks also fail under their deliberate breaks.

## Limits and cleanup

The verdict covers this task. It does not cover the wave C bodies or their component regions.
The user must review the changed grids before landing, as AC10 requires.
The first live attempt lacked the scratch home marker. The verifier corrected the setup and repeated every capture.
Only the repeated captures appear in the live results.
The verifier stopped and deleted `t137-check-verifier` after the live checks.
The verifier removed the scratch binary, overlays, board exports, browser profiles, and scratch home after recording the evidence.
No production file, worker `V<n>.md` file, or plan tracker changed during verification.
Raw captures keep terminal padding. Patch files keep their diff context.
