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
	"unicode/utf8"

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
	maxLogText    = 4096
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
	Design          *DesignCopy  `json:"design,omitempty"`
}

type createReadFailure struct {
	phase   string
	message string
}

func (e *createReadFailure) Error() string {
	return e.message
}

// Create makes the worktree of one task.
func Create(ctx context.Context, home string, options CreateOptions) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	logger, closeLog := openWorktreeLogger(home, options.Logger)
	defer closeLog()
	ctx = run.WithLogger(ctx, logger)
	failProjectID := ""
	failTaskID := options.Task
	fail := func(phase, message string) output.Envelope {
		logFailure(logger, commandCreate, failProjectID, failTaskID, phase, message)
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
				return &createReadFailure{"task", "task not found: " + options.Task}
			}
			return &createReadFailure{"store", "read task: " + oneLine(err.Error())}
		}
		if task.ID != options.Task || !records.ValidTaskID(task.ID) {
			return &createReadFailure{"task", "task record id mismatch: " + task.ID + " != " + options.Task}
		}
		if task.Worktree != nil {
			return &createReadFailure{"task", "task has a worktree: " + task.Worktree.Path}
		}
		if task.Status != "created" && task.Status != "assigned" {
			return &createReadFailure{"task", "task is " + task.Status + ": create needs created or assigned"}
		}
		if !records.ValidID(task.Project) {
			return &createReadFailure{"project", "invalid project id: " + task.Project}
		}
		if err := locked.Read("projects", task.Project, &project); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return &createReadFailure{"project", "project not found: " + task.Project}
			}
			return &createReadFailure{"store", "read project: " + oneLine(err.Error())}
		}
		if project.ID != task.Project || !records.ValidID(project.ID) {
			return &createReadFailure{"project", "project record id mismatch: " + project.ID + " != " + task.Project}
		}
		return nil
	})
	if readErr != nil {
		var failure *createReadFailure
		if errors.As(readErr, &failure) {
			return fail(failure.phase, failure.message)
		}
		return fail("store", "read records: "+oneLine(readErr.Error()))
	}
	failProjectID = project.ID
	scoped := logger
	if logger != nil {
		scoped = logger.With("project_id", project.ID, "task_id", options.Task)
	}
	ctx = run.WithLogger(ctx, scoped)
	worktreePath := filepath.Join(home, "projects", project.ID, "worktrees", options.Task)
	branch := "task/" + options.Task
	logEvent(logger, "worktree create start", commandCreate, project.ID, options.Task, "status", task.Status)

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
	featureBrief := featureBriefRelativePath(task)

	before, err := herdr.WorkspaceList(ctx, herdr.WithLogger(scoped))
	if err != nil {
		return fail("workspace", "list workspaces: "+oneLine(err.Error()))
	}
	params := herdr.WorktreeCreateParams{
		Cwd:    project.Path,
		Branch: branch,
		Base:   project.Base,
		Path:   worktreePath,
		Label:  options.Task,
		Focus:  options.Focus,
	}
	logEvent(logger, "worktree create request", commandCreate, project.ID, options.Task, "cwd", project.Path, "branch", branch, "base", project.Base, "path", worktreePath, "focus", options.Focus)
	result, callErr := herdr.WorktreeCreate(ctx, params, herdr.WithDeadline(createTimeout), herdr.WithLogger(scoped))
	if callErr != nil {
		var herdrErr *herdr.Error
		if errors.As(callErr, &herdrErr) {
			return typedCreateFailure(ctx, logger, project, options.Task, branch, herdrErr)
		}
		return inspectUnknownCreate(ctx, registry, logger, scoped, project, options.Task, worktreePath, branch, before, callErr)
	}

	actualPath := result.Worktree.Path
	if !samePath(actualPath, worktreePath) {
		effects := createEffects(result.Workspace.WorkspaceID, actualPath, branch, "")
		logEffects(logger, commandCreate, project.ID, options.Task, effects)
		message := "worktree mismatch: path: " + actualPath + " does not match " + worktreePath
		logFailure(logger, commandCreate, project.ID, options.Task, "request", message)
		return partialEnvelope("request", message, effects, createResult(records.Task{}, result.Workspace.WorkspaceID, result.Tab.TabID, result.RootPane.PaneID, actualPath, branch, ""))
	}

	after, listErr := herdr.WorkspaceList(ctx, herdr.WithLogger(scoped))
	repoWorkspaceID := ""
	repoAmbiguous := false
	if listErr == nil {
		repoWorkspaceID, repoAmbiguous = findRepositoryWorkspace(before, after, project.Path)
	}
	worktree := records.Worktree{
		WorkspaceID: result.Workspace.WorkspaceID,
		TabID:       result.Tab.TabID,
		RootPane:    result.RootPane.PaneID,
		Path:        worktreePath,
		Branch:      branch,
		CreatedAt:   now(),
	}
	effects := createEffects(worktree.WorkspaceID, worktreePath, branch, repoWorkspaceID)
	var specCopyErr error
	var briefCopyErr error
	var design *DesignCopy
	if task.Feature != "" {
		specCopy, err := copySpecIn(home, worktreePath, task)
		if err == nil {
			err = writeSpecCopy(registry, specCopy)
		}
		if err != nil {
			specCopyErr = err
		} else {
			effects = append(effects,
				output.Effect{Resource: "spec", Path: filepath.Join(worktreePath, filepath.FromSlash(featureSpecRelDir(task.Feature))) + "/", Action: "copied"},
				output.Effect{Resource: "spec_copy", Path: SpecCopyRelPath(task.ID), Action: "created"},
			)
			logEvent(logger, "worktree spec copy-in", commandCreate, project.ID, options.Task, "path", featureSpecRelDir(task.Feature), "files", len(specCopy.Files))
		}
		counts, err := copyDesignIn(home, worktreePath, task)
		if err != nil {
			if specCopyErr == nil {
				specCopyErr = err
			}
		} else {
			design = &counts
			if counts.Features > 0 {
				effects = append(effects, output.Effect{Resource: "design", Path: fmt.Sprintf("%d features %d files", counts.Features, counts.Files), Action: "copied"})
			}
			logEvent(logger, "worktree design copy-in", commandCreate, project.ID, options.Task, "features", counts.Features, "files", counts.Files)
		}
		if err := copyFeatureBrief(home, worktreePath, task, featureBrief); err != nil {
			briefCopyErr = err
		} else {
			effects = append(effects, output.Effect{
				Resource: "brief",
				Path:     filepath.Join(worktreePath, filepath.FromSlash(featureBrief)),
				Action:   "copied",
			})
		}
	}
	logEffects(logger, commandCreate, project.ID, options.Task, effects)
	updated, phase, persistErr := persistCreate(registry, project, options.Task, &worktree, repoWorkspaceID)
	if persistErr != nil {
		message := createPersistMessage(phase, persistErr)
		logFailure(logger, commandCreate, project.ID, options.Task, phase, message)
		return partialEnvelope(phase, message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, repoWorkspaceID))
	}
	if listErr != nil {
		message := "list workspaces after create: " + oneLine(listErr.Error())
		logFailure(logger, commandCreate, project.ID, options.Task, "workspace", message)
		return partialEnvelope("workspace", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, ""))
	}
	if repoAmbiguous {
		message := "repository workspace is uncertain: more than one new workspace matches the repository"
		logFailure(logger, commandCreate, project.ID, options.Task, "workspace", message)
		return partialEnvelope("workspace", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, ""))
	}
	if specCopyErr != nil {
		message := oneLine(specCopyErr.Error())
		logFailure(logger, commandCreate, project.ID, options.Task, "spec", message)
		return partialEnvelope("spec", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, repoWorkspaceID))
	}
	if briefCopyErr != nil {
		message := "copy feature brief: " + oneLine(briefCopyErr.Error())
		logFailure(logger, commandCreate, project.ID, options.Task, "brief", message)
		return partialEnvelope("brief", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, repoWorkspaceID))
	}
	logEvent(logger, "worktree create done", commandCreate, project.ID, options.Task, "workspace_id", worktree.WorkspaceID, "path", worktreePath, "branch", branch)
	created := createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktreePath, branch, repoWorkspaceID)
	created.Design = design
	return output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: effects,
		Error:   nil,
		Result:  created,
	}
}

