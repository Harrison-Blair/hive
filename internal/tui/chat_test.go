package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	chatlib "github.com/Harrison-Blair/hive/internal/chat"
	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

func TestChatMatchesItsBoard(t *testing.T) {
	checkBoard(t, "Chat")
}

func TestCompactChatMatchesItsBoard(t *testing.T) {
	checkBoard(t, "CompactChat")
}

func TestChatStatusBoardsMatch(t *testing.T) {
	for _, board := range []string{"ChatStatus", "ChatEvents", "CompactChatStatus"} {
		t.Run(board, func(t *testing.T) { checkBoard(t, board) })
	}
}

// The fixture field ui.scroll.feed.top puts that message on row 1, below
// the earlier row of chat.md, and stops the follow of the end (rule 7).
func TestChatFixtureScrollPutsTheMessageAtTheTop(t *testing.T) {
	m := chatModel(t, 120, 40, 40)
	m.chatFeed.topItem = 5
	m.invalidateView()
	rows := viewRows(m)
	if !strings.Contains(rows[0], "· · · 5 earlier · · ·") || !strings.HasPrefix(rows[1], " ● coordinator ") || !strings.HasPrefix(rows[2], " line-06 ") {
		t.Fatalf("rows 0 to 2 = %q, want the earlier row and message 6", rows[:3])
	}
	if m.chatFeed.follow {
		t.Fatal("the scroll kept the follow of the end")
	}
}

