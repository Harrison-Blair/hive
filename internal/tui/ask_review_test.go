package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/ask"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

var enterKey = tea.KeyPressMsg{Code: tea.KeyEnter}

// requireNoCalls fails when the card sent an answer or a dismiss.
func requireNoCalls(t *testing.T, calls *[]askCall, when string) {
	t.Helper()
	if len(*calls) != 0 {
		t.Fatalf("%s: calls = %v, want none", when, *calls)
	}
}

// requireCalls fails when the calls are not want.
func requireCalls(t *testing.T, calls *[]askCall, want []askCall) {
	t.Helper()
	if fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

// rowOf gives the view row that contains text, or -1.
func rowOf(m *Model, text string) int {
	for index, line := range strings.Split(ansi.Strip(m.View().Content), "\n") {
		if strings.Contains(line, text) {
			return index
		}
	}
	return -1
}

// AC1: a card with one question shows the review after the pick. The review
// writes no answer. enter on Send sends it.
func TestAskReviewOneQuestion(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m = inEastern(m)
	m = run(t, m, key('2', "2"))
	requireNoCalls(t, calls, "the pick")
	requireView(t, m,
		"╭─ ? manager-hive asks · confirm ─",
		" open · 1 of 1 ╮",
		"│   Model → No ",
		"│ › Send ",
		"? manager-hive 07:30  asked · Model · in the composer, ^r to answer",
	)
	rejectView(t, m, "1 [ ] Yes", "2 [x] No")
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · review · ↑↓ question · enter send · x dismiss · J/K card · ^r input") {
		t.Fatalf("footer = %q", got)
	}
	run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{2}}}}})
}

// AC1: each path that completes the last answer opens the review and sends
// nothing: enter, a digit, a click, enter on a note, and enter on Other.
func TestAskEachCompletionPathOpensTheReview(t *testing.T) {
	for _, tc := range []struct {
		name   string
		keys   func(t *testing.T, m *Model) *Model
		row    string
		answer ask.Reply
	}{
		{"enter", func(t *testing.T, m *Model) *Model {
			m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
			return run(t, m, enterKey)
		}, "Model → No ", ask.Reply{Picks: []int{2}}},
		{"digit", func(t *testing.T, m *Model) *Model {
			return run(t, m, key('3', "3"))
		}, "Model → Later ", ask.Reply{Picks: []int{3}}},
		{"click", func(t *testing.T, m *Model) *Model {
			return run(t, m, click(12, rowOf(m, "2 [ ] No")))
		}, "Model → No ", ask.Reply{Picks: []int{2}}},
		{"note", func(t *testing.T, m *Model) *Model {
			m, _ = raise(t, m, key('n', "n"))
			for _, r := range "ok" {
				m, _ = raise(t, m, key(r, string(r)))
			}
			return run(t, m, enterKey)
		}, "Model → Yes · note: ok ", ask.Reply{Picks: []int{1}, Note: "ok"}},
		{"other", func(t *testing.T, m *Model) *Model {
			m, _ = raise(t, m, key('4', "4"))
			for _, r := range "soon" {
				m, _ = raise(t, m, key(r, string(r)))
			}
			return run(t, m, enterKey)
		}, "Model → soon ", ask.Reply{Picks: []int{4}, Other: "soon"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
			m = tc.keys(t, m)
			requireNoCalls(t, calls, tc.name)
			requireView(t, m, "? manager-hive asks · confirm", "│   "+tc.row, "│ › Send ")
			if m.chatData.asks[0].State != ask.StateOpen {
				t.Fatalf("state = %q, want open", m.chatData.asks[0].State)
			}
			run(t, m, enterKey)
			requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{tc.answer}}})
		})
	}
}

// AC1: an empty Other answer keeps its text line and opens no review.
func TestAskReviewKeepsTheEmptyOtherCheck(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m, _ = raise(t, m, key('4', "4"))
	m = run(t, m, enterKey)
	requireNoCalls(t, calls, "the empty Other")
	rejectView(t, m, "asks · confirm")
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · other ·") {
		t.Fatalf("footer = %q", got)
	}
}

