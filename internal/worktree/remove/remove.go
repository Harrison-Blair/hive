// Package remove removes the checkout of a branch that hive made, when it is
// clean and idle. The branch stays.
package remove

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/worktree"
)

// Options holds the values of one remove.
type Options struct {
	// Dir selects the Git repository: the working directory of the caller.
	Dir    string
	Branch string
	// CallerPane is the herdr pane of the caller (HERDR_PANE_ID), or empty.
	CallerPane string
	Logger     *slog.Logger
}

// Run removes the checkout of the branch. It refuses a checkout that hive did
// not make, a checkout with any change, untracked file, or ignored file, and
// a checkout that a live agent uses. It closes the herdr workspace open on
// the checkout, which holds only shells, removes the checkout, and removes
// the ownership record. For a branch without a checkout, it removes the
// ownership record that an earlier remove left.
func Run(ctx context.Context, home string, options Options) output.Envelope {
	logger := agent.Logger(options.Logger)
	// The layout lock covers the whole remove, so no spawn places an agent
	// in the checkout or its workspace between the checks and the removal.
	var envelope output.Envelope
	if err := agent.Hold(home, func() error {
		envelope = run(ctx, home, options, logger)
		return nil
	}); err != nil {
		envelope = output.Failed("lock", err.Error())
	}
	attrs := []any{"command", "worktree remove", "branch", options.Branch, "outcome", envelope.Outcome, "phase", envelope.Phase}
	if envelope.Error != nil {
		attrs = append(attrs, "error", fmt.Sprint(envelope.Error))
	}
	logger.Debug("worktree remove", attrs...)
	return envelope
}

func run(ctx context.Context, home string, options Options, logger *slog.Logger) output.Envelope {
	if options.Branch == "" {
		return output.Failed("validate", "BRANCH is required")
	}
	checkout, found, err := worktree.Lookup(ctx, home, options.Dir, options.Branch, logger)
	if err != nil {
		return output.Failed("worktree", err.Error())
	}
	if !found {
		return forgetStale(ctx, home, options, logger)
	}
	if !checkout.Owned {
		return output.Failed("owner", fmt.Sprintf("checkout %s of branch %s was not made by hive: hive removes only its own checkouts", checkout.Path, checkout.Branch))
	}
	if message := checkClean(ctx, checkout.Path); message != "" {
		return output.Failed("dirty", message+"; nothing was removed")
	}
	if message, err := liveAgents(ctx, home, checkout.Path, logger); err != nil {
		return output.Failed("agents", err.Error())
	} else if message != "" {
		return output.Failed("busy", message)
	}
	workspaces, message, err := checkoutWorkspaces(ctx, checkout.Path, logger)
	if err != nil {
		return output.Failed("workspace", err.Error())
	}
	if message != "" {
		return output.Failed("busy", message)
	}
	if options.CallerPane != "" && len(workspaces) > 0 {
		caller, err := herdr.PaneGet(ctx, options.CallerPane, herdr.WithLogger(logger))
		if err != nil {
			return output.Failed("caller", fmt.Sprintf("caller pane %s: %s", options.CallerPane, err))
		}
		// The close would end the shell of the caller and this remove.
		if slices.Contains(workspaces, caller.WorkspaceID) {
			return output.Failed("caller", fmt.Sprintf("remove closes workspace %s, which holds the caller pane %s: run it from a pane in another workspace; nothing was removed", caller.WorkspaceID, options.CallerPane))
		}
	}

	// The workspaces close first. A failure after a close keeps the
	// checkout, so a second remove finds it and finishes. A checkout removed
	// first would leave a workspace that hive can no longer find.
	effects := []output.Effect{}
	for _, workspace := range workspaces {
		if err := herdr.WorkspaceClose(ctx, workspace, false, herdr.WithLogger(logger)); err != nil {
			envelope := output.Failed("close", fmt.Sprintf("close workspace %s: %s", workspace, err)+notRemoved(effects))
			envelope.Effects = effects
			return envelope
		}
		effects = append(effects, output.Effect{Resource: "workspace", Path: workspace, Action: "closed"})
	}
	// git worktree remove deletes ignored files, so the check runs again
	// just before it. A process outside hive that writes into the checkout
	// after this check and before git ends can still lose a file: hive
	// expects such writers to be stopped.
	if message := checkClean(ctx, checkout.Path); message != "" {
		envelope := output.Failed("dirty", message+notRemoved(effects))
		envelope.Effects = effects
		return envelope
	}
	out, err := exec.CommandContext(ctx, "git", "-C", options.Dir, "worktree", "remove", "--", checkout.Path).CombinedOutput()
	logger.Debug("git", "component", "git", "args", []string{"worktree", "remove", checkout.Path}, "cwd", options.Dir)
	if err != nil {
		envelope := output.Failed("remove", fmt.Sprintf("git worktree remove %s: %s: %s", checkout.Path, err, strings.TrimSpace(string(out)))+notRemoved(effects))
		envelope.Effects = effects
		return envelope
	}
	effects = append(effects, output.Effect{Resource: "worktree", Path: checkout.Path, Action: "removed"})
	if err := worktree.Forget(home, checkout); err != nil {
		message := fmt.Sprintf("remove ownership record: %s; run hive worktree remove %s again to remove it", err, checkout.Branch)
		return output.Envelope{Outcome: output.OutcomePartial, Phase: "record", Effects: effects, Error: message, Result: checkout}
	}
	effects = append(effects, recordEffect(checkout))
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: effects, Error: nil, Result: checkout}
}

