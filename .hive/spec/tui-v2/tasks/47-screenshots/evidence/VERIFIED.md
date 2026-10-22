# Verification: 47-screenshots

Verifier: `t47-verifier-1`. Date: 2026-10-22.
Task commit: `1768e5d466037c5304998709d36c64e76cf40b5d`.
Base: `dev` at `2adaf0abc2dbf17014574fe158f949ec8058f14a`.

## Verdict

**pass** for the requirements of this brief.

| Item | Result | Observation |
| --- | --- | --- |
| V1 | pass | The new run gives 20 PNG files. Each file has the correct board name and is byte-identical to its committed file. Each screen was inspected against its capture text. No added wrap, clipped board row, or missing symbol other than the accepted `◐` box was found. |
| V-gate | pass | `git merge dev` said `Already up to date.` The diff has only the 20 PNG files, the brief, evidence, and the brief baseline row with 0. All five gate commands exit 0. |

No Go source changed. This task has no test item that needs a red run.

## V1 run

The design authority for this task is the signed-off capture set at
`/home/penguin/source/hive/.hive/tmp/reports/tui-v2-signoff/README.md`.
The verifier read that file and the feature README. These are screenshots
of the approved captures. No app screen or HTML board changed in this task.
The verifier rendered each capture again, then read each new PNG and its
capture text with SGR codes removed. This check did not use a contact sheet.

`freeze --version` gives `freeze version v0.2.2`. The installed Go build
record also gives `github.com/charmbracelet/freeze v0.2.2`.
The font is JetBrains Mono v2.304. Its source URL is
https://github.com/JetBrains/JetBrainsMono/releases/download/v2.304/JetBrainsMono-2.304.zip.
The verifier checked the TTF SHA-256:
`a0bf60ef0f83c5ed4d7a75d45838548b1f6873372dfac88f71804491898d138f`.

The verifier read `.hive/tmp/render.sh` and checked that it is the exact
command recorded in `V1.md`. The rerun was:

```sh
OUT=.hive/tmp/verify47/png bash .hive/tmp/render.sh
```

The command in that script is:

```sh
S=${S:-/home/penguin/source/hive/.hive/tmp/reports/tui-v2-signoff}; OUT=${OUT:-.hive/spec/tui-v2/screenshots}; FONT=.hive/tmp/font/fonts/ttf/JetBrainsMono-Regular.ttf; mkdir -p "$OUT"; for a in "$S"/*.ansi; do n=$(basename "$a" .ansi); read -r _ _ c r < "$S/$n.grid"; freeze -c base --font.file "$FONT" --font.size 28 --font.ligatures=false --line-height 1.32 --theme charm --background '#171717' --padding 40 --window=false --width "$(echo "$c*16.8+80" | bc)" --height "$(echo "$r*36.96+88.4" | bc)" --output "$OUT/$n.png" < "$a" || exit 1; done
```

### Screen review

