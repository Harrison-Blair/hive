package coordinator

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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

func view(m Picker) string { return m.View().Content }

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
	if !ok || choice != (Choice{"pi", "opencode-go/deepseek-v4-pro"}) {
		t.Fatalf("Result = %+v %v; want the last choice", choice, ok)
	}
	if cmd == nil || cmd() != tea.Quit() {
		t.Fatal("Enter on a model does not quit the program")
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
			if !strings.Contains(line, "! model list unavailable:") {
				t.Fatalf("failure line = %q", line)
			}
			if !strings.Contains(view(m), "│ type a model name: ▏") {
				t.Fatalf("view = %q; want the typed-entry field", view(m))
			}
			if name == "error" && !strings.Contains(line, "opencode models: not found") {
				t.Fatalf("failure line = %q; want the error", line)
			}
			m = typeText(t, m, "claude-opus-5-5")
			if !strings.Contains(view(m), "type a model name: claude-opus-5-5▏") {
				t.Fatalf("view = %q; want the typed name", view(m))
			}
			m = update(t, m, press(tea.KeyEnter))
			if choice, ok := m.Result(); !ok || choice != (Choice{"claude", "claude-opus-5-5"}) {
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
	if !strings.Contains(screen, "model: ▏") || !strings.Contains(screen, "enter start · esc back") {
		t.Fatalf("Other… view = %q", screen)
	}
	m = typeText(t, m, "glm-5.3:xhigx")
	m = update(t, m, press(tea.KeyBackspace))
	m = typeText(t, m, "h")
	m = update(t, m, press(tea.KeyEnter))
	if choice, ok := m.Result(); !ok || choice != (Choice{"pi", "glm-5.3:xhigh"}) {
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
	for _, screen := range screens {
		if strings.Contains(screen, "\x1b") {
			t.Fatalf("view has an escape code; state must read without colour: %q", screen)
		}
	}
	for i, glyph := range []string{"▸", "⠋", "▸", "!"} {
		if !strings.Contains(screens[i], glyph) {
			t.Errorf("screen %d lacks %q:\n%s", i, glyph, screens[i])
		}
	}
	if !strings.Contains(screens[0], "✓ last") || !strings.Contains(screens[2], "✓ last") {
		t.Error("a screen lacks ✓ last")
	}
}

func TestPickerRule6TitleShowsStep(t *testing.T) {
	m := newPiPicker()
	if top := strings.SplitN(view(m), "\n", 2)[0]; !strings.HasPrefix(top, "┌ hive · harness ") || !strings.HasSuffix(top, "┐") {
		t.Fatalf("step 1 top = %q", top)
	}
	if !strings.Contains(view(m), "↑↓ move · enter next · esc quit") {
		t.Fatalf("step 1 view = %q; want the step 1 keys", view(m))
	}
	m = enterAndDiscover(t, m)
	if top := strings.SplitN(view(m), "\n", 2)[0]; !strings.HasPrefix(top, "┌ hive · pi · model ") {
		t.Fatalf("step 2 top = %q", top)
	}
	if !strings.Contains(view(m), "↑↓ move · type filter · enter start · esc back") {
		t.Fatalf("step 2 view = %q; want the step 2 keys", view(m))
	}
}

func TestPickerBoxLinesHaveOneWidth(t *testing.T) {
	m := enterAndDiscover(t, newPiPicker())
	lines := strings.Split(strings.TrimRight(view(m), "\n"), "\n")
	width := len([]rune(lines[0]))
	for _, line := range lines {
		if len([]rune(line)) != width {
			t.Fatalf("line %q has width %d; want %d", line, len([]rune(line)), width)
		}
	}
}
