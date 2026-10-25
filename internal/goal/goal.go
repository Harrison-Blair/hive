package goal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/identity"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	goalsCollection = "goals"
	tasksCollection = "tasks"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var doneWhenValues = map[string]bool{
	"local":   true,
	"dev":     true,
	"pr":      true,
	"main":    true,
	"release": true,
}

// CreateOptions holds the values for one goal creation.
type CreateOptions struct {
	Project string
	Title   string
	Slug    string
	Grant   string
	Agent   string
	PaneID  string
}

// EditOptions holds the values for one goal edit.
type EditOptions struct {
	Goal           string
	Agent          string
	PaneID         string
	AddWave        string
	AddWaveSet     bool
	DoneWhen       string
	Wave           string
	AddTask        string
	AddTaskSet     bool
	NeedsUser      string
	NeedsUserSet   bool
	ClearNeedsUser bool
	WaveStatus     string
	WaveStatusSet  bool
	Status         string
	StatusSet      bool
}

// ShowResult includes wave progress from task delivery records.
type ShowResult struct {
	records.Goal
	WaveProgress []WaveProgress `json:"wave_progress"`
}

// WaveProgress reports the landed task count for one wave.
type WaveProgress struct {
	ID     string `json:"id"`
	Landed int    `json:"landed"`
	Total  int    `json:"total"`
}

// Create makes an active goal with no waves.
func Create(home string, options CreateOptions) output.Envelope {
	log := openOperationLog(home)
	defer log.close()
	actorID := options.Agent
	fail := func(phase, message string) output.Envelope {
		log.failure("goal create", options.Project, "", actorID, phase, message)
		return failed(phase, message)
	}
	if !records.ValidID(options.Project) {
		return fail("validate", "project requires a valid ID")
	}
	if strings.TrimSpace(options.Title) == "" {
		return fail("validate", "title requires a value")
	}
	if strings.TrimSpace(options.Grant) == "" {
		return fail("validate", "grant requires a value")
	}
	if len(options.Slug) > 178 || !slugPattern.MatchString(options.Slug) {
		return fail("validate", "slug must be kebab-case and fit in a goal ID")
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var project records.Project
		if err := locked.ReadIn([]string{"projects"}, options.Project, &project); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				envelope = fail("project", "project not found: "+options.Project)
			} else {
				envelope = fail("store", err.Error())
			}
			return nil
		}
		if project.ID != options.Project {
			envelope = fail("project", "project record ID does not match its path")
			return nil
		}
		caller, err := authorize(locked, options.Agent, options.PaneID, project.ID)
		if err != nil {
			envelope = fail("identity", err.Error())
			return nil
		}
		actorID = caller.Key
		number, err := locked.Next("goals")
		if err != nil {
			envelope = fail("counter", err.Error())
			return nil
		}
		counterEffect := output.Effect{Resource: "counter", Path: ".hive/state/counters/goals.json", Action: "updated"}
		log.effect("goal create", project.ID, "", actorID, counterEffect.Resource, counterEffect.Path, counterEffect.Action)
		now := time.Now().UTC().Format(time.RFC3339Nano)
		record := records.Goal{
			ID:        fmt.Sprintf("g%d-%s", number, options.Slug),
			Project:   project.ID,
			Title:     options.Title,
			Grant:     options.Grant,
			Status:    "active",
			Waves:     []records.GoalWave{},
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := locked.WriteIn([]string{goalsCollection}, record.ID, record); err != nil {
			log.failure("goal create", project.ID, record.ID, actorID, "goal", err.Error())
			envelope = output.Envelope{
				Outcome: "partial", Phase: "goal", Effects: []output.Effect{counterEffect},
				Error: err.Error(), Result: nil,
			}
			return nil
		}
		goalEffect := output.Effect{Resource: "goal", Path: goalPath(record.ID), Action: "created"}
		log.effect("goal create", project.ID, record.ID, actorID, goalEffect.Resource, goalEffect.Path, goalEffect.Action)
		log.created(project.ID, record.ID, actorID)
		envelope = output.Envelope{
			Outcome: "ok", Phase: "complete",
			Effects: []output.Effect{counterEffect, goalEffect},
			Result:  record,
		}
		return nil
	})
	if err != nil {
		return fail("store", err.Error())
	}
	return envelope
}

