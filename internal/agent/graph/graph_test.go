package graph

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/list"
	"github.com/Harrison-Blair/hive/internal/lib/style"
	"github.com/charmbracelet/x/ansi"
)

var clock = time.Date(2026, 10, 30, 14, 22, 5, 0, time.UTC)

func entry(name, parent, status, harness, model string, age time.Duration) list.Entry {
	return list.Entry{
		Record: agent.Record{Name: name, Parent: parent, Harness: harness, Model: model, Pane: "w1:" + name, StartedAt: clock.Add(-age)},
		Status: status,
	}
}

// fixture has two project pods and a home pod.
func fixture() []list.Entry {
	return []list.Entry{
		entry("api-coordinator", "coordinator", "idle", "codex", "gpt-6", 10*time.Minute),
		entry("api-worker", "api-coordinator", "working", "pi", "luna", 3*time.Minute),
		entry("coordinator", "", "working", "claude", "opus", time.Hour),
		entry("hive-coordinator", "coordinator", "working", "claude", "opus", 4*time.Minute+5*time.Second),
		entry("hive-reviewer", "hive-coordinator", "done", "pi", "sol", 2*time.Minute),
		entry("hive-worker", "hive-coordinator", "blocked", "codex", "gpt-6", 90*time.Second),
		entry("hive-worker-kid", "hive-worker", "idle", "pi", "luna", 30*time.Second),
		entry("loose", "", "gone", "pi", "luna", 5*time.Second),
	}
}

func fetchOf(entries []list.Entry) func(context.Context) ([]list.Entry, error) {
	return func(context.Context) ([]list.Entry, error) { return entries, nil }
}

