# Verification: 51-release-workflow

- Verifier: `t51-verifier-1` (Pi).
- Date: 2026-10-22.
- Tested commit: `110a581a0342`.
- Verdict: pass.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Tag checks | pass | All 14 cases passed. When the lightweight-tag refusal was removed, that case failed with exit 0 instead of the required exit 1. The script was restored. All 14 cases passed again. |
| V2 Package | pass | Two package runs produced the same bytes. Names, members, checksums, version, owner, mode, and time passed. When LICENSE was omitted, the test failed with `[hive], want [LICENSE hive]`. The script was restored and the test passed again. Live archive checks passed. |
| V3 Workflow | pass | actionlint printed nothing and exited 0. The workflow meets decisions 12 to 15 and 22. Details are below. |
| V-gate | pass | dev was merged once. The five gate commands passed. The only merge conflict was in the STE baseline. Both zero-count entries were retained. |

## V1

Commands:

```sh
go test -run TestCheckTag -v .
# Remove the lightweight-tag refusal for the red run.
go test -count=1 -run TestCheckTag -v .
# Restore the refusal for the green run.
go test -count=1 -run TestCheckTag -v .
```

The first green run took 0.253 seconds. The red run failed only
`TestCheckTag/lightweight`: `err <nil>, want exit status 1`, with stdout
`release`. The restored green run took 0.408 seconds.

The exact command initially used for the red run returned a cached green
result. The red run was repeated with `-count=1` to execute the changed
external script. No cached result was accepted as red-run evidence.

## V2

Commands:

```sh
go test -run TestPackage -v .
# Omit LICENSE from tar for the red run.
go test -count=1 -run TestPackage -v .
# Restore LICENSE for the green run.
go test -count=1 -run TestPackage -v .
.github/scripts/package.sh v0.1.0-rc.1 .hive/tmp/dist
tar --numeric-owner --full-time -tvzf .hive/tmp/dist/hive_v0.1.0-rc.1_linux_amd64.tar.gz
tar --numeric-owner --full-time -tvzf .hive/tmp/dist/hive_v0.1.0-rc.1_linux_arm64.tar.gz
(cd .hive/tmp/dist && sha256sum --check checksums.txt)
mkdir -p .hive/tmp/v51-host
tar -xzf .hive/tmp/dist/hive_v0.1.0-rc.1_linux_amd64.tar.gz -C .hive/tmp/v51-host
.hive/tmp/v51-host/hive version
```

The first TestPackage run passed in 9.17 seconds. The red run failed in
14.35 seconds because LICENSE was absent. The restored TestPackage run
passed in **14.55 seconds**. No run exceeded 60 seconds.

Both live archives held LICENSE, then hive. Each member had owner/group
`0/0` and the HEAD commit time, Unix `1792644825`. The local tar display
showed `2026-10-22 00:53:45`. LICENSE had mode 0644 and hive had mode 0755.
Both checksum lines reported `OK`.

The host binary printed:

```text
hive v0.1.0-rc.1 (commit 110a581a0342, go1.27.2)
```

The arm64 binary was inspected as archive data, not executed on this amd64
host. The workflow checks it on `ubuntu-24.04-arm`.

## V3

`actionlint .github/workflows/*.yml` printed nothing and exited 0.

Source review against the spec confirmed:

- The only release trigger is a push of a `v*` tag. There is no manual trigger.
- `check-tag` runs before `gate`. `gate` calls `ci.yml`. Packaging needs gate.
- Each archive check needs packaging. The matrix uses ubuntu-latest for
  amd64 and ubuntu-24.04-arm for arm64. Each runner checks the checksums
  and the version of its matching binary.
- Publishing needs the tag result and all archive checks. The release is
  created as a draft with generated notes and an install header. Upload
  precedes the step that removes draft status. A failed upload prevents
  publication.
- The tag script returns prerelease for `-rc.N`; publication sets
  `--prerelease` for that result.
- The install header links to the tagged root README at `#install`.
  The merged README has `## Install` at line 20.
- Top-level contents permission is read. Only publish grants contents write.
  Concurrency is grouped by release ref.
- The only change to ci.yml is the addition of workflow_call. Its existing
  triggers and steps remain unchanged.
- No workflow step pushes a tag.

No tag was pushed. No GitHub release command or remote workflow was run.
The GitHub execution itself remains unverified under the local-only grant.
No user interface changed, so no visual check applies.

## V-gate

At the start of verification:

```sh
git fetch origin dev && git merge --no-edit origin/dev
```

The only conflict was in `testdata/ste-baseline.json`: task 51 and task 52
had each added a zero-count brief entry at the same location. Both entries
were retained. Merge commit: `110a581a0342`.

| Command | Exit | Output |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | 0 | none |
| `go mod tidy -diff` | 0 | none |
| `go vet ./...` | 0 | none |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `go test -race ./...` | 0 | All packages passed; root 35.598s, internal/tui 40.754s. |

No listed flake occurred. The first four gate commands were also repeated
with explicit exit-status output; each exited 0. Both scripts match HEAD
after the mutation checks. Scratch archives, binaries, and logs were
removed after this report was recorded.
