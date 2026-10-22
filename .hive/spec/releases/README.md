# Hive releases

## Destination

The owner pushes an annotated tag `vX.Y.Z` on a commit of `main`. A GitHub workflow runs the gate, builds the Linux archives, checks them, and publishes one GitHub release. A user installs hive from that release or with `go install`, and keeps the home at the same tag. The six old Fledge releases stay, with a title and a top line that say that they do not contain hive.

## Problem

- `.github/workflows/ci.yml` runs the gate on each push and each pull request to `main` and `dev`. No workflow builds or publishes a binary.
- The tags `v0.0.1` to `v0.0.6` and their GitHub releases are Fledge, the predecessor of hive. The text of each release links to pull requests of `Harrison-Blair/fledge`.
- The releases `v0.0.4` to `v0.0.6` each have `fledge_vX.Y.Z_linux_amd64.tar.gz`, `fledge_vX.Y.Z_linux_arm64.tar.gz`, and `checksums.txt`. The releases `v0.0.1` and `v0.0.2` have `fledge_X.Y.Z_linux_ARCH.tar.gz` and `SHA256SUMS`. The release `v0.0.3` has no files.
- The tag `v0.0.6` points at `928cfa95`. This commit is the current `origin/main`.
- At each Fledge tag, `go.mod` gives the module path `github.com/Harrison-Blair/fledge`. Now it gives `github.com/Harrison-Blair/hive`. Thus `go install github.com/Harrison-Blair/hive@latest` finds no valid hive version.
- Commit `e45979f4` added the Fledge release pipeline, and commit `3dd44730` removed it. It had these parts:
  - `release.yml`: a `workflow_dispatch` with a patch, minor, or major choice, and the gate jobs.
  - `package.sh`: reproducible Linux amd64 and arm64 archives with `-trimpath`, the version through `-ldflags -X`, and `sha256sum` checksums.
  - `release.py`: it made a draft release, uploaded the files, and then published the release.
  - `fledge update`: it replaced the binary after a SHA-256 check.
- `internal/lib/version.Version` is `"dev"` when `-ldflags -X` does not set it. `hive version` adds the VCS commit and the modified flag from the Go build information.
- `hive version` does not read the module version of the Go build information. Thus a build from `go install ...@vX.Y.Z` reports `dev`.
- Hive needs a home, and the home is a checkout of this repository. `hive init` fails when no `.hive/README.md` marker is above the working directory.
- The skills, `AGENTS.md`, and the specs come from the checkout. A binary alone cannot run hive.
- Each release tag comes from `main`. `main` gets hive only when pull request #47 merges.
- `hive doctor` already shows a supervisor that runs an older binary.
- The repository has no root `README.md`.

## Out of scope

- macOS and Windows archives.
- A GitHub build attestation.
- A `hive update` command that replaces the binary.
- Release text that a person writes by hand for each release.
- The deletion of a Fledge tag or a Fledge release.
- A change to `main` or to pull request #47.

## Decisions

The user approved option A of each decision R1 to R9 on 2026-10-22.

### Start and version

1. R1. The owner starts a release. The owner pushes an annotated tag on a commit of `main`. The workflow starts on the push of a tag that matches `v*`.
2. R1. The workflow refuses a tag when its name does not match `vX.Y.Z` or `vX.Y.Z-rc.N`. X, Y, Z, and N are decimal numbers with no leading zero.
3. R1. A tag with the suffix `-rc.N` makes a prerelease. A tag with no suffix makes a release.
4. R1. The workflow refuses a lightweight tag. It refuses a tag when the tag commit is not on `origin/main`.
5. R2. The first hive version is `v0.1.0`. It sorts above `v0.0.6`. Thus `@latest` and the GitHub "Latest" mark go to hive.

### Files and build

6. R3. The release has Linux archives for amd64 and arm64 only.
7. R4. The release has three files:
   - `hive_vX.Y.Z_linux_amd64.tar.gz`
   - `hive_vX.Y.Z_linux_arm64.tar.gz`
   - `checksums.txt`
8. R4. Each archive holds the binary `hive` and `LICENSE`. `checksums.txt` holds the `sha256sum` line of each archive.
9. R4. The build uses `CGO_ENABLED=0` and `-trimpath`. It sets `internal/lib/version.Version` to the tag with `-ldflags -X`.
10. R4. The archives are reproducible. Two package runs on the same commit with the same tag and the same Go version make the same bytes. The archive time of each file comes from the tag commit. The owner and the group in the archive are 0. The file order is by name.
11. The script `.github/scripts/package.sh` builds and packages the archives. It comes from the Fledge script, with the name `hive` in place of `fledge`.

### Workflow

