# 53-module-case independent verification

Agent: `t53-verifier-1` (pi). Date: 2026-10-22.
Worker tip: `68c60b68`. Base: `19988d4c`.
Verified merge: `d35a626b349395dfc9d5841fd194018f2d54ca50`.

| Item | Result | Observation |
| --- | --- | --- |
| V1 Clash test | pass | Base fails with eight clashes. Task passes before and after the merge. |
| V2 Module install | pass | Local tag and file proxy install the task commit at v0.9.7. The binary prints help. |
| V-gate Gate | pass | The five commands exit 0. Format and tidy print nothing. A second race run with no test cache also passes. |

## Scope and rename checks

Read the brief with `hive task show 53-module-case --brief`.
Compared the base paths with `git ls-files --with-tree=19988d4c` and
`git ls-tree -r --name-only 19988d4c`. Grouping the base paths by lower case
finds exactly eight clashes. `git diff --name-status -M 19988d4c..68c60b68`
shows exactly eight R100 renames:

| Base path suffix | Task path suffix |
| --- | --- |
| 34-screen-chat/evidence/screens/chat-G.ansi | 34-screen-chat/evidence/screens/chat-shift-g.ansi |
| 34-screen-chat/evidence/screens/chat-G.png | 34-screen-chat/evidence/screens/chat-shift-g.png |
| 35-screen-work-goals/evidence/verifier3/G.ansi | 35-screen-work-goals/evidence/verifier3/shift-g.ansi |
| 35-screen-work-goals/evidence/verifier3/G.json | 35-screen-work-goals/evidence/verifier3/shift-g.json |
| 35-screen-work-goals/evidence/verifier3/G.txt | 35-screen-work-goals/evidence/verifier3/shift-g.txt |
| 35-screen-work-goals/evidence/verifier4/G.ansi | 35-screen-work-goals/evidence/verifier4/shift-g.ansi |
| 35-screen-work-goals/evidence/verifier4/G.json | 35-screen-work-goals/evidence/verifier4/shift-g.json |
| 35-screen-work-goals/evidence/verifier4/G.txt | 35-screen-work-goals/evidence/verifier4/shift-g.txt |

All paths above start with `.hive/spec/tui-v2/tasks/`.
Compared the bytes from `git show BASE:OLD` and `git show TASK:NEW` for each
rename: all eight are identical. Compared every other base `.txt` and `.json`
under `.hive/spec/tui-v2/tasks/` with the task: all are identical. This includes
the old diff stats and scope records.

The sole historical Markdown edit changes three `chat-G` references to
`chat-shift-g` in task 34 evidence/V2.md. The other task changes are the new
brief, test, evidence, and the zero STE baseline entry. `ste_test.go` already
includes `.hive/spec/*/tasks/*/brief.md`; no edit is needed there.

Evidence for the key: task 34 evidence/V2.md pairs `G` with the end-of-feed
capture in all three changed rows. Task 35 evidence/V2.md pairs `G` with the
last goal. In verifier3, `g.txt` selects the first goal, while `shift-g.txt`
selects the last goal. In verifier4/live.py line 13, `('G',['G'])` sends the
upper-case key, while `('g',['g'])` sends the lower-case key. These records
support the shift-g names. No UI code or screen content changes in this task.

The test invokes `git ls-files -co --exclude-standard`, uses
`strings.ToLower`, names both clashing paths, and excludes `projects/`.

## V1: red and green

Commands from the worktree root:

```sh
git worktree add --detach .hive/tmp/verify53/base 19988d4c
cp module_paths_test.go .hive/tmp/verify53/base/
(cd .hive/tmp/verify53/base && go test -run TestNoLetterCaseClash -v .)
go test -run TestNoLetterCaseClash -v .
```

Worktree creation exits 0 at `19988d4c`. Copy exits 0.
The base test exits 1:

