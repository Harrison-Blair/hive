# Help

Help is an overlay over the current screen. `?` opens it from each screen, and from an empty composer on Chat. `?` or `esc` closes it. `↑` and `↓` scroll it. Each other key does nothing.

## Rules

1. The overlay is a rounded box with the `accent` border at the center of the screen. Its title is `help · esc or ? closes · ↑↓ scroll`. The screen behind it is `border`.
2. The body has two columns. The left column lists SCREENS, CHAT, and ASK CARD. The right column lists AGENTS, WORK, FRICTION, HEALTH, and LOGS. The last rows list STATUS ROW with the segments and the config keys. Each section name is `accent` bold. Each key is `fg` bold. Each action is `dim`.
3. The SCREENS section lists each chord of README decision 16, `esc`, `?`, and `ctrl+c ×2`. When the config rebinds a chord, the overlay shows the bound key.
4. When the body is taller than the box, the box scrolls with the `┃` thumb on its right border.

## States

| State | Trigger | What shows | Board |
| --- | --- | --- | --- |
| help | `?` | rules 1 to 3 | `GfHelp` |
| scrolled | the body is taller than the box | rule 4 | no board |
