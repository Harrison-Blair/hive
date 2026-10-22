# T21 role playbooks

| Item | Value |
| --- | --- |
| Task ID | T21-role-playbooks |
| Wave | 5 |
| Tier | strong |
| Depends on | T07, T12, T13, T14, T15, CP3 |
| Blocks | CP5 |
| Branch | `task/T21-role-playbooks` |
| Worktree | `projects/hive/worktrees/T21-role-playbooks/` |
| Evidence | `evidence/T21-role-playbooks/` |

## Overview

This task writes the procedures that each role follows, as portable skills
under `.agents/skills/`. It also writes the final text of the generated project
instructions and points the root `AGENTS.md` at the skills. The procedures use
only commands that exist. A test checks that rule. The task adds no new
command.

A harness discovers a project skill only from the repository root of its
working directory. A worker in a worktree of another registered repository
does not see `.agents/skills/` of the home. The first prompt of a worker or a
verifier and the generated manager instructions therefore name the absolute
path of the playbook file, and the agent reads it from there.

## User flows

1. A new coordinator session loads `hive-coordinator` and follows its startup. The steps: register, read the handoff, read the inbox, list the agents, run `doctor`, start the supervisor when it is down, and report open decisions.
2. A manager loads `hive-manager` and reads the presets. It creates tasks with briefs and dependencies, makes worktrees, spawns workers by tier, and lands verified tasks under a grant.
3. A worker in a worktree of another repository gets the first prompt of T14. The prompt names `<home>/.agents/skills/hive-worker/SKILL.md`. The worker reads the file, then the brief. It works test-first and commits and pushes after each green run. It writes the evidence, runs the gate, and runs `task complete`.
4. A verifier gets its first prompt with the path of `hive-verifier`. It re-runs each item, writes `VERIFIED.md`, and runs `task verify`.
5. A Claude Code session in the home types `/hive-worker`. A Codex session in the home types `$hive-worker`. Each shows the skill.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.agents/skills/hive-coordinator/SKILL.md` | New. |
| `.agents/skills/hive-manager/SKILL.md` | New. |
| `.agents/skills/hive-worker/SKILL.md` | New. |
| `.agents/skills/hive-verifier/SKILL.md` | New. |
| `.claude/skills/<name>` | New. One tracked directory symbolic link for each skill, with the relative target `../../.agents/skills/<name>`. |
| `.gitignore` | Allow the four symbolic links with the narrowest pattern. |
| `internal/agent/templates/project-agents.md` | The final text of the manager instructions, with the absolute playbook path. |
| `internal/agent/templates/presets.md` | New. The default presets document with three tiers. |
| `internal/agent/instructions.go` | Write the default presets document when the file is absent, under the store lock. |
| `AGENTS.md` | Point the coordinator section at the skill. Keep the bootstrap clause. |
| `ste_test.go` | Add the four skills and the presets template to `steFiles`. |
| `testdata/ste-baseline.json` | Add one entry with 0 for each new file. |
| `skills_test.go` | New. Each hive command in a skill or a template exists in the command tree. |

This source table is the named exception to the common rule that a task does not
edit a skill, `AGENTS.md`, or a global setting. T21 may edit only the named skill
files, the relative `.claude/skills` link, `.gitignore`, `AGENTS.md`, `ste_test.go`,
and `testdata/ste-baseline.json`. It does not change a global setting or the
`agnostic-skill` files.

## Requirements

### AC1 Skills

- Each skill is a `SKILL.md` with the frontmatter `name` and `description`, as the user skill `agnostic-skill` at `~/.claude/skills/agnostic-skill/SKILL.md` describes. The worker follows that procedure for Claude Code and Codex discovery.
- Each skill is in ASD-STE100 Strict and names each command with its flags.
- The coordinator skill gives: startup, preference records in `overrides/`, the spawn of a manager, and the start of the supervisor. It also gives the handoff at the end and the rule that the coordinator never changes project source.
- The manager skill gives:
  - the project instructions, the presets, and the tiers,
  - task creation with briefs and dependencies, worktree creation, and worker spawn,
  - the inbox rule: read the inbox at startup and after each nudge. Act on each unread record. End the turn when no approved work is ready. Do not poll or wait in a loop,
  - the verifier spawn after a `completed` record,
  - landing under a grant from `overrides/grants.md`, worktree removal, memory, and handoff,
  - escalation with `hive inbox send --to coordinator`.
- The worker skill gives the brief, test-first work, evidence, and gate steps. It requires a commit and push on the task branch after each green run. The worker does not wait for a user request. It names `task complete` and says not to edit the tracker.
- The verifier skill gives: the items, `VERIFIED.md`, and `task verify` with or without `--fail`.
- Each skill that an agent reads from a path, not through the harness, starts with the same frontmatter and works as plain text.

### AC2 Discovery

- Codex discovers `.agents/skills/` itself. Claude Code discovers `.claude/skills/`. The task adds one tracked directory symbolic link `.claude/skills/<name>` for each skill, with the relative target `../../.agents/skills/<name>`.
- The `agnostic-skill` procedure asks for an absolute link target. A tracked link must work in each clone of this repository, so this task uses a relative target. This brief states that exception. The worker does not change the `agnostic-skill` files or a global setting.
- The discovery test checks the real Claude Code path and the real Codex path. The Claude Code session uses `.claude/skills/<name>`. The Codex session uses `.agents/skills/<name>`. The content review stays separate in V5.
- T14 writes the first sentence of the worker prompt and the verifier prompt: `Read <home>/.agents/skills/hive-worker/SKILL.md first.`, or the verifier path. This task gives the skill directories the names that those paths use.
- The project instructions template names `<home>/.agents/skills/hive-manager/SKILL.md` as an absolute path and tells the manager to read it.

### AC3 Templates

- The project instructions template has this order: the generated-file line, the role gate, the absolute path of the manager skill, and the delegate rule. The delegate rule sends each source change to a worker.
- The presets template has three headings: strong, decent, small. Each has one prose paragraph that says what kind of task it serves and one ordered list of harness and model choices.
- Hive writes the presets document to `overrides/presets.md` at a manager spawn when the file is absent. The rule covers a home used before T21. The check for the file and the write run under the store lock. An existing file keeps each byte, also when two first spawns run at the same time.
- A failed first write does not become an existing-file success. A partial temporary file does not become an existing-file success. A write error other than "file exists" fails the spawn with the error.
- Before the write, check the static type of the home overrides directory, each ancestor, and the destination. Reject a symbolic link and a wrong type.

### AC4 Command check

- `skills_test.go` reads each skill and the two templates. It finds each command line that starts with `hive`. It skips `--json` before the command path. It checks the command path against the tree from `cmd.NewRootCmd()`. It does not run a command.
- An unknown command fails the test with the file, the line, and the command.

### AC5 Gate

- The four gate commands pass from the root of the worktree. The ratchet covers the new files with a baseline of 0.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | The four skills exist with the content of the requirements. | V1, V5 |
| AC2 | Claude Code and Codex find the skills in the home, and the prompts and the template carry the absolute playbook path. | V3, V4 |
| AC3 | The templates have the content of the requirements, and the presets file appears once and keeps its bytes. | V2 |
| AC4 | Each command that a skill or a template names exists, also after persistent flags. | V1 |
| AC5 | The gate passes with the new ratchet entries. | V6 |

## Verification items

### V1 Command check test

- Proves: AC1, AC4.
- Command: `go test . -run 'TestSkillCommands|TestSTE' -v`, red and green. The red run has a checker with an empty body and a fixture skill with the lines `hive agent fly` and `hive --json agent fly`.
- Expected: the red run fails on the two unknown commands. The green run passes with the four skills, the two templates, and a ratchet count of 0 for each. The green run also accepts `hive task complete --task 1-step`.

### V2 Templates tests

- Proves: AC3.
- Command: `go test ./internal/agent/ -run 'TestPresets|TestProjectInstructions' -v`, red and green.
- Expected: the first spawn writes `overrides/presets.md` with the three headings. A second spawn with a changed file does not replace it. Two first spawns at the same time leave one file with the bytes of the first writer. A file that exists before the first spawn keeps each byte. A failed first write and a partial temporary file do not become an existing-file success. A linked overrides directory, a linked destination, and a wrong type fail. A home used before T21 gets the file at its next manager spawn. The rendered instructions have the generated-file line first, the role gate, the absolute path of the manager skill, and the delegate rule.

### V3 Prompt paths

- Proves: AC2.
- Command: `go test ./internal/agent/ -run TestBrief -v`, then `test -f .agents/skills/hive-worker/SKILL.md`, `test -f .agents/skills/hive-verifier/SKILL.md`, and `test -f .claude/skills/hive-worker/SKILL.md`, each with `echo "exit $?"`. One run, because T14 owns the test.
- Expected: the T14 test passes, and each path that the prompts name exists in the worktree, through the symbolic link too.

### V4 Live skills

- Proves: AC2.
- Command: three sessions. First, a new Claude Code session in the worktree: type `/hive-coordinator`, `/hive-manager`, `/hive-worker`, and `/hive-verifier`. Second, a new Codex session in the worktree: type `$hive-coordinator`, `$hive-manager`, `$hive-worker`, and `$hive-verifier`. Third, a new Claude Code session in a scratch repository under `.hive/tmp/` with no hive files. Give it the worker prompt of T14 with the home of the worktree, and record what it reads.
- Expected: each of the eight skill calls shows the frontmatter and the first lines of the skill. The Claude Code session finds the skills through the relative `.claude/skills` link. The Codex session finds `.agents/skills` itself. The third session reads the playbook from the absolute path. The evidence records the first lines of each session with `herdr pane read`.

### V5 Content review

- Proves: AC1.
- Command: none. The worker writes a table with one row for each bullet of AC1. Each row names the skill file and the heading that gives the procedure.
- Expected: each row has a heading. The verifier reads each heading and marks the row pass or fail.

### V6 Gate

- Proves: AC5.
- Command: the four gate commands of `AGENTS.md`.
- Expected: `gofmt -l` prints nothing. The other three pass.

## Out of scope

- The end of the bootstrap clause. The user decides at CP5.
- Skills for the Pi, OpenCode, and Cursor harnesses. The canonical skill files serve them when they discover `.agents/skills/`. The absolute path in the prompts serves them otherwise.

## Changes

2026-10-18. State the relative `.claude/skills` link exception. Keep the
`agnostic-skill` files and global settings unchanged. Add the absent-file rule
for a home used before T21, the failed-write rule, and the static type checks.
Keep the real Claude Code and Codex discovery check in V4 and the content review
in V5.
