# Task 137 repair: F1 and F2 of VERIFIED.md at c4badeaa

The earlier evidence, each `V<n>.md`, and `VERIFIED.md` did not change. No board
copy and no grid changed. `git status --short .hive/spec/golden-frames/design/`
prints nothing.

## F1: a resize breaks the ask border (decision 5)

- Cause: `ComposerCard.top` removed the right text when space was short, then
  truncated the full row. The truncation removed the `╮` corner with the title.
- Fix: `internal/tui/component/composer.go`. When the title does not fit, the
  title is truncated to `width-3` cells with `…`. The `╭─` and `╮` corners stay.
  A row that fits does not change, so no golden grid changes.
- Test: `TestAskResizeKeepsTheTopRightCorner` in
  `internal/tui/ask_resize_test.go`, from the verifier probe. It also checks that
  the top row is exactly the pane width.
- Rows (`ask-resize-rows.txt`, plain text without colour):

```text
120x40: ╭─ ? manager-project-with-a-long-name asks · Question0000, Question0001, Question0002, Question0003 ──── open · 1 of 1 ╮
80x24:  ╭─ ? manager-project-with-a-long-name asks · Question0000, Question0001, Quest…╮
```

## F2: the config refuses a valid Bubble Tea key name (decision 18)

- Cause: `validKeyName` accepted only one rune after the modifier prefixes.
- Fix: `internal/lib/home/config.go`. `validKeyName` also accepts each Bubble
  Tea named key. `namedKeys` takes the names from Bubble Tea itself
  (`tea.Key{Code: c}.String()` for the control codes, backspace, and each
  extended key code), so the list follows Bubble Tea. The prefixes `meta+`,
  `hyper+`, and `super+` of Bubble Tea are also accepted. Bubble Tea is
  already a module dependency; `go mod tidy -diff` prints nothing.
- Test: `TestParseConfigTUIAcceptsNamedKeys` in
  `internal/lib/home/config_tui_test.go`. Hand-written keys: `alt+up`,
  `ctrl+pgdown`, `f5`, `shift+enter`, `ctrl+space`, `alt+home`, `meta+x`,
  `super+left` are accepted. `alt+upp`, `ctrl+f99`, `foo`, `alt+` are refused
  with an error that names `tui.keys.help`.

## Runs

| File | Content | Result |
| --- | --- | --- |
| `red.txt` | The two new tests on `c4badeaa` code | fail: F2 names `"f5" is not a key name` (and 6 others); F1 names the top row that ends in `Questio` with no `╮` |
| `green.txt` | The two new tests after the fix, and the four verifier probes from `verifier/probes_test.go.txt` by Go overlay | pass, all six |
| `gate.txt` | The gate of `AGENTS.md` | `gofmt -l` and `go mod tidy -diff` print nothing; vet, lint (`0 issues.`), and `go test -race ./...` exit 0 |

The red run is the deliberate break of each fix: it is the code without the fix.
