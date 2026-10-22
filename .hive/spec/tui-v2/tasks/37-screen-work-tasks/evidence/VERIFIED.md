# VERIFIED 37-screen-work-tasks

Result: fail

Verifier: `t37-verifier-1`. Date: 2026-10-22.
Task commit: `0646674161249213009d86bc7d3d4dffe4522b65`.
Merged dev base: `86f48a68b932bc2d15005c4cef151ed5acba2549`.

## Failure that needs repair

After the required merge of current `dev`, the TUI tests do not compile.
`internal/tui/work_tasks_test.go:18,45` and
`internal/tui/friction_test.go:17,40` both define `checkBoard` and
`requireFits`. The two `requireFits` functions have different signatures.
Use distinct helper names and run the checks again. The verifier did not
change these test helpers. Full diagnostics are in `verify/final-green.log`
and `verify/final-gate.log`.

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | fail | Base and mutation checks pass their expectations on the earlier tree. After the final dev merge, the task tests do not compile because their helpers conflict with the Friction test helpers. See `verify/final-green.log`. |
| V2 Live check | pass | Work at 120×40 and Compact at 80×24 each have 0 differing cells. Keys act as AC1 and AC4 require. See captures, diffs, PNG files, and `verify/keys.txt`. |
| V-gate Diff scope and gate | fail | Scope is permitted. Formatting and module checks print nothing and exit 0. Vet, lint, and race tests exit 1 because the TUI test helpers conflict. Full output: `verify/final-gate.log`. Scope: `verify/scope.log`. |

## V1 commands and results

The base is `866bde2a`, before the Tasks implementation. I extracted it into
`.hive/tmp/verify37/base`. I copied the task test files into that folder.

```sh
go test ./internal/tui -run 'TestWorkTasks|TestGoalsOpenTasksListsEachProject' -v
go test ./internal/tui/component -run 'TestListFillGolden|TestSubTabsHint' -v
```

Each of the 12 `TestWorkTasks` tests fails on the base. These include the
normal-mode loaders, refresh, details, and keys. The component tests first
fail to compile because the new fields do not exist. After I added the fields
without behavior, both tests fail on their assertions. I extracted
`TestGoalsOpenTasksListsEachProject` into a separate base test file. It fails
because `p` hides a project's tasks. The base test log records that failure.

On the task branch, I ran this command before and after the mutations:

```sh
go test ./internal/tui ./internal/tui/component -run 'TestWorkTasks|TestGoalsOpenTasksListsEachProject|TestListFillGolden|TestSubTabsHint|TestQueueKeys|TestScreenSelectionsUseHoneyMarker' -v
```

Both runs exit 0 on the initial tree. The same tests also pass after the
Logs merge (`021361fc`), in `verify/postmerge-green.log`. After the Friction
merge (`06466741`), the test package does not compile. The tests cover the
exact Work and Compact grids, all
required sizes, empty data, long text, normal-mode records and refresh,
selection, file keys, overlay keys, and the Queue-to-Tasks route.

I changed `return doneStyle` to `return infoStyle` in `taskStatusStyle`.
`go test ./internal/tui -run TestWorkTasksMatchesCompactBoard -v` fails with
66 differing cells. I restored the token.

I changed `case "d":` to `case "D":` in `workTasksScreen.Update`.
`go test ./internal/tui -run 'TestWorkTasks(FileKeys|OverlayKeys)' -v` fails
because the dependency graph does not open. I restored the key. The source
diff was empty after restoration.

## V2 commands and observations

I built `.hive/tmp/verify37/hive` with `go build`. Each run used a new scratch
home with a `.hive/README.md` marker. Demo mode implies `--no-startup`.
No coordinator or supervisor ran in the scratch home.

```sh
HIVE_HOME="$PWD/.hive/tmp/verify37/home" COLORTERM=truecolor .hive/tmp/verify37/hive --demo .hive/spec/tui-v2/design/Work.fixture.json --size 120x40
HIVE_HOME="$PWD/.hive/tmp/verify37/home" COLORTERM=truecolor .hive/tmp/verify37/hive --demo .hive/spec/tui-v2/design/Compact.fixture.json --size 80x24
herdr pane read w9Y:p5 --source visible --format ansi
```

The final captures use tab `w9Y:t5`, pane `w9Y:p5`, after the Friction merge.
Earlier captures used owned panes `w9Y:p2` and `w9Y:p3`. I also created and
closed an unused scratch pane `w9Y:p4`. Two reads 500 ms apart
were equal for each board. The socket read records show `truncated: false`.
The pane records show scroll offset 0. The comparator crops the top-left
board area and calls `cellgrid.Convert`. It compares every non-space cell
fully. It ignores token and bold only when both cells are spaces, as the
manager decision requires. It uses no mask. Its source is `verify/compare.go.txt`.

