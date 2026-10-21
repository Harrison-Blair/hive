# 10-board-export grid review aid

The `t|` text of each golden grid, in the order of the board table of the spec. The person reviews this text at the end of wave 2. Each line is one row of cells. The token and bold lines are in the grid file.

## Main

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────────────────────────────────────────────
╭─ project agents ─────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                                                      │
│                           ╭─ hive · g1-tui-restart · done ───────────────────────────────╮                           │
│                           │                                                              │                           │
│                           │                   ╭──────────────────────╮                   │                           │
│                           │                   │ ◐ manager-hive       │                   │                           │
│                           │                   │ claude · idle · nud… │                   │                           │
│                           │                   ╰───────────┬──────────╯                   │                           │
│                           │                               │                              │                           │
│                           │                 ╭─────────────┴───────────────╮              │                           │
│                           │                 │                             │              │                           │
│                           │     ╭───────────┴──────────╮      ╭───────────┴──────────╮   │                           │
│                           │     │ ○ t4-worker-1        │      │ ○ t4-verifier-1      │   │                           │
│                           │     │ pi · stopped · task… │      │ pi · stopped · pass  │   │                           │
│                           │     ╰──────────────────────╯      ╰──────────────────────╯   │                           │
│                           │                                                              │                           │
│                           │ 2 stopped shown · h hides them                               │                           │
│                           ╰──────────────────────────────────────────────────────────────╯                           │
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
│ ── spawn   ● working   ◐ idle   ○ stopped   one column per project, all the same width                               │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
 [graph] ←↑↓→ move · enter details · a attach · p peek · t task · h hide stopped · ? help                               
```

## AgentsMany

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 1   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────────────────────────────────────────────
╭─ project agents · 5 projects ────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                                                      │
│ ╭─ hive ───────────────────╮ ╭─ example-api ────────────╮ ╭─ example-web ────────────╮ ╭─ example-docs ───────────╮  │
│ │ g2-tui-v2 · wave A       │ │ [goal] · wave B          │ │ needs you · approve mer… │ │ [goal] · wave A          │  │
│ │ ╭──────────────────────╮ │ │ ╭──────────────────────╮ │ │ ╭──────────────────────╮ │ │ ╭──────────────────────╮ │  │
│ │ │ ● manager-hive       │ │ │ │ ● manager-api        │ │ │ │ ◐ manager-web        │ │ │ │ ◐ manager-docs       │ │  │
│ │ │ claude · working · … │ │ │ │ codex · working · 2m │ │ │ │ claude · idle · 9m   │ │ │ │ pi · idle · 22m      │ │  │
│ │ ├──────────────────────╯ │ │ ├──────────────────────╯ │ │ ├──────────────────────╯ │ │ ╰──────────────────────╯ │  │
│ │ │                        │ │ │                        │ │ │                        │ │                          │  │
│ │ │╭──────────────────────╮│ │ │╭──────────────────────╮│ │ │╭──────────────────────╮│ │                          │  │
│ │ ├┤ ● t5-worker-1        ││ │ ├┤ ● t3-worker-1        ││ │ ╰┤ ○ t8-worker-1        ││ │  no workers              │  │
│ │ ││ pi · working · 3m    ││ │ ││ codex · working      ││ │  │ pi · stopped · done  ││ │                          │  │
│ │ │╰──────────────────────╯│ │ │╰──────────────────────╯│ │  ╰──────────────────────╯│ │                          │  │
│ │ │                        │ │ │                        │ │                          │ │                          │  │
│ │ │╭──────────────────────╮│ │ │╭──────────────────────╮│ │                          │ │                          │  │
│ │ ├┤ ● t6-worker-1        ││ │ ╰┤ ◐ t3-verifier-1      ││ │                          │ │                          │  │
│ │ ││ pi · working · 1m    ││ │  │ pi · idle · 6m       ││ │                          │ │                          │  │
│ │ │╰──────────────────────╯│ │  ╰──────────────────────╯│ │                          │ │                          │  │
│ │ │                        │ │                          │ │                          │ │                          │  │
│ │ │╭──────────────────────╮│ │                          │ │                          │ │                          │  │
│ │ ╰┤ ○ t4-verifier-1      ││ │                          │ │                          │ │                          │  │
│ │  │ pi · stopped · pass  ││ │                          │ │                          │ │                          │  │
│ │  ╰──────────────────────╯│ │                          │ │                          │ │                          │  │
│ │                          │ │                          │ │                          │ │                          │  │
│ │                          │ │                          │ │                          │ │                          │  │
│ ╰──────────────────────────╯ ╰──────────────────────────╯ ╰──────────────────────────╯ ╰──────────────────────────╯  │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│ ── spawn   ● working   ◐ idle   ○ stopped   columns share one width                                                  │
│ ←→ at the edge scrolls · projects 1–4 of 5                                                         example-ml →      │
╰━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━────────────────────────╯
 [graph] ←↑↓→ move · enter details · a attach · p peek · t task · h hide stopped · ? help                               
```

