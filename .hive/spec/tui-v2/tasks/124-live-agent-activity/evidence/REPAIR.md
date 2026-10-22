# Repair after the verification fail

Date: 2026-10-22. Agent: `t124-worker-2`. Base of the repair: `ab1b7fab`.
Finding: `VERIFIED.md`, "an older message can reverse the newest message
arrow".

## Cause and change

`litMessages` grouped messages by their raw sender and receiver names.
`manager-hive` (agent name) and `manager-hive-g2` (record key) name one card,
so two messages on one line stayed in two map entries. `lightLines` drew
both in map order, and the older path could be drawn last.

Change in `internal/tui/agents_screen.go`:

- `litMessages` resolves the sender and the receiver to card keys with
  `graphKey` first. Then it groups by the line (the unordered key pair) and
  keeps the newest message of each line. The result is sorted by time, then
  by sender and receiver key. No map order reaches the result.
- `lightLines` uses the resolved keys of `litMessages`.
- New `recentMessages` gives the messages of the last 5 seconds.
  `nextLitExpiry` uses it, as it used the raw time filter before. The expiry
  timer does not depend on the graph of the last view, because `litTick`
  runs after a data load and before the next render.

## V3 red and green (repair)

New test `TestLiveMessagesGroupByCard` in `internal/tui/agents_live_test.go`.
Two messages on the LivePod line from `manager-hive` to `t6-worker-1`, one by
name and one by record key, in both orders. Expected: one lit message, with
the literal keys of the newer message.

| Run | Command | Result | File |
| --- | --- | --- | --- |
| red, code at `ab1b7fab` | `go test ./internal/tui -run '^TestLiveMessagesGroupByCard$' -count=20 -v` | 20 of 20 runs fail; both subtests give 2 messages | `repair-group-red.txt` |
| red, code at `ab1b7fab` | `go test ./internal/tui -run '^TestLiveSendEventGivesTheReceiverByKey$' -count=200` | 30 of 200 runs fail (by chance) | `repair-receiver-key-red.txt` |
| green | `go test ./internal/tui -run '^TestLiveMessagesGroupByCard$' -count=20 -v` | pass | `repair-group-green.txt` |
| green | `go test ./internal/tui -run '^TestLiveSendEventGivesTheReceiverByKey$' -count=200 -v` | 200 of 200 pass | `repair-receiver-key-green.txt` |
| green | same with `-race` | 200 of 200 pass | `repair-receiver-key-race-green.txt` |
| green | `go test -race ./internal/tui -run '^TestLive' -count=50` | pass | `repair-live-repeat.txt` |

The new test fails in each run on the old code, because the old code always
returns two messages. It does not depend on map order.

## V4 live check (repair)

Build: `go build -o .hive/tmp/r124/hive .` of the repaired tree. Pane:
`wVR:p2`, a tab that this worker made in its own workspace, 206×57, herdr
0.9.5. Each run: `COLORTERM=truecolor hive --demo FIXTURE --size WxH`, read
`herdr pane read --source visible --format ansi` every 500 ms until two reads
are equal, crop to the board size, compare with `cellgrid.Compare` (space
cells compare by rune only), quit with two `ctrl+c` presses. A temporary
comparator in `.hive/tmp/r124/` did the crop. Before use, it gave the 2 arrow
diffs of the verifier on `verifier/reply-key-live.ansi`, and 0 diffs on
`verifier/LivePod.ansi`.

| Case | Result |
| --- | --- |
| `verifier/reply-key.fixture.json` at 120×40 against `verifier/reply-key.expected.grid`, 20 fresh runs | 20 of 20 have 0 diffs |
| LiveOverview, LivePod (120×40), CompactLiveOverview, CompactLivePod (80×24) | 0 diffs each |
| Main, AgentsMany, AgentDetails, AgentsHome, AgentsNested, AgentsHiddenParent, AgentsCoordinatorChild, AgentDetailsGeneral, Help (120×40); CompactAgents, CompactAgentsHome (80×24) | 0 diffs each |

Log: `repair-live.log`. Capture of run 20: `repair-reply-key-live.ansi` and
the plain text `repair-reply-key-live.ansi.txt`.

## V5 arrow points (repair)

In `repair-reply-key-live.ansi.txt` (colour removed), rows 15 to 19:

| Point | Result |
| --- | --- |
| Lit path is heavy from `t6-worker-1` (`┸` row 19) up to `manager-hive` (`┰` row 15), with no gap | pass |
| One arrow, `▲` at column 60 row 16, next to the receiver joint | pass |
| No arrow on the vertical run above `t6-worker-1` (`┃` at column 75 row 18) | pass |
| Each approved grid of the 15 Agents and Help boards still matches | pass |

No grid, fixture, or board changed in the repair.

## V-gate (repair)

`git merge origin/dev`: already up to date. Full output: `repair-gate.log`.
`gofmt -l` and `go mod tidy -diff` print nothing. `go vet`, `golangci-lint`
(`0 issues.`), and `go test -race ./...` exit 0. The repair changes only
`internal/tui/agents_screen.go`, `internal/tui/agents_live_test.go`, and this
evidence folder.

## Cleanup

The worker closed its tab `wVR:t2` and removed `.hive/tmp/r124/` (build,
comparator, captures).
