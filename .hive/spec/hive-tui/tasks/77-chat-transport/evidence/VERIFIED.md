# 77-chat-transport verification

| Item | Value |
| --- | --- |
| Task | 77-chat-transport |
| Commit verified | 0611e508 (task branch head; dev at 22f51204) |
| Date | 2026-10-20 |
| Verifier | t77-verifier-1, Claude Code, claude-opus-5-5 |

## Results

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Red: in a scratch worktree at 0611e508 I reverted `send.go`, `nudge.go`, `cmd/root.go`, `AGENTS.md`, the four skills, and the template to `dev`. I replaced `panelock.Lock` with the old `:`→`-` name rule, the bodies of `chat.Send`/`Reply`/`Log` with a `stub` failure, and `promptNudge` with the base behavior (prompt directly, no refusal, no lock). Every new test failed on its assertion: panelock (`invalid store lock: "pane-w1-pA"`, missing `pane-77313a7031`), agent (marker text missing, `agent.prompt requests = 1 while pane lock is held`, no user-tag refusal), supervise (`want user-tag refusal`, `prompts = 1 while pane lock is held`), chat (all 6 tests), cmd (`unknown command "chat"`, unknown flags, no user-tag refusal). Green: the same five `go test` commands on the task branch all print `ok` and exit 0. The scratch worktree was removed. |
| V2 | pass | Temp `HIVE_HOME` and `XDG_CONFIG_HOME`, `HERDR_PANE_ID` empty. `chat send hello` with no coordinator: `no live coordinator`, exit 1. `chat send --agent worker-one` and `chat send --pane w2:p1`: `only the user can send chat; caller is the live worker agent`, exit 1. `chat reply --agent worker-one`: `only the coordinator can run hive chat reply`, exit 1. `agent send --agent worker-one 'user: x'`: `agent send cannot send a user-tagged message`, exit 1. Neither home had `.hive/state/chat`. `chat log --json` returned 0 entries. The temp dir was removed. |
| V3 | pass | Live in herdr 0.9.5. Temp `HIVE_HOME` and `XDG_CONFIG_HOME`. Workspace `wKC`, Claude `haiku` in `wKC:p1`, registered as coordinator in the temp home. A two-line `chat send` and an `agent send` to the coordinator ran at the same time. Both exited 0 (`chat c1 sent`, `agent coordinator idle`). `herdr pane read` shows two separate submissions: `❯ ⬢ user: VERIFY_CHAT_ONE_77 …` / `VERIFY_CHAT_TWO_77` on lines 7–8 and `❯ ⬢ VERIFY_AGENT_77 …` on line 14, with the agent output between them. `chat reply --to c1 --pane wKC:p1` printed `chat c2 reply saved`. `chat log` printed `c1 user [sent]` and `c2 coordinator -> c1`. The raw file has a `pending` line, a `sent` delivery line for c1, and a reply with `reply_to: c1`. `herdr workspace get wKC` then returned `workspace_not_found`. The temp dirs were removed. The real coordinator got no chat. |
| V4 | pass | `git diff dev --stat`: 25 files. Each one is in the source table (the skills, `AGENTS.md`, `cmd/chat/chat.go`, `cmd/chat_root_test.go`, `cmd/root.go`, agent send, template, and their tests, `internal/chat/`, `internal/lib/panelock/`, supervise nudge and its test, `testdata/ste-baseline.json`), or it is the brief copy or evidence V1–V5. The baseline change adds the brief copy with 0, which is authorized. |
| V5 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.`, exit 0. `go test -race ./...` 38 packages `ok`, exit 0. |

## Scope change 1

The lock name is `pane-` + the lowercase hex of the pane ID bytes (`internal/lib/panelock/panelock.go`). `TestLockUsesHexPaneName` asserts the literal `pane-77313a7041` for `w1:pA`. `store` is unchanged.

## Notes (not failures)

- The pane prompt is `⬢ user: TEXT` with no chat ID, as R2 says. The coordinator skill tells it to "read each `⬢ user:` message and its ID", so the coordinator must get the ID from `hive chat log`.
- `chat reply` keeps the trailing newline from standard input in `text` (`"VERIFY_REPLY_77\n"`).
