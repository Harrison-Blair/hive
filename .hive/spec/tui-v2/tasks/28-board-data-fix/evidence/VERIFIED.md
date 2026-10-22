# Verification: 28-board-data-fix

Verifier: `t28-verifier-1` (pi). Reviewed commit: `51534c61`.
Verdict: **pass**.

Cleanup: removed `.hive/tmp/t28-verify/`, including all profiles, PNGs,
plain-text extracts, scratch programs, and temporary output. Confirmed that
the folder is absent. `git status --short` shows only this new report.
`git diff --check` prints nothing. No task source file remains modified.

`git merge dev --no-edit` printed `Already up to date.`
All commands ran from the task worktree root.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Boards | pass | All 21 `cmp` commands exited 0. Each of the seven copy PNGs equals its approved-board PNG byte for byte. No capture shows `{{`. |
| V2 Grid diff | pass | `cellgrid.ReadGrid` and `cellgrid.Compare` with no mask report only the AC2 cells. The other 14 grids have zero changed cells. A fresh export exits 0 and leaves no diff. |
| V3 Fixtures | pass | All 21 fixtures load. No fixture contains `w58:p3`. Each coordinator record uses `w58:p4`. Independent samples: AgentDetails, AgentsMany, Logs. |
| V-gate Diff scope and gate | pass | The diff contains only allowed files and evidence. The six sample-time literals have manager approval. All five gate commands exit 0. The first two print nothing. No health-test flake occurred. |

## V1 Capture check

For each of the 21 boards, ran `cmp /home/penguin/source/hive/plans/tui-v2/design/NAME.dc.html .hive/spec/tui-v2/design/NAME.dc.html`.
All exit statuses were 0.

Captured both versions of AgentDetails, AgentsMany, Chat, CompactChat,
CompactLogs, HealthAgents, and Logs. Used this command for each capture:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

Profiles and PNGs were in `.hive/tmp/t28-verify/`. The initial attempts
failed because the profile directories did not exist. Created the directories
and repeated all 14 commands. All 14 then exited 0. Each pair is byte-identical.
Inspected the seven copy images. The fixed values are visible. No image shows
an unresolved template. The copies equal the approved captures, including the
unchanged board elements. This task changes no application screen.

Removed colour by reading only the `t|` text rows of the grids. State words
and glyphs remain. The following table covers the board elements of the
seven captures. Shared elements have one row.

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Shared header and tabs | `working`, `live`, `needs you`, `●`; active tab underline | matches | Copy and approved PNGs are identical. Compact headers retain the state glyphs. |
| Shared frame and footer | box lines, region names, key words | matches | Same geometry and text as approved captures. |
| AgentDetails context and metadata | `manager`, `live · idle`, `done`, pane and parent names | matches | Same approved details and context. |
| AgentDetails actions | `a attach to pane`, `p peek pane`, `t open task · none` | matches | Same approved actions. |
| AgentDetails RECENT | `14:08`, `14:07`, `13:58`, `13:44` | matches | AC2 times are visible. |
| AgentsMany project headings | project names, goal IDs, `wave A`, `wave B`, `needs you` | matches | API and docs goal text agrees with AC2. |
| AgentsMany cards and selection | `working`, `idle`, `stopped`, `pass`, `done`; box outline | matches | States remain explicit in plain text. |
| AgentsMany legend and scrolling | `● working`, `◐ idle`, `○ stopped`, `projects 1–4 of 5`, arrows | matches | Same legend and scroll text. |
| Chat feed, table, code and earlier count | sender names, `42 earlier`, table words, code text | matches | Same approved feed. |
| Chat final user message | `14:22 you`, `1A, 2A, 3B` | matches | AC2 time is visible. |
| Chat composer | `message`, `>`, `[input]` | matches | Same approved prompt and cursor. |
| CompactChat feed | `14:21 coordinator`, choice table and recommendation | matches | Same compact feed. |
| CompactChat final message and composer | `14:22 you`, `1A, 2A, 3B`, `message`, `>`, `[input]` | matches | AC2 time is visible. |
| Logs and CompactLogs filter | `level:debug proj:hive`, day and arrows | matches | `proj:` uses dim and `hive` uses fg. No space separates these two board cells. |
| Logs event list and follow state | `events · 312`, times `13:19:40` through `13:45:01`, `WARN`, `DEBUG`, `follow`, `>` | matches | Fixed times and explicit level words remain. |
| Logs and CompactLogs event JSON | `event 13:38:02`, `13:38:02.114Z`, `WARN`, `verify fail`, pass/fail words | matches | Fixed popup and JSON times are visible. Compact popup hides its list. |
| HealthAgents view and list title | `[Agents]`, `agents · 5 · h hides 1 stopped` | matches | Correct stopped count. |
| HealthAgents rows and selection | `working`, `live`, `idle`, `unresolved`, `stopped`, `>` | matches | Each state has a word. |
| HealthAgents details and actions | `supervisor`, `live`, PID, pane, `r restart`, `a attach`, `p peek` | matches | Same approved details. |
| HealthAgents restart confirmation | `restart supervisor-g4?`, `live`, `stopped`, `y restart`, `esc cancel` | matches | Same confirmation and explanatory text. |

