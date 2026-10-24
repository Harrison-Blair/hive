# hive guide

This guide gives the flows of hive work and a reference of each command.
Read the brief of your task before you start.

Hive marks each message that it delivers with `⬢`. A message that starts with
`⬢ user:` comes from the user. Each other marked message comes from the
approved goals and grants. A message with no marker comes from the user.

## Task

1. Read the project instructions and each unread inbox record.
2. Read the brief with `hive task show ID --brief`.
3. Create a task with `hive task create`, and add it to its wave with
   `hive goal edit`.
4. Make the worktree with `hive worktree create --task ID`.
5. Spawn one agent for the task with `hive agent spawn`.
6. Do the work in the worktree. Write a failing test first.
7. Run the gate of `AGENTS.md`. Commit and push the task branch.
8. Run `hive task complete --task ID --agent NAME` when each verification
   item has a green run.

A completed task record starts an independent check. The agent that checks
the task reads the same brief and repeats each verification item.

## Verify

1. Read the brief and each verification item.
2. Run each item again on the task branch. Use a fresh checkout and a scratch
   home.
3. Write `VERIFIED.md` in the evidence folder. Do not commit it.
4. Run `hive task verify --task ID --agent NAME` for a pass.
5. Run `hive task verify --task ID --fail --reason TEXT` for a fail.
6. Send each fail to the agent that owns the project with `hive inbox send`.

## Land

1. Land a verified task with `hive land`. Pass the mode and the grant.
2. The mode `merge` merges the task branch into the base branch.
3. The mode `pr` opens a pull request. Pass the evidence URL.
4. After the merge, `hive land` copies each changed spec file into the home.
5. Remove the worktree with `hive worktree remove --task ID`.
6. A merge conflict stops the land. Reopen the task with `hive task reopen`.
   Spawn a repair agent, then a fresh independent agent.

## Inbox

1. Each agent has one inbox key. Hive gives the key to the agent.
2. `hive inbox send` appends one record to the inbox of a key.
3. `hive inbox` reads each unread record of the caller.
4. `hive inbox --all` prints each record and sets nothing.
5. A task command sends one record to the project inbox after a completed
   task and after a check.
6. Read each unread record at startup and after each nudge. Act on each
   record.
7. End the turn when no approved work is ready. Do not wait in a loop.

## Ask

1. Write the ask as JSON in a file, or give the JSON on standard input.
2. Run `hive ask` in your own pane. It writes one ask record and prints the
   ask ID.
3. The command stops at once. It does not wait for the answer.
4. Do other work, or stop and wait.
5. The answer comes to your pane as one prompt. Each chosen label appears in
   the prompt. The `Other` row gives the text that the user typed.
6. When the user dismisses the ask, a message says so and no answer comes.

## Chat

1. The user types a message in the Chat screen of the TUI.
2. The coordinator reads the message in its pane. It answers with
   `hive chat reply --to ID`.
3. `hive chat log` prints the chat history.
4. An agent does not send a user message. Use `hive agent report` for a
   result to your parent, or an inbox record for a peer.
5. `hive ask` is the way to ask the user a question.
