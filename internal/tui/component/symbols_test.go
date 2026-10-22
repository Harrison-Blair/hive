package component

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestSymbolWidth checks that each symbol of decision 13 takes one cell.
func TestSymbolWidth(t *testing.T) {
	symbols := strings.Fields("─ │ ╭ ╮ ╰ ╯ ├ ┤ ┼ ┴ ┬ └ ━ ┃ ┏ ┓ ┗ ┛ ┯ ┷ ┡ ┨ ● ◐ ○ ▶ ✓ • · – — … ↑ ↓ ← → › █ × ▸ ▾ ▲ ┰ ┸ ┺ ┹ ╃ ┻ ◀ ▼")
	if len(symbols) != 50 {
		t.Fatalf("decision 13 lists 50 symbols, the test has %d", len(symbols))
	}
	for _, s := range symbols {
		if w := lipgloss.Width(s); w != 1 {
			t.Errorf("lipgloss.Width(%q) = %d, want 1", s, w)
		}
	}
}
