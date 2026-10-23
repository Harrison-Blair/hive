# hive ask: a question card in the Chat dock

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g6-g6-hive-ask`, wave A, task 2 of 2 |
| Tier | strong |
| Depends on | tasks 126 and 129. Base: `dev` at e0799d49 or later |
| Grant | goal `g6-g6-hive-ask`: the user approved `plans/hive-ask.md` and the boards (canvas version 88). Build and land on `dev`. |
| Evidence | `.hive/spec/tui-v2/tasks/131-hive-ask/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The user approved the plan `/home/penguin/source/hive/plans/hive-ask.md` on
2026-10-22. Read it first. Task 126 did decisions 8 and 9 and the header
count at 0. This task does decisions 1 to 7:

1. The command `hive ask`. Any agent can run it. It returns at once.
2. The question card in a dock between the Chat feed and the input.
3. The answer goes to the asker as a `⬢` prompt. When the asker stopped, the
   answer goes to the parent of the asker.
4. The header count `? N asks` shows the number of open asks.
5. The hive role skills use `hive ask`.

The TUI must match each board cell for cell: the glyph, the colour token, and
the bold flag.

Rule slips: the ask boards are older than task 126. They use the short
palette, with no `user` token and no `code` token, and some footers and header
columns can differ from the spec rules of task 126 (decisions 79 to 90). When
a board breaks a spec rule that exists and adds no new design, follow the
rule, do not stop, and list each case (board, row, columns, rule) in
`evidence/board-rule-cases.md` and in your report. Stop and report only when
no rule applies.

## Design

