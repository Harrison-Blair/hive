package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
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

func TestChatScreenShowsCoordinatorHeading(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, key('2', "2"))
	requireView(t, m, "coordinator", "hive events")
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

func TestChatLoadsRecentConversationAndCoordinatorState(t *testing.T) {
	home := newTestHome(t)
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	clock := fixedClock(t, now)
	writeRecord(t, home, []string{"agents"}, "coordinator-g2", records.Agent{
		Key: "coordinator-g2", Name: "coordinator", Role: "coordinator", Harness: "claude", Model: "claude-opus-5-5",
		PaneID: "w1:p1", Status: records.AgentLive, AgentStatus: "working", Generation: 2,
	})
	appendChatFixture(t, home, chatlib.Entry{
		ID: "c1", At: "2026-10-13T11:59:00Z", From: "user", Text: "outside the seven-day window", Delivery: "pending",
	})
	appendChatFixture(t, home, chatlib.Entry{
		ID: "c2", At: "2026-10-19T12:00:00Z", From: "user", Text: "first message", Delivery: "pending",
	})
	appendChatFixture(t, home, chatlib.Entry{
		ID: "c3", At: "2026-10-19T13:00:00Z", From: "coordinator", Text: "answer", ReplyTo: "c2",
	})
	appendChatFixture(t, home, chatlib.Entry{
		ID: "c4", At: "2026-10-19T14:00:00Z", From: "user", Text: "delivered message", Delivery: "pending",
	})
	appendChatFixture(t, home, chatlib.Entry{ID: "c4", At: "2026-10-19T14:00:01Z", Delivery: "sent"})
	appendInboxFixture(t, home, records.Inbox{
		Seq: 1, Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-one", Text: "message first line\nmessage second line",
		CreatedAt: "2026-10-19T12:30:00Z",
	})
	appendInboxFixture(t, home, records.Inbox{
		Seq: 2, Role: "coordinator", Kind: records.InboxCompleted, Agent: "worker-two", Text: "hidden by default noise",
		CreatedAt: "2026-10-19T12:45:00Z",
	})

	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 150, Height: 40})
	m = loadChat(t, m)
	requireView(t, m, "coordinator", "live", "claude", "working", "claude-opus-5-5", "w1:p1")
	requireView(t, m, "you · Oct 19", "first message", "pending", "coordinator · Oct 19", "answer", "↳ c2", "delivered message")
	rejectView(t, m, "outside the seven-day window", "[sent]", "message second line")
	requireView(t, m, "2 hive events", "message first line", "hidden by default noise")

	view := ansi.Strip(m.View().Content)
	if strings.Index(view, "answer") > strings.Index(view, "delivered message") {
		t.Fatalf("newest conversation entry is not at the bottom:\n%s", view)
	}
}

func TestChatFrameFitsLargeConversationAndWrapsAtWords(t *testing.T) {
	sizes := []struct {
		width  int
		height int
	}{{width: 120, height: 40}, {width: 200, height: 57}}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			home := newTestHome(t)
			clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
			for index := 1; index <= 200; index++ {
				text := fmt.Sprintf("entry-%03d", index)
				if index%5 == 0 {
					text += strings.Repeat(" olderword", 12)
				}
				if index == 200 {
					text = "alpha beta " + strings.Repeat("x", 80) + " omega"
				}
				at := time.Date(2026, 10, 20, 10, 0, index, 0, time.UTC).Format(time.RFC3339)
				appendChatFixture(t, home, chatlib.Entry{
					ID: fmt.Sprintf("c%d", index), At: at, From: "coordinator", Text: text,
				})
			}

			m := newTestModel(t, home, clock)
			m, _ = raise(t, m, tea.WindowSizeMsg{Width: size.width, Height: size.height})
			m = loadChat(t, m)
			view := ansi.Strip(m.View().Content)
			lines := strings.Split(view, "\n")
			if len(lines) != size.height {
				t.Fatalf("Chat frame has %d lines, want %d:\n%s", len(lines), size.height, view)
			}
			for index, line := range lines {
				if got := ansi.StringWidth(line); got != size.width {
					t.Fatalf("Chat frame line %d has width %d, want %d: %q", index+1, got, size.width, line)
				}
			}
			if !strings.Contains(view, "alpha beta") || !strings.Contains(view, "omega") {
				t.Fatalf("Chat frame omitted the newest wrapped entry:\n%s", view)
			}
			if strings.Contains(view, "entry-001") {
				t.Fatalf("Chat frame did not scroll older entries out at the top:\n%s", view)
			}
		})
	}
}