// notRemoved ends the error of a remove that stops before git removed the
// checkout: it names the workspaces in effects that remove closed.
func notRemoved(effects []output.Effect) string {
	var closed []string
	for _, effect := range effects {
		if effect.Resource == "workspace" && effect.Action == "closed" {
			closed = append(closed, effect.Path)
		}
	}
	if len(closed) == 0 {
		return "; nothing was removed"
	}
	return "; the checkout was not removed; closed workspaces " + strings.Join(closed, ", ")
}

// forgetStale removes the ownership record that an earlier remove left after
// it removed the checkout. A branch without a checkout and without such a
// record fails.
func forgetStale(ctx context.Context, home string, options Options, logger *slog.Logger) output.Envelope {
	checkout, found, err := worktree.Stale(ctx, home, options.Dir, options.Branch, logger)
	if err != nil {
		return output.Failed("worktree", err.Error())
	}
	if !found {
		return output.Failed("worktree", fmt.Sprintf("branch %s has no checkout", options.Branch))
	}
	if err := worktree.Forget(home, checkout); err != nil {
		return output.Failed("record", "remove ownership record: "+err.Error())
	}
	return output.Envelope{Outcome: output.OutcomeOK, Phase: "complete", Effects: []output.Effect{recordEffect(checkout)}, Error: nil, Result: checkout}
}

func recordEffect(checkout worktree.Checkout) output.Effect {
	return output.Effect{Resource: "record", Path: filepath.Join(worktree.Dir, filepath.Base(filepath.Dir(checkout.Path)), "worktree.json"), Action: "removed"}
}

// checkClean returns a message when the checkout at path has a change, an
// untracked file, or an ignored file, or when Git cannot read it.
func checkClean(ctx context.Context, path string) string {
	out, err := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain", "--ignored").CombinedOutput()
	if err != nil {
		return fmt.Sprintf("git status %s: %s: %s", path, err, strings.TrimSpace(string(out)))
	}
	if len(strings.TrimSpace(string(out))) > 0 {
		return fmt.Sprintf("checkout %s is dirty: commit or delete its changes, untracked files, and ignored files first", path)
	}
	return ""
}

