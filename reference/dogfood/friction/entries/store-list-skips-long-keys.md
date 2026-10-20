---
title: "store.List skips an agent record with a key longer than 18 characters"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
`store.List` in `internal/lib/store/store.go` skips each record file whose ID
fails `records.ValidID`, which allows at most 18 characters. Agent keys such as
`manager-stenographer-g7` (23 characters) are valid agent keys, so a caller
that lists agents with `store.List` loses them with no error. `agent.List`
shows them. On 2026-10-20 the verifier of task 81-agents-screen found that the
new Agents screen missed `manager-stenographer-g1` to `-g7` on the real home
while the header counted them. Repair 1 of task 81 moves the screen to
`agent.List`. The silent skip stays for each other caller of `store.List`.

**Reproduction steps:**
1. Register a manager for a project with an ID of 12 or more characters, so
   its key is longer than 18 characters.
2. List the `agents` kind with `store.List`.
3. The record is not in the result, and no error shows.
