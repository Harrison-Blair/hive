# VERIFIED — 102-selection-marker

Verifier: `t102-verifier-1`. Task commit: `5ca5078b`. Base: `origin/dev` at `d7fdcf2d`.

Verdict: **fail**. One "fix" row of the task 92 table has no test that fails when the fix breaks. All other items pass.

## Results

| ID | Check | Result |
| --- | --- | --- |
| V1 | R2 tests red on the base, green on the task | pass, with one gap (see the task 92 table: Detail strip name) |
| V2 | Live screens on a scratch home | pass |
| V3 | `git diff origin/dev --stat` | pass |
| V4 | Gate | pass |
| M1 | Each selection on the 8 screens: `▶` + honey bold, no background, no `colorText` | pass |
| M2 | Task 92 table: each fix row has a test that fails on `origin/dev` | **fail**: Detail strip name |
| M3 | Colour is never the only signal | pass |
| M4 | Design visual check and element table | pass, with notes |
| M5 | Frames exactly w × h at 120 × 40 and 200 × 57 | pass for the task fixtures; Chat at 120 × 40 with real chat data is 46 rows on base and task (pre-existing) |

## V1 — red and green

I copied `selection_marker_test.go` and the task `agents_test.go` into a detached worktree at `origin/dev` and ran the new tests.

- `TestScreenSelectionsUseHoneyMarkerAndDefaultForeground`: all 8 subtests fail on the base (missing honey-bold `▶` span, or `38;2;232;228;218` present). They pass on the task.
- `TestAgentBoardShowsDistinctStatusColorAndDimWorkingCount`: fails on the base (status dot foreground `""`, working count foreground `""`). Passes on the task.
- `TestAgentGraphBoardRendersTabsLegendHeaderAndDetails`: on the base it stops at the tab line. I changed that one `Fatalf` to `Errorf` and guarded the marker index in the base copy only. Then the base run reported: unselected tab honey, legend honey, working count `""`, needs-you count `""`, selected node border `""`, detail strip border `""`, detail label `""`. The final check `requireHoneyBoldText(view, "coordinator")` passed on the base, because the explorer row already shows `coordinator` in honey bold.
- `TestAgentsGraphPodTitleSharesTopBorder`, `TestAgentsGraphLongPodTitleKeepsNameTogetherBeforeStem`, `TestAgentsGraphColorsMatchBoard` (task versions): fail on the base, pass on the task.

Mutation on the task code: I replaced `styled.WriteString(selectedStyle.Render(words[0]))` with `styled.WriteString(words[0])` in `agents.go` (detail strip name not honey bold). `go test ./internal/tui` passed. No test detects a loss of this fix.

Mutation on the task code: I put the legend directly after the tabs (`tabLine += "  " + legend`). The golden frame tests `graph-120x40` and `graph-200x57` failed. The legend alignment is covered by the goldens.

## V2 — live

- Binary: `go build -o .hive/tmp/v102/hive .` at `5ca5078b`.
- Scratch home: `.hive/tmp/v102/home` with a copy of the real home records (agents, goals, projects, tasks, briefs, chat, friction) and a scratch `XDG_CONFIG_HOME`. Started with `env -u HERDR_ENV -u HERDR_PANE_ID -u HERDR_WORKSPACE_ID -u HERDR_SOCKET_PATH`. The TUI showed `not in herdr: no coordinator or supervisor started`.
- Tab `wTQ:t2`, pane `wTQ:p2` in my workspace `wTQ`. Last task pid `3728661`, stopped with ctrl+c ×2. I closed tab `wTQ:t2`. No `hive supervise` and no coordinator tab started. I did not open the Chat input and sent no chat.
- Pane size 207 × 57. A `stty` override to 200 × 57 conflicted with the herdr resize, so I discarded those captures and ran at the native size.
- I moved the selection once on each screen (Chat `n`, then `Down` on the others; Agents graph via `Tab` then `Down`) and captured ANSI. An independent SGR parser found:

| Screen | Selected spans (honey bold, start with `▶`) | Background SGR | `38;2;232;228;218` |
| --- | --- | --- | --- |
| Chat | `▶ 1 Chat`, `▶ noise: decisions only` | none | none |
| Agents (graph) | `▶ [ GRAPH ]`, explorer `▶ ? supervisor`, graph node `▶ ? supervisor` | none | none |
| Goals | `▶ 3 Goals`, `▶ hive · Hive TUI dashboard · active · …` | none | none |
| Queue | `▶ 4 Queue`, `▶ 1  105-verified-close …` | none | none |
| Tasks | `▶ 5 Tasks`, `▶ [active 2]`, `▶ project: all`, `▶ 117-changelog… …` | none | none |
| Friction | `▶ 6 Friction`, `▶ [open 16]`, `▶ 2026-10-20   A completed task …` | none | none |
| Health | `▶ 7 Health`, `▶ herdr pane` | none | none |
| Logs | `▶ 8 Logs`, `▶ level: info`, `▶ 10:46:29 ERROR task task error` | none | none |

