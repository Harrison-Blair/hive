package land

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

	"github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/identity"
	hivelog "github.com/Harrison-Blair/hive/internal/lib/log"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	// mergeTimeout bounds the merge command.
	mergeTimeout = 60 * time.Second
	// pushTimeout bounds the push command.
	pushTimeout = 120 * time.Second
	// ghTimeout bounds one gh command.
	ghTimeout = 60 * time.Second
	// cleanupTimeout bounds the merge cleanup after a failed merge. It uses
	// a fresh context so an expired caller context cannot prevent the abort.
	cleanupTimeout = 30 * time.Second
	// maxLogText caps one log message.
	maxLogText = 4096
)

// Options holds the values of one landing.
type Options struct {
	Task     string
	Mode     string
	Grant    string
	Evidence string
	Draft    bool
	Logger   *slog.Logger
	// Agent is the value of --agent.
	Agent string
	// Pane is the caller pane of --pane or HERDR_PANE_ID.
	Pane string
	// WorkingDir is the working directory of the command. It is the user
	// rule input when the caller has no agent record.
	WorkingDir string
	// MergeTimeout and PushTimeout override the git bounds in a test. A zero
	// value uses the fixed bound.
	MergeTimeout time.Duration
	PushTimeout  time.Duration
}

// Result is the delivery of one landing.
type Result struct {
	Task   records.Task `json:"task"`
	Mode   string       `json:"mode"`
	Target string       `json:"target"`
	Commit string       `json:"commit"`
	Pushed bool         `json:"pushed"`
	PRURL  string       `json:"pr_url,omitempty"`
	Note   string       `json:"note,omitempty"`
}

type landFailure struct {
	phase   string
	message string
}

func (e *landFailure) Error() string {
	return e.message
}

// Land delivers one verified task branch to its project.
func Land(ctx context.Context, home string, options Options) output.Envelope {
	if ctx == nil {
		ctx = context.Background()
	}
	logger, closeLog := openLogger(home, options.Logger)
	defer closeLog()
	failProjectID := ""
	failTaskID := options.Task
	fail := func(phase, message string) output.Envelope {
		logFailure(logger, failProjectID, failTaskID, phase, message)
		return failedEnvelope(phase, message)
	}
	if !records.ValidTaskID(options.Task) {
		return fail("validate", "task not found: "+options.Task)
	}
	if options.Mode != "merge" && options.Mode != "pr" {
		return fail("validate", "mode must be merge or pr")
	}
	if options.Grant == "" {
		return fail("validate", "land refused: pass --grant")
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", "open store: "+oneLine(err.Error()))
	}

	var pre records.Task
	if err := readTask(registry, options.Task, &pre); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fail("read", "task not found: "+options.Task)
		}
		return fail("store", "read task: "+oneLine(err.Error()))
	}
	failProjectID = pre.Project

	if err := checkCaller(ctx, home, registry, pre, options); err != nil {
		return fail("identity", err.Error())
	}

	lock, err := store.LockNamed(ctx, home, "land-"+pre.Project)
	if err != nil {
		return fail("lock", "named lock: "+oneLine(err.Error()))
	}
	defer func() { _ = lock.Unlock() }()

	var task records.Task
	var project records.Project
	err = registry.WithLock(ctx, func(locked *store.LockedStore) error {
		if err := locked.ReadIn([]string{"tasks"}, options.Task, &task); err != nil {
			return &landFailure{"read", "task not found: " + options.Task}
		}
		if task.Status != "verified" {
			return &landFailure{"precondition", "task is " + task.Status + ": land needs verified"}
		}
		if task.Verification == nil || task.Verification.Result != "pass" {
			return &landFailure{"precondition", "task verification is not a pass"}
		}
		if task.Worktree == nil || task.Worktree.Branch == "" {
			return &landFailure{"precondition", "task has no worktree branch"}
		}
		if !records.ValidID(task.Project) {
			return &landFailure{"precondition", "invalid project id: " + task.Project}
		}
		if err := locked.Read("projects", task.Project, &project); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return &landFailure{"precondition", "project not found: " + task.Project}
			}
			return &landFailure{"store", "read project: " + oneLine(err.Error())}
		}
		if task.Delivery != nil && task.Delivery.Commit != "" && task.Delivery.Mode != options.Mode {
			return &landFailure{"precondition", "task delivered in mode " + task.Delivery.Mode}
		}
		return nil
	})
	if err != nil {
		var failure *landFailure
		if errors.As(err, &failure) {
			return fail(failure.phase, failure.message)
		}
		return fail("store", "read records: "+oneLine(err.Error()))
	}
	failProjectID = project.ID
	scoped := logger
	if logger != nil {
		scoped = logger.With("project_id", project.ID, "task_id", options.Task)
	}
	logEvent(scoped, "land start", project.ID, options.Task, "mode", options.Mode, "base", project.Base)

	if options.Mode == "merge" {
		return landSpec(home, scoped, task, landMerge(ctx, registry, scoped, options, task, project))
	}
	return landPR(ctx, registry, scoped, home, options, task, project)
}