## AgentDetails

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────────────────────────────────────────────
╭─ project agents ─────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                                                      │
│                           ╭─ hive · g1-tui-restart · done ───────────────────────────────╮                           │
│                           │                                                              │                           │
│                           │ ╭─ manager-hive · esc closes ──────────────────────────────╮ │                           │
│                           │ │                                                          │ │                           │
│                           │ │  role      manager                                       │ │                           │
│                           │ │  project   hive                                          │ │                           │
│                           │ │  status    ◐ live · idle                                 │ │                           │
│                           │ │  nudged    4m ago by supervisor                          │ │                           │
│                           │ │  harness   claude                                        │ │                           │
│                           │ │  model     claude-opus-5-5                               │ │                           │
│                           │ │  pane      w58:p6                                        │ │                           │
│                           │ │  tab       manager-hive                                  │ │                           │
│                           │ │  goal      g1-tui-restart · done                         │ │                           │
│                           │ │  parent    coordinator-g2                                │ │                           │
│                           │ │  replaces  manager-hive-g1                               │ │                           │
│                           ╰─│                                                          │─╯                           │
│                             │  ACTIONS                                                 │                             │
│                             │  a  attach to pane                                       │                             │
│                             │  p  peek pane                                            │                             │
│                             │  t  open task · none                                     │                             │
│                             │                                                          │                             │
│                             │                                                          │                             │
│                             │  RECENT                                                  │                             │
│                             │  15:08 inbox 28 sent to coordinator                      │                             │
│                             │  15:07 land 4-agents-equal-columns · ede94f3e            │                             │
│                             │  14:58 land 3-tui-restart · 3ccd93fb                     │                             │
│                             │  14:44 land 2-stale-coordinator-startup · 0e079272       │                             │
│                             │                                                          │                             │
│                             ╰──────────────────────────────────────────────────────────╯                             │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
 [details] a attach · p peek · t task · ↑↓ scroll · esc close                                                           
```

## CompactAgents

`hive-grid 1 80 24`

```text
 hive                                     coord ● · sup ● · needs you 0 · 14:22 
  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs                        
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────
╭─ project agents · 1 project ─────────────────────────────────────────────────╮
│ ╭─ hive · g1-tui-restart ──────────────────────────────────────────────────╮ │
│ │                                                                          │ │
│ │                         ╭──────────────────────╮                         │ │
│ │                         │ ◐ manager-hive       │                         │ │
│ │                         │ claude · idle · 4m   │                         │ │
│ │                         ╰──────────────────────╯                         │ │
│ │                                     │                                    │ │
│ │             ╭───────────────────────┴───────────────────────╮            │ │
│ │ ╭──────────────────────╮                        ╭──────────────────────╮ │ │
│ │ │ ○ t4-worker-1        │                        │ ○ t4-verifier-1      │ │ │
│ │ │ pi · stopped         │                        │ pi · stopped         │ │ │
│ │ ╰──────────────────────╯                        ╰──────────────────────╯ │ │
│ │                                                                          │ │
│ │                                                                          │ │
│ │ h hides 2 stopped                                                        │ │
│ │                                                                          │ │
│ │                                                                          │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
╰──────────────────────────────────────────────────────────────────────────────╯
 [graph] ←↑↓→ move · enter details · a attach · p peek · ?                      
```

## Chat

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
────────────━━━━━━━━────────────────────────────────────────────────────────────────────────────────────────────────────
╭─ chat · noise normal ───────────────────────────────────── ↑ 42 earlier ─────────────────────────────────────────────╮
│                                                                                                                      │
│ 14:02  you           I want to give the user an ability to restart the coordinator or supervisor within the hive     │
│                      UI.                                                                                             │
│                                                                                                                      │
│ 14:19  · manager-hi… goal g1-tui-restart created · wave A started · worker t2-worker-1 live                          │
│                                                                                                                      │
│ 14:21  coordinator   Interrogate, round 2                                                                            │
│                                                                                                                      │
│                      1. What happens to the old tab on restart?                                                      │
│                      Bare hive makes the coordinator tab, but its record has spawned=false. agent stop refuses it.   │
│                                                                                                                      │
│                       Choice  │ Answer                               │ Pros                 │ Cons                   │
│                      ─────────┼──────────────────────────────────────┼──────────────────────┼──────────────────      │
│                       A (REC) │ Record that bare hive started it     │ No tab pile-up       │ Field change           │
│                       B       │ Leave the old tab, open a new one    │ Smallest change      │ Dead tabs pile up      │
│                       C       │ Reuse the same pane                  │ One tab always       │ Must exit harness      │
│                                                                                                                      ┃
│                      • Recommended: A, because hive may close only panes it started.                                 ┃
│                      • Reply like 1A, 2A, 3A, or simplify N.                                                         ┃
│                                                                                                                      ┃
│                        hive agent stop manager-hive                                                                  ┃
│                        hive agent spawn --role manager                                                               ┃
│                                                                                                                      ┃
│ 14:24  you           1A, 2A, 3B                                                                                      ┃
│                                                                                                                      ┃
│                                                                                                                      ┃
│                                                                                                                      ┃
│                                                                                                                      ┃
│                                                                                                                      ┃
│                                                                                                                      ┃
│                                                                                                                      │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ message ────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ > 1A, 2A, 3B█                                                                                                        │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
 [input] enter send · shift+enter newline · esc leave input · pgup/pgdn scroll feed · tab region                        
```

## CompactChat

`hive-grid 1 80 24`

