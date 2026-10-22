# Brief copy wording changes

These changes keep the brief meaning. They remove violations from the STE check.

## `.hive/spec/agent-prompts/tasks/20-herdr-status-report/brief.md`

- Original: `For each status: who sets it (herdr detection, the harness, or an API call); whether a CLI command or a socket API method can set it or clear it; and whether a change persists or herdr overwrites it.`
- Revised: `For each status, record who sets it (herdr detection, the harness, or an API call). Record whether a CLI command or a socket API method can set or clear it. Record whether the change persists or herdr overwrites it.`
- Original: `Close the scratch workspace when the investigation is done.`
- Revised: `Close the scratch workspace after the investigation ends.`

## `.hive/spec/agent-prompts/tasks/21-first-prompt-report/brief.md`

- Original: `Close each scratch workspace and stop each scratch agent when you are done.`
- Revised: `Close each scratch workspace and stop each scratch agent after the checks.`

## `.hive/spec/home-spec-planning/tasks/25-home-spec-review/brief.md`

- Original: `(22-task-edit-brief, 23-verify-clean-branch, 24-19-evidence-fix; 23 and 24 are on \`dev\`).`
- Revised: `(22-task-edit-brief, 23-verify-clean-branch, and 24-19-evidence-fix). Tasks 23 and 24 are on \`dev\`.`

## `.hive/spec/home-spec-planning/tasks/48-spec-migration/brief.md`

- Original: `Move the old briefs, evidence, and spec docs of the hive repo into .hive/spec/<feature>/.`
- Revised: `Move the old briefs, evidence, and spec docs of the hive project into .hive/spec/<feature>/.`
- Original: `When the repo does not agree with the table, stop and report it with hive inbox send --to manager-hive TEXT --agent NAME.`
- Revised: `When the project does not agree with the table, stop and report it with hive inbox send --to manager-hive TEXT --agent NAME.`
- Original: `After the moves, spec/hive-v1/ and evidence/ at the repo root have no files.`
- Revised: `After the moves, spec/hive-v1/ and evidence/ at the project root have no files.`
- Original: `Each grep line is listed in the evidence with the reason that it stays.`
- Revised: `List each grep line in the evidence with its reason for staying.`

## `.hive/spec/pr-dev-main/tasks/10-pr47-final/brief.md`

- Original: `main is unchanged at 928cfa95.`
- Revised: `main remains at 928cfa95.`

## `.hive/spec/pr-dev-main/tasks/41-pr47-refresh-2/brief.md`

- Original: `main is unchanged at 928cfa95.`
- Revised: `main remains at 928cfa95.`

## `.hive/spec/pr-dev-main/tasks/7-pr-dev-main/brief.md`

- Original: `main is unchanged at 928cfa95.`
- Revised: `main remains at 928cfa95.`
- Original: `An evidence commit only is permitted.`
- Revised: `The branch may contain evidence commits only.`

## `.hive/spec/pr-dev-main/tasks/8-pr47-body/brief.md`

- Original: `Pull request #47 has the new title and a body with Summary, Changes, LOC by category, Verification, the italic footer with a model identifier, and both grant lines.`
- Revised: `Pull request #47 has the new title and a body with Summary, Changes, LOC by category, and Verification. The body also has the italic footer with a model identifier and both grant lines.`
- Original: `main is unchanged at 928cfa95.`
- Revised: `main remains at 928cfa95.`
- Original: `An evidence commit only is permitted.`
- Revised: `The branch may contain evidence commits only.`

## `.hive/spec/pr-dev-main/tasks/9-pr47-refresh/brief.md`

- Original: `The body of pull request #47 names the starting status in Changes, gives the gate results at the new head, has both grant lines, and has a footer with the reviewed commit equal to the head of dev at review time.`
- Revised: `The body of pull request #47 names the starting status in Changes and gives the gate results at the new head. It has both grant lines. Its footer gives the reviewed commit. That commit equals the head of dev at review time.`
- Original: `main is unchanged at 928cfa95.`
- Revised: `main remains at 928cfa95.`
- Original: `An evidence commit only is permitted.`
- Revised: `The branch may contain evidence commits only.`

## `.hive/spec/sidebar-layout/tasks/3-manager-adopts-parent/brief.md`

- Original: `Stop treats that error as "pane kept": the agent record is stopped, the project record keeps the workspace, and the result has one effect that names the refusal.`
- Revised: `Stop treats that error as "pane kept": the stop marks the agent record as stopped, the project record keeps the workspace, and the result returns one effect that names the refusal.`
- Original: `Hive never sends a confirm flag.`
- Revised: `Hive never sends a confirmation flag.`
- Original: `An old-layout record with cwd projects/<id>/ is cleared. No workspace.close goes for it. The spawn continues with (b) or (c).`
- Revised: `The spawn clears an old-layout record with cwd projects/<id>/. It sends no workspace.close request for that record. The spawn continues with (b) or (c).`
- Original: `Three fake stop cases. (b) adopted: topology "surviving", the record keeps the workspace ID and drops both tab fields. (c) created, no worktrees: topology "closed", the record clears. (c) created with a worktree group: pane.close gives confirmation_required, the agent record is stopped, the project record does not change, the result has an effect that names the refusal, and no workspace.close goes.`
- Revised: `Three fake stop cases. For (b), the adopted topology is surviving. The record keeps the workspace ID and drops both tab fields. For (c), with no worktrees, the topology is closed and the record clears. For (c), with a worktree group, pane.close returns confirmation_required. The stop marks the agent record as stopped. The project record does not change. The result has an effect that names the refusal. Hive sends no workspace.close call.`
- Original: `A fake agent.start failure after tab.create in an adopted workspace gives exactly one pane.close of the pane of the new tab, zero workspace.close, and a record with no tab fields.`
- Revised: `A fake agent.start failure after tab.create in an adopted workspace gives one pane.close call on the new tab's pane. It gives zero workspace.close calls and leaves a record with no tab fields.`
- Original: `The code search below lists exactly: the definitions in internal/lib/herdr, one WorkspaceCreate( call in the manager workspace resolver, one TabCreate( call in the manager workspace resolver, and the existing TabCreate( call in spawn_worker.go.`
- Revised: `The code search below must list the definitions in internal/lib/herdr. It must list one WorkspaceCreate( call in the manager workspace resolver. It must list one TabCreate( call there. It must list the existing TabCreate( call in spawn_worker.go.`
- Original: `With the fallback, a fake that gives agent_pane_busy once and then success gives two agent.start and a spawn that succeeds; a fake that always gives agent_pane_busy fails after the deadline with the busy message.`
- Revised: `With the fallback, a fake that gives agent_pane_busy once and then success gives two agent.start and a spawn that succeeds. A fake that always gives agent_pane_busy fails after the deadline with the busy message.`

## `.hive/spec/sidebar-layout/tasks/5-tab-labels/brief.md`

- Original: `Fake herdr records one tab.rename with the agent name for a worker spawn, for a verifier spawn, and for a manager spawn into the root pane of a workspace that hive made.`
- Revised: `Fake herdr records one tab.rename with the agent name for each of three spawns: worker, verifier, and manager. The manager spawn uses the root pane of a workspace that hive made.`
