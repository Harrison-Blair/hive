# Screenshots of the signed-off screens

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, wave 8 (Pull request) |
| Tier | strong |
| Depends on | 33-screen-agents (landed). Base: `dev` |
| Grant | `grants.md 2026-10-21 tui-v2 delivery on dev and PR 47 update` |
| Evidence | `.hive/spec/tui-v2/tasks/<task-id>/evidence/` in the worktree. `VERIFIED.md` goes in the same folder. |

## Overview

On 2026-10-21 the user signed off each of the 20 screens (decision 47). The
live captures of `dev` at `2adaf0ab` are in
`/home/penguin/source/hive/.hive/tmp/reports/tui-v2-signoff/`. Decision 52
gives the screenshots of the pull request. This task renders each signed-off
capture to a PNG file and commits the files. The manager then updates pull
request #47.

## Design

- `/home/penguin/source/hive/.hive/tmp/reports/tui-v2-signoff/README.md`
  lists the 20 screens, their sizes, and the files `NAME.ansi`.
- Decisions 13 and 52 of `.hive/spec/tui-v2/README.md`.

## Source files

| Path | Purpose of the edit |
| --- | --- |
| `.hive/spec/tui-v2/screenshots/NAME.png` | AC1. One file for each of the 20 screens. |
| `.gitignore` | Only when Git ignores the new folder. Use the narrowest rule. |
| `testdata/ste-baseline.json`, `ste_test.go` | To add the brief copy with 0. |
| `.hive/spec/tui-v2/tasks/<task-id>/brief.md` | The brief copy. Commit it. |

Do not edit other files. Do not add `freeze` to `go.mod`.

## Requirements

### AC1 The PNG files (decision 52)

- Install Charm `freeze` v0.2.2 with
  `go install github.com/charmbracelet/freeze@v0.2.2`.
- The font is JetBrains Mono. When the host has no such font, download the
  official TTF file of JetBrains Mono into `.hive/tmp/` of the worktree and
  give it to `freeze` with its font file option. Record the source URL and
  the version of the font.
- Render each `NAME.ansi` of the sign-off folder to
  `.hive/spec/tui-v2/screenshots/NAME.png`. Use one command for all files.
  The command sets the font, a dark theme, the padding, and the window size
  of the board. It does not wrap a line. Each PNG shows exactly the cells of
  its board size.
- Record the exact command in the evidence.

### AC2 Symbol check

- The font must draw each symbol of decision 13. Make one test page with
  each symbol and render it with the same command. Look at the PNG. A symbol
  that shows as an empty box or a fallback glyph is missing.
- The user accepted one missing symbol on 2026-10-21: JetBrains Mono has no
  `◐`, so the PNG shows an empty box for it. Each other symbol must show.
- Look at each PNG. Compare it with its capture text. Stop and report to the
  manager when another symbol is missing or a line wraps.

## Acceptance criteria

| ID | Criterion | Proven by |
| --- | --- | --- |
| AC1 | 20 PNG files with the board names, made by one recorded command | V1 |
| AC2 | No missing symbol other than `◐`, and no wrapped line | V1 |

## Verification items

### V1 Screenshots

- Proves: AC1, AC2.
- Command: run the recorded command again into a scratch folder. Compare the
  PNG sizes with the committed files. Look at the symbol page and at each
  PNG next to its capture text.
- Expected: 20 files with the board names. Each image shows the full board
  with no wrap and no missing symbol other than `◐`. The new run gives the same image size
  for each file.

### V-gate Diff scope and gate

- Proves: each criterion.
- Command: `git diff dev...HEAD --stat`, after you merge `dev` into the task
  branch. Then the full gate of `AGENTS.md` from the worktree root.
- Expected: only the files of the source table and the evidence changed.
  `gofmt -l` and `go mod tidy -diff` print nothing. The other commands exit 0.
  Put the full gate output in the evidence.

## Rules

- Read `.hive/spec/tui-v2/README.md` first. When the code, a board, or the
  spec does not agree with this brief, stop and report it with
  `hive inbox send --to manager-hive TEXT --agent NAME`.
- Each Go change is test-first. The brief names the first failing test. See
  it fail for the correct reason.
- Do not run bare `hive` in herdr without `--no-startup` or
  `HIVE_NO_STARTUP=1`. Refer to the friction entry
  `scratch-hive-starts-real-agents`.
- A live check never changes the real home. Close each workspace, tab, and
  pane that the check makes.
- In a cleanup, do not use `rm -rf` on a shell variable without `"${VAR:?}"`.
  Push in a command of its own.

## Out of scope

The pull request. The `Components` sheet, which is not a screen.

## Changes

- 2026-10-21: the user accepts the empty box for `◐` (JetBrains Mono v2.304).
