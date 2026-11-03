---
title: "herdr does not resize the terminal of a pane in a background tab"
description: "A pane split or resize in an unfocused tab changes the pane rect but not the terminal size, so a live check of a small pane needs the tab in front"
status: "Open"
opened: 2026-11-03
decision: ""
---

**Issue:** In an unfocused tab, `herdr pane split` and the split ratio change
the pane rect that `herdr pane layout` reports, but the terminal of the pane
keeps its old size: `herdr pane get` still reports the old
`viewport_rows`, and the program in the pane gets no resize. A live check
that must see a TUI at a small size, for example the bare `hive` picker in a
48 × 11 pane, cannot get that size in a background tab. Setting the size
with `stty rows 11 cols 48` in the pane makes the program draw for 48
columns while the herdr terminal stays wide, and the Bubble Tea renderer
then draws rows with stray blanks (it moves the cursor with tabs and relies
on the right edge of the terminal), so the pane text is not a faithful view.

**Summary:** On 2026-11-03 the builder `picker-builder` live-checked the
picker of the branch `feat/picker-style`. It opened the tab
`picker-live-check-7f3a` with `--no-focus` so the user's view did not change,
split its pane down with ratio 0.2 and right with ratio 0.24, and got a rect
of 50 × 12. `pane get` kept `viewport_rows` 58 and the picker kept its
62 × 15 box. The builder checked the small size in a detached tmux session
of 48 × 11 instead, where the box shrank and grew back as expected.

**Reproduction steps:**
1. `herdr tab create --workspace W --label T --no-focus`, then run a
   Bubble Tea program in its root pane P.
2. `herdr pane split P --direction down --ratio 0.2 --no-focus`.
3. `herdr pane layout --pane P` shows a height of 12; `herdr pane get P`
   still shows the old `viewport_rows`, and the program does not redraw.
