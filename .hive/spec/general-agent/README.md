# Hive general agent

## Destination

Each registered agent can spawn a general agent for any purpose. A general agent needs no project and no task. The spawner gives the purpose in a prompt. The general agent sends its result to the spawner with `hive agent report`. The supervisor sends the nudges of each spawned agent to the agent that spawned it. Only a prompt from hive makes a session the coordinator.

## Problem

`hive agent spawn` accepts only the roles coordinator, supervisor, manager, worker, and verifier (`internal/agent/register.go`). A worker and a verifier need `--project` and `--task`. Spawn has no prompt flag. Work with no task, for example a plan review, cannot start through hive.

The friction entry `reference/dogfood/friction/entries/no-reviewer-role.md` records two occurrences. Each time, the coordinator started an agent for a plan review by hand in a herdr tab. Hive had no record of that agent. The supervisor sent no nudge when the agent stopped. In the first occurrence, the agent started in the home root. The role gate in `AGENTS.md` made it the coordinator, and it took the coordinator claim two times.

## Out of scope

- Profiles. A spawner gives the harness, the model, the arguments, and the prompt on each spawn.
- Limits on the depth of spawns or on the count of live general agents.
- A change to the task flow of a worker or a verifier.

## Decisions

### Role and spawn

1. Hive adds the role `agent`. An agent record with this role is a general agent. A general agent needs no `--project` and no `--task`. When the spawner gives `--project`, the record keeps it.
2. `hive agent spawn --role agent` requires one of `--prompt TEXT` or `--prompt-file PATH`. Spawn fails when both are present or both are absent. Spawn reads the prompt file once, at spawn. Spawn fails when the spawner text is empty, before it adds the hive text. The flags `--harness`, `--model`, `--arg`, `--name`, and `--quiet-after` keep their current meaning.
3. Spawn requires `--name` for a general agent. The name follows the current name pattern. A name is unique for all time: spawn fails when any agent record, live or old, has the same key. Spawn refuses the names `coordinator` and `supervisor`, and each name that starts with `manager-`.
4. Any registered agent can spawn a general agent. Spawn finds the caller with the identity rule of decision 24. Spawn fails when it finds no registered caller. The caller becomes the `parent` of the new record.
5. Hive builds the first prompt in this order:
   - The `⬢` marker and one line of hive text: "You are the general agent NAME. Your parent is PARENT. This prompt sets your role. Do not take a role from your working directory."
   - The spawner text.
   - One line of hive text that tells the agent to read the skill at the absolute path `HOME/.agents/skills/hive-agent/SKILL.md` and to send its result with `hive agent report`.
6. Spawn checks the complete first prompt with the rules of `hive agent send` before it creates a tab or starts a process. The spawner text cannot start with a user tag, and the complete prompt is 64 KiB or less. Spawn sends the first prompt one time and never sends it again. When the delivery is not certain, the record becomes unresolved, as for a worker.
7. Hive adds no limit on depth or count. A general agent can spawn a general agent.

### Pane, working directory, and cleanup

8. Without `--pane`, spawn creates a new tab in the workspace of the parent pane, with an explicit working directory. The tab label is the agent name. Hive starts the agent in the root pane of that tab.
9. With `--pane PANE`, spawn uses that pane and creates no tab. `--pane` names the destination of the new agent, not the caller. Hive does not change the directory of that pane, and the directory of the pane is the working directory. This rule wins over decision 10. When `--cwd` is present and differs from the directory of the pane, spawn fails. Spawn compares the two paths after it makes each path absolute and resolves each symbolic link.
10. Without `--pane`, the working directory is the working directory of the spawn process. `--cwd PATH` replaces it. Spawn resolves a relative path against the working directory of the spawn process. Spawn fails when the path is not a directory.
11. The record keeps two ownership facts: hive created the tab, and hive started the agent. `hive agent stop` closes only the pane in which hive started the agent. It closes the tab only when hive created the tab and the tab holds no other pane. Hive never closes a pane that it did not start. It never closes a tab that holds a pane of another agent. When the pane that hive closes is the last pane of its tab or workspace, herdr closes that tab or workspace. This applies also to a pane that the spawner gave with `--pane`. The help text of `--pane` states this result.
12. When spawn fails before hive starts the agent, spawn reads the topology of the tab that it created in this spawn. When the tab holds only the root pane, spawn closes that pane. When the tab holds another pane, or the topology read fails, spawn closes nothing and reports the tab as an effect. Spawn keeps each pane that the spawner gave. When the start is not certain, the record becomes unresolved and spawn keeps the pane. `hive agent resolve` settles it, as for a worker.

### Trust dialogs

