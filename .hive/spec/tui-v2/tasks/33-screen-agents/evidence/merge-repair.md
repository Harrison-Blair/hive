# Merge repair (2026-10-21)

The verifier passed V1, V2, and V-gate. The merge of `dev` failed. This file records the repair. The verifier evidence (`VERIFIED.md`, `verify3/`) does not change.

## Merges

| Commit | dev | Conflicts | Resolution |
| --- | --- | --- | --- |
| `e578f8b1` | `d3e1aa2c` (Chat, Health, Goals) | `internal/tui/shell_test.go`, `testdata/ste-baseline.json` | Both sides. The baseline keeps the entries of tasks 33 and 34. The screen-key test keeps `no project agents` for Agents (this task) and `no chat yet` for Chat (dev). |
| `656bac63` | `361b8848` (Work Tasks, task 37) | `internal/tui/agents.go`, `internal/tui/frame.go` | Both sides. Task 37 removed the Tasks case of `legacyDetails`. This task removed the Agents case. Thus `legacyDetails` has no case and goes. `t` on the graph opens the task as the Queue screen does, because task 37 removed `tasks.tab`. |

Duplicate helper: `TestAgentsBoardsMatchTheirGrids` had its own copy of the board compare. It now calls `checkBoardGolden` of dev.

## Removal of the old Agents screen (manager decision: option A)

After the first merge, Health › Agents has its own panel (task 39). No code drew the TREE, GRAPH, and TABLE views of `agents.go`. 16 tests that reached them through Health › Agents failed. The manager chose option A: remove the old views and their tests, and keep the loader, the peek, and the restart with their tests.

Removed code: `agentsScreen.View` now returns "" (the frame draws `agentsGraphView`). `agentsScreen.Update` keeps only the restart prompt, the running restart, and `esc` of the peek. The frame does not route each other key to it. The old views, their selection state (`selectedKey`, `showHistory`, `tab`, scroll offsets), `agentDetails`, `legacyDetails`, `selectedStyle`, and the node fields that only the old views read (`parentKey`, `inferred`, `ghost`, `landedGhost`, `landedLabel`, `taskStatus`, `nextTask`, `nudged`) go. The loader no longer reads the coordinator claim, because only the inferred parent used it. `golangci-lint` reports 0 issues.

### Removed tests

Each line gives the behavior that the test checked, then the test that covers it on the new screen, or the decision that removes it. "New" marks a test that this repair adds. The first 16 lines are the tests that failed after the merge.

