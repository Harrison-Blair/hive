// Package edit applies one goal or wave change.
package edit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	goallib "github.com/Harrison-Blair/hive/internal/goal"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

var doneWhenValues = map[string]bool{
	"local":   true,
	"dev":     true,
	"pr":      true,
	"main":    true,
	"release": true,
}

// Options holds the values for one goal edit.
type Options struct {
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

// Run applies one goal or wave change.
func Run(home string, options Options) output.Envelope {
	log := goallib.OpenLog(home)
	defer log.Close()
	actorID := options.Agent
	fail := func(projectID, phase, message string) output.Envelope {
		log.Failure("goal edit", projectID, options.Goal, actorID, phase, message)
		return goallib.Failed(phase, message)
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
		if err := locked.ReadIn([]string{goallib.GoalsCollection}, options.Goal, &record); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				envelope = fail("", "read", "goal not found: "+options.Goal)
			} else {
				envelope = fail("", "store", err.Error())
			}
			return nil
		}
		caller, err := goallib.Authorize(locked, options.Agent, options.PaneID, record.Project)
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
		if err := locked.WriteIn([]string{goallib.GoalsCollection}, record.ID, record); err != nil {
			envelope = fail(record.Project, "store", err.Error())
			return nil
		}
		goalEffect := output.Effect{Resource: "goal", Path: goallib.GoalPath(record.ID), Action: "updated"}
		log.Effect("goal edit", record.Project, record.ID, actorID, goalEffect.Resource, goalEffect.Path, goalEffect.Action)
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
		log.Changed(record.Project, record.ID, actorID, action, extra...)
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

// Text renders the effects of one goal edit.
func Text(envelope output.Envelope) string { return goallib.EffectsText(envelope.Effects) }

func editAction(options Options) (string, error) {
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

func applyEdit(locked *store.LockedStore, record *records.Goal, options Options, action string) error {
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
			if err := locked.ReadIn([]string{goallib.TasksCollection}, options.AddTask, &task); err != nil {
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
