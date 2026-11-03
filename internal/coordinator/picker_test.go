package coordinator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/lib/style"
	"github.com/charmbracelet/x/ansi"
)

var piModels = []string{"opencode-go/deepseek-v4.1-flash", "opencode-go/deepseek-v4-pro", "openai-codex/gpt-6-sol"}

func press(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func typeText(t *testing.T, m Picker, text string) Picker {
	t.Helper()
	for _, r := range text {
		m = update(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func update(t *testing.T, m Picker, msg tea.Msg) Picker {
	t.Helper()
	next, _ := m.Update(msg)
	return next.(Picker)
}

// enterAndDiscover presses Enter on step 1 and feeds the discovery result back.
func enterAndDiscover(t *testing.T, m Picker) Picker {
	t.Helper()
	next, cmd := m.Update(press(tea.KeyEnter))
	m = next.(Picker)
	for _, msg := range runCmd(cmd) {
		if _, ok := msg.(modelsMsg); ok {
			m = update(t, m, msg)
		}
	}
	return m
}

// runCmd runs cmd and each command of a batch, and returns the messages.
func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var msgs []tea.Msg
		for _, c := range batch {
			msgs = append(msgs, runCmd(c)...)
		}
		return msgs
	}
	return []tea.Msg{msg}
}

// view is the screen text without SGR codes. raw keeps them.
func view(m Picker) string { return ansi.Strip(raw(m)) }

func raw(m Picker) string { return m.View().Content }

func lineWith(t *testing.T, screen, part string) string {
	t.Helper()
	for line := range strings.SplitSeq(screen, "\n") {
		if strings.Contains(line, part) {
			return line
		}
	}
	t.Fatalf("screen has no line with %q:\n%s", part, screen)
	return ""
}

func fakeDiscover(models []string, err error) func(string) ([]string, error) {
	return func(string) ([]string, error) { return models, err }
}

func newPiPicker() Picker {
	return NewPicker(Choice{Harness: "pi", Model: "opencode-go/deepseek-v4-pro"}, true, fakeDiscover(piModels, nil))
}

func TestPickerRule1PreselectsLastChoiceAndEnterEnterStartsIt(t *testing.T) {
	m := newPiPicker()
	line := lineWith(t, view(m), " pi")
	if !strings.Contains(line, "▸ pi") || !strings.Contains(line, "✓ last") {
		t.Fatalf("step 1 pi line = %q; want cursor and ✓ last", line)
	}
	if strings.Contains(lineWith(t, view(m), "claude"), "▸") {
		t.Fatal("cursor is on claude; want pi")
	}
	m = enterAndDiscover(t, m)
	line = lineWith(t, view(m), "deepseek-v4-pro")
	if !strings.Contains(line, "▸ opencode-go/deepseek-v4-pro") || !strings.Contains(line, "✓ last") {
		t.Fatalf("step 2 last model line = %q; want cursor and ✓ last", line)
	}
	next, cmd := m.Update(press(tea.KeyEnter))
	choice, ok := next.(Picker).Result()
	if !ok || choice.Harness != "pi" || choice.Model != "opencode-go/deepseek-v4-pro" {
		t.Fatalf("Result = %+v %v; want the last choice", choice, ok)
	}
	if cmd == nil || cmd() != tea.Quit() {
		t.Fatal("Enter on a model does not quit the program")
	}
}

func TestPickerEnterPutsTheCursorOnTheRememberedModelOfThatHarness(t *testing.T) {
	last := Choice{Harness: "claude", Model: "claude-opus-5-5", Models: map[string]string{"claude": "claude-opus-5-5", "pi": "opencode-go/deepseek-v4-pro"}}
	m := NewPicker(last, true, fakeDiscover(piModels, nil))
	m = update(t, m, press(tea.KeyDown))
	m = update(t, m, press(tea.KeyDown))
	m = enterAndDiscover(t, m)
	if line := lineWith(t, view(m), "deepseek-v4-pro"); !strings.Contains(line, "▸ opencode-go/deepseek-v4-pro") {
		t.Fatalf("pi model line = %q; want the cursor on the remembered pi model", line)
	}
}

func TestPickerRule1NoLastStartsOnFirstRows(t *testing.T) {
	m := NewPicker(Choice{}, false, fakeDiscover(piModels, nil))
	if !strings.Contains(view(m), "▸ claude") || strings.Contains(view(m), "✓ last") {
		t.Fatalf("view = %q; want cursor on claude and no last mark", view(m))
	}
	m = update(t, m, press(tea.KeyDown))
	m = update(t, m, press(tea.KeyDown))
	m = enterAndDiscover(t, m)
	if !strings.Contains(view(m), "▸ opencode-go/deepseek-v4.1-flash") {
		t.Fatalf("view = %q; want cursor on the first model", view(m))
	}
}

func TestPickerRule2FilterCountAndOtherRow(t *testing.T) {
	m := enterAndDiscover(t, newPiPicker())
	if !strings.Contains(view(m), "3 of 3") {
		t.Fatalf("view = %q; want 3 of 3", view(m))
	}
	m = typeText(t, m, "DEEPS")
	screen := view(m)
	if !strings.Contains(lineWith(t, screen, "▏"), "DEEPS▏") || !strings.Contains(screen, "2 of 3") || strings.Contains(screen, "gpt-6-sol") {
		t.Fatalf("filter DEEPS view = %q; want 2 of 3 deepseek rows", screen)
	}
	if !strings.Contains(screen, "Other…") {
		t.Fatalf("view = %q; want Other… row", screen)
	}
	m = update(t, m, press(tea.KeyDown))
	m = update(t, m, press(tea.KeyDown))
	if !strings.Contains(view(m), "▸ Other…") {
		t.Fatalf("view = %q; want cursor on Other… as the last row", view(m))
	}
	m = update(t, m, press(tea.KeyUp))
	if !strings.Contains(view(m), "▸ opencode-go/deepseek-v4-pro") {
		t.Fatalf("view = %q; want cursor back on the second row", view(m))
	}
	m = typeText(t, m, "zzz")
	screen = view(m)
	if !strings.Contains(screen, "0 of 3") || !strings.Contains(screen, "▸ Other…") {
		t.Fatalf("view = %q; want 0 of 3 and the Other… row under the cursor", screen)
	}
	for range 3 {
		m = update(t, m, press(tea.KeyBackspace))
	}
	if !strings.Contains(view(m), "2 of 3") {
		t.Fatalf("view after backspace = %q; want 2 of 3", view(m))
	}
}

func TestPickerRule3EscClearsFilterThenGoesBackThenQuits(t *testing.T) {
	m := enterAndDiscover(t, newPiPicker())
	m = typeText(t, m, "gpt")
	m = update(t, m, press(tea.KeyEscape))
	if !strings.Contains(view(m), "hive · pi · model") || !strings.Contains(view(m), "3 of 3") {
		t.Fatalf("view = %q; want step 2 with the filter cleared", view(m))
	}
	m = update(t, m, press(tea.KeyEscape))
	if !strings.Contains(view(m), "hive · harness") {
		t.Fatalf("view = %q; want step 1", view(m))
	}
	next, cmd := m.Update(press(tea.KeyEscape))
	if _, ok := next.(Picker).Result(); ok {
		t.Fatal("Result ok after quit; want none")
	}
	if cmd == nil || cmd() != tea.Quit() {
		t.Fatal("Esc on step 1 does not quit")
	}
}

func TestPickerRule3CtrlCQuitsFromEachState(t *testing.T) {
	ctrlC := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	loading, _ := newPiPicker().Update(press(tea.KeyEnter))
	other := enterAndDiscover(t, newPiPicker())
	for range 3 {
		other = update(t, other, press(tea.KeyDown))
	}
	other = update(t, other, press(tea.KeyEnter))
	states := map[string]Picker{
		"harness": newPiPicker(),
		"loading": loading.(Picker),
		"models":  enterAndDiscover(t, newPiPicker()),
		"failure": enterAndDiscover(t, NewPicker(Choice{}, false, fakeDiscover(nil, errors.New("boom")))),
		"other":   other,
	}
	for name, m := range states {
		next, cmd := m.Update(ctrlC)
		if _, ok := next.(Picker).Result(); ok {
			t.Errorf("%s: Result ok after ctrl+c", name)
		}
		if cmd == nil || cmd() != tea.Quit() {
			t.Errorf("%s: ctrl+c does not quit", name)
		}
	}
}

func TestPickerRule4LoadingRunsDiscoveryWithSpinner(t *testing.T) {
	var asked []string
	m := NewPicker(Choice{}, false, func(h string) ([]string, error) {
		asked = append(asked, h)
		return []string{"gpt-6-sol"}, nil
	})
	m = update(t, m, press(tea.KeyDown))
	next, cmd := m.Update(press(tea.KeyEnter))
	m = next.(Picker)
	if !strings.Contains(view(m), "⠋ listing models…") {
		t.Fatalf("loading view = %q; want spinner line", view(m))
	}
	if len(asked) != 0 {
		t.Fatal("discovery ran inside Update; want it in the command")
	}
	for _, msg := range runCmd(cmd) {
		m = update(t, m, msg)
	}
	if len(asked) != 1 || asked[0] != "codex" {
		t.Fatalf("discovery calls = %v; want [codex]", asked)
	}
	if !strings.Contains(view(m), "▸ gpt-6-sol") {
		t.Fatalf("view = %q; want the model list", view(m))
	}
}

func TestPickerLoadingSpinnerAdvances(t *testing.T) {
	next, _ := newPiPicker().Update(press(tea.KeyEnter))
	m := update(t, next.(Picker), spinnerMsg{})
	if !strings.Contains(view(m), "⠙ listing models…") {
		t.Fatalf("view = %q; want the second spinner frame", view(m))
	}
}

func TestPickerLoadingEscGoesBackAndDropsTheLateList(t *testing.T) {
	next, cmd := newPiPicker().Update(press(tea.KeyEnter))
	m := update(t, next.(Picker), press(tea.KeyEscape))
	if !strings.Contains(view(m), "hive · harness") {
		t.Fatalf("view = %q; want step 1", view(m))
	}
	for _, msg := range runCmd(cmd) {
		m = update(t, m, msg)
	}
	if !strings.Contains(view(m), "hive · harness") {
		t.Fatalf("view = %q; want step 1 after a late list", view(m))
	}
}

func TestPickerRule4FailureShowsErrorAndTypedEntry(t *testing.T) {
	for name, discover := range map[string]func(string) ([]string, error){
		"error": fakeDiscover(nil, errors.New("opencode models: not found")),
		"empty": fakeDiscover(nil, nil),
	} {
		t.Run(name, func(t *testing.T) {
			m := enterAndDiscover(t, NewPicker(Choice{}, false, discover))
			line := lineWith(t, view(m), "!")
			if !strings.Contains(line, "│ ! model list unavailable ") {
				t.Fatalf("failure line = %q", line)
			}
			if !strings.Contains(view(m), "│ model ▏") {
				t.Fatalf("view = %q; want the typed-entry field", view(m))
			}
			if lines := strings.Split(view(m), "\n"); name == "error" && !strings.HasPrefix(lines[3], "│   opencode models: not found ") {
				t.Fatalf("view = %q; want the error on the row under the failure line", view(m))
			}
			m = typeText(t, m, "claude-opus-5-5")
			if !strings.Contains(view(m), "│ model claude-opus-5-5▏") {
				t.Fatalf("view = %q; want the typed name", view(m))
			}
			m = update(t, m, press(tea.KeyEnter))
			if choice, ok := m.Result(); !ok || choice.Harness != "claude" || choice.Model != "claude-opus-5-5" {
				t.Fatalf("Result = %+v %v", choice, ok)
			}
		})
	}
}

func TestPickerFailureFitsAnEightyColumnPane(t *testing.T) {
	long := errors.New(`opencode models: exec: "opencode": executable file not found in $PATH`)
	m := enterAndDiscover(t, NewPicker(Choice{}, false, fakeDiscover(nil, long)))
	m = typeText(t, m, "claude-opus-5-5")
	for line := range strings.SplitSeq(strings.TrimRight(view(m), "\n"), "\n") {
		if n := len([]rune(line)); n > 80 {
			t.Fatalf("line %q has width %d; want at most 80", line, n)
		}
	}
}

func TestPickerFailureEscGoesBackAndEmptyEnterWaits(t *testing.T) {
	m := enterAndDiscover(t, NewPicker(Choice{}, false, fakeDiscover(nil, errors.New("boom"))))
	m = update(t, m, press(tea.KeyEnter))
	if _, ok := m.Result(); ok {
		t.Fatal("empty Enter gave a result")
	}
	m = update(t, m, press(tea.KeyEscape))
	if !strings.Contains(view(m), "hive · harness") {
		t.Fatalf("view = %q; want step 1", view(m))
	}
}

func TestPickerOtherRowTakesAnyTypedName(t *testing.T) {
	m := enterAndDiscover(t, newPiPicker())
	for range 3 {
		m = update(t, m, press(tea.KeyDown))
	}
	m = update(t, m, press(tea.KeyEnter))
	screen := view(m)
	if !strings.Contains(screen, "│ model ▏") || !strings.Contains(screen, "[model] enter start · esc back") {
		t.Fatalf("Other… view = %q", screen)
	}
	m = typeText(t, m, "glm-5.3:xhigx")
	m = update(t, m, press(tea.KeyBackspace))
	m = typeText(t, m, "h")
	m = update(t, m, press(tea.KeyEnter))
	if choice, ok := m.Result(); !ok || choice.Harness != "pi" || choice.Model != "glm-5.3:xhigh" {
		t.Fatalf("Result = %+v %v; want the typed name", choice, ok)
	}
}

func TestPickerOtherEscReturnsToList(t *testing.T) {
	m := enterAndDiscover(t, newPiPicker())
	m = update(t, m, press(tea.KeyUp)) // from the last model, up stays on rows
	for range 3 {
		m = update(t, m, press(tea.KeyDown))
	}
	m = update(t, m, press(tea.KeyEnter))
	m = update(t, m, press(tea.KeyEscape))
	if !strings.Contains(view(m), "3 of 3") || !strings.Contains(view(m), "▸ Other…") {
		t.Fatalf("view = %q; want the list with the cursor on Other…", view(m))
	}
}

func TestPickerRule5GlyphsCarryStateWithNoColour(t *testing.T) {
	m := newPiPicker()
	screens := []string{view(m)}
	next, _ := m.Update(press(tea.KeyEnter))
	screens = append(screens, view(next.(Picker)), view(enterAndDiscover(t, m)))
	screens = append(screens, view(enterAndDiscover(t, NewPicker(Choice{}, false, fakeDiscover(nil, errors.New("boom"))))))
	screens = append(screens, view(NewPicker(Choice{Harness: "claude", Model: "claude-opus-5-5"}, true, fakeDiscover(piModels, nil))))
	for i, glyph := range []string{"▸", "⠋", "▸", "!"} {
		if !strings.Contains(screens[i], glyph) {
			t.Errorf("screen %d lacks %q:\n%s", i, glyph, screens[i])
		}
	}
	if !strings.Contains(screens[0], "✓ last") || !strings.Contains(screens[2], "✓ last") {
		t.Error("a screen lacks ✓ last")
	}
	if !strings.Contains(lineWith(t, screens[4], "codex"), "–") {
		t.Errorf("codex row without a model lacks –:\n%s", screens[4])
	}
}

func TestPickerRule6TitleShowsStep(t *testing.T) {
	m := newPiPicker()
	if top := strings.SplitN(view(m), "\n", 2)[0]; !strings.HasPrefix(top, "╭─ hive · harness ") || !strings.HasSuffix(top, "╮") {
		t.Fatalf("step 1 top = %q", top)
	}
	if !strings.Contains(view(m), "\n [harness] ↑↓ move · enter next · esc quit") {
		t.Fatalf("step 1 view = %q; want the step 1 keys", view(m))
	}
	m = enterAndDiscover(t, m)
	if top := strings.SplitN(view(m), "\n", 2)[0]; !strings.HasPrefix(top, "╭─ hive · pi · model ") {
		t.Fatalf("step 2 top = %q", top)
	}
	if !strings.Contains(view(m), "\n [model] ↑↓ move · type filter · enter start · esc back") {
		t.Fatalf("step 2 view = %q; want the step 2 keys", view(m))
	}
}

func TestPickerBoxLinesHaveOneWidth(t *testing.T) {
	m := enterAndDiscover(t, newPiPicker())
	lines := strings.Split(view(m), "\n")
	width := len([]rune(lines[0]))
	for _, line := range lines[:len(lines)-1] {
		if len([]rune(line)) != width {
			t.Fatalf("line %q has width %d; want %d", line, len([]rune(line)), width)
		}
	}
}

func TestPickerLogsTheFullDiscoveryErrorAndShowsItClipped(t *testing.T) {
	long := errors.New(`opencode models: exec: "opencode": executable file not found in $PATH`)
	var out bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&out, nil))
	m := enterAndDiscover(t, NewPicker(Choice{}, false, logFailures(fakeDiscover(nil, long), logger)))
	if !strings.Contains(view(m), `opencode models: exec: "opencode": executable file not …`) || strings.Contains(view(m), long.Error()) {
		t.Fatalf("view does not show the clipped error:\n%s", view(m))
	}
	var event map[string]any
	if err := json.Unmarshal(out.Bytes(), &event); err != nil {
		t.Fatalf("log is not one JSON event: %v\n%s", err, out.String())
	}
	if event["error"] != long.Error() || event["harness"] != "claude" || event["level"] != "ERROR" {
		t.Fatalf("log event = %v; want the full error and harness claude at ERROR", event)
	}
}

