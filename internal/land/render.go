package land

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// LandText renders the delivery and the effects of one landing.
func LandText(envelope output.Envelope) string {
	lines := []string{}
	if result, ok := envelope.Result.(Result); ok {
		if result.Commit != "" {
			lines = append(lines, "commit: "+result.Commit)
		}
		if result.Mode != "" {
			lines = append(lines, "mode: "+result.Mode)
		}
		if result.Target != "" {
			lines = append(lines, "target: "+result.Target)
		}
		if result.Pushed {
			lines = append(lines, "pushed: true")
		}
		if result.PRURL != "" {
			lines = append(lines, "pr_url: "+result.PRURL)
		}
		if result.Note != "" {
			lines = append(lines, "note: "+result.Note)
		}
	}
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
