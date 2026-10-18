package worktree

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// CreateText renders the result and effects of one worktree create.
func CreateText(envelope output.Envelope) string {
	lines := []string{}
	if result, ok := envelope.Result.(CreateResult); ok {
		if result.WorkspaceID != "" {
			lines = append(lines, "workspace: "+result.WorkspaceID)
		}
		if result.TabID != "" {
			lines = append(lines, "tab: "+result.TabID)
		}
		if result.RootPane != "" {
			lines = append(lines, "root_pane: "+result.RootPane)
		}
		if result.Path != "" {
			lines = append(lines, "path: "+result.Path)
		}
		if result.Branch != "" {
			lines = append(lines, "branch: "+result.Branch)
		}
	}
	lines = append(lines, effectLines(envelope.Effects)...)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// RemoveText renders the effects and the kept branch of one worktree remove.
func RemoveText(envelope output.Envelope) string {
	lines := effectLines(envelope.Effects)
	if result, ok := envelope.Result.(RemoveResult); ok && result.Branch != "" {
		lines = append(lines, "branch kept: "+result.Branch)
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func effectLines(effects []output.Effect) []string {
	lines := make([]string, 0, len(effects))
	for _, effect := range effects {
		switch {
		case effect.Resource == "workspace" && effect.Action == "created":
			lines = append(lines, "workspace created: "+effect.Path)
		case effect.Resource == "workspace" && effect.Action == "opened":
			lines = append(lines, "repository workspace opened: "+effect.Path)
		case effect.Resource == "workspace" && effect.Action == "closed":
			lines = append(lines, "workspace closed: "+effect.Path)
		case effect.Resource == "worktree" && effect.Action == "created":
			lines = append(lines, "worktree created: "+effect.Path)
		case effect.Resource == "worktree" && effect.Action == "removed":
			lines = append(lines, "worktree removed: "+effect.Path)
		case effect.Resource == "branch" && effect.Action == "created":
			lines = append(lines, "branch created: "+effect.Path)
		default:
			lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
		}
	}
	return lines
}
