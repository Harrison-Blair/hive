# Agent graph screenshots

Renders of `hive agent graph` made on 2026-11-25. They are not live
sessions. A test in `internal/agent/graph` built the model with fixed sample
agents and wrote the text of its real `View()`, with its SGR colours. A script
then drew that text to PNG on a fixed cell grid of 10 × 20 pixels, one glyph
in each cell, with a margin of one cell. The script used DejaVu Sans Mono
16.6 px, its bold face for bold text, and DejaVu Sans for the braille spinner,
which DejaVu Sans Mono does not have. Neither the test nor the script is in
the repository.

| File | Mode | Size in cells | Pixels |
| --- | --- | --- | --- |
| `agent-graph-wide.png` | tree | 200 × 48 | 2020 × 1000 |
| `agent-graph-narrow.png` | tree | 80 × 45 | 820 × 940 |
| `agent-graph-list.png` | list | 200 × 48 | 2020 × 1000 |

Each file shows the same 11 sample agents in five pods: `agents` (a
coordinator with two children), `hive` (a coordinator with three children and
one grandchild), `qmeter`, `ror-desktop`, and `home`. The agents cover each
status except unknown and gone, and the harnesses claude, codex, and pi. The
selected card is `graph-builder`, with its heavy outline.

The layout:

- Cards and pods are on a grid of tracks 40 cells apart. A card is 33 cells
  wide and starts 3 cells into its track. A pod of n tracks is 40n - 1 cells
  wide. A row of pods holds (W + 1) / 40 tracks of the screen width W, and a
  pod that does not fit starts a new row.
- Tree mode: each subtree takes one track for each leaf, and a parent is over
  its first child. The join row below a parent has `├`, `┬`, and `╮`, or `│`
  for one child.
- List mode: a pod is one track. Each child starts 3 cells right of its
  parent, is 3 cells narrower, and ends in the column of the root. A rail
  joins the children with a tick in the middle row of each child.
- Header: the title and the status start at column 3, and the clock ends 5
  cells from the right edge.
- Footer: a bar with a dark background. At 200 columns, the legend is on the
  left and the keys are on the right of one row. At 80 columns it has two
  rows. At 80 columns, the `hive` pod is wider than the screen and is cut at
  the right edge, where the board scrolls sideways, and the `home` pod is
  below the pod area.