func TestChatNarrowViewportWrapsWordsAndBreaksLongWords(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	m.chatData.entries = []chatlib.Entry{{
		ID: "c1", At: "2026-10-20T11:00:00Z", From: "coordinator",
		Text: "alpha beta gamma abcdefghijklmnopqrstu omega",
	}}
	m.rebuildChat()
	m.chatView(34, 20)
	got := strings.Split(strings.TrimRight(ansi.Strip(m.chatViewport.View()), "\n"), "\n")
	for index := range got {
		got[index] = strings.TrimRight(got[index], " ")
	}
	want := []string{"  alpha beta", "gamma", "abcdefghijklmn", "opqrstu omega"}
	found := false
	for start := 0; start+len(want) <= len(got); start++ {
		if strings.Join(got[start:start+len(want)], "\n") == strings.Join(want, "\n") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Chat wrapped text as %q, want adjacent literal lines %q", got, want)
	}
}

func TestChatPagesScrollTheConversation(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	for index := 1; index <= 40; index++ {
		at := time.Date(2026, 10, 20, 0, index, 0, 0, time.UTC).Format(time.RFC3339)
		appendChatFixture(t, home, chatlib.Entry{
			ID: fmt.Sprintf("c%d", index), At: at, From: "coordinator", Text: fmt.Sprintf("line-%02d", index),
		})
	}
	m := newTestModel(t, home, clock)
	// Decision 7: 80×24 is the smallest frame, so the feed has 40 lines.
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m = loadChat(t, m)
	requireView(t, m, "line-40")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyPgUp})
	rejectView(t, m, "line-40")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyPgDown})
	requireView(t, m, "line-40")
}

func TestChatNoiseAndEventFoldingDoNotMarkRecordsRead(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	appendChatFixture(t, home, chatlib.Entry{ID: "c1", At: "2026-10-20T10:00:00Z", From: "user", Text: "before", Delivery: "sent"})
	appendChatFixture(t, home, chatlib.Entry{ID: "c2", At: "2026-10-20T11:00:00Z", From: "coordinator", Text: "after"})
	appendInboxFixture(t, home, records.Inbox{
		Seq: 1, Role: "coordinator", Kind: records.InboxMessage, Agent: "older-sender", Text: "older event\nexpanded detail",
		CreatedAt: "2026-10-20T10:15:00Z",
	})
	appendInboxFixture(t, home, records.Inbox{
		Seq: 2, Role: "coordinator", Kind: records.InboxBlocked, Agent: "newer-sender", Text: "newer event",
		CreatedAt: "2026-10-20T10:30:00Z",
	})
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 150, Height: 40})
	m = loadChat(t, m)
	requireView(t, m, "2 hive events", "older event", "newer event")

	m, _ = pressScreenKey(t, m, key('n', "n"))
	requireView(t, m, "noise: decisions only", "2 hive events", "older event", "newer event")
	m, _ = pressScreenKey(t, m, key('n', "n"))
	allEvents := ansi.Strip(m.View().Content)
	if strings.Index(allEvents, "newer-sender") > strings.Index(allEvents, "older-sender") {
		t.Fatalf("hive events are not newest first:\n%s", allEvents)
	}
	requireView(t, m, "2 hive events", "newer event")

	m, _ = pressScreenKey(t, m, key('e', "e"))
	requireView(t, m, "hive event ·", "expanded detail")
	m, _ = pressScreenKey(t, m, key('e', "e"))
	requireView(t, m, "── 2 hive events ──")
	rejectView(t, m, "expanded detail")
	readBack, err := inboxlib.All(home, "coordinator")
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range readBack {
		if event.ReadAt != "" {
			t.Fatalf("Chat marked inbox record %d read", event.Seq)
		}
	}
}

