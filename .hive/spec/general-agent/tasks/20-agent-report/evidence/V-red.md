# V-red Tests red and green

- Proves: AC1, AC2, AC3, AC4.
- Task commit: `1c83186b`. Base: `dev` at `ae14fe29`.

## New tests

| Test | Criterion |
| --- | --- |
| `internal/agent` `TestReportStdinWritesOneParentInboxRecord` (first failing test of the brief) | AC1, AC2 |
| `TestReportStdinLimits` (4096 bytes, 4097 bytes, invalid UTF-8) | AC1, decision 19 |
| `TestReportFileKeepsAbsolutePathAndCutsAtRuneBoundary`, `TestReportFileRefusals` | AC1, decision 20 |
| `TestReportRefusesEachOtherRole` | AC1, decision 18 |
| `TestReportRoutesWithTheParentRule` | AC2, decision 28 |
| `TestReportCallerOverrides` (`--agent` and `--pane` overrides) | AC2, decisions 24, 25 |
| `TestReportSetsReportedAtAndNotFinished`, `TestReportTwiceWritesTwoRecords` | AC3, decisions 21, 22 |
| `TestReportAgentWriteFailureListsTheInboxRecord` | AC3, decision 21 failure rule |
| `cmd/agent/report` `TestReportReadsStandardInputToTheParentInbox`, `TestReportJSONAndFile`, `TestReportCallerOverrides`, `TestReportOutsideHerdrUsesTheAgentFlag` | AC1, AC2, decision 42 |
| `internal/lib/inbox` `TestAppendReportKindKeepsPath` | AC2 |
| `internal/inbox` `TestReadTextShowsTheReportPath` | AC2 |
| root `TestGeneralAgentSkill` | AC4 |

## Red on the base

The new test files go into a detached worktree of `dev`. The build fails
because the report API, the kind `report`, the field `path`, and
`reported_at` do not exist. The skill on the base fails the command check,
because `hive agent report` does not exist.

```
# github.com/Harrison-Blair/hive/internal/lib/inbox [github.com/Harrison-Blair/hive/internal/lib/inbox.test]
internal/lib/inbox/inbox_test.go:795:99: unknown field Path in struct literal of type AppendOptions
internal/lib/inbox/inbox_test.go:800:70: all[0].Path undefined (type records.Inbox has no field or method Path)
# github.com/Harrison-Blair/hive/internal/inbox [github.com/Harrison-Blair/hive/internal/inbox.test]
internal/inbox/render_test.go:55:22: undefined: records.InboxReport
internal/inbox/render_test.go:59:3: unknown field Path in struct literal of type records.Inbox
# github.com/Harrison-Blair/hive/internal/agent [github.com/Harrison-Blair/hive/internal/agent.test]
internal/agent/report_test.go:51:14: undefined: Report
internal/agent/report_test.go:51:27: undefined: ReportOptions
internal/agent/report_test.go:56:129: found[0].Path undefined (type records.Inbox has no field or method Path)
internal/agent/report_test.go:63:17: undefined: Report
internal/agent/report_test.go:63:30: undefined: ReportOptions
internal/agent/report_test.go:67:12: record.ReportedAt undefined (type records.Agent has no field or method ReportedAt)
internal/agent/report_test.go:75:18: undefined: Report
internal/agent/report_test.go:75:31: undefined: ReportOptions
internal/agent/report_test.go:94:16: undefined: Report
internal/agent/report_test.go:94:29: undefined: ReportOptions
internal/agent/report_test.go:94:29: too many errors
FAIL	github.com/Harrison-Blair/hive/internal/agent [build failed]
# github.com/Harrison-Blair/hive/cmd/agent/report_test [github.com/Harrison-Blair/hive/cmd/agent/report.test]
cmd/agent/report/report_test.go:109:91: response.Result.Record.Path undefined (type records.Inbox has no field or method Path)
FAIL	github.com/Harrison-Blair/hive/cmd/agent/report [build failed]
FAIL	github.com/Harrison-Blair/hive/internal/lib/inbox [build failed]
FAIL	github.com/Harrison-Blair/hive/internal/inbox [build failed]
FAIL	github.com/Harrison-Blair/hive	0.014s
```

