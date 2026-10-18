package supervise

import "time"

// defaultQuietAfterMin is the quiet period when a record gives none.
const defaultQuietAfterMin = 10

// Clock gives the supervisor its time source.
type Clock interface {
	Now() time.Time
	NewTimer(d time.Duration) Timer
}

// Timer is one clock timer.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
}

type realClock struct{}

type realTimer struct {
	timer *time.Timer
}

func (realClock) Now() time.Time {
	return time.Now()
}

func (realClock) NewTimer(d time.Duration) Timer {
	return realTimer{timer: time.NewTimer(d)}
}

func (t realTimer) C() <-chan time.Time {
	return t.timer.C
}

func (t realTimer) Stop() bool {
	return t.timer.Stop()
}

type quietTimer struct {
	timer  Timer
	cancel chan struct{}
}

// quietManager owns one quiet timer for each idle agent.
type quietManager struct {
	clock  Clock
	fired  chan string
	timers map[string]*quietTimer
	done   chan struct{}
}

func newQuietManager(clock Clock) *quietManager {
	return &quietManager{
		clock:  clock,
		fired:  make(chan string, 8),
		timers: map[string]*quietTimer{},
		done:   make(chan struct{}),
	}
}

// start replaces the quiet timer of one agent.
func (m *quietManager) start(key string, after time.Duration) {
	m.cancel(key)
	timer := m.clock.NewTimer(after)
	entry := &quietTimer{timer: timer, cancel: make(chan struct{})}
	m.timers[key] = entry
	go func() {
		select {
		case <-timer.C():
			select {
			case m.fired <- key:
			case <-entry.cancel:
			case <-m.done:
			}
		case <-entry.cancel:
		case <-m.done:
		}
	}()
}

// cancel stops the quiet timer of one agent.
func (m *quietManager) cancel(key string) {
	entry, ok := m.timers[key]
	if !ok {
		return
	}
	delete(m.timers, key)
	entry.timer.Stop()
	close(entry.cancel)
}

// close stops each quiet timer.
func (m *quietManager) close() {
	close(m.done)
	for key := range m.timers {
		m.cancel(key)
	}
}
