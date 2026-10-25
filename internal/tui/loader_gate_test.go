package tui

import (
	"log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/doctor"
)

type loaderDispatch struct {
	name      string
	screen    int
	setScreen bool
	start     func(*Model) tea.Cmd
	check     func(*testing.T, tea.Msg)
}

// loaderDispatches lists each home-access loader callback of the task 179
// inventory.
func loaderDispatches() []loaderDispatch {
	screenCheck := func(request dataRequest) func(*testing.T, tea.Msg) {
		return func(t *testing.T, msg tea.Msg) {
			t.Helper()
			loaded, ok := msg.(screenLoadMsg)
			if !ok {
				t.Fatalf("loader gave %T; want screenLoadMsg", msg)
			}
			if loaded.request != request {
				t.Fatalf("screen load request = %d; want %d", loaded.request, request)
			}
		}
	}
	return []loaderDispatch{
		{
			name:  "header",
			start: func(m *Model) tea.Cmd { return m.startHeaderLoad() },
			check: func(t *testing.T, msg tea.Msg) {
				t.Helper()
				if _, ok := msg.(headerLoadMsg); !ok {
					t.Fatalf("loader gave %T; want headerLoadMsg", msg)
				}
			},
		},
		{
			name:   "agents",
			screen: 1, setScreen: true,
			start: func(m *Model) tea.Cmd { return m.startCurrentScreenLoad() },
			check: screenCheck(agentsDataRequest),
		},
		{
			name:   "chat",
			screen: 0, setScreen: true,
			start: func(m *Model) tea.Cmd { return m.startCurrentScreenLoad() },
			check: screenCheck(chatDataRequest),
		},
		{
			name:   "health composite",
			screen: 6, setScreen: true,
			start: func(m *Model) tea.Cmd { return m.startCurrentScreenLoad() },
			check: screenCheck(healthDataRequest),
		},
		{
			name:   "work",
			screen: 2, setScreen: true,
			start: func(m *Model) tea.Cmd { return m.startCurrentScreenLoad() },
			check: screenCheck(tasksDataRequest),
		},
		{
			name:   "friction",
			screen: 5, setScreen: true,
			start: func(m *Model) tea.Cmd { return m.startCurrentScreenLoad() },
			check: screenCheck(frictionDataRequest),
		},
		{
			name:   "logs",
			screen: 7, setScreen: true,
			start: func(m *Model) tea.Cmd { return m.startCurrentScreenLoad() },
			check: screenCheck(logsDataRequest),
		},
		{
			name:  "refresh",
			start: func(m *Model) tea.Cmd { return m.startRefreshCheck() },
			check: func(t *testing.T, msg tea.Msg) {
				t.Helper()
				if _, ok := msg.(refreshCheckMsg); !ok {
					t.Fatalf("loader gave %T; want refreshCheckMsg", msg)
				}
			},
		},
		{
			name: "task file",
			start: func(m *Model) tea.Cmd {
				return m.startTaskFileRead(filepath.Join(m.home, ".hive", "README.md"), "README.md")
			},
			check: func(t *testing.T, msg tea.Msg) {
				t.Helper()
				file, ok := msg.(taskFileMsg)
				if !ok {
					t.Fatalf("loader gave %T; want taskFileMsg", msg)
				}
				if file.err != nil || file.content != "home\n" {
					t.Fatalf("task file load = %q, %v; want the README text", file.content, file.err)
				}
			},
		},
	}
}

// TestLoaderGateSettlesEachDispatch holds one admitted callback of each
// inventory row. Shutdown must not return until the callback settles.
func TestLoaderGateSettlesEachDispatch(t *testing.T) {
	for _, tc := range loaderDispatches() {
		t.Run(tc.name, func(t *testing.T) {
			home := newTestHome(t)
			model := New(home, nil)
			defer model.Stop()
			if tc.setScreen {
				model.screenIndex = tc.screen
			}
			var begins, ends atomic.Int32
			started := make(chan struct{})
			release := make(chan struct{})
			var startOnce sync.Once
			model.loads.testHook = func(event string) {
				switch event {
				case "begin":
					startOnce.Do(func() { close(started) })
					<-release
					begins.Add(1)
				case "end":
					ends.Add(1)
				}
			}

			cmd := tc.start(model)
			if cmd == nil {
				t.Fatal("the dispatch returned no command")
			}
			result := make(chan tea.Msg, 1)
			go func() { result <- cmd() }()
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("the admitted loader did not start")
			}
			closed := make(chan struct{})
			go func() {
				model.loads.close()
				close(closed)
			}()
			select {
			case <-closed:
				t.Fatal("shutdown returned while the admitted loader was outstanding")
			case <-time.After(200 * time.Millisecond):
			}
			close(release)
			var msg tea.Msg
			select {
			case msg = <-result:
			case <-time.After(5 * time.Second):
				t.Fatal("the admitted loader did not settle")
			}
			select {
			case <-closed:
			case <-time.After(5 * time.Second):
				t.Fatal("shutdown did not return after the loader settled")
			}
			if got := begins.Load(); got != 1 {
				t.Fatalf("admissions = %d; want 1", got)
			}
			if got := ends.Load(); got != 1 {
				t.Fatalf("settlements = %d; want 1", got)
			}
			tc.check(t, msg)
		})
	}
}