// Chat loads the chat of the last seven days and the coordinator inbox, with
// the same renderer as demo mode (decisions 15, 37, 38), in the transcript
// form of golden-frames chat.md rules 2 to 5.
func TestChatLoadsRecentConversationAndInboxEvents(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	appendChatFixture(t, home, chatlib.Entry{ID: "c1", At: "2026-10-13T11:59:00Z", From: "user", Text: "outside the seven-day window", Delivery: "sent"})
	appendChatFixture(t, home, chatlib.Entry{ID: "c2", At: "2026-10-19T12:00:00Z", From: "user", Text: "first message", Delivery: "pending"})
	appendChatFixture(t, home, chatlib.Entry{ID: "c3", At: "2026-10-19T13:00:00Z", From: "coordinator", Text: "**answer**", ReplyTo: "c2"})
	appendChatFixture(t, home, chatlib.Entry{ID: "c4", At: "2026-10-19T14:00:00Z", From: "user", Text: "delivered message", Delivery: "pending"})
	appendChatFixture(t, home, chatlib.Entry{ID: "c4", At: "2026-10-19T14:00:01Z", Delivery: "sent"})
	appendInboxFixture(t, home, records.Inbox{Seq: 1, Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-one", Text: "message first line\nmessage second line", CreatedAt: "2026-10-19T12:30:00Z"})
	appendInboxFixture(t, home, records.Inbox{Seq: 2, Role: "coordinator", Kind: records.InboxExited, Agent: "worker-two", Text: "exited is noise", CreatedAt: "2026-10-19T12:45:00Z"})

	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m = loadChat(t, m)
	// Chat shows each time in the local zone, as the status row does.
	local := func(at string) string {
		parsed, err := time.Parse(time.RFC3339, at)
		if err != nil {
			t.Fatal(err)
		}
		return parsed.Local().Format("15:04")
	}
	requireMessage(t, m, " › you "+local("2026-10-19T12:00:00Z")+" · pending ", " first message ")
	requireView(t, m, " ▸ worker-one "+local("2026-10-19T12:30:00Z")+"  message first line")
	requireMessage(t, m, " ● coordinator "+local("2026-10-19T13:00:00Z")+" ", " answer ")
	requireMessage(t, m, " › you "+local("2026-10-19T14:00:00Z")+" ", " delivered message ")
	rejectView(t, m, local("2026-10-19T14:00:00Z")+" · pending")
	rejectView(t, m, "outside the seven-day window", "message second line", "exited is noise", "**")

	readBack, err := inboxlib.All(home, "coordinator")
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range readBack {
		if event.ReadAt != "" {
			t.Fatalf("Chat marked inbox record %d read", event.Seq)
		}
	}

	// A new load shows a new reply. On Chat, R types (decision 19).
	appendChatFixture(t, home, chatlib.Entry{ID: "c5", At: "2026-10-20T11:00:00Z", From: "coordinator", Text: "a new reply"})
	m = loadChat(t, m)
	requireView(t, m, " a new reply ")
}

// requireMessage fails when no row of the view starts with head and has a
// next row that starts with body.
func requireMessage(t *testing.T, m *Model, head, body string) {
	t.Helper()
	rows := viewRows(m)
	for index, row := range rows[:len(rows)-1] {
		if strings.HasPrefix(row, head) && strings.HasPrefix(rows[index+1], body) {
			return
		}
	}
	t.Fatalf("view has no %q above %q:\n%s", head, body, strings.Join(rows, "\n"))
}

// Each noise level shows the inbox kinds of the earlier Chat screen.
func TestChatNoiseLevelsFilterInboxKinds(t *testing.T) {
	m := newTestModel(t, newTestHome(t), fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	for index, kind := range []string{records.InboxMessage, records.InboxCompleted, records.InboxVerified, records.InboxBlocked, records.InboxExited, records.InboxQuiet, records.InboxUnresolved} {
		m.chatData.inbox = append(m.chatData.inbox, records.Inbox{Seq: uint64(index + 1), Kind: kind})
	}
	for _, tt := range []struct {
		noise string
		want  []string
	}{
		{chatNoiseNormal, []string{"message", "completed", "verified", "blocked", "quiet", "unresolved"}},
		{chatNoiseQuiet, []string{"message", "blocked", "unresolved"}},
		{chatNoiseAll, []string{"message", "completed", "verified", "blocked", "exited", "quiet", "unresolved"}},
	} {
		m.chatNoise = tt.noise
		var got []string
		for _, event := range m.visibleChatEvents() {
			got = append(got, event.Kind)
		}
		if strings.Join(got, " ") != strings.Join(tt.want, " ") {
			t.Errorf("noise %s shows %v; want %v", tt.noise, got, tt.want)
		}
	}
}

// chatModel gives a model on the Chat tab with n coordinator replies.
func chatModel(t *testing.T, width, height, n int) *Model {
	t.Helper()
	m := newTestModel(t, newTestHome(t), fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	for index := 1; index <= n; index++ {
		at := time.Date(2026, 10, 20, 10, index, 0, 0, time.UTC).Format(time.RFC3339)
		m.chatData.entries = append(m.chatData.entries, chatlib.Entry{ID: fmt.Sprintf("c%d", index), At: at, From: "coordinator", Text: fmt.Sprintf("line-%02d", index)})
	}
	m.rebuildChat()
	m.invalidateView()
	return m
}

// Golden-frames chat.md rule 7 and decision 19: the transcript follows its
// end, and pgup and pgdn page it while the input line keeps the focus. j, k,
// g, and G type.
func TestChatKeysScrollTheFeed(t *testing.T) {
	m := chatModel(t, 120, 40, 40)
	requireView(t, m, "line-40")
	rejectView(t, m, "line-01")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyPgUp})
	rejectView(t, m, "line-40")
	if !m.chatInputActive {
		t.Fatal("pgup moved the focus")
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	requireView(t, m, "line-40")
	top := m.chatFeed.top
	for _, character := range "gjkG" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	if m.chatFeed.top != top || m.chatInput.Value() != "gjkG" {
		t.Fatalf("top %d→%d, draft %q, want no scroll and gjkG typed", top, m.chatFeed.top, m.chatInput.Value())
	}
}

// Rule 9 and the key table of chat.md: each printable key types, also the
// keys of other screens, and ? once the line has text. shift+enter adds a
// line, and the second row starts with two spaces. A chord opens its screen
// and keeps the draft.
func TestChatKeysFollowTheFocusRules(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	for _, character := range "x12?/jkgGi" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	if m.activeTab() != tabChat || m.helpOpen {
		t.Fatalf("a key in the input left the input: tab %d, help %v", m.activeTab(), m.helpOpen)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModShift})
	for _, character := range "two" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	if got := m.chatInput.Value(); got != "x12?/jkgGi\ntwo" {
		t.Fatalf("draft = %q", got)
	}
	requireView(t, m, "│ > x12?/jkgGi ", "│   two█ ")
	m, _ = raise(t, m, ctrl('t'))
	if m.activeTab() != tabWork {
		t.Fatalf("^t gave tab %d", m.activeTab())
	}
	if got := m.chatInput.Value(); got != "x12?/jkgGi\ntwo" {
		t.Fatalf("draft after the chord = %q", got)
	}
}

// Golden-frames decision 33: below 120 columns the ask status row removes
// keys from the right, before the asks segment and the clock.
func TestChatCompactAskStatusRowCutsKeys(t *testing.T) {
	m, _ := askModel(t, 80, 24, oneQuestionAsk("a1", "manager-hive"), oneQuestionAsk("a2", "coordinator"))
	m = inEastern(m)
	got := viewRow(m, 23)
	if !strings.HasPrefix(got, " ○ none · ask · ↑↓ move · 1..4 pick · ") || !strings.HasSuffix(got, " · ? 2 asks    08:00 ") {
		t.Fatalf("compact ask status row = %q", got)
	}
	if strings.Contains(got, "J/K card") {
		t.Fatalf("compact ask status row = %q, want the last keys cut", got)
	}
}

// Golden-frames chat.md rules 2 and 9 and decision 11: the coordinator
// state is on the status row, a pending message shows · pending on its line,
// and the composer frame has no title.
func TestChatPendingMessageAndStateLeaveTheComposerBorder(t *testing.T) {
	home := newTestHome(t)
	writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, Status: records.AgentLive,
		AgentStatus: "working", StatusChangedAt: "2026-10-20T11:58:00Z",
	})
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	t.Cleanup(m.Stop)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	state, err := readHeader(home)
	if err != nil {
		t.Fatal(err)
	}
	m.header = state
	m.chatData.entries = []chatlib.Entry{{ID: "pending", At: "2026-10-20T11:59:00Z", From: "user", Text: "hello", Delivery: "pending"}}
	m.invalidateView()
	requireMessage(t, m, " › you "+time.Date(2026, 10, 20, 11, 59, 0, 0, time.UTC).Local().Format("15:04")+" · pending ", " hello ")
	if got := viewRow(m, 39); !strings.HasPrefix(got, " ● working 2m · ") {
		t.Fatalf("status row = %q", got)
	}
	if got := viewRow(m, 36); got != "╭"+strings.Repeat("─", 118)+"╮" {
		t.Fatalf("composer top = %q, want a plain border", got)
	}
	requireView(t, m, "│ > type a message · ? help ")
}

