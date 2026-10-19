# 24-19-evidence-fix VERIFIED

| Item | Value |
| --- | --- |
| Task | 24-19-evidence-fix |
| Commit | 43f7d2f8 |
| Date | 2026-10-19 |
| Agent | verifier t24-verifier-1, Claude Code, claude-opus-5-5 |
| Verdict | pass |

| Item | Proves | Result | Remark |
| --- | --- | --- | --- |
| V1 | R1 | pass | `diff` of each committed V1-V3 file with its copy in the old 19-worker-respawn worktree printed nothing. |
| V2 | R1 | pass | In V1-V4, `39a35e6c` occurs only in V1.md lines 29, 62, 91. Each is the `commit` line of a red run that restores the base files with `git show HEAD:`. Each green run names `2c69b4ad`. Other matches are in VERIFIED.md, which the task does not change. |
| V3 | R1, R2 | pass | `git diff dev --stat` lists the four 19-worker-respawn evidence files, the brief copy, `testdata/ste-baseline.json` (one added line, value 0), and the task evidence in `evidence/24-19-evidence-fix/`. The V4.md diff changes only the Commit row. The brief copy differs from `task show --brief` in one sentence (line 15), the STE reword that V3.md records. |
| V4 | R1, R2 | pass | `gofmt -l` printed nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race ./...` exited 0. |

## Runs

```sh
OLD=/home/penguin/source/hive/projects/hive/worktrees/19-worker-respawn/spec/hive-v1/evidence/19-worker-respawn
E=spec/hive-v1/evidence/19-worker-respawn
for f in V1 V2 V3; do diff $E/$f.md $OLD/$f.md && echo same; done
grep -n 39a35e6c $E/V*.md
git diff dev --stat
git diff dev -- $E/V4.md testdata/ste-baseline.json
gofmt -l $(git ls-files -co --exclude-standard '*.go'); go vet ./...; golangci-lint run ./...; go test -race ./...
```

```text
same
same
same
V1.md:29:commit 39a35e6c   (red run)
V1.md:62:commit 39a35e6c   (red run)
V1.md:91:commit 39a35e6c   (red run)
 10 files changed, 371 insertions(+), 28 deletions(-)
V4.md: -| Commit | 39a35e6c |  +| Commit | 2c69b4ad |
gofmt: no output; vet exit 0; golangci-lint 0 issues exit 0; go test -race exit 0
```
