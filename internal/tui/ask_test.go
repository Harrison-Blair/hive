package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Harrison-Blair/hive/internal/ask"
	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestAskBoardsMatch(t *testing.T) {
	for _, board := range []string{"ChatAsk", "ChatAskMulti", "ChatAskPreview", "ChatAskStopped", "CompactChatAsk", "ChatAskReview", "ChatAskReviewMulti", "CompactChatAskReview"} {
		t.Run(board, func(t *testing.T) { checkBoard(t, board) })
	}
}

func askOption(label, description string) ask.Option {
	return ask.Option{Label: label, Description: description}
}

func oneQuestionAsk(id, asker string) ask.Record {
	return ask.Record{
		ID: id, Asker: asker, AskerKey: asker, Parent: "coordinator", ParentKey: "coordinator-g1",
		CreatedAt: "2026-10-20T11:30:00Z", State: ask.StateOpen,
		Questions: []ask.Question{{Question: "Restart " + asker + "?", Header: "Model", Options: []ask.Option{
			askOption("Yes", "stop it"), askOption("No", "no change"), askOption("Later", "after the goal"),
		}}},
	}
}

func twoQuestionAsk(id string) ask.Record {
	return ask.Record{
		ID: id, Asker: "coordinator", AskerKey: "coordinator-g1",
		CreatedAt: "2026-10-20T11:31:00Z", State: ask.StateOpen,
		Questions: []ask.Question{
			{Question: "Which parts?", Header: "Scope", MultiSelect: true, Options: []ask.Option{
				askOption("Wheel", "scroll"), askOption("Click", "focus"), askOption("Modes", "i in"),
			}},
			{Question: "Which first?", Header: "Order", Options: []ask.Option{
				askOption("Modes first", "new modes"), askOption("Card first", "later"),
			}},
		},
	}
}

type askCall struct {
	action  string
	id      string
	replies []ask.Reply
}

// askModel gives a Chat model with the asks. Each asker is live. The calls
// slice records each answer and dismiss in place of the store.
func askModel(t *testing.T, width, height int, asks ...ask.Record) (*Model, *[]askCall) {
	t.Helper()
	m := chatModel(t, width, height, 2)
	calls := &[]askCall{}
	m.askAnswer = func(_ context.Context, _ string, id string, replies []ask.Reply) output.Envelope {
		*calls = append(*calls, askCall{"answer", id, replies})
		return output.Envelope{Outcome: "ok"}
	}
	m.askDismiss = func(_ context.Context, _ string, id string) output.Envelope {
		*calls = append(*calls, askCall{"dismiss", id, nil})
		return output.Envelope{Outcome: "ok"}
	}
	agents := []records.Agent{{Key: "coordinator-g1", Name: "coordinator", Status: records.AgentLive}}
	for _, record := range asks {
		agents = append(agents, records.Agent{Key: record.AskerKey, Name: record.Asker, Status: records.AgentLive})
	}
	m.chatData = buildChatScreenData(m.chatData.entries, nil, asks, agents, m.now())
	m.header.asks = len(m.openAsks())
	m.rebuildChat()
	m.invalidateView()
	// Golden-frames chat.md rule 17: the input line has the focus at start,
	// and ctrl+r gives the card the focus.
	if len(m.openAsks()) > 0 {
		m, _ = raise(t, m, ctrl('r'))
	}
	return m, calls
}

// run raises msg and then each message of the commands that it gives, so a
// send reaches the injected answer function.
func run(t *testing.T, m *Model, msg tea.Msg) *Model {
	t.Helper()
	m, cmd := raise(t, m, msg)
	for _, next := range flatten(cmd) {
		switch next.(type) {
		case askDoneMsg:
			m, _ = raise(t, m, next)
		}
	}
	return m
}

// flatten runs cmd and each batch in it, and gives the messages.
func flatten(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, inner := range batch {
			out = append(out, flatten(inner)...)
		}
		return out
	}
	return []tea.Msg{msg}
}