```
--- FAIL: TestSkillCommands (0.00s)
    --- FAIL: TestSkillCommands/each_real_file_names_a_command (0.00s)
        skills_test.go:50: .agents/skills/hive-agent/SKILL.md:47: unknown command: agent report
FAIL
FAIL	github.com/Harrison-Blair/hive	1.249s
FAIL
```

## Green on the task branch

```
--- PASS: TestManagerNotRepositoryReportsWorktreeListError (0.01s)
--- PASS: TestReportStdinWritesOneParentInboxRecord (0.02s)
--- PASS: TestReportSetsReportedAtAndNotFinished (0.02s)
--- PASS: TestReportTwiceWritesTwoRecords (0.03s)
--- PASS: TestReportRefusesEachOtherRole (0.03s)
    --- PASS: TestReportRefusesEachOtherRole/worker (0.00s)
    --- PASS: TestReportRefusesEachOtherRole/verifier (0.00s)
    --- PASS: TestReportRefusesEachOtherRole/manager (0.01s)
    --- PASS: TestReportRefusesEachOtherRole/coordinator (0.01s)
    --- PASS: TestReportRefusesEachOtherRole/supervisor (0.01s)
--- PASS: TestReportStdinLimits (0.03s)
    --- PASS: TestReportStdinLimits/4096_bytes (0.02s)
    --- PASS: TestReportStdinLimits/4097_bytes (0.01s)
    --- PASS: TestReportStdinLimits/invalid_UTF-8 (0.01s)
--- PASS: TestReportFileKeepsAbsolutePathAndCutsAtRuneBoundary (0.02s)
--- PASS: TestReportFileRefusals (0.03s)
    --- PASS: TestReportFileRefusals/absent (0.01s)
    --- PASS: TestReportFileRefusals/directory (0.01s)
    --- PASS: TestReportFileRefusals/invalid_UTF-8 (0.01s)
--- PASS: TestReportRoutesWithTheParentRule (0.06s)
    --- PASS: TestReportRoutesWithTheParentRule/stopped_parent_goes_to_its_live_successor (0.04s)
    --- PASS: TestReportRoutesWithTheParentRule/no_parent_goes_to_coordinator (0.02s)
--- PASS: TestReportCallerOverrides (0.11s)
    --- PASS: TestReportCallerOverrides/agent_override_in_a_pane (0.02s)
    --- PASS: TestReportCallerOverrides/pane_override_in_a_pane (0.01s)
    --- PASS: TestReportCallerOverrides/own_agent_in_a_pane (0.03s)
    --- PASS: TestReportCallerOverrides/own_pane_in_a_pane (0.02s)
    --- PASS: TestReportCallerOverrides/agent_outside_herdr (0.02s)
    --- PASS: TestReportCallerOverrides/pane_outside_herdr (0.02s)
--- PASS: TestReportAgentWriteFailureListsTheInboxRecord (0.01s)
--- PASS: TestRestartSupervisorKeepsUnownedPaneAndReportsIt (0.01s)
--- PASS: TestStopGeneralAgentClosesItsPaneAndReportsTheTab (0.03s)
    --- PASS: TestStopGeneralAgentClosesItsPaneAndReportsTheTab/hive_tab_closes_with_the_pane (0.01s)
    --- PASS: TestStopGeneralAgentClosesItsPaneAndReportsTheTab/hive_tab_with_another_pane_stays (0.01s)
    --- PASS: TestStopGeneralAgentClosesItsPaneAndReportsTheTab/given_pane (0.01s)
ok  	github.com/Harrison-Blair/hive/internal/agent	0.431s
--- PASS: TestReportReadsStandardInputToTheParentInbox (0.03s)
--- PASS: TestReportJSONAndFile (0.03s)
--- PASS: TestReportCallerOverrides (0.09s)
    --- PASS: TestReportCallerOverrides/agent_override (0.01s)
    --- PASS: TestReportCallerOverrides/pane_override (0.01s)
    --- PASS: TestReportCallerOverrides/own_agent (0.03s)
    --- PASS: TestReportCallerOverrides/own_pane (0.03s)
--- PASS: TestReportOutsideHerdrUsesTheAgentFlag (0.03s)
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	0.178s
--- PASS: TestAppendReportKindKeepsPath (0.01s)
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	0.017s
--- PASS: TestReadTextShowsTheReportPath (0.00s)
ok  	github.com/Harrison-Blair/hive/internal/inbox	0.005s
```

