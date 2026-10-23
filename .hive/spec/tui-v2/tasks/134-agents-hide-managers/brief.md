# Agents: hide stopped managers and empty pods

| Item | Value |
| --- | --- |
| Feature | `tui-v2`, goal `g9-agents-hide-managers`, wave A |
| Tier | strong |
| Depends on | none. Base: `dev` at `6a586f1a` or later. |
| Grant | Goal `g9-agents-hide-managers`. The user approved the plan in chat `c175` and asks `a3` and `a4`. Land on `dev`. |
| Evidence | The `evidence/` folder beside this brief. Put `VERIFIED.md` there. |

## User decisions: approved on 2026-10-23

Coordinator inbox record 536 relays the user decisions from ask `a5` and chat.
The user approved the code grids as golden files, as in tasks 126 and 131.
Update the final scenario grids to follow these decisions.
Do not request a second approval unless a rendered grid disagrees with the approved behavior.

1. A direct child of a hidden manager stays in its project pod, under the pod top border.
2. A child of a hidden worker keeps decision 39 and moves to home.
3. With registered projects and all project pods hidden, show the notice `no project agents` and the live home cards.
4. Keep the coordinator and the live home agents visible in that state.
5. Show the notice alone only when nothing at all is visible in that state.
6. Keep decision 10 unchanged for an empty project registry.

This section replaces the earlier approval hold.
The manager clears task and wave `needs_user`.
The worker completes the task after the tests, approved grids, live checks, and gate pass.
The manager then spawns an independent verifier and lands the verified branch.

## Plan

1. Write failing tests for the approved behavior.
2. Make the smallest source change.
3. Export code grids and captures for user approval.
4. Apply the user decisions relayed in coordinator record 536.
5. Apply the approved grids and run each verification item.

## Design

Read `/home/penguin/source/hive/.hive/spec/tui-v2/README.md`.
The approved boards are in `/home/penguin/source/hive/.hive/spec/tui-v2/design/`.
Read these boards as the layout references:

- `/home/penguin/source/hive/.hive/spec/tui-v2/design/Main.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/AgentsMany.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/CompactAgents.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/AgentsHome.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/AgentsCoordinatorChild.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/LiveOverview.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/LivePod.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/CompactLiveOverview.dc.html`
- `/home/penguin/source/hive/.hive/spec/tui-v2/design/CompactLivePod.dc.html`

Goal g9 changes the hide rule of these references.
The user approved the resulting code grids through ask `a5`, as coordinator record 536 states.
Follow the visual check procedure in the worker skill.

## Requirements

### AC1 Managers follow the hide rule

- Hide each current stopped or stale manager when stopped cards hide.
- Keep the existing default: stopped cards hide at TUI start.
- Remove the manager exemption from task 129.
- Keep current-generation selection and deduplication unchanged.
- Keep Health behavior unchanged.

### AC2 Pods follow visible cards

- Remove a pod from Agents when it has no visible card.
- Keep the pod border and title when a hidden manager has a live child.
- Attach each visible top child to the pod top border.
- Use the existing tree of the home pod without a manager as the reference.
- Apply this rule in the overview and the one-pod view.
- Keep nested child relationships that remain visible.
- Keep a direct general-agent child of a hidden manager in the project pod.
- Keep the home rule for a child of a hidden worker.
- When all project pods hide, draw the notice and the visible coordinator and home cards.
- Show only the notice when nothing at all is visible in that state.
- Keep decision 10 unchanged for no registered project.
- Keep selection and scrolling valid when a pod disappears or returns.

### AC3 Counts and discovery

- Count a hidden manager in the pod text `h shows N stopped`.
- Count a shown stopped manager in `h hides N stopped`.
- Keep the meaning of `h`. Add no key.
- Show one persistent hint for `h` on the Agents tab when cards hide.
- Keep the pod count text required above.
- Keep this hint visible when every project pod disappears, also at 80 columns.
- Use a glyph or a word for each state. Colour only adds to the state.

### AC4 Code grids before golden changes

