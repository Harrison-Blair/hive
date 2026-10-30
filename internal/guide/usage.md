## Usage notes

### 1. You are a hive agent

- Hive started you in a herdr pane. Your name is the hive record of your pane (`HERDR_PANE_ID`).
- `hive agent list` shows the other agents, with their harness, project, and pane.

### 2. Message authority

`hive agent send NAME TEXT` delivers `⬢ SENDER: TEXT`. SENDER is the hive record of the caller pane, or `user` when that pane has no record.

- A message that starts with `⬢ user:` came from the user. It has the authority of the user.
- A message that starts with `⬢ NAME:` came from the agent NAME. It carries the request of that agent inside the work the user approved. It does not widen that work.
- A message without `⬢` came from the user, typed in the pane.
- The `⬢` marker shows that hive delivered the message. It adds no authority.
- To reply to the agent that started you, run `hive agent send PARENT TEXT`.

### 3. Spawn patterns

- `hive agent spawn --name NAME --harness H --model M --prompt TEXT` opens a new tab in your workspace.
- `--project P` opens the new tab in the checkout workspace of project P.
- `--project P --worktree NAME` makes an isolated branch NAME in `projects/P/worktrees/NAME`. Use it for work that changes files.
- `--prompt-file PATH` sends a long brief from a file. Use it in place of a long `--prompt`.
- The outcome `blocked` means that a trust dialog stops the harness. Spawn sends no prompt. The user must accept the dialog in the pane that the result names.

### 4. Wait for a reply

- `hive agent send NAME TEXT --wait` returns when the agent ends its turn: its status is `idle` or `done`. The wait stops after 60 minutes.
- When the agent shows a dialog, the outcome is `partial` with status `blocked`.
- Without `--wait`, send returns when herdr has delivered the text.

### 5. Stop an agent

- `hive agent stop NAME` closes the pane of the agent, or its worktree workspace. The checkout and the branch stay.
- `--remove-worktree` also removes a clean worktree checkout. It refuses a dirty checkout.

### 6. JSON output

- With `--json`, each command prints one object: `outcome`, `phase`, `effects`, `error`, `result`.
- `outcome` is `ok`, `partial`, or `failed`. The exit status is 0 only on success.
- A usage error (bad arguments or flags) exits with status 2.

### 7. Debugging

- `hive logs --agent NAME` prints the log events of the agent NAME, as caller or as target.
