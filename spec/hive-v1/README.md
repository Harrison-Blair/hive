# Hive v1 specification

| Item | Value |
| --- | --- |
| Status | Design authority for hive v1 |
| Record date | 2026-10-17 |
| Source baseline | `c317d456` on `dev` |
| Level | Goals, features, architecture, risks, and the task plan |
| Terms | [STE-GLOSSARY.md](../../STE-GLOSSARY.md) gives one meaning for each term. |

Hive is a Go CLI and TUI that coordinates coding agents in herdr from one home.
[AGENTS.md](../../AGENTS.md) gives the procedures for agents that work in this
repository. This directory gives the design and the plan. When the two disagree,
correct this directory first, then the procedure.

## Files

| File | Content |
| --- | --- |
| [01-goals.md](01-goals.md) | What hive is for, and when v1 is complete |
| [02-features.md](02-features.md) | What v1 includes, what comes later, and what is out |
| [03-architecture.md](03-architecture.md) | Home, roles, herdr layout, agent lifecycle, supervision, state, worktrees, memory, delivery, presets, logging, TUI, bounds, recovery |
| [04-risks.md](04-risks.md) | Each recorded risk and its answer |
| [05-tracker.md](05-tracker.md) | The task graph, the waves, the checkpoints, the status of each task, and the procedures |
| [tasks/](tasks/README.md) | One brief for each task, in the format that `tasks/README.md` gives |
| [evidence/](evidence/README.md) | The evidence files of each task, in the format that `evidence/README.md` gives |

## History

Fledge was the earlier tool in this repository. Hive is its reauthoring, with a
clean start on 2026-10-17. The v1 planning record of 2026-10-14 is on the branch
`plan/hive-v1`. A partial implementation is on `integration/hive-v1`. Both
branches are history. This directory does not import them and does not reuse
their decision IDs.

## Changes to this specification

A user decision changes the design. Record the change in the file that owns it,
with the date. Keep the four design files at the level of goals, features,
architecture, and risks. Tasks, waves, and status live in `05-tracker.md` and
`tasks/`. A checkpoint critique that changes the design goes into the design
file first, then into the tasks it affects.