## V2 Cell comparison

A scratch Go program reads each `dev` grid with `git show dev:PATH`, reads
the task grid, and calls `cellgrid.ReadGrid` for both inputs. It then calls
`cellgrid.Compare(old, current, nil)` and prints `cellgrid.FormatDiffs`.
Ran `go run ./.hive/tmp/t28-verify/griddiff`. Exit: 0.
Coordinates below are zero-based. Each changed cell is listed in the output.

The counts are AgentDetails 4, AgentsMany 38, Chat 1, CompactChat 1,
CompactLogs 20, HealthAgents 1, and Logs 36. Only the filter cells change
foreground tokens. No bold flag changes. All changes agree with AC2.

### Golden-test red and green runs

Ran:

```sh
go test ./internal/tui/component -run '^(TestChatGolden|TestCompactChatGolden|TestPopupGolden)$' -count=1 -v
```

For the red run, temporarily used the `dev` AgentDetails, Chat, and
CompactChat grids. All three golden tests failed on the six expected time
cells. Restored the task grids. The same command then passed all three tests.
No source or grid mutation remains.

## V3 Fixture checks

Ran:

```sh
go test ./internal/tui -run '^TestEachBoardFixtureLoadsAtItsSpecSize$' -count=1 -v
rg -n 'w58:p3' .hive/spec/tui-v2/design/*.fixture.json
```

The load test passes all 21 subtests. The search prints nothing and exits 1
(no matches). For its red run, moved AgentDetails.fixture.json to scratch.
The load test failed because that fixture was missing. Restored the fixture
and repeated the same test: all 21 pass. This checks the loader, not the
semantic values; the following independent comparisons check the values.

Source paths below are relative to `.hive/spec/tui-v2/design/`.

| Chosen board | Board script lines | Fixture lines | Independent observation |
| --- | --- | --- | --- |
| AgentDetails | AgentDetails.dc.html:60 | AgentDetails.fixture.json:34–37 | Four log times are 13:44:20, 13:58:10, 14:07:30, 14:08:05. Their minute values equal the four RECENT entries, in reverse display order. |
| AgentsMany | AgentsMany.dc.html:70,72 | AgentsMany.fixture.json:32,34 | example-api has g4-rate-limits with first open wave B. example-docs has g2-guides with first open wave A. Both equal the board goal strings. |
| Logs | Logs.dc.html:44 | Logs.fixture.json:345 | Filter is level:debug proj:hive. Board uses dim for proj: and fg for hive. |
| Logs | Logs.dc.html:48–63 | Logs.fixture.json:320–335 | All 16 visible event times match an independent expected list, from 13:19:40 to 13:45:01. |
| Logs | Logs.dc.html:74–75 | Logs.fixture.json:328,343 | Selected event and selection both use 2026-10-21T13:38:02.114Z. Popup title uses 13:38:02. |
| Coordinator panes, all fixtures | Health.dc.html:51; CompactHealth.dc.html:52 | Coordinator record at line 12, except AgentsMany line 16 | Every existing coordinator record uses w58:p4. TooSmall has no coordinator record. No fixture holds w58:p3. |

