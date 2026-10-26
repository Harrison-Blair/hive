// Package list returns goals in ID order.
package list

import (
	"encoding/json"
	"fmt"
	"strings"

	goallib "github.com/Harrison-Blair/hive/internal/goal"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Run returns goals in ID order, with an optional project filter.
func Run(home, project string) output.Envelope {
	log := goallib.OpenLog(home)
	defer log.Close()
	fail := func(phase, message string) output.Envelope {
		log.Failure("goal list", project, "", "", phase, message)
		return goallib.Failed(phase, message)
	}
	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	raw, err := registry.ListIn([]string{goallib.GoalsCollection}, nil)
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

// Text renders a goal list.
func Text(envelope output.Envelope) string {
	goals, ok := envelope.Result.([]records.Goal)
	if !ok || envelope.Error != nil {
		return ""
	}
	if len(goals) == 0 {
		return "no goals\n"
	}
	lines := make([]string, 0, len(goals))
	for _, goal := range goals {
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", goal.ID, goal.Status, goal.Project, goallib.OneLine(goal.Title)))
	}
	return strings.Join(lines, "\n") + "\n"
}
