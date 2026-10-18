package supervise

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// fakeClock is a clock that fires only when a test asks.
type fakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
}

type fakeTimer struct {
	ch      chan time.Time
	stopped bool
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) NewTimer(time.Duration) Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	timer := &fakeTimer{ch: make(chan time.Time, 1)}
	c.timers = append(c.timers, timer)
	return timer
}

func (c *fakeClock) fire() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, timer := range c.timers {
		if timer.stopped {
			continue
		}
		select {
		case timer.ch <- c.now:
		default:
		}
	}
}

func (t *fakeTimer) C() <-chan time.Time { return t.ch }
func (t *fakeTimer) Stop() bool {
	t.stopped = true
	return true
}

// fakeStream is one open subscription.
type fakeStream struct {
	events chan herdr.Event
	errs   chan error
	mu     sync.Mutex
	closed bool
}

func newFakeStream() *fakeStream {
	return &fakeStream{events: make(chan herdr.Event, 32), errs: make(chan error, 1)}
}

func (s *fakeStream) Events() <-chan herdr.Event { return s.events }
func (s *fakeStream) Errors() <-chan error       { return s.errs }
func (s *fakeStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		close(s.events)
	}
	return nil
}

func (s *fakeStream) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

// fakeOps is the scripted herdr and agent surface.
type fakeOps struct {
	mu        sync.Mutex
	agent     map[string]herdr.AgentInfo
	agentErr  map[string]error
	pane      map[string]herdr.PaneInfo
	paneErr   map[string]error
	list      []herdr.AgentInfo
	listErr   error
	streams   []*fakeStream
	subs      []herdr.SubscribeParams
	subErr    error
	prompts   []herdr.AgentPromptParams
	promptErr error
	stops     []records.Agent
	getPanes  []string
	stopFn    func(records.Agent) (agent.StopResult, output.Envelope)
}

func newFakeOps() *fakeOps {
	return &fakeOps{
		agent:    map[string]herdr.AgentInfo{},
		agentErr: map[string]error{},
		pane:     map[string]herdr.PaneInfo{},
		paneErr:  map[string]error{},
	}
}

func (o *fakeOps) GetAgent(_ context.Context, paneID string) (herdr.AgentInfo, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.getPanes = append(o.getPanes, paneID)
	if err, ok := o.agentErr[paneID]; ok {
		return herdr.AgentInfo{}, err
	}
	info, ok := o.agent[paneID]
	if !ok {
		return herdr.AgentInfo{}, &herdr.Error{Code: herdr.CodeAgentNotFound, Message: "no agent"}
	}
	return info, nil
}

func (o *fakeOps) GetPane(_ context.Context, paneID string) (herdr.PaneInfo, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err, ok := o.paneErr[paneID]; ok {
		return herdr.PaneInfo{}, err
	}
	info, ok := o.pane[paneID]
	if !ok {
		return herdr.PaneInfo{}, &herdr.Error{Code: herdr.CodePaneNotFound, Message: "no pane"}
	}
	return info, nil
}

func (o *fakeOps) ListAgents(context.Context) ([]herdr.AgentInfo, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]herdr.AgentInfo(nil), o.list...), o.listErr
}

func (o *fakeOps) Prompt(_ context.Context, params herdr.AgentPromptParams) (herdr.AgentInfo, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.prompts = append(o.prompts, params)
	return herdr.AgentInfo{}, o.promptErr
}

func (o *fakeOps) Stop(_ context.Context, _ string, record records.Agent) (agent.StopResult, output.Envelope) {
	o.mu.Lock()
	o.stops = append(o.stops, record)
	fn := o.stopFn
	o.mu.Unlock()
	if fn != nil {
		return fn(record)
	}
	return agent.StopResult{Agent: record}, output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}}
}

func (o *fakeOps) Subscribe(_ context.Context, params herdr.SubscribeParams) (EventStream, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.subErr != nil {
		return nil, o.subErr
	}
	o.subs = append(o.subs, params)
	stream := newFakeStream()
	o.streams = append(o.streams, stream)
	return stream, nil
}

func (o *fakeOps) lastStream() *fakeStream {
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.streams) == 0 {
		return nil
	}
	return o.streams[len(o.streams)-1]
}

func (o *fakeOps) promptCount() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.prompts)
}

func (o *fakeOps) stopCount() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.stops)
}

// testHome makes one home with the marker and the state root.
func testHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	_ = registry
	return home
}

func writeAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

func writeClaim(t *testing.T, home string, claim records.Claim) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, claim.Key, claim); err != nil {
		t.Fatal(err)
	}
}

func readAgent(t *testing.T, home, key string) records.Agent {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatalf("read agent %s: %v", key, err)
	}
	return record
}

func readInbox(t *testing.T, home, role string) []records.Inbox {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var all []records.Inbox
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		raw, err := locked.ListIn([]string{"inbox", role}, nil)
		if err != nil {
			return err
		}
		for _, data := range raw {
			var record records.Inbox
			if err := json.Unmarshal(data, &record); err != nil {
				return err
			}
			all = append(all, record)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return all
}

// handlerSupervisor returns a supervisor with a test home and the given ops.
func handlerSupervisor(t *testing.T, home string, ops Operations, clock Clock) *Supervisor {
	t.Helper()
	s := newSupervisor(Options{Home: home, Operations: ops, Clock: clock})
	s.generation = 1
	s.ownKey = "supervisor-g1"
	return s
}

// eventOf builds one pushed event from the literal data object.
func eventOf(name, data string) herdr.Event {
	return herdr.Event{Event: name, Data: []byte(data)}
}