// Golden-frames chat.md rules 14, 15, and 17 and decision 33: the card sits
// in the composer frame above the input line, the ask line in the transcript
// names the composer, and the status row gives the keys of the card.
func TestAskCardSitsInTheComposerFrame(t *testing.T) {
	m, _ := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	requireView(t, m,
		"╭─ ? manager-hive asks · Model ─",
		" open · 1 of 1 ╮",
		"│ Restart manager-hive?",
		"│ › 1 [ ] Yes                        stop it",
		"│   2 [ ] No                         no change",
		"│   4 [ ] Other: type an answer",
		"│ note › none · n adds one",
		"├──────",
		"? manager-hive 07:30  asked · Model · in the composer, ^r to answer",
	)
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · ask · ↑↓ move · 1..4 pick · enter review · n note · x dismiss · J/K card · ^r input · esc input") {
		t.Fatalf("status row = %q", got)
	}
	top := rowOf(m, "╭─ ? manager-hive")
	grid := renderGrid(t, m, 120, 40)
	requireCell(t, grid, top, 0, "╭", cellgrid.Accent)
	requireCell(t, grid, top+3, 2, "›", cellgrid.Accent)
	if bottom := rowOf(m, "╰─"); bottom != 38 {
		t.Fatalf("frame bottom on row %d, want 38, above the status row", bottom)
	}
}

// Rule 16: the frame is at most half of the content rows, and the card
// scrolls to the current question.
func TestAskCardSitsAboveTheInputAndTakesAtMostHalfTheScreen(t *testing.T) {
	record := oneQuestionAsk("a1", "manager-hive")
	record.Questions = nil
	for index := range 4 {
		record.Questions = append(record.Questions, ask.Question{
			Question: fmt.Sprintf("Question %d?", index+1), Header: fmt.Sprintf("H%d", index+1),
			Options: []ask.Option{askOption("A", "a"), askOption("B", "b"), askOption("C", "c"), askOption("D", "d"), askOption("E", "e")},
		})
	}
	m, _ := askModel(t, 80, 24, record)
	view := requireFits(t, m, 80, 24)
	rows := strings.Split(view, "\n")
	top, rule, bottom := -1, -1, -1
	for index, row := range rows {
		switch {
		case strings.HasPrefix(row, "╭─"):
			top = index
		case strings.HasPrefix(row, "├─"):
			rule = index
		case strings.HasPrefix(row, "╰─"):
			bottom = index
		}
	}
	if top < 0 || rule <= top || bottom != rule+2 || !strings.HasPrefix(rows[rule+1], "│ > ") {
		t.Fatalf("no card above the input line:\n%s", view)
	}
	// 23 content rows above the status row: the frame has at most 11.
	if height := bottom - top + 1; height > 23/2 {
		t.Fatalf("frame height = %d, want at most 11:\n%s", height, view)
	}
	requireView(t, m, "1. Question 1?")
	for range 3 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	requireView(t, m, "4. Question 4?")
	rejectView(t, m, "1. Question 1?")
}

func TestAskDigitPicksAndSendsASingleQuestion(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m = run(t, m, key('2', "2"))
	run(t, m, enterKey)
	want := []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{2}}}}}
	if fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
	if m.screenIndex != 0 {
		t.Fatalf("the digit opened screen %d", m.screenIndex)
	}
}

func TestAskArrowsAndEnterPickTheCurrentOption(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	requireView(t, m, "│ › 3 [ ] Later")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyUp})
	requireView(t, m, "│ › 2 [ ] No")
	m = run(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	run(t, m, enterKey)
	if want := []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{2}}}}}; fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

