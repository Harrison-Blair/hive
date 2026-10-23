# 136-golden-frames-spec VERIFIED

| Item | Value |
| --- | --- |
| Task | `136-golden-frames-spec` |
| Commit | `22cd5027` |
| Date | 2026-10-23 |
| Agent | verifier, pi, gpt-6.1-sol, `t136-verifier-1` |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | AC1 | pass | All eight byte comparisons pass. The old folder is absent. The ignore check prints nothing and exits 1. |
| V2 | AC2 | pass | The old ratchet paths fail through a temporary Go overlay. The task paths pass. The baseline has nine feature keys, all 0. |
| V3 | AC3 | pass | No reference remains outside the task folder. |
| V4 | AC1 | pass | Eight `R100` renames. Only the two named files, brief, and evidence also change. |
| V5 | AC2 | pass | All five gate commands pass. The format and module checks print nothing. |

The task moves documents and changes the ratchet paths. It changes no user interface. A visual check does not apply.

## Runs

Each command ran from the worktree root. The verifier did not change source files or worker evidence.

### V1 Move

```sh
set -o pipefail
for path in README.md screens/agents.md screens/chat.md screens/friction.md screens/health.md screens/help.md screens/logs.md screens/work.md; do
  printf '%s\n' "$path"
  git show "0effc449:redesign-spec/$path" | cmp - ".hive/spec/golden-frames/$path"
  echo "exit $?"
done
test ! -e redesign-spec
echo "exit $?"
git check-ignore -v .hive/spec/golden-frames/README.md .hive/spec/golden-frames/screens/*.md
echo "exit $?"
```

```text
README.md
exit 0
screens/agents.md
exit 0
screens/chat.md
exit 0
screens/friction.md
exit 0
screens/health.md
exit 0
screens/help.md
exit 0
screens/logs.md
exit 0
screens/work.md
exit 0
exit 0
exit 1
```

### V2 Ratchet

#### Red run

The overlay supplies `ste_test.go` from `0effc449`. That file differs from the task version only at the two spec paths. The assertion fails because the old path matches no file. Compilation succeeds. No worktree source file changes.

Setup:

```sh
scratch=$(mktemp -d "$PWD/.hive/tmp/verify-136.XXXXXX")
trap 'rm -rf "$scratch"' EXIT
git show 0effc449:ste_test.go > "$scratch/old-ste_test.go"
python3 - "$PWD/ste_test.go" "$scratch/old-ste_test.go" "$scratch/overlay.json" <<'PY'
import json
import sys
from pathlib import Path
Path(sys.argv[3]).write_text(json.dumps({"Replace": {sys.argv[1]: sys.argv[2]}}))
PY
```

```sh
GOFLAGS="${GOFLAGS:-} -overlay=$scratch/overlay.json" go test ./ -run 'STE|Gitignore' -count=1 -v
echo "exit $?"
```

```text
commit 22cd5027
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
    ste_test.go:533: expand steFiles: pattern redesign-spec/README.md matches no file
--- FAIL: TestSTERatchet (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive	0.009s
FAIL
exit 1
```

#### Green run

```sh
go test ./ -run 'STE|Gitignore' -count=1 -v
echo "exit $?"
grep -c golden-frames testdata/ste-baseline.json
echo "exit $?"
```

```text
commit 22cd5027
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
--- PASS: TestSTERatchet (0.58s)
PASS
ok  	github.com/Harrison-Blair/hive	0.584s
exit 0
9
exit 0
```

The V2 filter selects the STE tests. V5 covers the ignore-rule tests. A separate JSON check found exactly the eight expected spec keys and the brief key, all 0. No old path key remains. The diff preserves the list format and changes only the two `steFiles` paths.

### V3 References

```sh
git grep -n redesign-spec -- ':!.hive/spec/golden-frames/tasks/'
echo "exit $?"
```

```text
exit 1
```

### V4 Diff scope

```sh
git merge dev
echo "exit $?"
git diff --stat -M dev...HEAD
echo "exit $?"
```

```text
Already up to date.
exit 0
 .../spec/golden-frames}/README.md                  |   0
 .../spec/golden-frames}/screens/agents.md          |   0
 .../spec/golden-frames}/screens/chat.md            |   0
 .../spec/golden-frames}/screens/friction.md        |   0
 .../spec/golden-frames}/screens/health.md          |   0
 .../spec/golden-frames}/screens/help.md            |   0
 .../spec/golden-frames}/screens/logs.md            |   0
 .../spec/golden-frames}/screens/work.md            |   0
 .../tasks/136-golden-frames-spec/brief.md          | 113 +++++++++++++++++++++
 .../tasks/136-golden-frames-spec/evidence/V1.md    |  26 +++++
 .../tasks/136-golden-frames-spec/evidence/V2.md    |  21 ++++
 .../tasks/136-golden-frames-spec/evidence/V3.md    |   8 ++
 .../tasks/136-golden-frames-spec/evidence/V4.md    |  20 ++++
 .../tasks/136-golden-frames-spec/evidence/V5.md    |  96 +++++++++++++++++
 ste_test.go                                        |   4 +-
 testdata/ste-baseline.json                         |  19 ++--
 16 files changed, 296 insertions(+), 11 deletions(-)
exit 0
```

The stat precedes this verifier file. This file adds only evidence. The independent `git diff --name-status -M dev...HEAD` check reports `R100` for each of the eight moves. `.gitignore` does not change.

### V5 Gate

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
echo "exit $?"
go mod tidy -diff
echo "exit $?"
go vet ./...
echo "exit $?"
golangci-lint run ./...
echo "exit $?"
go test -race ./...
echo "exit $?"
```

```text
exit 0
exit 0
exit 0
0 issues.
exit 0
ok  	github.com/Harrison-Blair/hive	(cached)
ok  	github.com/Harrison-Blair/hive/cmd	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/report	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	(cached)
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/ask	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	(cached)
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	(cached)
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	(cached)
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	(cached)
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	(cached)
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	(cached)
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	(cached)
ok  	github.com/Harrison-Blair/hive/internal/ask	(cached)
ok  	github.com/Harrison-Blair/hive/internal/boardexport	(cached)
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	(cached)
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui	(cached)
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	(cached)
exit 0
```

### Cleanup

The shell trap removed the temporary overlay folder at process exit. The final check printed nothing.

```sh
find .hive/tmp -maxdepth 1 -name 'verify-136.*' -print
echo "exit $?"
```

```text
exit 0
```
