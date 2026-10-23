# Review cases of the approval commit 219eddba

The manager asked for two cases after the review of `219eddba`
(inbox record 533). This file records the scratch test, its output, the spec
rule, and the analysis. The source, the approval grids, the golden files, and
the board sources did not change. The decision is open.

Task source: commit `c970d9a7`. Command:
`go test ./internal/tui/ -run TestZZCases -count=1 -v` (exit 0; the test
only prints the views). The test used the helpers of `agents_hide_test.go`:
projects `alpha` and `beta`, a live coordinator, a live supervisor claim.
The scratch test file is removed after this record.

## Case 1: every project pod hides, a live home agent stays

Records: `manager-alpha` and `a1-worker` stopped, `manager-beta` stopped,
`helper` (role `agent`, parent `coordinator`) live.

Observed at 120x40 and 80x24: `agentsPods` = `[home]`. The title is
`project agents · 0 projects · h shows 3 stopped`. The body shows only
`no project agents`. The coordinator card and the live home card are not
drawn. The footer still shows `←→ select pod · ↑ coordinator`.

Spec: decision 10 of `tui-v2`: "With no project, Agents shows the notice
'no project agents'." Here 2 projects are registered, and a home card is
visible. The code tests `projects == 0`, the count of visible project pods.
On the base, this state does not occur, because a stopped manager keeps its
pod.

Analysis: AC2 removes a pod only when it has no visible card. The home pod
has a visible card. AC3 keeps the hint "when every project pod disappears".
In this reading, AC2 and AC3 expect the overview with the coordinator and
home, and the hint on the title. Recommendation: show the notice only when
no card is visible, or only when no project is registered.

## Case 2: a hidden manager with one direct live child of role agent

Records: `manager-beta` stopped; `beta-helper` (role `agent`, project
`beta`, parent `manager-beta`) live; no beta worker.

Observed at 120x40 and 80x24: pods = `alpha[manager-alpha a1-worker]`,
`home[beta-helper]`. The beta pod is gone. `beta-helper` is in home. The
title is `project agents · 1 project · h shows 1 stopped`.

Cause: `allPods` puts no `podOf` entry for a hidden manager, so `place()`
sends its general child to home.

Spec: decision 39 of `general-agent`: a general agent is under the card of
its parent when that card is on the graph; each other general agent is in
`home`. Literally, a hidden parent sends the child to home (the
`AgentsHiddenParent` rule for a hidden worker parent).

