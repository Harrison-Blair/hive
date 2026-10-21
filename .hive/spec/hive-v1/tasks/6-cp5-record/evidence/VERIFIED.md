# CP5 record verification

Date: 2026-10-21
Verifier: `t6-verifier-1`
Model: `openai-codex/gpt-6-luna:xhigh`

| Item | Result | Re-run and findings |
| --- | --- | --- |
| V1 Citation check | fail | I read the cited goal, task, agent, and inbox records and the cited log events. Their stated values match the artifact. Citation C15 does not meet the brief's citation rule. It cites the instruction not to run and an unspecified worker statement. No record field, log event, or commit shows that no new run occurred. The ASD-STE100 linter also flags `repair` at `reference/dogfood/cp5.md:31`; `STE-GLOSSARY.md` requires `fix`. |
| V2 Landings | pass | `git log --merges --oneline origin/dev` and full-SHA checks show each cited SHA is the merge of its named task branch, and `origin/dev` contains each SHA: `0e079272924127b69f78b1142929eaa02e7229cf`, `3ccd93fbf0a4a213b9ff790321ea72ca65d6e04e`, and `ede94f3ea4b8a7fd63d99d23eae5ea74613e5253`. |
| V3 Tracker | fail | `git diff dev...HEAD -- .hive/spec/hive-v1/05-tracker.md` changes only the CP5 row and adds the dated receipt. `go test ./ -run STE` passes. AC3 fails because the CP5 verifier-model cell is empty. The model supplied before verification is `openai-codex/gpt-6-luna:xhigh`. |
| V4 main | pass | Before and after, local `main` is `3692162633f896b8fee6197469adba91f666a00c`; `origin/main` is `928cfa95e0906315fcb4a65a5cdbdef1470a0503`; `git ls-remote origin refs/heads/main` returns `928cfa95e0906315fcb4a65a5cdbdef1470a0503`. Each value stayed the same. |
| V5 Diff scope and gate | fail | The task diff lists only paths named by the brief or its evidence. The required `git merge dev` failed with a content conflict in `testdata/ste-baseline.json`; I aborted the merge and restored the clean task branch. The five gate commands pass on the unmerged task tip: `gofmt -l` and `go mod tidy -diff` print nothing, `go vet ./...` exits 0, `golangci-lint run ./...` reports `0 issues`, and `go test -race ./...` passes. The ASD-STE100 linter flags `repair` at lines 35 and 63 of the tracked brief copy; `STE-GLOSSARY.md` requires `fix`. |
