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
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/gitx"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/run"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	commandCreate = "worktree create"
	commandRemove = "worktree remove"
	createTimeout = 60 * time.Second
)

// CreateOptions holds the values of one worktree create.
type CreateOptions struct {
	Task   string
	Focus  bool
	Logger *slog.Logger
}

// CreateResult is the record and topology of one worktree create.
type CreateResult struct {
	Task            records.Task `json:"task"`
	WorkspaceID     string       `json:"workspace_id"`
	TabID           string       `json:"tab_id"`
	RootPane        string       `json:"root_pane"`
	Path            string       `json:"path"`
	Branch          string       `json:"branch"`
	RepoWorkspaceID string       `json:"repo_workspace_id,omitempty"`
}

// Create makes the worktree of one task.
func Create(ctx context.Context, home string, options CreateOptions) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	logger, closeLog := openWorktreeLogger(home, options.Logger)
	defer closeLog()
	ctx = run.WithLogger(ctx, logger)
	fail := func(phase, message string) output.Envelope {
		logFailure(logger, commandCreate, "", options.Task, phase, message)
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
	if err := registry.ReadIn([]string{"tasks"}, options.Task, &task); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fail("task", "task not found: "+options.Task)
		}
		return fail("store", "read task: "+oneLine(err.Error()))
	}
	if task.Worktree != nil {
		return fail("task", "task has a worktree: "+task.Worktree.Path)
	}
	if task.Status != "created" && task.Status != "assigned" {
		return fail("task", "task is "+task.Status+": create needs created or assigned")
	}
	var project records.Project
	if err := registry.Read("projects", task.Project, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fail("project", "project not found: "+task.Project)
		}
		return fail("store", "read project: "+oneLine(err.Error()))
	}
	if !records.ValidID(project.ID) {
		return fail("project", "invalid project id: "+project.ID)
	}
	worktreePath := filepath.Join(home, "projects", project.ID, "worktrees", task.ID)
	branch := "task/" + task.ID
	logEvent(logger, "worktree create start", commandCreate, project.ID, task.ID, "status", task.Status)

	linked, err := gitx.IsLinkedWorktree(ctx, project.Path)
	if err != nil {
		return fail("precheck", "read repository: "+oneLine(err.Error()))
	}
	if linked {
		return fail("precheck", "repository is a linked worktree: register the primary checkout")
	}
	exists, err := gitx.RefExists(ctx, project.Path, project.Base)
	if err != nil {
		return fail("precheck", "read base ref: "+oneLine(err.Error()))
	}
	if !exists {
		return fail("precheck", "base ref not found: "+project.Base)
	}
	exists, err = gitx.BranchExists(ctx, project.Path, branch)
	if err != nil {
		return fail("precheck", "read branch: "+oneLine(err.Error()))
	}
	if exists {
		return fail("precheck", "branch exists: "+branch)
	}
	if _, err := os.Lstat(worktreePath); err == nil {
		return fail("precheck", "path exists: "+worktreePath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fail("precheck", "check path: "+oneLine(err.Error()))
	}

	before, err := herdr.WorkspaceList(ctx, herdr.WithLogger(logger))
	if err != nil {
		return fail("workspace", "list workspaces: "+oneLine(err.Error()))
	}
	params := herdr.WorktreeCreateParams{
		Cwd:    project.Path,
		Branch: branch,
		Base:   project.Base,
		Path:   worktreePath,
		Label:  task.ID,
		Focus:  options.Focus,
	}
	logEvent(logger, "worktree create request", commandCreate, project.ID, task.ID, "cwd", project.Path, "branch", branch, "base", project.Base, "path", worktreePath, "focus", options.Focus)
	result, callErr := herdr.WorktreeCreate(ctx, params, herdr.WithDeadline(createTimeout), herdr.WithLogger(logger))
	if callErr != nil {
		var herdrErr *herdr.Error
		if errors.As(callErr, &herdrErr) {
			return typedCreateFailure(ctx, registry, logger, project, task.ID, branch, herdrErr)
		}
		return inspectUnknownCreate(ctx, registry, logger, project, task.ID, worktreePath, branch, before, callErr)
	}

	actualPath := result.Worktree.Path
	if !samePath(actualPath, worktreePath) {
		effects := createEffects(result.Workspace.WorkspaceID, actualPath, branch, "")
		message := "worktree mismatch: path: " + actualPath + " does not match " + worktreePath
		logFailure(logger, commandCreate, project.ID, task.ID, "request", message)
		return partialEnvelope("request", message, effects, createResult(records.Task{}, result.Workspace.WorkspaceID, result.Tab.TabID, result.RootPane.PaneID, actualPath, branch, ""))
	}

	after, listErr := herdr.WorkspaceList(ctx, herdr.WithLogger(logger))
	repoWorkspaceID, repoAmbiguous := findRepositoryWorkspace(before, after, project.Path)
	worktree := records.Worktree{
		WorkspaceID: result.Workspace.WorkspaceID,
		TabID:       result.Tab.TabID,
		RootPane:    result.RootPane.PaneID,
		Path:        worktreePath,
		Branch:      branch,
		CreatedAt:   now(),
	}
	updated, persistErr := persistWorktree(registry, task.ID, worktree)
	effects := createEffects(worktree.WorkspaceID, worktreePath, branch, "")
	if persistErr != nil {
		message := "persist worktree: " + oneLine(persistErr.Error())
		logFailure(logger, commandCreate, project.ID, task.ID, "task", message)
		return partialEnvelope("task", message, effects, createResult(records.Task{}, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, ""))
	}
	logEffect(logger, commandCreate, project.ID, task.ID, "workspace", worktree.WorkspaceID, "created")
	logEffect(logger, commandCreate, project.ID, task.ID, "worktree", worktreePath, "created")
	logEffect(logger, commandCreate, project.ID, task.ID, "branch", branch, "created")
	if listErr != nil {
		message := "list workspaces after create: " + oneLine(listErr.Error())
		logFailure(logger, commandCreate, project.ID, task.ID, "workspace", message)
		return partialEnvelope("workspace", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, ""))
	}
	if repoAmbiguous {
		message := "repository workspace is uncertain: more than one new workspace matches the repository"
		logFailure(logger, commandCreate, project.ID, task.ID, "workspace", message)
		return partialEnvelope("workspace", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, ""))
	}
	if repoWorkspaceID != "" {
		effects = append(effects, output.Effect{Resource: "workspace", Path: repoWorkspaceID, Action: "opened"})
		logEffect(logger, commandCreate, project.ID, task.ID, "workspace", repoWorkspaceID, "opened")
		if err := persistRepoWorkspace(registry, project.ID, repoWorkspaceID); err != nil {
			message := "persist repository workspace: " + oneLine(err.Error())
			logFailure(logger, commandCreate, project.ID, task.ID, "project", message)
			return partialEnvelope("project", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, repoWorkspaceID))
		}
	}
	logEvent(logger, "worktree create done", commandCreate, project.ID, task.ID, "workspace_id", worktree.WorkspaceID, "path", worktreePath, "branch", branch)
	return output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: effects,
		Error:   nil,
		Result:  createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, repoWorkspaceID),
	}
}