// manyModels is a list longer than the box.
func manyModels() []string {
	var models []string
	for i := range 20 {
		models = append(models, fmt.Sprintf("provider/model-%02d", i))
	}
	return models
}

// steps returns the picker in each step, after msgs.
func steps(t *testing.T, msgs ...tea.Msg) map[string]Picker {
	t.Helper()
	last := Choice{Harness: "pi", Model: "provider/model-10", Models: map[string]string{"pi": "provider/model-10"}}
	fresh := func(discover func(string) ([]string, error)) Picker {
		m := NewPicker(last, true, discover)
		for _, msg := range msgs {
			m = update(t, m, msg)
		}
		return m
	}
	harness := fresh(fakeDiscover(manyModels(), nil))
	loading, _ := harness.Update(press(tea.KeyEnter))
	models := enterAndDiscover(t, harness)
	other := models
	for range 21 {
		other = update(t, other, press(tea.KeyDown))
	}
	other = update(t, other, press(tea.KeyEnter))
	if other.step != stepOther {
		t.Fatalf("other step not reached:\n%s", view(other))
	}
	return map[string]Picker{
		"harness": harness,
		"loading": loading.(Picker),
		"models":  models,
		"failure": enterAndDiscover(t, fresh(fakeDiscover(nil, errors.New("boom")))),
		"other":   other,
	}
}

