---
title: "Worktree create opens a second workspace for the project checkout"
status: "Open"
opened: 2026-10-28
decision: ""
---

**Issue:** `hive worktree create` opened a herdr workspace with the label
`stenographer` for the registered checkout. The project already had the
workspace `hive:stenographer` of its manager. The herdr sidebar then showed
two stenographer workspaces, and the worktree workspace grouped below the
new one, not below `hive:stenographer`.

**Status:** Open.

**Summary:**
On 2026-10-28 at 22:35:22 UTC, `hive worktree create --task
205-changelog-v0170` logged the effect `action=opened resource=workspace
path=w14P`. Workspace `w14P` has label `stenographer`, worktree metadata
with `checkout_path` `/home/penguin/source/stenographer`, and one idle shell
pane. The manager workspace `w14N` (`hive:stenographer`, cwd
`projects/stenographer`) has no worktree metadata, so herdr does not group
the task worktree `w14Q` under it. The user saw the two workspaces as an
error. It is not yet checked whether `hive worktree remove` closes `w14P`.

**Reproduction steps:**
1. Spawn the owner planner of a registered project. Observe the workspace
   `hive:<project>`.
2. Run `hive worktree create --task ID` for a task of that project.
3. Run `herdr workspace list`. Observe a new workspace with the label
   `<project>` and the checkout path of the registered repository, beside
   `hive:<project>`.