func typedCreateFailure(ctx context.Context, registry *store.Store, logger *slog.Logger, project records.Project, taskID, branch string, herdrErr *herdr.Error) output.Envelope {
	message := herdrErr.Message
	if message == "" {
		message = herdrErr.Code
	}
	switch herdrErr.Code {
	case "worktree_create_failed":
		effects := []output.Effect{}
		exists, err := gitx.BranchExists(ctx, project.Path, branch)
		if err == nil && exists {
			effects = append(effects, output.Effect{Resource: "branch", Path: branch, Action: "created"})
			message += "; branch " + branch + " stays"
			logEffect(logger, commandCreate, project.ID, taskID, "branch", branch, "created")
		}
		logFailure(logger, commandCreate, project.ID, taskID, "request", message)
		return failedEnvelopeWithEffects("request", message, effects)
	case "worktree_operation_in_progress":
		message = "worktree operation in progress: retry later"
		logFailure(logger, commandCreate, project.ID, taskID, "request", message)
		return failedEnvelope("request", message)
	default:
		logFailure(logger, commandCreate, project.ID, taskID, "request", message)
		return failedEnvelope("request", message)
	}
}

func inspectUnknownCreate(ctx context.Context, registry *store.Store, logger *slog.Logger, project records.Project, taskID, worktreePath, branch string, before []herdr.WorkspaceInfo, callErr error) output.Envelope {
	message := oneLine(callErr.Error())
	effects := []output.Effect{}
	worktree := records.Worktree{}
	pathExists := false
	if _, err := os.Lstat(worktreePath); err == nil {
		pathExists = true
		worktree.Path = worktreePath
		worktree.Branch = branch
		worktree.CreatedAt = now()
		effects = append(effects, output.Effect{Resource: "worktree", Path: worktreePath, Action: "created"})
	}
	exists, err := gitx.BranchExists(ctx, project.Path, branch)
	if err == nil && exists {
		effects = append([]output.Effect{{Resource: "branch", Path: branch, Action: "created"}}, effects...)
	}
	repoWorkspaceID := ""
	after, listErr := herdr.WorkspaceList(ctx, herdr.WithLogger(logger))
	if listErr == nil {
		workspace, repoID, repoAmbiguous := findWorktreeWorkspace(before, after, project.Path, worktreePath)
		if workspace != nil {
			worktree.WorkspaceID = workspace.WorkspaceID
			worktree.TabID = workspace.ActiveTabID
			effects = append([]output.Effect{{Resource: "workspace", Path: workspace.WorkspaceID, Action: "created"}}, effects...)
			if !pathExists {
				worktree.Path = worktreePath
				worktree.Branch = branch
				worktree.CreatedAt = now()
				pathExists = true
			}
		}
		if !repoAmbiguous {
			repoWorkspaceID = repoID
			if repoWorkspaceID != "" {
				effects = append(effects, output.Effect{Resource: "workspace", Path: repoWorkspaceID, Action: "opened"})
			}
		}
	} else {
		message += "; list workspaces: " + oneLine(listErr.Error())
	}
	if pathExists {
		updated, persistErr := persistWorktree(registry, taskID, worktree)
		if persistErr != nil {
			message += "; persist worktree: " + oneLine(persistErr.Error())
			logFailure(logger, commandCreate, project.ID, taskID, "task", message)
			return partialEnvelope("task", message, effects, createResult(records.Task{}, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktree.Path, worktree.Branch, ""))
		}
		if repoWorkspaceID != "" {
			if err := persistRepoWorkspace(registry, project.ID, repoWorkspaceID); err != nil {
				message += "; persist repository workspace: " + oneLine(err.Error())
				logFailure(logger, commandCreate, project.ID, taskID, "project", message)
				return partialEnvelope("project", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktree.Path, worktree.Branch, repoWorkspaceID))
			}
		}
		logFailure(logger, commandCreate, project.ID, taskID, "request", message)
		return partialEnvelope("request", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktree.Path, worktree.Branch, repoWorkspaceID))
	}
	logFailure(logger, commandCreate, project.ID, taskID, "request", message)
	return partialEnvelope("request", message, effects, createResult(records.Task{}, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktree.Path, worktree.Branch, repoWorkspaceID))
}