// Decision 11: the state segment forms and the whole-minute age.
func TestChatStateSegmentShowsTheGlyphAndWholeMinuteAge(t *testing.T) {
	cases := []struct {
		name, status, changedAt, want string
	}{
		{"working zero", "working", "2026-10-20T12:00:00Z", "● working 0m"},
		{"working one", "working", "2026-10-20T11:59:00Z", "● working 1m"},
		{"working fifty-nine", "working", "2026-10-20T11:01:00Z", "● working 59m"},
		{"idle", "idle", "2026-10-20T11:59:00Z", "○ idle 1m"},
		{"blocked", "blocked", "2026-10-20T12:00:00Z", "▲ blocked 0m"},
		{"other", "waiting", "2026-10-20T11:59:00Z", "● starting"},
		{"no time", "idle", "", "○ idle"},
		{"no live coordinator", "", "", "○ none"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := newTestHome(t)
			if tc.name != "no live coordinator" {
				writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
					Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, Status: records.AgentLive,
					AgentStatus: tc.status, StatusChangedAt: tc.changedAt,
				})
			}
			m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
			defer m.Stop()
			state, err := readHeader(home)
			if err != nil {
				t.Fatal(err)
			}
			m.header = state
			if got := ansi.Strip(m.stateSegment()); got != tc.want {
				t.Fatalf("state segment = %q, want %q", got, tc.want)
			}
		})
	}
}