// liveAgents returns a message that names each agent record in path whose
// pane herdr still has.
func liveAgents(ctx context.Context, home, path string, logger *slog.Logger) (string, error) {
	records, err := agent.List(home)
	if err != nil {
		return "", err
	}
	var live []string
	for _, record := range records {
		if !within(record.Cwd, path) && !within(record.Worktree, path) {
			continue
		}
		_, err := herdr.PaneGet(ctx, record.Pane, herdr.WithLogger(logger))
		if agent.IsCode(err, herdr.CodePaneNotFound) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("check agent %s: %w", record.Name, err)
		}
		live = append(live, fmt.Sprintf("%s (pane %s)", record.Name, record.Pane))
	}
	if len(live) == 0 {
		return "", nil
	}
	return fmt.Sprintf("checkout %s has live agents %s: stop them first; nothing was removed", path, strings.Join(live, ", ")), nil
}

// checkoutWorkspaces returns the herdr workspaces of the checkout at path,
// which remove closes: the workspace open on the checkout, and each
// workspace with a pane whose shell or foreground process is in the
// checkout. message is not empty when remove
// must stop: a pane of such a workspace runs an agent, or is outside the
// checkout, so the close would stop a process that the checkout does not
// own.
func checkoutWorkspaces(ctx context.Context, path string, logger *slog.Logger) (workspaces []string, message string, err error) {
	open, err := openWorkspace(ctx, path, logger)
	if err != nil {
		return nil, "", err
	}
	panes, err := herdr.PaneList(ctx, "", herdr.WithLogger(logger))
	if err != nil {
		return nil, "", fmt.Errorf("pane list: %w", err)
	}
	if open != "" {
		workspaces = append(workspaces, open)
	}
	for _, pane := range panes {
		if inCheckout(pane, path) && !slices.Contains(workspaces, pane.WorkspaceID) {
			workspaces = append(workspaces, pane.WorkspaceID)
		}
	}
	for _, workspace := range workspaces {
		for _, pane := range panes {
			if pane.WorkspaceID != workspace {
				continue
			}
			if pane.Agent != "" {
				return nil, fmt.Sprintf("pane %s of workspace %s runs agent %s: stop it first; nothing was removed", pane.PaneID, workspace, pane.Agent), nil
			}
			if !within(pane.Cwd, path) || pane.ForegroundCwd != "" && !within(pane.ForegroundCwd, path) {
				var inside []string
				for _, other := range panes {
					if other.WorkspaceID == workspace && inCheckout(other, path) {
						inside = append(inside, other.PaneID)
					}
				}
				which := "is open on"
				if len(inside) > 0 {
					which = "has panes " + strings.Join(inside, ", ") + " in"
				}
				return nil, fmt.Sprintf("workspace %s %s checkout %s, and its pane %s is outside the checkout: close the panes in the checkout, or pane %s, first; nothing was removed", workspace, which, path, pane.PaneID, pane.PaneID), nil
			}
		}
	}
	return workspaces, "", nil
}

// inCheckout reports whether the shell or the foreground process of pane
// runs in the checkout at path.
func inCheckout(pane herdr.PaneInfo, path string) bool {
	return within(pane.Cwd, path) || within(pane.ForegroundCwd, path)
}

// openWorkspace returns the ID of the herdr workspace open on the checkout at
// path, or an empty string.
func openWorkspace(ctx context.Context, path string, logger *slog.Logger) (string, error) {
	listed, err := herdr.WorktreeList(ctx, herdr.WorktreeListParams{Cwd: path}, herdr.WithLogger(logger))
	if err != nil {
		return "", fmt.Errorf("worktree list %s: %w", path, err)
	}
	for _, tree := range listed.Worktrees {
		if tree.OpenWorkspaceID != "" && worktree.SamePath(tree.Path, path) {
			return tree.OpenWorkspaceID, nil
		}
	}
	return "", nil
}

// within reports whether dir is root or below root, with each symlink of
// both paths resolved when the path exists.
func within(dir, root string) bool {
	if dir == "" || root == "" {
		return false
	}
	rel, err := filepath.Rel(resolve(root), resolve(dir))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// resolve returns path with each symlink resolved, or the clean path when it
// cannot be resolved, for example when it does not exist.
func resolve(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

// Text renders the effects of one remove.
func Text(envelope output.Envelope) string {
	var lines []string
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