```text
=== RUN   TestNoLetterCaseClash
    module_paths_test.go:24: paths differ only in letter case: .hive/spec/tui-v2/tasks/34-screen-chat/evidence/screens/chat-G.ansi and .hive/spec/tui-v2/tasks/34-screen-chat/evidence/screens/chat-g.ansi
    module_paths_test.go:24: paths differ only in letter case: .hive/spec/tui-v2/tasks/34-screen-chat/evidence/screens/chat-G.png and .hive/spec/tui-v2/tasks/34-screen-chat/evidence/screens/chat-g.png
    module_paths_test.go:24: paths differ only in letter case: .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3/G.ansi and .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3/g.ansi
    module_paths_test.go:24: paths differ only in letter case: .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3/G.json and .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3/g.json
    module_paths_test.go:24: paths differ only in letter case: .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3/G.txt and .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3/g.txt
    module_paths_test.go:24: paths differ only in letter case: .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier4/G.ansi and .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier4/g.ansi
    module_paths_test.go:24: paths differ only in letter case: .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier4/G.json and .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier4/g.json
    module_paths_test.go:24: paths differ only in letter case: .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier4/G.txt and .hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier4/g.txt
--- FAIL: TestNoLetterCaseClash (0.01s)
FAIL
FAIL github.com/Harrison-Blair/hive 0.025s
FAIL
```

Task test exits 0:

```text
=== RUN   TestNoLetterCaseClash
--- PASS: TestNoLetterCaseClash (0.01s)
PASS
ok github.com/Harrison-Blair/hive 0.023s
```

After the merge, `go test -count=1 -run TestNoLetterCaseClash -v .` also
exits 0: PASS, package time 0.038s.

## V2: local file proxy

Commands from the worktree root, after the merge:

```sh
git worktree remove --force .hive/tmp/verify53/base
git clone --bare . .hive/tmp/verify53/repo.git
git -C .hive/tmp/verify53/repo.git tag v0.9.7 d35a626b
printf '[url "file://%s/.hive/tmp/verify53/repo.git"]\n\tinsteadOf = https://github.com/Harrison-Blair/hive\n' "$PWD" > .hive/tmp/verify53/gitconfig
env GIT_CONFIG_GLOBAL="$PWD/.hive/tmp/verify53/gitconfig" GOMODCACHE="$PWD/.hive/tmp/verify53/proxy-cache" GOPROXY=direct GOPRIVATE=github.com/Harrison-Blair/hive GOFLAGS=-modcacherw go mod download -json github.com/Harrison-Blair/hive@v0.9.7
```

Each command exits 0. Clone output:

```text
Cloning into bare repository '.hive/tmp/verify53/repo.git'...
done.
```

The other setup commands print nothing. Download output below uses `S` for
`/home/penguin/source/hive/projects/hive/worktrees/53-module-case/.hive/tmp/verify53`:

```text
go: downloading go1.27.2 (linux/amd64)
{
  "Path": "github.com/Harrison-Blair/hive",
  "Version": "v0.9.7",
  "Info": "S/proxy-cache/cache/download/github.com/!harrison-!blair/hive/@v/v0.9.7.info",
  "GoMod": "S/proxy-cache/cache/download/github.com/!harrison-!blair/hive/@v/v0.9.7.mod",
  "Zip": "S/proxy-cache/cache/download/github.com/!harrison-!blair/hive/@v/v0.9.7.zip",
  "Dir": "S/proxy-cache/github.com/!harrison-!blair/hive@v0.9.7",
  "Sum": "h1:YemwtDe+GU0HWa873G3yaLqdLpitLmVpzPiOtBVp0Yg=",
  "GoModSum": "h1:7pe/zkQFWcI3RUR5+C34RelPTWPslqqoQqgUxk5nveQ=",
  "Origin": {
    "VCS": "git",
    "URL": "https://github.com/Harrison-Blair/hive",
    "Hash": "d35a626b349395dfc9d5841fd194018f2d54ca50",
    "Ref": "refs/tags/v0.9.7"
  }
}
```

Install uses a separate empty cache. Hive comes from the file proxy.
The public proxy supplies the toolchain and dependencies.

```sh
env GOMODCACHE="$PWD/.hive/tmp/verify53/install-cache" GOPROXY="file://$PWD/.hive/tmp/verify53/proxy-cache/cache/download,https://proxy.golang.org" GOPRIVATE=github.com/Harrison-Blair/hive GONOPROXY=none GOFLAGS=-modcacherw GOBIN="$PWD/.hive/tmp/verify53/bin" go install github.com/Harrison-Blair/hive@v0.9.7
```

