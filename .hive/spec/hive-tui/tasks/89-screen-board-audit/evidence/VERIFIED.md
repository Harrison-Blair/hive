# VERIFIED: 89-screen-board-audit

Verifier: `t89-verifier-1`. Date: 2026-10-20. Branch head before this file: `3fbce3ba`.

I did not start the TUI. No check opened the Chat input.

| ID | Check | Observation | Result |
| --- | --- | --- | --- |
| V1 | 7 captures, 7 sections | The evidence has 7 `*.txt` captures. Each capture selects its screen in the sidebar: `▶ 1 Chat`, `▶ 3 Goals`, `▶ 4 Queue`, `▶ 5 Tasks`, `▶ 6 Friction`, `▶ 7 Health`, `▶ 8 Logs`. `GAPS.md` has 7 screen headings (`grep -c` gives 7). | pass |
| V2 | 3 random gaps against board and capture | `shuf -n 3` selected Q1, F1, and T2. Q1: the `Queue.dc.html` header has `runs without you 12 · stops 2 · waits behind a stop 3`. The `Queue.txt` header has only project, agent, supervisor, and needs_user counts. Decision 2 and G8 apply. F1: the `Friction.dc.html` columns are `opened issue goal`. The `Friction.txt` columns are `opened title decision`. Decision 3 applies. T2: the `Tasks.dc.html` controls are `v open VERIFIED.md` and `d deps graph`. The `Tasks.txt` footer has `tab filter · enter evidence · b brief · p project · ask in Chat · ↑/↓ move`, with no `v` or `d`. Decision 14 allows read-only views. G12 does not remove these controls, because it removes only held nudges, question cards, interruption, edits, restart, and cleanup. Each gap is real. Each decision column is correct. | pass |
| V3 | `git diff origin/dev --stat` | 15 files changed. They are the brief copy, `testdata/ste-baseline.json` (one added line: the brief with 0), and 13 files in this evidence folder. All of these paths are in the source table. | pass |
| V4 | `go test ./...` | Exit 0. All packages with tests show `ok`. | pass |

The manager decided to skip F1 and to keep the capacity panel out (decision 6). This verdict judges the report only.
