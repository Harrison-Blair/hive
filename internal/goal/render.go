package goal

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// CreateText renders the new goal ID.
func CreateText(envelope output.Envelope) string {
	if record, ok := envelope.Result.(records.Goal); ok && envelope.Error == nil {
		return fmt.Sprintf("id: %s\n", record.ID)
	}
	return effectsText(envelope.Effects)
}

// EditText renders the effects of one goal edit.
func EditText(envelope output.Envelope) string { return effectsText(envelope.Effects) }

// ShowText renders a goal and its wave progress.
func ShowText(envelope output.Envelope) string {
	result, ok := envelope.Result.(ShowResult)
	if !ok || envelope.Error != nil {
		return ""
	}
	goal := result.Goal
	lines := []string{
		"id: " + goal.ID,
		"project: " + goal.Project,
		"title: " + oneLine(goal.Title),
		"grant: " + oneLine(goal.Grant),
		"status: " + goal.Status,
	}
	for i, wave := range goal.Waves {
		progress := result.WaveProgress[i]
		lines = append(lines,
			fmt.Sprintf("wave %s: %s", wave.ID, oneLine(wave.Title)),
			"done_when: "+wave.DoneWhen,
			"tasks: "+taskList(wave.Tasks),
			"landed: "+fmt.Sprintf("%d/%d landed", progress.Landed, progress.Total),
			"status: "+wave.Status,
		)
		if wave.NeedsUser != "" {
			lines = append(lines, "needs_user: "+oneLine(wave.NeedsUser))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// ListText renders a goal list.
func ListText(envelope output.Envelope) string {
	goals, ok := envelope.Result.([]records.Goal)
	if !ok || envelope.Error != nil {
		return ""
	}
	if len(goals) == 0 {
		return "no goals\n"
	}
	lines := make([]string, 0, len(goals))
	for _, goal := range goals {
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", goal.ID, goal.Status, goal.Project, oneLine(goal.Title)))
	}
	return strings.Join(lines, "\n") + "\n"
}

func effectsText(effects []output.Effect) string {
	lines := make([]string, 0, len(effects))
	for _, effect := range effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func taskList(tasks []string) string {
	if len(tasks) == 0 {
		return "-"
	}
	return strings.Join(tasks, ",")
}

func oneLine(value string) string { return strings.Join(strings.Fields(value), " ") }
