# 48-spec-migration verification

| Item | Value |
| --- | --- |
| Task | 48-spec-migration |
| Commit verified | fa7c3345 (implementation 21edabc4) |
| Base | 89ceb001 (`dev`, also the merge base) |
| Date | 2026-10-19 |
| Agent | t48-verifier-1, Claude Code, claude-opus-5-5 |
| Verdict | pass |

## Results

| ID | Result | Observation |
| --- | --- | --- |
| V1 | pass | Own script (below) over the approved table `/home/penguin/source/hive/.hive/tmp/reports/h2-migration-table.md`: 109 move/copy rows. For each of the 96 `move` rows the source is tracked on `dev` and absent in the worktree. A file row has a tracked target. A folder row has the same set of file names under the target as under the source on `dev`. For each of the 13 `copy` rows the target is tracked. `.hive/state/` rows (stay, delete) were not checked; they belong to the manager. 46 and 47 briefs and evidence are under `.hive/spec/home-spec-planning/tasks/`. See the note on copy content. |
| V2 | pass | `git log --follow --oneline` shows commits from before 21edabc4 for briefs `ci/tasks/11-ci-workflow` (81c7ad0e), `hive-v1/tasks/T05-herdr-client-and-doctor` (b81600f6 and later), `repair-flow/tasks/18-task-reopen` (c3c7e1ee), `home-spec-planning/tasks/46-task-feature` (9ab94f2b), and evidence files `T12-land/evidence/V1.md` (64f54cb0), `36-nudge-no-home/evidence/VERIFIED.md` (5f3b131c, fb03fb1f), `47-brief-copy/evidence/V1.md` (89b40b66). |
| V3 | pass | `git ls-files spec/hive-v1 evidence` prints nothing. The grep prints 1444 lines at fa7c3345 (the exclude `':!.hive/spec/*/tasks'` does not remove the files below each task folder). Outside `.hive/spec/*/tasks/`, each hit is a new path that contains the substring `spec/hive-v1`: `.hive/README.md:16`, `AGENTS.md:30,36,76,131`, `ste_test.go:25`, `testdata/ste-baseline.json` keys. All other hits are in past briefs and evidence (history, kept as is), or in this task's own evidence, which quotes the grep. `V3-grep-hits.md` gives a reason per file. |
| V4 | pass | `git check-ignore -q .hive/spec/hive-v1/README.md; echo $?` prints `1`. `git diff dev --stat --find-renames`: 378 files, 351 renames (347 at R100; README R088, tracker R099, BRIEF-FORMAT R090, EVIDENCE-FORMAT R086, all Q3 link edits), 21 added (the 13 copies, this brief, this task's evidence), 6 modified: `.agents/skills/hive-worker/SKILL.md`, `.hive/README.md`, `AGENTS.md`, `STE-GLOSSARY.md`, `ste_test.go`, `testdata/ste-baseline.json`. Each is in the source table (Q3). `.gitignore`, `gitignore_test.go`, and the verifier skill already have the `.hive/spec/` changes on `dev`. |
| V5 | pass | `gofmt -l` printed nothing. `go vet ./...` exit 0. `golangci-lint run ./...` `0 issues.`, exit 0. `go test -race ./...` exit 0. |

## R4 check

`steFiles` globs are `.hive/spec/hive-v1/05-tracker.md`, `.hive/spec/*/tasks/*/brief.md`, `.hive/spec/*.md`. Each baseline value is 0 on `dev` and on the task branch. No count rose.

## Note on copy content (for the manager)

10 of the 13 `copy` targets differ from their `.hive/state/briefs/` source: 3, 5, 7, 8, 9, 10, 20, 21, 25, 41. 4, 6, and 26 are byte-identical. The worker reworded sentences so that the new files pass the STE check at baseline 0, and recorded each change in `brief-copy-changes.md` (all 10 files are listed there). The brief gives this permission in R5 for "the copy". The table action is `copy`, and D7 says the tracked copy shows the final brief. Most rewordings keep the meaning. One drops a word with meaning: in `sidebar-layout/tasks/3-manager-adopts-parent/brief.md` AC7, "lists exactly: ..." became "must list ..." with no "exactly". The V items do not check copy content, so this note does not change a result. The manager decides if the copies must be byte-identical.

## V1 script

```python
import re,subprocess,pathlib,filecmp,sys
T='/home/penguin/source/hive/.hive/tmp/reports/h2-migration-table.md'
H=pathlib.Path('/home/penguin/source/hive')
tracked=set(subprocess.run(['git','ls-files'],capture_output=True,text=True).stdout.split())
devfiles=set(subprocess.run(['git','ls-tree','-r','--name-only','dev'],capture_output=True,text=True).stdout.split())
n=f=0
for l in open(T):
    m=re.match(r'^\| `([^`]+)` \| (`([^`]+)`|\(none\)) \| (\w+) \|',l)
    if not m: continue
    s,t,a=m.group(1),m.group(3),m.group(4)
    if a in ('stay','delete'): continue
    n+=1; errs=[]
    if a=='move' and s.endswith('/'):
        old={x[len(s):] for x in devfiles if x.startswith(s)}
        new={x[len(t):] for x in tracked if x.startswith(t)}
        if not old: errs.append('source folder not tracked on dev')
        if pathlib.Path(s).exists(): errs.append('source present')
        if old!=new: errs.append(f'file sets differ: {sorted(old^new)}')
    elif a=='move':
        if s not in devfiles: errs.append('source not tracked on dev')
        if pathlib.Path(s).exists(): errs.append('source present')
        if t not in tracked: errs.append('target not tracked')
    else:
        if t not in tracked: errs.append('target not tracked')
        elif not filecmp.cmp(H/s,t,shallow=False): errs.append('content differs from state')
    if errs: f+=1; print('FAIL',a,s,'->',t,errs)
print(f'{n} move/copy rows, {f} fail')
```

Output (the 10 lines are the content comparison only; each target exists and is tracked):

```text
FAIL copy .hive/state/briefs/3-manager-adopts-parent.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/5-tab-labels.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/7-pr-dev-main.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/8-pr47-body.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/9-pr47-refresh.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/10-pr47-final.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/20-herdr-status-report.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/21-first-prompt-report.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/25-home-spec-review.md -> ... ['content differs from state']
FAIL copy .hive/state/briefs/41-pr47-refresh-2.md -> ... ['content differs from state']
109 move/copy rows, 10 fail
```
