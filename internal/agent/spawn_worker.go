package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/gitx"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// taskChoice is the task, the worktree, and the pane of one worker or verifier
// spawn. It also holds the topology that this invocation made.
type taskChoice struct {
	PaneID           string
	Task             records.Task
	Project          records.Project
	Worktree         records.Worktree
	Branch           string
	CreatedWorkspace bool
	CreatedTab       bool
}

// resolveTaskPane checks the task and the pane of one worker or verifier
// spawn. It returns the choice and an error message.
func resolveTaskPane(ctx context.Context, home string, options SpawnOptions, logger *slog.Logger) (taskChoice, string) {
	if !records.ValidTaskID(options.Task) {
		return taskChoice{}, "task not found: " + displayValue(options.Task)
	}
	registry, err := store.Open(home)
	if err != nil {
		return taskChoice{}, "store: " + err.Error()
	}
	var task records.Task
	if err := registry.ReadIn([]string{"tasks"}, options.Task, &task); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return taskChoice{}, "task not found: " + displayValue(options.Task)
		}
		return taskChoice{}, "store: " + err.Error()
	}
	if options.Project != "" && options.Project != task.Project {
		return taskChoice{}, fmt.Sprintf("task %s is in project %s", task.ID, task.Project)
	}
	var project records.Project
	if err := registry.ReadIn([]string{"projects"}, task.Project, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return taskChoice{}, "project not found: " + displayValue(task.Project)
		}
		return taskChoice{}, "store: " + err.Error()
	}
	choice := taskChoice{Task: task, Project: project}
	switch options.Role {
	case "worker":
		if message := checkWorkerTask(&choice); message != "" {
			return taskChoice{}, message
		}
	case "verifier":
		if message := checkVerifierTask(&choice, registry); message != "" {
			return taskChoice{}, message
		}
	default:
		return taskChoice{}, "invalid role: " + displayValue(options.Role)
	}
	if options.Role == "verifier" {
		resolved, message := resolveVerifierTopology(ctx, home, choice, logger)
		if message != "" {
			return taskChoice{}, message
		}
		choice = resolved
	}
	pane, message := checkWorktreePane(ctx, home, choice.Worktree, project, logger)
	if message != "" {
		choice.cleanup(ctx, home, logger)
		return taskChoice{}, message
	}
	choice.PaneID = pane.PaneID
	if choice.Branch == "" {
		branch, err := gitx.CurrentBranch(ctx, choice.Worktree.Path)
		if err != nil {
			choice.cleanup(ctx, home, logger)
			return taskChoice{}, "worktree branch: " + err.Error()
		}
		choice.Branch = branch
	}
	return choice, ""
}

// checkWorkerTask holds the task rules of one worker spawn.
func checkWorkerTask(choice *taskChoice) string {
	if choice.Task.Status != "created" && choice.Task.Status != "assigned" {
		return "task is " + choice.Task.Status
	}
	if choice.Task.Worktree == nil || choice.Task.Worktree.RootPane == "" {
		return "task has no worktree: run hive worktree create --task " + choice.Task.ID
	}
	choice.Worktree = *choice.Task.Worktree
	return ""
}

// checkVerifierTask holds the task rules of one verifier spawn.
func checkVerifierTask(choice *taskChoice, registry *store.Store) string {
	if choice.Task.Status != "completed" {
		return "task is " + choice.Task.Status
	}
	agents, err := loadAgentRecords(registry)
	if err != nil {
		return "store: " + err.Error()
	}
	for _, record := range agents {
		if record.Task != choice.Task.ID || record.Role != "worker" {
			continue
		}
		switch record.Status {
		case records.AgentLive:
			return "worker still live: " + record.Name
		case records.AgentUnresolved:
			return "pane has an unresolved record: " + record.Name
		case records.AgentStopped, records.AgentExited:
		default:
			return "worker is " + record.Status + ": " + record.Name
		}
	}
	if choice.Task.Worktree == nil || choice.Task.Worktree.RootPane == "" {
		return "task has no worktree: run hive worktree create --task " + choice.Task.ID
	}
	choice.Worktree = *choice.Task.Worktree
	return ""
}

