package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// managerChoice is the manager workspace and pane that one spawn selected.
type managerChoice struct {
	PaneID    string
	ProjectID string
	Workspace records.ProjectWorkspace
	Created   bool
}

// resolveManagerWorkspace reuses, adopts, or makes the workspace of one
// manager. It returns the chosen pane and an error message.
func resolveManagerWorkspace(ctx context.Context, home string, options SpawnOptions, logger *slog.Logger) (managerChoice, string) {
	project, message := readProjectRecord(home, options.Project)
	if message != "" {
		return managerChoice{}, message
	}
	projectDir := filepath.Join(home, "projects", project.ID)
	label := "hive:" + project.ID
	if recorded := project.Workspace; recorded != nil && recorded.ID != "" {
		choice, reuseMessage, cleared := reuseManagerWorkspace(ctx, home, projectDir, label, project, *recorded, logger)
		if !cleared {
			return choice, reuseMessage
		}
	}
	if options.Adopt != "" {
		return adoptManagerWorkspace(ctx, home, projectDir, label, project, options.Adopt, logger)
	}
	workspaces, err := herdr.WorkspaceList(ctx, herdr.WithLogger(logger))
	if err != nil {
		return managerChoice{}, "workspace list: " + err.Error()
	}
	for _, workspace := range workspaces {
		if workspace.Label == label {
			return managerChoice{}, fmt.Sprintf("workspace %s has the label %s but hive did not make it: close it or pass --adopt %s", workspace.WorkspaceID, label, workspace.WorkspaceID)
		}
	}
	return createManagerWorkspace(ctx, home, projectDir, label, project, logger)
}

// reuseManagerWorkspace checks one recorded workspace. When cleared is true,
// the record no longer names a workspace and the caller continues without one.
func reuseManagerWorkspace(ctx context.Context, home, projectDir, label string, project records.Project, recorded records.ProjectWorkspace, logger *slog.Logger) (managerChoice, string, bool) {
	workspace, err := herdr.WorkspaceGet(ctx, recorded.ID, herdr.WithLogger(logger))
	if err != nil {
		if !isHerdrCode(err, herdr.CodeWorkspaceNotFound) {
			return managerChoice{}, "workspace get: " + err.Error(), false
		}
		if clearErr := clearProjectWorkspace(home, project.ID); clearErr != nil {
			return managerChoice{}, "clear project workspace: " + clearErr.Error(), false
		}
		logWorkspaceReset(logger, project.ID, recorded.ID, "workspace not found")
		return managerChoice{}, "", true
	}
	if workspace.Label != label {
		if clearErr := clearProjectWorkspace(home, project.ID); clearErr != nil {
			return managerChoice{}, "clear project workspace: " + clearErr.Error(), false
		}
		logWorkspaceReset(logger, project.ID, recorded.ID, "label differs")
		return managerChoice{}, "", true
	}
	if recorded.RootPane == "" {
		return managerChoice{}, recordedRootRefusal(recorded.ID), false
	}
	pane, paneErr := herdr.PaneGet(ctx, recorded.RootPane, herdr.WithLogger(logger))
	if paneErr != nil {
		return managerChoice{}, recordedRootRefusal(recorded.ID), false
	}
	if pane.PaneID != recorded.RootPane {
		return managerChoice{}, recordedRootRefusal(recorded.ID), false
	}
	if !sameDirectory(paneCwd(pane), projectDir) {
		if clearErr := clearProjectWorkspace(home, project.ID); clearErr != nil {
			return managerChoice{}, "clear project workspace: " + clearErr.Error(), false
		}
		logWorkspaceReset(logger, project.ID, recorded.ID, "directory differs")
		return managerChoice{}, "", true
	}
	tabs, tabErr := herdr.TabList(ctx, recorded.ID, herdr.WithLogger(logger))
	if tabErr != nil {
		return managerChoice{}, "tab list: " + tabErr.Error(), false
	}
	if len(tabs) == 0 || tabs[0].TabID != pane.TabID {
		return managerChoice{}, recordedRootRefusal(recorded.ID), false
	}
	if message := shellUnavailable(pane); message != "" {
		return managerChoice{}, message, false
	}
	return managerChoice{PaneID: pane.PaneID, ProjectID: project.ID, Workspace: recorded}, "", false
}

// adoptManagerWorkspace records a workspace that hive did not make.
func adoptManagerWorkspace(ctx context.Context, home, projectDir, label string, project records.Project, adoptID string, logger *slog.Logger) (managerChoice, string) {
	workspace, err := herdr.WorkspaceGet(ctx, adoptID, herdr.WithLogger(logger))
	if err != nil {
		if isHerdrCode(err, herdr.CodeWorkspaceNotFound) {
			return managerChoice{}, fmt.Sprintf("workspace %s not found: pass --adopt with an open workspace", adoptID)
		}
		return managerChoice{}, "workspace get: " + err.Error()
	}
	if workspace.Label != label {
		return managerChoice{}, fmt.Sprintf("workspace %s has the label %s: want %s", adoptID, workspace.Label, label)
	}
	tabs, err := herdr.TabList(ctx, adoptID, herdr.WithLogger(logger))
	if err != nil {
		return managerChoice{}, "tab list: " + err.Error()
	}
	if len(tabs) == 0 {
		return managerChoice{}, "first tab has 0 panes: pass --pane"
	}
	first := tabs[0]
	panes, err := herdr.PaneList(ctx, adoptID, herdr.WithLogger(logger))
	if err != nil {
		return managerChoice{}, "pane list: " + err.Error()
	}
	count := 0
	var pane herdr.PaneInfo
	for _, candidate := range panes {
		if candidate.TabID != first.TabID {
			continue
		}
		count++
		pane = candidate
	}
	if count != 1 {
		return managerChoice{}, fmt.Sprintf("first tab has %d panes: pass --pane", count)
	}
	if !sameDirectory(paneCwd(pane), projectDir) {
		return managerChoice{}, "workspace cwd mismatch"
	}
	if message := shellUnavailable(pane); message != "" {
		return managerChoice{}, message
	}
	record := records.ProjectWorkspace{ID: adoptID, RootPane: pane.PaneID, CreatedByHive: false}
	if err := writeProjectWorkspace(home, project.ID, record); err != nil {
		return managerChoice{}, "store: " + err.Error()
	}
	return managerChoice{PaneID: pane.PaneID, ProjectID: project.ID, Workspace: record}, ""
}