// loaded returns a model of size width×height that has the entries.
func loaded(t *testing.T, entries []list.Entry, width, height int) Model {
	t.Helper()
	m := New(Options{Fetch: fetchOf(entries), Now: func() time.Time { return clock }})
	m = update(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	return update(t, m, fetchedMsg{entries: entries, at: clock})
}

func update(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	next, _ := m.Update(msg)
	return next.(Model)
}

func key(t *testing.T, m Model, code rune) Model {
	t.Helper()
	return update(t, m, tea.KeyPressMsg{Code: code})
}

func screen(m Model) string {
	return ansi.Strip(m.View().Content)
}

func TestPodsGroupProjectCoordinatorsAndHome(t *testing.T) {
	pods := Pods(fixture())
	var titles []string
	for _, pod := range pods {
		titles = append(titles, pod.Title)
	}
	if strings.Join(titles, ",") != "api,hive,home" {
		t.Fatalf("pod titles = %v; want api, hive, home", titles)
	}
	var hive []string
	for _, n := range pods[1].Order() {
		hive = append(hive, n.Entry.Name)
	}
	if strings.Join(hive, ",") != "hive-coordinator,hive-reviewer,hive-worker,hive-worker-kid" {
		t.Fatalf("hive pod order = %v", hive)
	}
	var home []string
	for _, n := range pods[2].Order() {
		home = append(home, n.Entry.Name)
	}
	if strings.Join(home, ",") != "coordinator,loose" {
		t.Fatalf("home pod order = %v", home)
	}
}

func TestPodsSurviveAParentCycle(t *testing.T) {
	pods := Pods([]list.Entry{entry("a", "b", "idle", "pi", "m", 0), entry("b", "a", "idle", "pi", "m", 0)})
	if len(pods) != 1 || len(pods[0].Order()) != 2 {
		t.Fatalf("pods = %+v; want one home pod with both agents", pods)
	}
}

func TestGraphCardsShortenModelLabelsAndDetailsKeepFullValue(t *testing.T) {
	for _, tc := range []struct{ model, label string }{
		{"openai-codex/gpt-6-sol", "6-sol"},
		{"claude-opus-5-5", "opus-5-5"},
		{"opencode-go/deepseek-v4-pro", "deepseek-v4-pro"},
		{"anthropic/claude-opus-5-5", "opus-5-5"},
		{"gpt-6", "6"},
		{"deepseek-v4-pro", "deepseek-v4-pro"},
		{"", ""},
	} {
		t.Run(tc.model, func(t *testing.T) {
			e := entry("solo", "", "idle", "pi", tc.model, time.Minute)
			m := loaded(t, []list.Entry{e}, 100, 25)
			for _, mode := range []string{"tree", "list"} {
				text := screen(m)
				want := "pi"
				if tc.label != "" {
					want += " · " + tc.label
				}
				if !strings.Contains(text, want+" ") {
					t.Errorf("%s card missing %q:\n%s", mode, want, text)
				}
				m = key(t, m, 'g')
			}
			m = key(t, m, tea.KeyEnter)
			if tc.model != "" && !strings.Contains(screen(m), tc.model) {
				t.Errorf("details missing full model %q:\n%s", tc.model, screen(m))
			}
			if e.Model != tc.model || m.entries[0].Model != tc.model {
				t.Errorf("model record changed: entry %q, graph %q; want %q", e.Model, m.entries[0].Model, tc.model)
			}
		})
	}
}

func TestCardRow3ColoursTheHarnessNameAndKeepsTheModelDim(t *testing.T) {
	for _, tc := range []struct{ harness, model, want string }{
		{"claude", "opus", "\x1b[" + style.Claude + "mclaude\x1b[0m\x1b[" + style.Dim + "m · opus"},
		{"codex", "gpt-6.1-sol", "\x1b[" + style.Codex + "mcodex\x1b[0m\x1b[" + style.Dim + "m · 6.1-sol"},
		{"pi", "luna", " pi\x1b[" + style.Dim + "m · luna"},
	} {
		e := entry("solo", "", "idle", tc.harness, tc.model, time.Minute)
		row3 := cardLines(e, "1m00s", false, 0, 28)[3]
		if !strings.Contains(row3, tc.want) {
			t.Errorf("%s row 3 = %q; want %q", tc.harness, row3, tc.want)
		}
	}
}

func TestCardShowsNameCheckoutThenHarnessAndModel(t *testing.T) {
	e := entry("hive-coordinator", "", "idle", "claude", "opus", 4*time.Minute+5*time.Second)
	m := loaded(t, []list.Entry{e}, 80, 20)
	m = update(t, m, fetchedMsg{entries: []list.Entry{e}, checkouts: map[string]string{e.Name: "feature-branch"}})
	text := screen(m)
	if !strings.Contains(text, "feature-branch") || strings.Index(text, "feature-branch") > strings.Index(text, "claude · opus") {
		t.Fatalf("checkout must appear before harness and model:\n%s", text)
	}
	lines := cardLines(e, "4m05s", false, 0, 28)
	if len(lines) != 5 {
		t.Fatalf("card has %d lines; want 5", len(lines))
	}
	row1, row3 := ansi.Strip(lines[1]), ansi.Strip(lines[3])
	if !strings.HasPrefix(row1, "│ ○ hive-coordinator") || !strings.HasSuffix(row1, " 4m05s │") {
		t.Errorf("row 1 = %q; want the idle glyph, the name, and 4m05s", row1)
	}
	if row3 != "│ claude · opus"+strings.Repeat(" ", 12)+"│" {
		t.Errorf("row 3 = %q; want harness · model and no status", row3)
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w != 28 {
			t.Errorf("line %q is %d cells; want 28", ansi.Strip(line), w)
		}
	}
}

func TestCardKeepsACoordinatorNameWithALongElapsedTime(t *testing.T) {
	for age, want := range map[time.Duration]string{10 * time.Minute: "10m00s", 26*time.Hour + 4*time.Minute + 3*time.Second: "26h04m03s"} {
		e := entry("hive-coordinator", "", "idle", "claude", "opus", time.Hour)
		m := loaded(t, []list.Entry{e}, 140, 20)
		e.Status = "working"
		m = update(t, m, fetchedMsg{entries: []list.Entry{e}, at: clock})
		m = update(t, m, clockMsg(clock.Add(age)))
		if row := cardRow1(t, m, "hive-coordinator"); !strings.Contains(row, " "+want+" ") {
			t.Errorf("age %v: card row 1 %q has no full name and time in state", age, row)
		}
	}
	e := entry("hive-coordinator", "", "idle", "claude", "opus", 26*time.Hour+4*time.Minute+3*time.Second)
	if text := screen(loaded(t, []list.Entry{e}, 140, 20)); !strings.Contains(text, " hive-coordinator ≥26h04m03s ") {
		t.Errorf("a long time with ≥ does not fit whole beside the full name:\n%s", text)
	}
}

func TestElapsedHasSeconds(t *testing.T) {
	for age, want := range map[time.Duration]string{
		7 * time.Second:                              "7s",
		4*time.Minute + 5*time.Second:                "4m05s",
		26*time.Hour + 4*time.Minute + 3*time.Second: "26h04m03s",
		-time.Second:                                 "0s",
	} {
		if got := elapsed(age); got != want {
			t.Errorf("elapsed(%v) = %q; want %q", age, got, want)
		}
	}
}

func TestViewDrawsPodsTreesAndClock(t *testing.T) {
	text := screen(listed(t, fixture(), 140, 50))
	for _, want := range []string{
		"14:22:05", "╭─ api ", "╭─ hive ", "╭─ home ",
		"│ │╭─", "├│ ✓ hive-reviewer", "╰│ × hive-worker ", "╰│ ○ hive-worker-kid",
		"┃ ○ api-coordinator", "⠋ working", "○ idle", "× blocked", "✓ done", "– gone",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("view has no %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "· idle") || strings.Contains(text, "· working") {
		t.Errorf("a card shows a status word:\n%s", text)
	}
}

func TestSelectedCardHasHeavyBorder(t *testing.T) {
	m := listed(t, fixture(), 140, 50)
	text := screen(m)
	if strings.Count(text, "┏") != 1 || !strings.Contains(text, "┃ ○ api-coordinator") {
		t.Fatalf("first card is not the one heavy card:\n%s", text)
	}
	m = key(t, m, tea.KeyDown)
	text = screen(m)
	if strings.Count(text, "┏") != 1 || !strings.Contains(text, "╰┃ ⠋ api-worker") {
		t.Fatalf("down did not move the heavy border:\n%s", text)
	}
}

func TestNavigation(t *testing.T) {
	m := listed(t, fixture(), 140, 50)
	if m.selected != "api-coordinator" {
		t.Fatalf("selected = %q; want the first card", m.selected)
	}
	m = key(t, m, tea.KeyDown)
	if m.selected != "api-worker" {
		t.Fatalf("down: selected = %q", m.selected)
	}
	m = key(t, m, tea.KeyRight)
	if m.selected != "hive-reviewer" {
		t.Fatalf("right: selected = %q; want the card at the same row of the next pod", m.selected)
	}
	m = key(t, m, tea.KeyRight)
	m = key(t, m, tea.KeyRight)
	if m.selected != "loose" {
		t.Fatalf("right at the last pod: selected = %q; want loose", m.selected)
	}
	m = key(t, m, tea.KeyLeft)
	m = key(t, m, tea.KeyUp)
	if m.selected != "hive-coordinator" {
		t.Fatalf("left then up: selected = %q", m.selected)
	}
	m = key(t, m, 'j')
	m = key(t, m, 'k')
	if m.selected != "hive-coordinator" {
		t.Fatalf("j then k: selected = %q", m.selected)
	}
	m = key(t, m, tea.KeyUp)
	if m.selected != "api-worker" {
		t.Fatalf("up at the top of a pod: selected = %q; want the last card of the pod before", m.selected)
	}
}

func TestEnterShowsDetailsAndEscCloses(t *testing.T) {
	entries := fixture()
	entries[0].Cwd, entries[0].Worktree, entries[0].HarnessSessionID = "/src/api", "api-wt", "sess-1"
	m := key(t, loaded(t, entries, 120, 40), tea.KeyEnter)
	text := screen(m)
	for _, want := range []string{"api-coordinator", "status", "idle", "codex", "gpt-6", "w1:api-coordinator", "coordinator", "/src/api", "api-wt", "sess-1", "10m00s", "esc close"} {
		if !strings.Contains(text, want) {
			t.Errorf("details have no %q:\n%s", want, text)
		}
	}
	m = key(t, m, tea.KeyEscape)
	if strings.Contains(screen(m), "esc close") {
		t.Fatalf("esc did not close the details:\n%s", screen(m))
	}
}

func TestQuitKeys(t *testing.T) {
	for _, msg := range []tea.KeyPressMsg{{Code: 'q'}, {Code: 'c', Mod: tea.ModCtrl}} {
		_, cmd := loaded(t, fixture(), 120, 40).Update(msg)
		if cmd == nil {
			t.Fatalf("%v gives no command", msg)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%v does not quit", msg)
		}
	}
}

func TestNarrowStacksPodsInTheWidth(t *testing.T) {
	m := listed(t, fixture(), 36, 120)
	content := m.View().Content
	for _, line := range strings.Split(content, "\n") {
		if w := ansi.StringWidth(line); w > 36 {
			t.Errorf("line %q is %d cells; want at most 36", ansi.Strip(line), w)
		}
	}
	text := ansi.Strip(content)
	api, hive := strings.Index(text, "╭─ api "), strings.Index(text, "╭─ hive ")
	if api < 0 || hive < 0 || strings.Count(text[:hive], "\n") <= strings.Count(text[:api], "\n") {
		t.Fatalf("pods are not stacked:\n%s", text)
	}
	if !strings.Contains(text, "hive-worker") {
		t.Fatalf("narrow cards drop the names:\n%s", text)
	}
}

func TestViewScrollsToTheSelection(t *testing.T) {
	m := listed(t, fixture(), 36, 20)
	for range 7 {
		m = key(t, m, tea.KeyDown)
	}
	if m.selected != "loose" {
		t.Fatalf("selected = %q; want loose", m.selected)
	}
	text := screen(m)
	if !strings.Contains(text, "loose") {
		t.Fatalf("the selected card is off screen:\n%s", text)
	}
	if lines := strings.Count(m.View().Content, "\n") + 1; lines > 20 {
		t.Fatalf("view has %d lines; want at most 20", lines)
	}
}

func TestRefreshErrorKeepsTheLastSnapshot(t *testing.T) {
	m := loaded(t, fixture(), 140, 50)
	m = update(t, m, fetchedMsg{err: errors.New("herdr: boom")})
	text := screen(m)
	if !strings.Contains(text, "herdr: boom") || !strings.Contains(text, "hive-worker-kid") {
		t.Fatalf("view after a failed refresh:\n%s", text)
	}
	m = update(t, m, fetchedMsg{entries: fixture()[:1]})
	if text := screen(m); strings.Contains(text, "boom") || strings.Contains(text, "hive-worker") {
		t.Fatalf("a good refresh does not replace the error and the snapshot:\n%s", text)
	}
}

func TestFirstFetchErrorShowsNoCards(t *testing.T) {
	m := New(Options{Fetch: fetchOf(nil), Now: func() time.Time { return clock }})
	m = update(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = update(t, m, fetchedMsg{err: errors.New("records: bad")})
	if text := screen(m); !strings.Contains(text, "records: bad") {
		t.Fatalf("view = %s", text)
	}
}

func TestNoAgents(t *testing.T) {
	if text := screen(loaded(t, nil, 80, 24)); !strings.Contains(text, "no agents") {
		t.Fatalf("view = %s", text)
	}
}

func TestSelectionFollowsTheNameAcrossRefreshes(t *testing.T) {
	m := key(t, loaded(t, fixture(), 140, 50), tea.KeyDown)
	m = update(t, m, fetchedMsg{entries: fixture()[1:]})
	if m.selected != "api-worker" {
		t.Fatalf("selected = %q; want api-worker kept", m.selected)
	}
	m = update(t, m, fetchedMsg{entries: fixture()[2:]})
	if m.selected != "hive-coordinator" {
		t.Fatalf("selected = %q; want the first card when the selection is gone", m.selected)
	}
}

func TestClockTickRefreshesAndR(t *testing.T) {
	calls := 0
	m := New(Options{Fetch: func(context.Context) ([]list.Entry, error) { calls++; return fixture(), nil }, Now: func() time.Time { return clock }})
	m = update(t, m, tea.WindowSizeMsg{Width: 140, Height: 50})
	m = update(t, m, fetchedMsg{entries: fixture()})
	next, cmd := m.Update(clockMsg(clock.Add(time.Second)))
	m = next.(Model)
	runAll(t, &m, cmd)
	if calls != 1 {
		t.Fatalf("a clock tick fetched %d times; want 1", calls)
	}
	if !strings.Contains(screen(m), "14:22:06") {
		t.Fatalf("clock did not advance:\n%s", screen(m))
	}
	next, cmd = m.Update(tea.KeyPressMsg{Code: 'r'})
	m = next.(Model)
	runAll(t, &m, cmd)
	if calls != 2 {
		t.Fatalf("r fetched %d times in all; want 2", calls)
	}
}

func TestOneFetchAtATime(t *testing.T) {
	m := loaded(t, fixture(), 140, 50)
	_, first := m.Update(tea.KeyPressMsg{Code: 'r'})
	next, _ := m.Update(tea.KeyPressMsg{Code: 'r'})
	_, second := next.(Model).Update(tea.KeyPressMsg{Code: 'r'})
	if first == nil || second != nil {
		t.Fatalf("first r = %v, r while fetching = %v; want a fetch and then none", first != nil, second != nil)
	}
}

// runAll runs cmd and feeds each fetchedMsg back to the model. It skips the
// timer commands, which would block.
func runAll(t *testing.T, m *Model, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		switch msg := msg.(type) {
		case tea.BatchMsg:
			for _, c := range msg {
				runAll(t, m, c)
			}
		case fetchedMsg:
			*m = update(t, *m, msg)
		}
	case <-time.After(200 * time.Millisecond):
	}
}

func TestSpinnerRunsOnlyWhileAWorkingCardIsVisible(t *testing.T) {
	m := loaded(t, fixture(), 140, 50)
	if !m.spinning {
		t.Fatal("no spinner with working cards on screen")
	}
	before := screen(m)
	next, cmd := m.Update(frameMsg{gen: m.frameGen})
	m = next.(Model)
	if cmd == nil {
		t.Fatal("a frame with working cards schedules no next frame")
	}
	after := screen(m)
	if !strings.Contains(before, "⠋ api-worker") || !strings.Contains(after, "⠙ api-worker") {
		t.Fatalf("frame 1 does not show the second braille frame:\nbefore\n%s\nafter\n%s", before, after)
	}
	if !strings.Contains(after, "○ api-coordinator") {
		t.Fatalf("an idle card spins:\n%s", after)
	}

	idle := loaded(t, []list.Entry{entry("a", "", "idle", "pi", "m", 0)}, 140, 50)
	if idle.spinning {
		t.Fatal("spinner on with no working card")
	}
	m = update(t, m, fetchedMsg{entries: []list.Entry{entry("a", "", "idle", "pi", "m", 0)}})
	if _, cmd := m.Update(frameMsg{gen: m.frameGen}); cmd != nil {
		t.Fatal("a frame with no working card schedules a next frame")
	}
}

func TestSpinnerStopsWhenWorkingCardsScrollOff(t *testing.T) {
	entries := []list.Entry{entry("a-working", "", "working", "pi", "m", 0)}
	for _, name := range []string{"b", "c", "d", "e", "f", "g"} {
		entries = append(entries, entry(name, "", "idle", "pi", "m", 0))
	}
	m := listed(t, entries, 40, 16)
	for range 6 {
		m = key(t, m, tea.KeyDown)
	}
	if strings.Contains(screen(m), "a-working") {
		t.Fatalf("working card still on screen:\n%s", screen(m))
	}
	if _, cmd := m.Update(frameMsg{gen: m.frameGen}); cmd != nil {
		t.Fatal("spinner runs with the working card off screen")
	}
	m = key(t, m, tea.KeyEnter)
	m = key(t, m, tea.KeyEscape)
	for range 6 {
		m = key(t, m, tea.KeyUp)
	}
	if !m.spinning {
		t.Fatal("spinner does not restart when the working card comes back")
	}
}

// TestStaleFrameDoesNotSpinTwice proves that a frame of a stopped spinner
// does not start a second frame loop beside the new one.
func TestStaleFrameDoesNotSpinTwice(t *testing.T) {
	m := loaded(t, fixture(), 140, 50)
	stale := frameMsg{gen: m.frameGen}
	m = key(t, m, tea.KeyEnter)
	m = key(t, m, tea.KeyEscape)
	if !m.spinning || stale.gen == m.frameGen {
		t.Fatalf("spinning = %v, gen = %d; want a new spinner", m.spinning, m.frameGen)
	}
	next, cmd := m.Update(stale)
	if cmd != nil || next.(Model).frame != m.frame {
		t.Fatal("a stale frame advanced the spinner or scheduled a frame")
	}
}

// cardRow1 returns the stripped row 1 of the card of name: the line that
// holds the name.
func cardRow1(t *testing.T, m Model, name string) string {
	t.Helper()
	for _, line := range strings.Split(screen(m), "\n") {
		if strings.Contains(line, " "+name+" ") {
			return line
		}
	}
	t.Fatalf("no card row of %s:\n%s", name, screen(m))
	return ""
}

// rawCardRow1 is cardRow1 with its SGR sequences.
func rawCardRow1(t *testing.T, m Model, name string) string {
	t.Helper()
	for _, line := range strings.Split(m.View().Content, "\n") {
		if strings.Contains(ansi.Strip(line), " "+name+" ") {
			return line
		}
	}
	t.Fatalf("no card row of %s:\n%s", name, screen(m))
	return ""
}

func legendLine(m Model) string {
	lines := strings.Split(m.View().Content, "\n")
	return lines[len(lines)-2]
}

func TestEachStateHasItsGlyphAndStyle(t *testing.T) {
	for _, tc := range []struct{ status, glyph, style string }{
		{"working", "⠋", "38;2;110;230;154"},
		{"idle", "○", "38;2;245;185;66"},
		{"blocked", "×", "38;2;232;112;95"},
		{"done", "✓", "38;2;163;157;144"},
		{"unknown", "·", "38;2;163;157;144"},
		{"gone", "–", "38;2;163;157;144"},
		{"strange", "·", "38;2;163;157;144"},
	} {
		e := entry("solo", "", tc.status, "pi", "m", time.Minute)
		m := loaded(t, []list.Entry{entry("first", "", "idle", "pi", "m", 0), e}, 140, 30)
		row := cardRow1(t, m, "solo")
		if !strings.Contains(row, tc.glyph+" solo ") {
			t.Errorf("%s: card row %q has no %q", tc.status, row, tc.glyph)
		}
		if raw := rawCardRow1(t, m, "solo"); !strings.Contains(raw, "\x1b["+tc.style+"m"+tc.glyph+"\x1b[0m") {
			t.Errorf("%s: card glyph %q is not in style %q: %q", tc.status, tc.glyph, tc.style, raw)
		}
	}
}

func TestLegendShowsEachGlyphAndNoStaticWorkingGlyph(t *testing.T) {
	m := loaded(t, []list.Entry{entry("a", "", "idle", "pi", "m", 0)}, 140, 30)
	legend := legendLine(m)
	for _, want := range []string{
		"\x1b[38;2;110;230;154m⠋\x1b[0m working",
		"\x1b[38;2;245;185;66m○\x1b[0m idle",
		"\x1b[38;2;232;112;95m×\x1b[0m blocked",
		"\x1b[38;2;163;157;144m✓\x1b[0m done",
		"\x1b[38;2;163;157;144m·\x1b[0m unknown",
		"\x1b[38;2;163;157;144m–\x1b[0m gone",
	} {
		if !strings.Contains(legend, want) {
			t.Errorf("legend %q has no %q", ansi.Strip(legend), ansi.Strip(want))
		}
	}
	m = loaded(t, fixture(), 140, 50)
	if strings.Contains(screen(m), "●") {
		t.Fatalf("the static working glyph shows:\n%s", screen(m))
	}
	m = update(t, m, frameMsg{gen: m.frameGen})
	if !strings.Contains(ansi.Strip(legendLine(m)), "⠙ working") {
		t.Fatalf("legend does not follow the spinner: %q", ansi.Strip(legendLine(m)))
	}
	if !strings.Contains(screen(m), "⠙ api-worker") {
		t.Fatalf("card does not follow the spinner:\n%s", screen(m))
	}
}

func TestDetailStatusOfAWorkingAgentSpins(t *testing.T) {
	m := loaded(t, []list.Entry{entry("w", "", "working", "pi", "m", time.Minute)}, 120, 40)
	m = key(t, m, tea.KeyEnter)
	if !strings.Contains(screen(m), "status    ⠋ working") {
		t.Fatalf("detail status has no braille frame:\n%s", screen(m))
	}
	next, cmd := m.Update(frameMsg{gen: m.frameGen})
	m = next.(Model)
	if cmd == nil || !strings.Contains(screen(m), "status    ⠙ working") {
		t.Fatalf("detail status does not spin:\n%s", screen(m))
	}
}

func TestCardTimerShowsAtLeastSinceStartUntilAChange(t *testing.T) {
	e := entry("w", "", "idle", "pi", "m", time.Hour+2*time.Minute+5*time.Second)
	m := loaded(t, []list.Entry{e}, 140, 30)
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " ≥1h02m05s ┃") {
		t.Fatalf("row before a change = %q; want ≥1h02m05s", row)
	}
	m = update(t, m, clockMsg(clock.Add(3*time.Second)))
	m = update(t, m, fetchedMsg{entries: []list.Entry{e}})
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " ≥1h02m08s ┃") {
		t.Fatalf("row after a fetch with no change = %q; want ≥1h02m08s", row)
	}
	e.Status = "working"
	m = update(t, m, fetchedMsg{entries: []list.Entry{e}, at: clock.Add(3 * time.Second)})
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " 0s ┃") || strings.Contains(row, "≥") {
		t.Fatalf("row at the change = %q; want 0s without ≥", row)
	}
	m = update(t, m, clockMsg(clock.Add(10*time.Second)))
	m = update(t, m, fetchedMsg{entries: []list.Entry{e}})
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " 7s ┃") || strings.Contains(row, "≥") {
		t.Fatalf("row 7 s after the change = %q; want 7s", row)
	}
	m = update(t, m, clockMsg(clock.Add(12*time.Second)))
	e.Status = "idle"
	m = update(t, m, fetchedMsg{entries: []list.Entry{e}, at: clock.Add(12 * time.Second)})
	m = update(t, m, clockMsg(clock.Add(14*time.Second)))
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " 2s ┃") {
		t.Fatalf("row 2 s after a second change = %q; want 2s", row)
	}
}

