# V-red Tests red and green

Task code: `0a64edea` on `task/18-general-spawn`. Base: `dev` at `77021a80`.

## New tests

| Test | File | Criterion |
| --- | --- | --- |
| `TestSpawnGeneralAgentRecordsRoleAndCallerParent` (first failing test) | `cmd/agent/spawn/general_cli_test.go` | AC1 |
| `TestSpawnGeneralAgentFlags` | `cmd/agent/spawn/general_cli_test.go` | AC1, AC3, AC5 |
| `TestSpawnPaneHelpStatesStopResult` | `cmd/agent/spawn/general_cli_test.go` | AC4 |
| `TestGeneralSpawnNewTabRecordsParentAndSendsPromptOnce` | `internal/agent/spawn_general_test.go` | AC1, AC2, AC3, AC4 |
| `TestGeneralSpawnRefusesBeforeAnyHerdrCall` (21 cases) | `internal/agent/spawn_general_test.go` | AC1, AC2, AC3, AC5 |
| `TestGeneralSpawnNoTrustMethodNamesHarnessAndDirectory` | `internal/agent/spawn_general_test.go` | AC5 |
| `TestGeneralSpawnAllowTrustDialogKeepsPromptAndDoesNotSend` | `internal/agent/spawn_general_test.go` | AC5 |
| `TestGeneralSpawnAllowTrustDialogIdleStartKeepsPrompt` | `internal/agent/spawn_general_test.go` | AC5 |
| `TestGeneralSpawnClaudeTrustEnvironment` | `internal/agent/spawn_general_test.go` | AC5 |
| `TestGeneralSpawnClaudeInGivenPaneNeedsAllowTrustDialog` | `internal/agent/spawn_general_test.go` | AC5 |
| `TestGeneralSpawnGivenPaneUsesItsDirectoryAndMakesNoTab` (`--pane` differs from the caller pane, `--agent` override) | `internal/agent/spawn_general_test.go` | AC3, AC5 |
| `TestGeneralSpawnGivenPaneRefusesOtherCwd` | `internal/agent/spawn_general_test.go` | AC3 |
| `TestGeneralSpawnRelativeCwdUsesProcessDirectory` | `internal/agent/spawn_general_test.go` | AC3 |
| `TestGeneralSpawnOutsideHerdrUsesAgentFlag` | `internal/agent/spawn_general_test.go` | AC1, AC5 |
| `TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane` | `internal/agent/spawn_general_test.go` | AC4 |
| `TestGeneralSpawnUncertainStartKeepsPane` | `internal/agent/spawn_general_test.go` | AC4 |
| `TestGeneralSpawnUncertainPromptMakesRecordUnresolved` | `internal/agent/spawn_general_test.go` | AC2 |
| `TestGeneralSpawnStalledPromptIsNotSentAgain` | `internal/agent/spawn_general_test.go` | AC2 |
| `TestStopGeneralAgentThatHiveDidNotStartClosesNothing` | `internal/agent/stop_general_test.go` | AC4 |
| `TestStopGeneralAgentClosesItsPaneAndReportsTheTab` | `internal/agent/stop_general_test.go` | AC4 |

The package tests use the fake herdr server of `internal/agent`. The CLI tests
use the fake server of `cmd/agent/spawn`.

## Red: first failing test before the code

```text
$ go test ./cmd/agent/spawn/ -run TestSpawnGeneralAgent
--- FAIL: TestSpawnGeneralAgentRecordsRoleAndCallerParent (0.02s)
    general_cli_test.go:40: execute error = unknown flag: --prompt
FAIL
```

## Red: stop tests before the stop change

```text
$ go test ./internal/agent/ -run TestStopGeneral
--- FAIL: TestStopGeneralAgentThatHiveDidNotStartClosesNothing (0.01s)
    fakeherdr_test.go:147: unexpected method "pane.get"
    stop_general_test.go:30: envelope = ... Error:"pane get: bad reply: empty reply" ...; want a refusal
--- FAIL: TestStopGeneralAgentClosesItsPaneAndReportsTheTab (0.03s)
    --- FAIL: .../hive_tab_closes_with_the_pane: tab effect = (*output.Effect)(nil); want &output.Effect{Resource:"tab", Path:"w1:t1", Action:"closed"}
    --- FAIL: .../hive_tab_with_another_pane_stays: tab effect = (*output.Effect)(nil); want &output.Effect{Resource:"tab", Path:"w1:t1", Action:"kept: 1 other panes"}
FAIL
```

## Red: all new tests on the base

The three new test files, copied onto a detached checkout of `77021a80`:

```text
$ go test ./internal/agent/ ./cmd/agent/spawn/
# github.com/Harrison-Blair/hive/internal/agent [github.com/Harrison-Blair/hive/internal/agent.test]
internal/agent/spawn_general_test.go:27:3: unknown field Prompt in struct literal of type SpawnOptions
internal/agent/spawn_general_test.go:28:3: unknown field PromptSet in struct literal of type SpawnOptions
internal/agent/spawn_general_test.go:29:3: unknown field Cwd in struct literal of type SpawnOptions
internal/agent/spawn_general_test.go:30:3: unknown field CwdSet in struct literal of type SpawnOptions
internal/agent/spawn_general_test.go:91:13: stored.TabCreatedByHive undefined (type records.Agent has no field or method TabCreatedByHive)
internal/agent/spawn_general_test.go:91:93: stored.FirstPrompt undefined (type records.Agent has no field or method FirstPrompt)
internal/agent/spawn_general_test.go:92:96: stored.TabCreatedByHive undefined (type records.Agent has no field or method TabCreatedByHive)
internal/agent/spawn_general_test.go:92:158: stored.FirstPrompt undefined (type records.Agent has no field or method FirstPrompt)
internal/agent/spawn_general_test.go:145:57: o.PromptSet undefined (type *SpawnOptions has no field or method PromptSet)
internal/agent/spawn_general_test.go:145:78: o.Prompt undefined (type *SpawnOptions has no field or method Prompt)
internal/agent/spawn_general_test.go:145:78: too many errors
FAIL	github.com/Harrison-Blair/hive/internal/agent [build failed]
--- FAIL: TestSpawnGeneralAgentRecordsRoleAndCallerParent (0.00s)
    general_cli_test.go:40: execute error = unknown flag: --prompt
--- FAIL: TestSpawnGeneralAgentFlags (0.01s)
    --- FAIL: TestSpawnGeneralAgentFlags/both_prompts: execute error = unknown flag: --prompt
    --- FAIL: TestSpawnGeneralAgentFlags/prompt_file: execute error = unknown flag: --prompt-file
    ... (each case: unknown flag)
--- FAIL: TestSpawnPaneHelpStatesStopResult (0.00s)
    general_cli_test.go:89: help = "... --pane string       herdr pane ID ..."; want the --pane stop result
FAIL	github.com/Harrison-Blair/hive/cmd/agent/spawn
```

The base has no role `agent`, no prompt or `--cwd` flags, and no ownership or
kept prompt fields. Each failure names a missing part of the feature.

## Green on the task branch

```text
$ go test ./internal/agent/ -run 'General' -v | grep -E '^(--- |ok)'
--- PASS: TestGeneralSpawnNewTabRecordsParentAndSendsPromptOnce
--- PASS: TestGeneralSpawnRefusesBeforeAnyHerdrCall
--- PASS: TestGeneralSpawnNoTrustMethodNamesHarnessAndDirectory
--- PASS: TestGeneralSpawnAllowTrustDialogKeepsPromptAndDoesNotSend
--- PASS: TestGeneralSpawnClaudeTrustEnvironment
--- PASS: TestGeneralSpawnClaudeInGivenPaneNeedsAllowTrustDialog
--- PASS: TestGeneralSpawnGivenPaneUsesItsDirectoryAndMakesNoTab
--- PASS: TestGeneralSpawnGivenPaneRefusesOtherCwd
--- PASS: TestGeneralSpawnRelativeCwdUsesProcessDirectory
--- PASS: TestGeneralSpawnOutsideHerdrUsesAgentFlag
--- PASS: TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane
--- PASS: TestGeneralSpawnUncertainStartKeepsPane
--- PASS: TestGeneralSpawnUncertainPromptMakesRecordUnresolved
--- PASS: TestGeneralSpawnStalledPromptIsNotSentAgain
--- PASS: TestGeneralSpawnAllowTrustDialogIdleStartKeepsPrompt
--- PASS: TestStopGeneralAgentThatHiveDidNotStartClosesNothing
--- PASS: TestStopGeneralAgentClosesItsPaneAndReportsTheTab
ok  	github.com/Harrison-Blair/hive/internal/agent
$ go test ./cmd/agent/spawn/
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn
```

The full `go test -race ./...` run is in V-gate.

## Broken on purpose

Each mutation changed one line of the task code. The run used the named tests.
The file was restored after each run.

