package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestIdleTickDoesNotReloadUnchangedData(t *testing.T) {
	m := newTestModel(t, newTestHome(t), fixedClock(t, time.Now()))
	m.screenIndex = 2
	headerLoads, screenLoads := 0, 0
	m.headerLoader = func(string) (headerState, error) {
		headerLoads++
		return headerState{}, nil
	}
	m.goalsQueueLoader = func(string) (goalsQueueData, error) {
		screenLoads++
		return goalsQueueData{}, nil
	}

	runIdleTick(t, m)
	if headerLoads != 0 || screenLoads != 0 {
		t.Fatalf("unchanged tick started %d header loads and %d screen loads; want none", headerLoads, screenLoads)
	}
}

func TestIdleTickReloadsChangedScreenOnce(t *testing.T) {
	home := newTestHome(t)
	m := newTestModel(t, home, fixedClock(t, time.Now()))
	m.screenIndex = 2
	headerLoads, screenLoads := 0, 0
	m.headerLoader = func(string) (headerState, error) {
		headerLoads++
		return headerState{}, nil
	}
	m.goalsQueueLoader = func(string) (goalsQueueData, error) {
		screenLoads++
		return goalsQueueData{}, nil
	}
	path := filepath.Join(home, ".hive", "state", "goals", "changed.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("change"), 0o644); err != nil {
		t.Fatal(err)
	}

	runIdleTick(t, m)
	if headerLoads != 0 || screenLoads != 1 {
		t.Fatalf("changed tick started %d header loads and %d screen loads; want 0 and 1", headerLoads, screenLoads)
	}
}

func TestHeaderRefreshReusesUnchangedScreenRender(t *testing.T) {
	model := newTestModel(t, newTestHome(t), fixedClock(t, time.Now()))
	renders := 0
	model.screens[0] = countingScreen{renders: &renders}
	model.View()
	if renders != 1 {
		t.Fatalf("initial view rendered screen %d times; want 1", renders)
	}

	model, _ = raise(t, model, headerLoadMsg{state: headerState{projects: 1}})
	model.View()
	if renders != 1 {
		t.Fatalf("header refresh rendered screen %d times; want cached screen", renders)
	}

	model, _ = raise(t, model, screenLoadMsg{request: chatDataRequest})
	model.View()
	if renders != 2 {
		t.Fatalf("screen data refresh rendered screen %d times; want 2", renders)
	}
}

func TestMinuteRefreshUpdatesOnlyTheHeader(t *testing.T) {
	now := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)
	clock := fixedClock(t, now)
	model := newTestModel(t, newTestHome(t), clock)
	renders := 0
	model.screens[0] = countingScreen{renders: &renders}
	before := ansi.Strip(model.View().Content)
	now = now.Add(time.Minute)
	clock.set(now)
	after := ansi.Strip(model.View().Content)

	if renders != 1 {
		t.Fatalf("minute refresh rendered screen %d times; want cached screen", renders)
	}
	stamp := clock.now().Format("15:04")
	if strings.Contains(before, stamp) || !strings.Contains(after, stamp) {
		t.Fatalf("minute refresh did not update the header to %q:\nbefore:\n%s\nafter:\n%s", stamp, before, after)
	}
}

type countingScreen struct{ renders *int }

func (s countingScreen) Update(tea.Msg) tea.Cmd { return nil }

func (s countingScreen) View(int, int) string {
	*s.renders++
	return "screen"
}

func (countingScreen) KeyHints() string { return "" }

func (countingScreen) DataRequest() dataRequest { return chatDataRequest }

func runIdleTick(t *testing.T, model *Model) *Model {
	t.Helper()
	model, cmd := raise(t, model, tickMsg{})
	if cmd == nil {
		t.Fatal("tick did not schedule its refresh check")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok || len(batch) < 2 {
		t.Fatalf("tick returned %T with no refresh check", batch)
	}
	for _, command := range batch[1:] {
		model = runIdleCommand(t, model, command)
	}
	return model
}

func runIdleCommand(t *testing.T, model *Model, cmd tea.Cmd) *Model {
	t.Helper()
	if cmd == nil {
		return model
	}
	msg := cmd()
	model, next := raise(t, model, msg)
	if next == nil {
		return model
	}
	message := next()
	if batch, ok := message.(tea.BatchMsg); ok {
		for _, command := range batch {
			model = runIdleCommand(t, model, command)
		}
		return model
	}
	return runIdleCommand(t, model, func() tea.Msg { return message })
}
