---
title: "A new herdr tab reports 120 columns but its terminal has 119"
status: "Open"
opened: 2026-10-23
decision: "Use an isolated larger scratch server for live captures"
---

**Summary:**
On 2026-10-23, verifier `t134-verifier-2` reported a width mismatch during
task `134-agents-hide-managers`. A new tab had a 119-column terminal, but
herdr `pane.layout` reported 120 columns. The live check required 120x40
and 80x24 captures.

The verifier used an isolated herdr 0.9.5 scratch server at 140x48 for the
captures. It ran no TUI on the real home. It will remove the scratch
session after the check.

Source: manager inbox record 546. The verifier will record the exact
commands and dimensions in the task evidence. The cause and repeatability
remain unknown. The manager did not independently repeat this mismatch.

**Reproduction steps:**
1. On herdr 0.9.5, create a new tab for a live TUI capture at 120x40.
2. Read the pane size through `pane.layout`.
3. Check the terminal columns in that pane.
4. Compare the two sizes. The verifier reported 120 and 119 columns.
5. For the capture, use an isolated scratch server larger than the board.