func TestReplacedRecordStartsAFreshTimer(t *testing.T) {
	e := entry("w", "", "idle", "pi", "m", time.Hour)
	m := loaded(t, []list.Entry{e}, 140, 30)
	e.Status = "working"
	m = update(t, m, fetchedMsg{entries: []list.Entry{e}, at: clock})
	m = update(t, m, clockMsg(clock.Add(5*time.Second)))
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " 5s ┃") {
		t.Fatalf("row = %q; want 5s", row)
	}
	fresh := entry("w", "", "idle", "pi", "m", -3*time.Second)
	m = update(t, m, fetchedMsg{entries: []list.Entry{fresh}})
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " ≥2s ┃") {
		t.Fatalf("row of the replaced record = %q; want ≥2s since its own start", row)
	}
	m = update(t, m, fetchedMsg{entries: nil})
	m = update(t, m, fetchedMsg{entries: []list.Entry{e}})
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " ≥1h00m05s ┃") {
		t.Fatalf("row of an agent that left and came back = %q; want ≥1h00m05s", row)
	}
}

func TestDetailsShowTimeInState(t *testing.T) {
	e := entry("w", "", "idle", "pi", "m", 10*time.Minute)
	m := key(t, loaded(t, []list.Entry{e}, 120, 40), tea.KeyEnter)
	if text := screen(m); !strings.Contains(text, "status    ○ idle · ≥10m00s ") || !strings.Contains(text, "elapsed   10m00s") {
		t.Fatalf("details before a change:\n%s", text)
	}
	e.Status = "working"
	m = update(t, m, fetchedMsg{entries: []list.Entry{e}, at: clock})
	m = update(t, m, clockMsg(clock.Add(4*time.Second)))
	if text := screen(m); !strings.Contains(text, "status    ⠋ working · 4s ") || !strings.Contains(text, "elapsed   10m04s") {
		t.Fatalf("details after a change:\n%s", text)
	}
}

