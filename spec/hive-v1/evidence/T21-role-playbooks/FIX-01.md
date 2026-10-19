# T21-role-playbooks FIX-01

| Item | Value |
| --- | --- |
| Task | T21-role-playbooks |
| Item | FIX-01 |
| Proves | AC1, AC4; findings B1 and B2 of `t21-verify-01.md` |
| Base | `1b986c5` |
| Source commit | `d379a58` |
| Date | 2026-10-19 |
| Agent | worker, pi, opencode-go/deepseek-v4.1-flash:high |

## Source map

| Path | Change |
| --- | --- |
| `skills_test.go` | Two new `TestSkillCommands` subtests, a new `TestCoordinatorPreference`, the prose-sentence skip, and the root rejection in `resolveCommandPath`. The unused `runnable` helper goes. |
| `.agents/skills/hive-coordinator/SKILL.md` | The preference procedure writes `overrides/NAME.md`. The memory command goes. |

The change removes the `runnable` escape from `resolveCommandPath`. An unknown
token now fails when the current command has children. A leaf command still
takes the token as a positional argument. The two templates start with a prose
line that starts with the word `hive`, so `unknownSkillCommands` skips a line
that ends with sentence punctuation.

## B1 red run

The root is runnable because bare `hive` starts the TUI. The old code treats
an unknown root command as a positional argument.

```sh
go test . -run 'TestSkillCommands/an_absent_root_command' -v
```

```text
commit 1b986c5
$ go test . -run 'TestSkillCommands/an_absent_root_command' -v
=== RUN   TestSkillCommands
=== RUN   TestSkillCommands/an_absent_root_command_is_unknown
    skills_test.go:65: got 0 problems []; want 1
--- FAIL: TestSkillCommands (0.00s)
    --- FAIL: TestSkillCommands/an_absent_root_command_is_unknown (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive	0.003s
FAIL
exit 1
```

## B1 green run

The same command with the fix.

```sh
go test . -run 'TestSkillCommands/(an_absent_root_command|a_valid_positional_argument)' -count=1 -v
```

```text
commit d379a58
$ go test . -run 'TestSkillCommands/(an_absent_root_command|a_valid_positional_argument)' -count=1 -v
=== RUN   TestSkillCommands
=== RUN   TestSkillCommands/an_absent_root_command_is_unknown
=== RUN   TestSkillCommands/a_valid_positional_argument_is_accepted
--- PASS: TestSkillCommands (0.00s)
    --- PASS: TestSkillCommands/an_absent_root_command_is_unknown (0.00s)
    --- PASS: TestSkillCommands/a_valid_positional_argument_is_accepted (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive	0.004s
exit 0
```

## B2 red run

The preference procedure writes a memory. The new test fails on the assertion.

```sh
go test . -run TestCoordinatorPreference -v
```

```text
commit 1b986c5
$ go test . -run TestCoordinatorPreference -v
=== RUN   TestCoordinatorPreference
    skills_test.go:93: preference procedure does not name overrides/:
        
        
        The user states a preference. Write one memory at home scope. The command reads
        the memory text from standard input.
        
        ```sh
        hive memory write --scope home NAME
        ```
    skills_test.go:96: preference procedure writes a memory:
        
        
        The user states a preference. Write one memory at home scope. The command reads
        the memory text from standard input.
        
        ```sh
        hive memory write --scope home NAME
        ```
--- FAIL: TestCoordinatorPreference (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive	0.003s
FAIL
exit 1
```

## B2 green run

The same command with the fix.

```sh
go test . -run TestCoordinatorPreference -count=1 -v
```

```text
commit d379a58
$ go test . -run TestCoordinatorPreference -count=1 -v
=== RUN   TestCoordinatorPreference
--- PASS: TestCoordinatorPreference (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive	0.004s
exit 0
```

## Full item green run

The V1 and V5 command, at the fix commit. `TestCoordinatorPreference` joins the
`TestSkillCommands` and STE run.

```sh
go test . -run 'TestSkillCommands|TestCoordinatorPreference|TestSTE' -count=1 -v
```

