# Hive home

This directory marks its parent directory as a hive home. Git tracks this file.
It ignores the other paths in this directory, `.hive/spec/` too. Hive never
creates a home. A clone of this repository is a home because this file is in it.

The home uses these subdirectories:

| Path | Content |
| --- | --- |
| `state/` | The state store: one JSON file for each record, and the lock file |
| `logs/` | Log events, one JSON line each, in one file for each day |
| `tmp/` | Scratch files |
| `spec/` | Feature specs, task files, and evidence: local state that hive copies into a worktree and back |

`.hive/spec/hive-v1/03-architecture.md` gives the full home layout.