// resolveVerifierTopology reopens the worktree or makes a new tab for one
// verifier spawn. It records the topology that this invocation made.
func resolveVerifierTopology(ctx context.Context, home string, choice taskChoice, logger *slog.Logger) (taskChoice, string) {
	workspace, err := herdr.WorkspaceGet(ctx, choice.Worktree.WorkspaceID, herdr.WithLogger(logger))
	if err == nil {
		created, createErr := herdr.TabCreate(ctx, herdr.TabCreateParams{
			WorkspaceID: workspace.WorkspaceID,
			Cwd:         choice.Worktree.Path,
			Focus:       false,
		}, herdr.WithLogger(logger))
		if createErr != nil {
			return choice, "tab create: " + createErr.Error()
		}
		choice.Worktree.WorkspaceID = workspace.WorkspaceID
		choice.Worktree.TabID = created.Tab.TabID
		choice.Worktree.RootPane = created.RootPane.PaneID
		choice.CreatedTab = true
		if err := writeTaskWorktree(home, choice.Task.ID, choice.Worktree); err != nil {
			choice.cleanup(ctx, home, logger)
			return choice, "store: " + err.Error()
		}
		return choice, ""
	}
	if !isHerdrCode(err, herdr.CodeWorkspaceNotFound) {
		return choice, "workspace get: " + err.Error()
	}
	opened, openErr := herdr.WorktreeOpen(ctx, herdr.WorktreeOpenParams{
		Cwd:   choice.Project.Path,
		Path:  choice.Worktree.Path,
		Label: choice.Task.ID,
		Focus: false,
	}, herdr.WithLogger(logger))
	if openErr != nil {
		return choice, "worktree open: " + openErr.Error()
	}
	choice.Worktree.WorkspaceID = opened.Workspace.WorkspaceID
	choice.Worktree.TabID = opened.Tab.TabID
	choice.Worktree.RootPane = opened.RootPane.PaneID
	choice.CreatedWorkspace = !opened.AlreadyOpen
	choice.Branch = opened.Worktree.Branch
	if err := writeTaskWorktree(home, choice.Task.ID, choice.Worktree); err != nil {
		choice.cleanup(ctx, home, logger)
		return choice, "store: " + err.Error()
	}
	return choice, ""
}

// checkWorktreePane checks the recorded root pane and the workspace metadata of
// one worktree.
func checkWorktreePane(ctx context.Context, home string, worktree records.Worktree, project records.Project, logger *slog.Logger) (herdr.PaneInfo, string) {
	pane, err := herdr.PaneGet(ctx, worktree.RootPane, herdr.WithLogger(logger))
	if err != nil {
		if isHerdrCode(err, herdr.CodePaneNotFound) {
			return herdr.PaneInfo{}, "worktree pane mismatch: pane_id"
		}
		return herdr.PaneInfo{}, "pane get: " + err.Error()
	}
	if pane.PaneID != worktree.RootPane {
		return herdr.PaneInfo{}, "worktree pane mismatch: pane_id"
	}
	if pane.WorkspaceID != worktree.WorkspaceID {
		return herdr.PaneInfo{}, "worktree pane mismatch: workspace_id"
	}
	if !sameDirectory(paneCwd(pane), worktree.Path) {
		return herdr.PaneInfo{}, "worktree pane mismatch: cwd"
	}
	if pane.RestoreError != nil {
		return herdr.PaneInfo{}, "pane not restored: " + *pane.RestoreError
	}
	if pane.Agent != "" || pane.AgentStatus != "unknown" {
		return herdr.PaneInfo{}, busyPaneMessage(ctx, pane, logger)
	}
	if name, found := unresolvedRecordForPane(home, pane.PaneID); found {
		return herdr.PaneInfo{}, "pane has an unresolved record: " + name
	}
	workspace, err := herdr.WorkspaceGet(ctx, worktree.WorkspaceID, herdr.WithLogger(logger))
	if err != nil {
		if isHerdrCode(err, herdr.CodeWorkspaceNotFound) {
			return herdr.PaneInfo{}, "worktree pane mismatch: workspace_id"
		}
		return herdr.PaneInfo{}, "workspace get: " + err.Error()
	}
	if workspace.Worktree == nil {
		return herdr.PaneInfo{}, "worktree pane mismatch: worktree"
	}
	if !sameDirectory(workspace.Worktree.RepoRoot, project.Path) {
		return herdr.PaneInfo{}, "worktree pane mismatch: repo_root"
	}
	if !sameDirectory(workspace.Worktree.CheckoutPath, worktree.Path) {
		return herdr.PaneInfo{}, "worktree pane mismatch: checkout_path"
	}
	return pane, ""
}

// busyPaneMessage names the agent that holds one pane.
func busyPaneMessage(ctx context.Context, pane herdr.PaneInfo, logger *slog.Logger) string {
	if info, err := herdr.AgentGet(ctx, pane.PaneID, herdr.WithLogger(logger)); err == nil && info.Name != "" {
		return "pane busy: " + info.Name
	}
	if pane.Agent != "" {
		return "pane busy: " + pane.Agent + " agent"
	}
	return "pane busy: " + pane.AgentStatus + " agent"
}