Exit 0. Output:

```text
go: downloading github.com/Harrison-Blair/hive v0.9.7
go: downloading go1.27.2 (linux/amd64)
go: downloading charm.land/bubbles/v2 v2.2.1
go: downloading github.com/spf13/cobra v1.10.2
go: downloading charm.land/bubbletea/v2 v2.1.0
go: downloading github.com/charmbracelet/x/ansi v0.11.9
go: downloading github.com/charmbracelet/x/term v0.2.2
go: downloading charm.land/lipgloss/v2 v2.0.6
go: downloading charm.land/glamour/v2 v2.0.1
go: downloading github.com/yuin/goldmark v1.7.8
go: downloading github.com/spf13/pflag v1.0.9
go: downloading golang.org/x/sys v0.47.0
go: downloading github.com/lucasb-eyer/go-colorful v1.4.1
go: downloading github.com/charmbracelet/colorprofile v0.4.3
go: downloading github.com/charmbracelet/ultraviolet v0.0.0-20261008173134-6b8d4baf91b4
go: downloading github.com/muesli/cancelreader v0.2.2
go: downloading github.com/microcosm-cc/bluemonday v1.0.27
go: downloading github.com/alecthomas/chroma/v2 v2.14.0
go: downloading github.com/yuin/goldmark-emoji v1.0.5
go: downloading github.com/charmbracelet/x/exp/slice v0.0.0-20250327172914-2fdc97757edf
go: downloading golang.org/x/text v0.24.0
go: downloading github.com/clipperhouse/displaywidth v0.11.0
go: downloading github.com/mattn/go-runewidth v0.0.30
go: downloading github.com/clipperhouse/uax29/v2 v2.7.0
go: downloading github.com/rivo/uniseg v0.4.7
go: downloading github.com/atotto/clipboard v0.1.4
go: downloading github.com/charmbracelet/x/termios v0.1.1
go: downloading github.com/charmbracelet/x/windows v0.2.2
go: downloading github.com/xo/terminfo v1.0.0
go: downloading golang.org/x/sync v0.22.0
go: downloading github.com/aymerick/douceur v0.2.0
go: downloading golang.org/x/net v0.39.0
go: downloading github.com/dlclark/regexp2 v1.11.0
go: downloading github.com/gorilla/css v1.0.1
```

`.hive/tmp/verify53/bin/hive --help` exits 0. Output starts:

```text
hive

Usage:
  hive [flags]
  hive [command]
```

`git ls-remote --tags origin refs/tags/v0.9.7` exits 0 and prints nothing.
No tag was pushed. No release command was run.

## V-gate: merge and gate

At the start of V-gate, ran `git fetch origin`, then one
`git merge --no-edit origin/dev`. Origin/dev was `77deb3ec` (task 52 README).
The merge had one conflict, in testdata/ste-baseline.json. Retained both zero
entries for task 52 and task 53. Committed the merge as `d35a626b` with
`git add testdata/ste-baseline.json && git commit --no-edit`.
This conflict is a merge note. No task source repair was needed.

| Command | Output | Exit |
| --- | --- | --- |
| `gofmt -l $(git ls-files -co --exclude-standard '*.go')` | none | 0 |
| `go mod tidy -diff` | none | 0 |
| `go vet ./...` | none | 0 |
| `golangci-lint run ./...` | `0 issues.` | 0 |
| `go test -race ./...` | all packages pass; some cached | 0 |
| `go test -count=1 -race ./...` | all packages pass; none cached | 0 |

The fresh race run includes root (23.964s), cmd (7.473s), store (28.461s),
TUI (40.856s), and TUI components (20.556s). No flake or retry was needed.

## Cleanup

The scratch base worktree was removed before V2. After the results were
recorded, removed `.hive/tmp/verify53`, including the bare clone, local tag,
Git config, proxy cache, install cache, binary, and logs. Checked that the
folder no longer exists and that no scratch worktree remains. Only this
verification record and the required dev merge are verifier changes.
