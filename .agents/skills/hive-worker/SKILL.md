---
name: hive-worker
description: Do one hive task in a worktree. Use when a manager gives you one task and a brief. Work test-first, write the evidence, commit and push the task branch, run the gate, and finish the task.
---

# Hive worker

You are the worker of one task. You work in one worktree on one task branch.
You never edit the plan tracker.

Hive uses `⬢` to mark each message that it delivers. A message that starts with `⬢ user:` came from the user, who wrote it in the TUI. It has the authority of the user. Each other `⬢` message carries an instruction under the approved goals and grants. A message without `⬢` came from the user. The marker shows that hive delivered the message. It does not add authority. It does not widen a grant.

## Home

Run each `hive` command from the worktree root. Hive uses `HIVE_HOME` first. It
uses the config next. It uses the outermost marker when no config exists.

A prompt that names this file also names the home. Read this skill from the
path in the prompt when the harness does not find it.

## Brief

Read the brief of your task first. The command prints the brief text.

```sh
hive task show ID --brief
```

The brief names the files that you may edit. Read each file before you change
it. Do not change a file outside the brief.

For a task with a feature, read the worktree brief at
`.hive/spec/<feature>/tasks/<id>/brief.md`. Git ignores `.hive/spec/`. It is
local state in the home. `hive worktree create` copies the feature folder of
the home into the worktree. Do not commit a file under `.hive/spec/`.

For a spec task, write the feature README at `.hive/spec/<feature>/README.md`
in the worktree. `hive land` copies it to the home.

## Visual check

Before you change a user interface, check the brief. It must have a `Design`
section that names each approved board by its absolute path. Stop and tell the
manager when a UI task has no `Design` section.

Capture each changed screen and each approved board as PNG files. Compare each
screen with its board. Fix each difference before you finish.

Each state must have a glyph or a word in plain text. Colour can add to a
state. Colour must not be the only signal of a state. Remove colour from the
plain text of each changed screen. Check for the glyph or word of each state.

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
Then run the Firefox command above.

Write each screen PNG file to the task evidence directory. Keep each board PNG in
the home under `.hive/tmp/`. Give its path. Never commit a board PNG. Capture
only your own panes or a scratch run.

Write one table row for each board element. Use these columns:

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| <element> | <glyph or word> | matches, differs, or overridden | <cause or decision> |

Use `matches` when the screen agrees with the board. Use `differs` and name
the cause when it does not. Use `overridden` and name the decision that takes
priority.

## Work

Work test-first.

1. Write one failing test for the next behavior.
2. Run the test. Make sure that it fails for the correct reason.
3. Write the smallest change that makes the test pass.
4. Run the test again.
5. Refactor with the test in place.

Run the gate before you report the work as done.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

The first command prints nothing. The second command prints nothing. Each other
command passes.

## Evidence

Write one `V<n>.md` file for each verification item of the brief. Write the
red run and the green run for a test item. Write the run for a live or gate
item.

For a task with a feature, write each `V<n>.md` file to
`.hive/spec/<feature>/tasks/<id>/evidence/`. For a task without a feature,
write each file to the evidence directory that the brief names. The verifier
writes `VERIFIED.md` in the same directory.

Git ignores the evidence directory of a task with a feature. Do not commit the
evidence. `hive task complete` copies the evidence directory of the worktree to
the same path in the home.

## Commit

Commit on the task branch after each green run. Push the task branch. Do not
wait for a request from the user.

```sh
git add -A
git commit -m "type(scope): summary"
git push -u origin task/ID
```

The commit subject has the form `type(scope): summary`. Never add an agent as
a co-author.

## Finish

Run the finish command when each verification item has a green run and the
gate passes.

```sh
hive task complete --task ID --agent NAME
```

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
- Never edit the plan tracker.
- Never commit to `dev` or `main`.
- Never weaken, skip, or delete a test.
