---
title: "The local gate does not run the module tidy check of CI"
status: "Open"
opened: 2026-10-20
decision: ""
---

**Summary:**
The CI job `gate` in `.github/workflows/ci.yml` runs `go mod tidy` and
`git diff --exit-code -- go.mod go.sum`. The gate in `AGENTS.md`, the worker
skill, and the verifier skill has four commands and no tidy check. Task
79-tui-shell added a direct import of `github.com/charmbracelet/x/ansi`, but
`go.mod` kept it as indirect. CI on `dev` failed from merge 7204d699 to
7def7814 (run 39388680419). Each worker and verifier gate passed. The manager
landed waves B and C on a red CI, because no step of the flow reads the CI
result.

**Reproduction steps:**
1. In a task, import a package that `go.mod` lists as indirect.
2. Run the four gate commands of `AGENTS.md`: all pass.
3. Run `go mod tidy && git diff --exit-code -- go.mod go.sum`: it fails.
4. Land the task on `dev`: the CI job `gate` fails.