// AC2: the review shows every question with its full picks and note. ↑↓
// highlights a question. enter on it opens its options with the picks. The
// edit returns to the review with Send selected, and the answer has the
// edited draft.
func TestAskReviewEditsAQuestionAndSendsTheEditedDraft(t *testing.T) {
	m, calls := askModel(t, 120, 40, twoQuestionAsk("a1"))
	m, _ = raise(t, m, key('1', "1"))
	m, _ = raise(t, m, key('4', "4"))
	for _, r := range "keys" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, enterKey)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = raise(t, m, key('n', "n"))
	for _, r := range "small" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, enterKey)
	requireNoCalls(t, calls, "the last pick")
	requireView(t, m,
		"╭─ ? coordinator asks · confirm ─",
		"│   Scope → Wheel, keys ",
		"│   Order → Card first · note: small ",
		"│ › Send ",
	)
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	requireView(t, m, "│ › Order → Card first · note: small ", "│   Send ")
	// Decision 33: the review keys stay enter send on a question row.
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · review · ↑↓ question · enter send ·") {
		t.Fatalf("footer = %q", got)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	requireView(t, m, "│ › Scope → Wheel, keys ", "│   Order → Card first · note: small ")
	m, _ = raise(t, m, enterKey)
	requireNoCalls(t, calls, "enter on a question")
	requireView(t, m, "│ 1. Which parts?", "│   1 [x] Wheel", "│ › 4 [x] Other: type an answer", "other › keys")
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · ask ·") {
		t.Fatalf("footer in the edit = %q", got)
	}
	m, _ = raise(t, m, key('2', "2"))
	m, _ = raise(t, m, enterKey)
	requireNoCalls(t, calls, "the edit")
	requireView(t, m, "│   Scope → Wheel, Click, keys ", "│   Order → Card first · note: small ", "│ › Send ")
	run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1, 2, 4}, Other: "keys"}, {Picks: []int{2}, Note: "small"}}}})
}

// AC2: an edit of the note of a question keeps the picks, and the review
// shows the new note.
func TestAskReviewEditsANote(t *testing.T) {
	m, calls := askModel(t, 120, 40, twoQuestionAsk("a1"))
	m, _ = raise(t, m, key('3', "3"))
	m, _ = raise(t, m, enterKey)
	m, _ = raise(t, m, key('1', "1"))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	m, _ = raise(t, m, enterKey)
	m, _ = raise(t, m, key('n', "n"))
	for _, r := range "later" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, enterKey)
	requireNoCalls(t, calls, "the note")
	requireView(t, m, "│   Scope → Modes · note: later ", "│   Order → Modes first ", "│ › Send ")
	run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{3}, Note: "later"}, {Picks: []int{1}}}}})
}

// AC3: in the review, digits, a click, space, tab, text, and backspace do
// not change the answer and send nothing. A digit does not open a screen.
func TestAskReviewIgnoresOtherInput(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m, _ = raise(t, m, key('n', "n"))
	m, _ = raise(t, m, key('o', "o"))
	m = run(t, m, enterKey)
	requireView(t, m, "│   Model → Yes · note: o ", "│ › Send ")
	presses := []tea.Msg{
		key('1', "1"), key('2', "2"), key('3', "3"), key('4', "4"), key('5', "5"),
		key(' ', "space"), tea.KeyPressMsg{Code: tea.KeyTab}, key('a', "a"), key('z', "z"),
		tea.KeyPressMsg{Code: tea.KeyBackspace},
	}
	for _, press := range presses {
		m = run(t, m, press)
	}
	// Each card row; a click on the input line below the ├─┤ rule gives it
	// the focus (rule 17).
	for row, line := range strings.Split(ansi.Strip(m.View().Content), "\n") {
		if strings.HasPrefix(line, "├") {
			break
		}
		if strings.HasPrefix(line, "│") {
			m = run(t, m, click(12, row))
		}
	}
	requireNoCalls(t, calls, "the review input")
	if m.screenIndex != 0 {
		t.Fatalf("a key opened screen %d", m.screenIndex)
	}
	requireView(t, m, "? manager-hive asks · confirm", "│   Model → Yes · note: o ", "│ › Send ")
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · review ·") {
		t.Fatalf("status row = %q", got)
	}
	run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1}, Note: "o"}}}})
}

