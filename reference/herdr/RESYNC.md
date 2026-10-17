# Resync procedure for a new herdr version

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md). Glossary: [STE-GLOSSARY.md](STE-GLOSSARY.md).

herdr can update itself with `herdr update`. The update channel comes from `herdr channel set <stable|preview>`. Thus the installed binary can be newer than these pages. A pass checks the pages against one herdr version. A schema diff alone is not sufficient, because behavior can change while the schema stays the same.

A pass has these steps: emit the raw artifacts again, compare them, probe each page, update the pages, apply the stamps, and do the checks.

## 1. Check the version

Run these commands:

```sh
herdr --version
herdr api schema | head -4   # shows protocol and schema_version
```

Compare the output with the stamp line at the top of this page. If the version, the protocol, and the schema version are the same, the pages are current. Then stop.

## 2. Emit the raw artifacts and compare them

Write the new artifacts to a scratch directory. Then compare each one with the file in `raw/`:

```sh
herdr api schema --json > /tmp/herdr-schema.json
herdr --skill            > /tmp/herdr-skill.md
herdr --default-config   > /tmp/herdr-default-config.toml
diff raw/schema.json /tmp/herdr-schema.json
diff raw/skill.md /tmp/herdr-skill.md
diff raw/default-config.toml /tmp/herdr-default-config.toml
```

To make the schema diff easier to read, compare the method lists first:

```sh
jq -r '.schemas.request.oneOf[].properties.method.const' raw/schema.json | sort > /tmp/old-methods
jq -r '.schemas.request.oneOf[].properties.method.const' /tmp/herdr-schema.json | sort > /tmp/new-methods
diff /tmp/old-methods /tmp/new-methods
```

Do the same for the event kinds (`.schemas.event."$defs".EventKind.enum[]`). Also do it for the result types (`.schemas.success_response."$defs".ResponseResult.oneOf[].properties.type.const`). Use `jq` only for these local files. A request to the socket does not go through `jq` (refer to step 4).

Also get new copies of `https://herdr.dev/agent-guide.md` and `https://herdr.dev/llms.txt`. Compare them with the files in `raw/`.

## 3. Probe each page

After each version change, probe each claim on each page. Do this also for pages whose methods are not in the schema diff. The shape of a request or a result can stay the same while its behavior changes. Examples of such changes are timing, validation sequence, error-code priority, event names, side effects, and default values. A schema diff does not show these changes.

For each page, do these steps:

1. Read the page.
2. Run each example again.
3. Check each **Errors** item and each **Events** item.
4. Record the result of each claim, with the evidence.

A claim that a pass cannot probe safely gets the status "not probed", with the reason. Do not guess its behavior.

## 4. Probe safely

The socket of the default session can hold a live session. Send only read-only methods to the socket of a live session. These methods are `*.get`, `*.list`, and `session.snapshot`. Send `pane.read` only to panes that you made. Send all other methods to a scratch server.

### Scratch servers