func TestAskTwoQuestionsWithMarksAndANote(t *testing.T) {
	m, calls := askModel(t, 120, 40, twoQuestionAsk("a1"))
	m, _ = raise(t, m, key(' ', "space"))
	m, _ = raise(t, m, key('3', "3"))
	requireView(t, m, "│   1 [x] Wheel", "│ › 3 [x] Modes")
	m, _ = raise(t, m, key('n', "n"))
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · note · type the note · enter keep note · ↑↓ option · space mark · ^r input") {
		t.Fatalf("status row = %q", got)
	}
	for _, r := range "small" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	requireView(t, m, "│ note › small█")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	requireView(t, m, "│ 2. Which first?", "│ › 1 [ ] Modes first")
	if len(*calls) != 0 {
		t.Fatalf("enter on question 1 sent %v", *calls)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	m = run(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	run(t, m, enterKey)
	want := []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1, 3}, Note: "small"}, {Picks: []int{2}}}}}
	if fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

func TestAskEnterFromANoteBeforeAPickPicksTheCurrentOption(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = raise(t, m, key('n', "n"))
	m, _ = raise(t, m, key('o', "o"))
	m, _ = raise(t, m, key('k', "k"))
	m = run(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	run(t, m, enterKey)
	if want := []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{2}, Note: "ok"}}}}; fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

// Rule 18 and ruling C14: tab does nothing on the card, n opens the note, n
// types in the note, enter keeps the note, and esc gives the input line the
// focus with the note kept.
func TestAskNoteKeys(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	before := m.View().Content
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.View().Content != before {
		t.Fatal("tab changed the card")
	}
	m, _ = raise(t, m, key('n', "n"))
	for _, r := range "no" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	requireView(t, m, "│ note › no█")
	m, _ = raise(t, m, escKey())
	if m.focusState() != "input" {
		t.Fatalf("focus after esc = %q, want input", m.focusState())
	}
	requireView(t, m, "│ note › no ")
	m, _ = raise(t, m, ctrl('r'))
	m = run(t, m, key('1', "1"))
	run(t, m, enterKey)
	if want := []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1}, Note: "no"}}}}; fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

func TestAskOtherOpensATextLine(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m, _ = raise(t, m, key('4', "4"))
	if len(*calls) != 0 {
		t.Fatalf("Other sent %v before the text", *calls)
	}
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · other · type the answer · enter keep answer · ^r input") {
		t.Fatalf("status row = %q", got)
	}
	for _, r := range "tomorrow" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyBackspace})
	requireView(t, m, "│ other › tomorro█")
	m = run(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	run(t, m, enterKey)
	if want := []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{4}, Other: "tomorro"}}}}; fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

func TestAskXDismisses(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	run(t, m, key('x', "x"))
	if want := []askCall{{"dismiss", "a1", nil}}; fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

func TestAskJAndKMoveBetweenCards(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "coordinator"), oneQuestionAsk("a2", "manager-hive"))
	requireView(t, m, "? coordinator asks · Model", " open · 1 of 2 · J next ╮",
		"? coordinator 07:30  asked · Model · in the composer, ^r to answer", "? manager-hive 07:30  asked · Model · waits · 2 of 2")
	m, _ = raise(t, m, key('J', "J"))
	requireView(t, m, "? manager-hive asks · Model", " open · 2 of 2 ╮",
		"? coordinator 07:30  asked · Model · waits · 1 of 2", "? manager-hive 07:30  asked · Model · in the composer, ^r to answer")
	m, _ = raise(t, m, key('K', "K"))
	requireView(t, m, "? coordinator asks · Model")
	run(t, m, key('x', "x"))
	if want := []askCall{{"dismiss", "a1", nil}}; fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