// AC3: x in the review dismisses the ask, as outside the review.
func TestAskReviewXDismisses(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m = run(t, m, key('1', "1"))
	run(t, m, key('x', "x"))
	requireCalls(t, calls, []askCall{{"dismiss", "a1", nil}})
}

// AC2 and AC3: J and K change cards in the review. Each card keeps its
// draft and its review.
func TestAskReviewSurvivesACardSwitch(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "coordinator"), oneQuestionAsk("a2", "manager-hive"))
	m = run(t, m, key('2', "2"))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	m, _ = raise(t, m, key('J', "J"))
	requireView(t, m, "? manager-hive asks · Model", "│ › 1 [ ] Yes")
	m, _ = raise(t, m, key('3', "3"))
	requireView(t, m, "? manager-hive asks · confirm", "│   Model → Later ")
	m, _ = raise(t, m, key('K', "K"))
	requireView(t, m, "? coordinator asks · confirm", "│ › Model → No ", "│   Send ")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	m = run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{2}}}}})
	requireView(t, m, "? manager-hive asks · confirm", "│   Model → Later ", "│ › Send ")
}

// AC2: a data refresh keeps the review and the draft.
func TestAskReviewSurvivesARefresh(t *testing.T) {
	m, calls := askModel(t, 120, 40, twoQuestionAsk("a1"))
	m, _ = raise(t, m, key('1', "1"))
	m, _ = raise(t, m, enterKey)
	m, _ = raise(t, m, key('2', "2"))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	agents := []records.Agent{{Key: "coordinator-g1", Name: "coordinator", Status: records.AgentLive}}
	data := buildChatScreenData(m.chatData.entries, nil, []ask.Record{twoQuestionAsk("a1")}, agents, m.now())
	m, _ = raise(t, m, screenLoadMsg{request: chatDataRequest, chat: data})
	requireView(t, m, "? coordinator asks · confirm", "│   Scope → Wheel ", "│ › Order → Card first ")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1}}, {Picks: []int{2}}}}})
}

// AC1: for a stopped asker, the review names the parent, and Send keeps the
// route of the store.
func TestAskReviewForAStoppedAsker(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "t1-worker"))
	agents := []records.Agent{
		{Key: "coordinator-g1", Name: "coordinator", Status: records.AgentLive},
		{Key: "t1-worker", Name: "t1-worker", Status: records.AgentStopped},
	}
	m.chatData = buildChatScreenData(m.chatData.entries, nil, m.chatData.asks, agents, m.now())
	m.rebuildChat()
	m = run(t, m, key('1', "1"))
	requireNoCalls(t, calls, "the pick")
	requireView(t, m, "? t1-worker asks · confirm", " asker stopped · answer goes to coordinator ╮",
		"│ › Send to coordinator ", " review · ↑↓ question · enter send · x dismiss")
	run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1}}}}})
}

// Decision 114: below 120 columns, the stopped-asker note is shorter when the
// full note would cover the review title.
func TestAskCompactReviewForAStoppedAskerKeepsTheTitle(t *testing.T) {
	record := oneQuestionAsk("a1", "asker-scratch")
	m, calls := askModel(t, 80, 24, record)
	agents := []records.Agent{
		{Key: "asker-scratch", Name: "asker-scratch", Status: records.AgentStopped},
		{Key: "parent-scratch", Name: "parent-scratch", Status: records.AgentLive},
	}
	record.Parent, record.ParentKey = "parent-scratch", "parent-scratch"
	m.chatData = buildChatScreenData(m.chatData.entries, nil, []ask.Record{record}, agents, m.now())
	m.rebuildChat()
	m = run(t, m, key('1', "1"))
	requireNoCalls(t, calls, "the pick")
	requireView(t, m,
		"╭─ ? asker-scratch asks · confirm ────────── asker stopped · to parent-scratch ╮",
		"│ › Send to parent-scratch ")
	run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1}}}}})
}