Also compared all 312 events of Logs and CompactLogs against `dev`. Each
timestamp is exactly one hour earlier. The worker evidence records the
remaining board and fixture source lines. No mismatch was found.

## Independent command output

### V1 comparisons

```text
AgentDetails.dc.html cmp exit 0
AgentsMany.dc.html cmp exit 0
Chat.dc.html cmp exit 0
Compact.dc.html cmp exit 0
CompactAgents.dc.html cmp exit 0
CompactChat.dc.html cmp exit 0
CompactFriction.dc.html cmp exit 0
CompactHealth.dc.html cmp exit 0
CompactLogs.dc.html cmp exit 0
Components.dc.html cmp exit 0
Friction.dc.html cmp exit 0
FrictionResolved.dc.html cmp exit 0
Health.dc.html cmp exit 0
HealthAgents.dc.html cmp exit 0
Help.dc.html cmp exit 0
Logs.dc.html cmp exit 0
Main.dc.html cmp exit 0
TooSmall.dc.html cmp exit 0
Work.dc.html cmp exit 0
WorkGoals.dc.html cmp exit 0
WorkQueue.dc.html cmp exit 0
```

### V1 capture rerun

```text
AgentDetails copy capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:30.813: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

AgentDetails approved capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:32.935: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

AgentDetails identical: True
AgentsMany copy capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:35.106: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

AgentsMany approved capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:37.790: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

AgentsMany identical: True
Chat copy capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:40.487: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

Chat approved capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:42.963: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

Chat identical: True
CompactChat copy capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:45.460: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

CompactChat approved capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:48.487: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

CompactChat identical: True
CompactLogs copy capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:51.075: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

CompactLogs approved capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:53.857: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

CompactLogs identical: True
HealthAgents copy capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:56.492: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

HealthAgents approved capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:03:58.833: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

HealthAgents identical: True
Logs copy capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:04:01.243: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

Logs approved capture exit 0  *** You are running in headless mode.
Gtk-Message: 20:04:03.693: Not loading module "atk-bridge": The functionality is provided by GTK natively. Please try to not load it.

Logs identical: True
```

### V2 each changed cell

