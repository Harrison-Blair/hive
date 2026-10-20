# 114-visual-check-rule VERIFIED

| Item | Value |
| --- | --- |
| Task | 114-visual-check-rule |
| Commit | d90368ab |
| Date | 2026-10-20 |
| Agent | verifier, Claude Code, claude-opus-5-5 (t114-verifier-1) |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1-R4, colour scope change | pass | Each point of the rule and the colour point is in the file that R1-R4 names. No script or command added. |
| V2 | recipes | pass | Board PNG and pane PNG both show the board and the screen. Not committed. |
| V3 | scope | pass | Only source-table files and task evidence changed. |
| V4 | gate | pass | `gofmt -l` and `go mod tidy -diff` print nothing. Other commands exit 0. |

## Runs

### V1

Read `git diff origin/dev` of the four files.

| Point | Worker skill (R1) | Verifier skill (R2) | BRIEF-FORMAT (R3) | EVIDENCE-FORMAT (R4) |
| --- | --- | --- | --- | --- |
| Trigger: `Design` section with board absolute paths | yes | yes | yes, new `Design` section and "A UI task must have" rule | - |
| Verifier flags a UI change with no `Design` section (`internal/tui/`, `*_view.py`, hub CSS) | - | yes | - | - |
| Worker step: capture, compare, fix each gap | yes | - | - | - |
| Verifier step: capture again independently, judge | - | yes | - | - |
| Board recipe (firefox headless, new scratch profile, 1220x780) | yes | yes | - | - |
| Textual recipe (`app.save_screenshot()`, `rsvg-convert`) | yes | yes | - | - |
| Pane recipe (`herdr pane read --source visible --format ansi`, SGR to `<pre>`, firefox) | yes | yes | - | - |
| Evidence: commit screen PNGs, board PNG under `.hive/tmp/`, give path, never commit a board | yes | - | - | yes |
| Element table: matches / differs (cause) / overridden (decision) | yes | yes | - | yes |
| Column `state without colour` | yes | yes | - | yes |
| Ownership: capture only own panes or a scratch run | yes | yes | - | - |
| Colour is never the only signal of a state | yes, with plain-text check | yes, with plain-text check and "Fail a state that only colour shows" | yes | yes, with "The verifier fails a state that only colour shows" |

`git diff origin/dev --stat` shows only Markdown and one JSON baseline line. No script or command is added.

### V2

Board recipe, new profile folder in scratch:

```sh
firefox --headless --no-remote --profile SCRATCH/ffprof1 --window-size=1220,780 \
  --screenshot /home/penguin/source/hive/.hive/tmp/verify-114/Main-board.png \
  file:///home/penguin/source/hive/.hive/tmp/coordinator/designs/hive-tui/Main.dc.html
```

Output: `PNG image data, 1220 x 780`. The image shows the Main board: header, AGENTS tree, GRAPH view, detail bar, and key hints.
Board PNG path: `/home/penguin/source/hive/.hive/tmp/verify-114/Main-board.png`.

Pane recipe: the verifier split pane `wV3:p2` in its own workspace `wV3`. It ran a scratch build of `hive` with a scratch `HIVE_HOME`, a temp `XDG_CONFIG_HOME`, and the `HERDR_*` variables and `TERM_PROGRAM` removed. TUI pid `3936291`. The TUI printed "not in herdr: no coordinator or supervisor started".

```sh
herdr pane read wV3:p2 --source visible --format ansi > SCRATCH/pane.ansi   # 8296 bytes
python3 -I SCRATCH/ansi2html.py SCRATCH/pane.ansi > SCRATCH/pane.html        # SGR 0/1/38;2 to <span> in <pre>
firefox --headless --no-remote --profile SCRATCH/ffprof2 --window-size=1220,780 --screenshot SCRATCH/pane.png file://SCRATCH/pane.html
```

Output: `PNG image data, 1220 x 780`. The image shows the TUI screen: header line, SCREENS list with `▶ 1 Chat`, coordinator panel "not live · no coordinator", hive events panel.
Cleanup: `kill 3936291` and `herdr pane close wV3:p2`. After that, `pgrep` found no scratch `hive` and `herdr pane list` shows only `wV3:p1`. No PNG is committed.

### V3

`git diff origin/dev --stat`:

```text
 .agents/skills/hive-verifier/SKILL.md              |  44 +++++++++
 .agents/skills/hive-worker/SKILL.md                |  45 +++++++++
 .hive/spec/BRIEF-FORMAT.md                         |  10 ++
 .hive/spec/EVIDENCE-FORMAT.md                      |  13 +++
 .../hive-tui/tasks/114-visual-check-rule/brief.md  | 103 +++++++++++++++++++++
 .../tasks/114-visual-check-rule/evidence/V1.md     |  81 ++++++++++++++++
 .../tasks/114-visual-check-rule/evidence/V2.md     |  75 +++++++++++++++
 .../tasks/114-visual-check-rule/evidence/V3.md     |  33 +++++++
 .../tasks/114-visual-check-rule/evidence/V4.md     | 102 ++++++++++++++++++++
 testdata/ste-baseline.json                         |   1 +
```

The baseline diff adds only `".hive/spec/hive-tui/tasks/114-visual-check-rule/brief.md": 0`. No count rose.

### V4

```text
gofmt -l $(git ls-files -co --exclude-standard '*.go')   # no output, rc=0
go mod tidy -diff                                        # no output, rc=0
go vet ./...                                             # rc=0
golangci-lint run ./...                                  # 0 issues., rc=0
go test -race ./...                                      # all ok, rc=0
```

`go test -count=1 -run 'STE|Ste|Skill' -v .` shows each STE and skill test as PASS.