// TestStateChangeTimeIsWhenTheFetchSawIt proves that the time in state counts
// from the end of the fetch that saw the change, not from the clock tick
// before it.
func TestStateChangeTimeIsWhenTheFetchSawIt(t *testing.T) {
	now := clock
	entries := []list.Entry{entry("w", "", "idle", "pi", "m", time.Hour)}
	m := New(Options{Fetch: func(context.Context) ([]list.Entry, error) { return entries, nil }, Now: func() time.Time { return now }})
	m = update(t, m, tea.WindowSizeMsg{Width: 140, Height: 30})
	m = update(t, m, m.fetch()())
	m = update(t, m, clockMsg(clock.Add(time.Second)))
	entries = []list.Entry{entry("w", "", "working", "pi", "m", time.Hour)}
	now = clock.Add(1900 * time.Millisecond)
	m = update(t, m, m.fetch()())
	m = update(t, m, clockMsg(clock.Add(2*time.Second)))
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " 0s ┃") {
		t.Fatalf("row 100 ms after the fetch saw the change = %q; want 0s", row)
	}
	m = update(t, m, clockMsg(clock.Add(3*time.Second)))
	if row := cardRow1(t, m, "w"); !strings.Contains(row, " 1s ┃") {
		t.Fatalf("row 1.1 s after the fetch saw the change = %q; want 1s", row)
	}
}