| ID | Broken behavior | Result |
| --- | --- | --- |
| M1 | Send the first prompt with `--allow-trust-dialog` | `TestGeneralSpawnAllowTrustDialogIdleStartKeepsPrompt` fails (unexpected `agent.get`) |
| M2 | No cleanup of the new tab after a failed start | `TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane` fails (both cases) |
| M3 | Cleanup ignores another pane in the tab | `.../other_pane` fails |
| M4 | No trust refusal | `no_trust_method`, `NoTrustMethodNamesHarnessAndDirectory`, `ClaudeInGivenPaneNeedsAllowTrustDialog` fail |
| M5 | Spawner text before the hive line | `TestGeneralSpawnNewTabRecordsParentAndSendsPromptOnce` fails |
| M6 | Uncertain prompt kept, not unresolved | `UncertainPromptMakesRecordUnresolved`, `StalledPromptIsNotSentAgain` fail (`record status "live"`) |
| M7 | No check of `--cwd` against the pane directory | `TestGeneralSpawnGivenPaneRefusesOtherCwd` fails |
| M8 | Caller ignores the herdr pane | `NewTab...`, `too_long`, `old_name`, `agent_override_in_herdr`, `cwd_not_a_directory` fail |
| M9 | An old (stopped) name is accepted | `.../old_name` fails (`want failed with "agent exists: gone"`) |
| M10 | Stop closes a pane that hive did not start | `TestStopGeneralAgentThatHiveDidNotStartClosesNothing` fails |
| M11 | Stop reports no tab | both tab cases of `TestStopGeneralAgentClosesItsPaneAndReportsTheTab` fail |
| M12 | Tab ownership not recorded | `TestGeneralSpawnNewTabRecordsParentAndSendsPromptOnce` fails |

M1 first passed: the allow test covered only a blocked start. The test
`TestGeneralSpawnAllowTrustDialogIdleStartKeepsPrompt` was added. Then M1
failed. The raw mutation output:

```text
### M1 send the prompt despite --allow-trust-dialog
--- FAIL: TestGeneralSpawnAllowTrustDialogIdleStartKeepsPrompt (60.19s)
### M2 no cleanup of the new tab
--- FAIL: TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane (0.03s)
    --- FAIL: TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane/root_pane_only (0.01s)
    --- FAIL: TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane/other_pane (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.069s
### M3 cleanup ignores another pane in the tab
--- FAIL: TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane (0.03s)
    --- FAIL: TestGeneralSpawnStartFailureClosesOnlyTheNewRootPane/other_pane (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.072s
FAIL
### M4 no trust refusal
--- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall (0.25s)
    --- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall/no_trust_method (0.01s)
--- FAIL: TestGeneralSpawnNoTrustMethodNamesHarnessAndDirectory (0.00s)
--- FAIL: TestGeneralSpawnClaudeInGivenPaneNeedsAllowTrustDialog (10.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	10.365s
### M5 wrong prompt order
--- FAIL: TestGeneralSpawnNewTabRecordsParentAndSendsPromptOnce (0.04s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.112s
FAIL
### M6 uncertain prompt kept instead of unresolved
--- FAIL: TestGeneralSpawnUncertainPromptMakesRecordUnresolved (0.04s)
--- FAIL: TestGeneralSpawnStalledPromptIsNotSentAgain (0.04s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.099s
### M7 --cwd differs from the pane
--- FAIL: TestGeneralSpawnGivenPaneRefusesOtherCwd (10.02s)
### M8 caller ignores the herdr pane
--- FAIL: TestGeneralSpawnNewTabRecordsParentAndSendsPromptOnce (0.01s)
--- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall (0.36s)
    --- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall/too_long (0.02s)
    --- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall/old_name (0.02s)
    --- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall/agent_override_in_herdr (0.02s)
    --- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall/cwd_not_a_directory (0.01s)
### M9 old name accepted
--- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall (0.25s)
    --- FAIL: TestGeneralSpawnRefusesBeforeAnyHerdrCall/old_name (0.01s)
FAIL
### M10 stop closes a pane hive did not start
--- FAIL: TestStopGeneralAgentThatHiveDidNotStartClosesNothing (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.081s
FAIL
### M11 stop reports no tab
--- FAIL: TestStopGeneralAgentClosesItsPaneAndReportsTheTab (0.04s)
    --- FAIL: TestStopGeneralAgentClosesItsPaneAndReportsTheTab/hive_tab_closes_with_the_pane (0.02s)
    --- FAIL: TestStopGeneralAgentClosesItsPaneAndReportsTheTab/hive_tab_with_another_pane_stays (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.062s
### M12 tab ownership not recorded
--- FAIL: TestGeneralSpawnNewTabRecordsParentAndSendsPromptOnce (0.03s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.104s
FAIL
```
