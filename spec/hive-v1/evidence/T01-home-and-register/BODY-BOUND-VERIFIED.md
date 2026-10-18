# T01 body bound: independent verification

To: coordinator.

| Item | Value |
| --- | --- |
| Verdict | PASS for the bounded shared logger fix |
| Pass | `t01-body-bound-verify-01` |
| Date | 2026-10-18 |
| Route | External Pi, `openai-codex/gpt-6.1-sol:high` |
| Branch | `task/T01-log-body-bound` |
| Base | `ebfad2a7658f5f66d40a6cecb3fd6516b2e9b7c1` |
| Base tree | `829eedbea9973108a245c6e95daed59c7ef270af` |
| Test commit | `4fe971a43f0d65bb503624da8698b7fe651a47df` |
| Source commit | `32bdd20c49865aed479359e47206bf2214941740` |
| Entry and sole verdict parent | `e696d9ca2b848464c57ad75cc15d34e2fdbde4ba` |
| Initial local, named tracking, direct remote | Equal to the entry commit |
| Verdict commit | The sole child that adds this file. The ignored report records its full SHA. |
| Report | `.hive/tmp/coordinator/reports/t01-body-bound-verify-01.md` |
| Captures | `.hive/tmp/coordinator/reports/t01-body-bound-verify-01-captures/` |
| Proves | T01 AC5 body semantics and AC7 at the final Go source |

## Authority and scope

The worker report ends with `END T01 BODY BOUND FIX 01`.
I read the report, received brief, source, tests, captures, and old logger guard evidence.
The coordinator attests that it read the complete worker result and that the worker exited with status 0.
The actual provider, model, and reasoning environment match the required route.
This pass started no agent or model child.

The entry worktree was clean.
The exact base-to-worker range has three paths:

- `internal/lib/log/log.go`
- `internal/lib/log/body_bound_test.go`
- `spec/hive-v1/evidence/T01-home-and-register/BODY-BOUND-FOLLOWUP.md`

This verdict adds one path only.
The worker tree has 317 entries. The verdict tree has 318 entries.
All 317 prior entries retain their modes, types, names, and blobs.
No source, test, old evidence, task, tracker, dependency, or baseline changed during this pass.

## Acceptance checklist

| ID | Criterion | Result |
| --- | --- | --- |
| B01 | Decoded body uses at most 4096 bytes | PASS. Real JSONL decode checks pass. |
| B02 | Largest complete-character prefix | PASS. Literal Unicode and independent byte-prefix checks pass. |
| B03 | Fitting valid ASCII and Unicode stay exact | PASS. Below, exact, mixed, and decomposed Unicode controls pass. |
| B04 | Each invalid byte keeps normal JSON replacement semantics | PASS. Mixed invalid sequences and exact replacement boundaries pass. |
| B05 | The flag is true only when the cap shortens the normal decoded sequence | PASS. Cut and unchanged controls pass. |
| B06 | One bounded cap path at `addAttr` | PASS. Source inspection finds one call site. |
| B07 | Ordinary, inherited, grouped, and nested bodies share the policy | PASS. Worker paths and independent nested group attributes pass. |
| B08 | Escapes and other fields stay unchanged | PASS. Exact metadata, 5000-byte non-body fields, and raw string tokens pass. |
| B09 | Clock, path guard, rotation, append, and concurrency stay unchanged | PASS. The diff preserves those paths. Original tests pass. |
| B10 | Tests detect the original defect and controlled regressions | PASS. Red replay and three assertion-red mutations pass sensitivity checks. |
| B11 | Actual shared consumer and CLI behavior | PASS. Real herdr calls and the built CLI pass. |
| B12 | Full uncached worktree gate and STE | PASS. Every gate command exits 0. |
| B13 | Preservation, custody, cleanup, and delivery scope | PASS. Manifests and Git maps retain the prior data. |

No source finding remains open.
No required verification item remains pending.

## Source and test decisions

`capBody` measures the sequence that standard `encoding/json` decoding produces.
`utf8.DecodeRuneInString` consumes each invalid byte separately.
The helper charges three decoded bytes for that byte.
It stops before the next character would exceed 4096 bytes.
The scan examines at most the retained budget plus one character.
The second scan builds only the retained prefix.
The builder requests at most 4096 bytes, subject to allocator size classes.
A fitting input returns unchanged without a builder.

This preserves the existing encoder behavior for fitting invalid strings.
A cut uses U+FFFD for each retained invalid byte.
No NFC, ellipsis, ASCII placeholder, rejection, new API, dependency, or caller workaround appears.
The cap does not bound JSON syntax or other fields.

The worker tests use real writers and real daily files.
Their oracle first applies standard JSON semantics, then uses the literal budget.
Literal expected prefixes also pin euro, four-byte, invalid-byte, and ASCII cases.
The tests neither read the production constant nor copy the helper scan.
All 19 captured JSONL cases have matching observation hashes.
The same test bytes appear in the red commit, final tree, and replay copies.

| File | SHA-256 | Git blob |
| --- | --- | --- |
| Old `log.go` | `cedbe1e03c5ba8c4df2f919e13f750edf2ab0c00baae6187f38a076613ceb90e` | `fbcc1259f8b0633c555a0ae2ce160d8b4cac99aa` |
| Final `log.go` | `2deb992bfa25f199ef4fdcc079398102bd6f4e347d3c5d84af342b3476d90d85` | `a3e0c8be19576cfe570c78ffe3e462b76e8b5b51` |
| New test | `0f0b9237b88344eab9a70dc38210439c69ec63c2531c2f0cc7e461e0687af3d8` | `e7375b1c3292a1b70c7b3397b2d79c53386eff0f` |