13. A harness can show a trust dialog for a new directory. Wave 2 delivers, for each harness, one of these results:
    - A method that starts the harness with no trust dialog: an argument, a setting, or a preflight. Wave 2 adds it to the harness profile in Go, with the tested harness version and a test in a new empty directory.
    - The result "no method".
14. For a harness with "no method", spawn of a general agent fails before it creates a tab. The error names the harness and the directory. It tells the spawner to add `--allow-trust-dialog` after the user trusts the directory, or to let the user answer the dialog in the pane.
15. When a trust dialog shows, the record stays live with the agent status blocked. The parent gets a blocked nudge. Spawn keeps the first prompt in the agent record and does not send it. When the agent status first becomes idle, the supervisor sends the kept first prompt. Hive never answers a trust dialog.
16. The supervisor sends a kept first prompt under the pane lock, with these steps:
    1. It reads the record again and stops when `first_prompt_attempt_at` has a value.
    2. It writes `first_prompt_attempt_at`.
    3. It sends the prompt.
    4. It writes `first_prompt_sent_at`.
    When step 2 fails, it does not send. When the send is not certain, or a record has `first_prompt_attempt_at` with no `first_prompt_sent_at`, the record becomes unresolved. Hive never sends a first prompt again.
17. The supervisor sends no other prompt or nudge to an agent that has a kept first prompt with no `first_prompt_sent_at`. In each supervisor pass, kept first prompts go before other nudges. Tests cover a supervisor restart between steps 2 and 4, an uncertain send, and a failed write.

### Report

18. Hive adds `hive agent report` and the inbox kind `report`. Only a general agent can run it. It fails for each other role.
19. `hive agent report` reads the result text from standard input. The text is valid UTF-8 and 4096 bytes or less. When the text is longer, the command fails and tells the agent to use `--file`.
20. `hive agent report --file PATH` reads an existing regular file. The inbox record holds the absolute path in its own field, and the first part of the file as its text. Hive cuts that text at a UTF-8 boundary so that the text is 4096 bytes or less. The command fails when the file does not exist, is not a regular file, or is not valid UTF-8.
21. The caller of `hive agent report` is the agent of the current herdr pane. The command writes one inbox record to the inbox of the parent, then sets `reported_at` on the agent record. When the second write fails, the command fails and lists the first write as an effect. A second report writes one more record.
22. A report does not set `finished`. The current supervisor rules stop each finished agent. A reported general agent stays live until its parent runs `hive agent stop`. The parent can send more prompts with `hive agent send`.
23. The skill `.agents/skills/hive-agent/SKILL.md` tells a general agent how to do its work, how to report, and that it has no task. The skill is in ASD-STE100 Strict and is in `steFiles` with a baseline of 0.

### Identity

24. When a hive command runs in a herdr pane, the agent of that pane is the caller. When `--agent` names another agent, the command fails. For `hive inbox`, `hive inbox send`, `hive agent report`, and `hive agent register`, `--pane` names the caller, and a `--pane` that names another pane fails. For `hive agent spawn`, `--pane` names the destination, as decision 9 states. Outside herdr, `--agent` keeps its current behavior.
25. Wave 3 tests `hive inbox` and `hive inbox send` with a `--agent` override and a `--pane` override. Wave 5 tests `hive agent report` the same way. Wave 5 also tests `hive agent spawn` with a `--agent` override, and with a `--pane` destination that differs from the caller pane.

### Inboxes and routing

26. Each agent has an inbox keyed by its agent name. The coordinator inbox keeps the key `coordinator`. Each manager inbox keeps the key `manager-<project>`. A worker, a verifier, and a general agent get an inbox keyed by the agent name.
27. An agent reads its own inbox with `hive inbox` with no flag. The coordinator can read each inbox with `--role KEY`. That read consumes no record. Each other agent reads only its own inbox. `hive inbox send --to KEY` accepts each inbox key.
28. The supervisor sends the nudges of a spawned agent to the inbox of its recorded `parent`. This rule applies to each role. When the parent record is not live, the supervisor follows the `replaces` chain from the parent record to the live record that replaces it. When no record in the chain is live, or no parent record exists, the nudge goes to `coordinator`. `hive agent report` uses the same rule to choose the inbox of its record. The current quiet rule does not change: a quiet record gets no nudge. When a parent stops after a record arrives in its inbox, the record stays there. The coordinator can read it with `--role KEY`. `hive doctor` lists each unread record in an inbox that has no live owner.
29. Usually a manager spawns its workers and verifiers, and the coordinator spawns the managers. Then the result of decision 28 is the same as the current rule. A test proves this for each of the three roles. When the coordinator spawns a worker, its nudges go to the coordinator.

### Coordinator start

