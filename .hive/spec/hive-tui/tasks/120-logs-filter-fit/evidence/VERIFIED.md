# VERIFIED: 120-logs-filter-fit

Verifier: Claude Opus 5.5 (claude-opus-5-5), workspace `wV6`. Task commit under test: `53596c93` (fix in `50fe51d9`).

| Item | Result | Observation |
| --- | --- | --- |
| R1 | pass | `internal/tui/shell.go:71` wraps at `width-4-lipgloss.Width("▶ ")`. The live ANSI capture shows `ESC[1m ESC[38;5;215m ▶ level: info ESC[0m`: the `▶ ` marker is kept with the bold accent `selectedStyle`. |
| R2 | pass | `TestLogsLongFilterKeepsPanelFrameAtTerminalSize` checks 120x40 and 200x57 with a wrapping filter: line count = h, each line = w cells, and line h-3 is the literal Logs panel bottom border. |
| V1 | pass | Green: `go test -race -count=1 -run TestLogsLongFilterKeepsPanelFrameAtTerminalSize ./internal/tui/` → `ok`. Red: I reverted the fix line to the base `wrapFilterLine(m.statusLine(), width-4)` (the only production change from `dev`). Both subtests fail: `last line of the Logs panel is "╰────────────────────────────╯│ ..."` (a body row in place of the bottom border). I restored the file with `git checkout`; the tree is clean. |
| V2 | pass | See "V2 live run". |
| V3 | pass | `git diff origin/dev...HEAD --stat`: `brief.md`, `evidence/{120x40.png,200x57.png,V1.md,V2.md,V3.md,V4.md}`, `internal/tui/logs_screen_test.go`, `internal/tui/shell.go`, `testdata/ste-baseline.json` (one added line for the brief copy at 0). All are in the source table or evidence. The two PNGs are screen captures, not board captures. |
| V4 | pass | `gofmt -l` printed nothing. `go mod tidy -diff` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` → `0 issues.`, exit 0. `go test -race ./...` exit 0, all packages `ok`. |

## V2 live run

- Binary: `go build -o <scratchpad>/hive .` from the worktree, exit 0.
- Herdr tab `wV6:t2`, pane `wV6:p2`, created by me in my own workspace `wV6`. I did not use `wBC`.
- The pane ran a script: `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH -u TMUX HIVE_HOME=<scratch home> XDG_CONFIG_HOME=<scratch xdg>`. It started a private tmux server `-L v120` (server pid 63642) with session `logs120` (`-x 120 -y 40`, hive pid 63643) and session `logs200` (`-x 200 -y 57`, hive pid 63646).
- `tmux display-message`: `logs120 120x40 pane 120x40`, `logs200 200x57 pane 200x57`.
- In each session: `8` (Logs), `/`, the filter text, `Enter`. The filter for 120x40 was `filter filter`. The filter for 200x57 was ten `filter` words. In both sessions, the filter line wraps to two lines inside the panel.
- `tmux capture-pane -p`: 40 rows and 57 rows. The Logs panel bottom border `╰…╯╰…╯` is on row 37 of 40 and on row 54 of 57. Row h-2 and row h-1 are the status lines, and row h is the footer. All rows with the box are 120 or 200 cells wide. Row 3 measures 1 cell more when the tab cells of the capture are expanded. This is a capture artifact: the pane is 120 or 200 cells wide.
- Cleanup: `tmux -L v120 kill-server` (no pid of 63642, 63643, or 63646 remains; `no server running on .../v120`), then `herdr tab close wV6:t2` (only `wV6:t1` remains). Scratch files stay in my session scratchpad, outside the repository.

## Visual check (filter line and panel border only)

Screen captures: ANSI → HTML `<pre>` → Firefox headless PNG. The board `/home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Logs.dc.html` was captured again with Firefox at 1220x780. No PNG is committed.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Filter line | `▶ level: info … text: filter filter …` | overridden | The board shows filter chips (`level ≥ info`, `component: all`, …). R1 requires to keep the `▶ ` marker and its style. The wrapped text stays inside the panel at both sizes. |
| Logs panel bottom border | `╰…╯` | matches | The rounded bottom border shows at 120x40 and 200x57, above the status lines and the footer. |
| Header status | `⬢ hive · 0 projects` | differs | not in the scope of this task |
| Screen navigation | `▶ 8 Logs` | differs | not in the scope of this task |
| Event table and rows | `no events today` | differs | not in the scope of this task (scratch home has no events) |
| Event detail panel | none | differs | not in the scope of this task (detail closed) |
| Footer hints | `1–8 screens · l level …` | differs | not in the scope of this task |

The selected state of the filter line and of the navigation has the `▶` glyph in plain text. Colour only adds to it.

## Verdict

pass