Start a scratch server under a session name that is new and that tells what you probe, such as `demo-probe`. Remove each `HERDR_*` variable from the environment first. A herdr pane has `HERDR_SOCKET_PATH` and other `HERDR_*` variables. A pane of a named session also has `HERDR_SESSION`. These variables can send a CLI command to the wrong server. [environment.md](environment.md#environment-variables) gives the variables.

```sh
env $(env | sed -n 's/^\(HERDR_[^=]*\)=.*/-u \1/p') herdr --session demo-probe server &
# socket: ~/.config/herdr/sessions/demo-probe/herdr.sock
```

[environment.md](environment.md#socket-path-and-permissions) gives the socket path rules. A socket path has a limit of 107 bytes (`sun_path`). If a path is too long, use a short symbolic link to its directory. Delete the link after the pass.

Do not touch a session directory that you did not make. Send `server.stop` and `server.live_handoff` only to a scratch server that you started.

The socket uses newline-delimited JSON, with one request on each connection. Connect, write one line, such as `{"id":"1","method":"...","params":{...}}`, and read one line. `events.subscribe` and the `*.wait` methods keep their connection open. Thus each of them needs its own connection. herdr has no `herdr api call` subcommand. To send a raw request, use a small script, such as a script that uses the `socket` module of Python.

After the pass, stop each scratch server. Then make sure that no process of that server continues to run:

```sh
HERDR_SOCKET_PATH=~/.config/herdr/sessions/demo-probe/herdr.sock herdr server stop
pgrep -af demo-probe
ss -xlp | grep demo-probe
```

After a `server.live_handoff`, the new process runs as `herdr server --handoff-import …` without `--session`. Thus `pgrep -af "session <name>"` does not find it. Stop it through its socket. Then make sure that `ss -xlp | grep <name>` finds nothing.

### Writes outside the scratch session

Some methods and CLI commands write files outside the session directory of a scratch server. Probe them only with the isolation aids below. If you cannot isolate a write, do not probe it. Record it as "not probed", with the reason.

On herdr 0.9.5, a pass saw these writes from a scratch server:

- The first successful `release_notes.dismiss` wrote the shared file `~/.config/herdr/release-notes.json`. It wrote this file also from a named session. Refer to [api/ui.md](api/ui.md#release_notesdismiss).
- `herdr plugin config-dir <ID>` made the directory `~/.config/herdr/plugins/config/<ID>`.
- A scratch server got the remote agent manifests when it started. It wrote them to the agent-detection cache in `~/.local/state/herdr`. `server.reload_agent_manifests` writes the same cache.
- After a restart, herdr resumed a `claude` agent session of kind `id` with no `resume_argv`. It ran the command `claude --resume <id>`. The default value of the setting `resume_agents_on_restore` is `true`. A test agent session for a known agent kind can thus start the real agent program.

### Isolation aids

These aids let a pass probe more methods safely:

- **Config and state directories.** herdr 0.9.5 uses `XDG_CONFIG_HOME` and `XDG_STATE_HOME` for the session directory, `config.toml`, the agent-detection overrides, and the remote cache. Set each `XDG_*` variable to a private directory under your scratch directory. Then start the scratch server.
- **A config file.** Set `HERDR_CONFIG_PATH` to a copy of the config file before you start the scratch server.
- **`worktree.create`.** Use a config file that sets `[worktrees] directory` to a scratch path. Alternatively, start the server with a private `HOME`. Then herdr writes nothing to `~/.herdr/worktrees`. Use Git repositories that you make for the test, not a real project.
- **Untrusted repositories.** To probe a repository that herdr does not trust, start the server with `GIT_TEST_ASSUME_DIFFERENT_OWNER=1`.
- **`plugin.*` methods that write `plugins.json`.** Start a scratch server with `XDG_CONFIG_HOME`, `XDG_STATE_HOME`, `XDG_DATA_HOME`, and `XDG_CACHE_HOME` all in one private directory. Then the registry write stays in that directory.
- **`integration.install` and `integration.uninstall`.** Start a scratch server with `HOME`, each `XDG_*` variable, `CLAUDE_CONFIG_DIR`, and `CODEX_HOME` in one private directory. Before and after the probe, compare the checksums of the real agent config files, such as `~/.claude/settings.json` and `~/.codex/config.toml`.

### Agent programs

Some claims need a real agent, such as `claude`, `codex`, `opencode`, or `pi`. Start an agent program only in a pane of a scratch server, with a scratch working directory. Obey these rules:

- Record the PID of each agent and each background process that you start. Stop exactly those processes at the end of the pass.
- An agent program can write its own user files when it starts. For example, a trust dialog can write a trust entry. Give the agent a private home or config directory, or do not accept the dialog.
- If an agent program uses a shared background daemon, start it without that daemon. For example, start `codex` with `--no-daemon`.
- An agent program can use tokens of a model provider. Use short prompts.

## 5. Update the pages

Put each change on the page that owns it:

- A method that is new, removed, or changed goes to its namespace page in `api/`. The table in [README.md](README.md#method-reference-102-methods) gives the namespace pages.
- A change to the events goes to [events.md](events.md).
- A change to an envelope goes to [protocol.md](protocol.md). A change to a type in `$defs` goes to [data-model.md](data-model.md).
- A new CLI command goes to [cli-mapping.md](cli-mapping.md). Use `herdr <group> <cmd> --help` to find new commands.

Behavior changes from step 3 go to the same pages, also when the schema did not change.

### Owners of shared facts

Each shared fact has one owner page. Other pages link to the owner and do not repeat the fact.

| Fact | Owner |
|---|---|
| Meaning of each `AgentStatus` value | [data-model.md](data-model.md#agentstatus) |
| Socket path, environment variables, session directory, and the Windows setting `allow_unelevated_clients` | [environment.md](environment.md) |
| ID scheme | [addressing.md](addressing.md#id-scheme) |
| Error codes | [errors.md](errors.md) |
| Event catalog | [events.md](events.md#event-catalog-eventkind--eventdata) |
| CLI conventions | [cli-mapping.md](cli-mapping.md#cli-conventions) |
| Admission rule of `pane run`, and its line limit on macOS | [cli-mapping.md](cli-mapping.md#pane-run-admission) |
| Agent identity across Ctrl+Z, and the line limit of `agent start` | [api/agent.md](api/agent.md#agent-identity-across-ctrlz) |
| A CLI or API close does not use the TUI last-pane confirmation | [api/pane.md](api/pane.md#paneclose) |

## 6. Apply the stamps

Each claim and each example gets one of these three stamps:

| Stamp | Use it when |
|---|---|
| `Checked YYYY-MM-DD against herdr X.Y.Z` | This pass tested the claim under the rules of step 4. |
| `Derived from schema; not live-checked (YYYY-MM-DD, herdr X.Y.Z)` | The claim comes only from the shapes in `raw/schema.json`. |
| `Documented by herdr <source> (herdr X.Y.Z); not live-checked: <reason>` | The claim comes from `--skill`, `--default-config`, `agent-guide`, `--help`, or `<cmd> --help`, and this pass cannot test it. The `--help` sources are the help output of the installed herdr binary. Give the reason. |

Put the Derived and Documented stamps in inline code. Their semicolon then does not count as prose.

A section gets a Checked stamp only for the parts that this pass tested. Give a Derived or Documented stamp to the other parts. Do not keep a Checked stamp from an earlier version unless this pass tested the claim again. A note about an earlier version, such as "changed in 0.9.3", can stay.

If an example contains a version, a path, a host, or a label, capture it again from a scratch server. Do not only change its stamp. Make each example anonymous with the placeholders in [README.md](README.md#examples). Do not put environment values, tokens, or secrets into a page.

Then do these steps:

1. Replace the files in `raw/` with the new artifacts from step 2. Do not change their text.
2. Change the stamp line at the top of each page that this pass probed. Also change it in `README.md` and in this page. The form is `herdr X.Y.Z · protocol N · schema_version M · captured YYYY-MM-DD`.

## 7. Write in STE

Write all prose in ASD-STE100 Simplified Technical English, Strict level. The rules apply to table cells too. They do not apply to fenced code, inline code, or `raw/`. Obey these rules:

- Use the active voice.
- Put one instruction in each sentence.
- Use a maximum of 20 words in an instruction and 25 words in a description.
- Do not use semicolons or phrasal verbs.
- Use a maximum of three words in a noun cluster.
- Use simple tenses.
- Keep each hedge, such as "may", "can", or "sometimes". Keep each fact, condition, and number. Do not make a claim stronger.

Use each term only with its meaning in [STE-GLOSSARY.md](STE-GLOSSARY.md).

A linter can help with these rules, but you do not have to use one. For example, an ASD-STE100 linter such as `ste-lint.py` can use the base glossary and the glossary of this folder:

```sh
python3 -I ste-lint.py --glossary base-glossary.md --glossary STE-GLOSSARY.md <page>
```

A page passes when the linter reports 0 hard violations. Do not run the linter on `raw/` or on `STE-GLOSSARY.md`.

## 8. Do the checks

### Completeness check

Each method in the schema must have a `## <method>` heading on exactly one page:

```sh
jq -r '.schemas.request.oneOf[].properties.method.const' raw/schema.json | sort > /tmp/methods
grep -hE '^## [a-z_.]+$' *.md api/*.md | sed 's/^## //' | sort > /tmp/documented
diff /tmp/methods /tmp/documented   # must be empty
```

The diff shows a method with no heading, a method with two headings, and a heading that is not a method.

### Link check

Each relative link must go to a file that exists. Each `#anchor` must agree with a heading of the target page, under the GitHub heading rules. Use a link checker that examines files and anchors.
