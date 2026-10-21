# CP5 record

This record follows goal `g1-tui-restart` from the manager to three landings on `dev`. [C01]

The goal belongs to project `hive`. It has waves A, B, and C. Each wave and the goal has status `done`. [C01]

This task made no new end-to-end run. The home has four agent records for this task: three workers and one verifier. Its commits change only the CP5 artifact, tracker, brief, citation and evidence files, and the STE baseline. [C15]

## Agents

| Role | Agent | Harness | Model | Citation |
| --- | --- | --- | --- | --- |
| Manager | `manager-hive` (`manager-hive-g2`, generation 2) | `claude` | `claude-opus-5-5` | C02 |
| Worker | `t2-worker-1` | `pi` | `openai-codex/gpt-6-luna:xhigh` | C03 |
| Verifier | `t2-verifier-1` | `pi` | `openai-codex/gpt-6-luna:xhigh` | C04 |
| Worker | `t2-worker-2` | `pi` | `openai-codex/gpt-6-luna:xhigh` | C05 |
| Verifier | `t2-verifier-2` | `pi` | `openai-codex/gpt-6-luna:xhigh` | C06 |
| Worker | `t3-worker-1` | `pi` | `openai-codex/gpt-6-luna:xhigh` | C07 |
| Verifier | `t3-verifier-1` | `pi` | `openai-codex/gpt-6-luna:xhigh` | C08 |
| Worker | `t4-worker-1` | `pi` | `openai-codex/gpt-6-luna:xhigh` | C09 |
| Verifier | `t4-verifier-1` | `pi` | `openai-codex/gpt-6-luna:xhigh` | C10 |

## Waves and task flow

- Wave A: `2-stale-coordinator-startup`, titled "Stale coordinator check at startup". [C01, C11]
- Wave B: `3-tui-restart`, titled "TUI restart of coordinator and supervisor". [C01, C12]
- Wave C: `4-agents-equal-columns`, titled "Equal column width in the Agents graph". [C01, C13]

Task `2-stale-coordinator-startup` first used worker `t2-worker-1`. Verifier `t2-verifier-1` reported a fail on V4 and the visual check. Hive changed the task from `completed` to `assigned`. [C03, C04, C11]

The manager then started a second worker, `t2-worker-2`. Verifier `t2-verifier-2` passed the task. Hive changed its status to `verified`. [C05, C06, C11]

Merge commit `0e079272924127b69f78b1142929eaa02e7229cf` landed task 2 on `dev`. [C12]

Task `3-tui-restart` used worker `t3-worker-1` and verifier `t3-verifier-1`. The verifier passed the task. Merge commit `3ccd93fbf0a4a213b9ff790321ea72ca65d6e04e` landed the task on `dev`. [C07, C08, C13]

Task `4-agents-equal-columns` used worker `t4-worker-1` and verifier `t4-verifier-1`. The verifier passed the task. Merge commit `ede94f3ea4b8a7fd63d99d23eae5ea74613e5253` landed the task on `dev`. [C09, C10, C14]