```text
== AgentDetails.grid: 4 changed cells (zero-based)
row 28 col 34: want '5' fg normal, got '4' fg normal
row 29 col 34: want '5' dim normal, got '4' dim normal
row 30 col 34: want '4' dim normal, got '3' dim normal
row 31 col 34: want '4' dim normal, got '3' dim normal
== AgentsMany.grid: 38 changed cells (zero-based)
row 6 col 33: want '[' dim normal, got 'g' dim normal
row 6 col 34: want 'g' dim normal, got '4' dim normal
row 6 col 35: want 'o' dim normal, got '-' dim normal
row 6 col 36: want 'a' dim normal, got 'r' dim normal
row 6 col 37: want 'l' dim normal, got 'a' dim normal
row 6 col 38: want ']' dim normal, got 't' dim normal
row 6 col 39: want ' ' dim normal, got 'e' dim normal
row 6 col 40: want '·' dim normal, got '-' dim normal
row 6 col 41: want ' ' dim normal, got 'l' dim normal
row 6 col 42: want 'w' dim normal, got 'i' dim normal
row 6 col 43: want 'a' dim normal, got 'm' dim normal
row 6 col 44: want 'v' dim normal, got 'i' dim normal
row 6 col 45: want 'e' dim normal, got 't' dim normal
row 6 col 46: want ' ' dim normal, got 's' dim normal
row 6 col 47: want 'B' dim normal, got ' ' dim normal
row 6 col 48: want ' ' dim normal, got '·' dim normal
row 6 col 50: want ' ' dim normal, got 'w' dim normal
row 6 col 51: want ' ' dim normal, got 'a' dim normal
row 6 col 52: want ' ' dim normal, got 'v' dim normal
row 6 col 53: want ' ' dim normal, got 'e' dim normal
row 6 col 55: want ' ' dim normal, got 'B' dim normal
row 6 col 91: want '[' dim normal, got 'g' dim normal
row 6 col 92: want 'g' dim normal, got '2' dim normal
row 6 col 93: want 'o' dim normal, got '-' dim normal
row 6 col 94: want 'a' dim normal, got 'g' dim normal
row 6 col 95: want 'l' dim normal, got 'u' dim normal
row 6 col 96: want ']' dim normal, got 'i' dim normal
row 6 col 97: want ' ' dim normal, got 'd' dim normal
row 6 col 98: want '·' dim normal, got 'e' dim normal
row 6 col 99: want ' ' dim normal, got 's' dim normal
row 6 col 100: want 'w' dim normal, got ' ' dim normal
row 6 col 101: want 'a' dim normal, got '·' dim normal
row 6 col 102: want 'v' dim normal, got ' ' dim normal
row 6 col 103: want 'e' dim normal, got 'w' dim normal
row 6 col 104: want ' ' dim normal, got 'a' dim normal
row 6 col 105: want 'A' dim normal, got 'v' dim normal
row 6 col 106: want ' ' dim normal, got 'e' dim normal
row 6 col 108: want ' ' dim normal, got 'A' dim normal
== Chat.grid: 1 changed cells (zero-based)
row 27 col 6: want '4' dim normal, got '2' dim normal
== Compact.grid: 0 changed cells (zero-based)
== CompactAgents.grid: 0 changed cells (zero-based)
== CompactChat.grid: 1 changed cells (zero-based)
row 17 col 6: want '4' dim normal, got '2' dim normal
== CompactFriction.grid: 0 changed cells (zero-based)
== CompactHealth.grid: 0 changed cells (zero-based)
== CompactLogs.grid: 20 changed cells (zero-based)
row 4 col 16: want 'a' dim normal, got 'p' dim normal
row 4 col 17: want 'g' dim normal, got 'r' dim normal
row 4 col 18: want 'e' dim normal, got 'o' dim normal
row 4 col 19: want 'n' dim normal, got 'j' dim normal
row 4 col 20: want 't' dim normal, got ':' dim normal
row 4 col 21: want ':' dim normal, got 'h' fg normal
row 4 col 22: want 'm' fg normal, got 'i' fg normal
row 4 col 23: want 'a' fg normal, got 'v' fg normal
row 4 col 24: want 'n' fg normal, got 'e' fg normal
row 4 col 25: want 'a' fg normal, got ' ' fg normal
row 4 col 26: want 'g' fg normal, got ' ' fg normal
row 4 col 27: want 'e' fg normal, got ' ' fg normal
row 4 col 28: want 'r' fg normal, got ' ' fg normal
row 4 col 29: want '-' fg normal, got ' ' fg normal
row 4 col 30: want 'h' fg normal, got ' ' fg normal
row 4 col 31: want 'i' fg normal, got ' ' fg normal
row 4 col 32: want 'v' fg normal, got ' ' fg normal
row 4 col 33: want 'e' fg normal, got ' ' fg normal
row 6 col 10: want '4' accent bold, got '3' accent bold
row 8 col 25: want '4' fg normal, got '3' fg normal
== Components.grid: 0 changed cells (zero-based)
== Friction.grid: 0 changed cells (zero-based)
== FrictionResolved.grid: 0 changed cells (zero-based)
== Health.grid: 0 changed cells (zero-based)
== HealthAgents.grid: 1 changed cells (zero-based)
row 4 col 24: want '2' accent bold, got '1' accent bold
== Help.grid: 0 changed cells (zero-based)
== Logs.grid: 36 changed cells (zero-based)
row 4 col 16: want 'a' dim normal, got 'p' dim normal
row 4 col 17: want 'g' dim normal, got 'r' dim normal
row 4 col 18: want 'e' dim normal, got 'o' dim normal
row 4 col 19: want 'n' dim normal, got 'j' dim normal
row 4 col 20: want 't' dim normal, got ':' dim normal
row 4 col 21: want ':' dim normal, got 'h' fg normal
row 4 col 22: want 'm' fg normal, got 'i' fg normal
row 4 col 23: want 'a' fg normal, got 'v' fg normal
row 4 col 24: want 'n' fg normal, got 'e' fg normal
row 4 col 25: want 'a' fg normal, got ' ' fg normal
row 4 col 26: want 'g' fg normal, got ' ' fg normal
row 4 col 27: want 'e' fg normal, got ' ' fg normal
row 4 col 28: want 'r' fg normal, got ' ' fg normal
row 4 col 29: want '-' fg normal, got ' ' fg normal
row 4 col 30: want 'h' fg normal, got ' ' fg normal
row 4 col 31: want 'i' fg normal, got ' ' fg normal
row 4 col 32: want 'v' fg normal, got ' ' fg normal
row 4 col 33: want 'e' fg normal, got ' ' fg normal
row 6 col 74: want '4' accent bold, got '3' accent bold
row 8 col 5: want '4' dim normal, got '3' dim normal
row 9 col 5: want '4' dim normal, got '3' dim normal
row 9 col 89: want '4' fg normal, got '3' fg normal
row 10 col 5: want '4' dim normal, got '3' dim normal
row 11 col 5: want '4' dim normal, got '3' dim normal
row 12 col 5: want '4' dim normal, got '3' dim normal
row 13 col 5: want '4' dim normal, got '3' dim normal
row 14 col 5: want '4' dim normal, got '3' dim normal
row 15 col 5: want '4' dim normal, got '3' dim normal
row 16 col 5: want '4' accent normal, got '3' accent normal
row 17 col 5: want '4' dim normal, got '3' dim normal
row 18 col 5: want '4' dim normal, got '3' dim normal
row 19 col 5: want '4' dim normal, got '3' dim normal
row 20 col 5: want '4' dim normal, got '3' dim normal
row 21 col 5: want '4' dim normal, got '3' dim normal
row 22 col 5: want '4' dim normal, got '3' dim normal
row 23 col 5: want '4' dim normal, got '3' dim normal
== Main.grid: 0 changed cells (zero-based)
== TooSmall.grid: 0 changed cells (zero-based)
== Work.grid: 0 changed cells (zero-based)
== WorkGoals.grid: 0 changed cells (zero-based)
== WorkQueue.grid: 0 changed cells (zero-based)
```