```text
 hive                                     coord ● · sup ● · needs you 0 · 14:22 
  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs                        
───────────━━━━━━━━─────────────────────────────────────────────────────────────
╭─ chat ───────────────────────────────────────────────────────────────────────╮
│ 14:21 coordinator                                                            │
│ Interrogate, round 2                                                         │
│                                                                              │
│ Bare hive makes the coordinator tab, but its record has                      │
│ spawned=false. agent stop refuses it.                                        ┃
│                                                                              ┃
│  Choice  │ Answer                       │ Pros            │ Cons             ┃
│ ─────────┼──────────────────────────────┼─────────────────┼───────────────   ┃
│  A (REC) │ Record that bare hive start… │ No tab pile-up  │ Field change     ┃
│  B       │ Leave the old tab            │ Smallest change │ Tabs pile up     ┃
│                                                                              │
│ • Recommended: A, hive closes only panes it started.                         │
│                                                                              │
│ 14:24 you  1A, 2A, 3B                                                        │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ message ────────────────────────────────────────────────────────────────────╮
│ > █                                                                          │
╰──────────────────────────────────────────────────────────────────────────────╯
 [input] enter send · esc leave input · pgup/pgdn feed · tab region             
```

## WorkGoals

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─────────────────────━━━━━━━━───────────────────────────────────────────────────────────────────────────────────────────
 [Goals]   Queue    Tasks                                                                             [ ] switch view   
╭─ goals · 2 ──────────────────────────────────────╮╭─ g1-tui-restart ─────────────────────────────────────────────────╮
│   GOAL                    STATUS    WAVES        ││                                                                  │
│   g2-tui-v2               planned   0/5          ││ title     TUI restart and stale coordinator startup              │
│ ▶ g1-tui-restart          done      3/3          ││ status    done                                                   │
│                                                  ││ project   hive                                                   │
│ g2-tui-v2 is example data                        ││ grant     grants.md 2026-10-18 hive delivery on dev              │
│                                                  ││ manager   manager-hive                                           │
│                                                  ││                                                                  │
│                                                  ││ WAVES                                                            │
│                                                  ││ WAVE TITLE                      DONE WHEN  LANDED  STATUS        │
│                                                  ││ A    Stale coordinator startup  dev        1/1     ✓ done        │
│                                                  ││ B    TUI restart                dev        1/1     ✓ done        │
│                                                  ││ C    Agents graph columns       dev        1/1     ✓ done        │
│                                                  ││                                                                  │
│                                                  ││ TASKS                                                            │
│                                                  ││ A   2-stale-coordinator-startup   merge 0e079272                 │
│                                                  ││ B   3-tui-restart                 merge 3ccd93fb                 │
│                                                  ││ C   4-agents-equal-columns        merge ede94f3e                 │
│                                                  ││                                                                  │
│                                                  ││ NEEDS YOU                                                        │
│                                                  ││ nothing                                                          │
│                                                  ││                                                                  │
│                                                  ││ o goal design   t open task in Tasks                             │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · tab region · [ ] view · enter open · o design · t task · / filter · ? help                            
```

## WorkQueue

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─────────────────────━━━━━━━━───────────────────────────────────────────────────────────────────────────────────────────
  Goals   [Queue]   Tasks                                                                             [ ] switch view   
╭─ queue · hive · 4 tasks · 1 stop ────────────────╮╭─ 1-cp5-end-to-end ───────────────────────────────────────────────╮
│                                                  ││                                                                  │
│ GRANT  grants.md 2026-10-18 hive delivery on dev ││ title     CP5 v1 end to end                                      │
│                                                  ││ status    assigned                                               │
│   2-stale-coordinator-startup         landed     ││ goal      none                                                   │
│   └▶ 3-tui-restart                    landed     ││ grant     none                                                   │
│     └▶ 4-agents-equal-columns         landed     ││ worker    cp5-worker-01 · unresolved                             │
│                                                  ││ depends   none                                                   │
│ NO GRANT                                         ││                                                                  │
│                                                  ││ STOP                                                             │
│ ▶ 1-cp5-end-to-end                    assigned   ││ The worker record is live, but its pane is gone.                 │
│   ! stop: worker unresolved                      ││ The task cannot move until the user decides.                     │
│                                                  ││                                                                  │
│                                                  ││ NEXT STEP                                                        │
│                                                  ││ Ask the coordinator in 2 Chat.                                   │
│                                                  ││                                                                  │
│                                                  ││ t open task in Tasks  s show stops only                          │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · tab region · [ ] view · enter open · s stops only · t task · / filter · ? help                        
```

## Work

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─────────────────────━━━━━━━━───────────────────────────────────────────────────────────────────────────────────────────
  Goals    Queue   [Tasks]                                                                            [ ] switch view   