The approved boards (canvas "Hive TUI Redesign",
https://claude.ai/artifact/VrhcA98g9G2G9cFQao5Z9h), in
`/home/penguin/source/hive/.hive/tmp/g6/boards/`:

- `ChatAsk.dc.html` (120×40): one open card with one question.
- `ChatAskMulti.dc.html` (120×40): two cards. The first card has two
  questions, with multi-select and an open `note` line.
- `ChatAskPreview.dc.html` (120×40): a question with a preview for each
  option.
- `ChatAskStopped.dc.html` (120×40): the asker stopped. The answer goes to
  its parent. The feed shows an answered ask.
- `CompactChatAsk.dc.html` (80×24): one card at 80 columns.

Each board has a `NAME.txt` text copy beside it. The `.dc.html` file is the
authority. The staged folder has no `support.js`. Use the copy in
`.hive/spec/tui-v2/design/`. The other boards in the folder are from task 126.
Do not copy them again.

Follow the visual check rule in the worker skill. The user is colourblind.
Each state has a glyph and a word. Colour only adds to a state.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/design/<Board>.dc.html`, `.grid`, `.fixture.json` for the 5 boards | AC1. New copies, grids, and fixtures. |
| `internal/boardexport/check.go` and its tests | AC1. Add the 5 boards to `Sizes`. |
| `.hive/spec/tui-v2/design/*.grid` of the boards that AC8 changes | AC8 only. |
| `.hive/spec/tui-v2/README.md` | AC2. Decisions 91 and later. |
| `cmd/ask/`, `cmd/root.go` | AC3. The command wiring. |
| `internal/ask/` | AC3, AC5. A new package for the ask records, the answer, and the delivery. |
| `internal/tui/*.go`, `internal/tui/component/*.go`, and their tests | AC4 to AC6. |
| `.agents/skills/hive-*/SKILL.md` | AC7. |
| `reference/` files that list the hive commands | AC3. Add `hive ask`. |
| `testdata/ste-baseline.json` | To add the brief copy with 0, and to lower a count. |

Do not edit other files. Do not add a mask. Do not change the exporter
except `check.go`. When you must edit a file that is not in this table, stop
and tell the manager first.

## Requirements

### AC1 Boards and grids

- Copy the 5 boards into `.hive/spec/tui-v2/design/`. `cmp` against the
  staged copies gives no difference.
- Add each board to `Sizes`. Add each new symbol to decision 13 and the
  symbol list, if a board uses one.
- Run the exporter. Commit the 5 new grids. Write one demo fixture for each
  board.

### AC2 Spec

Add the decisions of this task to the tui-v2 README in ASD-STE100 Strict,
from 91. Keep the meaning of the plan and of this brief. Record the ask
record format, the command, and the prompt text of AC5.

### AC3 The command

- `hive ask --agent NAME --file FILE` reads one ask as JSON. `--file -` reads
  standard input. The command writes one ask record to
  `.hive/state/asks/ID.json` and returns at once. It does not wait for the
  answer. It prints the ask ID. With `--json` it prints the standard result
  object of `AGENTS.md`.
- The JSON has the shape of the native question tool: `questions`, a list of
  1 to 4 items. Each item has `question`, `header` (12 characters at most),
  `options` (2 to 5 items, each with `label` and `description`, and an
  optional `preview`), and an optional `multiSelect`. Reject an ask that
  breaks a limit, with a clear error and no record.
- Hive adds the row `Other: type an answer` after the options of each
  question. The ask JSON does not give it.
- The record keeps the asker, the parent of the asker at the time of the ask,
  the time, the questions, the state (`open`, `answered`, `dismissed`), and
  the answer.
- Each ask record is a durable state. Log each ask, each answer, each
  dismiss, and each delivery as other state transitions.

### AC4 The card and the dock

- Each open ask is a card in a full-width dock between the feed and the
  input. The dock is at most half of the screen height. The dock scrolls
  inside, with the current question in view.
- The card title is `? ASKER asks · HEADERS`. The right side of the title
  shows `open · I of N`, and `J next` when a next card exists. When the asker
  stopped, it shows `asker stopped · answer goes to PARENT`.
- The oldest card has the focus first. `J` and `K` move to the next and the
  previous card.
- The feed shows one line for each ask: `? asked · HEADERS · open below` for
  the card with the focus, and the other forms of the boards for the other
  states. Answers and dismissed cards stay in the chat history, as the
  `ChatAskStopped` board shows.
- The options use the marks `[ ]` and `[x]`. `(REC)` stays a part of the
  label text. The row `›` marks the current option.
- A preview shows in a box on the right of the options, as the
  `ChatAskPreview` board shows. The preview follows the current option.

### AC5 Answer

- The card has the focus in normal mode when an ask is open. The footer starts
  with `[ask]`. `↑` `↓` move. `1` to `5` or a click picks an option. `space`
  marks an option of a multi-select question. `enter` picks and moves to the
  next question. `enter` on the last question sends the answer. `tab` opens a
  `note` line for the question. The footer then starts with `[note]`. `x`
  dismisses the card. While a card is open, the wheel and `pgup` `pgdn`
  scroll the feed. A click on the feed gives the feed the focus. `i` goes to
  input mode as task 126 sets.
- `Other` opens a text line for the free answer.
- Hive sends the answer as one prompt to the pane of the asker, with the path
  that `hive chat` uses for a user message. The text is:

  `⬢ user: answer to ask ID from ASKER · HEADER → LABEL; HEADER → LABEL, LABEL · note: TEXT`

  Give one `HEADER → LABELS` part for each question, in order, and put a
  `· note: TEXT` part after a question that has a `note`. For `Other`, the label
  is the typed text.
- When the asker stopped, hive sends the same text to the parent of the
  asker, with `(asker stopped)` after `from ASKER`.
- On `x`, hive sends `⬢ hive: the user dismissed ask ID from ASKER · HEADERS.
  No answer comes.` to the asker, or to its parent when the asker stopped.
- When a delivery fails (for example, the pane waits for a prompt), the ask
  keeps the answer and the state `answered`, and hive tries the delivery
  again with the rules that hive uses for other prompts. Log each try.

### AC6 Header count

- The header count of task 126 gives the number of open asks, on each tab, at
  both widths. Use the singular `ask` when N is 1.

### AC7 Role skills

- Edit each `.agents/skills/hive-*/SKILL.md`: an agent asks the user with
  `hive ask`, never with the native question tool of its harness. Give the
  command, the JSON shape, and the answer prompt of AC5.
- Do not edit the user's `interrogate` skill or other skills.
- Write the skill text in ASD-STE100 Strict. The counts in
  `testdata/ste-baseline.json` must not rise.

### AC8 Boards that change

The dock and the header count can change boards that exist (for example
`Help` for the new keys). Do the same as in task 126:

- Render the new grid of each board that changes from the code at its board
  size. Write them to `evidence/new-grids/` in the grid format of the
  exporter. Commit and push them.
- Write `evidence/new-grids/README.md`. For each board, give the rows that
  change.
- Send the path to the manager. Do not change the golden files yet. Hold until
  the manager sends the approval. Then do as the manager says, run the
  exporter, and check that each grid equals its grid from `new-grids`. In the
  spec, record which grids the approved grids replace.
- Do the other work while you hold.

### AC9 Tests

- Write each test first and see it fail for the correct reason.
- The golden check passes for each of the 5 new boards and for each board of
  AC8 after the approval.
- Command tests: a valid ask makes one record and returns at once. Each limit
  of AC3 gives an error and no record.
- Model tests with literal expected text and tokens: the card, the title, the
  feed line, the focus order, `J` and `K`, each key of AC5, a click on an
  option, the `note` line, `Other`, a multi-select answer, the dock height cap
  and its scroll, and the header count at 0, 1, and 2 at both widths.
- Delivery tests: the exact prompt text of AC5 for one question, for two
  questions with a `note` and `Other`, for a stopped asker, and for a dismiss.
  Mock only the herdr pane. A failed delivery keeps the answer and tries
  again.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Board copies, grids, fixtures | V1, V2 |
| AC2 | The spec has the new decisions | V-gate |
| AC3 | The command makes a valid record and rejects a bad ask | V3, V6 |
| AC4 to AC6 | The card, the answer, and the header match the boards and the plan | V3, V4, V5, V6 |
| AC7 | The role skills use `hive ask` | V-gate |
| AC8 | The user approved the new grids of the changed boards | V2, V4 |
| AC9 | The tests fail on the base and pass on the task branch | V3 |

## Verification items

### V1 Boards

- Proves: AC1.
- Command: `cmp` of each staged board against its copy. Capture both with
  the Firefox command of the verifier skill. Put the profile folder and the
  captures in `.hive/tmp/` of the worktree. Remove them after the check.
- Expected: no difference. No capture shows `{{`.

### V2 Grids

- Proves: AC1, AC8.
- Command: run the exporter on the design folder. Compare each grid with the
  committed grid, cell by cell, with `internal/lib/cellgrid`.
- Expected: the commit has the 5 new grids and the approved grids. No other
  grid changes.

### V3 Tests red and green

- Proves: AC3 to AC6, AC9.
- Command: run the new tests on the base code, then on the task branch. Run
  each new test with `-count=50` and with `-race`.
- Expected: each fails on the base and passes on the task branch in each run.
  Exception: a regression guard for behavior that the base already has can
  pass on the base. The evidence names each guard, and a deliberate break of
  the guarded behavior makes it fail.

### V4 Live check of the boards

- Proves: AC4 to AC6, AC8.
- Command: the live check of decision 43 for each new board and each board of
  AC8 at its board size, with `COLORTERM=truecolor hive --demo FILE --size
  WxH`. Then in the `ChatAskMulti` demo, press each key of AC5 and send a
  click on an option and on the feed with the raw SGR mouse sequences. Read
  the pane after each key and each event.
- Expected: each capture equals its grid. Each key and each event gives the
  result of AC5.

### V5 Visual alignment

- Proves: AC4 to AC6.
- Command: run the real TUI in a herdr pane at 120×40 and at 80×24 with an
  open card. Compare the capture at the board size with the board grid, cell
  by cell. Look at each capture and check each point:
  - Each panel border and corner lines up. The dock is between the feed and
    the input.
  - The card title, the marks, the `›` row, and the preview box are at the
    board columns.
  - The header count is at the board columns.
- Expected: each capture at the board size equals its grid. Each point
  passes at both sizes. Put each capture, each diff, and a table of the points
  with pass or fail in the evidence. A mismatch is a fail.
- The worker and the verifier each do this item.

### V6 End to end

- Proves: AC3, AC5.
- Command: make a scratch home in `.hive/tmp/` of the worktree, with a
  scratch herdr session of your own. Start the TUI there with
  `HIVE_NO_STARTUP=1`. In a second pane of that session, run a shell that has
  an agent record as asker. Run `hive ask` from it with a two-question ask.
  Answer the card in the TUI. Read the asker pane. Then make a second ask,
  stop the asker record, answer, and read the pane of the parent. Then make a
  third ask and press `x`.
- Expected: `hive ask` returns at once. Each pane gets the exact text of AC5.
  The header count goes up and down. Put each capture in the evidence.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`.
- Never run `hive ask`, open the Chat input, or send chat on the real home.
  Use a demo or a scratch home for each check.
- Close only the panes, the sessions, and the processes that you started. A
  demo quits with two `ctrl+c` presses.
- When a pane is smaller than the board size, use a scratch herdr session of
  your own at a larger size. Stop it after the check.
- Write a removal of a path in a variable as `rm -f "${VAR:?}"/...`, or use a
  literal path.
- Remove each scratch file that you make in `.hive/tmp/` after the check.

## Out of scope

The canvas. The mouse and the chat modes (task 126). The user's `interrogate`
skill.

## Finish

Commit on the task branch with a subject such as `feat(ask): add hive ask and
the question card`. Push the task branch. Write the evidence. Run `hive task
complete`. Send the commit SHA and the V1 to V6 results to the manager with
`hive inbox send --to manager-hive`.
