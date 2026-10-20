# Hive TUI

## Destination

The user can open an eight-screen TUI with bare `hive`. It shows agents and their lineage, goals, task queues, task evidence, friction, health, logs, and chat with the coordinator. The user can attach to an agent pane.

## Out of scope

- Capacity data from qmeter.
- Needs You and Timeline screens.
- TUI actions beyond reading data, chatting, and attaching to a pane.
- Changes that bypass the coordinator.

## Decisions

1. Agents has tree, graph, and table views. The graph groups agents under home and each project with its goal. It shows coordinator-to-manager-to-worker edges and dashed nudge edges. Stopped agents appear dim. `h` hides them.
2. Queue shows task chains under each grant and the stops that need the user.
3. Tasks shows evidence. Friction has its own screen.
4. Chat lets the user talk to the coordinator from the TUI. Hive sends user text to the coordinator pane as `⬢ user: TEXT` and reads the replies. Hive folds events into one line, and a noise level filters the feed.
5. Logs filters by level, component, project, agent, and text. It follows logs, shows one line per command with its duration, shows one event as JSON, and groups events by request. Startup replaces the `hive logs --follow` pane in the herdr `logs` tab.
6. Health checks the home and agents, an old supervisor binary, and stale worktrees. Capacity data from qmeter is out of scope.
7. Sidebar order is Chat, Agents, Goals, Queue, Tasks, Friction, Health, Logs. Needs You and Timeline are not screens.
8. Each wave has a done-when point chosen at dispatch: local implementation, landed on `dev`, PR open, merged to `main`, or release. The wave runs to that point under its grant.
9. Queue shows L2 run order, with one panel for each project and no view cycle.
10. The TUI uses the stenographer B panel look with a honey accent.
11. The coordinator sends each reply with `hive chat reply`. The TUI shows coordinator replies.
12. The user writes in the TUI. The supervisor does not hold nudges. Hive sends each chat message and nudge to the pane as one complete prompt.
13. A goal record and `hive goal` command hold the project, title, done-when point, grant, waves, and tasks.
14. The TUI reads data, chats, and attaches to a pane. Each other change goes through the coordinator.
15. Chat decisions use plain text in v1. Question cards are out of scope.
16. Chat history uses one JSONL file per day under `.hive/state/chat/`. Text typed directly into the coordinator pane is not in chat history.
17. A task has a short `needs_user` field set by its manager. Queue shows it as a stop.
18. The initial `hive tui` plan opened a TUI tab named `hive` and started missing coordinator and supervisor processes. Decision 20 sets bare `hive` as the entry point.
19. Work runs in three waves after H2 wave 1. Wave A adds data. Wave B adds the shell, Chat, and Agents graph. Wave C adds the other screens and startup changes. Each wave ends when it lands on `dev`.
20. Bare `hive` opens the TUI in a tab named `hive`, the coordinator in its own tab, and the supervisor in its own tab. There is no separate `hive tui` command. The old `logs` tab goes away.
21. The config has `coordinator.harness` and `coordinator.model` keys in `~/.config/hive/config.toml`. Defaults are `claude` and `claude-opus-5-5`.
22. Withdrawn.
23. The coordinator pane has no protection for typing. The user can type there on purpose.
24. A completion target does not widen a grant. If the user picks a target beyond `dev`, the coordinator writes the named grant at dispatch. The wave runs to that target without a stop.
25. The plan review is `.hive/tmp/reports/plan-review-hive-tui.md`. The manager decides G2-G5, G7-G12, and G14. G13 is void because decision 22 was withdrawn.
26. The user quits the TUI by pressing Ctrl+C twice. The first press shows `press ctrl+c again to quit`. The `q` key does not quit. The coordinator and supervisor keep running after the TUI quits.
27. Hive sends each user chat message with the marker and tag `⬢ user: TEXT`. The marker shows that Hive delivered it. The tag shows that the user wrote it in the TUI and carries the user's authority. The coordinator treats it as the user's own message.

## Manager decisions

