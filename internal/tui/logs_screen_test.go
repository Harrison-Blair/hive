package tui

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/logs"
	"github.com/charmbracelet/x/ansi"
)

type synchronizedOutput struct {
	mu      sync.Mutex
	output  strings.Builder
	changed chan struct{}
}

func newSynchronizedOutput() *synchronizedOutput {
	return &synchronizedOutput{changed: make(chan struct{}, 1)}
}

func (w *synchronizedOutput) Write(data []byte) (int, error) {
	w.mu.Lock()
	n, err := w.output.Write(data)
	w.mu.Unlock()
	select {
	case w.changed <- struct{}{}:
	default:
	}
	return n, err
}

func (w *synchronizedOutput) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.output.String()
}

func waitForOutput(t *testing.T, output *synchronizedOutput, wants ...string) string {
	t.Helper()
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()
	for {
		view := ansi.Strip(output.String())
		found := true
		for _, want := range wants {
			if !strings.Contains(view, want) {
				found = false
				break
			}
		}
		if found {
			return view
		}
		select {
		case <-output.changed:
		case <-timeout.C:
			t.Fatalf("program output did not reach state containing %q:\n%s", strings.Join(wants, ", "), view)
			return view
		}
	}
}

func logsFixture(t *testing.T, day time.Time) []logs.Event {
	t.Helper()
	home := t.TempDir()
	path := dailyPath(home, day)
	writeEvents(t, path,
		logsFixtureEvent(t, day, "DEBUG", "cli", "command start", map[string]any{"command": "task list", "request_id": "r-11111111"}),
		logsFixtureEvent(t, day.Add(time.Second), "INFO", "task", "queued", map[string]any{"duration_ms": 4, "project_id": "demo", "agent_id": "worker-1", "request_id": "r-11111111"}),
		logsFixtureEvent(t, day.Add(2*time.Second), "DEBUG", "cli", "command end", map[string]any{"command": "task list", "duration_ms": 12, "outcome": "ok", "request_id": "r-11111111"}),
		logsFixtureEvent(t, day.Add(3*time.Second), "WARN", "herdr", "network retry", map[string]any{"project_id": "other", "agent_id": "worker-2", "request_id": "r-22222222"}),
	)
	reader := logs.NewReaderWithClock(home, func() time.Time { return day })
	events, _, err := reader.Read(t.Context(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func logsFixtureEvent(t *testing.T, at time.Time, level, component, message string, extra map[string]any) string {
	t.Helper()
	event := map[string]any{"time": at.Format(time.RFC3339Nano), "level": level, "component": component, "msg": message}
	for key, value := range extra {
		event[key] = value
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
