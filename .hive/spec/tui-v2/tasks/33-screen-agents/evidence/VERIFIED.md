# Verification: 33-screen-agents

Verifier: `t33-verifier-3`. Task commit: `4b7007774f94d99e8658238430f534819f15a10e`.

## Verdict: fail

The required merge at the start of V-gate failed. The merge target was
`origin/dev` at `d3e1aa2c4b2f0a63cef6e417ca86a40bedbff6e4`.
`internal/tui/shell_test.go` and `testdata/ste-baseline.json` have conflicts.
The verifier aborted the merge. A worker must repair the task. This verdict
does not depend on a later change to `dev`.

The selection repair passes the plain-text check. The selected card has a
heavy border. The idle symbol does not serve as the selection signal.

## Items

| Item | Result | Observation |
| --- | --- | --- |
| V1 | fail | Incomplete base-code proof. The independent green run, full Help check, and deliberate token and key breaks behaved as expected. The complete new screen test file cannot compile on base because the graph state is absent. Per-test behavioral red proof remains incomplete. |
| V2 | pass | All five new live captures have zero differing cells under the space-cell rule. The key checks and plain-text selection check pass. |
| V-gate | fail | The required merge has two conflicts. All five gate commands pass on the unmerged task commit. No pass of the integrated code is claimed. |

## V1: independent runs

Full output: `verify3/green.log`, `base-red.log`, `help-red.log`,
`token-red.log`, and `key-red.log`.

Green command:

```sh
go test ./internal/tui ./internal/tui/component -run 'TestAgents|TestHelpMatchesItsBoard|TestCard|TestFitWave4' -count=1 -v
```

Exit 0. This includes all four Agents golden checks, the full Help grid,
35 responsive cases, movement, popup, action, footer, and scratch-home
normal-mode record and refresh tests. It also includes the heavy card and
mixed-weight joint checks.

Base run: a scratch worktree at local `dev`, `f1847c7b`, received the new
screen tests, the frame tests, and the two changed component test files.
The screen package failed to compile: `m.graph` and `agentsNoneText` are
absent. The component tests failed for behavior: the base draws rounded
selected borders instead of heavy borders. With the new screen test file
removed, the full Help test failed with 294 cells that differ. A later
attempt to isolate the base screen tests had unused imports; it supplies
no additional proof. No behavioral base proof is claimed for each test.

Deliberate breaks were made only in a second scratch worktree of the task:

- Idle dot: `accentStyle.Render("◐")` changed to `dimStyle.Render("◐")`.
  All four Agents golden cases failed. The project-agent test reported a
  dim idle dot instead of accent. The component golden test also failed.
- Hide key: `case "h":` changed to `case "H":`.
  `TestAgentsHHidesStoppedAgents` failed: `h` did not hide stopped agents.

Both files were restored after their red run. No task source changed.

## V2: independent live run

Built with `go build -o .hive/tmp/verify33/hive .`. Herdr client and server
were 0.9.5, protocol 22. The verifier made scratch tab `wAG:t2`, pane
`wAG:p2`, in its own workspace. The pane had 57 rows and no scroll offset.
The command for each board was:

```sh
HIVE_HOME=SCRATCH_HOME COLORTERM=truecolor SCRATCH_BINARY --demo WORKTREE/.hive/spec/tui-v2/design/BOARD.fixture.json --size WxH
```

Demo mode implies no startup. The scratch home contained a home marker.
No real home record was used by these runs. Sizes: CompactAgents 80×24;
the other four boards 120×40.

Each capture used `herdr pane read wAG:p2 --source visible --format ansi`.
Two reads 500 ms apart had to be equal. The captured viewport had all
required rows; pane metadata records offset 0. The CLI returns ANSI text
rather than a JSON read envelope with a truncation field. There was no
truncation notice. The converter cropped the top-left board rectangle.
The comparison used `internal/lib/cellgrid` and ignored token and weight
only when both cells were spaces.

| Board | Strict differences | Space-cell differences | Live-rule differences |
| --- | --- | --- | --- |
| Main | 3438 | 3438 | 0 |
| AgentsMany | 2193 | 2193 | 0 |
| AgentDetails | 942 | 942 | 0 |
| CompactAgents | 779 | 779 | 0 |
| Help | 1420 | 1420 | 0 |

`verify3/` holds each ANSI capture, pane metadata, screen PNG, new board
PNG, and colour-free plain text. The comparison program is the existing
`evidence/screens/compare-program.go.txt`, rebuilt independently.
`verify3/live.log` records the comparisons and key assertions. Its final
Firefox error was a missing scratch profile directory. The verifier then
created new profile directories and captured all twelve PNG files.
The PNGs were opened and compared. Font metrics and browser canvas edges
differ; terminal cells agree.

