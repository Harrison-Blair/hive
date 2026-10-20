# Hive TUI feature spec

| Item | Value |
| --- | --- |
| Feature | `hive-tui`. This is the spec task of the feature. |
| Tier | decent |
| Depends on | none. Base: `dev` at `c7cd9a2c` or later |
| Grant | `grants.md 2026-10-20 hive TUI goal` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

Write the feature README `.hive/spec/hive-tui/README.md`. The user approved
the goal. This task changes no code. When a source does not agree with this
brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Sources

Read each source first. All paths are in the home `/home/penguin/source/hive`.

- The goal: `.hive/tmp/coordinator/goal-hive-tui.md`, decisions 1 to 27.
  Decision 22 is withdrawn.
- The plan review: `.hive/tmp/reports/plan-review-hive-tui.md`, items G1 to
  G14.
- The designs: `.hive/tmp/coordinator/designs/hive-tui/*.dc.html`. When a
  design and a decision do not agree, the decision wins.
- The README shape: MAP decision 5 in
  `projects/hive/plans/home-spec-planning/MAP.md`.

## Requirements

### R1 README shape

The README has these sections, in this order:

1. Destination: what the user can do when the feature is complete.
2. Out of scope.
3. Decisions: one line for each goal decision 1 to 27, with its number. Keep
   decision 22 as one line that says it is withdrawn.
4. Manager decisions: one line for each item in R2, with its G number.
5. Waves: the task list in R3, with the done-when point of each wave.
6. Fog of war.
7. Tasks: one link for each task folder below `tasks/`.

### R2 Manager decisions

Write each of these decisions in the README. Do not change their meaning.

- G1: No protection for typing in the coordinator pane (goal decision 23).
- G2: Each prompt into an agent pane (a chat message or a nudge) holds one
  lock file for that pane, `.hive/state/locks/pane-<pane-id>.lock`, with
  `flock`. The lock is outside the state lock. The text and the Enter go in
  one `agent.prompt` call. Hive never splits them. When the delivery is not
  certain, hive marks it `unknown` and does not send it again.
- G3: Hive sends each user chat message to the coordinator pane with the
  marker and a user tag: `⬢ user: TEXT` (goal decision 27). The coordinator
  treats it as the user's own message, with the authority of the user. The
  chat transport task changes the role gate text in `AGENTS.md`, the role
  skills, and the manager template: the marker shows that hive delivered the
  message, and `⬢ user:` shows that the user wrote it in the TUI. Each chat entry has an ID `c<n>`. A user
  message has a delivery state: `pending`, `sent`, or `unknown`.
  `hive chat reply` reads the text from standard input. Only the coordinator
  can reply. A reply can name the ID of the message that it answers. Each
  append to the chat JSONL file holds the home lock. The coordinator skill
  tells the coordinator to send each reply to the user with `hive chat reply`.
  The manager tells a coordinator that is already live once with a hive
  message.
- G4: Each new agent record has `parent` (the key of the agent that spawned
  it) and `replaces` (the key of the generation before it). A verifier has its
  manager as parent. An old record without these fields shows in the graph
  grouped by role and project, with the edge labeled `inferred`.
- G5: Each agent record has `status_changed_at`. Hive sets it only when
  `agent_status` changes. The TUI shows `done` as idle. When the supervisor
  claim is stale, the TUI shows each agent status as `unknown`. Activity age is
  the time since `status_changed_at`. Without that field the age is unknown.
- G6: A completion target does not widen a grant. When the user picks a target
  beyond `dev`, the coordinator writes the named grant at dispatch (goal
  decision 24).
- G7: A goal record is `.hive/state/goals/<id>.json`, with ID `g<n>-<slug>`.
  It has project, title, grant (the grant line), status (`active` or `done`),
  and waves. Each wave has an ID, a title, a done-when point (`local`, `dev`,
  `pr`, `main`, or `release`), task IDs, an optional `needs_user` text, and a
  status. `hive goal create`, `show`, `list`, and `edit` change and show it.
  The TUI shows wave progress from the task records. Only the manager sets a
  wave or goal status. A task that no goal names shows as unassigned.
