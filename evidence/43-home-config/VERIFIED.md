# 43-home-config verification

| Item | Value |
| --- | --- |
| Task | 43-home-config |
| Commit | 7a6ddba8 |
| Base | `dev` at b16c5dff (merge base) |
| Date | 2026-10-19 |
| Agent | t43-verifier-1, Claude Code, claude-opus-5-5 |
| Verdict | FAIL (V4) |

Each run used a new temp `XDG_CONFIG_HOME` below the session scratchpad and
`HIVE_HOME` unset. `~/.config/hive/` does not exist before or after the runs.

## Results

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | The 5 changed test files on base b16c5dff: 8 home tests and 6 cmd tests fail, each for the expected reason (HIVE_HOME conflict, no config step, no init guidance, `unknown command "init"`). `internal/tui` passes on the base (isolation edit only). Task commit: all 3 packages pass. |
| V2 | pass | All observations match. See V2 below. |
| V3 | pass | `go test -count=1 -race ./...` with a hostile config exits 0. The config atime stays at 2000-01-01, so no test read the file. |
| V4 | fail | `git diff dev --stat` shows 5 files outside the source table: `spec/hive-v1/evidence/43-home-config/V1.md` to `V5.md`. The brief puts evidence in `evidence/<task-id>/` in the worktree. All source files are in the source table. |
| V5 | pass | `gofmt -l` prints nothing. `go vet`, `golangci-lint` (0 issues), and `go test -race ./...` exit 0. |

## V1

```sh
git worktree add --detach .hive/tmp/43-verify-base b16c5dff
for f in $(git diff dev --name-only -- '*_test.go'); do cp $f .hive/tmp/43-verify-base/$f; done
(cd .hive/tmp/43-verify-base && go test -count=1 ./internal/lib/home ./cmd ./internal/tui)
git worktree remove --force .hive/tmp/43-verify-base
go test -count=1 ./internal/lib/home ./cmd ./internal/tui
```

```text
# base
--- FAIL: TestResolveEnvWinsOverAncestor
--- FAIL: TestResolveNoMarker
--- FAIL: TestResolveMarkerMustBeRegularFile
--- FAIL: TestResolveUsesConfigFromUnmarkedDirectory
--- FAIL: TestResolveConfigRequiresMarkedHome
--- FAIL: TestResolveConfigConflictNamesBothHomes
--- FAIL: TestResolveEnvWinsOverConfigAndAncestor
--- FAIL: TestResolveConfigErrorsNameConfigPath
FAIL	github.com/Harrison-Blair/hive/internal/lib/home
--- FAIL: TestInitCreatesConfigForOutermostMarkedHome   (unknown command "init")
--- FAIL: TestInitJSONReportsConfigWriteEffect          (unknown command "init")
--- FAIL: TestInitSameHomeDoesNotChangeConfig           (unknown command "init")
--- FAIL: TestInitRejectsDifferentConfigWithoutWriting  (unknown command "init")
--- FAIL: TestInitWithoutMarkerIgnoresOtherHomeSources  (unknown command "init")
--- FAIL: TestProjectListRejectsLinkedHiveHome          (no "run hive init" text)
FAIL	github.com/Harrison-Blair/hive/cmd
ok  	github.com/Harrison-Blair/hive/internal/tui
exit 1
# task
ok  	github.com/Harrison-Blair/hive/internal/lib/home
ok  	github.com/Harrison-Blair/hive/cmd
ok  	github.com/Harrison-Blair/hive/internal/tui
exit 0
```

## V2

Binary built from 7a6ddba8. `OTHER` is a temp folder with `.hive/README.md`.

```text
$ hive init                                   (worktree)
Initialized hive home at /home/penguin/source/hive
exit 0
config: home = "/home/penguin/source/hive"
$ hive init                                   (worktree)
Hive home already initialized at /home/penguin/source/hive
exit 0
stat and sha1 of the config: unchanged
$ hive --json init
{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":{"home":"/home/penguin/source/hive"}}
exit 0
$ hive task list --project hive               (/tmp)
32 rows, includes "43-home-config completed t43-worker-1 Home config and hive init"
exit 0
# config now names OTHER
$ hive task list --project hive               (worktree)
hive: home conflict: config is OTHER, working directory is in /home/penguin/source/hive
exit 1
$ hive init                                   (worktree)
hive: home conflict: config is OTHER, working directory is in /home/penguin/source/hive
exit 1
cmp with the copy: config unchanged
$ hive --json init
{"outcome":"failed","phase":"command","effects":[],"error":"home conflict: config is OTHER, working directory is in /home/penguin/source/hive","result":null}
exit 1
# config removed
$ hive task list --project hive               (/tmp)
hive: not a hive home: no marker above /tmp; run hive init in the home
exit 1
```

## V3

```sh
# X/hive/config.toml: home = "OTHER", OTHER is a marked temp home, atime set to 2000-01-01
XDG_CONFIG_HOME=X go test -count=1 -race ./...
```

```text
exit 0, 0 lines other than "ok" and "[no test files]"
config atime after the run: 2000-01-01 00:00:00 (unchanged)
```

## V4

```text
$ git diff dev --stat
 .agents/skills/hive-coordinator/SKILL.md   |  20 +--
 cmd/homeinit/homeinit.go                   |  83 ++++++++++++
 cmd/homeinit_root_test.go                  | 205 +++++++++++++++++++++++++++++
 cmd/project_static_home_test.go            |   2 +-
 cmd/root.go                                |   3 +-
 cmd/root_test.go                           |  13 ++
 internal/homeinit/homeinit.go              |  46 +++++++
 internal/lib/home/config.go                |  78 +++++++++++
 internal/lib/home/doc.go                   |   5 +
 internal/lib/home/home.go                  |  27 +++-
 internal/lib/home/home_test.go             | 146 +++++++++++++++++++-
 internal/tui/tui_test.go                   |   2 +
 spec/hive-v1/evidence/43-home-config/V1.md |  88 +++++++++++++
 spec/hive-v1/evidence/43-home-config/V2.md | 104 +++++++++++++++
 spec/hive-v1/evidence/43-home-config/V3.md | 101 ++++++++++++++
 spec/hive-v1/evidence/43-home-config/V4.md |  50 +++++++
 spec/hive-v1/evidence/43-home-config/V5.md | 105 +++++++++++++++
 spec/hive-v1/tasks/43-home-config.md       | 136 +++++++++++++++++++
 testdata/ste-baseline.json                 |   3 +-
 19 files changed, 1195 insertions(+), 22 deletions(-)
```

Difference: the 5 evidence files are below `spec/hive-v1/evidence/`. The brief
gives `evidence/<task-id>/`. Repair: move them to `evidence/43-home-config/`.
The worker V4.md also shows the diff at 912b1d17, before the evidence commit.

## V5

```text
$ gofmt -l $(git ls-files -co --exclude-standard '*.go')
(no output) exit 0
$ go vet ./...
exit 0
$ golangci-lint run ./...
0 issues.
exit 0
$ go test -race ./...
exit 0, no FAIL lines
```
