package spawn

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// placement is the pane of a new agent and the facts that the record keeps.
// A failed placement keeps the effects that happened before the failure.
// rootPane is true when the agent takes the root pane of a new workspace, so
// its tab still needs the agent name as its label.
type placement struct {
	pane     herdr.PaneInfo
	cwd      string
	worktree string
	rootPane bool
	effects  []output.Effect
}

// placeInCaller opens a tab with label in the workspace of the caller pane.
func placeInCaller(ctx context.Context, callerPane, cwd, label string, env map[string]string, logger *slog.Logger) (placement, string, error) {
	if callerPane == "" {
		return placement{}, "caller", fmt.Errorf("HERDR_PANE_ID is not set: run hive agent spawn in a herdr pane, or pass --project")
	}
	caller, err := herdr.PaneGet(ctx, callerPane, herdr.WithLogger(logger))
	if err != nil {
		return placement{}, "caller", fmt.Errorf("caller pane %s: %w", callerPane, err)
	}
	return openTab(ctx, caller.WorkspaceID, cwd, label, env, logger, nil)
}

// placeInProject opens a tab with label in the checkout workspace of project,
// or takes the root pane of the checkout workspace that it makes.
func placeInProject(ctx context.Context, project records.Project, label string, env map[string]string, logger *slog.Logger) (placement, string, error) {
	workspace, created, effects, err := checkoutWorkspace(ctx, project, logger)
	if err != nil {
		return placement{}, "workspace", err
	}
	if created != nil {
		return placement{pane: *created, cwd: project.Path, rootPane: true, effects: effects}, "", nil
	}
	return openTab(ctx, workspace, project.Path, label, env, logger, effects)
}

// placeInWorktree makes a worktree of project from its checkout workspace and
// takes the root pane of the new worktree workspace.
func placeInWorktree(ctx context.Context, home string, project records.Project, name string, logger *slog.Logger) (placement, string, error) {
	workspace, _, effects, err := checkoutWorkspace(ctx, project, logger)
	if err != nil {
		return placement{}, "workspace", err
	}
	path := filepath.Join(home, "projects", project.ID, "worktrees", name)
	created, err := herdr.WorktreeCreate(ctx, herdr.WorktreeCreateParams{
		WorkspaceID: workspace,
		Branch:      name,
		Base:        project.Base,
		Path:        path,
	}, herdr.WithLogger(logger))
	if err != nil {
		return placement{effects: effects}, "worktree", fmt.Errorf("worktree create %s: %w", path, err)
	}
	effects = append(effects, output.Effect{Resource: "worktree", Path: path, Action: "created"})
	return placement{pane: created.RootPane, cwd: path, worktree: path, rootPane: true, effects: effects}, "", nil
}

// checkoutWorkspace returns the ID of the workspace that herdr has open on the
// registered checkout of project. herdr worktree.list gives it as the
// open_workspace_id of the checkout entry that is not a linked worktree. The
// workspace.list metadata is absent for a workspace that workspace.create
// made (herdr 0.9.5), so it is not used. When none is open, checkoutWorkspace
// makes one with the herdr default label and returns its root pane too.
func checkoutWorkspace(ctx context.Context, project records.Project, logger *slog.Logger) (string, *herdr.PaneInfo, []output.Effect, error) {
	listed, err := herdr.WorktreeList(ctx, herdr.WorktreeListParams{Cwd: project.Path}, herdr.WithLogger(logger))
	if err != nil {
		return "", nil, nil, fmt.Errorf("worktree list %s: %w", project.Path, err)
	}
	for _, tree := range listed.Worktrees {
		if !tree.IsLinkedWorktree && tree.OpenWorkspaceID != "" && samePath(tree.Path, project.Path) {
			return tree.OpenWorkspaceID, nil, nil, nil
		}
	}
	created, err := herdr.WorkspaceCreate(ctx, herdr.WorkspaceCreateParams{Cwd: project.Path}, herdr.WithLogger(logger))
	if err != nil {
		return "", nil, nil, fmt.Errorf("workspace create %s: %w", project.Path, err)
	}
	effects := []output.Effect{{Resource: "workspace", Path: created.Workspace.WorkspaceID, Action: "created"}}
	return created.Workspace.WorkspaceID, &created.RootPane, effects, nil
}

func openTab(ctx context.Context, workspace, cwd, label string, env map[string]string, logger *slog.Logger, effects []output.Effect) (placement, string, error) {
	created, err := herdr.TabCreate(ctx, herdr.TabCreateParams{WorkspaceID: workspace, Cwd: cwd, Label: label, Env: env}, herdr.WithLogger(logger))
	if err != nil {
		return placement{effects: effects}, "tab", fmt.Errorf("tab create in %s: %w", workspace, err)
	}
	effects = append(effects, output.Effect{Resource: "tab", Path: created.Tab.TabID, Action: "created"})
	return placement{pane: created.RootPane, cwd: cwd, effects: effects}, "", nil
}

func samePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if resolved, err := filepath.EvalSymlinks(a); err == nil {
		a = resolved
	}
	if resolved, err := filepath.EvalSymlinks(b); err == nil {
		b = resolved
	}
	return filepath.Clean(a) == filepath.Clean(b)
}
