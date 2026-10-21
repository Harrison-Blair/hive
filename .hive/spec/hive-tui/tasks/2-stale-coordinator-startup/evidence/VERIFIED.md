# Verification

| Item | Result | Evidence |
| --- | --- | --- |
| V1 Tests | pass | I ran the three new test commands on `dev` at `12a56f8a`. Each failed for the expected reason. The same commands passed on this branch. The changed-package suite passed. The gone-pane test used a coordinator record with no claim. Startup marked it `stale` and kept its file. |
| V2 Brief copy and ratchet | pass | `go test ./ -run STE` passed. |
| V3 Live check | pass | I used Herdr 0.9.5, protocol 22, a named session, and a temporary home and config. The first run made coordinator and supervisor tabs. Pi did not accept the startup prompt, so I registered the coordinator from its Hive-created pane. Both records had `started_by_hive: true`. I closed coordinator pane `w1:p2` and ran the task binary again. It made coordinator tab `w1:t4`. Record `coordinator-g1` became `stale`, and its file stayed on disk. The log showed `outcome=gone` and `pane_id=w1:p2`. Pi did not start in the new pane, so I registered the coordinator from pane `w1:p4`. Records `coordinator-g2` and `supervisor-g1` had `started_by_hive: true`. I stopped the TUI and supervisor, closed the workspace, and removed the named session. |
| V4 Diff scope | pass | After the `dev` merge, `git diff dev...HEAD --stat` listed only files from the brief source table and the task evidence folder. |
| V5 Gate | pass | All five gate commands passed. `gofmt -l` and `go mod tidy -diff` printed nothing. `go vet ./...` printed no diagnostics. `golangci-lint run ./...` reported `0 issues`. `go test -race ./...` passed. |
| Visual check | pass | The brief has a Design section and says that no new board exists. I captured the failed-check screen at 120x40 and 200x57. Plain text shows `error` and pane ID `w1:p9` in both images. |