func TestAskStoppedAskerShowsTheParent(t *testing.T) {
	m, _ := askModel(t, 120, 40, oneQuestionAsk("a1", "t1-worker"))
	agents := []records.Agent{
		{Key: "coordinator-g1", Name: "coordinator", Status: records.AgentLive},
		{Key: "t1-worker", Name: "t1-worker", Status: records.AgentStopped},
	}
	m.chatData = buildChatScreenData(m.chatData.entries, nil, m.chatData.asks, agents, m.now())
	m.rebuildChat()
	m.invalidateView()
	requireView(t, m, " asker stopped · answer goes to coordinator ╮",
		"asked · Model · asker stopped · in the composer, ^r to answer")
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · ask · ↑↓ move · 1..4 pick · enter review for parent · n note · x dismiss · J/K card") {
		t.Fatalf("status row = %q", got)
	}
}

// Rule 6: after an answer or a dismiss, the line from you follows the ask
// line. Two ask lines have no blank row between them.
func TestAskHistoryShowsTheAnswerAndTheDismissal(t *testing.T) {
	answered := twoQuestionAsk("a1")
	answered.State, answered.ClosedAt = ask.StateAnswered, "2026-10-20T11:40:00Z"
	answered.Replies = []ask.Reply{{Picks: []int{1, 4}, Other: "Header", Note: "small"}, {Picks: []int{2}}}
	dismissed := oneQuestionAsk("a2", "manager-hive")
	dismissed.State, dismissed.ClosedAt = ask.StateDismissed, "2026-10-20T11:41:00Z"
	m, _ := askModel(t, 120, 40, answered, dismissed)
	requireView(t, m,
		"? manager-hive 07:30  asked · Model",
		"? coordinator 07:31  asked · Scope, Order",
		"› you 07:40",
		"answered coordinator · Scope → Wheel, Header · note: small; Order → Card first",
		"› you 07:41",
		"dismissed manager-hive · Model",
	)
	rejectView(t, m, "╭─ ?", "in the composer")
	if next := viewRow(m, rowOf(m, "? manager-hive 07:30")+1); !strings.Contains(next, "? coordinator 07:31") {
		t.Fatalf("row after the first ask line = %q, want the second ask line", next)
	}
	if got := viewRow(m, 39); !strings.Contains(got, "^g agents") {
		t.Fatalf("status row = %q, want the chords", got)
	}
}

// A click on an option row picks it and gives the card the focus, also from
// the input line.
func TestAskMouseFocusesAndScrolls(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m, _ = raise(t, m, ctrl('r'))
	if m.focusState() != "input" {
		t.Fatalf("focus after ctrl+r = %q, want input", m.focusState())
	}
	grid := renderGrid(t, m, 120, 40)
	requireCell(t, grid, rowOf(m, "╭─ ? manager-hive"), 0, "╭", cellgrid.Accent)
	// The pick of the one question opens the review.
	m = run(t, m, click(12, rowOf(m, "2 [ ] No")))
	if m.focusState() != "review" {
		t.Fatalf("focus after a click on an option = %q, want review", m.focusState())
	}
	run(t, m, enterKey)
	if want := []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{2}}}}}; fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

// Rule 17: a click on the card gives it the focus, a click on the input line
// gives the input line the focus, and ctrl+r moves the focus both ways.
func TestAskClickOnTheCardGivesItTheFocus(t *testing.T) {
	m, _ := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m, _ = raise(t, m, click(60, rowOf(m, "│ > ")))
	if m.focusState() != "input" {
		t.Fatalf("focus after an input click = %q, want input", m.focusState())
	}
	m, _ = raise(t, m, click(60, rowOf(m, "Restart manager-hive?")))
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · ask ·") {
		t.Fatalf("status row after a card click = %q", got)
	}
	m, _ = raise(t, m, ctrl('r'))
	if m.focusState() != "input" {
		t.Fatalf("focus after ctrl+r = %q, want input", m.focusState())
	}
	m, _ = raise(t, m, ctrl('r'))
	if m.focusState() != "ask" {
		t.Fatalf("focus after a second ctrl+r = %q, want ask", m.focusState())
	}
}

