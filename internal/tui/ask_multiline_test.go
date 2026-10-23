package tui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/ask"
	"github.com/charmbracelet/x/ansi"
)

// requireReviewDock fails when a dock body row lacks its left or right
// border, or when the dock has more than limit rows. It gives the body rows.
func requireReviewDock(t *testing.T, m *Model, limit int) []string {
	t.Helper()
	var body []string
	inside, rows := false, 0
	for line := range strings.SplitSeq(ansi.Strip(m.View().Content), "\n") {
		switch {
		case strings.HasPrefix(line, "┏"):
			inside, rows = true, 1
		case strings.HasPrefix(line, "┗"):
			inside, rows = false, rows+1
		case inside:
			rows++
			if !strings.HasPrefix(line, "┃") || !strings.HasSuffix(line, "┃") {
				t.Fatalf("dock row has a broken border: %q", line)
			}
			body = append(body, line)
		}
	}
	if rows > limit {
		t.Fatalf("dock has %d rows, want at most %d", rows, limit)
	}
	return body
}

// multilineAsk gives the record of an accepted request file in the
// evidence of the first verification.
func multilineAsk(t *testing.T, name string) ask.Record {
	t.Helper()
	data, err := os.ReadFile("../../.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/verifier/multiline/" + name)
	if err != nil {
		t.Fatal(err)
	}
	request, err := ask.Parse(data)
	if err != nil {
		t.Fatalf("ask.Parse: %v", err)
	}
	record := oneQuestionAsk("a1", "manager-hive")
	record.Questions = request.Questions
	return record
}

// AC2: an accepted label, note, and Other answer with line feeds show each
// line on its own row inside the dock borders at both widths. The reply
// keeps the line feeds.
func TestAskReviewKeepsMultilineTextInsideTheDock(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}} {
		t.Run(fmt.Sprint(size[0]), func(t *testing.T) {
			record := multilineAsk(t, "short-request.json")
			record.Questions = append(record.Questions, ask.Question{
				Question: "Which other?", Header: "Other", MultiSelect: true,
				Options: []ask.Option{askOption("Plain", "one line")},
			})
			m, calls := askModel(t, size[0], size[1], record)
			m, _ = raise(t, m, ctrl('g'))
			m, _ = raise(t, m, escKey())
			m.askCards["a1"].others[1] = "other one\nother two"
			m.askCards["a1"].notes[1] = "note one\nnote two"
			m = run(t, m, enterKey)
			requireNoCalls(t, calls, "the review")
			requireView(t, m, "asks · confirm", "┃   Label → First line ", "┃   Second line ",
				"┃   Other → other one ", "┃   other two · note: note one ", "┃   note two ", "┃ › Send ")
			requireFits(t, m, size[0], size[1])
			requireReviewDock(t, m, size[1]/2)
			run(t, m, enterKey)
			requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{
				{Picks: []int{1}}, {Picks: []int{2}, Other: "other one\nother two", Note: "note one\nnote two"},
			}}})
		})
	}
}

// AC2: an accepted label of 31 lines stays inside the dock cap at both
// widths, and the arrows reach each of its lines with no cut.
func TestAskReviewReachesEachLineOfATallLabel(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}} {
		t.Run(fmt.Sprint(size[0]), func(t *testing.T) {
			m, calls := askModel(t, size[0], size[1], multilineAsk(t, "tall-request.json"))
			m = run(t, m, ctrl('g'))
			requireNoCalls(t, calls, "the review")
			requireView(t, m, "┃ › Send ")
			seen := map[string]bool{}
			look := func() {
				requireFits(t, m, size[0], size[1])
				for _, line := range requireReviewDock(t, m, size[1]/2) {
					if strings.Contains(line, "…") {
						t.Fatalf("dock row has a cut: %q", line)
					}
				}
				for _, word := range strings.Fields(cardText(m)) {
					seen[word] = true
				}
			}
			look()
			for range 12 {
				m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
				look()
			}
			requireView(t, m, "┃ › Label → ROW01 ")
			for range 12 {
				m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
				look()
			}
			requireView(t, m, "┃ › Send ")
			for row := 1; row <= 30; row++ {
				if word := fmt.Sprintf("ROW%02d", row); !seen[word] {
					t.Fatalf("%q never showed", word)
				}
			}
			if !seen["FINAL-LABEL-SENTINEL"] {
				t.Fatal("FINAL-LABEL-SENTINEL never showed")
			}
			run(t, m, enterKey)
			requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1}}}}})
		})
	}
}
