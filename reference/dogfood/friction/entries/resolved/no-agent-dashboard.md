---
title: "No dashboard shows the agents, their lineage, their activity, and their tasks"
status: "Resolved 2026-10-23"
opened: 2026-10-19
decision: "Goal g1-hive-tui lands the dashboard: the Agents, Chat, Work, Friction, Health, and Logs screens with lineage, activity, and tasks"
---

**Summary:**
On 2026-10-19, the user asked the coordinator for a list of the work of each
active agent. The user cannot see this state easily. To answer, the
coordinator ran `hive agent list` and `hive task list` and joined the output
by hand. The data has these problems:

- `hive agent list` shows the name, the role, the project, the task, and the
  pane. It does not show the parent agent (lineage), if the agent is working,
  idle, or blocked, or the time since its last activity.
- `hive task list` shows `verified` for a landed task. The user cannot tell a
  landed task from a task that waits for a landing.
- The goal of a manager is only in the coordinator notes and in the inbox.
  No hive view shows it.
- Bare `hive` opens a TUI with only a header and the log view
  (`internal/tui/doc.go`).

The user wants a dashboard in the bare `hive` TUI, as btop and htop show
processes. It shows:

1. Each agent as a tree: the coordinator, then each manager, then the workers
   and the verifiers of that manager.
2. For each agent: the harness, the model, the state (working, idle, blocked,
   exited), the time since the last activity, and the pane.
3. For each manager: its project, its goal, and its tasks with their status
   and dependencies.
4. For each worker and verifier: its task, the task status, and the branch.

**Reproduction steps:**
1. Spawn a manager. The manager spawns 2 workers.
2. Run bare `hive`. Only the header and the log view show.
3. Run `hive agent list`. No parent, no activity state, and no goal show.
4. Run `hive task list`. A landed task shows `verified`.