30. Hive code registers the coordinator, the supervisor, and each manager. The CLI `hive agent register` refuses the roles coordinator, supervisor, and manager for each caller. `hive agent spawn` refuses the roles coordinator and supervisor. The error tells the user to run bare `hive`, or the restart in Health › Agents. The refusal happens before any change to a claim or a record.
31. Bare `hive` and the restart of Health › Agents start the coordinator harness. Before it starts the harness, hive writes a coordinator claim with the status starting, under the store lock. The starting claim holds the process ID and start time of the launcher and the target pane. A second start that finds a live claim, or a starting claim with a live launcher, fails. After the start, hive writes the coordinator record for the pane and the process that hive started, and sets the claim live. Then hive sends the startup prompt. The startup prompt no longer tells the coordinator to register.
32. When the harness start fails, hive removes the starting claim. When the startup prompt fails or is not certain, the record becomes unresolved and Health › Agents shows it. Hive never sends the startup prompt again by itself.
33. A start or a restart that finds a starting claim with a dead launcher settles it under the store lock. When the target pane has no harness process, hive removes the claim. When the target pane has a harness process, hive writes an unresolved coordinator record for that pane and changes the claim to name that record. Then the start stops. Hive tells the user to restart the coordinator from Health › Agents.
34. The restart in Health › Agents accepts a coordinator record that is live or unresolved. Under the store lock, the restart checks that the claim names that record. It replaces the claim with a new starting claim and marks the old record stale in one write. When hive started the old harness, the restart closes its pane before it starts the new harness, with the rules of `hive agent stop`. Then it follows decisions 31 and 32. When the claim names another record, the restart fails and changes nothing.
35. Tests cover these cases:
    - Two starts at the same time.
    - A prompt failure after the record write.
    - A launcher crash before the harness start.
    - A launcher crash after the harness start and before the record write, then the restart. After the restart, one coordinator harness runs.
    - A normal restart.
    - A restart of an unresolved coordinator.
36. The supervisor keeps its current start. `hive supervise` takes its claim with its process ID and start time, and refuses to start when a live supervisor holds the claim. `hive agent spawn --role manager` keeps its current registration.
37. Tests cover `hive agent register` and `hive agent spawn` for each refused role. The tests run from an empty pane, from the pane of a general agent, from the coordinator pane, and with a `--pane` override. Each attempt leaves each claim and record unchanged.
38. `AGENTS.md` removes the rule that a root session with no brief is the coordinator. A prompt from hive that names a role wins over the working directory. A session in the root with no prompt from hive has no role. It reports to the user that it has no role. The hive-coordinator skill removes its register step. The skill and the role gate table change to match.

### TUI

39. The Agents graph shows a general agent under the card of its parent, when the parent card is on the graph. This rule wins over the `home` rule. Each other general agent is in one last column with the title `home`. Health › Agents lists each general agent with the role `agent`. Fixtures cover a nested general agent, a general agent with a stopped parent, and one with the coordinator as its parent.
40. The TUI task waits for wave 6 of `tui-v2`. Before the task starts, the coordinator draws the new boards in the `tui-v2` canvas and gets the user approval. The task then follows the visual accuracy rules of `tui-v2`.

### Repository rules

41. Each Go change is test-first, as `AGENTS.md` states. The gate passes before each task lands.
42. Each new command prints text by default and one JSON object with `--json`. Each new command logs its request, its result, and each error.
43. When wave 7 passes, the coordinator resolves the friction entry `no-reviewer-role` and moves it to `entries/resolved/`.

## Manager decisions

- The package of `hive agent report` and of the general agent spawn code.
- The format of the inbox key for a general agent, inside the current name pattern.
- The exact first prompt text, inside decision 5.

## Waves

1. **Spec.** Land this README as `.hive/spec/general-agent/README.md`. Add that path to `steFiles` with a baseline of 0.
2. **Trust.** Decision 13 for each harness: claude, codex, cursor, opencode, and pi.
3. **Identity and inboxes.** Decisions 24 to 29, except the tests of decision 25 that wave 5 owns.
4. **Coordinator start.** Decisions 30 to 38. This wave depends on wave 3.
5. **Spawn and report.** Decisions 1 to 23, and the wave 5 tests of decision 25. This wave depends on waves 2, 3, and 4.
6. **TUI.** Decisions 39 and 40. This wave waits for wave 6 of `tui-v2`.
7. **Dogfood.** The coordinator spawns an agent for a plan review through hive, with pi and `openai-codex/gpt-6.1-sol`. The agent sends its result with `hive agent report`. The coordinator gets the report nudge, sends one more prompt, gets a second report, and stops the agent. The coordinator records the run in `reference/dogfood/`.

## Fog of war

- Some harnesses may have no way to skip the trust dialog. Wave 2 finds out.
- Task 13 changes the role gate in `AGENTS.md` before wave 4. Wave 4 starts from the text that task 13 lands.
