package tui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/ask"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// Names at the 32-character limit of the agent name pattern.
const (
	longAsker  = "asker-scratch-with-long-name-abc"
	longParent = "parent-scratch-with-long-name-ab"
)

// stoppedAskModel gives a model at width x height with the accepted request
// of the third verification, from a stopped asker with a live parent. Later
// chat messages move the ask event above the feed viewport.
func stoppedAskModel(t *testing.T, width, height int, asker, parent string) (*Model, *[]askCall) {
	t.Helper()
	data, err := os.ReadFile("../../.hive/spec/tui-v2/tasks/135-ask-confirm/evidence/verifier3/no-caption/request.json")
	if err != nil {
		t.Fatal(err)
	}
	request, err := ask.Parse(data)
	if err != nil {
		t.Fatalf("ask.Parse: %v", err)
	}
	record := oneQuestionAsk("a1", asker)
	record.Parent, record.ParentKey = parent, parent
	record.Questions = request.Questions
	m, calls := askModel(t, width, height, record)
	agents := []records.Agent{
		{Key: asker, Name: asker, Status: records.AgentStopped},
		{Key: parent, Name: parent, Status: records.AgentLive},
	}
	var entries []chatlib.Entry
	for index := 1; index <= 30; index++ {
		entries = append(entries, chatlib.Entry{
			ID: fmt.Sprintf("c%d", index), At: fmt.Sprintf("2026-10-20T11:40:%02dZ", index),
			From: "coordinator", Text: fmt.Sprintf("Later chat message %d", index),
		})
	}
	m.chatData = buildChatScreenData(entries, nil, []ask.Record{record}, agents, m.now())
	m.rebuildChat()
	m.invalidateView()
	return m, calls
}

// askCardText gives the plain rows of the ask card, from its top border to
// its bottom border.
func askCardText(t *testing.T, m *Model) string {
	t.Helper()
	var card []string
	inside := false
	for line := range strings.SplitSeq(ansi.Strip(m.View().Content), "\n") {
		if strings.HasPrefix(line, "┏") {
			inside = true
		}
		if inside {
			card = append(card, line)
		}
		if strings.HasPrefix(line, "┗") {
			inside = false
		}
	}
	return strings.Join(card, "\n")
}

// requireCard fails when the card lacks one of the texts.
func requireCard(t *testing.T, m *Model, texts ...string) {
	t.Helper()
	card := askCardText(t, m)
	for _, text := range texts {
		if !strings.Contains(card, text) {
			t.Fatalf("card lacks %q:\n%s", text, card)
		}
	}
}

// AC1 and decision 114: the stopped review card keeps the full title, an
// explicit asker stopped state, the full parent in Send, and the parent in
// the hint at both widths, for names up to the 32-character limit. The state
// stays in the card when the ask event leaves the feed viewport, through an
// edit, the return to review, and a resize. Send keeps the parent route.
func TestAskStoppedReviewKeepsTheStateInTheCard(t *testing.T) {
	type view struct {
		caption string
		hint    string
	}
	for _, tc := range []struct {
		name, asker, parent string
		wide, compact       view
	}{
		{
			name: "full caption at both widths", asker: "t1", parent: "p1",
			wide:    view{"━ asker stopped · answer goes to p1 ━┓", "┗━ enter sends to p1 · ↑↓ question to change · x drop ━"},
			compact: view{"━ asker stopped · answer goes to p1 ━┓", "┗━ enter sends to p1 · ↑↓ question to change · x drop ━"},
		},
		{
			name: "short caption at 80", asker: "asker-scratch", parent: "parent-scratch",
			wide:    view{"━ asker stopped · answer goes to parent-scratch ━┓", "┗━ enter sends to parent-scratch · ↑↓ question to change · x drop ━"},
			compact: view{"━ asker stopped · to parent-scratch ━┓", "┗━ enter sends to parent-scratch · ↑↓ question to change · x drop ━"},
		},
		{
			name: "no route caption at 80", asker: longAsker, parent: "parent-scratch",
			wide:    view{"━ asker stopped · answer goes to parent-scratch ━┓", "┗━ enter sends to parent-scratch · ↑↓ question to change · x drop ━"},
			compact: view{"━ asker stopped ━┓", "┗━ enter sends to parent-scratch · ↑↓ question to change · x drop ━"},
		},
		{
			name: "short caption at 120 and no route caption at 80", asker: longAsker, parent: longParent,
			wide:    view{"━ asker stopped · to " + longParent + " ━┓", "┗━ enter sends to " + longParent + " · ↑↓ question to change · x drop ━"},
			compact: view{"━ asker stopped ━┓", "┗━ enter sends to " + longParent + " · ↑↓ change · x drop ━"},
		},
	} {
		for _, start := range []demoSize{{120, 40}, {80, 24}} {
			t.Run(fmt.Sprintf("%s/%dx%d", tc.name, start.Width, start.Height), func(t *testing.T) {
				sizes := map[int]view{120: tc.wide, 80: tc.compact}
				other := demoSize{80, 24}
				if start.Width == 80 {
					other = demoSize{120, 40}
				}
				check := func(m *Model, width, height int) {
					t.Helper()
					requireFits(t, m, width, height)
					requireCard(t, m, "┏━ ? "+tc.asker+" asks · confirm ━", sizes[width].caption,
						"┃ › Send to "+tc.parent+" ", sizes[width].hint)
				}
				m, calls := stoppedAskModel(t, start.Width, start.Height, tc.asker, tc.parent)
				if strings.Contains(ansi.Strip(m.View().Content), "? asked · Change") {
					t.Fatal("the ask event is in the feed viewport; the check needs it above")
				}
				m = run(t, m, ctrl('g'))
				requireNoCalls(t, calls, "the pick")
				check(m, start.Width, start.Height)

				m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
				m = run(t, m, enterKey)
				m = run(t, m, escKey())
				requireNoCalls(t, calls, "the edit")
				requireCard(t, m, "Change → No")
				check(m, start.Width, start.Height)

				m, _ = raise(t, m, tea.WindowSizeMsg{Width: other.Width, Height: other.Height})
				requireCard(t, m, "Change → No")
				check(m, other.Width, other.Height)

				run(t, m, enterKey)
				requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{2}}}}})
			})
		}
	}
}

// The editor card of a stopped asker keeps the state when no route caption
// fits at 80 columns.
func TestAskStoppedEditorKeepsTheStateInTheCard(t *testing.T) {
	m, _ := stoppedAskModel(t, 80, 24, longAsker, "parent-scratch")
	requireFits(t, m, 80, 24)
	requireCard(t, m, "┏━ ? "+longAsker+" asks · Change ━", "━ asker stopped ━┓")
}
