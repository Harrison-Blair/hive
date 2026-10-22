# 19-kept-first-prompt V-red Tests red and green

| Item | Value |
| --- | --- |
| Task | 19-kept-first-prompt |
| Item | V-red |
| Proves | AC1, AC2, AC3 |
| Commit | f1d6673f; guard runs at 1e791a74 |
| Date | 2026-10-21 |
| Agent | worker, claude, claude-opus-5-5 |

## New tests

| Test | Criterion |
| --- | --- |
| `TestGeneralSpawnBlockedRecordsBlockedStatusAndNudgesParent` (`internal/agent/spawn_general_test.go`) | AC1: live record, agent status blocked, one blocked record in the parent inbox, no `agent.prompt`, `pane.send_text`, or `pane.send_keys` |
| `TestAgentFirstPromptTimesRoundTripAndLegacyJSON` (`internal/lib/records/agent_test.go`) | AC2: `first_prompt_attempt_at` and `first_prompt_sent_at` |
| `TestKeptFirstPromptSentOneTimeAcrossTwoPasses` | AC2: the first failing test of the brief |
| `TestKeptFirstPromptSentWhenDone` | AC2: herdr `done` is the idle status of a tab that nobody saw |
| `TestKeptFirstPromptWaitsWhileBlocked` | AC1, AC2: no send while blocked |
| `TestKeptFirstPromptRestartBetweenAttemptAndSent` | AC3: restart between steps 2 and 4 |
| `TestKeptFirstPromptUncertainSendIsUnresolvedAndNotSentAgain` | AC3: uncertain send |
| `TestKeptFirstPromptFailedAttemptWriteDoesNotSend` | AC3: failed write of step 2 |
| `TestKeptFirstPromptFailedSentWriteIsUnresolvedAndNotSentAgain` | AC3: failed write of step 4 |
| `TestKeptFirstPromptBlocksOtherNudges` | AC3: no nudge before the kept prompt |
| `TestKeptFirstPromptGoesBeforeOtherNudges` | AC3: kept prompts go first in a pass |

The `KeptFirstPrompt` tests are in `internal/supervise/firstprompt_test.go`.

## Red run

Base `origin/dev` at `ae14fe29` in a scratch worktree, with the new test files
copied in. The records test and the agent test ran on the base code as it is.
Each `$` line is the exact command.

```text
commit ae14fe29
$ go test ./internal/lib/records/ -run TestAgentFirstPromptTimesRoundTripAndLegacyJSON
--- FAIL: TestAgentFirstPromptTimesRoundTripAndLegacyJSON (0.00s)
    agent_test.go:97: first_prompt_attempt_at = <nil>; want "2026-01-02T03:04:05Z"
    agent_test.go:97: first_prompt_sent_at = <nil>; want "2026-01-02T03:04:06Z"
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/records	0.003s
FAIL
exit 1
$ go test ./internal/agent/ -run TestGeneralSpawnBlockedRecordsBlockedStatusAndNudgesParent
--- FAIL: TestGeneralSpawnBlockedRecordsBlockedStatusAndNudgesParent (0.03s)
    spawn_general_test.go:260: record status "live" agent status ""; want live and blocked
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.039s
FAIL
exit 1
```

The supervise tests need the two record fields and `firstPromptRule`. On the
same base, the scratch worktree took the record fields of the task and this
empty `internal/supervise/firstprompt.go`:

```go
package supervise

import "context"

func (s *Supervisor) firstPromptRule(ctx context.Context) error { return nil }
```

```text
$ go test ./internal/supervise/ -run KeptFirstPrompt -v
    firstprompt_test.go:64: prompts = 0; want one
--- FAIL: TestKeptFirstPromptSentOneTimeAcrossTwoPasses (0.02s)
    firstprompt_test.go:87: prompts = []; want the kept first prompt
--- FAIL: TestKeptFirstPromptSentWhenDone (0.01s)
--- PASS: TestKeptFirstPromptWaitsWhileBlocked (0.01s)
    firstprompt_test.go:125: status = "live"; want unresolved
--- FAIL: TestKeptFirstPromptRestartBetweenAttemptAndSent (0.01s)
    firstprompt_test.go:145: prompts = 0; want one
--- FAIL: TestKeptFirstPromptUncertainSendIsUnresolvedAndNotSentAgain (0.02s)
--- PASS: TestKeptFirstPromptFailedAttemptWriteDoesNotSend (0.01s)
    firstprompt_test.go:194: firstPromptRule() = nil; want the failed sent write
--- FAIL: TestKeptFirstPromptFailedSentWriteIsUnresolvedAndNotSentAgain (0.01s)
    firstprompt_test.go:225: prompts = 1; want no nudge before the kept first prompt
--- FAIL: TestKeptFirstPromptBlocksOtherNudges (0.02s)
    firstprompt_test.go:243: prompts = [{Target:w1:p1 Text:⬢ hive: 1 new inbox record, run: hive inbox Wait:<nil>}]; want the kept first prompt, then the nudge
--- FAIL: TestKeptFirstPromptGoesBeforeOtherNudges (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.125s
FAIL
exit 1
```