| Test (removed) | Behavior | Covered by, or removed by |
| --- | --- | --- |
| `TestAgentsDefaultShowsCurrentGenerationsAndActiveTaskGhosts` | only the current generation shows; a stopped worker shows only with an open task, dim | Current generation: `TestAgentsGraphShowsTheCurrentGenerationOnly` (new). Ghost rule: removed by decision 14 and the task 33 decision "Stopped text": the graph shows each stopped current agent, and `h` hides them. |
| `TestAgentsHistoryToggleShowsAllGenerations` | `h` shows old generations and the landed label | Removed by decision 14: `h` hides stopped agents, and no view shows old generations. `h`: `TestAgentsHHidesStoppedAgents`. Landed label: removed by the task 33 decision "Card sub line". |
| `TestAgentsScreenRendersTopologyTabsAndDetail` | TREE, GRAPH, and TABLE tabs, the edge legend, the detail strip | Removed by decision 14 and the task 29 registry (one graph region, no views). Popup: `TestAgentsDetailsPopupKeys`, `TestAgentsBoardsMatchTheirGrids/AgentDetails`. Legend: `TestAgentsBoardsMatchTheirGrids/Main`. |
| `TestAgentsLoadsLongAgentKeysAsNodesAndParents` | a long agent key loads and shows under its project | `TestAgentsPodHoldsAgentsWithoutAParentRecord` (new). The TABLE row: removed by decision 14. |
| `TestAgentsScreenLogsTabAndSelectionActions` | the selection, `h`, and `a` log their action | `TestAgentsKeysLogTheirActions` (new). The tab action: removed by decisions 6 and 14 (`tab` moves the region, and Agents has no views). |
| `TestAgentsScreenStatusHidesStoppedAndUsesUnknownWhenSupervisorStale` | a stale agent and each agent with a stale supervisor is unknown; stopped agents hide and are dim | `TestAgentsCardsShowUnknownWhenTheSupervisorIsStale` (new). Hide: `TestAgentsHHidesStoppedAgents`. Dim: `TestAgentsBoardsMatchTheirGrids/Main` (tokens). |
| `TestAgentsGraphKeepsSelectedAgentInViewWithFortyAgents` | a tall pod keeps the selected card visible | `TestAgentsFitEachSize/forty_workers` (new data set), `TestAgentsArrowsScrollThePods`. |
| `TestAgentsSelectionMovesWithViKeysAndDetailShowsInferredParent` | `j` `k` move; the detail shows an inferred parent | Keys: `TestAgentsUpDownMoveInThePod`. Inferred parent: removed by the task 33 decision "Pod" (a pod holds the agents of its project) and the AgentDetails board (the popup shows the record parent). Placement without a parent: `TestAgentsPodHoldsAgentsWithoutAParentRecord` (new). |
| `TestAgentsRefreshKeepsSelectionByAgentKey` | refresh keeps the selection | `TestAgentsNormalModeLoadsAndRefreshes`. |
| `TestAgentsProjectsFollowProjectRecordOrder` | projects follow the record order | `TestAgentsPodsFollowProjectRecordOrder` (new). |
| `TestAgentsBoardDetailShowsTaskStatusAndNextTask` | the detail shows the task status and the next task | Removed by decision 42: the AgentDetails board gives the popup fields, and it has no task status or next task. `t`: `TestAgentsTOpensTheTask`. |
| `TestAgentsGhostsRequireAnOpenTask` | a stopped worker shows only when its task is open | Removed by decision 14 and the task 33 decision "Stopped text". |
| `TestAgentsSupervisorRowUsesLifecycleClaimPIDAndStartAge` | the supervisor pid and start age | Kept and changed, not removed: it now checks the loader node (`pid`, `age`), which Health › Agents shows. The graph row: removed by decision 14 (the supervisor is not on the graph). |
| `TestAgentsNudgeEdgesRequireUnreadNudgedRoleInbox` | a nudge shows only for an unread nudged coordinator or manager inbox | Nudge edges: removed by the task 33 decision "Legend" (`── spawn` only). The rule: `TestAgentsDetailsShowAnUnreadNudge` (new), which reads the `nudged` field of the popup. |
| `TestAgentsGraphConnectsSiblingWorkersToManagerStem` | sibling workers join the manager stem | Task 33 decision "Tree": `TestAgentsBoardsMatchTheirGrids/Main` (`┬`, `╭─┴─╮`, `┴`), `TestAgentsSelectedCardJoints`. |
| `TestAgentsGraphColumnsHaveEqualWidth` | equal project columns | Decision 10: `TestAgentsBoardsMatchTheirGrids/AgentsMany`, `TestAgentsFitEachSize`. |
| `TestAgentsNudgeEdgesDoNotTargetWorkersOrVerifiers` | no nudge for a worker or verifier | `TestAgentsDetailsShowAnUnreadNudge/nudged_worker` (new). |
| `TestAgentsGraphCardsKeepTheirWidths` | cards keep their width | Decision 10 (24×4): `component/card_test.go`, `TestAgentsBoardsMatchTheirGrids`. |
| `TestAgentsGraphSingleColumnKeepsItsPod` | one project keeps its pod | `TestAgentsBoardsMatchTheirGrids/CompactAgents`, `TestAgentsFitEachSize/forty_workers`. |
| `TestAgentsGraphLongCardFitsInEqualColumns` | long text fits the column | `TestAgentsFitEachSize/long_text`. |
| `TestAgentsGraphArrowSelectionMovesBetweenPodsAndScrolls` | `←` `→` move between pods and scroll | `TestAgentsArrowsScrollThePods`. |
| `TestAgentsGraphPodTitleSharesTopBorder` | the pod title is in the top border | Task 33 decision "Pod": `TestAgentsBoardsMatchTheirGrids`. |
| `TestAgentsGraphLongPodTitleKeepsNameTogetherBeforeStem` | a long pod title and the old stem | `TestAgentsFitEachSize/long_text`. The stem to the title: removed by the task 33 decision "Pod". |
| `TestAgentsGraphColorsMatchBoard` | graph colors | Decision 42: `TestAgentsBoardsMatchTheirGrids` compares each token. |
| `TestAgentsExplorerUsesIndentationWithoutConnectors` | TREE explorer indent | Removed by decision 14 (no TREE view). |
| `TestAgentsGraphHasNoStemForManagerWithOnlyHiddenChildren` | no stem when `h` hides each child | `TestAgentsHHidesStoppedAgents`; the task 33 decision "A pod with no child shows `no workers`". |
| `TestAgentsGraphDelegationBusUsesOpenEndGlyphs` | bus glyphs | Task 33 decision "Tree": `TestAgentsBoardsMatchTheirGrids/Main`. |
| `TestAgentsTreeLastVisibleChildUsesBottomConnector` | the last child uses the bottom connector | Task 33 decision "Rail" (`╰┤`): `TestAgentsBoardsMatchTheirGrids/AgentsMany`, `TestAgentsSelectedCardJoints`. |
| `TestAgentsGraphNodeSubtitlesHaveNoExtraSeparator` | the card sub line | Task 33 decision "Card sub line": `TestAgentsNormalModeLoadsAndRefreshes` (`pi · working · 3m`), `TestAgentsBoardsMatchTheirGrids`. |
| `TestAgentsGraphDelegationStemJoinsPodBorder` | the stem joins the pod border | `TestAgentsBoardsMatchTheirGrids`, `TestAgentsSelectedCardJoints`. |
| `TestAgentsGraphNodeCardsStayInsetInLongTitlePods` | cards stay in a long-title pod | `TestAgentsFitEachSize/long_text`. |
| `TestAgentsGraphPodsHaveRoundedCorners` | rounded pod corners | Decision 27: `TestAgentsBoardsMatchTheirGrids`. |
| `TestAgentsGraphNodeHasOneCellOfPadding` | card padding | `component/card_test.go`, `TestAgentsBoardsMatchTheirGrids`. |
| `TestAgentsDetailFieldsUseSpaces` | detail fields use spaces | `TestAgentsBoardsMatchTheirGrids/AgentDetails` (`component.DetailField`). |
| `TestAgentBoardShowsDistinctStatusColorAndDimWorkingCount` | explorer dot colors and the dim working count | The explorer: removed by decision 14. Dot colors: `TestAgentsBoardsMatchTheirGrids` (`working` tokens). |
| `TestAgentGraphDetailStripNameIsHoneyBold` | the selected name is accent and bold | `TestScreenSelectionsUseHoneyMarkerAndDefaultForeground/Agents` (changed to the new graph). |
| `TestAgentGraphBoardRendersTabsLegendHeaderAndDetails` | tab line, legend, header colors, detail strip | Tabs and the old legend: removed by decision 14 and the task 33 decision "Legend". Header tokens: `TestAgentsBoardsMatchTheirGrids/AgentsMany` (`needs you 1` in accent, `●` in working). |

