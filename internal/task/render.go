package task

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// FailureReport renders the known effects and the inspection commands of one
// finish failure. It never replays the finish group.
func FailureReport(envelope output.Envelope, taskID, agent string) string {
	project := ""
	if result, ok := envelope.Result.(records.Task); ok {
		taskID = result.ID
		project = result.Project
		if agent == "" {
			if result.Verification != nil && result.Verification.Agent != "" {
				agent = result.Verification.Agent
			} else {
				agent = result.ActingAgent
			}
		}
	}
	lines := []string{
		"outcome: " + envelope.Outcome,
		"phase: " + envelope.Phase,
	}
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("effect: %s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if envelope.Error != nil {
		lines = append(lines, "error: "+OneLine(fmt.Sprint(envelope.Error)))
	}
	if taskID != "" {
		lines = append(lines, "inspect: hive task show "+taskID)
	}
	if agent != "" {
		lines = append(lines, "inspect: hive agent show "+OneLine(agent))
	}
	if project != "" {
		lines = append(lines, "inspect: hive inbox --inbox "+ProjectAddress(project)+" --all")
	} else {
		lines = append(lines, "inspect: hive inbox")
	}
	return strings.Join(lines, "\n") + "\n"
}

// TransitionText renders one transition result.
func TransitionText(envelope output.Envelope) string {
	record, ok := envelope.Result.(records.Task)
	if !ok || envelope.Error != nil {
		return ""
	}
	return fmt.Sprintf("id: %s\nstatus: %s\n", record.ID, record.Status)
}

// EffectsText renders each completed effect.
func EffectsText(effects []output.Effect) string {
	lines := make([]string, 0, len(effects))
	for _, effect := range effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// QuoteIfControl quotes a value that holds a control character.
func QuoteIfControl(value string) string {
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Sprintf("%q", value)
		}
	}
	return value
}

// OneLine replaces each control character with a space for one-line output.
func OneLine(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
}