func longAsk(id string) ask.Record {
	return ask.Record{
		ID: id, Asker: "coordinator", AskerKey: "coordinator-g1",
		CreatedAt: "2026-10-20T11:31:00Z", State: ask.StateOpen,
		Questions: []ask.Question{
			{Question: "Which parts?", Header: "Scope", MultiSelect: true, Options: []ask.Option{
				askOption("Wheel scroll over the feed and the card", "scroll"), askOption("Click focus", "focus"),
				askOption("Modes", "i in"),
			}},
			{Question: "Which first?", Header: "Order", Options: []ask.Option{
				askOption("Modes first", "new modes"), askOption("Card first", "later"),
			}},
			{Question: "Which model?", Header: "Model", Options: []ask.Option{
				askOption("Opus", "strong"), askOption("Sonnet", "fast"),
			}},
			{Question: "Which wave?", Header: "Wave", Options: []ask.Option{askOption("A", "now"), askOption("B", "later")}},
			{Question: "Which tier?", Header: "Tier", Options: []ask.Option{askOption("Strong", "slow"), askOption("Fast", "weak")}},
			{Question: "Land now?", Header: "Land", Options: []ask.Option{askOption("Yes", "now"), askOption("No", "hold")}},
		},
	}
}

const (
	longNote  = "keep the old tab key for one release, then remove it after the users read the change log of the next version"
	longOther = "a typed answer that is longer than one row of the card at eighty columns and must wrap without a cut"
)

// longReview gives a model with the review of longAsk: long picks, a long
// note, and a long Other answer.
func longReview(t *testing.T, width, height int) (*Model, *[]askCall) {
	t.Helper()
	m, calls := askModel(t, width, height, longAsk("a1"))
	m, _ = raise(t, m, key('1', "1"))
	m, _ = raise(t, m, key('2', "2"))
	m, _ = raise(t, m, key('4', "4"))
	for _, r := range longOther {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, enterKey)
	m, _ = raise(t, m, key('n', "n"))
	for _, r := range longNote {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, enterKey)
	for _, digit := range "2111" {
		m, _ = raise(t, m, key(digit, string(digit)))
	}
	return m, calls
}

// cardText gives the body text of the card in view, with the rows joined
// by spaces and the frame, the ┃ thumb, and the input line removed.
func cardText(m *Model) string {
	var words []string
	for line := range strings.SplitSeq(ansi.Strip(m.View().Content), "\n") {
		if strings.HasPrefix(line, "├") {
			break // the input line is below the rule
		}
		if !strings.HasPrefix(line, "│") {
			continue
		}
		line = strings.TrimRight(strings.TrimRight(line, " │┃"), " ")
		words = append(words, strings.Fields(strings.Trim(line, "│ ›"))...)
	}
	return strings.Join(words, " ")
}