// unresolvedRecordForPane returns the name of one unresolved record that
// holds a pane.
func unresolvedRecordForPane(home, paneID string) (string, bool) {
	registry, err := store.Open(home)
	if err != nil {
		return "", false
	}
	agents, err := loadAgentRecords(registry)
	if err != nil {
		return "", false
	}
	for _, record := range agents {
		if record.PaneID == paneID && record.Status == records.AgentUnresolved {
			return record.Name, true
		}
	}
	return "", false
}

// loadAgentRecords reads each agent record of one store.
func loadAgentRecords(registry *store.Store) ([]records.Agent, error) {
	raw, err := registry.ListIn([]string{"agents"}, func(string) bool { return true })
	if err != nil {
		return nil, err
	}
	found := make([]records.Agent, 0, len(raw))
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, err
		}
		found = append(found, record)
	}
	return found, nil
}

// writeTaskWorktree writes the worktree fields of one task record.
func writeTaskWorktree(home, taskID string, worktree records.Worktree) error {
	registry, err := store.Open(home)
	if err != nil {
		return err
	}
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var task records.Task
		if err := locked.ReadIn([]string{"tasks"}, taskID, &task); err != nil {
			return err
		}
		copied := worktree
		task.Worktree = &copied
		task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		return locked.WriteIn([]string{"tasks"}, taskID, task)
	})
}

// cleanup closes only the topology that this invocation made. It checks the
// topology again before each close.
func (choice *taskChoice) cleanup(ctx context.Context, home string, logger *slog.Logger) {
	if choice == nil {
		return
	}
	if choice.CreatedWorkspace {
		if _, err := herdr.WorkspaceGet(ctx, choice.Worktree.WorkspaceID, herdr.WithLogger(logger)); err == nil {
			if err := herdr.WorkspaceClose(ctx, choice.Worktree.WorkspaceID, false, herdr.WithLogger(logger)); err != nil {
				logTaskCleanup(logger, choice, "workspace", err.Error())
			}
		}
	}
	if choice.CreatedTab {
		pane, err := herdr.PaneGet(ctx, choice.Worktree.RootPane, herdr.WithLogger(logger))
		if err == nil && pane.WorkspaceID == choice.Worktree.WorkspaceID && pane.TabID == choice.Worktree.TabID {
			if err := herdr.PaneClose(ctx, choice.Worktree.RootPane, herdr.WithLogger(logger)); err != nil {
				logTaskCleanup(logger, choice, "tab", err.Error())
			}
		}
	}
}

func logTaskCleanup(logger *slog.Logger, choice *taskChoice, resource, message string) {
	if logger == nil {
		return
	}
	logger.Debug("task spawn cleanup failed", "command", "agent spawn", "task_id", choice.Task.ID,
		"resource", resource, "workspace_id", choice.Worktree.WorkspaceID, "error", message)
}

// spawnTaskAgent runs the T10 core with the resolved pane, then assigns and
// prompts the worker or prompts the verifier.
func spawnTaskAgent(ctx context.Context, home string, options SpawnOptions, choice *taskChoice, logger *slog.Logger) output.Envelope {
	inner := options
	inner.Project = choice.Task.Project
	inner.PaneID = choice.PaneID
	inner.PaneSet = true
	envelope := Spawn(ctx, home, inner)
	if envelope.Outcome == "failed" {
		choice.cleanup(ctx, home, logger)
	}
	record := spawnResultAgent(envelope)
	if record.Key == "" {
		return envelope
	}
	if options.Role == "worker" {
		return finishWorkerSpawn(ctx, home, choice, record, envelope, logger)
	}
	return finishVerifierSpawn(ctx, home, choice, record, envelope, logger)
}

// finishWorkerSpawn assigns the task and sends the worker prompt after a
// settled start.
func finishWorkerSpawn(ctx context.Context, home string, choice *taskChoice, record records.Agent, envelope output.Envelope, logger *slog.Logger) output.Envelope {
	if envelope.Outcome == "ok" {
		if message := assignTask(home, choice.Task.ID, record.Name); message != "" {
			envelope.Outcome = "partial"
			envelope.Phase = "assign"
			envelope.Error = message
			return envelope
		}
		envelope.Effects = append(envelope.Effects, output.Effect{Resource: "task", Path: statePath("tasks", choice.Task.ID), Action: "assigned"})
		return promptTaskAgent(ctx, home, choice, record, envelope, logger)
	}
	if envelope.Outcome == "partial" && isBlockedEnvelope(envelope) {
		hint := "hive task assign --task " + choice.Task.ID + " --agent " + record.Name
		hint += " and hive agent send --agent " + record.Name + " " + shellQuote(workerPrompt(home, choice.Task.ID, choice.Worktree.Path, choice.Branch))
		envelope.Error = appendHint(envelope.Error, hint)
	}
	return envelope
}

