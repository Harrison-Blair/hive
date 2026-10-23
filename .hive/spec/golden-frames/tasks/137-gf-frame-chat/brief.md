# Golden frames wave 1: the frame, the status rows, the keys, and Chat

| Item | Value |
| --- | --- |
| Feature | `golden-frames`, goal `g12-golden-frames`, wave B |
| Tier | strong |
| Depends on | `136-golden-frames-spec`, landed on `dev` at `55df4ace`. Base: `dev` after that landing. |
| Blocks | each wave C task of `golden-frames` |
| Grant | Goal `g12-golden-frames`. The user signed off the boards in chat `c248` and ask `a10`. The user chose the golden retirement in ask `a11`. |
| Evidence | The `evidence/` folder beside this brief. Put `VERIFIED.md` there. |

## Overview

This task builds the frame of the restyled TUI and the Chat screen. Bare
`hive` opens Chat. No screen has a header row or a tab bar. A chord opens each
other screen. The task adds the status rows, the notice line, the key map, and
the `error` palette token. Chat matches its 14 boards cell for cell.

The spec is `.hive/spec/golden-frames/README.md` and
`.hive/spec/golden-frames/screens/chat.md`. The spec is the authority for each
behavior. This brief gives the scope and the proof. When the spec, a board, and
this brief do not agree, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`. Do not choose.

The six other screens (Work, Friction, Health, Logs, Agents, Help) are the wave
C tasks. In this task, each of these screens gets the frame of README
decisions 2, 3, 5, and 6: a title row with the screen name, the rule on row 1,
the notice line, and the status rows. Its body below stays as it is today. The
Help overlay keeps its layout. Its text lists the keys of README decision 16 in
place of the old keys. Do not restyle a body. A wave C task does that.

## User decision: ask a11, 2026-10-23

1. Retire each `tui-v2` golden grid that the new frame replaces. The grid
   files stay in `.hive/spec/tui-v2/design/` as history. Do not delete or
   change them. The golden checks stop checking them.
2. Rewrite each layout test or key test whose rule the new spec replaces. Each
   rewritten test cites the decision or the rule that replaces the old rule, in
   a comment. A test of a rule that the spec keeps stays and passes unchanged.
3. Retire the `Compact*` grids. Keep the property tests at 80×24 on each
   screen: no line wider than the pane, no broken border, and the selected
   item visible.
4. A `tui-v2` grid that still matches the new code stays checked. One example
   is `TooSmall`, because README decision 4 keeps the too-small notice.

## Design

The boards are pinned at version `1792795470-106f` of the canvas. The source
folder is `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/`
with `SHA256SUMS` beside it. Copy from this folder. Do not fetch the canvas.

- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChat.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatEmpty.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatNoCoordinator.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatEvents.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatTall.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatMarkdown.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatNotice.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatAsk.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatAskInput.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatAskMulti.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatAskPreview.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatAskStopped.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatAskReview.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfChatAskMany.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/GfComponents.dc.html`
- `/home/penguin/source/hive/.hive/tmp/golden-frames-1792795470-106f/project/canvas.json`

