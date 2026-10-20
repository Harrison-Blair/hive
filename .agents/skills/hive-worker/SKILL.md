---
name: hive-worker
description: Do one hive task in a worktree. Use when a manager gives you one task and a brief. Work test-first, commit and push the task branch, write the evidence, run the gate, and finish the task.
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
`.hive/spec/<feature>/tasks/<id>/brief.md`. Commit this file with your work.

When hive worktree create shows allowlist_needed, add !.hive/spec/ and
!.hive/spec/** to the project .gitignore after the rules that ignore the path.
Commit this change with the feature README.

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
go vet ./...
golangci-lint run ./...
go test -race ./...
```

The first command prints nothing. Each other command passes.

## Evidence

Write one `V<n>.md` file for each verification item of the brief. Write the
red run and the green run for a test item. Write the run for a live or gate
item.

For a task with a feature, write each `V<n>.md` file to
`.hive/spec/<feature>/tasks/<id>/evidence/`. For a task without a feature,
write each file to the evidence directory that the brief names. The verifier
writes `VERIFIED.md` in the same directory.

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

## Rules

- Never edit the plan tracker.
- Never commit to `dev` or `main`.
- Never weaken, skip, or delete a test.
