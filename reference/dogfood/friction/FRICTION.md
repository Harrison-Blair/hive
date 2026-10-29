# Friction log

This folder records each bug, missing capability, and workaround that an agent
finds when it uses hive, herdr, or a harness. The section "Friction log" in
`AGENTS.md` gives the procedure. `_template.md` gives the fields of an entry.

Each entry has one line in this file, under `## Open` or `## Resolved`:
`- [TITLE](entries/SLUG.md) — DESCRIPTION`. The title and the description are
the same as in the frontmatter of the entry. `friction_test.go` fails when this
file and the entry files do not agree.

The entries from before the teardown are at the tag `pre-teardown`, below
`reference/dogfood/friction/`.

## Open

## Resolved