╭─ tasks · 4 ──────────────────────────────────────╮╭─ 3-tui-restart ──────────────────────────────────────────────────╮
│   ID                        STATUS    WAVE       ││                                                                  │
│   4-agents-equal-columns    landed    C          ││ title     Restart of the coordinator and the supervisor from th… │
│ ▶ 3-tui-restart             landed    B          ││ status    landed                                                 │
│   2-stale-coordinator-star… landed    A          ││ goal      g1-tui-restart · wave B                                │
│   1-cp5-end-to-end          assigned  —          ││ worker    t3-worker-1 · stopped                                  │
│                                                  ││ verifier  t3-verifier-1 · pass                                   │
│                                                  ││ branch    task/3-tui-restart                                     │
│                                                  ││ merge     3ccd93fb on dev                                        │
│                                                  ││ depends   2-stale-coordinator-startup                            │
│                                                  ││                                                                  │
│                                                  ││ ACCEPTANCE                                                       │
│                                                  ││ ✓ AC1 r restarts a coordinator or supervisor, confirms only whe… │
│                                                  ││ ✓ AC2 coordinator restart closes only a pane that hive started   │
│                                                  ││ ✓ AC3 supervisor restart runs hive supervise in a new tab        │
│                                                  ││ ✓ AC4 old records get stale or stopped, stay on disk, h hides t… │
│                                                  ││ ✓ AC5 header shows old and r to restart for an old supervisor    │
│                                                  ││ ✓ AC6 coordinator start waits for herdr readiness                │
│                                                  ││                                                                  │
│                                                  ││ EVIDENCE                                                         │
│                                                  ││ b brief.md    v VERIFIED.md   d deps graph                       │
│                                                  ││                                                                  │
│                                                  ││ V1 go test -race ./internal/tui/...      pass                    │
│                                                  ││ V3 live restart in a scratch home       pass                     │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · tab region · [ ] view · enter open · b brief · v verified · d deps · / filter · ? help                
```

## Compact

`hive-grid 1 80 24`

```text
 hive                                     coord ● · sup ● · needs you 0 · 14:22 
  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs                        
───────────────────━━━━━━━━─────────────────────────────────────────────────────
  Goals    Queue   [Tasks]                                      list › details  
╭─ 3-tui-restart · esc back ───────────────────────────────────────────────────╮
│ status    landed                                                             ┃
│ goal      g1-tui-restart · wave B                                            ┃
│ verifier  t3-verifier-1 · pass                                               ┃
│ merge     3ccd93fb on dev                                                    ┃
│                                                                              ┃
│ ACCEPTANCE                                                                   ┃
│ ✓ AC1 r restarts a coordinator or supervisor, confirms only when live        ┃
│ ✓ AC2 coordinator restart closes only a pane that hive started               ┃
│ ✓ AC3 supervisor restart runs hive supervise in a new tab                    ┃
│ ✓ AC4 old records get stale or stopped, stay on disk, h hides them           ┃
│ ✓ AC5 header shows old and r to restart for an old supervisor                ┃
│ ✓ AC6 coordinator start waits for herdr readiness                            ┃
│                                                                              │
│ EVIDENCE                                                                     │
│ b brief.md    v VERIFIED.md   d deps graph                                   │
│ ↓ 3 more lines                                                               │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 [details] ↑↓ scroll · esc back to list · b brief · v verified · d deps         
```

## Friction

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
──────────────────────────────━━━━━━━━━━━━──────────────────────────────────────────────────────────────────────────────
 [Open]   Resolved                                                                                    [ ] switch view   
╭─ open · 15 ──────────────────────────────────────╮╭─ scratch-hive-starts-real-agents ────────────────────────────────╮
│   ENTRY                             OPENED       ││                                                                  │
│   agent-send-flag-means-target      10-21        ││ title     Bare hive in a scratch home starts real tabs           ┃
│   agent-can-send-user-chat          10-20        ││ status    Open                                                   ┃
│   completed-task-cannot-reopen      10-20        ││ opened    2026-10-20                                             ┃
│   local-gate-misses-ci-tidy         10-20        ││ seen      3 times · last 2026-10-21                              ┃
│ ▶ scratch-hive-starts-real-agents   10-20        ││ file      entries/scratch-hive-starts-real-agents.md             ┃
│   ste-baseline-row-conflicts        10-20        ││                                                                  ┃
│   verify-fail-lost-on-pass          10-20        ││ SUMMARY                                                          ┃
│   worker-has-no-inbox               10-20        ││ Bare hive runs the startup of G9: in herdr, it starts a          ┃
│   bypass-still-prompts-dangerous-rm 10-19        ││ coordinator and a supervisor in their own tabs when they are     ┃
│   no-agent-dashboard                10-19        ││ missing. A worker or verifier that runs bare hive with           ┃
│   no-reviewer-role                  10-19        ││ HIVE_HOME set to a scratch home gets no live coordinator and a   ┃
│   nudge-interrupts-user-typing      10-19        ││ stale supervisor claim, so each run starts a new coordinator     ┃
│   old-features-lack-readme          10-19        ││ tab and a new hive supervise process in the caller workspace.    ┃
│   verified-md-path-unspecified      10-19        ││                                                                  ┃
│   worker-gate-adds-golangci-lint    10-19        ││ REPRODUCTION                                                     ┃
│                                                  ││ 1. In a herdr pane, make a scratch home with a marker.           ┃
│                                                  ││ 2. Run HIVE_HOME=<scratch> hive and quit the TUI.                ┃
│                                                  ││ 3. herdr tab list shows a new coordinator and supervisor tab.    ┃
│                                                  ││                                                                  │
│                                                  ││ OCCURRENCES                                                      │
│                                                  ││ 2026-10-20  worker of 101 left 4 supervisors, 8 tabs             │
│                                                  ││ 2026-10-20  t91-worker-1 closed the user TUI tab                 │
│                                                  ││ 2026-10-21  third scratch hive tab leak                          │
│                                                  ││                                                                  │
│                                                  ││ ↓ 12 more lines                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · tab region · [ ] view · enter open · / filter · g/G top/end · ? help                                  
```