// Rule 11: the frame grows up to one third of the content rows. Beyond that
// the text scrolls, the cursor row stays visible, and the ┃ thumb is on the
// right border. esc does nothing on the input line.
func TestChatInputCapsAndKeepsCursorVisible(t *testing.T) {
	for _, size := range []struct{ width, height int }{{120, 40}, {80, 24}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			m := chatModel(t, size.width, size.height, 0)
			defer m.Stop()
			m.chatInput.SetValue(strings.Repeat("release notes stay readable ", 60))
			m.invalidateView()
			before := viewRows(m)
			content := size.height - 1
			frameTop := content - content/3
			if !strings.HasPrefix(before[frameTop], "╭") || !strings.HasPrefix(before[content-1], "╰") {
				t.Fatalf("frame is not rows %d to %d:\n%s", frameTop, content-1, strings.Join(before, "\n"))
			}
			cursorRow := content - 2
			last := []rune(before[cursorRow])
			if !strings.Contains(before[cursorRow], "█") || last[size.width-1] != '┃' {
				t.Fatalf("cursor or thumb is not on the last input row:\n%s", strings.Join(before[frameTop:content], "\n"))
			}
			m, _ = raise(t, m, escKey())
			if after := viewRows(m); strings.Join(after, "\n") != strings.Join(before, "\n") || !m.chatInputActive {
				t.Fatalf("esc changed the view or the focus:\n%s", strings.Join(after, "\n"))
			}
		})
	}
}

func TestChatCtrlOOpensAndFoldsEventsFromBothRegions(t *testing.T) {
	m := chatModel(t, 120, 40, 0)
	defer m.Stop()
	m.chatData.inbox = []records.Inbox{{
		Seq: 1, Kind: records.InboxMessage, Agent: "manager-hive",
		Text: "first event line\nsecond event line", CreatedAt: "2026-10-20T11:00:00Z",
	}}
	m.rebuildChat()
	m.invalidateView()
	requireView(t, m, "▸ manager-hi", "first event line")
	eventToggle := tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl}
	m, _ = raise(t, m, eventToggle)
	requireView(t, m, "▾ manager-hi", "first event line", "second event line")
	m, _ = raise(t, m, eventToggle)
	requireView(t, m, "▸ manager-hi")
	m, _ = raise(t, m, key('i', "i"))
	m, _ = raise(t, m, eventToggle)
	requireView(t, m, "▾ manager-hi", "second event line")
}

// Demo mode sends nothing and shows the notice (decision 35).
func TestDemoChatSendShowsTheNotice(t *testing.T) {
	m := newTestDemoModel(t, "Chat.fixture.json", demoSize{})
	m, cmd := raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		if _, ok := cmd().(chatSendMsg); ok {
			t.Fatal("demo mode started a chat send")
		}
	}
	requireView(t, m, "demo mode: no action", "> 1A, 2A, 3B█")
}

// checkFit checks the properties of decision 46 at width×height: the view
// has height rows of width cells, each panel border is whole, and the newest
// message is visible.
func checkFit(t *testing.T, m *Model, width, height int, newest string) {
	t.Helper()
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: width, Height: height})
	view := ansi.Strip(m.View().Content)
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("%d rows; want %d:\n%s", len(lines), height, view)
	}
	for index, line := range lines {
		if got := ansi.StringWidth(line); got != width {
			t.Fatalf("row %d is %d cells; want %d: %q", index, got, width, line)
		}
	}
	// Golden-frames chat.md layout: the transcript rows have no border, the
	// composer frame of one input row is on the 3 rows above the status row.
	for index, want := range map[int]string{height - 4: "╭╮", height - 3: "││", height - 2: "╰╯"} {
		row := []rune(lines[index])
		if string(row[0]) != want[:len("╭")] || string(row[width-1]) != want[len("╭"):] {
			t.Fatalf("row %d = %q, want the composer frame", index, lines[index])
		}
	}
	for index := range height - 4 {
		if strings.ContainsAny(lines[index], "│╭╰") {
			t.Fatalf("transcript row %d has a border: %q", index, lines[index])
		}
	}
	if !strings.Contains(view, newest) {
		t.Fatalf("the newest message %q is not visible:\n%s", newest, view)
	}
}

