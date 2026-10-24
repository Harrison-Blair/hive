---
title: "Supplemental verifier evidence stays in the worktree"
status: "Open"
opened: 2026-10-24
decision: ""
---

**Issue:** `hive task verify` copies only `VERIFIED.md` to the home.

**Summary:**
A verifier also writes captures, PNGs, diffs, scripts, and run logs in the
task evidence folder. Those files stay in the worktree after verification.
On 2026-10-24, task 146 passed verification. The manager sent home paths for
its review. The coordinator found no `verifier-2/` folder there. The folder
held 251 files in the worktree. The manager copied them to the home and
checked each relative path and SHA-256 hash before worktree removal.

The spec defines the single-file copy. The workflow gives no check that
supplemental evidence reached the home. A worktree removal can delete the
only copy of that ignored evidence. A review link can also name a missing
file before removal.

**Reproduction steps:**
1. In a scratch task, write `VERIFIED.md` and supplemental evidence in the worktree.
2. Run `hive task verify`.
3. Inspect the home evidence folder. It has the verdict but no new supplemental files.
4. Check a home review link to those files. The target does not exist.

**Workaround:** copy the full supplemental evidence folder to the home before
review or worktree removal. Check the relative paths and SHA-256 hashes.
Keep the worktree until that check passes. Keep the manifest with the evidence.

No source fix or new task is requested in this entry.