## FrictionResolved

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
──────────────────────────────━━━━━━━━━━━━──────────────────────────────────────────────────────────────────────────────
  Open   [Resolved]                                                                                   [ ] switch view   
╭─ resolved · 20 ──────────────────────────────────╮╭─ supervisor-runs-stale-binary ───────────────────────────────────╮
│   ENTRY                             RESOLVED     ││                                                                  │
│   stale-coordinator-blocks-startup  10-21        ┃│ title     The supervisor keeps running an old hive binary        │
│ ▶ supervisor-runs-stale-binary      10-21        ┃│ status    Resolved 2026-10-21                                    │
│   spec-task-needs-allowlist-first   10-20        ┃│ opened    2026-10-19                                             │
│   store-list-skips-long-keys        10-20        ┃│ decision  Task 3 adds r restart and an old-binary mark           │
│   verified-task-cannot-cancel       10-20        ┃│ file      entries/resolved/supervisor-runs-stale-binary.md       │
│   agent-spawn-skips-first-prompt    10-19        ┃│                                                                  │
│   any-agent-read-another-role-inbox 10-19        ┃│ SUMMARY                                                          │
│   exited-agent-name-blocks-reuse    10-19        ┃│ On 2026-10-19, the supervisor (pid 1718014) started at 13:22     │
│   full-pi-bootstrap-flow            10-19        ┃│ from the hive binary of 72a33b5b. The managers rebuilt           │
│   idle-session-supervisor-nudges    10-19        ┃│ ~/go/bin/hive after waves 2, 3, and 4. The supervisor kept the   │
│   manager-blocked-by-harness-permi… 10-19        ┃│ old code, so the wave 2 change was not live. hive doctor         │
│   manager-spawn-shell-ready-race    10-19        ┃│ reported the supervisor as live and did not compare its binary   │
│   no-hive-path-repairs-verified-ta… 10-19        ┃│ with the installed one.                                          │
│   shell-ready-spawn-race            10-19        ┃│                                                                  │
│   task-edit-brief-command           10-19        ││ RESOLVED BY                                                      │
│   verifier-spawn-pane-requires-pro… 10-19        ││ 3-tui-restart   merge 3ccd93fb                                   │
│   verifier-verdict-without-commit   10-19        ││                                                                  │
│   verify-fail-no-reverify-workspace 10-19        ││ ↓ 9 more lines                                                   │
│   verify-fail-repair-worker-no-wor… 10-19        ││                                                                  │
│   worker-spawn-keeps-old-acting-ag… 10-19        ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · tab region · [ ] view · enter open · / filter · g/G top/end · ? help                                  
```

## CompactFriction

`hive-grid 1 80 24`

```text
 hive                                     coord ● · sup ● · needs you 0 · 14:22 
  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs                        
───────────────────────────━━━━━━━━━━━━─────────────────────────────────────────
 [Open]   Resolved                                            [ ] switch view   
╭─ open · 15 ──────────────────────────────────────────────────────────────────╮
│   ENTRY                                 OPENED      SEEN                     │
│   agent-send-flag-means-target          2026-10-21  1                        │
│   agent-can-send-user-chat              2026-10-20  1                        │
│   completed-task-cannot-reopen          2026-10-20  1                        │
│   local-gate-misses-ci-tidy             2026-10-20  1                        │
│ ▶ scratch-hive-starts-real-agents       2026-10-20  3                        │
│   ste-baseline-row-conflicts            2026-10-20  1                        │
│   verify-fail-lost-on-pass              2026-10-20  1                        │
│   worker-has-no-inbox                   2026-10-20  1                        │
│   bypass-still-prompts-dangerous-rm     2026-10-19  1                        │
│   no-agent-dashboard                    2026-10-19  1                        │
│   no-reviewer-role                      2026-10-19  2                        │
│   nudge-interrupts-user-typing          2026-10-19  1                        │
│   old-features-lack-readme              2026-10-19  1                        │
│   verified-md-path-unspecified          2026-10-19  1                        │
│   worker-gate-adds-golangci-lint        2026-10-19  1                        │
│                                                         enter opens details  │
╰──────────────────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · enter details · [ ] view · / filter · ?                       
```

## Health

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
───────────────────────────────────────────━━━━━━━━━━───────────────────────────────────────────────────────────────────
 [Home]   Agents                                                                                      [ ] switch view   
╭─ checks · 17 · 1 warn ───────────────────────────╮╭─ unresolved agents ──────────────────────────────────────────────╮
│   CHECK                   RESULT                 ││                                                                  │
│   ✓ herdr socket          protocol 22            ││ result    warn · 1 agent                                         │
│   ✓ herdr pane            w58:p4                 ││ source    hive doctor                                            │
│   ✓ home                  ~/source/hive          ││ checked   14:22:05 · R to check again                            │
│   ✓ state store           76 records             ││                                                                  │
│   ✓ starting agents       0                      ││ AGENTS                                                           │
│ ▶ ! unresolved agents     1                      ││ ID                ROLE      STATE       PANE                     │
│   ✓ git                   2.43.0                 ││ cp5-worker-01     worker    unresolved  gone                     │
│   ✓ supervisor            live · pid 1620686     ││                                                                  │
│   ✓ supervisor binary     current                ││ WHAT IT MEANS                                                    │
│   ✓ stale worktrees       none                   ││ The record says the agent is live, but hive cannot find its      │
│   ✓ unpushed branches     none                   ││ pane. Hive does not change the record by itself.                 │
│   ✓ harness claude        found                  ││                                                                  │
│   ✓ harness codex         found                  ││ NEXT STEP                                                        │
│   ✓ harness opencode      found                  ││ Ask the coordinator in 2 Chat. Health only shows state.          │
│   ✓ harness pi            found                  ││                                                                  │
│   ✓ harness cursor-agent  found                  ││                                                                  │
│   ✓ gh                    2.102.0                ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · tab region · [ ] view · enter open · R refresh · ? help                                               
```