// AC2: long picks, a long note, and a long Other answer wrap in full at
// both widths. The card scrolls to keep the highlighted question in view.
func TestAskReviewWrapsAndScrollsLongContent(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}} {
		t.Run(fmt.Sprint(size[0]), func(t *testing.T) {
			m, calls := longReview(t, size[0], size[1])
			requireNoCalls(t, calls, "the review")
			requireFits(t, m, size[0], size[1])
			if strings.Contains(cardText(m), "…") {
				t.Fatalf("card text = %q, want no cut", cardText(m))
			}
			requireView(t, m, "│ › Send ")
			for range 6 {
				m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
			}
			requireView(t, m, "│ › Scope → Wheel scroll over the feed and the card,")
			text := cardText(m)
			if size[0] == 120 {
				want := "Scope → Wheel scroll over the feed and the card, Click focus, " + longOther + " Order → Modes first · note: " + longNote
				if !strings.Contains(text, want) {
					t.Fatalf("card text = %q\nwant %q", text, want)
				}
			}
			requireFits(t, m, size[0], size[1])
			for range 6 {
				m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
			}
			requireView(t, m, "│ › Send ")
			if size[0] == 80 {
				rejectView(t, m, "│ › Scope")
			}
			if !strings.Contains(cardText(m), "Model → Sonnet Wave → A Tier → Strong Land → Yes") {
				t.Fatalf("card text = %q, want the last question", cardText(m))
			}
			run(t, m, enterKey)
			requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{
				{Picks: []int{1, 2, 4}, Other: longOther},
				{Picks: []int{1}, Note: longNote},
				{Picks: []int{2}}, {Picks: []int{1}}, {Picks: []int{1}}, {Picks: []int{1}},
			}}})
		})
	}
}

// AC2: at 80x24 the edited question returns to the review in view with
// Send.
func TestAskReviewAfterAnEditAtEightyColumns(t *testing.T) {
	m, calls := longReview(t, 80, 24)
	for range 4 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	}
	m, _ = raise(t, m, enterKey)
	requireView(t, m, "3. Which model?")
	m, _ = raise(t, m, key('1', "1"))
	requireFits(t, m, 80, 24)
	requireView(t, m, "│   Model → Opus ", "│ › Send ")
	run(t, m, enterKey)
	requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{
		{Picks: []int{1, 2, 4}, Other: longOther},
		{Picks: []int{1}, Note: longNote},
		{Picks: []int{1}}, {Picks: []int{1}}, {Picks: []int{1}}, {Picks: []int{1}},
	}}})
}

// AC1: a blocked caller sends nothing from the review.
func TestAskReviewBlockedCallerSendsNothing(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m.chatCallerBlocked = true
	m = run(t, m, key('1', "1"))
	run(t, m, enterKey)
	requireNoCalls(t, calls, "a blocked caller")
}

