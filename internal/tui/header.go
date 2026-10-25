package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/ask"
	"github.com/Harrison-Blair/hive/internal/doctor"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/project"
	"github.com/Harrison-Blair/hive/internal/task"
)

var supervisorBinaryCache struct {
	sync.Mutex
	key string
	old bool
}

type headerState struct {
	coordinator          string
	coordinatorChangedAt string
	supervisor           string
	supervisorChangedAt  string
	supervisorOld        bool
	projects             int
	agents               int
	working              int
	idle                 int
	needsUser            int
	asks                 int
	askUndelivered       int
	queueRun             int
	queueStops           int
	queueWaits           int
}

func readHeader(ctx context.Context, home string) (headerState, error) {
	state := headerState{supervisor: "stale"}
	registry, err := store.Open(home)
	if err != nil {
		return state, err
	}
	var claim records.Claim
	err = registry.Read("claims", "supervisor", &claim)
	switch {
	case err == nil:
		if alive, checkErr := process.Alive(claim.PID, claim.PIDStart); checkErr == nil && alive {
			state.supervisor, state.supervisorChangedAt = "fresh", claim.ClaimedAt
			state.supervisorOld = supervisorBinaryIsOld(ctx, home, claim)
		}
	case errors.Is(err, os.ErrNotExist):
	default:
		return state, err
	}

	projectEnvelope := project.List(home)
	if projectEnvelope.Outcome != "ok" {
		return state, fmt.Errorf("projects: %v", projectEnvelope.Error)
	}
	projects, ok := projectEnvelope.Result.([]records.Project)
	if !ok {
		return state, errors.New("projects: unexpected result")
	}

	agentEnvelope := agent.List(home, agent.ListOptions{All: true})
	if agentEnvelope.Outcome != "ok" {
		return state, fmt.Errorf("agents: %v", agentEnvelope.Error)
	}
	agents, ok := agentEnvelope.Result.([]records.Agent)
	if !ok {
		return state, errors.New("agents: unexpected result")
	}

	taskEnvelope := task.List(home, task.ListOptions{})
	if taskEnvelope.Outcome != "ok" {
		return state, fmt.Errorf("tasks: %v", taskEnvelope.Error)
	}
	tasks, ok := taskEnvelope.Result.([]records.Task)
	if !ok {
		return state, errors.New("tasks: unexpected result")
	}
	asks, err := ask.List(home)
	if err != nil {
		return state, fmt.Errorf("asks: %w", err)
	}
	return buildHeader(state.supervisor == "fresh", state.supervisorOld, projects, agents, tasks, asks), nil
}

func buildHeader(supervisorFresh, supervisorOld bool, projects []records.Project, agents []records.Agent, tasks []records.Task, asks []ask.Record) headerState {
	state := headerState{supervisor: "stale", projects: len(projects)}
	if supervisorFresh {
		state.supervisor = "fresh"
		state.supervisorOld = supervisorOld
	}
	currentAgents := currentAgentKeys(agents)
	for _, record := range agents {
		if record.Status != records.AgentLive {
			continue
		}
		if record.Role == "coordinator" && currentAgents[record.Key] {
			state.coordinator = "live"
			if record.AgentStatus != "" {
				state.coordinator = record.AgentStatus
			}
			state.coordinatorChangedAt = record.StatusChangedAt
		}
		state.agents++
		if state.supervisor == "fresh" {
			if record.AgentStatus == "working" {
				state.working++
			} else {
				state.idle++
			}
		}
	}
	for _, record := range tasks {
		if record.NeedsUser != "" {
			state.needsUser++
		}
	}
	state.queueRun, state.queueStops, state.queueWaits = queueTaskCounts(tasks)
	for _, record := range asks {
		switch {
		case record.State == ask.StateOpen:
			state.asks++
		case record.DeliveredAt == "":
			state.askUndelivered++
		}
	}
	return state
}

func supervisorBinaryIsOld(ctx context.Context, home string, claim records.Claim) bool {
	hivePath, err := exec.LookPath("hive")
	if err != nil {
		return false
	}
	hiveInfo, err := os.Stat(hivePath)
	if err != nil {
		return false
	}
	supervisorInfo, err := os.Stat(fmt.Sprintf("/proc/%d/exe", claim.PID))
	if err != nil {
		return false
	}
	key := fmt.Sprintf("%s:%d:%d:%s:%d:%d:%d:%d", home, claim.PID, claim.PIDStart, hivePath, hiveInfo.Size(), hiveInfo.ModTime().UnixNano(), supervisorInfo.Size(), supervisorInfo.ModTime().UnixNano())
	supervisorBinaryCache.Lock()
	defer supervisorBinaryCache.Unlock()
	if supervisorBinaryCache.key == key {
		return supervisorBinaryCache.old
	}
	old := false
	for _, check := range doctor.Run(ctx, home, nil).Checks {
		if check.Name == "supervisor binary" {
			old = check.Status == doctor.StatusWarn && strings.Contains(check.Detail, "supervisor binary is older")
			break
		}
	}
	supervisorBinaryCache.key = key
	supervisorBinaryCache.old = old
	return old
}