// Edit applies one goal or wave change.
func Edit(home string, options EditOptions) output.Envelope {
	log := openOperationLog(home)
	defer log.close()
	actorID := options.Agent
	fail := func(projectID, phase, message string) output.Envelope {
		log.failure("goal edit", projectID, options.Goal, actorID, phase, message)
		return failed(phase, message)
	}
	if !records.ValidGoalID(options.Goal) {
		return fail("", "validate", "goal not found: "+options.Goal)
	}
	action, err := editAction(options)
	if err != nil {
		return fail("", "validate", err.Error())
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("", "store", err.Error())
	}
	var envelope output.Envelope
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		var record records.Goal
		if err := locked.ReadIn([]string{goalsCollection}, options.Goal, &record); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				envelope = fail("", "read", "goal not found: "+options.Goal)
			} else {
				envelope = fail("", "store", err.Error())
			}
			return nil
		}
		caller, err := authorize(locked, options.Agent, options.PaneID, record.Project)
		if err != nil {
			if caller.Key != "" {
				actorID = caller.Key
			}
			envelope = fail(record.Project, "identity", err.Error())
			return nil
		}
		actorID = caller.Key
		oldStatus := record.Status
		oldWaveStatus := ""
		if action == "wave-status" {
			for _, wave := range record.Waves {
				if wave.ID == options.Wave {
					oldWaveStatus = wave.Status
					break
				}
			}
		}
		if err := applyEdit(locked, &record, options, action); err != nil {
			envelope = fail(record.Project, "edit", err.Error())
			return nil
		}
		record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err := locked.WriteIn([]string{goalsCollection}, record.ID, record); err != nil {
			envelope = fail(record.Project, "store", err.Error())
			return nil
		}
		goalEffect := output.Effect{Resource: "goal", Path: goalPath(record.ID), Action: "updated"}
		log.effect("goal edit", record.Project, record.ID, actorID, goalEffect.Resource, goalEffect.Path, goalEffect.Action)
		var extra []any
		switch action {
		case "status":
			extra = []any{"from_status", oldStatus, "to_status", record.Status}
		case "wave-status":
			extra = []any{"wave_id", options.Wave, "from_status", oldWaveStatus, "to_status", options.WaveStatus}
		case "add-task":
			extra = []any{"wave_id", options.Wave, "task_id", options.AddTask}
		case "add-wave":
			extra = []any{"wave_id", record.Waves[len(record.Waves)-1].ID}
		default:
			extra = []any{"wave_id", options.Wave}
		}
		log.changed(record.Project, record.ID, actorID, action, extra...)
		envelope = output.Envelope{
			Outcome: "ok", Phase: "complete",
			Effects: []output.Effect{goalEffect},
			Result:  record,
		}
		return nil
	})
	if err != nil {
		return fail("", "store", err.Error())
	}
	return envelope
}

// Show reads a goal and derives landed progress from its task records.
func Show(home, id string) output.Envelope {
	log := openOperationLog(home)
	defer log.close()
	fail := func(projectID, phase, message string) output.Envelope {
		log.failure("goal show", projectID, id, "", phase, message)
		return failed(phase, message)
	}
	if !records.ValidGoalID(id) {
		return fail("", "validate", "goal not found: "+id)
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("", "store", err.Error())
	}
	var result ShowResult
	var inner error
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		if err := locked.ReadIn([]string{goalsCollection}, id, &result.Goal); err != nil {
			inner = err
			return nil
		}
		result.WaveProgress = make([]WaveProgress, 0, len(result.Waves))
		for _, wave := range result.Waves {
			progress := WaveProgress{ID: wave.ID, Total: len(wave.Tasks)}
			for _, taskID := range wave.Tasks {
				var task records.Task
				if err := locked.ReadIn([]string{tasksCollection}, taskID, &task); err != nil {
					if errors.Is(err, os.ErrNotExist) {
						continue
					}
					inner = err
					return nil
				}
				if task.Delivery != nil && task.Delivery.Commit != "" && task.Delivery.Pushed {
					progress.Landed++
				}
			}
			result.WaveProgress = append(result.WaveProgress, progress)
		}
		return nil
	})
	if err != nil {
		return fail("", "store", err.Error())
	}
	if inner != nil {
		if errors.Is(inner, os.ErrNotExist) {
			return fail("", "read", "goal not found: "+id)
		}
		return fail(result.Project, "store", inner.Error())
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Result: result}
}

