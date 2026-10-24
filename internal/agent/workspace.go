package agent

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// projectChoice is the workspace and pane that one project agent spawn
// selected.
type projectChoice struct {
	PaneID     string
	ProjectID  string
	Directory  string
	Workspace  records.ProjectWorkspace
	Created    bool
	CreatedTab bool
}

// resolveProjectPane resolves one pane of a project agent. The agent works in
// projects/P/ of the home, in the hive:P workspace.
func resolveProjectPane(ctx context.Context, home string, options SpawnOptions, logger *slog.Logger) (projectChoice, string) {
	if options.projectRecord == nil {
		return projectChoice{}, "project not found: " + displayValue(options.Project)
	}
	project := *options.projectRecord
	directory := filepath.Join(home, "projects", project.ID)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return projectChoice{}, "create project directory: " + err.Error()
	}
	resolved, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return projectChoice{}, "project directory: " + err.Error()
	}
	directory = resolved

	if options.Adopt != "" {
		workspace, err := herdr.WorkspaceGet(ctx, options.Adopt, herdr.WithLogger(logger))
		if err != nil {
			return projectChoice{}, "workspace get: " + err.Error()
		}
		if workspace.WorkspaceID != options.Adopt {
			return projectChoice{}, "workspace get: returned workspace ID differs"
		}
		return createProjectTab(ctx, home, project, directory, records.ProjectWorkspace{ID: workspace.WorkspaceID}, options, logger)
	}

	var recorded *records.ProjectWorkspace
	if project.Workspace != nil && project.Workspace.ID != "" {
		workspace, err := herdr.WorkspaceGet(ctx, project.Workspace.ID, herdr.WithLogger(logger))
		if err != nil {
			if !isHerdrCode(err, herdr.CodeWorkspaceNotFound) {
				return projectChoice{}, "workspace get: " + err.Error()
			}
			if clearErr := clearProjectWorkspace(home, project.ID); clearErr != nil {
				return projectChoice{}, "clear project workspace: " + clearErr.Error()
			}
			logWorkspaceReset(logger, project.ID, project.Workspace.ID, "workspace not found")
		} else if workspace.WorkspaceID != project.Workspace.ID {
			return projectChoice{}, "workspace get: returned workspace ID differs"
		} else {
			copy := *project.Workspace
			recorded = &copy
		}
	}
	if recorded != nil {
		if recorded.TabID != "" {
			choice, reusable, message := reuseProjectTab(ctx, home, project, directory, *recorded, logger)
			if message != "" {
				return projectChoice{}, message
			}
			if reusable {
				return choice, ""
			}
		}
		if err := clearProjectTab(home, project.ID, *recorded); err != nil {
			return projectChoice{}, "clear project tab: " + err.Error()
		}
		recorded.RootPane = ""
		recorded.TabID = ""
		recorded.TabCreatedByHive = false
		return createProjectTab(ctx, home, project, directory, *recorded, options, logger)
	}
	return createProjectWorkspace(ctx, home, project, directory, logger)
}