func TestNarrowCardKeepsTheGlyph(t *testing.T) {
	e := entry("solo", "", "idle", "pi", "m", 0)
	for age, want := range map[string]string{
		"≥1000h00m00s": "│ ○ solo       │",
		"≥10m00s":      "│ ○ s… ≥10m00s │",
		"≥1h00m00s":    "│ ○  ≥1h00m00s │",
		"≥10h00m00s":   "│ ○ ≥10h00m00s │",
	} {
		if row := ansi.Strip(cardLines(e, age, false, 0, 16)[1]); row != want {
			t.Errorf("age %s: row 1 = %q; want %q", age, row, want)
		}
	}
}

func TestShortDetailsShowTimeInState(t *testing.T) {
	for height, want := range map[int]string{
		16: "pod status harness model pane parent cwd worktree session started elapsed",
		15: "pod status harness model pane parent cwd worktree session started",
	} {
		m := key(t, loaded(t, []list.Entry{entry("w", "", "idle", "pi", "m", 10*time.Minute)}, 80, height), tea.KeyEnter)
		text := screen(m)
		var labels []string
		for _, line := range strings.Split(text, "\n") {
			if fields := strings.Fields(strings.Trim(strings.TrimSpace(line), "┃")); strings.HasPrefix(strings.TrimSpace(line), "┃ ") && len(fields) > 0 {
				labels = append(labels, fields[0])
			}
		}
		if got := strings.Join(labels, " "); got != want {
			t.Errorf("80x%d detail rows = %q; want %q:\n%s", height, got, want, text)
		}
		if !strings.Contains(text, "status    ○ idle · ≥10m00s ") {
			t.Errorf("80x%d status row has no time in state:\n%s", height, text)
		}
	}
}

