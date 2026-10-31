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
	return update(t, m, fetchedMsg{entries: entries})
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

func TestCardShowsNameAndElapsedThenHarnessAndModel(t *testing.T) {
	lines := cardLines(entry("hive-coordinator", "", "idle", "claude", "opus", 4*time.Minute+5*time.Second), clock, false, 0, 28)
	if len(lines) != 4 {
		t.Fatalf("card has %d lines; want 4", len(lines))
	}
	row1, row2 := ansi.Strip(lines[1]), ansi.Strip(lines[2])
	if !strings.HasPrefix(row1, "│ ◐ hive-coordinator") || !strings.HasSuffix(row1, " 4m05s │") {
		t.Errorf("row 1 = %q; want the idle glyph, the name, and 4m05s", row1)
	}
	if row2 != "│ claude · opus"+strings.Repeat(" ", 12)+"│" {
		t.Errorf("row 2 = %q; want harness · model and no status", row2)
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w != 28 {
			t.Errorf("line %q is %d cells; want 28", ansi.Strip(line), w)
		}
	}
}

func TestCardKeepsACoordinatorNameWithALongElapsedTime(t *testing.T) {
	for _, age := range []time.Duration{10 * time.Minute, 26*time.Hour + 4*time.Minute + 3*time.Second} {
		e := entry("hive-coordinator", "", "idle", "claude", "opus", age)
		text := screen(loaded(t, []list.Entry{e}, 140, 20))
		if !strings.Contains(text, " hive-coordinator ") || !strings.Contains(text, " "+elapsed(age)+" ") {
			t.Errorf("age %v: card row 1 has no full name and elapsed time:\n%s", age, text)
		}
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
	text := screen(loaded(t, fixture(), 140, 50))
	for _, want := range []string{
		"14:22:05", "╭─ api ", "╭─ hive ", "╭─ home ",
		"├──", "├┤ ✓ hive-reviewer", "╰┤ ▲ hive-worker ", "╰┤ ◐ hive-worker-kid",
		"┃ ◐ api-coordinator", "● working", "◐ idle", "▲ blocked", "✓ done", "○ gone",
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
	m := loaded(t, fixture(), 140, 50)
	text := screen(m)
	if strings.Count(text, "┏") != 1 || !strings.Contains(text, "┃ ◐ api-coordinator") {
		t.Fatalf("first card is not the one heavy card:\n%s", text)
	}
	m = key(t, m, tea.KeyDown)
	text = screen(m)
	if strings.Count(text, "┏") != 1 || !strings.Contains(text, "╰┨ ● api-worker") {
		t.Fatalf("down did not move the heavy border:\n%s", text)
	}
}

func TestNavigation(t *testing.T) {
	m := loaded(t, fixture(), 140, 50)
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
	m := loaded(t, fixture(), 36, 120)
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
	m := loaded(t, fixture(), 36, 20)
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
	if !strings.Contains(after, "⠋ api-worker") || strings.Contains(before, "⠋") {
		t.Fatalf("frame 1 does not show the first braille frame:\nbefore\n%s\nafter\n%s", before, after)
	}
	if !strings.Contains(after, "◐ api-coordinator") {
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
	m := loaded(t, entries, 40, 16)
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
	for range 6 {
		m = key(t, m, tea.KeyEscape)
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