- Render each changed board at its board size from the code.
- Put these grids in `evidence/new-grids/` in the exporter format.
- Add representative fixtures for a hidden manager with a live child and for no visible project cards.
- Include the two edge cases from `evidence/review-cases.md` in the final scenario grids.
- Keep the approval package and final evidence consistent with the approved decisions.
- Include both 120x40 and 80x24 cases in the approval package.
- Write `evidence/new-grids/README.md` with the changed boards, changed rows, and causes.
- List each board that stays unchanged.
- Put terminal PNG captures and plain text captures beside the grids.
- Commit and push the approval package.
- Send its commit and path to `manager-hive` with `hive inbox send`.
- Do not change a golden grid, board source, or existing board fixture before approval.
- Do other independent work while approval is pending.
- The manager owns task `needs_user` because only the controlling role can edit that field.
- Make each approved board source export its approved code grid.
- Change only the approved golden files and the fixtures they require.

### AC5 Spec and tests

- Add the approved decisions to the TUI spec after decision 104.
- Write tests before the source change.
- Use literal expected cards, counts, text, and connector glyphs.
- Cover stopped, exited, and stale managers with and without live children.
- Cover the overview and the one-pod view at both widths.
- Cover `h` twice and selection after the visible pod set changes.
- Cover the hint when all project pods disappear.
- Keep the existing test checks. Update an expectation only for an approved behavior change.
- Record why each changed expectation follows goal g9.

## Source files

| Path | Purpose |
| --- | --- |
| `internal/tui/agents_screen.go`, `internal/tui/tui.go`, and tests beside them | The hide rule, pod layout, counts, hint, selection, and tests. |
| `internal/tui/frame.go` and its tests | The persistent hint, only if the existing footer needs a change. |
| `internal/tui/testdata/` | New record fixtures only when required. |
| `.hive/spec/tui-v2/README.md` | The approved g9 decisions. |
| `.hive/spec/tui-v2/design/*.dc.html`, `*.grid`, `*.fixture.json` | Only changes that the user approves after the code grids. |
| `testdata/ste-baseline.json` | Add this brief with baseline 0. Counts can fall. |
| This task folder | The brief and evidence. |

Do not change other files.
Do not change the exporter or add a golden mask.
Report a required scope change before you edit another source file.

## Verification items

### V1 Tests red and green

Run each new test against the base implementation.
Record its failure and the correct cause.
Run each new test against the task implementation.
Run the focused tests with `-count=50` and with `-race`.
For a guard that passes on the base, deliberately break the guarded behavior.
Record that the guard fails, then restore the source.
Expected: tests detect the broken behavior and pass with the implementation.

### V2 Grid approval and comparison

Run the exporter after the user approves the code grids.
Compare each exported grid with its approved code grid through `internal/lib/cellgrid`.
Run each existing golden check.
Expected: each approved grid matches cell for cell, including colour tokens and weight.
Expected: no unapproved golden file changes.
Record the approval reference and the approved commit in the evidence.

### V3 Live behavior and visual check

Build the task binary.
Run it in scratch herdr panes with `HIVE_NO_STARTUP=1` and `COLORTERM=truecolor`.
Use demo fixtures or a scratch home.
Capture the overview and the one-pod view at 120x40 and 80x24.
Check the hidden manager with a live child, all-hidden pods, and the `h` cycle.
Check the border, title, top-child connectors, count, hint, and selection.
Compare each changed screen capture with its approved grid cell for cell.
Include the visual table required by the worker skill.
Expected: each changed screen matches its approved grid.
Expected: every approved behavior has a live check.
The worker and verifier each run this item.

### V4 Diff scope and gate

Merge current `dev` into the task branch before the final gate.
Run `git diff dev...HEAD --stat`.
Run the full gate from the worktree root:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

Expected: only the named files change.
Expected: the first two commands print nothing.
Expected: each other command exits 0.
Record the full output in the evidence.

## Cleanup and finish

Never run bare hive without `--no-startup` or `HIVE_NO_STARTUP=1`.
Never run the test TUI on the real home.
Close only panes and processes that you started.
Remove task scratch files after you record their results.
Keep the final evidence and captures.
Check cleanup before you report completion.
Commit and push the task branch without an agent co-author.
Run `hive task complete` only after approval and each verification item passes.
Send the commit SHA and results to `manager-hive`.
