---
title: "The worker skill gate adds golangci-lint that the project gate does not have"
status: "Open"
opened: 2026-10-19
decision: ""
---

**Summary:**
On 2026-10-19, in project qmeter, workers t50-worker-1, t51-worker-2 and
t53-worker-1 each stopped before `hive task complete` and asked the manager
for a waiver. The hive-worker gate runs `golangci-lint run ./...`. qmeter's
`AGENTS.md` and the briefs name only gofmt, vet, race tests and the
next-version test. `dev` already has 27 to 33 golangci-lint findings, and
`--new-from-rev=dev` reports 0 for each task. Each stop cost one manager
round trip. Also, a `blocked` inbox record carries only the text `blocked`,
not the reason.

**Reproduction steps:**
1. Use a project whose gate in `AGENTS.md` has no golangci-lint, and whose
   base branch has lint findings.
2. Spawn a pi worker for a task of that project.
3. The worker runs the hive-worker gate, sees golangci-lint fail on
   unchanged files, and asks the manager instead of completing.
