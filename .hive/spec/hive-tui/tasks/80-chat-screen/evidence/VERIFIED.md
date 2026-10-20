# Verified: 80-chat-screen

Verifier: independent Claude Code session (Opus 5.5), 2026-10-20. Task commit `ae93950a`. Base `dev` at `7204d699`.

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Red: I copied `internal/tui/chat_screen_test.go` into a detached worktree at `dev` (`7204d699`). `go test ./internal/tui -run '^TestChat' -count=1 -v` failed all seven tests for the expected reasons: no `coordinator` heading, Chat data request did not start (×3), footer has no `i write`, enter did not start a send command, program output omitted Chat. Green: the same command on the task branch passed all seven. I removed the base worktree. |
| V2 | pass | See the live log below. |
| V3 | pass | `git diff dev --stat` lists 13 files: `internal/tui/{chat_screen.go,chat_screen_test.go,doc.go,shell.go,shell_test.go,tick.go,tui.go}`, `testdata/ste-baseline.json` (one line, the brief copy), the brief copy, and `evidence/V1.md`–`V4.md`. All are in the source table or are evidence. |
| V4 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` printed `0 issues.`, exit 0. `go test -race ./...` exit 0 with all packages `ok`. |

## V2 live log

- herdr 0.9.5, `HERDR_ENV=1`. I built the task binary into the session scratchpad. I used a temp `HIVE_HOME` (with a copy of `.hive/README.md`) and a temp `XDG_CONFIG_HOME` in the same scratch folder. I did not touch the real coordinator.
- I created workspace `wM1` (label `80-chat-verify`). In `wM1:p1` I started `pi --model opencode-go/minimax-m3.1-flash-free`. herdr reported `agent_status: idle`.
- I registered `coordinator` (role `coordinator`, harness `pi`, pane `wM1:p1`, pid of that pi process) with `hive agent register`. `hive agent list` showed it `live` in `wM1:p1`. (My first register used a wrong PID and was marked stale; generation 2 is the valid one.)
- In `wM1:p2` I ran the task binary. Chat showed `coordinator`, `live · pi · unknown`, the model, and `wM1:p1`; the events panel showed `no events` and `noise: messages`.
- I pressed `i`, typed `Line one`, pressed `shift+enter`, typed `Line two`. The input showed both lines and the footer `enter send · shift+enter newline · esc keep draft · ctrl+c×2 quit`.
- I started `hive agent send --agent coordinator --wait "Reply with the single word NUDGE-OK ..."` in the background. When `wM1:p1` reported `working`, I pressed `enter` in Chat, then `enter` again 0.5 s later, then `n`. The agent send returned `agent coordinator done`, exit 0.
- `hive chat log --json`: exactly one entry `c1`, `"Line one\nLine two"`, delivery `sent`. The second `enter` did not send again. The `n` went into the draft as text (input stays active), so the draft was kept as `Line one / Line twon`, since it changed during the send.
- `herdr pane read wM1:p1` showed, in order: `⬢ Reply with the single word NUDGE-OK ...`, `NUDGE-OK`, then a separate `⬢ user: Line one` / `Line two` prompt, and the agent's answer that quoted both lines.
- `echo "Got both lines." | hive chat reply --to c1 --agent coordinator --pane wM1:p1` wrote `c2` with `reply_to: c1`. After the next tick, `herdr pane read wM1:p2` showed `you · Oct 20 01:53` / `Line one` / `Line two`, then `coordinator · Oct 20 01:53 ↳ c1` / `Got both lines.` under it.
- Cleanup: I quit the TUI with `ctrl+c` twice and ran `herdr workspace close wM1`. `herdr workspace list` no longer showed `80-chat-verify`, the pi process had exited, and I removed the scratch home, config, and binary.

## Observations (not failures)

- While the input is open, the Chat panel renders 3 rows taller than the sidebar and event panels, so its bottom border falls below theirs. The brief has no layout criterion. A later task can fix it.
- The coordinator harness status shows `unknown` because the registered record has no `agent_status`. That is the record data, not a screen defect.
- R4: the load also reads the coordinator agent records (for the R1 heading), beyond the chat files and the inbox. R1 needs this data.
