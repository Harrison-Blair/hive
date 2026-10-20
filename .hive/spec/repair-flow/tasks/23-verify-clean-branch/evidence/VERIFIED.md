# 23-verify-clean-branch VERIFIED

| Item | Value |
| --- | --- |
| Task | 23-verify-clean-branch |
| Commit | a0329a29 |
| Base | dev 457a85a4 |
| Date | 2026-10-19 |
| Agent | verifier, claude-code, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R2, R3 | pass | On a detached base worktree at 457a85a4 with the new test files copied in: the four `TestVerify*` tests fail (Verify returns `ok` and the clean-branch test cannot read the fetched ref), `TestVerifierCommitsVerdictBeforeVerify` fails for all three files, and `TestFetchBranchTipReadsRemoteBranch` fails to build (`undefined: FetchBranchTip`). On the task branch all six pass. |
| V2 | R3 | pass | The rule text is in `.agents/skills/hive-manager/SKILL.md:84-85`, `.agents/skills/hive-verifier/SKILL.md:40-41`, and `internal/agent/templates/project-agents.md:54`. |
| V3 | scope | pass | `git diff origin/dev --stat` shows only the files of the source table, the tests beside them, and `spec/hive-v1/evidence/23-verify-clean-branch/`. |
| V4 | gate | pass | `gofmt -l` prints nothing. `go vet ./...`, `golangci-lint run ./...` (0 issues), and `go test -race ./...` exit 0. |

## Runs

### V1 red (base 457a85a4)

```text
--- FAIL: TestVerifyRefusesDirtyWorktreeWithoutChanges
    verify_test.go:115: Verify() = output.Envelope{Outcome:"ok", ...}; want dirty worktree refusal
--- FAIL: TestVerifyRefusesBranchWithoutRemoteWithoutChanges
    verify_test.go:129: Verify() = output.Envelope{Outcome:"ok", ...}; want missing remote branch refusal
--- FAIL: TestVerifyRefusesBranchAheadOfOriginWithoutChanges
    verify_test.go:143: Verify() = output.Envelope{Outcome:"ok", ...}; want unpushed commit refusal
--- FAIL: TestVerifyAcceptsCleanPushedWorktree
FAIL	github.com/Harrison-Blair/hive/internal/task
--- FAIL: TestVerifierCommitsVerdictBeforeVerify
    template / manager skill / verifier skill does not contain "The verifier commits and pushes ..."
FAIL	github.com/Harrison-Blair/hive/internal/agent
internal/lib/gitx/gitx_test.go:187:22: undefined: FetchBranchTip
FAIL	github.com/Harrison-Blair/hive/internal/lib/gitx [build failed]
```

### V1 green (task branch a0329a29)

```text
--- PASS: TestVerifyRefusesDirtyWorktreeWithoutChanges
--- PASS: TestVerifyRefusesBranchWithoutRemoteWithoutChanges
--- PASS: TestVerifyRefusesBranchAheadOfOriginWithoutChanges
--- PASS: TestVerifyAcceptsCleanPushedWorktree
ok  	github.com/Harrison-Blair/hive/internal/task
--- PASS: TestVerifierCommitsVerdictBeforeVerify
ok  	github.com/Harrison-Blair/hive/internal/agent
--- PASS: TestFetchBranchTipReadsRemoteBranch
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx
```

### V4 gate

```text
gofmt -l: (no output)
go vet ./...: exit 0
golangci-lint run ./...: 0 issues., exit 0
go test -race ./...: all packages ok, exit 0
```
