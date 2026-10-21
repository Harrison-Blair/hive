package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/Harrison-Blair/hive/internal/agent"
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
	supervisor    string
	supervisorOld bool
	projects      int
	agents        int
	working       int
	idle          int
	needsUser     int
	queueRun      int
	queueStops    int
	queueWaits    int
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
			state.supervisorOld = supervisorBinaryIsOld(home, claim)
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
	state.queueRun, state.queueStops, state.queueWaits = queueTaskCounts(tasks)
	return state, nil
}

func supervisorBinaryIsOld(home string, claim records.Claim) bool {
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
	for _, check := range doctor.Run(context.Background(), home, nil).Checks {
		if check.Name == "supervisor binary" {
			old = check.Status == doctor.StatusWarn && strings.Contains(check.Detail, "supervisor binary is older")
			break
		}
	}
	supervisorBinaryCache.key = key
	supervisorBinaryCache.old = old
	return old
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
	workCounts := fmt.Sprintf("%s  %d idle", workingStyle.Render(fmt.Sprintf("%d working", state.working)), state.idle)
	if state.supervisor != "fresh" {
		workCounts = dimStyle.Render("working unknown  idle unknown")
	}
	supervisor := dimStyle.Render("○")
	if state.supervisor == "fresh" {
		supervisor = workingStyle.Render("●")
	}
	if state.supervisorOld {
		supervisor = dimStyle.Render("old")
	}
	counts := fmt.Sprintf("⬢ hive  %d %s  %d %s  %s", state.projects, projectLabel, state.agents, agentLabel, workCounts)
	needsUser := lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Render(fmt.Sprint(state.needsUser))
	status := fmt.Sprintf("supervisor %s", supervisor)
	if state.supervisorOld {
		status += " · r to restart"
	}
	status += fmt.Sprintf("  needs you %s", needsUser)
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
	if m.screenIndex == 3 && m.headerErr == "" {
		return m.queueHeaderLine()
	}
	return m.headerLine()
}

func (m *Model) queueHeaderLine() string {
	line := fmt.Sprintf("⬢ hive · runs without you %d · stops %d · waits behind a stop %d", m.header.queueRun, m.header.queueStops, m.header.queueWaits)
	timeLabel := m.now().Format("15:04")
	if m.width < 100 {
		return line + "\n" + timeLabel
	}
	padding := max(1, m.width-lipgloss.Width(line)-lipgloss.Width(timeLabel))
	return line + strings.Repeat(" ", padding) + timeLabel
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
	supervisorLabel := state.supervisor
	if state.supervisorOld {
		supervisorLabel = "old"
	}
	status := fmt.Sprintf("supervisor %s · needs_user %s · %s",
		supervisorStyle.Render(supervisorLabel), infoStyle.Render(fmt.Sprint(state.needsUser)), m.now().Format("15:04"))
	if state.supervisorOld {
		status += " · r to restart"
	}
	if m.width < 100 {
		return counts + "\n" + status
	}
	return counts + " · " + status
}
