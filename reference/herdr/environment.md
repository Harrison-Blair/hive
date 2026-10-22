# herdr API: environment and access model

> herdr 0.9.5 · protocol 22 · schema_version 1 · captured 2026-10-17
> Part of the herdr reference. Index: [README.md](README.md). Wire format: [protocol.md](protocol.md). IDs: [addressing.md](addressing.md).

herdr has no token, session key, or API key. Two facts give a client access to a server. First, the client can open the Unix socket of the server, which the operating system controls with file permissions. Second, herdr puts the identity of a managed pane into the environment of the processes in that pane. This page gives the environment variables that herdr sets, the socket path and its permissions, and the absence of credential checks. It also gives the layout of the session directory on the disk, and how named sessions keep servers apart.

Sources: `env` output in panes of scratch servers and of a live session, `raw/skill.md`, `raw/agent-guide.md`, `raw/default-config.toml`, and the top-level CLI help.

## Environment variables

herdr sets the context of the caller in each managed pane. Thus a process can find its own location and the server socket without configuration. This table gives the variables that a pass saw:

| variable | meaning | source |
| --- | --- | --- |
| `HERDR_ENV` | herdr sets it to `1` in a managed pane. It is the gate for all control. Before a client sends commands, it must do the check `test "${HERDR_ENV:-}" = 1`. If the variable has no value, the client must stop. | skill.md (top section), agent-guide.md |
| `HERDR_SOCKET_PATH` | The absolute path to the Unix domain socket of the server for this session. This socket is the transport endpoint (see [protocol.md](protocol.md)). | `env` in a pane |
| `HERDR_SESSION` | The name of the session that holds this pane, for example `demo`. herdr sets it only in panes of a named session. A pane of the default session does not have it. | `env` in a pane |
| `HERDR_WORKSPACE_ID` | The ID of the workspace that holds the calling pane, for example `w1`. Use it for `--workspace` and for target resolution. | skill.md, section "Use IDs and caller context" |
| `HERDR_TAB_ID` | The ID of the tab that holds the calling pane, for example `w1:t1`. | skill.md |
| `HERDR_PANE_ID` | The ID of the calling pane, for example `w1:p1`. For pane commands, it has the same effect as `--current`. | skill.md |
| `HERDR_BIN_PATH` | The absolute path to the `herdr` binary of this server. Scripts use it to start the correct client. `herdr status client` shows the same path on its `binary:` line. | `env` in a pane |
| `HERDR_CONFIG_PATH` | It sets a different path for the config file. `herdr --help` gives this text: "HERDR_CONFIG_PATH overrides config file path". herdr does not set this variable in panes. | CLI help |
| `SSH_AUTH_SOCK` | herdr sets it to `<session dir>/herdr.sock.agent` only when the server got an `SSH_AUTH_SOCK` value at its start. Thus SSH in the pane uses the agent socket that herdr selects. For the symlink, see [Socket path and permissions](#socket-path-and-permissions). | `env` in a pane |

This capture shows the variables in pane `w1:p1` of a named session. The server of this session got no `SSH_AUTH_SOCK` value, so the pane has no `SSH_AUTH_SOCK`.

```text
$ env | grep -E '^(HERDR_|SSH_AUTH_SOCK)' | sort
HERDR_BIN_PATH=/home/user/.local/bin/herdr
HERDR_ENV=1
HERDR_PANE_ID=w1:p1
HERDR_SESSION=demo
HERDR_SOCKET_PATH=/home/user/.config/herdr/sessions/demo/herdr.sock
HERDR_TAB_ID=w1:t1
HERDR_WORKSPACE_ID=w1
```

A pane of the default session has the same variables, but it has no `HERDR_SESSION`.

Usage notes:

- The three ID variables are the caller context that herdr sets. Prefer `--current` or `--pane "$HERDR_PANE_ID"` to an omitted target. A command without a target can act on the focused pane of a different client. See [addressing.md](addressing.md).
- A process keeps its `HERDR_PANE_ID`, `HERDR_TAB_ID`, and `HERDR_WORKSPACE_ID` values after a `pane.move`. The move gives the pane new IDs for external callers. But in the process, all three variables continue to show the location before the move. A child process that this process starts gets the old values too.
- A process gets these values from the process that started it, not from the pane that it serves. This is a problem for any client tool that runs commands through one shared background daemon. That daemon keeps the `HERDR_*` values of the pane that started it first. Thus each later command through the daemon on the machine shows that pane, also after a close of that pane. Do not trust the caller context in a command that a shared daemon runs.
- Codex is one example of such a client tool. By default, Codex runs tool commands through a shared app-server daemon. With `codex --no-daemon`, the Codex process of the pane runs the tool commands, with the values of that pane.
- `HERDR_ENV` is a gate flag. The other variables are data. Read them. If a variable has no value, do not guess its value.
- The process that makes a pane can change only `HERDR_SESSION` and `SSH_AUTH_SOCK` in the new pane. The `env` parameter of `pane.split`, `tab.create`, and `workspace.create` can set these two variables, and the new pane keeps the values. herdr writes its real value over an `env` value for `HERDR_PANE_ID`, `HERDR_ENV`, `HERDR_WORKSPACE_ID`, `HERDR_TAB_ID`, `HERDR_SOCKET_PATH`, or `HERDR_BIN_PATH`. herdr shows no warning for this change.
- Thus, do not trust the `$HERDR_SESSION` value of a child process for security decisions. The CLI uses `HERDR_SESSION` for the log path in the `herdr --help` footer. The CLI uses `HERDR_SOCKET_PATH` for the connection. With a false `HERDR_SESSION`, `herdr --help` shows a log path for the false session. In the same pane, `herdr status` shows the real socket.

This capture shows a pane from `pane.split` with a false value for each of the eight variables in its `env` parameter. `tab.create` and `workspace.create` gave the same result.

```text
HERDR_BIN_PATH=/home/user/.local/bin/herdr
HERDR_ENV=1
HERDR_PANE_ID=w1:p2
HERDR_SESSION=fake-HERDR_SESSION
HERDR_SOCKET_PATH=/home/user/.config/herdr/sessions/demo/herdr.sock
HERDR_TAB_ID=w1:t1
HERDR_WORKSPACE_ID=w1
SSH_AUTH_SOCK=fake-SSH_AUTH_SOCK
```

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the table, the two captures, the moved-process note, the gate note, the `env` parameter note, and the `--help` footer note. The pass read `env` in new panes of scratch servers, and in its own pane of a live session.
- `Documented by herdr --skill (herdr 0.9.5); not live-checked:` the `--current` note needs an attached TUI client with a different focus.
- The shared-daemon note and the Codex note come from an observation on 2026-10-12 with Codex. Not live-checked on herdr 0.9.5: this pass started Codex only with `--no-daemon`. The installed `codex --help` lists the `--no-daemon` flag and the shared app-server daemon.

## Socket path and permissions

- The server listens on a Unix domain socket of the stream type at `$HERDR_SOCKET_PATH`. `herdr status server` shows the same path on its `socket:` line. This capture comes from a pane of the `demo` session:

  ```text
  $ herdr status server
  status: running
  version: 0.9.5
  endpoint_compatible: yes
  private_protocol: 22
  private_protocol_compatible: yes
  socket: /home/user/.config/herdr/sessions/demo/herdr.sock
  ```

- herdr makes the socket file with the permission mode `0600`, which gives read and write access only to the owner. Thus only the OS user that owns the file can connect. herdr opens no network listener. The socket is local to the host. For remote control, herdr uses SSH. You can use `herdr --remote <target>`. You can also save a profile with `herdr machine add` and connect with `herdr --machine <label-or-id>`. These ways do not open the socket to the network.
- The session directory also holds a client socket `herdr-client.sock`, with the mode `0600`. The server log calls it the "client protocol socket". It does not use the NDJSON and JSON-RPC format of `herdr.sock`. A correct `{"id":…,"method":"ping","params":{}}` line to this socket gets no pong and no error response. The peer resets the connection. The server log then shows an `oversized handshake from client` warning. That warning tells that the server read the first four bytes of the line as a length. Thus this socket probably expects a binary handshake with a length prefix. Use this socket only as an internal part of herdr. Do not use it as a second or fallback API endpoint.
- The session directory can also hold `herdr.sock.agent`, a symlink. The panes use this symlink as their `SSH_AUTH_SOCK`. Its presence and its target depend on the `SSH_AUTH_SOCK` value that the server got at its start:

  | `SSH_AUTH_SOCK` of the server at start | `herdr.sock.agent` | `SSH_AUTH_SOCK` in panes |
  | --- | --- | --- |
  | not set | herdr makes no symlink. | not set |
  | a path that does not exist | It points to `herdr.sock.unavailable` in the same directory, which does not exist. | `<session dir>/herdr.sock.agent` |
  | a path to an agent socket that exists | It points to that agent socket. | `<session dir>/herdr.sock.agent` |

- The method `server.ssh_agent.register` can change the target of this symlink. If the first agent socket of the server does not exist, the symlink points to the registered socket. This change stays only while the connection of the registration stays open. After the close of that connection, the symlink points to `herdr.sock.unavailable` again. If the first agent socket of the server exists, the registration does not change the symlink. For the method, see [api/server.md](api/server.md#serverssh_agentregister).
- A socket path can be too long for the `sun_path` field of `sockaddr_un` in the OS. If `$HERDR_SOCKET_PATH` points to such a path, `herdr status server` does not show a normal "not running" status. It fails with exit code 1 and this error: `Custom { kind: InvalidInput, error: "local socket name length exceeds capacity of sun_path of sockaddr_un" }`. herdr does not check the length of `<name>` in the socket paths of named sessions (`~/.config/herdr/sessions/<name>/herdr.sock`). Thus a long name can cause this error.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the stream socket, the `0600` mode of both sockets, the `--remote` and `--machine` flags in `herdr --help`, and the reset on `herdr-client.sock`. Also checked: the absence of a network listener, with `ss -tlnp` and `ss -ulnp` on scratch servers. Also checked: the three `herdr.sock.agent` cases on three scratch servers, and the `server.ssh_agent.register` effect.
- The `sun_path` note comes from an observation on 2026-10-06 with herdr 0.9.3, with a directory of 120 characters in `HERDR_SOCKET_PATH`. Not live-checked on herdr 0.9.5: the tools of this pass do not accept a socket path longer than 107 bytes.

## No token or API-key authentication

herdr does no authorization with credentials. These items do not exist:

- A bearer token, an API key, or a shared secret in the request. The [request envelope](protocol.md#request-envelope) has only `id`, `method`, and `params`. The server ignores other keys, such as `"token"` or `"auth"`, in the envelope or in `params`.
- A login step, a session token, or a handshake that makes a credential. The `ping` handshake sends the version, the protocol, and the capabilities. It does not do authorization (see [protocol.md](protocol.md)). A client can send other methods without a `ping` first.
- A signature or an authorization header for each request.

Only two things control access:

1. OS socket permissions. The client must run as the OS user that owns the socket, and it must open the socket path with the mode `0600`.
2. Context that herdr sets. herdr trusts the workspace, tab, and pane identity that it put in the environment of a managed pane.

Any process that can read `$HERDR_SOCKET_PATH` and open that socket has full control of the server. A process with an empty environment and no `HERDR_*` variables can also control the server through the socket path. Access to the session directory in the file system gives full control of the session. Use the isolation of OS users for access control. Do not use herdr for access control.

This capture shows a `ping` with an extra `"token"` key and an extra `"auth"` key. The server ignores both keys:

```text
→ {"id":"1","method":"ping","params":{},"token":"x","auth":"y"}
← {"id":"1","result":{"type":"pong","version":"0.9.5","protocol":22,"capabilities":{"live_handoff":true,"detached_server_daemon":false,"endpoint_protocol_generation":1,"surface_interest":true,"health_check":true,"ssh_agent_registration":true}}}
```

A saved remote machine adds no separate credential. You add it with `herdr machine add`, and then you use `herdr --machine <label-or-id> <command>`. herdr sends the API command to the socket on the remote host through the SSH connection. Thus, on that host, access also comes from OS controls. These controls are the socket permissions on that host and the SSH authentication for the host. The source is the skill.md section "To control a saved SSH machine". `herdr machine --help` lists the subcommands `list`, `status`, `reconnect`, `add`, `rename`, `remove`, `enable`, and `disable`. `--remote-keybindings <local|server>` selects only the keybindings for a TUI attach through `--remote`. The default is `local`. This flag has no effect on access.

On Windows, the `[server]` table of the config file has the `allow_unelevated_clients` setting. The default config file gives this text:

```toml
# Windows only: allow ordinary same-account clients to control an elevated server.
# Requires a server restart.
# allow_unelevated_clients = false
```

The default value is `false`. Thus, with the default value, a client that is not elevated probably cannot control an elevated server, also when both run under the same account. When you set the value to `true`, such clients can control an elevated server. A change to this setting has effect only after a server restart.

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the request envelope, the extra keys in the envelope and in `params`, and the `ping` capture. Also checked: the methods without a `ping` first, and the access with an empty environment through `env -i`. Also checked: the `herdr machine --help` subcommand list.
- `Documented by herdr --skill (herdr 0.9.5); not live-checked:` the `--machine` forwarding needs a real remote SSH host.
- `Documented by herdr --help (herdr 0.9.5); not live-checked:` `--remote-keybindings` needs an SSH target and a real TTY.
- `Documented by herdr --default-config (herdr 0.9.5); not live-checked:` `allow_unelevated_clients` is a Windows setting, and the host of the pass runs Linux.

## Session directory layout

Each named session has a directory in the herdr config root. On Linux and macOS, the config root is `~/.config/herdr/`. On Windows, it is `%APPDATA%\herdr\` (agent-guide.md, troubleshooting section). The files of the default session are directly in the config root. A named session has its own subdirectory:

```text
~/.config/herdr/
├── config.toml                     # user configuration (HERDR_CONFIG_PATH sets another path)
├── session.json                    # saved workspace/tab/pane/layout snapshot (undocumented; loaded again on restart)
└── sessions/
    └── <name>/                     # one directory for each named session, for example demo/
        ├── herdr.sock              # server API socket ($HERDR_SOCKET_PATH), mode 0600
        ├── herdr-client.sock       # client protocol socket, mode 0600
        ├── herdr-client.log        # client log; probably only after a TUI client attaches
        ├── herdr-server.log        # server log, mode 0644
        ├── herdr.sock.agent        # symlink for the SSH_AUTH_SOCK of the panes; only when the server started with SSH_AUTH_SOCK set
        ├── session.json            # saved snapshot of this session, mode 0644; absent until the session changes state
        └── session-snapshots/      # recovery copies of session.json; directory mode 0755, each file mode 0600
```

In the directory of a named session after some activity, the pass saw `herdr-client.sock`, `herdr-server.log`, `herdr.sock`, `session-snapshots/`, and `session.json`. The pass saw `herdr.sock.agent` only for servers that started with `SSH_AUTH_SOCK` set. The pass did not see `herdr-client.log` in a named session directory. The pass attached no TUI client there, and CLI commands from a pane did not make the file. The config root of a live session with attached clients has a `herdr-client.log`. Thus herdr probably writes this file when a TUI client attaches.

The config root also holds a `session-snapshots/` directory next to its `herdr.sock`. It can also hold `herdr.sock.agent` under the same condition as a named session. The config root of the live session in this pass also held `.plugins.lock` and `release-notes.json`.

A recovery copy has the name `session-<timestamp>-<server pid>-<n>.json`, with a timestamp that has leading zeros. It uses the `session.json` format, version 3. The scratch server wrote one recovery copy, and two restarts did not add a second copy. The log line for a recovery copy is `preserved session recovery copy`.

The `herdr --help` footer names a combined log. This capture comes from a pane of the `demo` session:

```text
Config: /home/user/.config/herdr/config.toml
Logs:   /home/user/.config/herdr/sessions/demo/herdr.log (plus herdr-client.log, herdr-server.log)
```

Outside a pane, or in a pane with no `HERDR_SESSION`, the footer shows `/home/user/.config/herdr/herdr.log`. The footer path comes from `$HERDR_SESSION`, not from the socket. herdr does not write the `herdr.log` file. The file does not occur in the default config root or in a named session directory. It also does not occur after client activity, server activity, and restarts. Only `herdr-client.log` and `herdr-server.log` occur. Thus, do not expect a file at the `herdr.log` path in the footer. The agent-guide.md sentence "Named-session logs live under `sessions/<name>/`" is correct for the two logs that occur.

Other herdr documents do not describe `session.json`. A new session has no `session.json`. After the first change of state, the config root and each named session directory hold one. It keeps a full snapshot of the workspaces, tabs, panes, and layout. This state stays after a server restart. You can stop the server of a named session and start a new server under the same name. Then the new server loads the same workspaces, tabs, and panes. The pane IDs and the custom labels stay the same. A restored pane can also start again a resume command that an agent reported. If the saved directory of a pane does not exist, herdr restores the pane as a placeholder with `restore_error`. See [data-model.md](data-model.md#appendix-on-disk-persistence-sessionjson-version-3).

`herdr status` and `herdr status server` give a summary of the runtime and show the resolved socket path on a `socket:` line. `herdr status` shows a `client:` block, a `server:` block, and an `update:` block. `herdr status client` does not show a socket path. `herdr status --json` gives the same data as JSON.

```text
$ herdr status client
version: 0.9.5
channel: stable
protocol: 22
endpoint_protocol_generation: 1
binary: /home/user/.local/bin/herdr
```

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: the Linux config root, the tree and the file modes, the recovery copy, and the `--help` footer. Also checked: the absence of `herdr.log` and the `restore_error` placeholder. Also checked: the `session.json` restore after two restarts of a scratch server, and the `herdr status` output.
- `Documented by herdr agent-guide (herdr 0.9.5); not live-checked:` the Windows config root needs a Windows host.
- The resume command note belongs to [data-model.md](data-model.md#appendix-on-disk-persistence-sessionjson-version-3). This pass did not check it on this page.

## Named sessions as isolation

A session is the unit of isolation. It has one server process, one socket, one namespace for workspaces, tabs, and panes, one set of logs, and one directory.

- **Select a session.** `herdr --session <name>` uses a named persistent session, or makes it if it does not exist. `herdr session attach <name>` attaches to it. `herdr session list [--json]`, `herdr session stop <name>`, and `herdr session delete <name>` manage sessions. In a pane of a named session, `$HERDR_SESSION` gives the name of the session. A pane of the default session has no `HERDR_SESSION`.
- **Isolation boundary.** Each session has its own socket path and its own set of IDs. Thus `w1` in one session has no relation to `w1` in a different session. A client of one socket cannot see or change the topology of a different session. Access to the server needs only access to the socket. Thus, if you can limit a user to one directory, each session is also a separate access domain.
- **Use for tests.** skill.md has the rule "Never kill the main Herdr process". It also has the rule "Use named test sessions for experiments that need an isolated server". Send each probe that changes state or removes items to a scratch server under a named session. Then an error cannot change the panes and processes of the live session. `herdr server stop` stops the running server through the API socket, and `herdr --help` lists it. Use it only on purpose. Do not use it on a live session that other people or agents use.
- **No `--no-session` mode.** The CLI has no `--no-session` mode, and it has no "monolithic" mode without a server and client. `herdr --no-session status` fails with `unknown option: --no-session` and exit code 2. The usage text and the options text of `herdr --help` do not contain `--no-session` or "monolithic".

This capture shows `herdr session list --json` with one named session. The row of the default session has the name `default`, and its `session_dir` is the config root.

```text
{"sessions":[{"default":false,"name":"demo","running":true,"session_dir":"/home/user/.config/herdr/sessions/demo","socket_path":"/home/user/.config/herdr/sessions/demo/herdr.sock"}, ...]}
```

`herdr session delete` does not delete the directory of a running session. Stop the session first:

```text
$ herdr session delete demo
{"error":{"code":"session_delete_failed","message":"session demo is running; stop it before deleting"}}
$ herdr session stop demo
stopped session demo
$ herdr session delete demo
deleted session demo
```

The first command fails with exit code 1. After the last command, the session directory does not exist. For error codes, see [errors.md](errors.md).

Stamps for this section:

- `Checked 2026-10-17 against herdr 0.9.5`: three scratch servers that ran at the same time, each with its own directory, socket, server log, and process. Also checked: the `herdr session --help` subcommands, `session list --json`, `session stop`, `session delete`, `herdr server stop`, and the `--no-session` error. Two sessions had a `w1` with different labels at the same time.
- `Documented by herdr --skill (herdr 0.9.5); not live-checked:` the skill.md rule about named test sessions is text, not behavior.
- `herdr session attach` occurs in `herdr session --help`. The pass did not do a TUI attach, because it needs a real terminal.