| Board | Grid result |
| --- | --- |
| Work | 0 cells differ |
| Compact | 0 cells differ |

| Live key sequence | Observation |
| --- | --- |
| Work: `j` | Task 2 is selected and its details show. |
| Work: `k b` | Task 3's fixture brief opens with its restart heading. |
| Work: `esc v` | The fixture `VERIFIED.md` opens with `Result: pass`. |
| Work: `d` | The dependency graph of task 3 opens. |
| Work: `esc enter` | Task details return; the footer shows `[details]` and its keys. |
| Compact: `down down` | The details scroll to the V1 and V3 rows. The overflow notice goes away. |
| Compact: `esc` | The list returns with task 3 selected and list key hints. |
| Compact: `j enter d` | The overlay opens for task 2 and shows its dependencies. |
| Compact: `esc` | Task 2 details return inside the overlay. |
| Compact: `esc` | The list returns with task 2 selected. |

## Visual review

I captured both approved boards again with Firefox, a new profile for each,
and `--headless --no-remote --window-size=1220,780 --screenshot`.
Board paths:

- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Work.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Compact.dc.html`

I converted each ANSI capture to an HTML `pre` and captured it with Firefox.
The board and live PNG files are in `verify/`. I reviewed all four PNG files.
The PNG font size and page margins differ; cell layout, text, tokens, and
weights match the grid checks. I also checked the captured text without
colour. States have the words or glyphs below.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Work header | `working`, `live`, `needs you 0`, `14:22` | matches | Fixture values. |
| Work tab bar | `3 Work` with underline | matches | Active tab. |
| Work view row | `[Tasks]`, `[ ] switch view` | matches | Active view. |
| Work list frame and title | `tasks · 4`, borders | matches | Focused list. |
| Work column labels | `ID`, `STATUS`, `WAVE` | matches | Board columns. |
| Work selection | `▶ 3-tui-restart` | matches | Selection has a glyph. |
| Work other task IDs | `2-stale-coordinator-star…` and other IDs | matches | Board truncation. |
| Work status and wave | `landed`, `assigned`, `C`, `B`, `A`, `—` | matches | States have words. |
| Work details frame | `3-tui-restart`, borders | matches | Side-by-side details. |
| Work detail fields | `title`, `status`, `goal`, `worker`, `verifier`, `branch`, `merge`, `depends` | matches | Field values match. |
| Work acceptance | `ACCEPTANCE`, `✓ AC1` through `✓ AC6` | matches | Pass marks have glyphs. |
| Work evidence | `EVIDENCE`, `b`, `v`, `d`, V1 and V3 `pass` | matches | Results have words. |
| Work footer | `[list]` with key hints | matches | List focus. |
| Compact header | `coord ●`, `sup ●`, `needs you 0`, `14:22` | matches | State glyphs remain without colour. |
| Compact tab bar | `3 Work` with underline | matches | Active tab. |
| Compact view row | `[Tasks]`, `list › details` | matches | Overlay focus. |
| Compact overlay frame | `3-tui-restart · esc back`, borders | matches | Full-width details. |
| Compact short fields | `status`, `goal`, `verifier`, `merge` | matches | Values include `landed` and `pass`. |
| Compact acceptance | `✓ AC1` through `✓ AC6` | matches | Pass marks have glyphs. |
| Compact evidence | `b brief.md`, `v VERIFIED.md`, `d deps graph` | matches | Action labels. |
| Compact overflow | `↓ 2 more lines`, `┃` | matches | Overflow and scroll position have glyphs. |
| Compact footer | `[details]` with key hints | matches | No Help hint inside the overlay. |

## Gate, scope, and cleanup

`verify/gate.log` has the full initial gate output; all commands exit 0.
`verify/postmerge-gate.log` has the gate after the Logs merge; all commands
exit 0. `verify/final-gate.log` has the gate after the Friction merge. Vet,
lint, and race tests fail on duplicate test helpers. These are compile
failures, not the known runtime flakes.

The two component changes are `List.Fill` and `SubTabs.Hint`. Other TUI edits
remove the old Tasks screen and connect its replacement. The Queue repair
removes references to the deleted status and project fields and updates its
route test, as the manager allowed.

At the manager's request, I merged `dev` twice: Logs at `fbca94d5`, then
Friction at `86f48a68`. I kept both decision sections and baseline entries.
For Friction, I kept `List.Fill` and `List.Hint`, both component tests, both
new screen names in the package map, and both new empty-list labels. I
removed both obsolete legacy-details cases, because both screens now use
the panel hooks. The merge did not resolve the duplicate test helpers
because Git saw no textual conflict in those files. I reported the failure
to the manager.

I closed the scratch pane and its tab. The tab list then showed only
`w9Y:t1`. I removed the scratch home, binary, extracted base, scripts, browser
profiles, and duplicate captures after I saved this evidence. No source
mutation remains. No worker evidence file was changed.
