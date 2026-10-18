# Hive home

This directory marks its parent directory as a hive home. Git tracks this file
and ignores everything else in this directory. Hive never creates a home. A
clone of this repository is a home because this file is in it.

Hive creates these subdirectories on first use:

| Path | Content |
| --- | --- |
| `state/` | The state store: one JSON file for each record, and the lock file |
| `logs/` | Log events, one JSON line each, in one file for each day |
| `tmp/` | Scratch files |

`spec/hive-v1/03-architecture.md` gives the full home layout.