// reuseProjectTab reuses the recorded hive tab only when it holds one free shell.
func reuseProjectTab(ctx context.Context, home string, project records.Project, directory string, recorded records.ProjectWorkspace, logger *slog.Logger) (projectChoice, bool, string) {
	tabs, err := herdr.TabList(ctx, recorded.ID, herdr.WithLogger(logger))
	if err != nil {
		return projectChoice{}, false, "tab list: " + err.Error()
	}
	var tab *herdr.TabInfo
	for index := range tabs {
		if tabs[index].TabID == recorded.TabID && tabs[index].WorkspaceID == recorded.ID {
			tab = &tabs[index]
			break
		}
	}
	if tab == nil || tab.PaneCount != 1 {
		return projectChoice{}, false, ""
	}
	panes, err := herdr.PaneList(ctx, recorded.ID, herdr.WithLogger(logger))
	if err != nil {
		return projectChoice{}, false, "pane list: " + err.Error()
	}
	var pane *herdr.PaneInfo
	for index := range panes {
		if panes[index].TabID != recorded.TabID {
			continue
		}
		if pane != nil {
			return projectChoice{}, false, ""
		}
		pane = &panes[index]
	}
	if pane == nil || pane.PaneID == "" || pane.WorkspaceID != recorded.ID || pane.TabID != recorded.TabID || !sameDirectory(paneCwd(*pane), directory) || shellUnavailable(*pane) != "" {
		return projectChoice{}, false, ""
	}
	processInfo, err := herdr.PaneProcessInfoGet(ctx, pane.PaneID, herdr.WithLogger(logger))
	if err != nil || processInfo.PaneID != pane.PaneID || processInfo.ShellPID == nil || processInfo.ForegroundProcessGroupID == nil || *processInfo.ShellPID != *processInfo.ForegroundProcessGroupID {
		return projectChoice{}, false, ""
	}
	if recorded.RootPane != pane.PaneID {
		recorded.RootPane = pane.PaneID
		if err := writeProjectWorkspace(home, project.ID, recorded); err != nil {
			return projectChoice{}, false, "store: " + err.Error()
		}
	}
	return projectChoice{PaneID: pane.PaneID, ProjectID: project.ID, Directory: directory, Workspace: recorded}, true, ""
}

// createProjectTab creates one project agent tab in an existing workspace.
func createProjectTab(ctx context.Context, home string, project records.Project, directory string, workspace records.ProjectWorkspace, options SpawnOptions, logger *slog.Logger) (projectChoice, string) {
	created, err := herdr.TabCreate(ctx, herdr.TabCreateParams{
		WorkspaceID: workspace.ID,
		Cwd:         directory,
		Label:       options.Name,
		Focus:       false,
	}, herdr.WithLogger(logger))
	if err != nil {
		return projectChoice{}, "tab create: " + err.Error()
	}
	if created.Tab.TabID == "" || created.Tab.WorkspaceID != workspace.ID || created.RootPane.PaneID == "" || created.RootPane.WorkspaceID != workspace.ID || created.RootPane.TabID != created.Tab.TabID {
		return projectChoice{}, "tab create: returned topology does not match the request"
	}
	workspace.RootPane = created.RootPane.PaneID
	workspace.TabID = created.Tab.TabID
	workspace.TabCreatedByHive = true
	choice := projectChoice{PaneID: created.RootPane.PaneID, ProjectID: project.ID, Directory: directory, Workspace: workspace, CreatedTab: true}
	if !sameDirectory(paneCwd(created.RootPane), directory) {
		choice.cleanup(ctx, home, logger)
		return projectChoice{}, "workspace cwd mismatch"
	}
	if err := writeProjectWorkspace(home, project.ID, workspace); err != nil {
		choice.cleanup(ctx, home, logger)
		return projectChoice{}, "store: " + err.Error()
	}
	if message := shellUnavailable(created.RootPane); message != "" {
		choice.cleanup(ctx, home, logger)
		return projectChoice{}, message
	}
	return choice, ""
}

