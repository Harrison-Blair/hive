# 52-readme independent verification

- Verifier: `t52-verifier-1` (pi).
- Date: 2026-10-22.
- Checked commit: `0a3aa18459b67dc7da7f89202461b4a5f8f84307`.
- Verdict: pass.

| Item | Result | Observation |
| --- | --- | --- |
| V1 README / AC1 | pass | Required sections and links are present. Local clone, source install, and init commands exit 0 with a scratch config. The real home and config do not change. |
| V-gate Gate | pass | Merge of dev reports already up to date. All five gate commands exit 0. Format and module checks print nothing. |

## V1 document check

Read `README.md` with decisions 16, 18, and 23 of the releases spec.
The introduction agrees with the hive v1 introduction, its goals, and the
Project section of `AGENTS.md`. The manager accepts those sources because
hive v1 has no Destination section.

- Requirements name Linux amd64 or arm64, herdr 0.9.5 (protocol 22), and
  Go 1.27.2 for a source build. The Go version agrees with `go.mod`.
- Install has the exact heading `## Install`. The four steps clone at a tag,
  get the matching binary, put it on PATH, and run init in the clone.
- Versions are explicit: `vX.Y.Z`, with `v0.1.0` as an example. No install
  command uses `@latest`.
- Commands are in code blocks. The archive check is exactly
  `sha256sum --check --ignore-missing checksums.txt`.
- Update gives fetch, checkout, binary replacement, and restart from
  Health › Agents. The `r` key agrees with `internal/tui/frame.go`.
  The manager accepts `HOME_DIR` as the path placeholder.
- Links point to the hive v1 spec, `AGENTS.md`, and `reference/`.
- The root README is not in `steFiles`. The brief has a baseline of 0.

## V1 live checks

Scratch root in the worktree: `.hive/tmp/52-verify/` (`S` below).
The manager permits an external clone for init because hive uses the
outermost home marker. External scratch root (`O` below):
`/tmp/hive-52-verifier-3C6UTU`.

Ran these commands from the worktree root. All exit 0:

```sh
git clone --quiet --depth 1 --branch v0.0.6 https://github.com/Harrison-Blair/hive.git "$S/tagged-clone"
git -C "$S/tagged-clone" describe --tags
git clone --quiet --branch task/52-readme "$PWD" "$O/hive"
(cd "$O/hive" && GOBIN="$S/bin" go install .)
(cd "$O/hive" && XDG_CONFIG_HOME="$S/config" "$S/bin/hive" init)
XDG_CONFIG_HOME="$S/config" "$S/bin/hive" version
git -C "$O/hive" fetch --tags
git -C "$O/hive" checkout v0.0.6
git -C "$O/hive" describe --tags
```

The remote clone and update checkout report `v0.0.6`. This is a Fledge tag;
it checks the tagged-clone command only. No hive release exists yet. The
current source build and init check use the task branch in the external
clone. Init prints:

```text
Initialized hive home at /tmp/hive-52-verifier-3C6UTU/hive
```

The scratch config contains exactly:

```toml
home = "/tmp/hive-52-verifier-3C6UTU/hive"
```

The installed binary reports:

```text
hive dev (commit 0a3aa18459b6, go1.27.2)
```

Also made a local archive with the built binary and LICENSE. The checksum
file lists that archive and an absent arm64 archive. Ran the documented
checksum, extraction, and install commands with a scratch binary directory.
The checksum prints `hive_vX.Y.Z_linux_amd64.tar.gz: OK`. Extraction and
`install -m 0755` exit 0. The installed binary reports the same version.
Appended corrupt bytes to the archive: the same checksum command prints
`FAILED` and exits 1. Restored the archive: the check prints `OK` and exits 0.
This checks checksum failure and success independently of a release download.

Compared the real config digest, real home HEAD, and real home Git status
before and after the live checks. The records are equal. The config file
`~/.config/hive/config.toml` is absent in both records. The real home HEAD is
`19988d4c6eec6c13d0b96c953fbfe7e268116a09`, with an empty Git status.

Limits: no hive release archive or module version is available for this
check. Did not download a hive release or install a tagged hive module.
Did not restart live agents. This task changes documentation only; no screen
or board check applies. No tag, release, or workflow was changed.

## V-gate

At the start of the gate, ran:

```sh
git fetch origin dev
git merge origin/dev
```

Fetch exits 0. Merge prints `Already up to date.` and exits 0.
`origin/dev` is `19988d4c6eec6c13d0b96c953fbfe7e268116a09`.

| Exact gate command | Exit | Output |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | 0 | empty |
| `go mod tidy -diff` | 0 | empty |
| `go vet ./...` | 0 | empty |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `go test -race ./...` | 0 | All packages pass or have no test files. Some package results are cached. Root: 22.478s; TUI: 42.344s. |

No gate retry was needed. No Go code or test changed in this task, so no
feature test mutation applies. The archive corruption check has both a
failed check and a successful check.

## Cleanup

Removed both scratch roots after recording these results. Checked that
`.hive/tmp/52-verify/` and `/tmp/hive-52-verifier-3C6UTU` no longer exist.
