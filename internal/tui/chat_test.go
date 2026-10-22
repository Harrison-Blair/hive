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

// pressScreenKey gives a key to the current screen of the tab. The frame
// gives some keys a new meaning: decision 24 gives a printable key in the
// Chat feed to the input, and decision 6 gives tab to the regions. A test of
// a key of a current screen calls the screen directly until wave 6 replaces
// the screen.
func pressScreenKey(t *testing.T, m *Model, press tea.KeyPressMsg) (*Model, tea.Cmd) {
	t.Helper()
	cmd := m.currentScreen().Update(press)
	if cmd != nil {
		m, _ = raise(t, m, cmd())
	}
	m.invalidateView()
	return m, cmd
}

// The fixture boards scroll the feed to the message at ui.scroll.feed.top,
// and the wide feed counts the messages above the view in its top border.
func TestChatFixtureScrollPutsTheMessageAtTheTop(t *testing.T) {
	m := newTestDemoModel(t, "Chat.fixture.json", demoSize{})
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	if !strings.Contains(lines[3], "chat · noise normal") || !strings.Contains(lines[3], "↑ 42 earlier") {
		t.Fatalf("feed title row = %q", lines[3])
	}
	if strings.TrimSpace(strings.Trim(lines[4], "│┃")) != "" || !strings.HasPrefix(lines[5], "│ 14:02  you           I want") {
		t.Fatalf("feed does not start with a blank row and message 43:\n%s", strings.Join(lines[3:8], "\n"))
	}
}

// Normal mode loads the chat of the last seven days and the coordinator
// inbox, with the same renderer as demo mode (decisions 15, 37, 38).
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
	// Normal mode shows each time in the local zone, as the header does.
	local := func(at string) string {
		parsed, err := time.Parse(time.RFC3339, at)
		if err != nil {
			t.Fatal(err)
		}
		return parsed.Local().Format("15:04")
	}
	requireView(t, m,
		"│ "+local("2026-10-19T12:00:00Z")+"  you · pending first message",
		"│ "+local("2026-10-19T12:30:00Z")+"  ▸ worker-one  message first line",
		"│ "+local("2026-10-19T13:00:00Z")+"  coordinator   answer",
		"│ "+local("2026-10-19T14:00:00Z")+"  you           delivered message",
	)
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

	// R reloads the records of the screen (decision 22).
	appendChatFixture(t, home, chatlib.Entry{ID: "c5", At: "2026-10-20T11:00:00Z", From: "coordinator", Text: "a new reply"})
	m, cmd := raise(t, m, key('R', "R"))
	m = runLoads(t, m, cmd)
	requireView(t, m, "coordinator   a new reply")
}

// runLoads runs cmd and gives each screen load message to m.
func runLoads(t *testing.T, m *Model, cmd tea.Cmd) *Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, next := range msg {
			m = runLoads(t, m, next)
		}
	case screenLoadMsg:
		m, _ = raise(t, m, msg)
	}
	return m
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

// The feed follows its end. pgup and pgdn page it from the feed and from the
// input, and j k g G move it from the feed (decisions 21, 22, 24).
func TestChatKeysScrollTheFeed(t *testing.T) {
	m := chatModel(t, 120, 40, 40)
	requireView(t, m, "line-40")
	rejectView(t, m, "line-01")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyPgUp})
	rejectView(t, m, "line-40")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	requireView(t, m, "line-40")
	m, _ = raise(t, m, key('g', "g"))
	requireView(t, m, "line-01")
	rejectView(t, m, "line-40")
	m, _ = raise(t, m, key('j', "j"))
	m, _ = raise(t, m, key('j', "j"))
	rejectView(t, m, "line-01")
	m, _ = raise(t, m, key('k', "k"))
	m, _ = raise(t, m, key('k', "k"))
	requireView(t, m, "line-01")
	m, _ = raise(t, m, key('G', "G"))
	requireView(t, m, "line-40")
	if m.chatInputActive {
		t.Fatal("a feed key gave the focus to the input")
	}

	m, _ = raise(t, m, key('i', "i"))
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyPgUp})
	rejectView(t, m, "line-40")
	if !m.chatInputActive {
		t.Fatal("pgup in the input moved the focus")
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	requireView(t, m, "line-40")
}