// createProjectWorkspace makes the hive:P workspace at the project directory.
func createProjectWorkspace(ctx context.Context, home string, project records.Project, directory string, logger *slog.Logger) (projectChoice, string) {
	created, err := herdr.WorkspaceCreate(ctx, herdr.WorkspaceCreateParams{
		Cwd:   directory,
		Label: "hive:" + project.ID,
		Focus: false,
	}, herdr.WithLogger(logger))
	if err != nil {
		return projectChoice{}, "workspace create: " + err.Error()
	}
	if created.Tab.TabID == "" || created.Tab.WorkspaceID != created.Workspace.WorkspaceID || created.RootPane.PaneID == "" || created.RootPane.WorkspaceID != created.Workspace.WorkspaceID || created.RootPane.TabID != created.Tab.TabID {
		return projectChoice{}, "workspace create: returned topology does not match the request"
	}
	workspace := records.ProjectWorkspace{
		ID:               created.Workspace.WorkspaceID,
		RootPane:         created.RootPane.PaneID,
		CreatedByHive:    true,
		CreatedAt:        time.Now().UTC().Format(time.RFC3339Nano),
		TabID:            created.Tab.TabID,
		TabCreatedByHive: false,
	}
	choice := projectChoice{PaneID: created.RootPane.PaneID, ProjectID: project.ID, Directory: directory, Workspace: workspace, Created: true}
	if !sameDirectory(paneCwd(created.RootPane), directory) {
		choice.cleanup(ctx, home, logger)
		return projectChoice{}, "workspace cwd mismatch"
	}
	if err := writeNewProjectWorkspace(home, project.ID, workspace); err != nil {
		choice.cleanup(ctx, home, logger)
		return projectChoice{}, "store: " + err.Error()
	}
	if message := shellUnavailable(created.RootPane); message != "" {
		choice.cleanup(ctx, home, logger)
		return projectChoice{}, message
	}
	return choice, ""
}
func (choice projectChoice) cleanup(ctx context.Context, home string, logger *slog.Logger) {
	if choice.Created {
		if err := herdr.WorkspaceClose(ctx, choice.Workspace.ID, false, herdr.WithLogger(logger)); err != nil {
			if logger != nil {
				logger.Debug("project workspace cleanup close failed", "command", "agent spawn", "project_id", choice.ProjectID, "workspace_id", choice.Workspace.ID, "error", err.Error())
			}
		} else if logger != nil {
			logger.Debug("project workspace cleanup closed", "command", "agent spawn", "project_id", choice.ProjectID, "workspace_id", choice.Workspace.ID)
		}
		if err := clearProjectWorkspaceIfMatch(home, choice.ProjectID, choice.Workspace.ID); err != nil && logger != nil {
			logger.Debug("project workspace cleanup clear failed", "command", "agent spawn", "project_id", choice.ProjectID, "error", err.Error())
		}
		return
	}
	if !choice.CreatedTab {
		return
	}
	pane, err := herdr.PaneGet(ctx, choice.PaneID, herdr.WithLogger(logger))
	if err == nil && pane.PaneID == choice.PaneID && pane.WorkspaceID == choice.Workspace.ID && pane.TabID == choice.Workspace.TabID {
		if err := herdr.PaneClose(ctx, choice.PaneID, herdr.WithLogger(logger)); err != nil && logger != nil {
			logger.Debug("manager tab cleanup close failed", "command", "agent spawn", "project_id", choice.ProjectID, "workspace_id", choice.Workspace.ID, "tab_id", choice.Workspace.TabID, "pane_id", choice.PaneID, "error", err.Error())
		}
	}
	if err := clearProjectTab(home, choice.ProjectID, choice.Workspace); err != nil && logger != nil {
		logger.Debug("manager tab cleanup clear failed", "command", "agent spawn", "project_id", choice.ProjectID, "workspace_id", choice.Workspace.ID, "tab_id", choice.Workspace.TabID, "error", err.Error())
	}
}

func clearProjectTab(home, projectID string, expected records.ProjectWorkspace) error {
	return mutateProjectWorkspace(home, projectID, func(project *records.Project) {
		if project.Workspace == nil || project.Workspace.ID != expected.ID || project.Workspace.TabID != expected.TabID {
			return
		}
		project.Workspace.RootPane = ""
		project.Workspace.TabID = ""
		project.Workspace.TabCreatedByHive = false
	})
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

// writeNewProjectWorkspace records a workspace only while the project has no
// workspace. A concurrent spawn keeps the workspace that it recorded first.
func writeNewProjectWorkspace(home, projectID string, workspace records.ProjectWorkspace) error {
	return mutateProjectWorkspace(home, projectID, func(project *records.Project) {
		if project.Workspace != nil && project.Workspace.ID != "" {
			return
		}
		copied := workspace
		project.Workspace = &copied
	})
}

// clearProjectWorkspaceIfMatch removes the recorded workspace only when it is
// the workspace of one cleanup.
func clearProjectWorkspaceIfMatch(home, projectID, workspaceID string) error {
	return mutateProjectWorkspace(home, projectID, func(project *records.Project) {
		if project.Workspace == nil || project.Workspace.ID != workspaceID {
			return
		}
		project.Workspace = nil
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