func TestUpdatedTimeIsWhenTheFetchReturned(t *testing.T) {
	now := clock
	m := New(Options{Fetch: fetchOf(fixture()), Now: func() time.Time { return now }})
	m = update(t, m, tea.WindowSizeMsg{Width: 140, Height: 50})
	m = update(t, m, clockMsg(clock.Add(time.Second)))
	now = clock.Add(2500 * time.Millisecond)
	m = update(t, m, m.fetch()())
	if text := screen(m); !strings.Contains(text, " updated 14:22:07") {
		t.Fatalf("status does not show the time when the fetch returned:\n%s", text)
	}
}

func TestNarrowDetailsKeepTheWholeStatusRow(t *testing.T) {
	e := entry("w", "", "working", "pi", "m", 26*time.Hour+4*time.Minute+3*time.Second)
	m := key(t, loaded(t, []list.Entry{e}, 40, 20), tea.KeyEnter)
	if text := screen(m); !strings.Contains(text, "┃ status    ⠋ working · ≥26h04m03s ") {
		t.Fatalf("a 40x20 status row is not whole:\n%s", text)
	}
}

func TestEscHidesTheOutlineAndAMoveKeyShowsIt(t *testing.T) {
	m := key(t, listed(t, fixture(), 140, 50), tea.KeyDown)
	m = key(t, m, tea.KeyEscape)
	if text := screen(m); strings.ContainsAny(text, "┏┃") {
		t.Fatalf("esc did not hide the outline:\n%s", text)
	}
	m = key(t, m, tea.KeyDown)
	if text := screen(m); strings.Count(text, "┏") != 1 || !strings.Contains(text, "╰┃ ⠋ api-worker") {
		t.Fatalf("the first down after esc does not show the outline on api-worker:\n%s", text)
	}
	m = key(t, m, tea.KeyDown)
	if m.selected != "hive-coordinator" {
		t.Fatalf("the second down did not move: selected = %q; want hive-coordinator", m.selected)
	}
	if text := screen(m); strings.Count(text, "┏") != 1 || !strings.Contains(text, "┃ ⠋ hive-coordinator") {
		t.Fatalf("the second down did not move the outline:\n%s", text)
	}
}