func createEffects(workspaceID, path, branch, repoWorkspaceID string) []output.Effect {
	effects := []output.Effect{}
	if workspaceID != "" {
		effects = append(effects, output.Effect{Resource: "workspace", Path: workspaceID, Action: "created"})
	}
	if path != "" {
		effects = append(effects, output.Effect{Resource: "worktree", Path: path, Action: "created"})
	}
	if branch != "" {
		effects = append(effects, output.Effect{Resource: "branch", Path: branch, Action: "created"})
	}
	if repoWorkspaceID != "" {
		effects = append(effects, output.Effect{Resource: "workspace", Path: repoWorkspaceID, Action: "opened"})
	}
	return effects
}

func createResult(task records.Task, workspaceID, tabID, rootPane, path, branch, repoWorkspaceID string) CreateResult {
	return CreateResult{
		Task:            task,
		WorkspaceID:     workspaceID,
		TabID:           tabID,
		RootPane:        rootPane,
		Path:            path,
		Branch:          branch,
		RepoWorkspaceID: repoWorkspaceID,
	}
}

func findRepositoryWorkspace(before, after []herdr.WorkspaceInfo, repoPath string) (string, bool) {
	_, repoID, ambiguous := findWorktreeWorkspace(before, after, repoPath, "")
	return repoID, ambiguous
}