// Decision 24: a printable key in the feed gives the focus to the input and
// types the key. i and enter give the focus. 1–6 and ? stay global in the
// feed. In the input each printable key types, esc gives the focus to the
// feed, and tab moves the region. The footer shows the keys of each focus
// state (decision 21).
func TestChatKeysFollowTheFocusRules(t *testing.T) {
	m := chatModel(t, 120, 40, 1)
	requireView(t, m, "[feed] i or type to write · j/k scroll · g/G top/end · ctrl+o open events · tab region · ? help")
	m, _ = raise(t, m, key('x', "x"))
	if !m.chatInputActive {
		t.Fatal("a printable key in the feed did not give the focus to the input")
	}
	requireView(t, m, "> x█", "[input] enter send · shift+enter newline · ctrl+o open events · esc leave input · pgup/pgdn scroll feed · tab region")
	for _, character := range "12?/jkgG" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	if m.activeTab() != tabChat || m.helpOpen {
		t.Fatalf("a key in the input left the input: tab %d, help %v", m.activeTab(), m.helpOpen)
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModShift})
	for _, character := range "two" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	if got := m.chatInput.Value(); got != "x12?/jkgG\ntwo" {
		t.Fatalf("draft = %q", got)
	}
	requireView(t, m, "  two█")

	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.chatInputActive {
		t.Fatal("esc did not give the focus to the feed")
	}
	requireView(t, m, "[feed]", "  two")
	rejectView(t, m, "two█")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.chatInputActive {
		t.Fatal("enter in the feed did not give the focus to the input")
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.chatInputActive {
		t.Fatal("tab in the input did not move the focus to the feed")
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyTab})
	if !m.chatInputActive {
		t.Fatal("tab in the feed did not move the focus to the input")
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	m, _ = raise(t, m, key('?', "?"))
	if !m.helpOpen {
		t.Fatal("? in the feed did not open help")
	}
	m, _ = raise(t, m, key('?', "?"))
	m, _ = raise(t, m, key('3', "3"))
	if m.activeTab() != tabWork {
		t.Fatalf("3 in the feed gave tab %d", m.activeTab())
	}
	if got := m.chatInput.Value(); got != "x12?/jkgG\ntwo" {
		t.Fatalf("draft after the feed keys = %q", got)
	}
}

// Below 120 columns the footer shows the compact hints of each focus state.
func TestChatCompactFooterHints(t *testing.T) {
	m := chatModel(t, 80, 24, 1)
	requireView(t, m, "[feed] i type · pgup/pgdn feed · ctrl+o events · tab region · ?")
	m, _ = raise(t, m, key('i', "i"))
	requireView(t, m, "[input] enter send · ctrl+o events · esc leave input · pgup/pgdn feed")
}

func TestChatInputBorderShowsCoordinatorStatusAndQueuedMark(t *testing.T) {
	home := newTestHome(t)
	writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Status: records.AgentLive,
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
	got := ansi.Strip(m.View().Content)
	if !strings.Contains(got, "message · ● working 2m · queued") {
		t.Fatalf("input border does not show status and pending delivery:\n%s", got)
	}
}

func TestChatInputShowsStatusGlyphAndWholeMinuteAge(t *testing.T) {
	cases := []struct {
		name, status, changedAt, want string
	}{
		{"working zero", "working", "2026-10-20T12:00:00Z", "message · ● working 0m"},
		{"working one", "working", "2026-10-20T11:59:00Z", "message · ● working 1m"},
		{"working fifty-nine", "working", "2026-10-20T11:01:00Z", "message · ● working 59m"},
		{"idle", "idle", "2026-10-20T11:59:00Z", "message · ○ idle 1m"},
		{"blocked", "blocked", "2026-10-20T12:00:00Z", "message · ▲ blocked 0m"},
		{"other", "waiting", "2026-10-20T11:59:00Z", "message · ○ waiting 1m"},
		{"no time", "idle", "", "message · ○ idle"},
		{"no live coordinator", "", "", "message · ○ none"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := newTestHome(t)
			if tc.name != "no live coordinator" {
				writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
					Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", Status: records.AgentLive,
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
			got := ansi.Strip(m.chatView(120, 36))
			if !strings.Contains(got, tc.want) {
				t.Fatalf("chat input border lacks %q:\n%s", tc.want, got)
			}
		})
	}
}