## Broken on purpose

Each break goes into the task code, the tests run, and `git checkout` restores the file.

```
### break: inbox lib refuses kind report
--- FAIL: TestAppendReportKindKeepsPath (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/lib/inbox	0.003s
--- FAIL: TestReportStdinWritesOneParentInboxRecord (0.01s)
--- FAIL: TestReportSetsReportedAtAndNotFinished (0.01s)
--- FAIL: TestReportTwiceWritesTwoRecords (0.01s)
--- FAIL: TestReportStdinLimits (0.04s)
    --- FAIL: TestReportStdinLimits/4096_bytes (0.01s)
--- FAIL: TestReportFileKeepsAbsolutePathAndCutsAtRuneBoundary (0.01s)
--- FAIL: TestReportRoutesWithTheParentRule (0.02s)
    --- FAIL: TestReportRoutesWithTheParentRule/stopped_parent_goes_to_its_live_successor (0.02s)
    --- FAIL: TestReportRoutesWithTheParentRule/no_parent_goes_to_coordinator (0.01s)
--- FAIL: TestReportCallerOverrides (0.07s)
    --- FAIL: TestReportCallerOverrides/own_agent_in_a_pane (0.01s)
    --- FAIL: TestReportCallerOverrides/own_pane_in_a_pane (0.01s)
    --- FAIL: TestReportCallerOverrides/agent_outside_herdr (0.01s)
    --- FAIL: TestReportCallerOverrides/pane_outside_herdr (0.01s)
--- FAIL: TestReportAgentWriteFailureListsTheInboxRecord (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.292s
FAIL
### break: report sets finished
--- FAIL: TestReportSetsReportedAtAndNotFinished (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.420s
FAIL
### break: no role check
--- FAIL: TestReportRefusesEachOtherRole (0.09s)
    --- FAIL: TestReportRefusesEachOtherRole/worker (0.01s)
    --- FAIL: TestReportRefusesEachOtherRole/verifier (0.01s)
    --- FAIL: TestReportRefusesEachOtherRole/manager (0.02s)
    --- FAIL: TestReportRefusesEachOtherRole/coordinator (0.02s)
    --- FAIL: TestReportRefusesEachOtherRole/supervisor (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.517s
FAIL
### break: no rune boundary cut
--- FAIL: TestReportFileKeepsAbsolutePathAndCutsAtRuneBoundary (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.459s
FAIL
### break: caller ignores the herdr pane
--- FAIL: TestReportStdinWritesOneParentInboxRecord (0.01s)
--- FAIL: TestReportSetsReportedAtAndNotFinished (0.01s)
--- FAIL: TestReportTwiceWritesTwoRecords (0.01s)
--- FAIL: TestReportRefusesEachOtherRole (0.05s)
    --- FAIL: TestReportRefusesEachOtherRole/worker (0.01s)
    --- FAIL: TestReportRefusesEachOtherRole/verifier (0.01s)
    --- FAIL: TestReportRefusesEachOtherRole/manager (0.01s)
    --- FAIL: TestReportRefusesEachOtherRole/coordinator (0.01s)
    --- FAIL: TestReportRefusesEachOtherRole/supervisor (0.01s)
--- FAIL: TestReportStdinLimits (0.03s)
    --- FAIL: TestReportStdinLimits/4096_bytes (0.01s)
--- FAIL: TestReportFileKeepsAbsolutePathAndCutsAtRuneBoundary (0.01s)
--- FAIL: TestReportRoutesWithTheParentRule (0.02s)
    --- FAIL: TestReportRoutesWithTheParentRule/stopped_parent_goes_to_its_live_successor (0.01s)
    --- FAIL: TestReportRoutesWithTheParentRule/no_parent_goes_to_coordinator (0.00s)
--- FAIL: TestReportCallerOverrides (0.11s)
    --- FAIL: TestReportCallerOverrides/agent_override_in_a_pane (0.02s)
    --- FAIL: TestReportCallerOverrides/pane_override_in_a_pane (0.02s)
--- FAIL: TestReportAgentWriteFailureListsTheInboxRecord (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.343s
--- FAIL: TestReportReadsStandardInputToTheParentInbox (0.01s)
--- FAIL: TestReportJSONAndFile (0.01s)
--- FAIL: TestReportCallerOverrides (0.12s)
    --- FAIL: TestReportCallerOverrides/agent_override (0.03s)
    --- FAIL: TestReportCallerOverrides/pane_override (0.03s)
FAIL
FAIL	github.com/Harrison-Blair/hive/cmd/agent/report	0.179s
FAIL
### break: agent write failure reported as ok
--- FAIL: TestReportAgentWriteFailureListsTheInboxRecord (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.367s
FAIL
### break: route to caller instead of parent rule
--- FAIL: TestReportRoutesWithTheParentRule (0.05s)
    --- FAIL: TestReportRoutesWithTheParentRule/stopped_parent_goes_to_its_live_successor (0.03s)
    --- FAIL: TestReportRoutesWithTheParentRule/no_parent_goes_to_coordinator (0.02s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.445s
FAIL
### break: no stdin limit
--- FAIL: TestReportStdinLimits (0.04s)
    --- FAIL: TestReportStdinLimits/4097_bytes (0.01s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/agent	0.443s
FAIL
### break: render drops the path
--- FAIL: TestReadTextShowsTheReportPath (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/inbox	0.176s
FAIL
```