// size gives the widest line and the line count of a view.
func size(screen string) (int, int) {
	lines := strings.Split(screen, "\n")
	width := 0
	for _, line := range lines {
		width = max(width, ansi.StringWidth(line))
	}
	return width, len(lines)
}

func TestPickerEveryStepHasTheSameSize(t *testing.T) {
	for _, msgs := range [][]tea.Msg{nil, {tea.WindowSizeMsg{Width: 80, Height: 24}}} {
		for name, m := range steps(t, msgs...) {
			w, h := size(raw(m))
			if w != 62 || h != 16 {
				t.Errorf("%s step after %v is %d × %d; want 62 × 16 (a 62 × 15 box and the key line):\n%s", name, msgs, w, h, view(m))
			}
			lines := strings.Split(raw(m), "\n")
			for _, line := range lines[:15] {
				if ansi.StringWidth(line) != 62 {
					t.Errorf("%s box line %q is %d cells; want 62", name, ansi.Strip(line), ansi.StringWidth(line))
				}
			}
		}
	}
}

func TestPickerShrinksToASmallPaneAndGrowsBack(t *testing.T) {
	small := tea.WindowSizeMsg{Width: 48, Height: 11}
	for name, m := range steps(t, small) {
		if w, h := size(raw(m)); w > 48 || h > 11 {
			t.Errorf("%s step in a 48 × 11 pane is %d × %d:\n%s", name, w, h, view(m))
		}
		m = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
		if w, h := size(raw(m)); w != 62 || h != 16 {
			t.Errorf("%s step after the pane grows is %d × %d; want 62 × 16", name, w, h)
		}
	}
	models := steps(t, small)["models"]
	if !strings.Contains(view(models), "▸ provider/model-10") {
		t.Errorf("small models step hides the cursor row:\n%s", view(models))
	}
	if strings.Count(view(models), "provider/model-") >= strings.Count(view(steps(t)["models"]), "provider/model-") {
		t.Errorf("small models step does not lose list rows:\n%s", view(models))
	}
}