```text
commit d379a58
$ go test . -run 'TestSkillCommands|TestCoordinatorPreference|TestSTE' -count=1 -v
=== RUN   TestSkillCommands
=== RUN   TestSkillCommands/fixture_reports_each_unknown_command
=== RUN   TestSkillCommands/each_real_file_names_a_command
=== RUN   TestSkillCommands/persistent_flags_are_skipped
=== RUN   TestSkillCommands/an_absent_root_command_is_unknown
=== RUN   TestSkillCommands/a_valid_positional_argument_is_accepted
--- PASS: TestSkillCommands (0.00s)
    --- PASS: TestSkillCommands/fixture_reports_each_unknown_command (0.00s)
    --- PASS: TestSkillCommands/each_real_file_names_a_command (0.00s)
    --- PASS: TestSkillCommands/persistent_flags_are_skipped (0.00s)
    --- PASS: TestSkillCommands/an_absent_root_command_is_unknown (0.00s)
    --- PASS: TestSkillCommands/a_valid_positional_argument_is_accepted (0.00s)
=== RUN   TestCoordinatorPreference
--- PASS: TestCoordinatorPreference (0.00s)
=== RUN   TestSTESemicolon
=== RUN   TestSTESemicolon/pass
=== RUN   TestSTESemicolon/fail
--- PASS: TestSTESemicolon (0.00s)
    --- PASS: TestSTESemicolon/pass (0.00s)
    --- PASS: TestSTESemicolon/fail (0.00s)
=== RUN   TestSTEPassiveVoice
=== RUN   TestSTEPassiveVoice/pass
=== RUN   TestSTEPassiveVoice/fail
--- PASS: TestSTEPassiveVoice (0.00s)
    --- PASS: TestSTEPassiveVoice/pass (0.00s)
    --- PASS: TestSTEPassiveVoice/fail (0.00s)
=== RUN   TestSTEPresentPerfect
=== RUN   TestSTEPresentPerfect/pass
=== RUN   TestSTEPresentPerfect/fail
=== RUN   TestSTEPresentPerfect/modal_perfect_stays_a_hedge
--- PASS: TestSTEPresentPerfect (0.00s)
    --- PASS: TestSTEPresentPerfect/pass (0.00s)
    --- PASS: TestSTEPresentPerfect/fail (0.00s)
    --- PASS: TestSTEPresentPerfect/modal_perfect_stays_a_hedge (0.00s)
=== RUN   TestSTELongSentence
=== RUN   TestSTELongSentence/pass
=== RUN   TestSTELongSentence/fail
--- PASS: TestSTELongSentence (0.00s)
    --- PASS: TestSTELongSentence/pass (0.00s)
    --- PASS: TestSTELongSentence/fail (0.00s)
=== RUN   TestSTEPhrasalVerb
=== RUN   TestSTEPhrasalVerb/pass
=== RUN   TestSTEPhrasalVerb/fail
--- PASS: TestSTEPhrasalVerb (0.00s)
    --- PASS: TestSTEPhrasalVerb/pass (0.00s)
    --- PASS: TestSTEPhrasalVerb/fail (0.00s)
=== RUN   TestSTEGlossaryTerm
=== RUN   TestSTEGlossaryTerm/pass
=== RUN   TestSTEGlossaryTerm/fail
=== RUN   TestSTEGlossaryTerm/inflections_count
=== RUN   TestSTEGlossaryTerm/silent_e_drops_before_-ing
=== RUN   TestSTEGlossaryTerm/silent_e_drops_before_-ing,_second_word
=== RUN   TestSTEGlossaryTerm/consonant_y_becomes_-ies
=== RUN   TestSTEGlossaryTerm/consonant_y_becomes_-ied
=== RUN   TestSTEGlossaryTerm/regular_-ing_survives_the_y_rule
=== RUN   TestSTEGlossaryTerm/multi-word_term
=== RUN   TestSTEGlossaryTerm/word_boundary
--- PASS: TestSTEGlossaryTerm (0.00s)
    --- PASS: TestSTEGlossaryTerm/pass (0.00s)
    --- PASS: TestSTEGlossaryTerm/fail (0.00s)
    --- PASS: TestSTEGlossaryTerm/inflections_count (0.00s)
    --- PASS: TestSTEGlossaryTerm/silent_e_drops_before_-ing (0.00s)
    --- PASS: TestSTEGlossaryTerm/silent_e_drops_before_-ing,_second_word (0.00s)
    --- PASS: TestSTEGlossaryTerm/consonant_y_becomes_-ies (0.00s)
    --- PASS: TestSTEGlossaryTerm/consonant_y_becomes_-ied (0.00s)
    --- PASS: TestSTEGlossaryTerm/regular_-ing_survives_the_y_rule (0.00s)
    --- PASS: TestSTEGlossaryTerm/multi-word_term (0.00s)
    --- PASS: TestSTEGlossaryTerm/word_boundary (0.00s)
=== RUN   TestSTECodeSpanIsolation
=== RUN   TestSTECodeSpanIsolation/live_rules
=== RUN   TestSTECodeSpanIsolation/code_span_holding_a_pipe_and_a_semicolon
=== RUN   TestSTECodeSpanIsolation/code_span_does_not_join_the_prose_around_it
--- PASS: TestSTECodeSpanIsolation (0.00s)
    --- PASS: TestSTECodeSpanIsolation/live_rules (0.00s)
    --- PASS: TestSTECodeSpanIsolation/code_span_holding_a_pipe_and_a_semicolon (0.00s)
    --- PASS: TestSTECodeSpanIsolation/code_span_does_not_join_the_prose_around_it (0.00s)
=== RUN   TestSTESegmentsSkipCode
=== RUN   TestSTESegmentsSkipCode/fenced_code
=== RUN   TestSTESegmentsSkipCode/tilde_fence
=== RUN   TestSTESegmentsSkipCode/inline_code_span_masked_to_its_own_width
=== RUN   TestSTESegmentsSkipCode/table_row_of_only_code
=== RUN   TestSTESegmentsSkipCode/table_cells_split
=== RUN   TestSTESegmentsSkipCode/pipe_inside_a_code_span_does_not_split_a_cell
--- PASS: TestSTESegmentsSkipCode (0.00s)
    --- PASS: TestSTESegmentsSkipCode/fenced_code (0.00s)
    --- PASS: TestSTESegmentsSkipCode/tilde_fence (0.00s)
    --- PASS: TestSTESegmentsSkipCode/inline_code_span_masked_to_its_own_width (0.00s)
    --- PASS: TestSTESegmentsSkipCode/table_row_of_only_code (0.00s)
    --- PASS: TestSTESegmentsSkipCode/table_cells_split (0.00s)
    --- PASS: TestSTESegmentsSkipCode/pipe_inside_a_code_span_does_not_split_a_cell (0.00s)
=== RUN   TestSTECompare
=== RUN   TestSTECompare/equal
=== RUN   TestSTECompare/above_baseline
=== RUN   TestSTECompare/below_baseline
=== RUN   TestSTECompare/missing_entry
=== RUN   TestSTECompare/stale_entry
--- PASS: TestSTECompare (0.00s)
    --- PASS: TestSTECompare/equal (0.00s)
    --- PASS: TestSTECompare/above_baseline (0.00s)
    --- PASS: TestSTECompare/below_baseline (0.00s)
    --- PASS: TestSTECompare/missing_entry (0.00s)
    --- PASS: TestSTECompare/stale_entry (0.00s)
=== RUN   TestSTERatchet
--- PASS: TestSTERatchet (0.41s)
PASS
ok  	github.com/Harrison-Blair/hive	0.417s
exit 0
```

