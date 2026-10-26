// Package show reads a goal and derives landed progress from its task records.
package show

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	goallib "github.com/Harrison-Blair/hive/internal/goal"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Result includes wave progress from task delivery records.
type Result struct {
	records.Goal
	WaveProgress []WaveProgress `json:"wave_progress"`
}

// WaveProgress reports the landed task count for one wave.
type WaveProgress struct {
	ID     string `json:"id"`
	Landed int    `json:"landed"`
	Total  int    `json:"total"`
}

// Run reads a goal and derives landed progress from its task records.
func Run(home, id string) output.Envelope {
	log := goallib.OpenLog(home)
	defer log.Close()
	fail := func(projectID, phase, message string) output.Envelope {
		log.Failure("goal show", projectID, id, "", phase, message)
		return goallib.Failed(phase, message)
	}
	if !records.ValidGoalID(id) {
		return fail("", "validate", "goal not found: "+id)
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("", "store", err.Error())
	}
	var result Result
	var inner error
	err = registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		if err := locked.ReadIn([]string{goallib.GoalsCollection}, id, &result.Goal); err != nil {
			inner = err
			return nil
		}
		result.WaveProgress = make([]WaveProgress, 0, len(result.Waves))
		for _, wave := range result.Waves {
			progress := WaveProgress{ID: wave.ID, Total: len(wave.Tasks)}
			for _, taskID := range wave.Tasks {
				var task records.Task
				if err := locked.ReadIn([]string{goallib.TasksCollection}, taskID, &task); err != nil {
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

// Text renders a goal and its wave progress.
func Text(envelope output.Envelope) string {
	result, ok := envelope.Result.(Result)
	if !ok || envelope.Error != nil {
		return ""
	}
	goal := result.Goal
	lines := []string{
		"id: " + goal.ID,
		"project: " + goal.Project,
		"title: " + goallib.OneLine(goal.Title),
		"grant: " + goallib.OneLine(goal.Grant),
		"status: " + goal.Status,
	}
	for i, wave := range goal.Waves {
		progress := result.WaveProgress[i]
		lines = append(lines,
			fmt.Sprintf("wave %s: %s", wave.ID, goallib.OneLine(wave.Title)),
			"done_when: "+wave.DoneWhen,
			"tasks: "+taskList(wave.Tasks),
			"landed: "+fmt.Sprintf("%d/%d landed", progress.Landed, progress.Total),
			"status: "+wave.Status,
		)
		if wave.NeedsUser != "" {
			lines = append(lines, "needs_user: "+goallib.OneLine(wave.NeedsUser))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func taskList(tasks []string) string {
	if len(tasks) == 0 {
		return "-"
	}
	return strings.Join(tasks, ",")
}