Each row covers the frame, content, borders, and footer of that screen.
Each rerun PNG is byte-identical to the committed PNG. The dimensions below
are pixels. The comparison uses the capture text as the board content.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| AgentDetails, 2096×1566 | `live · idle`, `ACTIONS`, `[details]` | matches | Full popup, dimmed frame, recent events, and footer. |
| AgentsMany, 2096×1566 | `working`, `idle`, `stopped`, heavy selected border | matches | Four visible pods, project range 1–4 of 5, scroll mark, legend. |
| Chat, 2096×1566 | `you`, `coordinator`, `[input]`, `█` | matches | Table, code, message input, and footer. The `UI.` row is present in the capture. |
| Compact, 1424×975 | `[Tasks]`, `landed`, `pass`, `✓` | matches | Details overlay, six AC rows, evidence, and `↓ 2 more lines`. |
| CompactAgents, 1424×975 | `idle`, `stopped`, `done`, `pass`, heavy selected border | matches | Manager, two children, joined tree, and footer. |
| CompactChat, 1424×975 | `coordinator`, `you`, `[input]`, `█` | matches | Table and input. The `stop refuses it.` row is present in the capture. |
| CompactFriction, 1424×975 | `[Open]`, `▶`, opened dates and seen counts | matches | All 15 rows, details hint, and footer. |
| CompactHealth, 1424×975 | `[Agents]`, `working`, `live`, `idle`, `unresolved`, `stopped` | matches | Five agents, panes, restart explanation, and footer. |
| CompactLogs, 1424×975 | `WARN`, `verify fail`, `[json]` | matches | Filter, complete JSON object, and footer. |
| Friction, 2096×1566 | `[Open]`, `▶`, `Open` | matches | List, entry fields, sections, occurrences, and overflow hint. |
| FrictionResolved, 2096×1566 | `[Resolved]`, `▶`, `Resolved 2026-10-21` | matches | List, resolution fields, summary, and reproduction. |
| Health, 2096×1566 | `[Home]`, `✓`, `!`, `warn`, `unresolved` | matches | Seventeen checks, warning details, explanation, and next step. |
| HealthAgents, 2096×1566 | `[Agents]`, `▶`, state words, `[restart]` | matches | Five rows, supervisor fields, prompt, and footer. |
| Help, 2096×1566 | `GLOBAL`, `ACTIONS`, `FILTER FIELDS`, `CHAT`, `[help]` | matches | Full key map, dimmed frame, and footer. |
| Logs, 2096×1566 | `DEBUG`, `WARN`, `▶`, `pass`, `fail`, `follow ●` | matches | Event list, JSON, filter, overflow ellipses, and footer. |
| Main, 2096×1566 | `idle`, `stopped`, `done`, `pass`, heavy selected border | matches | Manager, two children, joined tree, legend, and footer. |
| TooSmall, 1088×679 | `terminal too small`, `60×16`, `80×24` | matches | Full notice and last-row quit hint. The canvas keeps the 60-cell width. |
| Work, 2096×1566 | `[Tasks]`, `▶`, `landed`, `assigned`, `✓`, `pass` | matches | Four tasks, details, acceptance, evidence, and footer. |
| WorkGoals, 2096×1566 | `[Goals]`, `▶`, `planned`, `done`, `✓` | matches | Goal fields, waves, tasks, needs-you text, and footer. |
| WorkQueue, 2096×1566 | `[Queue]`, `▶`, `landed`, `assigned`, `! stop`, `unresolved` | matches | Dependency tree, grants, stop details, next step, and footer. |
| Idle glyph on all affected screens | `idle` | overridden | The user accepted the empty box for `◐` on 2026-10-21. No other missing glyph was found. |
| Bold weight | Heavy selection borders, `▶`, bracketed views, and state words remain | differs | As the worker recorded, freeze v0.2.2 renders regular weight. This tool limit does not remove text or symbols and does not fail AC1 or AC2. |

The colour-free capture text retains the state words and selection glyphs
listed above. Colour is not the only state signal.

### Symbol page

The verifier made a new 48×8 page with each decision 13 symbol, spaced
symbols, and joined light and heavy box grids. The page used one foreground
SGR code, so freeze detects ANSI on stdin. A first input without SGR failed
with `Language Unknown`; it produced no PNG. The ANSI rerun used the same
render command:

```sh
S=.hive/tmp/verify47/symsrc OUT=.hive/tmp/verify47/symbols bash .hive/tmp/render.sh
```

The verifier read the resulting PNG. All 38 supported symbols draw. Only
`◐` shows the accepted missing-glyph box. Light and heavy box joints join.

## V-gate run

At the start of V-gate, the verifier ran `git merge dev` once. It reported
`Already up to date.` There was no merge conflict.
Then `git diff dev...HEAD --stat` reported 24 files: 20 PNG files, the brief,
`V1.md`, `V-gate.md`, and one line in `testdata/ste-baseline.json`.
No dependency or Go source file changed.

The full command output is in [verifier-gate.txt](verifier-gate.txt).

| Command | Result | Output |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | pass | No output; exit 0. |
| `go mod tidy -diff` | pass | No output; exit 0. |
| `go vet ./...` | pass | No output; exit 0. |
| `golangci-lint run ./...` | pass | `0 issues.`; exit 0. |
| `go test -race ./...` | pass | All packages pass; exit 0. No retry was needed. |

## Cleanup

The verifier removed `.hive/tmp/verify47/` after it recorded the results.
Pre-existing worker scratch files and the source sign-off captures were
preserved. This run made no pane, tab, workspace, or scratch home.