Each test fails on an assertion, except two guard tests:
`TestKeptFirstPromptWaitsWhileBlocked` and
`TestKeptFirstPromptFailedAttemptWriteDoesNotSend` pass on a rule that never
sends. The breaks B3 and B2 below show that each one fails when its guard
breaks.

## Green run

```text
commit f1d6673f
$ go test -v ./internal/supervise/ ./internal/agent/ ./internal/lib/records/ -run 'KeptFirstPrompt|GeneralSpawnBlocked|FirstPromptTimes' | grep -E "^(--- |ok|FAIL)"
--- PASS: TestKeptFirstPromptSentOneTimeAcrossTwoPasses (0.03s)
--- PASS: TestKeptFirstPromptSentWhenDone (0.04s)
--- PASS: TestKeptFirstPromptWaitsWhileBlocked (0.01s)
--- PASS: TestKeptFirstPromptRestartBetweenAttemptAndSent (0.04s)
--- PASS: TestKeptFirstPromptUncertainSendIsUnresolvedAndNotSentAgain (0.04s)
--- PASS: TestKeptFirstPromptFailedAttemptWriteDoesNotSend (0.01s)
--- PASS: TestKeptFirstPromptFailedSentWriteIsUnresolvedAndNotSentAgain (0.04s)
--- PASS: TestKeptFirstPromptBlocksOtherNudges (0.02s)
--- PASS: TestKeptFirstPromptGoesBeforeOtherNudges (0.05s)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
--- PASS: TestGeneralSpawnBlockedRecordsBlockedStatusAndNudgesParent (0.07s)
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
--- PASS: TestAgentFirstPromptTimesRoundTripAndLegacyJSON (0.00s)
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
exit 0
```

## Breaks on purpose

Each break changed one guard in the task code, ran the tests, then restored
the file. `git status --short` printed nothing after the last restore.

```text
commit f1d6673f
## B1: no step 1 check, and no unresolved for an attempt with no sent time
$ go test ./internal/supervise/ ./internal/agent/ -run 'KeptFirstPrompt|GeneralSpawnBlocked'
--- FAIL: TestKeptFirstPromptRestartBetweenAttemptAndSent (0.02s)
    firstprompt_test.go:122: prompts = 1; want none after a restart between the steps
--- FAIL: TestKeptFirstPromptFailedSentWriteIsUnresolvedAndNotSentAgain (0.02s)
    firstprompt_test.go:205: prompts = 2; want one
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.166s
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
FAIL
exit 1

## B2: send after a failed step 2 write
$ go test ./internal/supervise/ ./internal/agent/ -run 'KeptFirstPrompt|GeneralSpawnBlocked'
--- FAIL: TestKeptFirstPromptFailedAttemptWriteDoesNotSend (0.01s)
    firstprompt_test.go:172: prompts = 1; want none after a failed attempt write
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.177s
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
FAIL
exit 1

## B3: send with any agent status
$ go test ./internal/supervise/ ./internal/agent/ -run 'KeptFirstPrompt|GeneralSpawnBlocked'
--- FAIL: TestKeptFirstPromptWaitsWhileBlocked (0.01s)
    firstprompt_test.go:101: prompts = 1; want none while blocked
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.186s
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
FAIL
exit 1

## B4: idle only, done not accepted
$ go test ./internal/supervise/ ./internal/agent/ -run 'KeptFirstPrompt|GeneralSpawnBlocked'
--- FAIL: TestKeptFirstPromptSentWhenDone (0.01s)
    firstprompt_test.go:87: prompts = []; want the kept first prompt
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.171s
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
FAIL
exit 1

## B5: nudges go first and go to an agent with a pending kept prompt
$ go test ./internal/supervise/ ./internal/agent/ -run 'KeptFirstPrompt|GeneralSpawnBlocked'
--- FAIL: TestKeptFirstPromptBlocksOtherNudges (0.02s)
    firstprompt_test.go:225: prompts = 1; want no nudge before the kept first prompt
--- FAIL: TestKeptFirstPromptGoesBeforeOtherNudges (0.03s)
    firstprompt_test.go:243: prompts = [{Target:w1:p1 Text:⬢ hive: 1 new inbox record, run: hive inbox Wait:<nil>} {Target:w1:p5 Text:⬢ You are the general agent helper. Your parent is boss. Wait:<nil>}]; want the kept first prompt, then the nudge
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.177s
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
FAIL
exit 1

## B6: spawn writes no blocked agent status and no parent record
$ go test ./internal/supervise/ ./internal/agent/ -run 'KeptFirstPrompt|GeneralSpawnBlocked'
ok  	github.com/Harrison-Blair/hive/internal/supervise	0.188s
--- FAIL: TestGeneralSpawnBlockedRecordsBlockedStatusAndNudgesParent (0.04s)
    spawn_general_test.go:260: record status "live" agent status ""; want live and blocked
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.050s
FAIL
exit 1

```

