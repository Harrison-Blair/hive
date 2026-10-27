---
title: "Task dependencies cannot be changed through Hive"
status: "Open"
opened: 2026-10-25
decision: "Ask a22: use supported replacement tasks and retain cancelled originals. No new CLI feature."
---

**Issue:** Hive has no supported command to change an existing task dependency.

**Status:** Open.

**Summary:**
The user cancelled task 166 in ask a22. Created task 167 depended on it.
`hive task edit` accepts a brief, a feature, or a user-decision field.
It has no dependency option. The ready list requires a verified dependency,
so a cancelled dependency keeps the child out of the ready list.

The manager replaced tasks 167, 168, and 169 with tasks 180, 181, and 182
through `hive task create`. The manager preserved the original briefs and
evidence, cancelled the originals with explicit replacement links, and added
the new tasks to their original goal waves. This uses the supported commands
and gives no cancelled task a fake verification or delivery record.

**Reproduction steps:**
1. In a scratch home, create a task and a child task with `--depends-on`.
2. Cancel the first task.
3. Run `hive task list --ready`. The child does not appear.
4. Run `hive task edit --help`. No flag can change the dependency.

**Remaining gap:**
Replacement tasks use new IDs. Existing dependents can require further
replacement records. Goal waves retain cancelled originals as history and
count them in their totals. The manager must check actual completion against
the accepted replacement graph.

**Occurrence 2026-10-27 (stenographer, goal g3-hub-redesign):** spec review
added two order edges after the tasks existed: 194 after 191 and 192, and
198 after 197. Replacement tasks would cascade to nine downstream tasks
(195, 197 to 203 depend on 194 or 198 by ID). The manager kept the original
tasks, stated each extra edge in its brief and in spec D17, and holds the
order when spawning. `hive task list --ready` therefore shows 194 and 198
ready earlier than the spec allows.