## HealthAgents

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
───────────────────────────────────────────━━━━━━━━━━───────────────────────────────────────────────────────────────────
  Home   [Agents]                                                                                     [ ] switch view   
╭─ agents · 5 · h hides 2 stopped ─────────────────╮╭─ supervisor-g4 ──────────────────────────────────────────────────╮
│   AGENT                 ROLE        STATE        ││                                                                  │
│   coordinator-g2        coordinator ● working    ││ role      supervisor                                             │
│ ▶ supervisor-g4         supervisor  ● live       ││ state     live                                                   │
│   manager-hive          manager     ◐ idle       ││ pid       1620686                                                │
│   cp5-worker-01         worker      ● unresolved ││ pane      w58:p5 · tab supervisor                                │
│   t4-worker-1           worker      ○ stopped    ││ binary    current · ~/go/bin/hive                                │
│                                                  ││ started   by bare hive                                           │
│                                                  ││                                                                  │
│                                                  ││ ACTIONS                                                          │
│                                                  ││ r restart · runs hive supervise in a new tab                     │
│                                                  ││ a attach to the pane                                             │
│                                                  ││ p peek at the pane                                               │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││       ╭─ restart supervisor-g4? ───────────────────────────╮     │
│                                                  ││       │                                                    │     │
│                                                  ││       │ It is live. Hive stops it and starts a new one.    │     │
│                                                  ││       │ The old record stays on disk as stopped.           │     │
│                                                  ││       │                                                    │     │
│                                                  ││       │ y restart   esc cancel                             │     │
│                                                  ││       ╰────────────────────────────────────────────────────╯     │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││ r asks to confirm only when the agent is live.                   │
│                                                  ││ A stale or stopped agent restarts at once.                       │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
│                                                  ││                                                                  │
╰──────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────────────────╯
 [restart] y restart · any other key cancels                                                                            
```

## CompactHealth

`hive-grid 1 80 24`

```text
 hive                                     coord ● · sup ● · needs you 0 · 14:22 
  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs                        
───────────────────────────────────────━━━━━━━━━━───────────────────────────────
  Home   [Agents]                                             [ ] switch view   
╭─ agents · 5 ─────────────────────────────────────────────────────────────────╮
│   AGENT                 ROLE          STATE         PANE                     │
│ ▶ coordinator-g2        coordinator   ● working     w58:p4                   │
│   supervisor-g4         supervisor    ● live        w58:p5                   │
│   manager-hive          manager       ◐ idle        w58:p6                   │
│   cp5-worker-01         worker        ● unresolved  gone                     │
│   t4-worker-1           worker        ○ stopped     —                        │
│                                                                              │
│                                                                              │
│ r restarts the coordinator or supervisor. It asks first only when it is live.│
│ h hides 1 stopped agent.                                                     │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 [list] ↑↓ move · enter details · [ ] view · r restart · R refresh · ?          
```

## Logs

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
──────────────────────────────────────────────────────━━━━━━━━──────────────────────────────────────────────────────────
╭─ filter ─────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ / level:debug agent:manager-hive                                                  follow ●  < 2026-10-21 >           │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
╭─ events · 312 ───────────────────────────────────────────────╮╭─ event 14:38:02 ─────────────────────────────────────╮
│                                                              ││                                                      │
│   14:19:40 DEBUG inbox     record sent to coordinator        ││ {                                                    ┃
│   14:19:40 DEBUG supervise nudge coordinator · inbox 2       ││   "time": "2026-10-21T14:38:02.114Z",                ┃
│   14:21:02 DEBUG agent     spawn worker t2-worker-1          ││   "level": "WARN",                                   ┃
│   14:21:09 DEBUG herdr     agent.start w5T:p1 · 812ms        ││   "msg": "verify fail",                              ┃
│   14:24:51 DEBUG task      status assigned → working         ││   "component": "task",                               ┃
│   14:33:17 DEBUG task      status working → completed        ││   "command": "task verify",                          ┃
│   14:33:18 DEBUG agent     stop t2-worker-1 · pane closed    ││   "project_id": "hive",                              ┃
│   14:33:40 DEBUG agent     spawn verifier t2-verifier-1      ││   "task_id": "2-stale-coordinator-startup",          ┃
│ ▶ 14:38:02 WARN  task      verify fail · V4 diff scope       ││   "agent_id": "t2-verifier-1",                       ┃
│   14:38:30 DEBUG agent     spawn worker t2-worker-2          ││   "request_id": "r-7f31c2",                          ┃
│   14:44:12 DEBUG task      status completed → verified       ││   "duration_ms": 412,                                ┃
│   14:44:20 DEBUG land      merge task/2 into dev · 0e079272  ││   "items": {                                         ┃
│   14:44:22 DEBUG land      push dev · 1.9s                   ││     "V1": "pass",                                    ┃
│   14:44:23 DEBUG worktree  remove 2-stale-coordinator-start… ││     "V2": "pass",                                    ┃
│   14:44:25 DEBUG inbox     record sent to coordinator        ││     "V3": "pass",                                    ┃
│   14:45:01 DEBUG agent     spawn worker t3-worker-1          ││     "V4": "fail",                                    ┃
│                                                              ││     "V5": "pass"                                     ┃
│                                                              ││   },                                                 ┃
│                                                              ││   "reason": "V4 diff scope; Design section missin…   ┃
│                                                              ││   "truncated": false                                 ┃
│                                                              ││ }                                                    ┃
│                                                              ││                                                      │
│                                                              ┃│                                                      │
│                                                              ┃│                                                      │
│                                                              ┃│                                                      │
│                                                              ┃│                                                      │
│                                                              ┃│                                                      │
│                                                              ┃│                                                      │
│                                                              ┃│                                                      │
│                                                              ││                                                      │
╰──────────────────────────────────────────────────────────────╯╰──────────────────────────────────────────────────────╯
 [json] ↑↓ scroll · tab region · f follow · < > day · / filter · g/G top/end · ? help                                   
```