func findWorktreeWorkspace(before, after []herdr.WorkspaceInfo, repoPath, worktreePath string) (*herdr.WorkspaceInfo, string, bool) {
	seen := make(map[string]bool, len(before))
	for _, workspace := range before {
		seen[workspace.WorkspaceID] = true
	}
	repoCanonical := canonicalPath(repoPath)
	worktreeCanonical := ""
	if worktreePath != "" {
		worktreeCanonical = canonicalPath(worktreePath)
	}
	var linked *herdr.WorkspaceInfo
	linkedCount := 0
	repositoryIDs := []string{}
	for _, workspace := range after {
		if seen[workspace.WorkspaceID] {
			continue
		}
		if workspace.Worktree == nil {
			continue
		}
		if !samePath(workspace.Worktree.RepoRoot, repoCanonical) {
			continue
		}
		if workspace.Worktree.IsLinkedWorktree {
			if worktreeCanonical == "" || !samePath(workspace.Worktree.CheckoutPath, worktreeCanonical) {
				continue
			}
			candidate := workspace
			linked = &candidate
			linkedCount++
			continue
		}
		if samePath(workspace.Worktree.CheckoutPath, repoCanonical) {
			repositoryIDs = append(repositoryIDs, workspace.WorkspaceID)
		}
	}
	if linkedCount > 1 {
		linked = nil
	}
	repoID := ""
	ambiguous := len(repositoryIDs) > 1
	if len(repositoryIDs) == 1 {
		repoID = repositoryIDs[0]
	}
	return linked, repoID, ambiguous
}

func persistWorktree(registry *store.Store, taskID string, worktree records.Worktree) (records.Task, error) {
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
			if record.Worktree != nil {
				return nil, fmt.Errorf("task has a worktree: %s", record.Worktree.Path)
			}
			record.Worktree = &worktree
			record.UpdatedAt = now()
			updated = record
			return json.Marshal(record)
		})
	})
	return updated, err
}

func persistRepoWorkspace(registry *store.Store, projectID, workspaceID string) error {
	return registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		return locked.UpdateIn([]string{"projects"}, projectID, func(raw json.RawMessage) (json.RawMessage, error) {
			if len(raw) == 0 {
				return nil, fmt.Errorf("project not found: %s", projectID)
			}
			var record records.Project
			if err := json.Unmarshal(raw, &record); err != nil {
				return nil, fmt.Errorf("record invalid: %s", projectID)
			}
			record.RepoWorkspaceID = workspaceID
			return json.Marshal(record)
		})
	})
}

func canonicalPath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return filepath.Clean(absolute)
	}
	return resolved
}

func samePath(left, right string) bool {
	return canonicalPath(left) == canonicalPath(right)
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func failedEnvelope(phase, message string) output.Envelope {
	return failedEnvelopeWithEffects(phase, message, nil)
}

func failedEnvelopeWithEffects(phase, message string, effects []output.Effect) output.Envelope {
	if effects == nil {
		effects = []output.Effect{}
	}
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: effects, Error: oneLine(message), Result: nil}
}

func partialEnvelope(phase, message string, effects []output.Effect, result any) output.Envelope {
	if effects == nil {
		effects = []output.Effect{}
	}
	return output.Envelope{Outcome: "partial", Phase: phase, Effects: effects, Error: oneLine(message), Result: result}
}

func oneLine(value string) string {
	return strings.Map(func(character rune) rune {
		if character < 0x20 || character == 0x7f {
			return ' '
		}
		return character
	}, value)
}

func openWorktreeLogger(home string, injected *slog.Logger) (*slog.Logger, func()) {
	if injected != nil {
		return injected, func() {}
	}
	writer, err := hivelog.NewWriter(home, "worktree", time.Now)
	if err != nil {
		return nil, func() {}
	}
	return slog.New(writer), func() { _ = writer.Close() }
}

func logEvent(logger *slog.Logger, message, command, projectID, taskID string, attrs ...any) {
	if logger == nil {
		return
	}
	fields := []any{"command", command}
	if projectID != "" {
		fields = append(fields, "project_id", projectID)
	}
	if taskID != "" {
		fields = append(fields, "task_id", taskID)
	}
	fields = append(fields, attrs...)
	logger.Debug(message, fields...)
}

func logEffect(logger *slog.Logger, command, projectID, taskID, resource, path, action string) {
	logEvent(logger, "worktree effect", command, projectID, taskID, "resource", resource, "path", path, "action", action)
}

func logFailure(logger *slog.Logger, command, projectID, taskID, phase, message string) {
	if logger == nil {
		return
	}
	fields := []any{"command", command}
	if projectID != "" {
		fields = append(fields, "project_id", projectID)
	}
	if taskID != "" {
		fields = append(fields, "task_id", taskID)
	}
	if phase != "" {
		fields = append(fields, "phase", phase)
	}
	fields = append(fields, "error", oneLine(message))
	logger.Error("worktree error", fields...)
}