func TestPickerHarnessRowsShowTheLastModelOfEachHarness(t *testing.T) {
	last := Choice{Harness: "claude", Model: "claude-opus-5-5", Models: map[string]string{"claude": "claude-opus-5-5", "codex": "gpt-6.1-sol"}}
	m := NewPicker(last, true, fakeDiscover(piModels, nil))
	screen := view(m)
	if line := lineWith(t, screen, "claude "); !strings.Contains(line, "▸ claude") || !strings.Contains(line, "✓ last") || !strings.Contains(line, "claude-opus-5-5") {
		t.Errorf("claude row = %q; want cursor, ✓ last, and claude-opus-5-5", line)
	}
	if line := lineWith(t, screen, "codex"); !strings.Contains(line, "gpt-6.1-sol") || strings.Contains(line, "✓ last") {
		t.Errorf("codex row = %q; want gpt-6.1-sol and no ✓ last", line)
	}
	if line := lineWith(t, raw(m), " pi"); !strings.Contains(line, "\x1b["+style.Dim+"m–\x1b[0m") {
		t.Errorf("pi row = %q; want a dim –", line)
	}
}

func TestPickerColoursFollowTheGraphPalette(t *testing.T) {
	last := Choice{Harness: "codex", Model: "gpt-6-sol", Models: map[string]string{"codex": "gpt-6-sol"}}
	m := NewPicker(last, true, fakeDiscover([]string{"gpt-6-sol", "gpt-6-luna"}, nil))
	screen := raw(m)
	top := strings.SplitN(screen, "\n", 2)[0]
	if !strings.Contains(top, "\x1b["+style.Accent+"mhive\x1b[0m") || !strings.Contains(top, "\x1b["+style.Dim+"m · \x1b[0m") {
		t.Errorf("harness title = %q; want hive in the accent and a dim separator", top)
	}
	for _, want := range []string{
		"\x1b[" + style.Selected + "m▸\x1b[0m",
		"\x1b[1;" + style.Codex + "mcodex\x1b[0m",
		"\x1b[" + style.Claude + "mclaude\x1b[0m",
		"\x1b[" + style.Dim + "m✓ last\x1b[0m",
		"\x1b[" + style.Border + "m╭",
		"\n \x1b[" + style.Dim + "m[harness] ",
	} {
		if !strings.Contains(screen, want) {
			t.Errorf("harness step lacks %q:\n%q", want, screen)
		}
	}
	next, _ := m.Update(press(tea.KeyEnter))
	loading := raw(next.(Picker))
	if !strings.Contains(loading, "\x1b["+style.Working+"m⠋\x1b[0m") || !strings.Contains(loading, "\x1b["+style.Dim+"mlisting models…") {
		t.Errorf("loading step = %q; want a green spinner and dim text", loading)
	}
	if top := strings.SplitN(loading, "\n", 2)[0]; !strings.Contains(top, "\x1b["+style.Codex+"mcodex\x1b[0m") {
		t.Errorf("model title = %q; want codex in its harness colour", top)
	}
	models := raw(enterAndDiscover(t, m))
	for _, want := range []string{
		"\x1b[" + style.Selected + "mgpt-6-sol\x1b[0m",
		"\x1b[" + style.Dim + "mOther…\x1b[0m",
		"\x1b[" + style.Dim + "mfilter\x1b[0m",
		"\x1b[" + style.Dim + "m2 of 2\x1b[0m",
		"\x1b[" + style.Border + "m│" + strings.Repeat("─", 60) + "│\x1b[0m",
	} {
		if !strings.Contains(models, want) {
			t.Errorf("models step lacks %q:\n%q", want, models)
		}
	}
	failure := raw(enterAndDiscover(t, NewPicker(Choice{}, false, fakeDiscover(nil, errors.New("boom")))))
	if !strings.Contains(failure, "\x1b["+style.Error+"m! model list unavailable\x1b[0m") || !strings.Contains(failure, "\x1b["+style.Dim+"mboom\x1b[0m") {
		t.Errorf("failure step = %q; want the red failure line and the dim cause", failure)
	}
}

