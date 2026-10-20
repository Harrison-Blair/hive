# Brief copy changes

The source brief permits changes to sentences that fail the STE check. The
copy keeps the meaning of each changed sentence. Its baseline is 0.

| Source wording | Copy wording | Reason |
| --- | --- | --- |
| A feature is `.hive/spec/<feature>/` in the project repo. | A feature is `.hive/spec/<feature>/` in the project repository. | Replace the glossary term `repo`. |
| `One function gives the paths of a task with a feature, relative to the repo root:` | `One function gives the paths of a task with a feature, relative to the project repository root:` | Replace the glossary term `repo`. |
| `The user confirmed it.` | `The user agreed to this plan.` | Replace the glossary term `confirmed` and keep the user approval meaning. |
| `Use temp Git repos for the README check.` | `Use temporary Git repositories for the README check.` | Replace the glossary term `repos`. |

The original V2 check sentence was:

```text
In a temp home (`HIVE_HOME`, temp `XDG_CONFIG_HOME`) with a registered temp repo: create without `--feature`; create with `--feature f` and no README on base; create with `--feature f --spec`; commit `.hive/spec/f/README.md` on base and create with `--feature f`; `task list --feature f`; `task show` for a task with no brief file on base; `task edit --feature g` on an old record with no feature, then again
```

The copy uses these short sentences for the V2 check cell:

```text
Set `HIVE_HOME` and `XDG_CONFIG_HOME` in a temporary home. Register a temporary Git project.<br>Create without `--feature`.<br>Create with `--feature f` while the base README is absent.<br>Create with `--feature f --spec`.<br>Commit `.hive/spec/f/README.md` on the base branch. Then create with `--feature f`.<br>Run `task list --feature f`.<br>Run `task show` for a task without a linked brief on the base branch.<br>Run `task edit --feature g` on an old record without a feature. Repeat the edit.
```

This change removes the six semicolons, replaces `repo`, and keeps the same
setup, actions, and order. The V2 expected-results cell stays unchanged. The R8
and V3 amendments and the source-table authorization in the updated source
brief were copied as given.

`go test . -run TestSTERatchet -count=1` passes. The new task file has 0
violations against its baseline of 0.