Each state has a glyph or a word in plain text. Colour can add to a state.
Colour must not be the only signal of a state. The user is colourblind. When a
board shows a state by colour alone, stop and report it.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/golden-frames/design/` | AC1, AC8. New. The 15 boards, `canvas.json`, the 15 golden grids, and the fixtures of the 14 screen boards. |
| `internal/lib/cellgrid/` | AC2. The `error` token and the converter map. |
| `internal/boardexport/` | AC1, AC8, AC9. The sizes of the Gf boards, the design folder of the feature, the golden checks, and the retired list. |
| `internal/lib/home/config.go` | AC4, AC6. The `[tui.status]` and `[tui.keys]` tables. |
| `internal/tui/` | AC3 to AC9. The frame, the status rows, the notice line, the keys, Chat, the interim frame of the other screens, and the Help text. |
| `internal/tui/component/` | AC3 to AC7. The components of the frame and of Chat. Remove the tab bar and the header when no caller remains. |
| `cmd/root_demo_test.go` | AC8. Only to point the demo checks to the new boards. |
| `cmd/` | Only to pass the config to the TUI, if the wiring needs it. |

Edit no other file. Do not edit `.hive/spec/golden-frames/README.md`, a
screen file, or a file in `.hive/spec/tui-v2/`.

## Requirements

### AC1 The boards

- Copy the 15 boards and `canvas.json` of the Design section into
  `.hive/spec/golden-frames/design/`. Do not change a byte. `sha256sum -c`
  against `SHA256SUMS` passes for each copied file.
- Export each golden grid with the exporter of `tui-v2` decision 41. The
  exporter learns the Gf board sizes of README decision 23 and the design
  folder of this feature. Commit the 15 `.grid` files.
- The exporter refuses a board with a hex value outside the palette, a symbol
  outside README decision 7, or a size outside the table.

### AC2 The error token

- README decision 21. The palette has nine tokens. `error` is `#e8705f` with
  the grid letter `e`. The converter maps `#e8705f` to `error`.
- The tests of the converter use hand-written ANSI input and hand-written
  expected grids, as `tui-v2` decision 40 gives.

### AC3 The frame

- README decisions 1 to 7. Bare `hive` opens Chat. No screen has a header row
  or a tab bar. Each screen other than Chat has the title row on row 0 and the
  rule on row 1. In this task the title row shows the screen name in `accent`
  bold. A wave C task adds the facts and the right text.
- A chord opens its screen at full size. `esc` returns to Chat. The same chord
  on its own screen returns to Chat.
- `ctrl+]` moves the focus between two boxes, both ways, on a screen with two
  boxes. It does nothing on Chat.
- A popup or an overlay dims the full screen behind it. A resize keeps the
  popup and its scroll position and never breaks a border.

### AC4 The status rows

- README decisions 8 to 13. Each segment, each glyph form, the ` · ` join, the
  clock at the right edge, and the removal from the right.
- `[tui.status] lines` gives one to three rows of segment names. Hive refuses a
  value outside 1 to 3 and an unknown segment name, with an error that names
  the key. Below 30 rows, only the first row shows.
- `[tui.status] command` runs at most once in 300 ms, never on the input loop,
  with the JSON object of decision 12 on its standard input. A failure gives
  an empty segment and one log event. Demo mode runs no process.

### AC5 The notice line

- README decisions 14 and 15. The order of the notices, the tokens, and the
  time when each notice goes away.

### AC6 The keys

- README decisions 16 to 20. The chords, `ctrl+o`, `ctrl+r`, `ctrl+]`, and
  `ctrl+c` twice. The textarea key map does not take `ctrl+g`, `ctrl+t`, or
  `ctrl+f`. `→` moves the cursor one character forward.
- `[tui.keys]` can rebind each chord and `?`. Hive refuses a config that binds
  one key to two actions, and a config that binds `tab`. The error names the
  key.
- No function key, no `tab` region move, and no normal mode on Chat. `i`, `j`,
  `k`, `g`, `G`, and `ctrl+]` do nothing on Chat.
- Log each chord and each screen change as a TUI user action, as `AGENTS.md`
  gives.

### AC7 Chat

- Each rule of `.hive/spec/golden-frames/screens/chat.md`: the transcript
  (rules 1 to 8), the composer (rules 9 to 13), and the ask card (rules 14 to
  19). Each state of its state table, also the states with `no board`.
- The rules of `tui-v2` that `chat.md` keeps stay. The prompt format, the
  routing, the retries, and the ask record format do not change.

### AC8 The golden checks

- `tui-v2` decisions 42 and 44 for each of the 14 screen boards. The fixture is
  in `.hive/spec/golden-frames/design/`. The check renders `View()` at the
  board size with the TrueColor profile and compares each cell with no mask.