func TestChatFitsEachSize(t *testing.T) {
	long := "alpha " + strings.Repeat("x", 150) + " omega"
	for _, size := range []struct{ width, height int }{{80, 24}, {80, 25}, {119, 40}, {120, 40}, {120, 41}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			checkFit(t, chatModel(t, 120, 40, 40), size.width, size.height, "line-40")
			checkFit(t, chatModel(t, 120, 40, 0), size.width, size.height, "type a message to the coordinator.")
			m := chatModel(t, 120, 40, 3)
			m.chatData.entries = append(m.chatData.entries, chatlib.Entry{ID: "u", At: "2026-10-20T11:00:00Z", From: "user", Text: long, Delivery: "sent"})
			m.rebuildChat()
			checkFit(t, m, size.width, size.height, "omega")
		})
	}
	m := chatModel(t, 120, 40, 40)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 79, Height: 24})
	requireView(t, m, "terminal too small")
}

func TestChatStartFromProjectsShowsRefusalAndBlocksInput(t *testing.T) {
	home := newTestHome(t)
	projectDir := filepath.Join(home, "projects", "demo", "worktrees", "task")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := New(home, clock.now)
	t.Cleanup(m.Stop)
	// Decision 4: the refusal is in the one footer row, so the pane is wide
	// enough to show all of it.
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 200, Height: 40})
	requireView(t, m, "chat send is for the user only", "rule 2")
	// Golden-frames decision 19: the input has the focus at start, except
	// for a blocked caller, who cannot type.
	m, _ = raise(t, m, key('i', "i"))
	if m.chatInputActive || m.chatInput.Value() != "" {
		t.Fatalf("Chat input took a key for a caller below projects/: active %v, text %q", m.chatInputActive, m.chatInput.Value())
	}
}

func TestChatRefusedSendShowsNoticeAndKeepsDraft(t *testing.T) {
	home := newTestHome(t)
	writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, PaneID: "w1:p1", WorkspaceID: "w1",
		Status: records.AgentLive, Generation: 1,
	})
	writeRecord(t, home, []string{"claims"}, "coordinator", records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
	projectDir := filepath.Join(home, "projects", "demo", "worktrees", "task")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := New(home, clock.now)
	t.Cleanup(m.Stop)
	// Rule 12 refuses a send while no live coordinator is in the header.
	state, err := readHeader(home)
	if err != nil {
		t.Fatal(err)
	}
	m.header = state
	m.chatInputActive = true
	m.chatInput.Focus()
	m.chatInput.InsertString("hello")
	msg := runChatSendCommand(t, m.sendChat())
	m, _ = raise(t, m, msg)
	requireView(t, m, "chat send is for the user only", "hello")
	rejectView(t, m, "› you")
	if got := m.chatInput.Value(); got != "hello" {
		t.Fatalf("draft = %q; want hello", got)
	}
}

func TestChatSendRunsInBackgroundAndKeepsFailedDraft(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, PaneID: "w1:p1", Status: records.AgentLive, Generation: 1,
	})
	writeRecord(t, home, []string{"claims"}, "coordinator", records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
	fake := newBlockingChatHerdr(t)
	m := newTestModel(t, home, clock)
	state, err := readHeader(home)
	if err != nil {
		t.Fatal(err)
	}
	m.header = state
	for _, character := range "hello" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	m, cmd := raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter did not start a send command")
	}
	batchMsg, ok := cmd().(tea.BatchMsg)
	if !ok || len(batchMsg) != 2 {
		t.Fatalf("send command = %#v; want log and send commands", cmd)
	}
	_ = batchMsg[0]()
	done := make(chan tea.Msg, 1)
	go func() { done <- batchMsg[1]() }()
	select {
	case <-fake.started:
	case <-time.After(time.Second):
		t.Fatal("fake send did not start")
	}
	m, second := raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if second != nil || fake.prompts.Load() != 1 {
		t.Fatalf("second enter started another send: command=%v sends=%d", second != nil, fake.prompts.Load())
	}
	m, _ = raise(t, m, key('8', "8"))
	requireView(t, m, "hello8")
	fake.releasePrompt()
	m, _ = raise(t, m, <-done)
	requireView(t, m, "stalled")
	requireView(t, m, "hello8")
	m = runIdleTick(t, m)
	requireView(t, m, "stalled")
	if fake.prompts.Load() != 1 {
		t.Fatalf("send retried without user input; calls=%d", fake.prompts.Load())
	}
}

