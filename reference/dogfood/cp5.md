# CP5 v1 end to end

Run date: 2026-10-18 local, 2026-10-19 UTC.

This file records the CP5 v1 end to end checkpoint.

## Definition of v1

The definition of v1: hive builds hive through a manager, a worker, a verifier,
and a landing.

## Checkpoint and paths

| Item | Value |
| --- | --- |
| Checkpoint | CP5 v1 end to end |
| Artifact | `reference/dogfood/cp5.md` |
| Tracked evidence | `spec/hive-v1/evidence/CP5-run-01/VERIFIED.md` |

## Routes

| Role | Harness and model |
| --- | --- |
| Manager | `pi opencode-go/deepseek-v4.1-flash:medium` |
| Coding worker | `pi opencode-go/deepseek-v4.1-flash:medium` |
| Source verifier | `pi openai-codex/gpt-6.1-sol:high` |

The DeepSeek route requested the model `opencode-go/deepseek-v4.1-flash:medium`, the actual effective model was `opencode-go/deepseek-v4.1-flash:high`, and Pi normalizes DeepSeek `:medium` to the effective `high`.

## Hive commands of the flow

One command per line:

```text
task create
worktree create
agent spawn
task complete
task verify
land --mode merge
worktree remove
handoff write
```

## Runner captures

The runner wrote its capture files to `.hive/tmp/coordinator/reports/cp5-captures/`.