## CompactLogs

`hive-grid 1 80 24`

```text
 hive                                     coord ● · sup ● · needs you 0 · 14:22 
  1 Agents  2 Chat  3 Work  4 Friction  5 Health  6 Logs                        
─────────────────────────────────────────────────━━━━━━━━───────────────────────
╭─ filter ─────────────────────────────────────────────────────────────────────╮
│ / level:debug agent:manager-hive                        < 2026-10-21 >       │
╰──────────────────────────────────────────────────────────────────────────────╯
╭─ event 14:38:02 · esc back ──────────────────────────────────────────────────╮
│ {                                                                            ┃
│   "time": "2026-10-21T14:38:02.114Z",                                        ┃
│   "level": "WARN",                                                           ┃
│   "msg": "verify fail",                                                      ┃
│   "component": "task",                                                       ┃
│   "command": "task verify",                                                  ┃
│   "project_id": "hive",                                                      ┃
│   "task_id": "2-stale-coordinator-startup",                                  ┃
│   "agent_id": "t2-verifier-1",                                               ┃
│   "request_id": "r-7f31c2",                                                  ┃
│   "duration_ms": 412,                                                        │
│   "reason": "V4 diff scope; Design section missing",                         │
│   "truncated": false                                                         │
│ }                                                                            │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 [json] ↑↓ scroll · esc back to events · f follow                               
```

## Help

`hive-grid 1 120 40`

```text
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────────────────────────────────────────────
╭─ project agents ─────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                                                      │
│                           ╭─ hive ───────────────────────────────────────────────────────╮                           │
│                 ╭─ keys · ? or esc closes ─────────────────────────────────────────────────────────╮                 │
│                 │                                                                                  │                 │
│                 │  GLOBAL · same on every screen            ACTIONS · one meaning each             │                 │
│                 │                                                                                  │                 │
│                 │  1–6        switch screen                 a     attach to pane                   │                 │
│                 │  tab        next region                   p     peek pane                        │                 │
│                 │  [ ]        previous / next view          r     restart · Health › Agents        │                 │
│                 │  ↑↓ j k     move                          h     hide stopped agents              │                 │
│                 │  ←→         move on the graph             t     open task                        │                 │
│                 │  pgup pgdn  page                          b     task brief                       │                 │
│                 │  g G        top / end                     v     VERIFIED.md                      │                 │
│                 │  enter      open / details                d     dependency graph                 │                 │
│                 │  esc        back / close                  o     goal design                      │                 │
│                 │  /          filter bar                    s     stops only                       │                 │
│                 │  ?          this help                     f     follow logs                      │                 │
│                 │  R          refresh now                   < >   previous / next day              │                 │
│                 │  ctrl+c ×2  quit                                                                 │                 │
│                 │                                                                                  │                 │
│                 │                                                                                  │                 │
│                 │  FILTER FIELDS                                                                   │                 │
│                 │  level:  comp:  proj:  agent:  task:  req:  text                                 │                 │
│                 │  example  / level:warn agent:manager-hive restart                                │                 │
│                 │                                                                                  │                 │
│                 │  CHAT                                                                            │                 │
│                 │  i or enter types in chat · esc leaves the input · 1–6 work again                │                 │
│                 │                                                                                  │                 │
│                 │  r restart: coordinator or supervisor, asks first only when live                 │                 │
│                 │                                                                                  │                 │
│                 ╰──────────────────────────────────────────────────────────────────────────────────╯                 │
│                                                                                                                      │
│                                                                                                                      │
│                                                                                                                      │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
 [help] ↑↓ scroll · ? or esc close                                                                                      
```

