package tui

import (
	"errors"
	"fmt"
	"math"
	"os"
	"syscall"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

type headerState struct {
	supervisor string
	agents     int
	unread     int
}

// refreshHeader reads the header once. A read error becomes a persistent,
// bounded error view: the next ticks do not read again, so one unchanged
// fault cannot log a new event on each tick.
func (m *Model) refreshHeader() {
	if m.headerStopped {
		return
	}
	state, err := readHeader(m.home)
	if err != nil {
		m.headerErr = boundedError(err.Error())
		m.headerStopped = true
		m.logError("header", err)
		return
	}
	m.headerErr = ""
	m.supervisor = state.supervisor
	m.agentCount = state.agents
	m.unread = state.unread
}

func readHeader(home string) (headerState, error) {
	var state headerState
	state.supervisor = "down"
	registry, err := store.Open(home)
	if err != nil {
		return state, err
	}
	var claim records.Claim
	err = registry.Read("claims", "supervisor", &claim)
	switch {
	case err == nil:
		if claim.PID > 0 && processAlive(claim.PID) {
			state.supervisor = "live"
		}
	case errors.Is(err, os.ErrNotExist):
	default:
		return state, err
	}

	agentEnvelope := agent.List(home, agent.ListOptions{All: true})
	if agentEnvelope.Outcome != "ok" {
		return state, fmt.Errorf("agents: %v", agentEnvelope.Error)
	}
	agents, ok := agentEnvelope.Result.([]records.Agent)
	if !ok {
		return state, errors.New("agents: unexpected result")
	}
	for _, record := range agents {
		if record.Status == records.AgentLive {
			state.agents++
		}
	}

	counts, err := inbox.UnreadCountAll(home)
	if err != nil {
		return state, fmt.Errorf("inbox: %w", err)
	}
	for _, count := range counts {
		state.unread += count
	}
	return state, nil
}

// maxPID is the largest signed 32-bit process ID. A larger positive value
// must not reach the process lookup or a signal: the platform truncates it,
// and a negative result such as -1 has process-wide meaning.
const maxPID = int64(math.MaxInt32)

func processAlive(pid int64) bool {
	if pid <= 0 || pid > maxPID {
		return false
	}
	process, err := os.FindProcess(int(pid))
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

func (m *Model) headerLine() string {
	if m.headerErr != "" {
		return "hive " + sanitizeText(m.home) + "  error: " + m.headerErr
	}
	return fmt.Sprintf("hive %s  supervisor: %s  agents: %d  inbox: %d unread",
		sanitizeText(m.home), m.supervisor, m.agentCount, m.unread)
}