- `GfComponents` has a golden check of each component on the sheet, as the
  `tui-v2` component check does.
- Add no mask. A mask needs the approval of the user.

### AC9 The retirement

- Follow the user decision of ask a11 above. One list in the test code names
  each retired `tui-v2` grid. Each entry gives the reason in one line.
- A test checks that each `tui-v2` grid is in the retired list or passes its
  golden check. A grid that is in neither fails the test.
- The property tests of `tui-v2` decision 46 run on each screen at 79×24,
  80×24, 119×40, and 120×40, with the new frame.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The 15 boards are copied byte for byte and their grids are exported | V1 |
| AC2 | The `error` token is in the palette and the converter | V2 |
| AC3 | The frame has no header or tab bar, the chords open screens, and `esc` returns | V2, V4 |
| AC4 | The status rows and `[tui.status]` follow decisions 8 to 13 | V2, V4 |
| AC5 | The notice line follows decisions 14 and 15 | V2, V4 |
| AC6 | The keys and `[tui.keys]` follow decisions 16 to 20 | V2, V4 |
| AC7 | Chat follows each rule and state of `chat.md` | V2, V3, V4 |
| AC8 | The 14 screen boards and the component sheet pass the golden check | V3, V4 |
| AC9 | The retired grids are listed and the property tests pass | V3 |

## Verification items

### V1 Board copies and export

- Proves: AC1.
- Command: `sha256sum -c` of the copied files against `SHA256SUMS`. Run the exporter for the Gf boards. Then `git status --short .hive/spec/golden-frames/design/`.
- Expected: each file is `OK`. A second export gives no change in the grids.

### V2 Behavior tests, red and green

- Proves: AC2 to AC7.
- Command: `go test ./internal/... -count=1 -run 'NAMES'` for each new or rewritten test.
- Expected: each test fails on the code before the change, for the reason that the test names. Each test passes after the change. Record both runs. For each rule that a test checks, also break the rule on purpose and show that the test fails. Record at least 10 deliberate breaks across AC2 to AC7.

### V3 Golden checks

- Proves: AC7, AC8, AC9.
- Command: the golden tests of `internal/boardexport` and `internal/tui` with `-count=1 -v`.
- Expected: the 14 screen boards and the component sheet pass with no mask. The retired list test passes. The property tests pass at each of the four sizes.

### V4 Live check

- Proves: AC3 to AC8.
- Command: `tui-v2` decision 43 for each of the 14 screen boards, with `COLORTERM=truecolor hive --demo FILE --size 120x40` in a herdr pane of that size or larger. Then in the same demo, press each chord, `esc`, `?`, `ctrl+o`, `ctrl+r`, and `ctrl+c` twice, and capture each result.
- Expected: each capture matches its golden grid with no difference. Each key gives the screen of the spec. Store the captures and the diffs in the evidence. Close each pane that the check made.

### V5 Gate

- Proves: AC1 to AC9.
- Command: the gate of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.

## Rules

- Do not run bare `hive` against the real home. Use `--demo` or `HIVE_HOME="$PWD"`.
- A live check gives each agent and pane a unique name with the prefix `t137-check-`.

## Out of scope

The bodies of Work, Friction, Health, Logs, Agents, and the Help layout. The
80×24 forms (README decision 27). Each item of README decisions 24 to 32.
`main`.

## Changes

- 2026-10-23, manager: C8. Chat has 14 boards, as the Design section gives.
- 2026-10-23, manager: C7 scope. The `GfComponents` rows of the wave C components (TaskCard, WaveColumn, Rail, Popup, CheckTile, DensityStrip, SeenBar) are out of scope. The component check of this task covers only the frame and Chat components of the sheet. The test excludes the wave C rows by component region, not by a mask. The evidence lists them as wave C.
