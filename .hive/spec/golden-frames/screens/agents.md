# Agents

Agents keeps the rules of `.hive/spec/tui-v2/README.md` decisions 10, 58 to 78, and 105 to 112, and `general-agent` decision 39. This file gives only the changes.

## Changes

1. The screen has no header row and no tab bar. Row 0 is the title row: `agents` in `accent` bold, then `· N projects · H home · h shows S stopped` in `dim`. Row 1 is the rule. The body starts on row 2. The outer box `project agents` of `tui-v2` decision 65 goes away. The body uses the full width with one cell of margin.
2. The title of a project pod is `PROJECT · wave W L/T`. W is the current wave of the goal of the pod. L is the count of landed tasks of that wave. T is the count of tasks of that wave. A pod with no goal shows `PROJECT` only. The goal line inside the pod stays.
3. The legend row of `tui-v2` decision 69 stays, on the row above the status rows. It ends with `wave W L/T = landed tasks`.
4. The status rows replace the footer. The `screen` segment shows `agents` and the keys of the view: `←→ pod · ↑↓ card · enter details · a attach · p peek · t task · h stopped` in the overview, and `←↑↓→ move · enter details · a attach · p peek · t task · esc pods · h stopped` in the one-pod view.
5. `esc` in the one-pod view returns to the overview. `esc` in the overview returns to Chat. With one project and no home pod, `esc` returns to Chat at once.

## States

| State | Trigger | What shows | Board |
| --- | --- | --- | --- |
| overview | two projects and a home pod | rules 1 to 3 | `GfAgents` |
| one-pod view | `enter` on a pod | the tree of `tui-v2` decision 74 with the title of rule 2 | `GfAgentsPod` |
| details popup | `enter` on a card | the popup of `tui-v2` decision 65 | `GfAgentsDetails` |
| no project agents | each pod is hidden or no project exists | the notice of `tui-v2` decision 110 | `GfAgentsNone` |
| many projects | more pods than fit | the pages of `tui-v2` decision 62 | no board |
| stopped shown | `h` | `tui-v2` decision 58 | no board |
| nested child | a child of a worker | `tui-v2` decision 74 | no board |
| wide tree | a tree wider than the region | `tui-v2` decision 75 | no board |
| lit line | a message in the last 5 seconds | `tui-v2` decisions 66 and 67 | `GfAgents` |
| starting, blocked, unresolved, stale | an agent in that state | the dot and the word of `tui-v2` decision 55 on the card | no board |