// finishVerifierSpawn sends the verifier prompt after a settled start.
func finishVerifierSpawn(ctx context.Context, home string, choice *taskChoice, record records.Agent, envelope output.Envelope, logger *slog.Logger) output.Envelope {
	if envelope.Outcome == "ok" {
		return promptTaskAgent(ctx, home, choice, record, envelope, logger)
	}
	if envelope.Outcome == "partial" && isBlockedEnvelope(envelope) {
		hint := "hive agent send --agent " + record.Name + " " + shellQuote(verifierPrompt(home, choice.Task.ID))
		envelope.Error = appendHint(envelope.Error, hint)
	}
	return envelope
}

// promptTaskAgent sends the first prompt of one task agent without a wait. A
// launch that is still pending gives agent_not_ready. The function waits for
// the interactive agent and sends the prompt again.
func promptTaskAgent(ctx context.Context, home string, choice *taskChoice, record records.Agent, envelope output.Envelope, logger *slog.Logger) output.Envelope {
	prompt := verifierPrompt(home, choice.Task.ID)
	if record.Role != "verifier" {
		prompt = workerPrompt(home, choice.Task.ID, choice.Worktree.Path, choice.Branch)
	}
	params := herdr.AgentPromptParams{Target: choice.PaneID, Text: prompt}
	_, err := herdr.AgentPrompt(ctx, params, herdr.WithLogger(logger))
	if err != nil && isHerdrCode(err, "agent_not_ready") {
		if waitErr := waitAgentInteractive(ctx, choice.PaneID, logger); waitErr == nil {
			_, err = herdr.AgentPrompt(ctx, params, herdr.WithLogger(logger))
		}
	}
	if err != nil {
		envelope.Outcome = "partial"
		envelope.Phase = "prompt"
		envelope.Error = err.Error() + "; then run: hive agent send --agent " + record.Name + " " + shellQuote(prompt)
		return envelope
	}
	envelope.Effects = append(envelope.Effects, output.Effect{Resource: "agent", Path: statePath("agents", record.Key), Action: "prompted"})
	return envelope
}

// promptReadyTimeout bounds the wait for one interactive agent.
const promptReadyTimeout = 60 * time.Second

// waitAgentInteractive waits until herdr reports an interactive agent.
func waitAgentInteractive(ctx context.Context, paneID string, logger *slog.Logger) error {
	deadline := time.Now().Add(promptReadyTimeout)
	for {
		info, err := herdr.AgentGet(ctx, paneID, herdr.WithLogger(logger))
		if err == nil && info.Interactive {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(deadline) {
			return errors.New("agent is not interactive")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}

// assignTask assigns one task under the store lock. A changed status fails with
// the current status.
func assignTask(home, taskID, agentName string) string {
	registry, err := store.Open(home)
	if err != nil {
		return "store: " + err.Error()
	}
	message := ""
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var task records.Task
		if err := locked.ReadIn([]string{"tasks"}, taskID, &task); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				message = "task not found: " + displayValue(taskID)
			} else {
				message = "store: " + err.Error()
			}
			return nil
		}
		if task.Status != "created" && task.Status != "assigned" {
			message = "task is " + task.Status
			return nil
		}
		task.ActingAgent = agentName
		task.Status = "assigned"
		task.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err := locked.WriteIn([]string{"tasks"}, taskID, task); err != nil {
			message = "store: " + err.Error()
			return nil
		}
		return nil
	})
	if err != nil {
		return "store: " + err.Error()
	}
	return message
}

// spawnResultAgent returns the agent record of one spawn envelope.
func spawnResultAgent(envelope output.Envelope) records.Agent {
	if result, ok := envelope.Result.(SpawnResult); ok {
		return result.Agent
	}
	return records.Agent{}
}

// isBlockedEnvelope reports a start that stopped at a trust or login dialog.
func isBlockedEnvelope(envelope output.Envelope) bool {
	if envelope.Error == nil {
		return false
	}
	return strings.HasPrefix(fmt.Sprint(envelope.Error), "agent blocked")
}

// appendHint adds the by-hand command hint to one error.
func appendHint(err any, hint string) string {
	text := fmt.Sprint(err)
	if text == "" {
		return "then run: " + hint
	}
	return text + "; then run: " + hint
}
