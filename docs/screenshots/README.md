# Agent graph screenshots

Live captures of `hive agent graph` in a kitty terminal on 2026-10-31. The
agents were real claude sessions started with bare `hive` and
`hive agent spawn` in herdr, with `HIVE_HOME` set to a scratch directory.

| File | Terminal size | Pixels |
| --- | --- | --- |
| `agent-graph-wide.png` | 132 × 36 cells | 1320 × 800 |
| `agent-graph-narrow.png` | 80 × 45 cells | 800 × 1000 |

Both captures show:

- Project coordinator names: bare `hive` in the checkouts of the projects
  `demo` and `docs-site` saved the records `demo-coordinator` and
  `docs-site-coordinator`. Each heads its own project pod.
- The `home` pod with `gs-scout`, an agent with no project coordinator.
- Card row 1: the status glyph, the agent name, and the elapsed time in
  seconds, for example `3m43s`.
- Card row 2: the harness and the model, for example
  `claude · claude-haiku-5-5`.
- The header clock with seconds at the top right, and the time of the last
  refresh below the title.
- A braille spinner frame on `gs-builder`, the one agent with status
  `working`. The other agents show the glyphs of the legend.
- The selected card, `demo-coordinator`, with a highlighted border.

At 132 columns the three pods fit on one row. At 80 columns the `home` pod
wraps to a second row.