func TestEnterWhileHiddenShowsTheOutlineAndNoDetails(t *testing.T) {
	m := key(t, loaded(t, fixture(), 140, 50), tea.KeyEscape)
	m = key(t, m, tea.KeyEnter)
	text := screen(m)
	if strings.Contains(text, "esc close") {
		t.Fatalf("enter while hidden opened the details:\n%s", text)
	}
	if strings.Count(text, "┏") != 1 || !strings.Contains(text, "┃ ○ api-coordinator") {
		t.Fatalf("enter while hidden does not show the outline on api-coordinator:\n%s", text)
	}
}

func TestRefreshKeepsTheOutlineHidden(t *testing.T) {
	m := key(t, loaded(t, fixture(), 140, 50), tea.KeyEscape)
	m = update(t, m, fetchedMsg{entries: fixture(), at: clock})
	if text := screen(m); strings.ContainsAny(text, "┏┃") {
		t.Fatalf("a refresh showed the hidden outline:\n%s", text)
	}
	m = update(t, m, fetchedMsg{entries: fixture()[2:], at: clock})
	if text := screen(m); strings.ContainsAny(text, "┏┃") {
		t.Fatalf("a refresh that dropped the selection showed the hidden outline:\n%s", text)
	}
	m = key(t, m, tea.KeyDown)
	if text := screen(m); m.selected != "hive-coordinator" || strings.Count(text, "┏") != 1 || !strings.Contains(text, "┃ ⠋ hive-coordinator") {
		t.Fatalf("down after the refresh does not show the outline on the first card (selected %q):\n%s", m.selected, text)
	}
}