// List returns goals in ID order, with an optional project filter.
func List(home, project string) output.Envelope {
	log := openOperationLog(home)
	defer log.close()
	fail := func(phase, message string) output.Envelope {
		log.failure("goal list", project, "", "", phase, message)
		return failed(phase, message)
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	raw, err := registry.ListIn([]string{goalsCollection}, nil)
	if err != nil {
		return fail("store", err.Error())
	}
	goals := make([]records.Goal, 0, len(raw))
	for _, data := range raw {
		var record records.Goal
		if err := json.Unmarshal(data, &record); err != nil {
			return fail("store", fmt.Sprintf("goal record invalid: %v", err))
		}
		if project == "" || record.Project == project {
			goals = append(goals, record)
		}
	}
	return output.Envelope{Outcome: "ok", Phase: "complete", Effects: []output.Effect{}, Result: goals}
}

func authorize(locked *store.LockedStore, agent, paneID, project string) (records.Agent, error) {
	caller, err := identity.ResolveHeld(locked, agent, paneID)
	if err != nil {
		return records.Agent{}, err
	}
	if caller.Kind == records.AgentKindCoordinator || caller.Owns == project {
		return caller, nil
	}
	return caller, fmt.Errorf("not the owner of project %s or the coordinator: %s", project, oneLine(caller.Name))
}

func editAction(options EditOptions) (string, error) {
	count := 0
	action := ""
	set := func(enabled bool, name string) {
		if enabled {
			count++
			action = name
		}
	}
	set(options.AddWaveSet, "add-wave")
	set(options.AddTaskSet, "add-task")
	set(options.NeedsUserSet, "needs-user")
	set(options.ClearNeedsUser, "clear-needs-user")
	set(options.WaveStatusSet, "wave-status")
	set(options.StatusSet, "status")
	if count != 1 {
		return "", fmt.Errorf("goal edit requires exactly one change")
	}
	switch action {
	case "add-wave":
		if strings.TrimSpace(options.AddWave) == "" || !doneWhenValues[options.DoneWhen] || options.Wave != "" {
			return "", fmt.Errorf("--add-wave needs a title and a valid --done-when value")
		}
	case "add-task", "needs-user", "clear-needs-user", "wave-status":
		if !validWaveID(options.Wave) {
			return "", fmt.Errorf("--wave requires a wave ID from the goal")
		}
		if action == "add-task" && !records.ValidTaskID(options.AddTask) {
			return "", fmt.Errorf("--add-task requires a valid task ID")
		}
		if action == "needs-user" && strings.TrimSpace(options.NeedsUser) == "" {
			return "", fmt.Errorf("--needs-user requires text")
		}
		if action == "wave-status" && options.WaveStatus != "open" && options.WaveStatus != "done" {
			return "", fmt.Errorf("wave status must be open or done")
		}
	case "status":
		if options.Status != "active" && options.Status != "done" {
			return "", fmt.Errorf("goal status must be active or done")
		}
	}
	return action, nil
}

func applyEdit(locked *store.LockedStore, record *records.Goal, options EditOptions, action string) error {
	switch action {
	case "add-wave":
		if record.Waves == nil {
			record.Waves = []records.GoalWave{}
		}
		record.Waves = append(record.Waves, records.GoalWave{
			ID: waveID(len(record.Waves) + 1), Title: options.AddWave,
			DoneWhen: options.DoneWhen, Tasks: []string{}, Status: "open",
		})
	case "status":
		record.Status = options.Status
	default:
		index := -1
		for i := range record.Waves {
			if record.Waves[i].ID == options.Wave {
				index = i
				break
			}
		}
		if index < 0 {
			return fmt.Errorf("wave not found: %s", options.Wave)
		}
		wave := &record.Waves[index]
		switch action {
		case "add-task":
			var task records.Task
			if err := locked.ReadIn([]string{tasksCollection}, options.AddTask, &task); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("task not found: %s", options.AddTask)
				}
				return err
			}
			if task.Project != record.Project {
				return fmt.Errorf("task is not in goal project %s: %s", record.Project, task.ID)
			}
			for _, existing := range record.Waves {
				for _, taskID := range existing.Tasks {
					if taskID == task.ID {
						return fmt.Errorf("task already belongs to wave %s", existing.ID)
					}
				}
			}
			wave.Tasks = append(wave.Tasks, task.ID)
		case "needs-user":
			wave.NeedsUser = options.NeedsUser
		case "clear-needs-user":
			wave.NeedsUser = ""
		case "wave-status":
			wave.Status = options.WaveStatus
		}
	}
	return nil
}

func validWaveID(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func waveID(number int) string {
	var result string
	for number > 0 {
		number--
		result = string(rune('A'+number%26)) + result
		number /= 26
	}
	return result
}

func goalPath(id string) string { return ".hive/state/goals/" + id + ".json" }

func failed(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}
