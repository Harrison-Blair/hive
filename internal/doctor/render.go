package doctor

import (
	"fmt"
	"strings"
)

// Render writes one line for each check.
func Render(result Result) string {
	var builder strings.Builder
	for _, check := range result.Checks {
		fmt.Fprintf(&builder, "%s %s: %s\n", check.Status, check.Name, check.Detail)
	}
	return builder.String()
}