func TestChatProgramRunsWithStringInput(t *testing.T) {
	home := newTestHome(t)
	var out strings.Builder
	done := make(chan error, 1)
	go func() { done <- RunHome(home, strings.NewReader("\x03\x03"), &out) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunHome: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("RunHome did not quit after the second ctrl+c")
	}
	// Golden-frames chat.md rules 8 and 10: the welcome block and the
	// placeholder of no coordinator.
	if !strings.Contains(out.String(), "type a message to the coordinator.") || !strings.Contains(out.String(), "> no coordinator · ^y health, r restarts it") {
		t.Fatalf("program output omitted Chat:\n%s", out.String())
	}
}

func runChatSendCommand(t *testing.T, command tea.Cmd) tea.Msg {
	t.Helper()
	if command == nil {
		t.Fatal("send returned no command")
	}
	result := command()
	if batch, ok := result.(tea.BatchMsg); ok {
		for _, task := range batch {
			if msg := task(); msg != nil {
				if _, isSend := msg.(chatSendMsg); isSend {
					return msg
				}
			}
		}
		t.Fatal("send command returned no chat result")
	}
	if _, ok := result.(chatSendMsg); !ok {
		t.Fatalf("send command returned %T; want chatSendMsg", result)
	}
	return result
}

func loadChat(t *testing.T, m *Model) *Model {
	t.Helper()
	cmd := m.startCurrentScreenLoad()
	if cmd == nil {
		t.Fatal("Chat data request did not start")
	}
	m, _ = raise(t, m, cmd())
	return m
}

func appendChatFixture(t *testing.T, home string, entry chatlib.Entry) {
	t.Helper()
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".hive", "state", "chat", entry.At[:10]+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func appendInboxFixture(t *testing.T, home string, event records.Inbox) {
	t.Helper()
	writeRecord(t, home, []string{"inbox", "coordinator"}, fmt.Sprint(event.Seq), event)
}

type blockingChatHerdr struct {
	listener net.Listener
	started  chan struct{}
	release  chan struct{}
	fail     bool
	prompts  atomic.Int32
	once     sync.Once
}

func newBlockingChatHerdr(t *testing.T) *blockingChatHerdr {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	fake := &blockingChatHerdr{listener: listener, started: make(chan struct{}, 1), release: make(chan struct{}), fail: true}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(func() {
		fake.releasePrompt()
		_ = listener.Close()
	})
	return fake
}

func (f *blockingChatHerdr) serve() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *blockingChatHerdr) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return
	}
	var request struct {
		ID     string `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(line, &request); err != nil {
		return
	}
	switch request.Method {
	case "ping":
		_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"pong","version":"0.9.5","protocol":22}}`+"\n", request.ID)
	case "agent.prompt":
		f.prompts.Add(1)
		f.started <- struct{}{}
		<-f.release
		if f.fail {
			_, _ = fmt.Fprintf(conn, `{"id":%q,"error":{"code":"agent_prompt_stalled","message":"stalled"}}`+"\n", request.ID)
			return
		}
		_, _ = fmt.Fprintf(conn, `{"id":%q,"result":{"type":"agent_prompted","agent":{"pane_id":"w1:p1","terminal_id":"term_a","workspace_id":"w1","tab_id":"w1:t1","agent_status":"idle","focused":false,"revision":1}}}`+"\n", request.ID)
	}
}

func (f *blockingChatHerdr) releasePrompt() { f.once.Do(func() { close(f.release) }) }