### V2 fresh exporter

```text
```

### V3 fixture red run

```text
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Main
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/AgentsMany
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/AgentDetails
    demo_fixtures_test.go:37: fixture: open ../../.hive/spec/tui-v2/design/AgentDetails.fixture.json: no such file or directory
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactAgents
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Chat
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactChat
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/WorkGoals
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/WorkQueue
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Work
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Compact
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Friction
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/FrictionResolved
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactFriction
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Health
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/HealthAgents
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactHealth
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Logs
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactLogs
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Help
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/TooSmall
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Components
--- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize (0.02s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Main (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/AgentsMany (0.00s)
    --- FAIL: TestEachBoardFixtureLoadsAtItsSpecSize/AgentDetails (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactAgents (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Chat (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactChat (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/WorkGoals (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/WorkQueue (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Work (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Compact (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Friction (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/FrictionResolved (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactFriction (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Health (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/HealthAgents (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactHealth (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Logs (0.01s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactLogs (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Help (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/TooSmall (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Components (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui	0.026s
FAIL
EXIT: 1
```

### V3 fixture green run

```text
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Main
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/AgentsMany
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/AgentDetails
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactAgents
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Chat
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactChat
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/WorkGoals
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/WorkQueue
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Work
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Compact
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Friction
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/FrictionResolved
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactFriction
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Health
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/HealthAgents
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactHealth
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Logs
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/CompactLogs
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Help
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/TooSmall
=== RUN   TestEachBoardFixtureLoadsAtItsSpecSize/Components
--- PASS: TestEachBoardFixtureLoadsAtItsSpecSize (0.01s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Main (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/AgentsMany (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/AgentDetails (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactAgents (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Chat (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactChat (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/WorkGoals (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/WorkQueue (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Work (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Compact (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Friction (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/FrictionResolved (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactFriction (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Health (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/HealthAgents (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactHealth (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Logs (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/CompactLogs (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Help (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/TooSmall (0.00s)
    --- PASS: TestEachBoardFixtureLoadsAtItsSpecSize/Components (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	0.020s
EXIT: 0
```

