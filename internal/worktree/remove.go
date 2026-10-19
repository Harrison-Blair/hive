package worktree

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/gitx"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/run"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const maxOutputText = 8192

// RemoveOptions holds the values of one worktree remove.
type RemoveOptions struct {
	Task   string
	Force  bool
	Logger *slog.Logger
}

// RemoveResult is the record and effects of one worktree remove.
type RemoveResult struct {
	Task            records.Task `json:"task"`
	Path            string       `json:"path"`
	Branch          string       `json:"branch"`
	WorkspaceClosed bool         `json:"workspace_closed"`
}

type removeReadFailure struct {
	phase   string
	message string
}

func (e *removeReadFailure) Error() string {
	return e.message
}

// Remove removes the worktree of one task.
func Remove(ctx context.Context, home string, options RemoveOptions) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	logger, closeLog := openWorktreeLogger(home, options.Logger)
	defer closeLog()
	ctx = run.WithLogger(ctx, logger)
	failProjectID := ""
	failTaskID := options.Task
	fail := func(phase, message string) output.Envelope {
		logFailure(logger, commandRemove, failProjectID, failTaskID, phase, message)
		return failedEnvelope(phase, message)
	}
	if !records.ValidTaskID(options.Task) {
		return fail("validate", "task not found: "+options.Task)
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", "open store: "+oneLine(err.Error()))
	}
	var task records.Task
	var project records.Project
	readErr := registry.WithLock(ctx, func(locked *store.LockedStore) error {
		if err := locked.ReadIn([]string{"tasks"}, options.Task, &task); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return &removeReadFailure{"task", "task not found: " + options.Task}
			}
			return &removeReadFailure{"store", "read task: " + oneLine(err.Error())}
		}
		if task.ID != options.Task || !records.ValidTaskID(task.ID) {
			return &removeReadFailure{"task", "task record id mismatch: " + task.ID + " != " + options.Task}
		}
		if task.Worktree == nil || task.Worktree.Path == "" || task.Worktree.Branch == "" {
			return &removeReadFailure{"task", "task has no worktree"}
		}
		if task.Status != "verified" && task.Status != "cancelled" {
			return &removeReadFailure{"task", "task is " + task.Status + ": complete, verify, or cancel it first"}
		}
		if !records.ValidID(task.Project) {
			return &removeReadFailure{"project", "invalid project id: " + task.Project}
		}
		if err := locked.Read("projects", task.Project, &project); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return &removeReadFailure{"project", "project not found: " + task.Project}
			}
			return &removeReadFailure{"store", "read project: " + oneLine(err.Error())}
		}
		if project.ID != task.Project || !records.ValidID(project.ID) {
			return &removeReadFailure{"project", "project record id mismatch: " + project.ID + " != " + task.Project}
		}
		return nil
	})
	if readErr != nil {
		var failure *removeReadFailure
		if errors.As(readErr, &failure) {
			return fail(failure.phase, failure.message)
		}
		return fail("store", "read records: "+oneLine(readErr.Error()))
	}
	path := task.Worktree.Path
	branch := task.Worktree.Branch
	if !filepath.IsAbs(path) {
		return fail("identity", "worktree mismatch: path")
	}
	permitted := filepath.Join(home, "projects", project.ID, "worktrees", options.Task)
	if !samePath(path, permitted) {
		return fail("identity", "worktree mismatch: path")
	}
	if branch != "task/"+options.Task {
		return fail("identity", "worktree mismatch: branch")
	}
	failProjectID = project.ID
	if message, occupied := findAgentOccupant(home, options.Task, task.Worktree.WorkspaceID, path); occupied {
		return fail("agent", message)
	}
	scoped := logger
	if logger != nil {
		scoped = logger.With("project_id", project.ID, "task_id", options.Task)
	}
	ctx = run.WithLogger(ctx, scoped)
	logEvent(logger, "worktree remove start", commandRemove, project.ID, options.Task, "status", task.Status, "path", path, "branch", branch, "force", options.Force)

	sourceLinked, err := gitx.IsLinkedWorktree(ctx, project.Path)
	if err != nil {
		return fail("identity", "repository check: "+oneLine(err.Error()))
	}
	if sourceLinked {
		return fail("identity", "repository is a linked worktree: register the primary checkout")
	}
	linked, err := gitx.IsLinkedWorktree(ctx, path)
	if err != nil {
		return fail("identity", "worktree check: "+oneLine(err.Error()))
	}
	if !linked {
		return fail("identity", "worktree mismatch: is_linked_worktree")
	}
	targetCommon, err := gitx.CommonDir(ctx, path)
	if err != nil {
		return fail("identity", "worktree common dir: "+oneLine(err.Error()))
	}
	repoCommon, err := gitx.CommonDir(ctx, project.Path)
	if err != nil {
		return fail("identity", "repository common dir: "+oneLine(err.Error()))
	}
	if targetCommon != repoCommon {
		return fail("identity", "worktree mismatch: repo_root")
	}
	current, err := gitx.CurrentBranch(ctx, path)
	if err != nil {
		return fail("identity", "read current branch: "+oneLine(err.Error()))
	}
	if current != branch {
		return fail("identity", "worktree mismatch: branch")
	}

	workspaceFound := false
	workspaceID := task.Worktree.WorkspaceID
	if workspaceID != "" {
		workspace, getErr := herdr.WorkspaceGet(ctx, workspaceID, herdr.WithLogger(scoped))
		if getErr != nil {
			var herdrErr *herdr.Error
			if errors.As(getErr, &herdrErr) && herdrErr.Code == herdr.CodeWorkspaceNotFound {
				workspaceFound = false
			} else {
				return fail("workspace", "workspace get: "+oneLine(getErr.Error()))
			}
		} else {
			workspaceFound = true
			if workspace.Worktree == nil || !workspace.Worktree.IsLinkedWorktree {
				return fail("identity", "worktree mismatch: is_linked_worktree")
			}
			if !samePath(workspace.Worktree.CheckoutPath, path) {
				return fail("identity", "worktree mismatch: checkout_path")
			}
			if !samePath(workspace.Worktree.RepoRoot, project.Path) {
				return fail("identity", "worktree mismatch: repo_root")
			}
		}
	}

	if !options.Force {
		dirty, err := gitx.IsDirty(ctx, path)
		if err != nil {
			return fail("guard", "read worktree state: "+oneLine(err.Error()))
		}
		if dirty {
			return fail("guard", "worktree is dirty")
		}
		merged, err := gitx.IsMerged(ctx, project.Path, project.Base, branch)
		if err != nil {
			return fail("guard", "read merge state: "+oneLine(err.Error()))
		}
		if !merged {
			return fail("guard", "branch not merged into "+project.Base)
		}
	}

	message, occupied, listErr := liveWorkspaceAgent(ctx, task.Worktree.WorkspaceID, scoped)
	if listErr != nil {
		return fail("agent", "agent occupancy unknown: "+oneLine(listErr.Error()))
	}
	if occupied {
		return fail("agent", message)
	}

	effects := []output.Effect{}
	closed := false
	if workspaceFound {
		closeErr := herdr.WorkspaceClose(ctx, workspaceID, false, herdr.WithLogger(scoped))
		if closeErr != nil {
			var herdrErr *herdr.Error
			if errors.As(closeErr, &herdrErr) {
				switch herdrErr.Code {
				case herdr.CodeWorkspaceNotFound:
				case "workspace_group_close_required":
					return fail("workspace", "workspace is a group root: close it by hand")
				default:
					message := herdrErr.Message
					if message == "" {
						message = closeErr.Error()
					}
					return fail("workspace", message)
				}
			} else {
				message := "workspace close outcome unknown: " + oneLine(closeErr.Error())
				logFailure(logger, commandRemove, project.ID, options.Task, "workspace", message)
				return partialEnvelope("workspace", message, nil, nil)
			}
		} else {
			closed = true
			effects = append(effects, output.Effect{Resource: "workspace", Path: workspaceID, Action: "closed"})
			logEffect(logger, commandRemove, project.ID, options.Task, "workspace", workspaceID, "closed")
		}
	}

	removeResult, removeErr := gitx.WorktreeRemove(ctx, project.Path, path, options.Force)
	gitFailure := ""
	if removeErr != nil {
		gitFailure = "git worktree remove: " + oneLine(removeErr.Error())
	} else if removeResult.ExitCode != 0 {
		gitFailure = worktreeRemoveFailure(removeResult)
	}
	if gitFailure != "" {
		logFailure(logger, commandRemove, project.ID, options.Task, "worktree", gitFailure)
		if closed {
			return partialEnvelope("worktree", gitFailure, effects, RemoveResult{Task: task, Path: path, Branch: branch, WorkspaceClosed: true})
		}
		return failedEnvelopeWithEffects("worktree", gitFailure, effects)
	}
	effects = append(effects, output.Effect{Resource: "worktree", Path: path, Action: "removed"})
	logEffect(logger, commandRemove, project.ID, options.Task, "worktree", path, "removed")

	updated, err := persistRemoved(registry, project.ID, options.Task, path, branch, task.Worktree.WorkspaceID, task.Worktree.TabID, task.Worktree.RootPane)
	if err != nil {
		message := "persist worktree removal: " + oneLine(err.Error())
		logFailure(logger, commandRemove, project.ID, options.Task, "task", message)
		return partialEnvelope("task", message, effects, RemoveResult{Task: task, Path: path, Branch: branch, WorkspaceClosed: closed})
	}
	logEvent(logger, "worktree remove done", commandRemove, project.ID, options.Task, "path", path, "branch", branch, "workspace_closed", closed)
	return output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: effects,
		Error:   nil,
		Result:  RemoveResult{Task: updated, Path: path, Branch: branch, WorkspaceClosed: closed},
	}
}