The `TestSTERatchet` run covers the coordinator file with a baseline of 0.

## B1 live check

Append the bad command to a real skill. The check fails with the file, the
line, and the command. The restore follows.

```sh
cp .agents/skills/hive-coordinator/SKILL.md .hive/tmp/t21-fix/SKILL.md.orig
printf '\nhive absent-command\n' >> .agents/skills/hive-coordinator/SKILL.md
go test . -run 'TestSkillCommands/each_real_file_names_a_command' -count=1 -v; echo "exit $?"
cp .hive/tmp/t21-fix/SKILL.md.orig .agents/skills/hive-coordinator/SKILL.md
git status --porcelain=v1
```

```text
commit d379a58
$ go test . -run 'TestSkillCommands/each_real_file_names_a_command' -count=1 -v
=== RUN   TestSkillCommands
=== RUN   TestSkillCommands/each_real_file_names_a_command
    skills_test.go:50: .agents/skills/hive-coordinator/SKILL.md:89: unknown command: absent-command
--- FAIL: TestSkillCommands (0.00s)
    --- FAIL: TestSkillCommands/each_real_file_names_a_command (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive	0.004s
FAIL
exit 1
$ git status --porcelain=v1
 M spec/hive-v1/evidence/T21-role-playbooks/V1.md
 M spec/hive-v1/evidence/T21-role-playbooks/V5.md
 M spec/hive-v1/evidence/T21-role-playbooks/V6.md
?? spec/hive-v1/evidence/T21-role-playbooks/FIX-01.md
```

The cleanup left the skill at its committed bytes. The skill file is absent
from the status output. The status lists only the evidence paths of this pass.

## Result

The check rejects the absent root command `hive absent-command`, still accepts
the valid positional argument `hive project show repo-one`, and keeps each
existing test. The coordinator preference procedure names `overrides/NAME.md`
and writes no memory.
