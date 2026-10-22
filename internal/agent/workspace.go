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

// managerChoice is the workspace and pane that one manager spawn selected.
type managerChoice struct {
	PaneID     string
	ProjectID  string
	Workspace  records.ProjectWorkspace
	Created    bool
	CreatedTab bool
}

// resolveManagerWorkspace resolves one manager pane from the project record or
// the parent workspace that herdr reports for the repository.
func resolveManagerWorkspace(ctx context.Context, home string, options SpawnOptions, logger *slog.Logger) (managerChoice, string) {
	project, message := readProjectRecord(home, options.Project)
	if message != "" {
		return managerChoice{}, message
	}

	var recorded *records.ProjectWorkspace
	if project.Workspace != nil && project.Workspace.ID != "" {
		workspace, err := herdr.WorkspaceGet(ctx, project.Workspace.ID, herdr.WithLogger(logger))
		if err != nil {
			if !isHerdrCode(err, herdr.CodeWorkspaceNotFound) {
				return managerChoice{}, "workspace get: " + err.Error()
			}
			if clearErr := clearProjectWorkspace(home, project.ID); clearErr != nil {
				return managerChoice{}, "clear project workspace: " + clearErr.Error()
			}
			logWorkspaceReset(logger, project.ID, project.Workspace.ID, "workspace not found")
		} else if workspace.WorkspaceID != project.Workspace.ID {
			return managerChoice{}, "workspace get: returned workspace ID differs"
		} else {
			copy := *project.Workspace
			recorded = &copy
		}
	}

	parentID, err := managerParentWorkspace(ctx, project.Path, logger)
	if err != nil {
		return managerChoice{}, "worktree list: " + err.Error()
	}
	if recorded != nil && parentID != recorded.ID {
		if clearErr := clearProjectWorkspace(home, project.ID); clearErr != nil {
			return managerChoice{}, "clear project workspace: " + clearErr.Error()
		}
		logWorkspaceReset(logger, project.ID, recorded.ID, "workspace does not hold parent checkout")
		recorded = nil
	}
	if recorded != nil {
		if recorded.TabID != "" {
			choice, reusable, message := reuseManagerTab(ctx, home, project, *recorded, logger)
			if message != "" {
				return managerChoice{}, message
			}
			if reusable {
				return choice, ""
			}
		}
		if err := clearProjectTab(home, project.ID, *recorded); err != nil {
			return managerChoice{}, "clear project tab: " + err.Error()
		}
		recorded.RootPane = ""
		recorded.TabID = ""
		recorded.TabCreatedByHive = false
		return createManagerTab(ctx, home, project, *recorded, options, logger)
	}

	if options.Adopt != "" {
		if parentID == "" || options.Adopt != parentID {
			return managerChoice{}, fmt.Sprintf("workspace %s does not hold the parent checkout", options.Adopt)
		}
		workspace, err := herdr.WorkspaceGet(ctx, options.Adopt, herdr.WithLogger(logger))
		if err != nil {
			return managerChoice{}, "workspace get: " + err.Error()
		}
		if workspace.WorkspaceID != options.Adopt {
			return managerChoice{}, "workspace get: returned workspace ID differs"
		}
		return createManagerTab(ctx, home, project, records.ProjectWorkspace{ID: workspace.WorkspaceID}, options, logger)
	}
	if parentID != "" {
		workspace, err := herdr.WorkspaceGet(ctx, parentID, herdr.WithLogger(logger))
		if err != nil {
			return managerChoice{}, "workspace get: " + err.Error()
		}
		if workspace.WorkspaceID != parentID {
			return managerChoice{}, "workspace get: returned workspace ID differs"
		}
		return createManagerTab(ctx, home, project, records.ProjectWorkspace{ID: workspace.WorkspaceID}, options, logger)
	}
	if options.Adopt != "" {
		return managerChoice{}, fmt.Sprintf("workspace %s does not hold the parent checkout", options.Adopt)
	}
	return createManagerWorkspace(ctx, home, project, logger)
}

// managerParentWorkspace gets the workspace that holds the repository's parent
// checkout. An empty ID means that no workspace holds it.
func managerParentWorkspace(ctx context.Context, repositoryPath string, logger *slog.Logger) (string, error) {
	listing, err := herdr.WorktreeList(ctx, herdr.WorktreeListParams{Cwd: repositoryPath}, herdr.WithLogger(logger))
	if err != nil {
		return "", err
	}
	return listing.Source.SourceWorkspaceID, nil
}