Analysis: AC2 keeps the pod border and title when a hidden manager has a
live child. Ask `a3` (answer "Hide manager: the children hang from the pod
top border") has no role limit. AC2 requires a change: a direct child of a
hidden manager, of each role, stays a top card of the project pod. A general
agent with a hidden worker parent keeps decision 39 and goes to home.

## Scratch test

```go
package tui

import (
	"fmt"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

func TestZZCases(t *testing.T) {
	// Case 1: every project pod hides, a live home agent stays.
	home, clock := hideHome(t, betaManager(records.AgentStopped))
	writeAgentRecord(t, home, records.Agent{Key: "manager-alpha", Name: "manager-alpha", Role: "manager", Project: "alpha", Harness: "claude", Status: records.AgentStopped, Parent: "coordinator", StartedAt: "2026-10-23T10:01:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "a1-worker", Name: "a1-worker", Role: "worker", Project: "alpha", Harness: "pi", Status: records.AgentStopped, Parent: "manager-alpha", StartedAt: "2026-10-23T10:02:00Z"})
	writeAgentRecord(t, home, records.Agent{Key: "helper", Name: "helper", Role: "agent", Harness: "pi", Status: records.AgentLive, AgentStatus: "idle", Parent: "coordinator", StartedAt: "2026-10-23T10:07:00Z"})
	for _, size := range hideSizes {
		m := openHide(t, home, clock, size)
		var ids []string
		for _, p := range m.agentsPods() {
			ids = append(ids, p.group.id)
		}
		fmt.Printf("=== case 1 %dx%d pods %v\n%s\n", size.Width, size.Height, ids, ansi.Strip(m.View().Content))
	}
	// Case 2: hidden manager with one direct live child of role agent.
	home, clock = hideHome(t, betaManager(records.AgentStopped))
	writeAgentRecord(t, home, records.Agent{Key: "beta-helper", Name: "beta-helper", Role: "agent", Project: "beta", Harness: "claude", Status: records.AgentLive, AgentStatus: "working", Parent: "manager-beta", StartedAt: "2026-10-23T10:06:00Z"})
	for _, size := range hideSizes {
		m := openHide(t, home, clock, size)
		var ids []string
		for _, p := range m.agentsPods() {
			var keys []string
			for _, n := range p.nodes() {
				keys = append(keys, n.record.Key)
			}
			ids = append(ids, fmt.Sprintf("%s%v", p.group.id, keys))
		}
		fmt.Printf("=== case 2 %dx%d pods %v\n%s\n", size.Width, size.Height, ids, ansi.Strip(m.View().Content))
	}
}
```

## Output

The temporary home path is replaced with `/tmp/HOME`. Blank body rows are
kept.

```text
=== RUN   TestZZCases
=== case 1 120x40 pods [home]
 hive   /tmp/HOME                 coordinator ○ stopped   supervisor ○ stopped   ? 0 asks   needs you 0   08:00 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────────────────────────────────────────────
╭─ project agents · 0 projects · h shows 3 stopped ────────────────────────────────────────────────────────────────────╮
│                                                                                                                      │
│  no project agents                                                                                                   │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
 [graph] ←→ select pod · ↑ coordinator · enter open pod · h stopped · ? help                                            
=== case 1 80x24 pods [home]
 hive                               coord ○ · sup ○ · ? 0 · needs you 0 · 08:00 
  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs                        
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────
╭─ project agents · 0 projects · h shows 3 stopped ────────────────────────────╮
│                                                                              │
│  no project agents                                                           │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 [graph] ←→ select pod · ↑ coordinator · enter open · ?                         
=== case 2 120x40 pods [alpha[manager-alpha a1-worker] home[beta-helper]]
 hive   /tmp/HOME                 coordinator ○ stopped   supervisor ○ stopped   ? 0 asks   needs you 0   08:00 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────────────────────────────────────────────
╭─ project agents · 1 project · h shows 1 stopped ─────────────────────────────────────────────────────────────────────╮
│                                                                                                                      │
│                                               ╭──────────────────────╮                                               │
│                                               │ ● coordinator        │                                               │
│                                               │ claude · working · … │                                               │
│                                               ╰───────────┬──────────╯                                               │
│                             ╭─────────────────────────────┴───────────────────────────╮                              │
│ ┏━ alpha ━━━━━━━━━━━━━━━━━━━┷━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ╭─ home ────────────────────┴───────────────────────────╮  │
│ ┃                                                       ┃ │                                                       │  │
│ ┃               ╭──────────────────────╮                ┃ │ ╭──────────────────────╮                              │  │
│ ┃               │ ◐ manager-alpha      │                ┃ │ │ ● beta-helper        │                              │  │
│ ┃               │ claude · idle · 2m   │                ┃ │ │ claude · working     │                              │  │
│ ┃               ╰───────────┬──────────╯                ┃ │ ╰──────────────────────╯                              │  │
│ ┃                           │                           ┃ │                                                       │  │
│ ┃                           │                           ┃ │                                                       │  │
│ ┃                           │                           ┃ │                                                       │  │
│ ┃               ╭───────────┴──────────╮                ┃ │                                                       │  │
│ ┃               │ ● a1-worker          │                ┃ │                                                       │  │
│ ┃               │ pi · working · 2m    │                ┃ │                                                       │  │
│ ┃               ╰──────────────────────╯                ┃ │                                                       │  │
│ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ╰───────────────────────────────────────────────────────╯  │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│ ── spawn   ● working   ◐ idle   ○ stopped   ━ ▶ message, lit 5s   one column per project                             │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
 [graph] ←→ select pod · ↑ coordinator · enter open pod · h stopped · ? help                                            
=== case 2 80x24 pods [alpha[manager-alpha a1-worker] home[beta-helper]]
 hive                               coord ○ · sup ○ · ? 0 · needs you 0 · 08:00 
  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs                        
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────
╭─ project agents · 1 project · h shows 1 stopped ─────────────────────────────╮
│                           ╭──────────────────────╮                           ┃
│                           │ ● coordinator        │                           ┃
│                           │ claude · working · … │                           ┃
│                           ╰───────────┬──────────╯                           ┃
│                   ╭───────────────────┴─────────────────╮                    ┃
│ ┏━ alpha ━━━━━━━━━┷━━━━━━━━━━━━━━━━━┓ ╭─ home ──────────┴─────────────────╮  ┃
│ ┃                                   ┃ │                                   │  ┃
│ ┃     ╭──────────────────────╮      ┃ │ ╭──────────────────────╮          │  ┃
│ ┃     │ ◐ manager-alpha      │      ┃ │ │ ● beta-helper        │          │  ┃
│ ┃     │ claude · idle · 2m   │      ┃ │ │ claude · working     │          │  ┃
│ ┃     ╰───────────┬──────────╯      ┃ │ ╰──────────────────────╯          │  ┃
│ ┃                 │                 ┃ │                                   │  ┃
│ ┃                 │                 ┃ │                                   │  ┃
│ ┃                 │                 ┃ │                                   │  ┃
│ ┃     ╭───────────┴──────────╮      ┃ │                                   │  ┃
│ ┃     │ ● a1-worker          │      ┃ │                                   │  ┃
│ ┃     │ pi · working · 2m    │      ┃ │                                   │  ┃
│ ┃     ╰──────────────────────╯      ┃ │                                   │  │
╰──────────────────────────────────────────────────────────────────────────────╯
 [graph] ←→ select pod · ↑ coordinator · enter open · ?                         
--- PASS: TestZZCases (0.00s)
PASS
ok  	github.com/Harrison-Blair/hive/internal/tui	0.011s
```