### V3 independent semantic checks

```text
PASS AgentDetails.fixture.json has no w58:p3
PASS AgentDetails.fixture.json coordinator pane w58:p4
PASS AgentsMany.fixture.json has no w58:p3
PASS AgentsMany.fixture.json coordinator pane w58:p4
PASS Chat.fixture.json has no w58:p3
PASS Chat.fixture.json coordinator pane w58:p4
PASS Compact.fixture.json has no w58:p3
PASS Compact.fixture.json coordinator pane w58:p4
PASS CompactAgents.fixture.json has no w58:p3
PASS CompactAgents.fixture.json coordinator pane w58:p4
PASS CompactChat.fixture.json has no w58:p3
PASS CompactChat.fixture.json coordinator pane w58:p4
PASS CompactFriction.fixture.json has no w58:p3
PASS CompactFriction.fixture.json coordinator pane w58:p4
PASS CompactHealth.fixture.json has no w58:p3
PASS CompactHealth.fixture.json coordinator pane w58:p4
PASS CompactLogs.fixture.json has no w58:p3
PASS CompactLogs.fixture.json coordinator pane w58:p4
PASS Components.fixture.json has no w58:p3
PASS Components.fixture.json coordinator pane w58:p4
PASS Friction.fixture.json has no w58:p3
PASS Friction.fixture.json coordinator pane w58:p4
PASS FrictionResolved.fixture.json has no w58:p3
PASS FrictionResolved.fixture.json coordinator pane w58:p4
PASS Health.fixture.json has no w58:p3
PASS Health.fixture.json coordinator pane w58:p4
PASS HealthAgents.fixture.json has no w58:p3
PASS HealthAgents.fixture.json coordinator pane w58:p4
PASS Help.fixture.json has no w58:p3
PASS Help.fixture.json coordinator pane w58:p4
PASS Logs.fixture.json has no w58:p3
PASS Logs.fixture.json coordinator pane w58:p4
PASS Main.fixture.json has no w58:p3
PASS Main.fixture.json coordinator pane w58:p4
PASS TooSmall.fixture.json has no w58:p3
PASS TooSmall.fixture.json coordinator pane w58:p4 (no coordinator record)
PASS Work.fixture.json has no w58:p3
PASS Work.fixture.json coordinator pane w58:p4
PASS WorkGoals.fixture.json has no w58:p3
PASS WorkGoals.fixture.json coordinator pane w58:p4
PASS WorkQueue.fixture.json has no w58:p3
PASS WorkQueue.fixture.json coordinator pane w58:p4
PASS AgentDetails four fixture RECENT times
PASS AgentDetails four times in board script line 60
PASS example-api fixture goal g4-rate-limits wave B
PASS example-api board goal matches fixture
PASS example-docs fixture goal g2-guides wave A
PASS example-docs board goal matches fixture
PASS Logs filter text
PASS Logs board filter text and tokens line 44
PASS Logs final 16 fixture event times match independent expected list
PASS Logs board event times lines 48-63 match expected list
PASS Logs selection timestamp
PASS Logs popup and JSON timestamp lines 74-75
PASS Logs all 312 events one hour earlier (312 changed)
PASS CompactLogs all 312 events one hour earlier (312 changed)
```

### V-gate diff scope