// reuseManagerTab reuses the recorded hive tab only when it holds one free shell.
func reuseManagerTab(ctx context.Context, home string, project records.Project, recorded records.ProjectWorkspace, logger *slog.Logger) (managerChoice, bool, string) {
	tabs, err := herdr.TabList(ctx, recorded.ID, herdr.WithLogger(logger))
	if err != nil {
		return managerChoice{}, false, "tab list: " + err.Error()
	}
	var tab *herdr.TabInfo
	for index := range tabs {
		if tabs[index].TabID == recorded.TabID && tabs[index].WorkspaceID == recorded.ID {
			tab = &tabs[index]
			break
		}
	}
	if tab == nil || tab.PaneCount != 1 {
		return managerChoice{}, false, ""
	}
	panes, err := herdr.PaneList(ctx, recorded.ID, herdr.WithLogger(logger))
	if err != nil {
		return managerChoice{}, false, "pane list: " + err.Error()
	}
	var pane *herdr.PaneInfo
	for index := range panes {
		if panes[index].TabID != recorded.TabID {
			continue
		}
		if pane != nil {
			return managerChoice{}, false, ""
		}
		pane = &panes[index]
	}
	if pane == nil || pane.PaneID == "" || pane.WorkspaceID != recorded.ID || pane.TabID != recorded.TabID || !sameDirectory(paneCwd(*pane), project.Path) || shellUnavailable(*pane) != "" {
		return managerChoice{}, false, ""
	}
	processInfo, err := herdr.PaneProcessInfoGet(ctx, pane.PaneID, herdr.WithLogger(logger))
	if err != nil || processInfo.PaneID != pane.PaneID || processInfo.ShellPID == nil || processInfo.ForegroundProcessGroupID == nil || *processInfo.ShellPID != *processInfo.ForegroundProcessGroupID {
		return managerChoice{}, false, ""
	}
	if recorded.RootPane != pane.PaneID {
		recorded.RootPane = pane.PaneID
		if err := writeProjectWorkspace(home, project.ID, recorded); err != nil {
			return managerChoice{}, false, "store: " + err.Error()
		}
	}
	return managerChoice{PaneID: pane.PaneID, ProjectID: project.ID, Workspace: recorded}, true, ""
}

// createManagerTab creates one manager tab in the workspace that holds the
// repository's parent checkout.
func createManagerTab(ctx context.Context, home string, project records.Project, workspace records.ProjectWorkspace, options SpawnOptions, logger *slog.Logger) (managerChoice, string) {
	created, err := herdr.TabCreate(ctx, herdr.TabCreateParams{
		WorkspaceID: workspace.ID,
		Cwd:         project.Path,
		Label:       managerAgentName(options),
		Focus:       false,
	}, herdr.WithLogger(logger))
	if err != nil {
		return managerChoice{}, "tab create: " + err.Error()
	}
	if created.Tab.TabID == "" || created.Tab.WorkspaceID != workspace.ID || created.RootPane.PaneID == "" || created.RootPane.WorkspaceID != workspace.ID || created.RootPane.TabID != created.Tab.TabID {
		return managerChoice{}, "tab create: returned topology does not match the request"
	}
	workspace.RootPane = created.RootPane.PaneID
	workspace.TabID = created.Tab.TabID
	workspace.TabCreatedByHive = true
	choice := managerChoice{PaneID: created.RootPane.PaneID, ProjectID: project.ID, Workspace: workspace, CreatedTab: true}
	if !sameDirectory(paneCwd(created.RootPane), project.Path) {
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

// createManagerWorkspace makes a workspace at the repository path.
func createManagerWorkspace(ctx context.Context, home string, project records.Project, logger *slog.Logger) (managerChoice, string) {
	created, err := herdr.WorkspaceCreate(ctx, herdr.WorkspaceCreateParams{
		Cwd:   project.Path,
		Label: "hive:" + project.ID,
		Focus: false,
	}, herdr.WithLogger(logger))
	if err != nil {
		return managerChoice{}, "workspace create: " + err.Error()
	}
	if created.Tab.TabID == "" || created.Tab.WorkspaceID != created.Workspace.WorkspaceID || created.RootPane.PaneID == "" || created.RootPane.WorkspaceID != created.Workspace.WorkspaceID || created.RootPane.TabID != created.Tab.TabID {
		return managerChoice{}, "workspace create: returned topology does not match the request"
	}
	workspace := records.ProjectWorkspace{
		ID:               created.Workspace.WorkspaceID,
		RootPane:         created.RootPane.PaneID,
		CreatedByHive:    true,
		CreatedAt:        time.Now().UTC().Format(time.RFC3339Nano),
		TabID:            created.Tab.TabID,
		TabCreatedByHive: false,
	}
	choice := managerChoice{PaneID: created.RootPane.PaneID, ProjectID: project.ID, Workspace: workspace, Created: true}
	if !sameDirectory(paneCwd(created.RootPane), project.Path) {
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

func managerAgentName(options SpawnOptions) string {
	if options.NameSet {
		return options.Name
	}
	return defaultSingletonName("manager", options.Project)
}

// cleanup closes only topology that this spawn made.
func (choice managerChoice) cleanup(ctx context.Context, home string, logger *slog.Logger) {
	if choice.Created {
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