The golden files `internal/tui/testdata/agents-board/*.golden` went in the first pass of this task (`V1.md`).

### Changed tests (kept)

- `TestAgentsLegacyGenerationFallbackUsesStatusChangeThenRecordTime`: reads `current` of the loader nodes.
- `TestAgentsSupervisorRowUsesLifecycleClaimPIDAndStartAge`: reads `pid` and `age` of the loader node.
- `TestAgentsPeekReadsInCommandAndEscClosesIt`: reads the frame at 120×40, where the peek panel shows each of the 20 lines.
- `TestScreenSelectionsUseHoneyMarkerAndDefaultForeground/Agents`: the selected card name is accent and bold, with the heavy border `┏━`.
- `TestDemoMainShowsFixtureDataAtTheFixtureClock`: reads `graph.selected`.
- `agentsBoardModel`: opens the Agents tab.

### New tests: deliberate breaks

Each new test passes on `656bac63`. Each break below changed one line of the code, ran the test, and restored the file (`cmp` showed no difference). Full output: `merge-repair/breaks.txt`.

| Break | Test | Result |
| --- | --- | --- |
| pods keep non-current children | `TestAgentsGraphShowsTheCurrentGenerationOnly` | FAIL |
| the loader drops `nudgedAt` | `TestAgentsDetailsShowAnUnreadNudge/unread_nudged_manager` | FAIL |
| the inbox check takes each role | `TestAgentsDetailsShowAnUnreadNudge/nudged_worker` | FAIL |
| pods skip an agent with no parent | `TestAgentsPodHoldsAgentsWithoutAParentRecord` | FAIL |
| pods in reverse order | `TestAgentsPodsFollowProjectRecordOrder` | FAIL |
| a stale supervisor does not give `unknown` | `TestAgentsCardsShowUnknownWhenTheSupervisorIsStale` | FAIL |
| the selection does not log | `TestAgentsKeysLogTheirActions` | FAIL |
| the graph does not scroll down | `TestAgentsFitEachSize/forty_workers/80x24` | FAIL |

