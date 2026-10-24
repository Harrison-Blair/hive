package agent

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// RegisterText renders the effects of one registration.
func RegisterText(envelope output.Envelope) string {
	return effectsText(envelope)
}

// ListText renders one line for each agent record.
func ListText(envelope output.Envelope) string {
	agents, ok := envelope.Result.([]records.Agent)
	if !ok || envelope.Error != nil {
		return ""
	}
	if len(agents) == 0 {
		return "no agents\n"
	}
	lines := make([]string, 0, len(agents))
	for _, record := range agents {
		lines = append(lines, fmt.Sprintf("%s\t%d\t%s\t%s\t%s\t%s\t%s", record.Name, record.Generation, record.MirrorRole(), record.Status, record.Project, record.Task, record.PaneID))
	}
	return strings.Join(lines, "\n") + "\n"
}

// ShowText renders each field of one agent record.
func ShowText(envelope output.Envelope) string {
	record, ok := envelope.Result.(records.Agent)
	if !ok || envelope.Error != nil {
		return ""
	}
	text := fmt.Sprintf("key: %s\nname: %s\nprofile: %s\nkind: %s\nrole: %s\nworktree: %s\nowns: %s\nproject: %s\ntask: %s\nharness: %s\nmodel: %s\npane_id: %s\nworkspace_id: %s\ntab_id: %s\nterminal_id: %s\ncwd: %s\nstatus: %s\nfinished: %t\nquiet_after_min: %d\ngeneration: %d\nstarted_at: %s\nupdated_at: %s\n",
		record.Key, record.Name, record.Profile, record.Kind, record.MirrorRole(), record.Worktree, record.Owns,
		record.Project, record.Task, record.Harness, record.Model,
		record.PaneID, record.WorkspaceID, record.TabID, record.TerminalID, record.Cwd, record.Status,
		record.Finished, record.QuietAfterMin, record.Generation, record.StartedAt, record.UpdatedAt)
	if record.Parent != "" {
		text += fmt.Sprintf("parent: %s\n", record.Parent)
	}
	if record.Replaces != "" {
		text += fmt.Sprintf("replaces: %s\n", record.Replaces)
	}
	if record.StatusChangedAt != "" {
		text += fmt.Sprintf("status since: %s\n", record.StatusChangedAt)
	}
	return text
}

func effectsText(envelope output.Envelope) string {
	lines := make([]string, 0, len(envelope.Effects))
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