func TestChatNoiseLevelsFilterInboxKinds(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	for index, event := range []records.Inbox{
		{Kind: records.InboxMessage, Agent: "decision", Text: "D"},
		{Kind: records.InboxCompleted, Agent: "completed", Text: "C"},
		{Kind: records.InboxVerified, Agent: "verified", Text: "V"},
		{Kind: records.InboxBlocked, Agent: "blocked", Text: "B"},
		{Kind: records.InboxExited, Agent: "exited", Text: "E"},
		{Kind: records.InboxQuiet, Agent: "quiet", Text: "Q"},
		{Kind: records.InboxUnresolved, Agent: "unresolved", Text: "U"},
	} {
		event.Seq = uint64(index + 1)
		event.Role = "coordinator"
		event.CreatedAt = time.Date(2026, 10, 20, 10, index, 0, 0, time.UTC).Format(time.RFC3339)
		appendInboxFixture(t, home, event)
	}
	m := newTestModel(t, home, clock)
	m, _ = raise(t, m, tea.WindowSizeMsg{Width: 200, Height: 57})
	m = loadChat(t, m)

	assertKinds := func(t *testing.T, want []string) {
		t.Helper()
		got := make([]string, 0, len(m.visibleChatEvents()))
		for _, event := range m.visibleChatEvents() {
			got = append(got, event.Kind)
		}
		expected := append([]string(nil), want...)
		sort.Strings(got)
		sort.Strings(expected)
		if len(got) != len(expected) {
			t.Fatalf("visible inbox kinds = %v; want %v", got, expected)
		}
		for index := range got {
			if got[index] != expected[index] {
				t.Fatalf("visible inbox kinds = %v; want %v", got, expected)
			}
		}
	}

	t.Run("landings and decisions", func(t *testing.T) {
		assertKinds(t, []string{"message", "blocked", "unresolved", "completed", "verified", "quiet"})
		requireView(t, m, "noise: landings and decisions", "6 hive events", "decision: D", "blocked: B", "unresolved: U", "completed: C", "verified: V", "quiet: Q")
		rejectView(t, m, "exited: E")
	})

	m, _ = pressScreenKey(t, m, key('n', "n"))
	t.Run("decisions only", func(t *testing.T) {
		assertKinds(t, []string{"message", "blocked", "unresolved"})
		requireView(t, m, "noise: decisions only", "3 hive events", "decision: D", "blocked: B", "unresolved: U")
		rejectView(t, m, "completed: C", "verified: V", "exited: E", "quiet: Q")
	})

	m, _ = pressScreenKey(t, m, key('n', "n"))
	t.Run("all", func(t *testing.T) {
		assertKinds(t, []string{"message", "completed", "verified", "blocked", "exited", "quiet", "unresolved"})
		requireView(t, m, "noise: all", "7 hive events", "decision: D", "completed: C", "verified: V", "blocked: B", "exited: E", "quiet: Q", "unresolved: U")
	})

	m, _ = pressScreenKey(t, m, key('n', "n"))
	t.Run("returns to landings and decisions", func(t *testing.T) {
		assertKinds(t, []string{"message", "blocked", "unresolved", "completed", "verified", "quiet"})
		requireView(t, m, "noise: landings and decisions", "6 hive events", "decision: D", "blocked: B", "unresolved: U", "completed: C", "verified: V", "quiet: Q")
		rejectView(t, m, "exited: E")
	})
}

func TestChatInputKeepsDraftAndTreatsScreenNumbersAsText(t *testing.T) {
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	m := newTestModel(t, home, clock)
	// Decision 21: the footer shows the keys of the feed, then of the input.
	requireView(t, m, "[feed]", "i type", "pgup/pgdn feed")
	m, _ = raise(t, m, key('i', "i"))
	requireView(t, m, "[input]", "enter send", "esc leave input")
	for _, character := range "line1" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModShift})
	for _, character := range "line2" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt})
	for _, character := range "line3" {
		m, _ = raise(t, m, key(character, string(character)))
	}
	for number := rune('1'); number <= '8'; number++ {
		m, _ = raise(t, m, key(number, string(number)))
	}
	if m.activeTab() != tabChat {
		t.Fatalf("a number key in the Chat input changed the tab to %d", m.activeTab())
	}
	requireView(t, m, "line1", "line2", "line312345678")
	m, _ = raise(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	requireView(t, m, "line1", "line2", "line312345678")
	rejectView(t, m, "[input]")
	m, _ = raise(t, m, key('i', "i"))
	requireView(t, m, "line1", "line2", "line312345678")
	m, _ = raise(t, m, ctrlC())
	requireView(t, m, "press ctrl+c again to quit")
	requireView(t, m, "line1", "line2", "line312345678")
	_, cmd := raise(t, m, ctrlC())
	quitMessage(t, cmd)
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
	if !strings.Contains(out.String(), "coordinator") || !strings.Contains(out.String(), "hive events") {
		t.Fatalf("program output omitted Chat:\n%s", out.String())
	}
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
