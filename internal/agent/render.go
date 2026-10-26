package agent

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// EffectsText renders one line for each effect.
func EffectsText(envelope output.Envelope) string {
	lines := make([]string, 0, len(envelope.Effects))
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