```text
 .hive/spec/tui-v2/design/AgentDetails.dc.html      |   2 +-
 .hive/spec/tui-v2/design/AgentDetails.fixture.json |  10 +-
 .hive/spec/tui-v2/design/AgentDetails.grid         |   8 +-
 .hive/spec/tui-v2/design/AgentsMany.dc.html        |   4 +-
 .hive/spec/tui-v2/design/AgentsMany.fixture.json   |   6 +-
 .hive/spec/tui-v2/design/AgentsMany.grid           |   2 +-
 .hive/spec/tui-v2/design/Chat.dc.html              |   2 +-
 .hive/spec/tui-v2/design/Chat.fixture.json         |   4 +-
 .hive/spec/tui-v2/design/Chat.grid                 |   2 +-
 .hive/spec/tui-v2/design/Compact.fixture.json      |   2 +-
 .../spec/tui-v2/design/CompactAgents.fixture.json  |   2 +-
 .hive/spec/tui-v2/design/CompactChat.dc.html       |   2 +-
 .hive/spec/tui-v2/design/CompactChat.fixture.json  |   4 +-
 .hive/spec/tui-v2/design/CompactChat.grid          |   2 +-
 .../tui-v2/design/CompactFriction.fixture.json     |   2 +-
 .hive/spec/tui-v2/design/CompactLogs.dc.html       |   6 +-
 .hive/spec/tui-v2/design/CompactLogs.fixture.json  | 188 ++++++++---------
 .hive/spec/tui-v2/design/CompactLogs.grid          |   8 +-
 .hive/spec/tui-v2/design/Components.fixture.json   |   2 +-
 .hive/spec/tui-v2/design/Friction.fixture.json     |   2 +-
 .../tui-v2/design/FrictionResolved.fixture.json    |   2 +-
 .hive/spec/tui-v2/design/Health.fixture.json       |   2 +-
 .hive/spec/tui-v2/design/HealthAgents.dc.html      |   2 +-
 .hive/spec/tui-v2/design/HealthAgents.fixture.json |   2 +-
 .hive/spec/tui-v2/design/HealthAgents.grid         |   2 +-
 .hive/spec/tui-v2/design/Help.fixture.json         |   2 +-
 .hive/spec/tui-v2/design/Logs.dc.html              |  38 ++--
 .hive/spec/tui-v2/design/Logs.fixture.json         | 188 ++++++++---------
 .hive/spec/tui-v2/design/Logs.grid                 |  38 ++--
 .hive/spec/tui-v2/design/Main.fixture.json         |   2 +-
 .hive/spec/tui-v2/design/Work.fixture.json         |   2 +-
 .hive/spec/tui-v2/design/WorkGoals.fixture.json    |   2 +-
 .hive/spec/tui-v2/design/WorkQueue.fixture.json    |   2 +-
 .hive/spec/tui-v2/tasks/28-board-data-fix/brief.md | 134 +++++++++++++
 .../tasks/28-board-data-fix/evidence/V-gate.md     | 223 +++++++++++++++++++++
 .../tui-v2/tasks/28-board-data-fix/evidence/V1.md  |  73 +++++++
 .../tui-v2/tasks/28-board-data-fix/evidence/V2.md  | 113 +++++++++++
 .../tui-v2/tasks/28-board-data-fix/evidence/V3.md  |  85 ++++++++
 internal/tui/component/chat_test.go                |   4 +-
 internal/tui/component/popup_test.go               |   4 +-
 testdata/ste-baseline.json                         |   1 +
 41 files changed, 905 insertions(+), 276 deletions(-)
```

### V-gate full output