// At both caps, Left must move the rendered cursor before the final D.
// The text widths are 74 and 114 cells. Each last wrapped row is xEND.
func TestChatWrappedInputCursorTracksInsertion(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		width, height, prefix int
	}{
		{"compact cap", 80, 24, 445},
		{"full cap", 120, 40, 1141},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := chatModel(t, tc.width, tc.height, 0)
			defer m.Stop()
			m, _ = raise(t, m, key('i', "i"))
			m.chatInput.SetValue(strings.Repeat("x", tc.prefix) + "END")
			m.invalidateView()
			before := ansi.Strip(m.View().Content)
			if !strings.Contains(before, "xEND█") {
				t.Fatalf("initial cursor is not after END:\n%s", before)
			}
			m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyLeft})
			after := ansi.Strip(m.View().Content)
			if !strings.Contains(after, "xEN█") {
				t.Errorf("Left must put the cursor before D; expected xEN█:\n%s", after)
			}
			m, _ = raise(t, m, key('Z', "Z"))
			if !strings.HasSuffix(m.chatInput.Value(), "xENZD") {
				t.Fatalf("Z was not inserted before D: %q", m.chatInput.Value())
			}
			after = ansi.Strip(m.View().Content)
			if !strings.Contains(after, "xENZ█") {
				t.Errorf("cursor must stay before D after inserting Z; expected xENZ█:\n%s", after)
			}
		})
	}
}

// Text that wraps at spaces must keep the cursor on the insertion point.
// At 80×24 the text width is 74 cells. Each row holds 15 "abcd" words, and
// the third row is "ZEND".
func TestChatWordWrappedInputCursorTracksInsertion(t *testing.T) {
	m := chatModel(t, 80, 24, 0)
	defer m.Stop()
	m, _ = raise(t, m, key('i', "i"))
	m.chatInput.SetValue(strings.Repeat("abcd ", 30) + "END")
	m.invalidateView()
	for range 3 {
		m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyLeft})
	}
	m, _ = raise(t, m, key('Z', "Z"))
	if !strings.HasSuffix(m.chatInput.Value(), "abcd ZEND") {
		t.Fatalf("Z was not inserted before END: %q", m.chatInput.Value())
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "│   Z█ND ") {
		t.Errorf("cursor must sit on E after inserting Z; expected Z█ND:\n%s", view)
	}
}

// Golden-frames chat.md rule 8: the welcome block names the home with ~ for
// the user home folder, the state, the first hint, the chords, and ? help.
func TestChatWelcomeNamesTheHomeWithTilde(t *testing.T) {
	home := newTestHome(t)
	t.Setenv("HOME", filepath.Dir(home))
	m := newTestModel(t, home, fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)))
	t.Cleanup(m.Stop)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	rows := viewRows(m)
	want := []string{" hive · ~/" + filepath.Base(home), " ○ none", " type a message to the coordinator.", " ^g agents  ^t work  ^f friction  ^y health  ^l logs", " ? help"}
	for index, line := range want {
		if got := strings.TrimRight(rows[index+1], " "); got != line {
			t.Errorf("row %d = %q, want %q", index+1, got, line)
		}
	}
}

// Golden-frames chat.md rule 12 and decision 14: enter with no live
// coordinator shows error: no live coordinator on the notice line above the
// composer, sends nothing, and keeps the text.
func TestChatSendWithNoCoordinatorKeepsTheText(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	for _, character := range "hi" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.chatSending || m.chatInput.Value() != "hi" {
		t.Fatalf("sending %v, draft %q, want no send and hi kept", m.chatSending, m.chatInput.Value())
	}
	rows := viewRows(m)
	if got := strings.TrimRight(rows[35], " "); got != " error: no live coordinator" {
		t.Fatalf("row 35 = %q, want the notice line", got)
	}
	if !strings.HasPrefix(rows[36], "╭") {
		t.Fatalf("row 36 = %q, want the composer below the notice", rows[36])
	}
}

// Golden-frames decisions 11 and 19: the screen segment on Chat shows the
// keys of the input line from the chat.md key table. ctrl+] has no function
// on Chat, so the segment does not show it.
func TestChatScreenSegmentShowsTheInputLineKeys(t *testing.T) {
	for _, width := range []int{120, 80} {
		m := chatModel(t, width, 40, 1)
		cfg := m.tuiConfig
		cfg.StatusLines = [][]string{{"screen"}}
		cfg.StatusLinesExplicit = true
		m.setTUIConfig(cfg)
		row := strings.TrimRight(viewRow(m, 39), " ")
		if !strings.HasPrefix(row, " chat · enter send · ") || strings.Contains(row, "ctrl+]") {
			t.Errorf("status row at %d = %q, want the chat keys with no ctrl+]", width, row)
		}
	}
}