func TestChatInputCapsAndKeepsCursorVisibleAfterFocusLeaves(t *testing.T) {
	for _, size := range []struct{ width, height int }{{120, 40}, {80, 24}} {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			m := chatModel(t, size.width, size.height, 0)
			defer m.Stop()
			m.chatInput.SetValue(strings.Repeat("release notes stay readable ", 60))
			m.chatInputActive = true
			m.chatInput.Focus()
			m.invalidateView()
			before := strings.Split(ansi.Strip(m.View().Content), "\n")
			panelCap := (size.height - 4) / 3
			inputTop := size.height - panelCap - 1
			cursorRow := inputTop + panelCap - 2
			if !strings.HasPrefix(before[inputTop], "╭") || !strings.Contains(before[inputTop], "message") {
				t.Fatalf("capped input starts at row %d, got %q:\n%s", inputTop, before[inputTop], strings.Join(before, "\n"))
			}
			lastInputRow := []rune(before[cursorRow])
			if !strings.Contains(before[cursorRow], "█") || lastInputRow[size.width-1] != '┃' {
				t.Fatalf("input cursor or scroll mark is not visible at the cap (cursor=%v, height=%d):\n%s", m.chatInput.Cursor(), m.chatInput.Height(), strings.Join(before[inputTop:size.height-1], "\n"))
			}
			m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
			after := strings.Split(ansi.Strip(m.View().Content), "\n")
			if !strings.HasPrefix(after[inputTop], "╭") || !strings.HasPrefix(after[size.height-2], "╰") {
				t.Fatalf("input height changed after esc:\n%s", strings.Join(after[inputTop:size.height-1], "\n"))
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
	feedBottom, inputTop := height-5, height-4
	for index := 3; index < height-1; index++ {
		row := []rune(lines[index])
		first, last := "│", "│┃"
		switch index {
		case 3, inputTop:
			first, last = "╭", "╮"
		case feedBottom, height - 2:
			first, last = "╰", "╯"
		}
		if string(row[0]) != first || !strings.ContainsRune(last, row[width-1]) {
			t.Fatalf("row %d has a broken border %q", index, lines[index])
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
			checkFit(t, chatModel(t, 120, 40, 0), size.width, size.height, "no chat yet")
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
	m, cmd := raise(t, m, key('i', "i"))
	if cmd != nil || m.chatInputActive {
		t.Fatal("Chat input opened for a caller below projects/")
	}
}

func TestChatRefusedSendShowsNoticeAndKeepsDraft(t *testing.T) {
	home := newTestHome(t)
	writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", WorkspaceID: "w1",
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
	m.chatInputActive = true
	m.chatInput.Focus()
	m.chatInput.InsertString("hello")
	msg := runChatSendCommand(t, m.sendChat())
	m, _ = raise(t, m, msg)
	requireView(t, m, "chat send is for the user only", "hello")
	rejectView(t, m, "you ·")
	if got := m.chatInput.Value(); got != "hello" {
		t.Fatalf("draft = %q; want hello", got)
	}
}

func TestChatSendRunsInBackgroundAndKeepsFailedDraft(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	writeRecord(t, home, []string{"agents"}, "coordinator-g1", records.Agent{
		Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:p1", Status: records.AgentLive, Generation: 1,
	})
	writeRecord(t, home, []string{"claims"}, "coordinator", records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1})
	fake := newBlockingChatHerdr(t)
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, key('i', "i"))
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
	if !strings.Contains(out.String(), "no chat yet") || !strings.Contains(out.String(), "message") {
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