## Independent runs

All Go commands use child environments with `TMPDIR=/tmp/t01bv01.FpPfAVmJ/t`.
That path is below the unique owned root and outside every marked home.
`commands.json` records exact arguments, working directories, statuses, durations, and limits.
Raw output, probes, patches, source snapshots, and synthetic JSONL remain in the captures.

The retained original production source and unchanged new test compile together in an owned module.

```sh
go test -count=1 -v -run '^TestBodyBound' .
```

The replay exits 1 on assertions.
Euro produces 4098 decoded bytes. Four-byte unaligned input produces 4102 bytes.
The 2000 invalid bytes produce 6000 bytes without a flag.
Nine leaf cases fail. Four parent tests also report failure.
The same selection passes with final source and race detection.
Both runs retain 19 actual JSONL cases.

| Owned mutation | Result |
| --- | --- |
| Restore old raw slice | Exit 1. Nine leaf failures. |
| Raise the cap to 4099 | Exit 1. Twelve leaf failures. |
| Break the flag | Exit 1. Twelve leaf failures. |
| Restore exact final source | Exit 0. Full copied log suite passes with race detection. |

No compile error, panic, or timeout counts as a red result.
Failure totals of 13 and 18 include parent tests.
The worker mutation record counts only top-level failures in its `fail_lines` fields.
The underlying raw outputs agree.

The independent matrix has 110 inputs across three paths, for 330 cases.
It includes 14 literal inputs and 96 deterministic mixed inputs.
The oracle normalizes with standard JSON, slices at literal byte 4096, then removes an incomplete suffix.
It does not copy the production rune scan.
Checks cover exact decoded prefixes, raw encoded prefixes, invalid bytes, escapes, inherited groups, nested attributes, and metadata.
All 330 cases pass with race detection.

The actual production herdr call logs both a large request and a large reply.
Both retain the exact complete UTF-8 prefix and a true flag.
The operation still returns the full 3000-euro result.
The probe uses an owned project copy and a synthetic Unix socket, not native agents.

The actual built CLI runs text and JSON `project list` against an owned synthetic home.
Text remains `no projects`. The JSON envelope retains `phase: complete` and an empty result array.
Its four log events retain CLI metadata and contain no body.
An initial verifier assertion expected `phase: done`. That expectation was a probe error.
The corrected assertion passes on the same retained output. No product rerun or source change was necessary.

## Full gate

The gate ran once from the actual worktree at final Go source.
Only verification evidence follows that gate.

```sh
gofmt -l $(git ls-files -co --exclude-standard '*.go')
go vet ./...
golangci-lint run --allow-serial-runners ./...
go test -race -count=1 ./...
go test -run TestSTE -count=1 -v .
```

`gofmt` prints nothing. `vet` prints nothing. Both exit 0.
The linter prints `0 issues.` and exits 0.
The full race suite passes every package and exits 0.
The following exact package lines come from `gate-race.txt`:

```text
ok  	github.com/Harrison-Blair/hive/internal/lib/herdr	6.507s
ok  	github.com/Harrison-Blair/hive/internal/lib/log	1.091s
ok  	github.com/Harrison-Blair/hive/internal/lib/store	17.240s
```

The complete package output remains in the captures.
The focused log, herdr, CLI, and STE checks also exit 0.
Both glossaries give zero hard findings for the new report and evidence.
The dependency and STE baseline blobs equal the exact base blobs.
The post-evidence source comparison retains every tested Go byte.

## Custody and preservation

The original frozen inventory has 1552 entries, including its root entry.
The fresh inventory has 1553 entries and omits the root entry.
Their 1551 shared entries match in type, mode, size, inode, owner, and regular-file hash.
Only the fresh inventory includes the original inventory and its hash receipt.
All 47 compact manifest files match their retained hashes.
All original summary hashes match. All five original JSONL fixtures match their observations.
The original probe and raw Unicode failure retain their specified hashes.
The original root remains absent.
These checks support retained content equality and the recorded cleanup receipt.
They do not independently replay the past deletion operation.
Later custody receipts are later records, not entries retroactively added to the original frozen inventory.

The worker final inventory has 195 entries with no hash mismatch.
This pass preserves the entire worker report and capture tree against its entry manifest.
It also preserves all 317 tracked file hashes before the verdict.
Root activation and the separately active T10 worktree are explicitly outside this freeze.
This pass writes neither location and requires no global topology freeze.
The helper worktree and shell remain in place. No other agent resource closes.
Credentials, global harness configuration, and historical attribution remain untouched.

## Cleanup and limits

The owned root is `/tmp/t01bv01.FpPfAVmJ`.
Compact evidence precedes cleanup.
The no-follow ownership manifest records each path, type, mode, size, inode, and regular-file hash.
Cleanup removes only still-matching listed entries below that root.
Cleanup removed 411 listed entries and 14095239 regular-file bytes.
The absence receipt records ENOENT at 2026-10-18T13:48:51Z. The inherited `TMPDIR` remains unset.
This pass removed no old capture, harness log, shared cache, worktree runtime root, or unrelated path.

The approved boundary is local user and cooperating writer.
Hostile replacement between `Lstat` and `Open` remains deferred for T01.
This verdict does not bound total JSON syntax or accept a general logger framework.
It makes no root integration, Task10, checkpoint, or landing acceptance claim.
Recommendation: accept this bounded source fix for the coordinator's separate landing decision.

END T01 BODY BOUND VERIFY 01
