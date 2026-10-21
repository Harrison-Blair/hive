# Equal column width in the Agents graph

| Item | Value |
| --- | --- |
| Feature | `hive-tui`, restart goal, wave 3 |
| Tier | decent |
| Depends on | 3-tui-restart. Base: `dev` after it lands |
| Blocks | none |
| Grant | `grants.md 2026-10-18 hive delivery on dev` |
| Evidence | `.hive/spec/hive-tui/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

The Agents graph has one column for each group: home, and each project. Now
each column has the width of its own pod, and a card can be cut. The user
approved this change on 2026-10-21: each column is as wide as the widest
column, so that no card is cut.

Read `.hive/spec/hive-tui/README.md` decision 1 first. When the code does not
agree with this brief, stop and report it with
`hive inbox send --to manager-hive TEXT --agent NAME`.

## Known facts

- `graphLayout` in `internal/tui/agents.go` sets the width of each column
  from the `podWidth` that `graphPod` returns for that group.
- The edges and the card positions use the column offset and the card width
  of each pod.

## Design

There is no new board. Keep the Agents graph look. Record a screenshot at
120x40 and at 200x57 in the evidence, before and after the change, with at
least two groups of different widths.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `internal/tui/agents.go` and its tests | AC1, AC2. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/hive-tui/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files.

## Requirements

### AC1 Equal columns

- Each column of the graph has the width of the widest column.
- Each card in each column shows its full text. No card is cut.
- Each card stays centred in its column, as now.

### AC2 Edges stay correct

- Each edge, solid or dashed, still joins the correct cards after the change.
- The selection marker and attach still act on the selected card.
- When the graph is wider than the screen, the existing behavior for a wide
  graph does not change. Record that behavior in the evidence.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | Each column has the width of the widest column, and no card is cut | V1, V3 |
| AC2 | Edges, selection, and the wide-graph behavior stay correct | V1, V3 |

## Verification items

### V1 Tests

- Proves: AC1, AC2.
- Command: run each new test against the base code, then against the task
  code. Test at the model with `Update` and `View`.
- Expected: each new test fails on the base for the correct reason and passes
  on the task branch. Cases: two groups with different widths, one group, and
  a card whose text is wider than the other cards.

### V2 Brief copy and ratchet

- Proves: AC1.
- Command: `go test ./ -run STE`
- Expected: pass.

### V3 Screenshots

- Proves: AC1, AC2.
- Command: run the task binary with a temp `HIVE_HOME` that has fixture
  records in at least two groups. Take screenshots at 120x40 and 200x57.
- Expected: the columns have equal width, and no card is cut. Close each
  workspace and pane that the check made.

### V4 Diff scope

- Proves: AC1, AC2.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch.
- Expected: only the files of the source table and the evidence changed.

### V5 Gate

- Proves: AC1, AC2.
- Command: the gate of `AGENTS.md` from the worktree root.
- Expected: `gofmt -l` and `go mod tidy -diff` print nothing. The other
  commands exit 0.

## Rules

- Each Go change is test-first. See each new test fail for the correct reason.
- Live checks never change the real home.

## Out of scope

Other changes to the Agents screen. `main`.

## Changes

- 2026-10-21: V4 uses `git diff dev...HEAD`, because `dev` gets friction
  commits after the branch starts.