## Live run

A built binary with a temp `HIVE_HOME` and no herdr. The records are fixtures.

```
$ printf "plan review: sound" | hive agent report --agent rev
reported to boss as record 1
exit 0
$ hive --json agent report --agent rev --file result.md
{"outcome":"ok","phase":"complete","effects":[{"resource":"counter","path":".hive/state/counters/inbox:boss.json","action":"updated"},{"resource":"inbox","path":".hive/state/inbox/boss/2.json","action":"created"},{"resource":"agent","path":".hive/state/agents/rev.json","action":"reported"}],"error":null,"result":{"inbox":"boss","record":{"seq":2,"role":"boss","kind":"report","task":"","agent":"rev","text":"a long result\n","path":"SCRATCH/result.md","created_at":"2026-10-22T00:06:05.4475308Z","read_at":"","nudged_at":"","nudged_generation":0}}}
exit 0
$ echo x | hive agent report --agent wk
hive: only a general agent can report: wk has the role worker
exit 1
$ head -c 4097 /dev/zero | tr "\0" a | hive agent report --agent rev
hive: report text is longer than 4096 bytes: use --file
exit 1
$ HERDR_PANE_ID=w9:p2 hive agent report --agent boss
hive: --agent boss is not the caller rev
exit 1
$ hive inbox --role boss --all --agent boss
1	report	rev	2026-10-22T00:06:05.430203774Z	plan review: sound
2	report	rev	2026-10-22T00:06:05.4475308Z	"a long result\n"	SCRATCH/result.md
exit 0
$ jq reported_at,finished,status rev.json
[true,false,"live"]
```

Log events of `agent report` (decision 42):

```
{"msg":"command start","level":"DEBUG","outcome":null,"error":null,"inbox":null,"agent_id":null}
{"msg":"agent report request","level":"DEBUG","outcome":null,"error":null,"inbox":null,"agent_id":null}
{"msg":"agent report text","level":"DEBUG","outcome":null,"error":null,"inbox":null,"agent_id":null}
{"msg":"agent report end","level":"DEBUG","outcome":"ok","error":null,"inbox":"boss","agent_id":"rev"}
{"msg":"command end","level":"DEBUG","outcome":"ok","error":null,"inbox":null,"agent_id":null}
{"msg":"command start","level":"DEBUG","outcome":null,"error":null,"inbox":null,"agent_id":null}
{"msg":"agent report request","level":"DEBUG","outcome":null,"error":null,"inbox":null,"agent_id":null}
{"msg":"agent report text","level":"DEBUG","outcome":null,"error":null,"inbox":null,"agent_id":null}
{"msg":"agent report end","level":"DEBUG","outcome":"ok","error":null,"inbox":"boss","agent_id":"rev"}
{"msg":"command end","level":"DEBUG","outcome":"ok","error":null,"inbox":null,"agent_id":null}
{"msg":"command start","level":"DEBUG","outcome":null,"error":null,"inbox":null,"agent_id":null}
{"msg":"agent report request","level":"DEBUG","outcome":null,"error":null,"inbox":null,"agent_id":null}
```