// TestLoaderGateRefusesEachDispatchAfterShutdown checks that no inventory row
// starts home access after shutdown.
func TestLoaderGateRefusesEachDispatchAfterShutdown(t *testing.T) {
	for _, tc := range loaderDispatches() {
		t.Run(tc.name, func(t *testing.T) {
			home := newTestHome(t)
			model := New(home, nil)
			defer model.Stop()
			if tc.setScreen {
				model.screenIndex = tc.screen
			}
			var began atomic.Bool
			model.loads.testHook = func(event string) {
				if event == "begin" {
					began.Store(true)
				}
			}
			model.loads.close()

			cmd := tc.start(model)
			if cmd == nil {
				t.Fatal("the dispatch returned no command")
			}
			if msg := cmd(); msg != nil {
				if _, ok := msg.(loaderStoppedMsg); !ok {
					t.Fatalf("a callback after shutdown gave %T; want loaderStoppedMsg", msg)
				}
			} else {
				t.Fatal("a callback after shutdown gave no message; want loaderStoppedMsg")
			}
			if began.Load() {
				t.Fatal("a loader began home access after shutdown")
			}
		})
	}
}

// TestLoaderGateHealthCompositeLoadsAgentsAndChecks checks that one admitted
// Health callback runs the agents loader and then the doctor loader.
func TestLoaderGateHealthCompositeLoadsAgentsAndChecks(t *testing.T) {
	home := newTestHome(t)
	model := New(home, nil)
	defer model.Stop()
	model.screenIndex = 6
	realAgents, realHealth := model.agentsLoader, model.healthLoader
	var agentsCalls, healthCalls atomic.Int32
	model.agentsLoader = func(home string, now time.Time) (agentsData, error) {
		agentsCalls.Add(1)
		return realAgents(home, now)
	}
	model.healthLoader = func(home string, logger *slog.Logger) (doctor.Result, error) {
		healthCalls.Add(1)
		return realHealth(home, logger)
	}

	cmd := model.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("the Health dispatch returned no command")
	}
	raw := cmd()
	msg, ok := raw.(screenLoadMsg)
	if !ok {
		t.Fatalf("Health loader gave %T; want screenLoadMsg", raw)
	}
	if msg.request != healthDataRequest {
		t.Fatalf("Health loader request = %d; want %d", msg.request, healthDataRequest)
	}
	if got := agentsCalls.Load(); got != 1 {
		t.Fatalf("agents loader calls = %d; want 1", got)
	}
	if got := healthCalls.Load(); got != 1 {
		t.Fatalf("doctor loader calls = %d; want 1", got)
	}
}

// TestLoaderGateDoesNotWaitForUndispatchedCallbacks checks that a command that
// Bubble Tea never starts does not hold shutdown open.
func TestLoaderGateDoesNotWaitForUndispatchedCallbacks(t *testing.T) {
	for _, tc := range loaderDispatches() {
		t.Run(tc.name, func(t *testing.T) {
			home := newTestHome(t)
			model := New(home, nil)
			defer model.Stop()
			if tc.setScreen {
				model.screenIndex = tc.screen
			}
			if cmd := tc.start(model); cmd == nil {
				t.Fatal("the dispatch returned no command")
			}
			closed := make(chan struct{})
			go func() {
				model.loads.close()
				close(closed)
			}()
			select {
			case <-closed:
			case <-time.After(5 * time.Second):
				t.Fatal("shutdown waited for an undispatched loader")
			}
		})
	}
}