func TestEscInDetailsClosesAndKeepsTheOutline(t *testing.T) {
	m := key(t, loaded(t, fixture(), 140, 50), tea.KeyEnter)
	m = key(t, m, tea.KeyEscape)
	text := screen(m)
	if strings.Contains(text, "esc close") {
		t.Fatalf("esc did not close the details:\n%s", text)
	}
	if strings.Count(text, "┏") != 1 || !strings.Contains(text, "┃ ○ api-coordinator") {
		t.Fatalf("esc in the details hid the outline:\n%s", text)
	}
}

func TestHiddenOutlineDoesNotScroll(t *testing.T) {
	m := listed(t, fixture(), 36, 20)
	for range 7 {
		m = key(t, m, tea.KeyDown)
	}
	m = key(t, m, tea.KeyEscape)
	if m.scroll == 0 {
		t.Fatalf("the view did not scroll to loose")
	}
	m = update(t, m, fetchedMsg{entries: fixture()[:7], at: clock})
	if text := screen(m); m.selected != "api-coordinator" || strings.Contains(text, "api-coordinator") {
		t.Fatalf("selected %q; the view followed the selection while hidden:\n%s", m.selected, text)
	}
	m = key(t, m, tea.KeyDown)
	if text := screen(m); !strings.Contains(text, "┃ ○ api-coordinator") {
		t.Fatalf("the view does not follow the outline when it shows:\n%s", text)
	}
}

func TestGraphKeysShowEscHide(t *testing.T) {
	m := loaded(t, fixture(), 140, 50)
	if text := screen(m); !strings.Contains(text, " [tree] ←↑↓→ move · enter details · esc hide · g list · r refresh · q quit") {
		t.Fatalf("tree key hints have no esc hide:\n%s", text)
	}
	if text := screen(key(t, m, 'g')); !strings.Contains(text, " [list] ←↑↓→ move · enter details · esc hide · g tree · r refresh · q quit") {
		t.Fatalf("list key hints have no esc hide:\n%s", text)
	}
}
