---
title: "A new task can go only in a feature that has a README"
status: "Open"
opened: 2026-10-19
decision: ""
---

**Summary:**
After the spec migration (task 48), only `hive-v1` has
`.hive/spec/<feature>/README.md`. `hive task create` without `--spec` needs
that README on the base branch. On 2026-10-19 the manager made task
70-spec-allowlist, a fix to the H2 spec layout. Its feature is
`home-spec-planning`, but that feature has no README. The manager put the task
in `hive-v1` to start it. Until each feature has a README from its spec task,
a manager must put new tasks in `hive-v1` or make a spec task first.

**Reproduction steps:**
1. Run `hive task create --project hive --feature home-spec-planning --title T --slug s --brief FILE`.
2. It fails: the README is missing on the base branch.
3. Run the same command with `--feature hive-v1`. It succeeds.