func typedCreateFailure(ctx context.Context, logger *slog.Logger, project records.Project, taskID, branch string, herdrErr *herdr.Error) output.Envelope {
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
		}
		logEffects(logger, commandCreate, project.ID, taskID, effects)
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

func inspectUnknownCreate(ctx context.Context, registry *store.Store, logger, scoped *slog.Logger, project records.Project, taskID, worktreePath, branch string, before []herdr.WorkspaceInfo, callErr error) output.Envelope {
	message := oneLine(callErr.Error())
	worktree := records.Worktree{}
	worktreeProven := false
	branchObserved := false
	if _, err := os.Lstat(worktreePath); err == nil {
		linked, linkedErr := gitx.IsLinkedWorktree(ctx, worktreePath)
		switch {
		case linkedErr != nil:
			message += "; inspect worktree: " + oneLine(linkedErr.Error())
		case !linked:
			message += "; path is not a linked worktree: " + worktreePath
		default:
			targetCommon, targetErr := gitx.CommonDir(ctx, worktreePath)
			repoCommon, repoErr := gitx.CommonDir(ctx, project.Path)
			switch {
			case targetErr != nil:
				message += "; inspect worktree common dir: " + oneLine(targetErr.Error())
			case repoErr != nil:
				message += "; inspect repository common dir: " + oneLine(repoErr.Error())
			case targetCommon != repoCommon:
				message += "; path worktree repo_root mismatch: " + worktreePath
			default:
				current, branchErr := gitx.CurrentBranch(ctx, worktreePath)
				switch {
				case branchErr != nil:
					message += "; inspect current branch: " + oneLine(branchErr.Error())
				case current != branch:
					message += "; worktree branch is " + current + ", want " + branch
				default:
					worktree.Path = worktreePath
					worktree.Branch = branch
					worktree.CreatedAt = now()
					worktreeProven = true
					branchObserved = true
				}
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		message += "; inspect path: " + oneLine(err.Error())
	}
	exists, branchErr := gitx.BranchExists(ctx, project.Path, branch)
	if branchErr != nil {
		message += "; inspect branch: " + oneLine(branchErr.Error())
	} else if exists {
		branchObserved = true
	}
	repoWorkspaceID := ""
	after, listErr := herdr.WorkspaceList(ctx, herdr.WithLogger(scoped))
	if listErr != nil {
		message += "; list workspaces: " + oneLine(listErr.Error())
	} else {
		workspace, repoID, repoAmbiguous := findWorktreeWorkspace(before, after, project.Path, worktreePath)
		if workspace != nil {
			worktree.WorkspaceID = workspace.WorkspaceID
			worktree.TabID = workspace.ActiveTabID
			if !worktreeProven {
				worktree.Path = worktreePath
				worktree.CreatedAt = now()
				worktreeProven = true
			}
		}
		if repoAmbiguous {
			message += "; repository workspace is uncertain: more than one new workspace matches the repository"
		} else if repoID != "" {
			repoWorkspaceID = repoID
		}
	}
	effects := createUnknownEffects(worktreeProven, worktree.WorkspaceID, worktreePath, branchObserved, branch, repoWorkspaceID)
	logEffects(logger, commandCreate, project.ID, taskID, effects)
	var persisted *records.Worktree
	if worktreeProven {
		persisted = &worktree
	}
	updated, phase, persistErr := persistCreate(registry, project, taskID, persisted, repoWorkspaceID)
	if persistErr != nil {
		message += "; " + createPersistMessage(phase, persistErr)
		logFailure(logger, commandCreate, project.ID, taskID, phase, message)
		return partialEnvelope(phase, message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktree.Path, worktree.Branch, repoWorkspaceID))
	}
	logFailure(logger, commandCreate, project.ID, taskID, "request", message)
	return partialEnvelope("request", message, effects, createResult(updated, worktree.WorkspaceID, worktree.TabID, worktree.RootPane, worktree.Path, worktree.Branch, repoWorkspaceID))
}

func createUnknownEffects(worktreeProven bool, workspaceID, path string, branchObserved bool, branch, repoWorkspaceID string) []output.Effect {
	effects := []output.Effect{}
	if workspaceID != "" {
		effects = append(effects, output.Effect{Resource: "workspace", Path: workspaceID, Action: "created"})
	}
	if worktreeProven {
		effects = append(effects, output.Effect{Resource: "worktree", Path: path, Action: "created"})
	}
	if branchObserved {
		effects = append(effects, output.Effect{Resource: "branch", Path: branch, Action: "created"})
	}
	if repoWorkspaceID != "" {
		effects = append(effects, output.Effect{Resource: "workspace", Path: repoWorkspaceID, Action: "opened"})
	}
	return effects
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

func featureBriefRelativePath(task records.Task) string {
	if task.Feature == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Join(".hive", "spec", task.Feature, "tasks", task.ID, "brief.md"))
}

func copyFeatureBrief(home, worktreePath string, task records.Task, relativePath string) error {
	source := filepath.Join(home, ".hive", "state", "briefs", task.ID+".md")
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read state brief: %s", oneLine(err.Error()))
	}
	destination := filepath.Join(worktreePath, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create feature brief directory: %s", oneLine(err.Error()))
	}
	if err := os.WriteFile(destination, data, 0o644); err != nil {
		return fmt.Errorf("write feature brief: %s", oneLine(err.Error()))
	}
	return nil
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

func persistCreate(registry *store.Store, project records.Project, taskID string, worktree *records.Worktree, repoWorkspaceID string) (records.Task, string, error) {
	var updated records.Task
	phase := ""
	err := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		if worktree != nil {
			if err := locked.UpdateIn([]string{"tasks"}, taskID, func(raw json.RawMessage) (json.RawMessage, error) {
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
				if record.Project != project.ID {
					return nil, fmt.Errorf("task record changed: project %s", record.Project)
				}
				if record.Worktree != nil {
					return nil, fmt.Errorf("task has a worktree: %s", record.Worktree.Path)
				}
				record.Worktree = worktree
				record.UpdatedAt = now()
				updated = record
				return json.Marshal(record)
			}); err != nil {
				phase = "task"
				return err
			}
		}
		if repoWorkspaceID != "" {
			if err := locked.UpdateIn([]string{"projects"}, project.ID, func(raw json.RawMessage) (json.RawMessage, error) {
				if len(raw) == 0 {
					return nil, fmt.Errorf("project not found: %s", project.ID)
				}
				var record records.Project
				if err := json.Unmarshal(raw, &record); err != nil {
					return nil, fmt.Errorf("record invalid: %s", project.ID)
				}
				if record.ID != project.ID {
					return nil, fmt.Errorf("project record changed: %s", record.ID)
				}
				if !samePath(record.Path, project.Path) {
					return nil, fmt.Errorf("project record changed: path %s", record.Path)
				}
				if record.RepoWorkspaceID != project.RepoWorkspaceID && record.RepoWorkspaceID != repoWorkspaceID {
					return nil, fmt.Errorf("project record changed: repo workspace %s", record.RepoWorkspaceID)
				}
				record.RepoWorkspaceID = repoWorkspaceID
				return json.Marshal(record)
			}); err != nil {
				phase = "project"
				return err
			}
		}
		return nil
	})
	if err != nil && phase == "" {
		phase = "store"
	}
	return updated, phase, err
}

func createPersistMessage(phase string, err error) string {
	switch phase {
	case "task":
		return "persist worktree: " + oneLine(err.Error())
	case "project":
		return "persist repository workspace: " + oneLine(err.Error())
	default:
		return "persist records: " + oneLine(err.Error())
	}
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

func logEffects(logger *slog.Logger, command, projectID, taskID string, effects []output.Effect) {
	for _, effect := range effects {
		logEffect(logger, command, projectID, taskID, effect.Resource, effect.Path, effect.Action)
	}
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
	text, truncated := boundLogText(oneLine(message))
	fields = append(fields, "error", text)
	if truncated {
		fields = append(fields, "truncated", true)
	}
	logger.Error("worktree error", fields...)
}

func boundLogText(text string) (string, bool) {
	if len(text) <= maxLogText {
		return text, false
	}
	cut := maxLogText
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut], true
}