The new board captures used each absolute board path from the brief and:

```sh
firefox --headless --no-remote --profile SCRATCH --window-size=1220,780 --screenshot OUT.png file://BOARD.dc.html
```

Screen captures used a scratch ANSI-to-HTML conversion and the same
Firefox command. No colour was used to check the `.plain.txt` files.

### Keys

- Main: `h` hides two stopped agents; a second `h` shows them. `r` leaves
  the graph unchanged. Down and right move the heavy selection border to
  the two children. Enter opens details. `r` also leaves details unchanged.
  Escape closes details. Attach and peek show `demo mode: no action`.
  `t` opens task `4-agents-equal-columns` in Work Tasks.
- AgentsMany: `k` selects the manager with joint `┡`. `j` and four right
  keys show projects 2–5. Four left keys restore projects 1–4.
- AgentDetails: arrow and page input are accepted; escape closes details.
  Herdr rejected `pgdn` and `pagedown` as key names before the final run.
  The final run sent the terminal page sequences with `pane send-text`.
- CompactAgents: enter opens details; escape closes them; `?` opens help;
  escape restores graph focus and its compact hints.
- Help: arrow input stays in help. Both `?` and escape close help.
- The first Ctrl+C shows the quit notice. The second returns to the shell.

### Visual check

| Board element | state without colour | Result | Cause or decision |
| --- | --- | --- | --- |
| Header | coordinator working, supervisor live, needs you count, clock | matches | Main, AgentsMany, AgentDetails, Help |
| Compact header | coord, sup, needs you, clock | matches | CompactAgents |
| Active tab | `1 Agents` with `━` underline | matches | All five boards |
| Outer graph title | `project agents · N projects` | matches | Singular form on Main and CompactAgents |
| Pod title | Project name | matches | One or four visible equal-width pods |
| Goal line | Goal and status or wave | matches | Task 33 manager decisions |
| Needs-user state | `needs you · approve mer…` | matches | AgentsMany |
| Working agent | `●`, `working` | matches | AgentsMany |
| Idle agent | `◐`, `idle` | matches | Accent token also checked by cell comparison |
| Stopped agent | `○`, `stopped`, result | matches | Main and AgentsMany |
| Selected card | `┏━┓┃┗┛` | matches | User option A; visible in colour-free capture |
| Tree manager joint | `┯` | matches | Main and CompactAgents |
| Tree child joint | `┷` | matches | New main-down capture |
| Rail child joint | `┨` | matches | AgentsMany |
| Rail manager joint | `┡` | matches | New many-k capture |
| Tree connectors | `│`, bus, `┴` | matches | Main and CompactAgents |
| Rail connectors | `│`, `├┤`, `╰┤` | matches | AgentsMany |
| No children | `no workers` | matches | AgentsMany docs pod |
| Stopped toggle | `h hides N stopped`, `h shows N stopped` | matches | Live hide and show check |
| Legend | Spawn and state glyphs with words | matches | Wide graphs |
| Horizontal scroll | Projects range, arrows, heavy bottom mark | matches | AgentsMany and right/left checks |
| Graph footer | `[graph]` and action hints | matches | Wide and compact hints |
| Details title and fields | Name, `esc closes`, role, project, status, record fields | matches | AgentDetails |
| Details sections | `ACTIONS`, `RECENT` | matches | AgentDetails |
| Details footer | `[details]`, attach, peek, task, scroll, close | matches | Wide and compact popups |
| Help headings and map | `GLOBAL`, `ACTIONS`, `FILTER FIELDS`, `CHAT` | matches | Full Help board |
| Help footer | `[help]`, scroll and close hints | matches | Help |
| Screen behind popup or help | Graph text remains; focus word changes to details or help | matches | Full screen has the board dim token |

## V-gate

`verify3/merge.log` contains the failed integration attempt. The verifier
fetched `origin/dev` once and ran `git merge --no-edit origin/dev` once.
It did not resolve source conflicts. `git merge --abort` restored the task.
`verify3/scope.log` is the diff stat against that fetched target. It has
only the allowed source paths and task evidence, but it is not a completed
merge result.

`verify3/gate.log` contains all output and exit statuses of the five gate
commands on the restored task commit:

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go mod tidy -diff
go vet ./...
golangci-lint run ./...
go test -race ./...
```

All exit 0. The first two print nothing. No flake rerun was required.
These results do not prove the code after a conflict repair.

## Cleanup

The scratch live tab was closed after its foreground process returned to
bash. The workspace again has only verifier tab `wAG:t1`. Both scratch Git
worktrees were removed. The temporary binary, comparison program, browser
profiles, scratch home, scripts, and extra logs were removed after the
results were copied here. Final captures and logs remain in `verify3/`.
No worker evidence file or plan tracker changed.