// createManagerWorkspace makes the project directory and its workspace.
func createManagerWorkspace(ctx context.Context, home, projectDir, label string, project records.Project, logger *slog.Logger) (managerChoice, string) {
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return managerChoice{}, "create project directory: " + err.Error()
	}
	created, err := herdr.WorkspaceCreate(ctx, herdr.WorkspaceCreateParams{Cwd: projectDir, Label: label, Focus: false}, herdr.WithLogger(logger))
	if err != nil {
		return managerChoice{}, "workspace create: " + err.Error()
	}
	workspace := records.ProjectWorkspace{
		ID:            created.Workspace.WorkspaceID,
		RootPane:      created.RootPane.PaneID,
		CreatedByHive: true,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	choice := managerChoice{PaneID: created.RootPane.PaneID, ProjectID: project.ID, Workspace: workspace, Created: true}
	if !sameDirectory(paneCwd(created.RootPane), projectDir) {
		choice.cleanup(ctx, home, logger)
		return managerChoice{}, "workspace cwd mismatch"
	}
	if err := writeProjectWorkspace(home, project.ID, workspace); err != nil {
		choice.cleanup(ctx, home, logger)
		return managerChoice{}, "store: " + err.Error()
	}
	if message := shellUnavailable(created.RootPane); message != "" {
		choice.cleanup(ctx, home, logger)
		return managerChoice{}, message
	}
	return choice, ""
}

// cleanup closes a workspace that this spawn made and clears the record.
func (choice managerChoice) cleanup(ctx context.Context, home string, logger *slog.Logger) {
	if !choice.Created {
		return
	}
	if err := herdr.WorkspaceClose(ctx, choice.Workspace.ID, false, herdr.WithLogger(logger)); err != nil {
		if logger != nil {
			logger.Debug("manager workspace cleanup close failed", "command", "agent spawn", "project_id", choice.ProjectID, "workspace_id", choice.Workspace.ID, "error", err.Error())
		}
	} else if logger != nil {
		logger.Debug("manager workspace cleanup closed", "command", "agent spawn", "project_id", choice.ProjectID, "workspace_id", choice.Workspace.ID)
	}
	if err := clearProjectWorkspace(home, choice.ProjectID); err != nil && logger != nil {
		logger.Debug("manager workspace cleanup clear failed", "command", "agent spawn", "project_id", choice.ProjectID, "error", err.Error())
	}
}

func recordedRootRefusal(workspaceID string) string {
	return fmt.Sprintf("workspace %s has no recorded root pane: pass --pane <shell pane>", workspaceID)
}

// shellUnavailable reports whether one pane cannot hold the manager.
func shellUnavailable(pane herdr.PaneInfo) string {
	if pane.RestoreError != nil || pane.AgentStatus != "unknown" || pane.Agent != "" {
		return "first tab busy: stop the agent or pass --pane"
	}
	return ""
}

func paneCwd(pane herdr.PaneInfo) string {
	if pane.Cwd != "" {
		return pane.Cwd
	}
	return pane.ForegroundCwd
}

func sameDirectory(a, b string) bool {
	resolvedA, errA := filepath.EvalSymlinks(a)
	if errA != nil {
		resolvedA = filepath.Clean(a)
	}
	resolvedB, errB := filepath.EvalSymlinks(b)
	if errB != nil {
		resolvedB = filepath.Clean(b)
	}
	return resolvedA == resolvedB
}

func readProjectRecord(home, projectID string) (records.Project, string) {
	registry, err := store.Open(home)
	if err != nil {
		return records.Project{}, "store: " + err.Error()
	}
	var project records.Project
	if err := registry.ReadIn([]string{"projects"}, projectID, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return records.Project{}, "project not found: " + projectID
		}
		return records.Project{}, "store: " + err.Error()
	}
	return project, ""
}

func writeProjectWorkspace(home, projectID string, workspace records.ProjectWorkspace) error {
	return mutateProjectWorkspace(home, projectID, func(project *records.Project) {
		copied := workspace
		project.Workspace = &copied
	})
}

func clearProjectWorkspace(home, projectID string) error {
	return mutateProjectWorkspace(home, projectID, func(project *records.Project) {
		project.Workspace = nil
	})
}

func mutateProjectWorkspace(home, projectID string, apply func(*records.Project)) error {
	registry, err := store.Open(home)
	if err != nil {
		return err
	}
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var project records.Project
		if err := locked.ReadIn([]string{"projects"}, projectID, &project); err != nil {
			return err
		}
		apply(&project)
		return locked.WriteIn([]string{"projects"}, projectID, project)
	})
}

func isHerdrCode(err error, code string) bool {
	var serverError *herdr.Error
	return errors.As(err, &serverError) && serverError.Code == code
}

func logWorkspaceReset(logger *slog.Logger, projectID, workspaceID, reason string) {
	if logger == nil {
		return
	}
	logger.Debug("manager workspace record cleared", "command", "agent spawn", "project_id", projectID, "workspace_id", workspaceID, "reason", reason)
}
