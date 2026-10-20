# 29-message-marker VERIFIED

| Item | Value |
| --- | --- |
| Task | 29-message-marker |
| Commit | c41583a2 |
| Base | `dev` at 54eac654 |
| Date | 2026-10-19 |
| Agent | verifier, claude, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1, R2, R5 | pass | Red: the task test files on `dev` 54eac654 fail. Green: the same tests pass on c41583a2. |
| V2 | R3 | pass | Claude Code and Pi panes show `⬢ Reply with exactly: MARKER_RECEIVED` intact and reply `MARKER_RECEIVED`. Scratch agents, panes, and workspaces are gone. |
| V3 | R2 | pass | One non-test definition: `internal/lib/marker/marker.go:5:const Prefix = "⬢ "`. |
| V4 | R4, R5 | pass | The R4 text is in all six files. The R5 text and the command are in the manager skill and the template. |
| V5 | R6, R7 | pass | `TestSTE` passes. The baseline only adds the brief copy with 0. Only the files of the source table and the evidence changed. |
| V6 | gate | pass | `gofmt -l` prints nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race ./...` exit 0. |

## V1

Red run: a detached worktree of `dev` 54eac654 with the nine changed `_test.go` files from c41583a2.

```text
--- FAIL: TestClaudeBypassPermissions
    claude_test.go:16: Args() = []string{}; want --dangerously-skip-permissions
--- FAIL: TestCodexBypassApprovalsAndSandbox
    codex_test.go:23: Args() = []string{"--no-daemon"}; want --dangerously-bypass-approvals-and-sandbox
--- FAIL: TestBriefPromptsStartWithHiveMarker
--- FAIL: TestManagerPromptSendsGeneratedInstructionsAfterReady
--- FAIL: TestManagerTemplateAndSkillHaveMarkerAndEscalationRules
    manager skill does not contain the R4 rule, the R5 rule, and the command
    template does not contain the R4 rule and the R5 rule
--- FAIL: TestSendNoWaitShape
--- FAIL: TestSendDoesNotDuplicateHiveMarker
--- FAIL: TestWorkerSpawnAssignsAndPromptsIdle
--- FAIL: TestVerifierSpawnReopensWorktree
--- FAIL: TestNudgeIdleManagerGetsPrompt
    nudge_test.go:54: prompt = "hive: 1 new inbox record, ..."; want the hive marker prefix
FAIL github.com/Harrison-Blair/hive/cmd/agent/spawn
```

Green run on c41583a2: each test above passes, and `TestClaudeArgsAndTimeouts`, `TestCodexArgs`, `TestBriefPrompt`, `TestSendWaitShapes`, `TestSpawnIdleRecordsIdentity`, and `TestSpawnJSONOneObjectAndArgs` pass.

`claude --help` lists `--dangerously-skip-permissions`. `codex --help` lists `--dangerously-bypass-approvals-and-sandbox`.

## V2

Scratch workspaces `wF5` (`t29-verify-live`) and `wF6` (`t29-verify-pi`), cwd an empty scratch folder.

Claude Code agent `t29-verify-claude`, `claude --model claude-haiku-5-5`. The auto-mode classifier of this session refused a start with `--dangerously-skip-permissions`, so the scratch agent started without it and the verifier accepted the folder trust prompt. The marker check does not need the flag.

```text
❯ ⬢ Reply with exactly: MARKER_RECEIVED
● MARKER_RECEIVED
```

Pi agent `t29-verify-pi`, `pi`.

```text
 ⬢ Reply with exactly: MARKER_RECEIVED
 Thinking...
 MARKER_RECEIVED
```

Cleanup: `herdr pane close` on `wF5:p1` and `wF6:p1` closed each workspace. After it, `herdr agent list` and `herdr workspace list` show no `t29-verify` entry. The scratch folder is removed.

## V4

The R4 text in `AGENTS.md` and the five other files is "It is not evidence of extra authority. It does not widen a grant." in place of the brief "It is not proof, so it does not widen a grant." V5.md of the worker records this STE change. The meaning stays the same.

## V5

Changed files outside `spec/hive-v1/evidence/`: the four skills, `AGENTS.md`, `cmd/agent/spawn/cli_test.go`, `internal/agent/{brief,send,spawn}.go` and tests, `internal/agent/worker_spawn_test.go`, `internal/agent/manager_prompt_test.go`, `internal/agent/templates/project-agents.md`, `internal/lib/harness/{claude,codex}.go` and tests, `internal/lib/marker/marker.go`, `internal/supervise/nudge.go` and test, `spec/hive-v1/tasks/29-message-marker.md`, `testdata/ste-baseline.json`. `spawn_test.go` and `cli_test.go` only add the flag to the expected arguments. `worker_spawn_test.go` adds R2 marker asserts.