func TestPickerClipsALongModelName(t *testing.T) {
	long := "provider/" + strings.Repeat("x", 80)
	m := enterAndDiscover(t, NewPicker(Choice{}, false, fakeDiscover([]string{long, "short"}, nil)))
	line := lineWith(t, view(m), "provider/")
	if !strings.Contains(line, "x…") || ansi.StringWidth(line) != 62 {
		t.Fatalf("long model row = %q (%d cells); want it clipped with … in the 62-cell box", line, ansi.StringWidth(line))
	}
}

func TestPickerKeyLines(t *testing.T) {
	want := map[string]string{
		"harness": " [harness] ↑↓ move · enter next · esc quit",
		"loading": " [model] esc back",
		"models":  " [model] ↑↓ move · type filter · enter start · esc back",
		"failure": " [model] enter start · esc back",
		"other":   " [model] enter start · esc back",
	}
	for name, m := range steps(t) {
		lines := strings.Split(view(m), "\n")
		if got := lines[len(lines)-1]; got != want[name] {
			t.Errorf("%s key line = %q; want %q", name, got, want[name])
		}
	}
}

func TestPickerNeverGoesBeyondATinyPane(t *testing.T) {
	for _, pane := range []tea.WindowSizeMsg{{Width: 20, Height: 8}, {Width: 10, Height: 4}, {Width: 3, Height: 2}, {Width: 1, Height: 1}} {
		for name, m := range steps(t, pane) {
			lines := strings.Split(raw(m), "\n")
			if len(lines) > pane.Height {
				t.Errorf("%s step in a %d × %d pane has %d lines:\n%s", name, pane.Width, pane.Height, len(lines), view(m))
			}
			for _, line := range lines {
				if w := ansi.StringWidth(line); w > pane.Width {
					t.Errorf("%s step in a %d × %d pane has a line of %d cells: %q", name, pane.Width, pane.Height, w, ansi.Strip(line))
				}
			}
		}
	}
}

func TestPickerKeepsTheWholeBoxInASmallPane(t *testing.T) {
	for name, m := range steps(t, tea.WindowSizeMsg{Width: 20, Height: 8}) {
		lines := strings.Split(view(m), "\n")
		box := lines[:len(lines)-1]
		if !strings.HasPrefix(box[0], "╭") || !strings.HasPrefix(box[len(box)-1], "╰") {
			t.Errorf("%s box has no top or bottom edge:\n%s", name, view(m))
		}
		for _, line := range box {
			if w := ansi.StringWidth(line); w != 20 || !strings.ContainsAny(string([]rune(line)[19]), "╮│╯") {
				t.Errorf("%s box line %q (%d cells) lacks the right edge at column 20", name, line, w)
			}
		}
	}
}
