# Verification: 46-agents-home-graph

Verdict: **pass**. Verifier: `t46-verifier-1`. Task code: `4f9ea6d6`.

## Items

| Item | Result | Independent run and evidence |
| --- | --- | --- |
| V1 Tests | pass | Base compile run, base run with an unset `levels` field, task run, 20 deliberate breaks, and restored task runs. See `verifier/base-compile.txt`, `base.txt`, `green.txt`, `mutations.txt`, `B1.txt` through `B20.txt`, `restored-green.txt`, and `extra-restored-green.txt`. |
| V2 Live rows | pass | Rebuilt the binary. Ran all seven new fixtures and `AgentsMany` in scratch pane `wAT:p2`. Each capture has 0 differences under the live space-cell rule. See `verifier/live-compare.txt`, `capture-metadata.json`, ANSI captures, plain-text captures, and PNG pairs. Report rows also match: `verifier/report-check.txt`. |
| V-gate Diff scope and gate | pass | Fetched `origin/dev`; `dev` and `origin/dev` both name `2adaf0ab`. `git merge dev` returned `Already up to date.` No merge conflict. All five gate commands exit 0. See `verifier/diff-stat.txt` and the full output in `verifier/gate.txt`. |
| AC5 Boards and grids | pass | All eight board files and `canvas.json` equal the approved files in `/home/penguin/source/hive/plans/tui-v2/design/`. Re-exported the design folder. Every grid equals its committed copy. Only the seven new grids and `AgentsMany.grid` differ from `dev`. All seven new golden checks pass. See `verifier/boards.txt`. |

## V1 method

Ran the new tests with `go test ./internal/tui -count=1 -run 'TestAgentsHome|TestAgentsNest|TestAgentsChildren|TestAgentsHMoves|TestAgentsCoordinatorChildren|TestAgentsManagerWithout|TestAgentsArrowsReach|TestAgentsCompactHome|TestAgentsDetailsOfAGeneral|TestHealthAgentsRestartRefuses|TestGeneralAgentBoard' -v`.

Used a detached scratch worktree for the base and each break. The base lacks `graphPod.levels`; the first run fails to compile for that reason. Added only the unset field for the base behavior run. The first requested test fails with `review-1 is in pod ""; want home`. The no-project notice test and the Health fixture subtest pass on the base, which already meets those rules. B4 and B16 fail tests when those rules break. All task tests pass before and after the breaks.

The 20 breaks cover home placement, project count, empty home, the no-project notice, home order, parent placement, creation order, nesting indent, rail form, hidden parent, the empty-child text, coordinator children, distinct identities, empty popup fields, restart refusal, Health role, nested joints, home scroll text, card dimensions, and card subline. Each break fails a behavioral or golden test. No compile failure counts as a successful mutation check. Full commands and outputs are in the numbered logs. The scripts are saved as `verifier/tests.py.txt` and `extra.py.txt`.

## V2 method and keys

Herdr 0.9.5, protocol 22. The scratch pane has 57 rows. Used a scratch `HIVE_HOME` with the home marker, `HIVE_NO_STARTUP=1`, and `COLORTERM=truecolor`. For each board, ran the newly built binary as `hive --demo ABSOLUTE_FIXTURE --size WxH`. Each initial capture and each key capture has two equal reads 500 ms apart, `truncated: false`, and scroll offset 0. Cropped the top-left fixture size before comparison.

The report rows, tokens, and bold flags equal the corresponding golden rows. The report gives all seven new boards and exactly the changed rows and columns of `AgentsMany` (178 changed cells). Section (a) correctly gives the `AgentsNested` right-border scroll mark on rows 4–34.

