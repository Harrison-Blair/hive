package spawn

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/worktree"
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
		return placement{}, "caller", fmt.Errorf("HERDR_PANE_ID is not set: run hive agent spawn in a herdr pane, or pass --branch or --directory")
	}
	caller, err := herdr.PaneGet(ctx, callerPane, herdr.WithLogger(logger))
	if err != nil {
		return placement{}, "caller", fmt.Errorf("caller pane %s: %w", callerPane, err)
	}
	return openTab(ctx, caller.WorkspaceID, cwd, label, env, logger, nil)
}

// placeInBranch finds or makes the checkout of options.Branch in the Git
// repository of options.Cwd, and places the agent in it.
func placeInBranch(ctx context.Context, home string, options Options, env map[string]string, logger *slog.Logger) (placement, string, error) {
	checkout, err := worktree.Ensure(ctx, home, options.Cwd, options.Branch, options.Base, logger)
	effects := worktree.Effects(checkout)
	if err != nil {
		return placement{effects: effects}, "worktree", err
	}
	place, phase, err := placeInDirectory(ctx, home, checkout.Path, options.Name, env, logger, effects)
	place.worktree = checkout.Path
	return place, phase, err
}

// placeInDirectory places the agent in dir. It opens a tab in the herdr
// workspace of dir: the workspace that herdr has open on the checkout at dir,
// or else the workspace of a pane whose shell is in dir. When dir has no
// workspace, it takes the root pane of a new workspace on dir.
func placeInDirectory(ctx context.Context, home, dir, label string, env map[string]string, logger *slog.Logger, effects []output.Effect) (placement, string, error) {
	workspace, err := workspaceOf(ctx, dir, logger)
	if err != nil {
		return placement{effects: effects}, "workspace", err
	}
	if workspace != "" {
		return openTab(ctx, workspace, dir, label, env, logger, effects)
	}
	created, err := herdr.WorkspaceCreate(ctx, herdr.WorkspaceCreateParams{Cwd: dir, Env: env}, herdr.WithLogger(logger))
	if err != nil {
		return placement{effects: effects}, "workspace", fmt.Errorf("workspace create %s: %w", dir, err)
	}
	effects = append(effects, output.Effect{Resource: "workspace", Path: created.Workspace.WorkspaceID, Action: "created"})
	return placement{pane: created.RootPane, cwd: dir, rootPane: true, effects: effects}, "", nil
}

// workspaceOf returns the ID of the workspace of dir, or an empty string.
// herdr worktree.list gives the workspace open on a checkout as the
// open_workspace_id of its entry, also for a workspace that workspace.create
// made (herdr 0.9.5). A directory outside Git, or below the root of a
// checkout, has no entry. Then the workspace of the first pane whose shell is
// in dir is the workspace of dir: the root pane of a workspace on dir, or the
// pane of an agent in dir.
func workspaceOf(ctx context.Context, dir string, logger *slog.Logger) (string, error) {
	listed, err := herdr.WorktreeList(ctx, herdr.WorktreeListParams{Cwd: dir}, herdr.WithLogger(logger))
	switch {
	case agent.IsCode(err, herdr.CodeNotGitWorktree):
	case err != nil:
		return "", fmt.Errorf("worktree list %s: %w", dir, err)
	default:
		for _, tree := range listed.Worktrees {
			if tree.OpenWorkspaceID != "" && worktree.SamePath(tree.Path, dir) {
				return tree.OpenWorkspaceID, nil
			}
		}
	}
	panes, err := herdr.PaneList(ctx, "", herdr.WithLogger(logger))
	if err != nil {
		return "", fmt.Errorf("pane list: %w", err)
	}
	for _, pane := range panes {
		if worktree.SamePath(pane.Cwd, dir) {
			return pane.WorkspaceID, nil
		}
	}
	return "", nil
}

func openTab(ctx context.Context, workspace, cwd, label string, env map[string]string, logger *slog.Logger, effects []output.Effect) (placement, string, error) {
	created, err := herdr.TabCreate(ctx, herdr.TabCreateParams{WorkspaceID: workspace, Cwd: cwd, Label: label, Env: env}, herdr.WithLogger(logger))
	if err != nil {
		return placement{effects: effects}, "tab", fmt.Errorf("tab create in %s: %w", workspace, err)
	}
	effects = append(effects, output.Effect{Resource: "tab", Path: created.Tab.TabID, Action: "created"})
	return placement{pane: created.RootPane, cwd: cwd, effects: effects}, "", nil
}