- G8: `needs_user` is a short text on a task or a wave. The controlling role
  sets it and clears it with `hive task edit` or `hive goal edit`. The Queue
  shows a stop on that task and on each task that depends on it. "Landed"
  means that the task delivery has a commit and `pushed` is true.
- G9: Bare `hive` checks for the herdr environment. Without it, it runs only
  the TUI and shows a notice. With it, each start step checks first and is
  safe to run again: Hive keeps a live coordinator and a fresh supervisor
  claim. Otherwise, Hive starts each in its own tab. When the TUI quits, the
  coordinator and the supervisor keep running.
- G10: Startup never closes a pane that it did not start. When an old `logs`
  tab exists, the coordinator skill tells the coordinator to close it by hand
  after the new supervisor tab is live.
- G11: The config file accepts `home` and a `[coordinator]` table with
  `harness` and `model`. `hive init` changes only the `home` line and keeps
  the other lines. Unknown keys fail with an error that names the key.
- G12: The TUI reads, chats, and attaches to a pane. It has no other action.
  Remove the design controls for held nudges, question cards, interruption,
  edits, restart, and cleanup. Where a design shows such a control, the TUI
  shows the hint "ask in Chat".
- G13: Void (goal decision 22 is withdrawn).
- G14: Each TUI screen loads data in a background command, not in the input
  loop. Each CLI command logs one `request_id`. The Logs screen groups the
  events of one request with it. Health gets new checks: the supervisor
  binary is older than the installed `hive`, and a worktree exists for a task
  that landed. Waves A and B each have one live check of a chat message and a
  nudge to the same pane at the same time.

### R3 Waves

Write this list in the Waves section. The manager fills in the task IDs as
the tasks start.

| Wave | Task | Tier | Depends on |
| --- | --- | --- | --- |
| A, done when landed on `dev` | Goal record and `hive goal` (G7, G8 wave part) | decent | this spec task |
| A | Task `needs_user` (G8 task part) | small | this spec task |
| A | Agent lineage and status time (G4, G5) | decent | this spec task |
| A | Chat transport, `hive chat send/reply/log`, chat history, pane lock, `⬢ user:` tag and role gate text (G2, G3, decision 27) | strong | this spec task |
| B, done when landed on `dev` | TUI shell: sidebar of 8 screens, look, background loading, attach, quit with two ctrl+c (decision 26) | strong | wave A |
| B | Chat screen | strong | shell, chat transport |
| B | Agents screen: tree, graph, table | strong | shell, lineage |
| C, done when landed on `dev` | Goals and Queue screens | decent | wave B |
| C | Tasks and Friction screens | decent | wave B |
| C | Health screen and new checks | decent | wave B |
| C | Logs screen and `request_id` | decent | wave B |
| C | Startup: bare `hive` setup, coordinator config keys, coordinator skill steps (G9-G11) | strong | wave B |

### R4 Fog of war

List the open points: the herdr layout of the three tabs, the graph layout
for many agents, the design files stay untracked in the home, and the noise
levels of the Chat feed.

### R5 Text

Write the README in ASD-STE100 Strict. `ste_test.go` checks the brief copy.
When the brief copy causes a violation, change the words of that sentence in
the copy and keep its meaning. Record each change in the evidence.

## Not in scope

Code. Other features. `main`.

## Verification

| ID | Check | Expected result |
| --- | --- | --- |
| V1 | Compare the README with R1 to R4 | Each section is there in order. Each goal decision 1 to 27 and each G item has one line. |
| V2 | `git diff dev --stat` | Only `.hive/spec/hive-tui/README.md`, the brief copy, the evidence, and `testdata/ste-baseline.json` (only to add or lower a count) changed. |
| V3 | The gate of `AGENTS.md` from the worktree root | `gofmt -l` prints nothing. The other three commands exit 0. |

## Finish

Commit on the task branch with a subject such as `docs(spec): add the hive-tui
feature`. Push the task branch. Write the evidence. Run `hive task complete`.
Send the commit SHA and the V1-V3 results to the manager with
`hive inbox send`.
