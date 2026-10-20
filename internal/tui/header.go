package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	"github.com/Harrison-Blair/hive/internal/project"
	"github.com/Harrison-Blair/hive/internal/task"
)

type headerState struct {
	supervisor string
	projects   int
	agents     int
	working    int
	idle       int
	needsUser  int
}

func readHeader(home string) (headerState, error) {
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
			state.supervisor = "fresh"
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
	state.projects = len(projects)

	agentEnvelope := agent.List(home, agent.ListOptions{All: true})
	if agentEnvelope.Outcome != "ok" {
		return state, fmt.Errorf("agents: %v", agentEnvelope.Error)
	}
	agents, ok := agentEnvelope.Result.([]records.Agent)
	if !ok {
		return state, errors.New("agents: unexpected result")
	}
	for _, record := range agents {
		if record.Status != records.AgentLive {
			continue
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

	taskEnvelope := task.List(home, task.ListOptions{})
	if taskEnvelope.Outcome != "ok" {
		return state, fmt.Errorf("tasks: %v", taskEnvelope.Error)
	}
	tasks, ok := taskEnvelope.Result.([]records.Task)
	if !ok {
		return state, errors.New("tasks: unexpected result")
	}
	for _, record := range tasks {
		if record.NeedsUser != "" {
			state.needsUser++
		}
	}
	return state, nil
}

func (m *Model) agentsHeaderLine() string {
	state := m.header
	projectLabel := "projects"
	if state.projects == 1 {
		projectLabel = "project"
	}
	agentLabel := "agents"
	if state.agents == 1 {
		agentLabel = "agent"
	}
	workCounts := fmt.Sprintf("%d working  %d idle", state.working, state.idle)
	if state.supervisor != "fresh" {
		workCounts = "working unknown  idle unknown"
	}
	supervisor := dimStyle.Render("○")
	if state.supervisor == "fresh" {
		supervisor = workingStyle.Render("●")
	}
	counts := fmt.Sprintf("⬢ hive  %d %s  %d %s  %s", state.projects, projectLabel, state.agents, agentLabel, workCounts)
	status := fmt.Sprintf("supervisor %s  needs you %d", supervisor, state.needsUser)
	timeLabel := m.now().Format("15:04")
	width := m.width
	if width < 1 {
		width = defaultWidth
	}
	if width < 100 {
		padding := max(1, width-lipgloss.Width(status)-lipgloss.Width(timeLabel))
		return counts + "\n" + status + strings.Repeat(" ", padding) + timeLabel
	}
	line := counts + "  " + status
	padding := max(1, width-lipgloss.Width(line)-lipgloss.Width(timeLabel))
	return line + strings.Repeat(" ", padding) + timeLabel
}

func (m *Model) activeHeaderLine() string {
	if m.screenIndex == 1 {
		return m.agentsHeaderLine()
	}
	return m.headerLine()
}

func (m *Model) headerLine() string {
	state := m.header
	if m.headerErr != "" {
		return fmt.Sprintf("⬢ hive · error: %s\n%s", m.headerErr, m.now().Format("15:04"))
	}
	supervisorStyle := dimStyle
	if state.supervisor == "fresh" {
		supervisorStyle = workingStyle
	}
	projectLabel := "projects"
	if state.projects == 1 {
		projectLabel = "project"
	}
	agentLabel := "live agents"
	if state.agents == 1 {
		agentLabel = "live agent"
	}
	workCount := workingStyle.Render(fmt.Sprintf("%d working", state.working))
	idleCount := doneStyle.Render(fmt.Sprintf("%d idle", state.idle))
	if state.supervisor != "fresh" {
		workCount = dimStyle.Render("working unknown")
		idleCount = dimStyle.Render("idle unknown")
	}
	counts := fmt.Sprintf("⬢ hive · %d %s · %d %s · %s / %s",
		state.projects, projectLabel, state.agents, agentLabel, workCount, idleCount)
	status := fmt.Sprintf("supervisor %s · needs_user %s · %s",
		supervisorStyle.Render(state.supervisor), infoStyle.Render(fmt.Sprint(state.needsUser)), m.now().Format("15:04"))
	if m.width < 100 {
		return counts + "\n" + status
	}
	return counts + " · " + status
}