## Live check on `656bac63` (decision 43)

Setup as `V2.md`: binary from `656bac63`, scratch tab `t33-merge2` (pane `wAN:p3`) with a scratch `HIVE_HOME`, `COLORTERM=truecolor`, and `HIVE_NO_STARTUP=1`. The demo renders only while its tab has the focus, so the check gave the focus to the scratch tab, then gave it back to `t33-worker-3`. The scratch tab is closed. `herdr tab list --workspace wAN` shows only `t33-worker-3`.

```text
Main.grid: strict differences 3438, space-cell token or bold only 3438, differences under the live rule 0
AgentsMany.grid: strict differences 2193, space-cell token or bold only 2193, differences under the live rule 0
AgentDetails.grid: strict differences 942, space-cell token or bold only 942, differences under the live rule 0
CompactAgents.grid: strict differences 779, space-cell token or bold only 779, differences under the live rule 0
Help.grid: strict differences 1420, space-cell token or bold only 1420, differences under the live rule 0
```

The counts are the same as in `V2.md`. Captures: `merge-repair/*.ansi`.

Keys on Main 120×40 (`merge-repair/keys-*.txt`):

| Keys | Result |
| --- | --- |
| (start) | `h hides 2 stopped`, `[graph] ←↑↓→ move · enter details · a attach · p peek · t task · h hide stopped · ? help` |
| `h` | `h shows 2 stopped` |
| `h`, `r` | the screen is the same as at the start (clock column ignored) |
| `down`, `enter` | popup `t4-worker-1 · esc closes`, `[details] a attach · p peek · t task · ↑↓ scroll · esc close` |
| `esc`, `a` | `[graph] demo mode: no action` |
| `t` | Work › `[Tasks]` with `4-agents-equal-columns` |

## V-gate on `656bac63`

`git diff origin/dev...HEAD --stat` (`merge-repair/diff-stat.txt`): only `internal/tui/` and its tests, `internal/tui/component/`, `.hive/spec/tui-v2/README.md` (section `Manager decisions`), `testdata/ste-baseline.json`, and the brief and evidence of this task. The README decision about `agents.go` now says that task 33 removed the old views.

Full gate output: `merge-repair/gate.txt`. Each command exits 0. `gofmt -l` and `go mod tidy -diff` print nothing. `golangci-lint` gives `0 issues.`