| Fixture | Keys and observation |
| --- | --- |
| AgentsHome | `enter` opens `plan-review-1`; `esc left` selects `t3-worker-1` in the project; `right` selects `plan-review-1` in home again. |
| AgentsNested | `up` selects `t5-research-1`; a second `up` selects `t5-worker-1`; `down down down` selects `t6-worker-1`. Each selection has a heavy border and joint. |
| AgentsHiddenParent | `h` shows `t4-verifier-1`; the selected `t4-review-1` moves under it on the rail. A second `h` moves the selected child back to home. The stopped text changes with it. |
| AgentsCoordinatorChild | Both the coordinator child and its nested child show in home. `left` selects `t4-worker-1` in the project; `right` first selects its tree sibling `t4-verifier-1`, as the graph key rule requires. |
| CompactAgentsHome | `left left` reaches the first page with `home →`; `right right` reaches home with `projects 2–2 of 2 · home`. |
| AgentDetailsGeneral | The popup shows role, status, harness, model, pane, tab, parent, actions, and recent events. Empty fields are absent. `esc` closes it; `enter` opens it again. |
| HealthAgentsGeneral | The selected row has role `agent`. `r` shows `restart: coordinator and supervisor only`. No restart prompt appears. |
| AgentsMany | The verifier precedes the workers by creation time. The empty manager pod says `no child agents`. |

## Visual review

Captured the eight approved boards and eight initial screens again as PNG files in `verifier/`. Board files use the approved source. Screen files use the captured ANSI SGR text in an HTML `pre`. Firefox used a new profile for each capture, `--headless --no-remote --window-size=1220,780 --screenshot OUT.png file://SOURCE`. See `verifier/live.py.txt` and `compare.go.txt` for the capture and comparison steps.

Reviewed each PNG pair. The browser page padding, font metrics, and background outside the cell grid differ between the two capture methods. These do not change a cell. All non-space cells match their board character, token, and weight. Removed SGR from every capture and checked the state signals below.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header and tabs | `hive`, coordinator and supervisor state words or dots, clock, tab labels, active underline | matches | Eight board pairs; compact header uses its approved form. |
| Frame and footer | borders, `[graph]`, `[details]`, or `[list]`, key words | matches | Eight board pairs. |
| Home column | `home`, last column, no goal text | matches | AgentsHome and coordinator-child boards. |
| Project count | `1 project` or `2 projects`; home is excluded | matches | Graph boards. |
| General-agent card | dot and name, harness, state word, age | matches | Cards retain the 24×4 form. |
| Selected card | heavy `┏━┓┃┗┛` border and mixed-weight joint | matches | Selection remains visible without colour. |
| Nested rails | `│`, `╰┤`, `├┤`, parent-bottom joint, two-cell steps | matches | AgentsNested and AgentsCoordinatorChild. |
| Nested overflow | right-border `┃`, cut pod | matches | AgentsNested rows 4–34; corrected report agrees. |
| Hidden parent | child in `home`, `h shows 1 stopped` | matches | AgentsHiddenParent; live `h` returns the child to its parent. |
| Compact page | `projects 2–2 of 2 · home`, `← hive`, horizontal mark | matches | CompactAgentsHome. |
| Popup | title with `esc closes`, `role agent`, populated fields, ACTIONS, RECENT | matches | AgentDetailsGeneral; no empty field. |
| Popup backdrop | visible text and borders behind the popup | matches | Approved dimmed backdrop. |
| Health row and selection | `▶ plan-review-1`, `agent`, `● working` | matches | HealthAgentsGeneral. |
| Health action restriction | `r restarts only the coordinator and the supervisor` | matches | Board details; live refusal also shows the footer notice. |
| Empty-child notice | `no child agents` | matches | AgentsMany. |
| Child order | verifier, then workers | matches | AgentsMany; approved creation-time order. |
| Card states | `● working`, `◐ idle`, `○ stopped` | matches | Words and glyphs retain the states without colour. |
| Health exceptional state | `unresolved` | matches | Word is present without colour. |

## Scope and cleanup

The diff stays within the brief source table and evidence. The spec board table adds seven data rows. `boardexport.Sizes` adds seven data entries with no logic change. The fixture-count test changes 21 to 28. No current fixture changes. The general-agent spec changes only Manager decisions. The brief is covered by the existing `steFiles` glob and has baseline 0.

The full gate passed on the first run; no known-flake exception was needed. `gofmt -l` and `go mod tidy -diff` printed nothing. `go vet ./...`, `golangci-lint run ./...`, and `go test -race ./...` exited 0.

Closed the scratch tab and removed the detached worktree, binary, scratch home, Firefox profiles, generated HTML, and temporary scripts after saving the evidence. No source file or worker evidence file was changed by the verifier.