// checkCaller applies the caller rules of D8 to one landing. An agent caller
// must be the live owner of the task project or the coordinator. A caller with
// no agent record is the user when chat.ValidateCaller accepts its pane and
// working directory.
func checkCaller(ctx context.Context, home string, registry *store.Store, task records.Task, options Options) error {
	var caller records.Agent
	var callerErr error
	if err := registry.WithLock(ctx, func(locked *store.LockedStore) error {
		caller, callerErr = identity.ResolveHeld(locked, options.Agent, options.Pane)
		return nil
	}); err != nil {
		return err
	}
	if callerErr == nil {
		if caller.Kind == records.AgentKindCoordinator || caller.Owns == task.Project {
			return nil
		}
		return fmt.Errorf("land refused: %s is not the owner of project %s or the coordinator", oneLine(caller.Name), task.Project)
	}
	if !errors.Is(callerErr, identity.ErrCallerUnknown) {
		return errors.New("caller unknown: pass --agent")
	}
	if err := chat.ValidateCaller(ctx, home, options.Agent, options.Pane, options.WorkingDir); err != nil {
		return fmt.Errorf("land refused: the caller is not the user: %s", err.Error())
	}
	return nil
}

// storeDelivery writes the delivery fields under the store lock.
func storeDelivery(registry *store.Store, taskID, projectID string, delivery records.Delivery) (records.Task, error) {
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
			record.Delivery = &delivery
			record.UpdatedAt = now()
			updated = record
			return json.Marshal(record)
		})
	})
	return updated, err
}

// resumeDelivery returns the delivery of a retry. A retry without a citation
// flag keeps the recorded grant and evidence values and updates at.
func resumeDelivery(recorded records.Delivery, mode, target string, options Options) records.Delivery {
	delivery := recorded
	delivery.Mode = mode
	delivery.Target = target
	if options.Grant != "" {
		delivery.Grant = options.Grant
	}
	if options.Evidence != "" {
		delivery.Evidence = options.Evidence
	}
	delivery.At = now()
	return delivery
}

// deliveryEnvelope writes the delivery and builds the result envelope.
func deliveryEnvelope(registry *store.Store, logger *slog.Logger, task records.Task, project records.Project, delivery records.Delivery, outcome, phase, message, note string, effects []output.Effect) output.Envelope {
	updated, err := storeDelivery(registry, task.ID, project.ID, delivery)
	if err != nil {
		logFailure(logger, project.ID, task.ID, "record", err.Error())
		return failedEnvelopeWithEffects("record", "persist delivery: "+oneLine(err.Error()), effects)
	}
	result := Result{
		Task:   updated,
		Mode:   delivery.Mode,
		Target: delivery.Target,
		Commit: delivery.Commit,
		Pushed: delivery.Pushed,
		PRURL:  delivery.PRURL,
		Note:   note,
	}
	if outcome == "partial" {
		return partialEnvelope(phase, message, effects, result)
	}
	return okEnvelope(result, effects)
}

func readTask(registry *store.Store, id string, target *records.Task) error {
	return registry.ReadIn([]string{"tasks"}, id, target)
}

func bound(override, fallback time.Duration) time.Duration {
	if override > 0 {
		return override
	}
	return fallback
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func okEnvelope(result any, effects []output.Effect) output.Envelope {
	if effects == nil {
		effects = []output.Effect{}
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: effects, Error: nil, Result: result}
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

func openLogger(home string, injected *slog.Logger) (*slog.Logger, func()) {
	if injected != nil {
		return injected, func() {}
	}
	writer, err := hivelog.NewWriter(home, "land", time.Now)
	if err != nil {
		return nil, func() {}
	}
	return slog.New(writer), func() { _ = writer.Close() }
}

func logEvent(logger *slog.Logger, message, projectID, taskID string, attrs ...any) {
	if logger == nil {
		return
	}
	fields := []any{"command", "land"}
	if projectID != "" {
		fields = append(fields, "project_id", projectID)
	}
	if taskID != "" {
		fields = append(fields, "task_id", taskID)
	}
	fields = append(fields, attrs...)
	logger.Debug(message, fields...)
}

func logEffect(logger *slog.Logger, projectID, taskID, resource, path, action string) {
	logEvent(logger, "land effect", projectID, taskID, "resource", resource, "path", path, "action", action)
}

func logFailure(logger *slog.Logger, projectID, taskID, phase, message string) {
	if logger == nil {
		return
	}
	fields := []any{"command", "land"}
	if projectID != "" {
		fields = append(fields, "project_id", projectID)
	}
	if taskID != "" {
		fields = append(fields, "task_id", taskID)
	}
	if phase != "" {
		fields = append(fields, "phase", phase)
	}
	text := oneLine(message)
	if len(text) > maxLogText {
		text = text[:maxLogText]
		fields = append(fields, "truncated", true)
	}
	fields = append(fields, "error", text)
	logger.Error("land error", fields...)
}

// briefPath returns the brief path of one task.
func briefPath(home, taskID string) string {
	return filepath.Join(home, ".hive", "state", "briefs", taskID+".md")
}
