package worktree

import (
	"fmt"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// EffectLines renders one text line for each effect.
func EffectLines(effects []output.Effect) []string {
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
