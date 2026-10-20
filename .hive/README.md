# Hive home

This directory marks its parent directory as a hive home. Git tracks this file
and `.hive/spec/`. It ignores the other paths in this directory. Hive never
creates a home. A clone of this repository is a home because this file is in it.

The home uses these subdirectories:

| Path | Content |
| --- | --- |
| `state/` | The state store: one JSON file for each record, and the lock file |
| `logs/` | Log events, one JSON line each, in one file for each day |
| `tmp/` | Scratch files |
| `spec/` | Feature specs, task files, and evidence |

`spec/hive-v1/03-architecture.md` gives the full home layout.
