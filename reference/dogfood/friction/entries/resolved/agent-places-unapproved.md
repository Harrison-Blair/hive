---
title: "Agents make scratch directories and workspaces in places the user did not approve"
description: "The guide had no rule for scratch files or agent places, so agents made ~/stenographer-acceptance and started agents there with --directory, each in its own top-level workspace"
status: "Resolved 2026-11-12"
opened: 2026-11-12
decision: "Spawn and bare hive create <home>/tmp/<project>/ and set HIVE_TMP to it; guide note 10 tells agents to put scratch files below $HIVE_TMP, to use their workspace or --branch scratch/NAME, and to ask the user before they use another place"
---

**Issue:** The startup guide had no rule for the place of scratch files or for
the place of an agent. Hive gave an agent no scratch directory. An agent that
needed a directory for its work made one where it chose, and `spawn
--directory` with a plain directory outside the repository gave each such
directory its own top-level herdr workspace.

**Summary:** Agents of the stenographer project made the directory
`~/stenographer-acceptance` and started agents in it with `hive agent spawn
--directory`. The user did not approve that path. Each plain directory gave a
separate top-level herdr workspace.

**Reproduction steps:**
1. Give an agent work that needs a directory of scratch files, with no place named in its brief.
2. The agent makes a directory of its choice, for example below `~`, and runs `hive agent spawn --directory` with it.
3. Herdr shows a new top-level workspace for that directory.