// findAgentOccupant returns the first agent record that occupies the worktree
// by the task, the workspace, or the path, unless its status is stopped or
// exited.
func findAgentOccupant(home, taskID, workspaceID, path string) (string, bool) {
	registry, err := store.Open(home)
	if err != nil {
		return "", false
	}
	raw, err := registry.ListIn([]string{"agents"}, nil)
	if err != nil {
		return "", false
	}
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}
		if record.Status == records.AgentStopped || record.Status == records.AgentExited {
			continue
		}
		occupied := record.Task == taskID
		if !occupied && workspaceID != "" && record.WorkspaceID == workspaceID {
			occupied = true
		}
		if !occupied && path != "" && record.Cwd != "" && samePath(record.Cwd, path) {
			occupied = true
		}
		if !occupied {
			continue
		}
		name := record.Name
		if name == "" {
			name = record.Key
		}
		return "agent " + record.Status + ": " + name, true
	}
	return "", false
}

// liveWorkspaceAgent returns the first live herdr agent in the workspace. A
// list failure returns an error, because an unknown occupancy cannot permit a
// removal.
func liveWorkspaceAgent(ctx context.Context, workspaceID string, logger *slog.Logger) (string, bool, error) {
	if workspaceID == "" {
		return "", false, nil
	}
	agents, err := herdr.AgentList(ctx, herdr.WithLogger(logger))
	if err != nil {
		if logger != nil {
			logger.Debug("worktree remove agent list failed", "command", commandRemove, "workspace_id", workspaceID, "error", err.Error())
		}
		return "", false, err
	}
	for _, agent := range agents {
		if agent.WorkspaceID == workspaceID {
			return "agent in workspace: " + agent.PaneID, true, nil
		}
	}
	return "", false, nil
}