func TestAskWheelOverTheCardScrollsTheFeed(t *testing.T) {
	m, _ := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	for index := 3; index <= 40; index++ {
		m.chatData.entries = append(m.chatData.entries, m.chatData.entries[0])
		m.chatData.entries[len(m.chatData.entries)-1].ID = fmt.Sprintf("c%d", index)
	}
	m.rebuildChat()
	m.invalidateView()
	m.View()
	before := m.chatFeed.top
	m, _ = raise(t, m, wheel(60, 32, tea.MouseWheelUp))
	if m.chatFeed.top >= before {
		t.Fatalf("feed top = %d after the wheel, want less than %d", m.chatFeed.top, before)
	}
	if m.focusState() != "ask" {
		t.Fatalf("focus after the wheel = %q, want ask", m.focusState())
	}
}

func TestReadHeaderCountsTheOpenAsks(t *testing.T) {
	home := newTestHome(t)
	open := oneQuestionAsk("a1", "manager-hive")
	second := oneQuestionAsk("a2", "manager-hive")
	closed := oneQuestionAsk("a3", "manager-hive")
	closed.State, closed.DeliveredAt = ask.StateAnswered, "2026-10-20T11:40:00Z"
	waiting := oneQuestionAsk("a4", "manager-hive")
	waiting.State = ask.StateDismissed
	for _, record := range []ask.Record{open, second, closed, waiting} {
		writeRecord(t, home, []string{"asks"}, record.ID, record)
	}
	state, err := readHeader(home)
	if err != nil {
		t.Fatal(err)
	}
	if state.asks != 2 || state.askUndelivered != 1 {
		t.Fatalf("asks = %d, undelivered = %d; want 2 and 1", state.asks, state.askUndelivered)
	}
}

func TestTickDeliversAWaitingAnswer(t *testing.T) {
	m, _ := askModel(t, 120, 40)
	delivered := make(chan struct{}, 4)
	m.askDeliver = func(context.Context, string) error {
		delivered <- struct{}{}
		return nil
	}
	// tickRuns runs each command of one tick in the background, because the
	// next tick waits.
	tickRuns := func() {
		_, cmd := raise(t, m, tickMsg{})
		msg := cmd()
		if inner, ok := msg.(tea.BatchMsg); ok {
			for _, next := range inner {
				go next()
			}
		}
	}
	tickRuns()
	select {
	case <-delivered:
		t.Fatal("tick delivered with no waiting answer")
	case <-time.After(300 * time.Millisecond):
	}
	m.header.askUndelivered = 1
	tickRuns()
	select {
	case <-delivered:
	case <-time.After(2 * time.Second):
		t.Fatal("tick did not deliver the waiting answer")
	}
}

func TestDemoBlocksTheAnswer(t *testing.T) {
	m := newTestDemoModel(t, "ChatAsk.fixture.json", demoSize{})
	m, _ = raise(t, m, key('1', "1"))
	m, _ = raise(t, m, enterKey)
	requireView(t, m, demoNotice)
}

// A new ask record refreshes the header count and the Chat screen.
func TestANewAskRefreshesTheHeaderAndChat(t *testing.T) {
	home := newTestHome(t)
	writeRecord(t, home, []string{"agents"}, "manager-hive", records.Agent{Key: "manager-hive", Name: "manager-hive"})
	header := pathSignature(headerRefreshPaths(home))
	chat := pathSignature(screenRefreshPaths(home, chatDataRequest, time.Now(), tasksData{}))
	writeRecord(t, home, []string{"asks"}, "a1", oneQuestionAsk("a1", "manager-hive"))
	if pathSignature(headerRefreshPaths(home)) == header {
		t.Error("a new ask does not refresh the header")
	}
	if pathSignature(screenRefreshPaths(home, chatDataRequest, time.Now(), tasksData{})) == chat {
		t.Error("a new ask does not refresh Chat")
	}
}