## TooSmall

`hive-grid 1 60 16`

```text
 hive                                                       
                                                            
                                                            
                                                            
         ╭─ terminal too small ───────────────────╮         
         │                                        │         
         │           this pane is 60×16           │         
         │       hive needs at least 80×24        │         
         │                                        │         
         │     resize or zoom the herdr pane      │         
         │                                        │         
         ╰────────────────────────────────────────╯         
                                                            
                                                            
                                                            
 ctrl+c ×2 quit                                             
```

## Components

`hive-grid 1 120 67`

```text
HIVE TUI COMPONENTS  Lip Gloss + Bubbles · each block below is one reusable Go component                                
                                                                                                                        
Palette  terminal default background and foreground · lipgloss.Color                                                    
██ fg        ██ border    ██ dim       ██ accent    ██ working   ██ done      ██ info      ██ user      ██ code         
   default      #3d3a33      #a39d90      #f5b942      #6ee69a      #7ccf6a      #6fb7ff      #e9cf9b new  #9fd0ff new  
                                                                                                                        
Header  bold · full width · left: home · right: coordinator, supervisor, needs you, clock                               
 hive  ~/source/hive                                    coordinator ● working   supervisor ● live   needs you 0   14:22 
                                                                                                                        
TabBar  keys 1–6 · active: accent bold + ━ underline · others dim · rule in border color                                
  1 Agents   2 Chat   3 Work   4 Friction   5 Health   6 Logs                                                           
─━━━━━━━━━━─────────────────────────────────────────────────────────────────────────────────────────────────────────────
                                                                                                                        
SubTabs  [ ] switch · active in brackets, accent bold · hint at right                                                   
  Goals    Queue   [Tasks]                                                                            [ ] switch view   
                                                                                                                        
Panel  RoundedBorder · Padding(0,1) · title in top border · focus = accent border · tab moves focus                     
╭─ unfocused ──────────────────────────╮    ╭─ focused ────────────────────────────╮                                    
│                                      │    │                                      │                                    
│ border #3d3a33                       │    │ border #f5b942 · takes the keys      │                                    
│                                      │    │                                      │                                    
╰──────────────────────────────────────╯    ╰──────────────────────────────────────╯                                    
                                                                                                                        
Viewport + ScrollMark  only regions scroll · ┃ thumb drawn on the right border · horizontal ━ on the bottom border      
╭─ viewport ───────────────────────────────────────────────╮    ╭─ horizontal ─────────────────────────────────────────╮
│ line 41 of 120                                           │    │                                                      │
│ line 42 of 120                                           ┃    │ projects 1–4 of 5 · ←→ at the edge                   │
│ line 43 of 120                                           ┃    │                                                      │
│ line 44 of 120                                           │    │                                                      │
│ line 45 of 120                                           │    │                                                      │
╰──────────────────────────────────────────────────────────╯    ╰━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━────────────╯
                                                                                                                        
List  ▶ selected · fixed columns · truncates with …             Details  labels dim, 10 cols · sections accent          
╭─ tasks · 4 ──────────────────────────────────────────────╮    ╭─ 3-tui-restart ──────────────────────────────────────╮
│   ID                        STATUS    WAVE               │    │ status    landed                                     │
│   4-agents-equal-columns    landed    C                  │    │ goal      g1-tui-restart · wave B                    │
│ ▶ 3-tui-restart             landed    B                  │    │ ACCEPTANCE                                           │
│   2-stale-coordinator-star… landed    A                  │    │ ✓ AC1 r restarts, confirms only when live            │
╰──────────────────────────────────────────────────────────╯    ╰──────────────────────────────────────────────────────╯
                                                                                                                        
FilterBar  / opens · typed fields level: comp: proj: agent: task: req: · fixed above the list                           
╭─ filter ─────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ / level:warn agent:manager-hive █                                                                                    │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
                                                                                                                        
Input + Chat text  fixed at the bottom · user text in user color · agent replies render markdown                        
╭─ message ────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│ > restart the supervisor please█                                                                                     │
╰──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
                                                                                                                        
Card  graph node · 24×4 · dot = state · one column per project, all the same width                                      
╭──────────────────────╮  ╭──────────────────────╮  ╭──────────────────────╮                                            
│ ● manager-hive       │  │ ◐ t3-verifier-1      │  │ ○ t4-worker-1        │    selected = accent border                
│ claude · working     │  │ pi · idle · 6m       │  │ pi · stopped         │    ● working ◐ idle ○ stopped              
╰──────────────────────╯  ╰──────────────────────╯  ╰──────────────────────╯                                            
                                                                                                                        
Popup  centered over a dimmed screen · enter opens · esc closes · same Panel + Details inside                           
╭─ manager-hive · esc closes ──────────────╮                                                                            
│                                          │    below 120 columns, every details panel                                  
│ a attach · p peek · t task               │    opens as a full-width overlay instead                                   
│                                          │                                                                            
╰──────────────────────────────────────────╯                                                                            
                                                                                                                        
KeyHints  footer · [region] first in accent · hints dim · always the last row                                           
 [graph] ←↑↓→ move · enter details · a attach · p peek · t task · ? help                                                
                                                                                                                        
Notice  below 80×24 the whole screen shows only the too-small notice                                                    
```
