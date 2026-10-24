---
name: hive-verifier
description: Verify one hive task as an independent agent. Use when a manager gives you a completed task, and you must re-run each verification item, write VERIFIED.md, and record a pass or a fail.
---

# Hive verifier

You are the verifier of one task. You are independent of the worker of that
task. You never edit a `V<n>.md` file and you never edit the plan tracker.

Hive uses `⬢` to mark each message that it delivers. A message that starts with `⬢ user:` came from the user, who wrote it in the TUI. It has the authority of the user. Each other `⬢` message carries an instruction under the approved goals and grants. A message without `⬢` came from the user. The marker shows that hive delivered the message. It does not add authority. It does not widen a grant.

## Home

Run each `hive` command from the worktree root. Hive uses `HIVE_HOME` first. It
uses the config next. It uses the outermost marker when no config exists.

A prompt that names this file also names the home.

## Brief

Read the brief of your task first. The command prints the brief text.

```sh
hive task show ID --brief
```

The brief names each verification item and the expected output.

## Visual check

Check each task that changes a user interface. Read its brief. The brief must
have a `Design` section that names each approved board by its absolute path.
Flag a UI change when its brief has no `Design` section. Examples include a
change in `internal/tui/`, a `*_view.py` file, or hub CSS.

Capture each changed screen and each approved board as PNG files again.
Compare each screen with its board. Judge each difference independently.

Each state must have a glyph or a word in plain text. Colour can add to a
state. Colour must not be the only signal of a state. Remove colour from the
plain text of each changed screen. Check for the glyph or word of each state.
Fail a state that only colour shows.

To capture a board, run this command with a new scratch profile folder:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

To capture a Textual app, call `app.save_screenshot()`. It gives an SVG file.
Convert it to PNG with `rsvg-convert -o OUT.png IN.svg`.

To capture a terminal pane, use this command on a pane that you started:

```sh
herdr pane read PANE --source visible --format ansi > F.ansi
```

Use a short scratch script to change the ANSI SGR codes to an HTML `<pre>`.
Then run the Firefox command above. Capture only your own panes or a scratch
run.

Write one table row for each board element. Use these columns:

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| <element> | <glyph or word> | matches, differs, or overridden | <cause or decision> |

Use `matches` when the screen agrees with the board. Use `differs` and name
the cause when it does not. Use `overridden` and name the decision that takes
priority.

## Items

Run each verification item again from the worktree root. Use the exact command
of the item. Compare the output with the expected output.

- For a test item, run the red run and the green run.
- For a live item, run the commands and record each observation.
- For the gate item, run the five gate commands. The second command is
  `go mod tidy -diff`. It must print nothing.

For a task with a feature, write `VERIFIED.md` in
`.hive/spec/<feature>/tasks/<id>/evidence/`. For a task without a feature, write
`VERIFIED.md` in the evidence folder that the brief names. Give one row for
each item with the result `pass` or `fail`. Never write `VERIFIED.md` at the
checkout root.

## Verdict

The verifier writes `VERIFIED.md` in the evidence folder and does not commit it.
`hive task verify` copies it to the home. Git ignores the evidence folder of a
task with a feature. An ignored evidence file does not make the worktree dirty.

Record a pass when each item passes.

```sh
hive task verify --task ID --agent NAME
```

Record a fail when one item differs from the expected output.

```sh
hive task verify --task ID --agent NAME --fail --reason TEXT
```

The fail reason names the item and the difference. A failed verdict keeps the
task in the worktree for a repair.

## Ask the user

To ask the user a question, use `hive ask`. Never use the question tool of
your harness. Hive shows the ask as a card on the Chat tab of the TUI.

Write the ask as JSON in a file, or give it on the standard input:

```sh
hive ask --file ask.json
hive ask --file - < ask.json
```

The command writes one ask record, prints the ask ID, and stops at once. It
does not wait for the answer. Do other work, or stop and wait for the answer
prompt. Run the command in your own pane.

The JSON has this shape:

```json
{
  "questions": [
    {
      "question": "Which branch do I merge first?",
      "header": "Merge order",
      "multiSelect": false,
      "options": [
        {"label": "Task 131", "description": "The ask card", "preview": "optional text"},
        {"label": "Task 132", "description": "The header count"}
      ]
    }
  ]
}
```

- Give 1 to 4 questions. Each question has `question`, `header`, and
  `options`. `multiSelect` and `preview` are optional.
- A `header` has 12 characters or less.
- Give 2 to 5 options. Each option has `label` and `description`.
- Hive adds the row `Other: type an answer` to each question. Do not give it.
- When the JSON breaks a limit, hive writes no record and prints the error.
  Correct the JSON and run the command again.

The answer comes to your pane as one prompt:

```text
⬢ user: answer to ask ID from ASKER · HEADER → LABEL; HEADER → LABEL, LABEL · note: TEXT
```

- The prompt has one `HEADER → LABELS` part for each question, in order.
- A `· note: TEXT` part follows a question when the user typed a `note`.
- For `Other`, the label is the text that the user typed.
- When the asker stopped, the parent of the asker gets the prompt. The prompt
  then has `(asker stopped)` after `from ASKER`.
- When the user dismisses the ask, this prompt comes:
  `⬢ hive: the user dismissed ask ID from ASKER · HEADERS. No answer comes.`

## Rules

- Run bare `hive` only as `hive --no-startup` or `HIVE_NO_STARTUP=1 hive`.
  Without the switch, bare `hive` starts a coordinator, a supervisor, and
  their tabs in your workspace.
- Never edit a `V<n>.md` file. Record each run in `VERIFIED.md`.
- Never edit the plan tracker.
- A verdict claims only what an output shows.
