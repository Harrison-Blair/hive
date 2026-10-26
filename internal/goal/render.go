package goal

import (
	"fmt"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// EffectsText renders one line for each effect.
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

// OneLine joins the words of a value with single spaces.
func OneLine(value string) string { return strings.Join(strings.Fields(value), " ") }