// AC4: a fixture with the focus review shows the review of the card, with
// review_row as the highlighted question. Without review_row, Send has the
// highlight.
func TestDemoFixtureShowsTheReview(t *testing.T) {
	base, err := os.ReadFile(filepath.Join(designDir, "ChatAskMulti.fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		row  string
		want []string
	}{
		{"", []string{"│   Scope → Mouse wheel (REC), Click to focus, Normal and input mode ", "│   Order → Modes first (REC) · note: keep the old tab key for one release ", "│ › Send "}},
		{`, "review_row": 1`, []string{"│ › Scope → Mouse wheel (REC),", "│   Send "}},
	} {
		data := strings.Replace(string(base), `"focus": "note"`, `"focus": "review"`, 1)
		data = strings.Replace(data, `"note_open": true`, `"note_open": false`+tc.row, 1)
		path := filepath.Join(t.TempDir(), "review.fixture.json")
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
		loaded, err := loadDemoFixture(path)
		if err != nil {
			t.Fatal(err)
		}
		m := newDemoModel(t.TempDir(), &demoState{fixture: loaded})
		t.Cleanup(m.Stop)
		requireView(t, m, append(tc.want, "? coordinator asks · confirm")...)
		if got := viewRow(m, 39); !strings.Contains(got, " review · ↑↓ question · enter send") {
			t.Fatalf("status row = %q", got)
		}
	}
}

// AC3 and golden-frames rule 18: in the review, only ↑↓, enter, x, J/K, and
// the focus keys act. i, pgup, pgdown, ?, and other frame keys do nothing.
// The status row names the review keys. esc focuses the input line. ctrl+c
// twice still quits.
func TestAskReviewKeepsOnlyTheReviewKeys(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	for index := 3; index <= 40; index++ {
		m.chatData.entries = append(m.chatData.entries, m.chatData.entries[0])
		m.chatData.entries[len(m.chatData.entries)-1].ID = fmt.Sprintf("c%d", index)
	}
	m.rebuildChat()
	m = run(t, m, key('2', "2"))
	before := ansi.Strip(m.View().Content)
	top := m.chatFeed.top
	presses := []tea.KeyPressMsg{
		key('i', "i"), {Code: tea.KeyPgUp}, {Code: tea.KeyPgDown},
		key('?', "?"), key('r', "r"), key('h', "h"), ctrlBracket(),
		{Code: tea.KeyLeft}, {Code: tea.KeyRight}, key('g', "g"), key('G', "G"),
	}
	for _, press := range presses {
		m = run(t, m, press)
		if got := ansi.Strip(m.View().Content); got != before {
			t.Fatalf("%s changed the view:\n%s", press.String(), got)
		}
	}
	if m.chatFeed.top != top || m.chatInputActive || m.helpOpen || m.screenIndex != 0 {
		t.Fatalf("a review key acted: top %d→%d, input %v, help %v, screen %d", top, m.chatFeed.top, m.chatInputActive, m.helpOpen, m.screenIndex)
	}
	requireNoCalls(t, calls, "the frame keys")
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · review · ↑↓ question · enter send · x dismiss · J/K card · ^r input") {
		t.Fatalf("status row = %q", got)
	}
	// Rule 18: esc gives the focus to the input line, also from the review.
	m, _ = raise(t, m, escKey())
	if m.focusState() != "input" {
		t.Fatalf("focus after esc = %q, want input", m.focusState())
	}
	_, cmd := raise(t, m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("the first ctrl+c gave no command")
	}
	_, cmd = raise(t, m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("the second ctrl+c did not quit")
	}
}

// longNoteAsk gives a one-question ask and a note of words w001 to wNNN with
// MIDDLE-SENTINEL in the middle and FINAL-SENTINEL at the end.
func longNoteText(words int) string {
	var parts []string
	for index := 1; index <= words; index++ {
		parts = append(parts, fmt.Sprintf("w%03d", index))
		if index == words/2 {
			parts = append(parts, "MIDDLE-SENTINEL")
		}
	}
	return strings.Join(append(parts, "FINAL-SENTINEL"), " ")
}

// AC2: a note longer than the dock can show is readable in full with ↑↓
// only. ↑ gives a row the highlight at its first line. ↓ pages through the
// highlighted row before it moves to the next row, and ↑ pages back. Each
// word shows at some step, and the first visible line of the highlighted
// row has the › mark.
func TestAskReviewReachesEachPartOfALongRow(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}} {
		t.Run(fmt.Sprint(size[0]), func(t *testing.T) {
			note := longNoteText(900)
			m, calls := askModel(t, size[0], size[1], oneQuestionAsk("a1", "manager-hive"))
			m, _ = raise(t, m, key('n', "n"))
			m.askCards["a1"].notes[0] = note
			m = run(t, m, enterKey)
			requireView(t, m, "│ › Send ")
			seen := map[string]bool{}
			look := func() {
				requireFits(t, m, size[0], size[1])
				view := ansi.Strip(m.View().Content)
				if !strings.Contains(view, "│ › ") {
					t.Fatalf("no › mark in view:\n%s", view)
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
			requireView(t, m, "│ › Model → Yes · note: w001")
			m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
			look()
			rejectView(t, m, "note: w001")
			m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
			requireView(t, m, "│ › Model → Yes · note: w001")
			for range 12 {
				m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
				look()
			}
			requireView(t, m, "│ › Send ")
			for _, word := range append(strings.Fields(note), "Model", "Send") {
				if !seen[word] {
					t.Fatalf("%q never showed", word)
				}
			}
			run(t, m, enterKey)
			requireCalls(t, calls, []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1}, Note: note}}}})
		})
	}
}