func worktreeRemoveFailure(result *run.Result) string {
	text := strings.TrimSpace(result.Stderr)
	if text == "" {
		text = strings.TrimSpace(result.Stdout)
	}
	if text == "" {
		return fmt.Sprintf("git worktree remove exited with code %d", result.ExitCode)
	}
	line := text
	if index := strings.IndexByte(line, '\n'); index >= 0 {
		line = line[:index]
	}
	line = oneLine(strings.TrimSpace(line))
	if len(line) > maxOutputText {
		line = line[:maxOutputText]
	}
	return "git worktree remove: " + line
}

func persistRemoved(registry *store.Store, projectID, taskID, path, branch, workspaceID, tabID, rootPane string) (records.Task, error) {
	var updated records.Task
	err := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		return locked.UpdateIn([]string{"tasks"}, taskID, func(raw json.RawMessage) (json.RawMessage, error) {
			if len(raw) == 0 {
				return nil, fmt.Errorf("task not found: %s", taskID)
			}
			var record records.Task
			if err := json.Unmarshal(raw, &record); err != nil {
				return nil, fmt.Errorf("record invalid: %s", taskID)
			}
			if record.ID != taskID {
				return nil, fmt.Errorf("task record changed: %s", record.ID)
			}
			if record.Project != projectID {
				return nil, fmt.Errorf("task record changed: project %s", record.Project)
			}
			if record.Worktree == nil {
				return nil, fmt.Errorf("task has no worktree")
			}
			if record.Worktree.Path != path || record.Worktree.Branch != branch {
				return nil, fmt.Errorf("task worktree changed: %s", record.Worktree.Path)
			}
			currentTopology := record.Worktree.WorkspaceID + "\x00" + record.Worktree.TabID + "\x00" + record.Worktree.RootPane
			capturedTopology := workspaceID + "\x00" + tabID + "\x00" + rootPane
			clearedTopology := record.Worktree.WorkspaceID == "" && record.Worktree.TabID == "" && record.Worktree.RootPane == ""
			if currentTopology != capturedTopology && !clearedTopology {
				return nil, fmt.Errorf("task worktree changed: topology %s/%s/%s", record.Worktree.WorkspaceID, record.Worktree.TabID, record.Worktree.RootPane)
			}
			if record.Worktree.RemovedAt != "" {
				return nil, fmt.Errorf("task worktree already removed: %s", record.Worktree.RemovedAt)
			}
			record.Worktree.RemovedAt = now()
			record.Worktree.WorkspaceID = ""
			record.Worktree.TabID = ""
			record.Worktree.RootPane = ""
			record.UpdatedAt = now()
			updated = record
			return json.Marshal(record)
		})
	})
	return updated, err
}