12. The workflow is `.github/workflows/release.yml`. It does these steps in this order:
    1. It checks the tag with decisions 2 to 4.
    2. It runs the gate of `ci.yml`.
    3. It builds and packages the archives with decision 11.
    4. It checks the archives.
    5. It makes a draft release, uploads each file, and then publishes the release.
13. The release workflow runs the gate jobs of `ci.yml` before it builds. `ci.yml` accepts `workflow_call` for this. The current triggers of `ci.yml` do not change.
14. The workflow checks the binary of each archive on a runner with a matching architecture. `hive version` must print the tag. The workflow also checks each archive against `checksums.txt`.
15. A failed step leaves no published release. The workflow publishes the draft only after it uploads each file.

### Install and update

16. R5. A user installs hive with these steps:
    1. Clone the repository at the tag. The clone is the home.
    2. Get the binary of the same tag. Use the release archive and check it against `checksums.txt`, or run `go install github.com/Harrison-Blair/hive@vX.Y.Z`.
    3. Put the binary on `PATH`.
    4. Run `hive init` in the clone.
17. R5. `hive version` reports the module version of the Go build information when `-ldflags -X` does not set `Version`. Thus a `go install ...@vX.Y.Z` build reports `vX.Y.Z`. A build from a checkout with no module version keeps `dev`.
18. R6. A user updates hive with these steps:
    1. Run `git -C HOME fetch --tags`.
    2. Run `git -C HOME checkout vX.Y.Z`.
    3. Replace the binary with the binary of the same tag, from the archive or with `go install`.
    4. Restart the supervisor and the coordinator from Health › Agents.
19. R6. A later task adds a `hive doctor` check. The check compares the version of the binary with the tag of the home checkout. It shows a warning when the two differ.

### Fledge releases, release text, and README

20. R7. Each Fledge tag and release stays. The title of each Fledge release becomes `Fledge vX.Y.Z`.
21. R7. The first line of each Fledge release text says this: the release is the Fledge predecessor of hive, and it does not contain hive.
22. R8. The workflow makes the release text with the GitHub generated text, with `generate_release_notes`. A short install header goes above the generated text. The header links to the install steps.
23. R9. The repository gets a root `README.md`. It tells what hive is, how to install hive, and how to update hive. It links to `.hive/spec/hive-v1/README.md`.

### Approvals

24. No task pushes a tag. No task edits a GitHub release. No task runs `gh release`.
25. The user approves the edit of the Fledge releases with its own grant. The user approves the first tag with a separate grant, or the owner pushes it.

## Tasks

| Wave | Task | Decisions | Depends on |
| --- | --- | --- | --- |
| 1 | Spec: this README, in `steFiles` with a baseline of 0. | all | none |
| 2 | Version: `hive version` reports the module version. Test-first. | 17 | wave 1 |
| 2 | Package and workflow: `.github/scripts/package.sh`, `.github/workflows/release.yml`, and `workflow_call` in `ci.yml`. Check with a local package run, `actionlint`, and the archive checks. Push no tag. | 1 to 15, 22 | wave 1 |
| 2 | Docs: the root `README.md`, and the install and update steps. | 16, 18, 23 | wave 1 |
| 3 | Fledge releases: apply R7 with `gh release edit`, under its own grant. | 20, 21, 25 | wave 1 |
| 4 | First release: after pull request #47 merges into `main`, the owner pushes `v0.1.0`, or the manager pushes it under a grant. Check the workflow run, the files, the checksums, and an install from the release in a new home. | 5, 16, 25 | wave 2, pull request #47 |
| 5 | Doctor check: the `hive doctor` check of R6. | 19 | wave 4 |

The tasks of wave 2 can run in parallel on `dev`. Each release friction goes in `reference/dogfood/friction/`.

## Risks

| Risk | Result | Control |
| --- | --- | --- |
| The Go module proxy has the Fledge tags with the module path `github.com/Harrison-Blair/fledge`. | `go install github.com/Harrison-Blair/hive@v0.0.6` fails, and the proxy keeps each version that it fetched. | Decision 5 starts hive at `v0.1.0`. The install steps give an explicit version. The first release checks `go install` of that version. |
| A user runs a new binary with an older home, or an older binary with a new home. | The skills and the specs of the home do not agree with the commands of the binary. | Decisions 16 and 18 move the home and the binary to the same tag. Decision 19 adds a warning. |
| The owner pushes a tag on a commit that is not on `main`. | The release contains code that has no pull request approval. | Decision 4 refuses the tag before the gate. |
| A step fails after the workflow makes the draft. | A draft with some files stays on GitHub. | Decision 15 publishes only after each upload. A draft is not public. |
| The workflow pushes a wrong tag or edits a Fledge release with no approval. | A public change to GitHub that a user can see. | Decisions 24 and 25. |
