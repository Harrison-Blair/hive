package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/ask"
	"github.com/charmbracelet/x/ansi"
)

// Golden-frames decision 5: a resize never breaks a border. A long asker
// and four full headers do not fit the top border at 80 columns. Task 137
// repair F1, from the verifier probe.
func TestAskResizeKeepsTheTopRightCorner(t *testing.T) {
	record := oneQuestionAsk("a1", "manager-project-with-a-long-name")
	record.Questions = nil
	for i := range 4 {
		record.Questions = append(record.Questions, ask.Question{
			Question: "Which option?", Header: fmt.Sprintf("Question%04d", i),
			Options: []ask.Option{askOption("Yes", "accept"), askOption("No", "decline")},
		})
	}
	data, err := json.Marshal(ask.Request{Questions: record.Questions})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ask.Parse(data); err != nil {
		t.Fatalf("invalid ask: %v", err)
	}
	m, _ := askModel(t, 120, 40, record)
	checkTop := func(width int) {
		t.Helper()
		found := false
		for _, row := range viewRows(m) {
			if !strings.HasPrefix(row, "╭─ ?") {
				continue
			}
			found = true
			if !strings.HasSuffix(row, "╮") || ansi.StringWidth(row) != width {
				t.Fatalf("at %d columns the ask top border is %q; want %d cells that end with ╮", width, row, width)
			}
		}
		if !found {
			t.Fatalf("at %d columns no ask top border", width)
		}
	}
	checkTop(120)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	checkTop(80)
}