## Guard tests

The brief has an exception for guard tests (Changes, 2026-10-21). A guard
test checks that hive sends nothing. It can pass on the base, because the
base sends no kept first prompt. To prove each guard test, the guard was
removed on purpose. Each test must then fail.

| Guard test | Guard | Base | Guard removed |
| --- | --- | --- | --- |
| `TestKeptFirstPromptWaitsWhileBlocked` | Send only when the agent status is idle or done | pass | fail: `prompts = 1; want none while blocked` |
| `TestKeptFirstPromptFailedAttemptWriteDoesNotSend` | No send when the step 2 write fails | pass | fail: `prompts = 1; want none after a failed attempt write` |

Each run below is on the merge of `origin/dev` at `1e791a74`. Each break
changed one guard in `internal/supervise/firstprompt.go`, ran the test, and
restored the file with `git checkout`. `git status --short` printed nothing
after the last restore.

```text
commit 1e791a74
## G1: remove the status guard of TestKeptFirstPromptWaitsWhileBlocked
guard (lines 37-39):
		if record.AgentStatus != "idle" && record.AgentStatus != "done" {
			continue
		}
break: line 37 becomes `if false {`
$ go test -count=1 ./internal/supervise/ -run TestKeptFirstPromptWaitsWhileBlocked
--- FAIL: TestKeptFirstPromptWaitsWhileBlocked (0.02s)
    firstprompt_test.go:101: prompts = 1; want none while blocked
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.028s
FAIL
exit 1

## G2: remove the failed attempt write guard of TestKeptFirstPromptFailedAttemptWriteDoesNotSend
guard (lines 63-66):
	if err != nil {
		s.logger.Debug("supervisor first prompt attempt write failed", ...)
		return nil
	}
break: line 65 becomes `attempted = true`, so the send continues
$ go test -count=1 ./internal/supervise/ -run TestKeptFirstPromptFailedAttemptWriteDoesNotSend
--- FAIL: TestKeptFirstPromptFailedAttemptWriteDoesNotSend (0.02s)
    firstprompt_test.go:172: prompts = 1; want none after a failed attempt write
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/supervise	0.020s
FAIL
exit 1

## Guards restored
$ go test -count=1 ./internal/supervise/ -run "TestKeptFirstPromptWaitsWhileBlocked|TestKeptFirstPromptFailedAttemptWriteDoesNotSend" -v
--- PASS: TestKeptFirstPromptWaitsWhileBlocked (0.01s)
--- PASS: TestKeptFirstPromptFailedAttemptWriteDoesNotSend (0.01s)
ok  	github.com/Harrison-Blair/hive/internal/supervise	0.021s
exit 0
```

The breaks B3 and B2 above at `f1d6673f` gave the same failures.

## Result

Nine new tests fail on the base for the correct reason. The two guard tests
`TestKeptFirstPromptWaitsWhileBlocked` and
`TestKeptFirstPromptFailedAttemptWriteDoesNotSend` pass on the base, which the
guard test exception of the brief permits. Each guard test fails when its
guard is removed on purpose (G1, G2). Each new test passes on the task branch.
Each of the six breaks B1 to B6 fails at least one new test.