```text
$ gofmt -l $(git ls-files -co --exclude-standard "*.go")
EXIT: 0
$ go mod tidy -diff
EXIT: 0
$ go vet ./...
EXIT: 0
$ golangci-lint run ./...
0 issues.
EXIT: 0
$ go test -race ./...
ok  	github.com/Harrison-Blair/hive	28.737s
ok  	github.com/Harrison-Blair/hive/cmd	7.582s
?   	github.com/Harrison-Blair/hive/cmd/agent	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/list	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/register	1.208s
ok  	github.com/Harrison-Blair/hive/cmd/agent/resolve	1.498s
?   	github.com/Harrison-Blair/hive/cmd/agent/send	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/agent/show	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/agent/spawn	1.455s
?   	github.com/Harrison-Blair/hive/cmd/agent/stop	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/chat	1.155s
?   	github.com/Harrison-Blair/hive/cmd/common	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/doctor	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/goal	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/handoff	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/handoff/write	2.448s
?   	github.com/Harrison-Blair/hive/cmd/homeinit	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/inbox	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/land	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/logs	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/memory	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/memory/write	2.375s
?   	github.com/Harrison-Blair/hive/cmd/project	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/register	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/project/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/supervise	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task	1.373s
?   	github.com/Harrison-Blair/hive/cmd/task/assign	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/cancel	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/complete	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/task/create	1.163s
ok  	github.com/Harrison-Blair/hive/cmd/task/edit	1.162s
?   	github.com/Harrison-Blair/hive/cmd/task/list	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/reopen	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/show	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/task/verify	[no test files]
ok  	github.com/Harrison-Blair/hive/cmd/version	1.036s
?   	github.com/Harrison-Blair/hive/cmd/worktree	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/create	[no test files]
?   	github.com/Harrison-Blair/hive/cmd/worktree/remove	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/agent	16.615s
ok  	github.com/Harrison-Blair/hive/internal/boardexport	1.691s
?   	github.com/Harrison-Blair/hive/internal/boardexport/cmd/boardexport	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/chat	(cached)
ok  	github.com/Harrison-Blair/hive/internal/doctor	6.002s
?   	github.com/Harrison-Blair/hive/internal/goal	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/handoff	(cached)
ok  	github.com/Harrison-Blair/hive/internal/homeinit	(cached)
ok  	github.com/Harrison-Blair/hive/internal/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/land	5.512s
ok  	github.com/Harrison-Blair/hive/internal/lib/cellgrid	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/gitx	1.414s
ok  	github.com/Harrison-Blair/hive/internal/lib/harness	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/home	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/identity	1.386s
ok  	github.com/Harrison-Blair/hive/internal/lib/inbox	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/log	(cached)
?   	github.com/Harrison-Blair/hive/internal/lib/marker	[no test files]
?   	github.com/Harrison-Blair/hive/internal/lib/output	[no test files]
ok  	github.com/Harrison-Blair/hive/internal/lib/panelock	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/process	1.040s
ok  	github.com/Harrison-Blair/hive/internal/lib/records	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/run	4.427s
ok  	github.com/Harrison-Blair/hive/internal/lib/scope	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/store	(cached)
ok  	github.com/Harrison-Blair/hive/internal/lib/version	(cached)
ok  	github.com/Harrison-Blair/hive/internal/logs	(cached)
ok  	github.com/Harrison-Blair/hive/internal/memory	(cached)
ok  	github.com/Harrison-Blair/hive/internal/project	(cached)
ok  	github.com/Harrison-Blair/hive/internal/supervise	(cached)
ok  	github.com/Harrison-Blair/hive/internal/task	11.915s
ok  	github.com/Harrison-Blair/hive/internal/tui	76.629s
ok  	github.com/Harrison-Blair/hive/internal/tui/component	(cached)
ok  	github.com/Harrison-Blair/hive/internal/worktree	7.784s
EXIT: 0
```

### Golden-test output

```text
=== RUN   TestChatGolden
    chat_test.go:45: 1 cells differ from Chat.grid (board row and column):
        row 27 col 6: want '4' dim normal, got '2' dim normal
--- FAIL: TestChatGolden (0.01s)
=== RUN   TestCompactChatGolden
    chat_test.go:73: 1 cells differ from CompactChat.grid (board row and column):
        row 17 col 6: want '4' dim normal, got '2' dim normal
--- FAIL: TestCompactChatGolden (0.00s)
=== RUN   TestPopupGolden
    popup_test.go:31: 4 cells differ from AgentDetails.grid (board row and column):
        row 28 col 34: want '5' fg normal, got '4' fg normal
        row 29 col 34: want '5' dim normal, got '4' dim normal
        row 30 col 34: want '4' dim normal, got '3' dim normal
        row 31 col 34: want '4' dim normal, got '3' dim normal
--- FAIL: TestPopupGolden (0.00s)
FAIL
FAIL	github.com/Harrison-Blair/hive/internal/tui/component	0.053s
FAIL
EXIT: 1
=== RUN   TestChatGolden
--- PASS: TestChatGolden (0.01s)
=== RUN   TestCompactChatGolden
--- PASS: TestCompactChatGolden (0.00s)
=== RUN   TestPopupGolden
--- PASS: TestPopupGolden (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui/component	0.044s
EXIT: 0
```