Logs note: with the full 120 663-event log copy, the selected event row is never in the viewport (follow mode shows the bottom, the selection is row 0). The base binary does the same. With a 40-line log the selected row shows. This is pre-existing, not a task change.

## V3 — scope

`git diff origin/dev --stat`: `internal/tui/*.go` and tests, `internal/tui/testdata/agents-board/*.golden`, `testdata/ste-baseline.json` (one added line for the brief copy), and the task folder (brief copy, V1–V4, 8 screen PNGs). No board PNG is committed. Pass.

## V4 — gate

| Command | Result |
| --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | no output |
| `go mod tidy -diff` | no output |
| `go vet ./...` | exit 0 |
| `golangci-lint run ./...` | `0 issues.`, exit 0 |
| `go test -race ./...` | all packages `ok`, exit 0 |

## M1 — rules 2 and 3

- `shell.go`: `colorText` removed; `headerStyle` and `panelStyle` have no foreground. `grep colorText\|e8e4da internal/tui/*.go` (non-test): no match.
- Every `selectedStyle.Render` of a selection starts with `▶` (sidebar, Chat noise, Agents tabs/explorer/node/tree/table, Goals goal, Queue row, Tasks tab/filter/row, Friction tab/row, Health check, Logs filter/row). Two non-selection uses stay without `▶`: the Agents detail strip name and the Goals `◎` done-when stage. Both are correct.
- Model probe at 120 × 40 and 200 × 57 for all 8 screens: no SGR with `48;`, `40`–`47`, `100`–`107`, or `232;228;218`.

## M2 — task 92 table

| Element | Action | Code | Test that fails on `origin/dev` | Result |
| --- | --- | --- | --- | --- |
| Header working count | fix | green | `TestAgentGraphBoard…`: `""`, want green | pass |
| Header needs-you count | fix | honey | same test: `""`, want honey | pass |
| Tabs | fix | `▶ [ GRAPH ]` honey, others dim | same test + selection test | pass |
| Legend | fix | dim, right-aligned | same test (dim); goldens (alignment, mutation fails) | pass |
| Explorer status dots | fix | live `●` green | `TestAgentBoardShowsDistinct…` | pass |
| Explorer working count | fix | `· 1 working` dim | same test | pass |
| Explorer selected row | keep | `▶`, honey bold, no fill | — | pass |
| Pod title text | fix | full title when space | `TestAgentsGraphPodTitle…`, `…LongPodTitle…` | pass |
| Selected node border | fix | honey | `TestAgentGraphBoard…`, `TestAgentsGraphColorsMatchBoard` | pass |
| Detail strip border | fix | honey | `TestAgentGraphBoard…` | pass |
| Detail strip name | fix | honey bold in code | `requireHoneyBoldText(view, "coordinator")` passes on the base (it matches the explorer row) and passes with the fix removed | **fail** |
| Detail strip labels and values | fix | labels dim, values default | `TestAgentGraphBoard…` (label `""` on base) | pass |
| Key bar keys | keep | plain text, no fill | — | pass |

## M3 — colour is never the only signal

Plain text (SGR removed) shows a glyph or a word for each state on each changed screen: agents `●` live, `○` idle/stopped, `×` gone, `?` unknown, `✓` done; selection `▶`; Health `✓`, `!`, `✕`; header `supervisor stale`/`fresh`; task status words; friction `status: Open`; log level words. `TestScreenSelectionsUseHoneyMarkerAndDefaultForeground` and `TestAgentBoardShowsDistinct…` read these from plain text. I found no state that only colour shows. Pass.

## M4 — design visual check

Boards rendered with headless Firefox to `/home/penguin/source/hive/projects/hive/worktrees/102-selection-marker/.hive/tmp/v102/boards/` (not committed; removed after this record). Screens rendered from my own ANSI captures. I compared each board with my screen PNG.

The element table in `V2.md` agrees with my screens for selection, marker, colour, and "state without colour", with these notes:

- Agents header shows `working unknown  idle unknown` (dim) in a scratch home outside herdr, so the green count is visible only in the test. Cause: environment.
- Chat `Transcript` is marked `matches`. In my capture, long messages break mid-word (`meant for t` / `he screen`) and have no left bar as on the board. The base binary does the same. Pre-existing; not a change of this task, but the row should say `differs`.
- Goals, Tasks, Queue keep their text layout where the boards draw chips and status colours. These are earlier screen decisions; this task changes only selection and text colour.
- Key chips with a fill on each board: overridden by rule 1.

None of these is a fail for this task.

## M5 — frame size

Model probe on the task code: all 8 screens with the test fixtures are exactly 120 × 40 and 200 × 57. Agents goldens: 40 × 120 and 57 × 200.

With the real chat copy, Chat at 120 × 40 renders 46 rows (the conversation panel grows with wrapped messages). The base code renders the same 46 rows. Pre-existing; not caused by this task.

## Repair needed

Add a test for the Detail strip name that reads the detail strip line (not the whole view) and checks that the agent name is honey bold. See it fail on `origin/dev` and with the `selectedStyle.Render(words[0])` call removed.
