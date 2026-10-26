// Package show renders the agent record of the parent agent.Show API.
package show

import (
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// Text renders each field of one agent record.
func Text(envelope output.Envelope) string {
	record, ok := envelope.Result.(records.Agent)
	if !ok || envelope.Error != nil {
		return ""
	}
	text := fmt.Sprintf("key: %s\nname: %s\nprofile: %s\nkind: %s\nworktree: %s\nowns: %s\nproject: %s\ntask: %s\nharness: %s\nmodel: %s\npane_id: %s\nworkspace_id: %s\ntab_id: %s\nterminal_id: %s\ncwd: %s\nstatus: %s\nfinished: %t\nquiet_after_min: %d\ngeneration: %d\nstarted_at: %s\nupdated_at: %s\n",
		record.Key, record.Name, record.Profile, record.Kind, record.Worktree, record.Owns,
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