- G1: The coordinator pane has no typing protection, as stated in decision 23.
- G2: Each chat message or nudge prompt takes `.hive/state/locks/pane-<pane-id>.lock` with `flock`. The pane lock is outside the state lock. One `agent.prompt` call sends the text and Enter together. Hive never splits the prompt. If delivery is uncertain, Hive marks it `unknown` and does not send it again.
- G3: Hive sends each user chat message to the coordinator pane as `⬢ user: TEXT`. The marker shows that Hive delivered the message. The tag shows that the user wrote it in the TUI and carries the user's authority. The coordinator treats it as the user's own message. The chat transport task changes the role gate in `AGENTS.md`, the role skills, and the manager template. Each chat entry has ID `c<n>`. User messages have delivery state `pending`, `sent`, or `unknown`. `hive chat reply` reads text from standard input. Only the coordinator can reply, and a reply can name the ID of the message it answers. Each chat JSONL append takes the home lock. The coordinator skill tells the coordinator to send each reply with `hive chat reply`. The manager sends one hive message to a coordinator that is already live.
- G4: Each new agent record has `parent`, the key of the agent that spawned it, and `replaces`, the key of the prior generation. A verifier has its manager as parent. The graph groups old records without these fields by role and project. It labels each edge `inferred`.
- G5: Each agent record has `status_changed_at`. Hive sets it only when `agent_status` changes. The TUI shows `done` as idle. If the supervisor claim is stale, the TUI shows each agent status as `unknown`. Activity age uses `status_changed_at`. Without that field, age is unknown.
- G6: A completion target does not widen a grant. If the user picks a target beyond `dev`, the coordinator writes the named grant at dispatch.
- G7: A goal record is `.hive/state/goals/<id>.json`, with ID `g<n>-<slug>`. It has project, title, grant line, status (`active` or `done`), and waves. Each wave has an ID, title, done-when point (`local`, `dev`, `pr`, `main`, or `release`), task IDs, optional `needs_user` text, and status. `hive goal create`, `show`, `list`, and `edit` change or show the record. The TUI shows wave progress from task records. Only the manager sets wave or goal status. A task that no goal names appears as unassigned.
- G8: A task or wave can have short `needs_user` text. The controlling role sets or clears it with `hive task edit` or `hive goal edit`. Queue shows a stop on that task and each task that depends on it. “Landed” means the task delivery has a commit and `pushed` is true.
- G9: Bare `hive` checks for herdr. Without herdr, it runs only the TUI and shows a notice. With herdr, each start step checks first and can run again safely. Hive keeps a live coordinator and a fresh supervisor claim. Otherwise, it starts each in its own tab. The coordinator and supervisor keep running after the TUI quits.
- G10: Startup never closes a pane that it did not start. If an old `logs` tab exists, the coordinator skill tells the coordinator to close it by hand after the new supervisor tab is live.
- G11: Config accepts `home` and a `[coordinator]` table with `harness` and `model`. `hive init` changes only the `home` line and keeps all other lines. An unknown key causes an error that names the key.
- G12: The TUI reads data, chats, and attaches to a pane. It has no other action. Remove design controls for held nudges, question cards, interruption, edits, restart, and cleanup. Show the hint `ask in Chat` where a design has one of those controls.
- G13: Void because goal decision 22 is withdrawn.
- G14: Each TUI screen loads data in a background command, outside the input loop. Each CLI command logs one `request_id`. Logs groups events by request. Health checks if the supervisor binary is older than the installed `hive` and if a worktree exists for a landed task. Waves A and B each include a live check of a chat message and a nudge sent to the same pane at the same time.

## Waves

The manager assigns task IDs when each task starts.

| Wave | Task | Tier | Depends on |
| --- | --- | --- | --- |
| A, done when landed on `dev` | Goal record and `hive goal` (G7, G8 wave part) | decent | this spec task |
| A | Task `needs_user` (G8 task part) | small | this spec task |
| A | Agent lineage and status time (G4, G5) | decent | this spec task |
| A | Chat transport, `hive chat send/reply/log`, chat history, pane lock, `⬢ user:` tag and role gate text (G2, G3, decision 27) | strong | this spec task |
| B, done when landed on `dev` | TUI shell: sidebar of 8 screens, look, background loading, attach, quit with two Ctrl+C presses (decision 26) | strong | wave A |
| B | Chat screen | strong | shell, chat transport |
| B | Agents screen: tree, graph, table | strong | shell, lineage |
| C, done when landed on `dev` | Goals and Queue screens | decent | wave B |
| C | Tasks and Friction screens | decent | wave B |
| C | Health screen and new checks | decent | wave B |
| C | Logs screen and `request_id` | decent | wave B |
| C | Startup: bare `hive` setup, coordinator config keys, coordinator skill steps (G9-G11) | strong | wave B |

## Fog of war

- The herdr layout of the three tabs.
- The graph layout for many agents.
- The design files stay untracked in the home.
- The noise levels for the Chat feed.

## Tasks

- [73-hive-tui-spec](tasks/73-hive-tui-spec/)