// Help lists the keys of the ask card.
func TestHelpListsTheAskKeys(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	m, _ = raise(t, m, key('?', "?"))
	// The ask lines are at the end of the help text.
	for range 40 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	// Golden-frames chat.md rules 17 and 18 give the card keys.
	requireView(t, m, "│  ctrl+r moves the focus between input and ask card · esc to input ",
		"│  ask card: ↑↓ move · 1–9 pick · space mark · n note · x dismiss · J/K card ",
		"│  ask review: enter on the last question · ↑↓ question · enter send or change")
}

// The status row shows the count of the open ask records at 0, 1, and 2, at
// both widths (golden-frames decision 11).
func TestHeaderShowsTheOpenAskRecords(t *testing.T) {
	for _, tc := range []struct {
		open, width int
		want        string
	}{
		{0, 120, " · ? 0 asks · needs you 0 · "},
		{1, 120, " · ? 1 ask · needs you 0 · "},
		{2, 120, " · ? 2 asks · needs you 0 · "},
		{0, 80, " · ? 0 asks · needs you 0 "},
		{1, 80, " · ? 1 ask · needs you 0 "},
		{2, 80, " · ? 2 asks · needs you 0 "},
	} {
		home := newTestHome(t)
		closed := oneQuestionAsk("a9", "manager-hive")
		closed.State = ask.StateDismissed
		writeRecord(t, home, []string{"asks"}, closed.ID, closed)
		for index := range tc.open {
			id := fmt.Sprintf("a%d", index+1)
			writeRecord(t, home, []string{"asks"}, id, oneQuestionAsk(id, "manager-hive"))
		}
		state, err := readHeader(home)
		if err != nil {
			t.Fatal(err)
		}
		m := chatModel(t, tc.width, 40, 1)
		m.now = func() time.Time { return time.Date(2026, 10, 20, 8, 0, 0, 0, time.Local) }
		m.header.asks = 7
		m.invalidateView()
		m.View()
		m, _ = raise(t, m, headerLoadMsg{state: state})
		rows := viewRows(m)
		if got := rows[len(rows)-1]; !strings.Contains(got, tc.want) || !strings.HasSuffix(got, " 08:00 ") {
			t.Errorf("%d open at %d: status row = %q, want %q and the clock 08:00", tc.open, tc.width, got, tc.want)
		}
	}
}

// In a multi-select question, a mark on Other opens its text line, and the
// answer keeps the marks and the text.
func TestAskOtherInAMultiSelectQuestion(t *testing.T) {
	m, calls := askModel(t, 120, 40, twoQuestionAsk("a1"))
	m, _ = raise(t, m, key('1', "1"))
	m, _ = raise(t, m, key('4', "4"))
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ○ none · other ·") {
		t.Fatalf("status row = %q, want the Other line", got)
	}
	for _, r := range "keys" {
		m, _ = raise(t, m, key(r, string(r)))
	}
	requireView(t, m, "│ other › keys█")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m = run(t, m, key('2', "2"))
	run(t, m, enterKey)
	want := []askCall{{"answer", "a1", []ask.Reply{{Picks: []int{1, 4}, Other: "keys"}, {Picks: []int{2}}}}}
	if fmt.Sprint(*calls) != fmt.Sprint(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

// The answer prompt has the authority of the user. A TUI that the chat rule
// blocks (rule 2: a working directory below projects/) sends no answer and no
// dismiss.
func TestABlockedCallerCannotAnswer(t *testing.T) {
	m, calls := askModel(t, 120, 40, oneQuestionAsk("a1", "manager-hive"))
	m.chatCallerBlocked = true
	m.chatErr = "chat send is for the user only: rule 2"
	run(t, m, key('1', "1"))
	run(t, m, key('x', "x"))
	if len(*calls) != 0 {
		t.Fatalf("a blocked caller sent %v", *calls)
	}
	requireView(t, m, "chat send is for the user only: rule 2")
}
