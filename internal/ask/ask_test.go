package ask

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func askHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

func seedAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatal(err)
	}
}

// seedPair writes a live parent agent and a live worker agent whose parent
// is that agent.
func seedPair(t *testing.T, home string) {
	t.Helper()
	seedAgent(t, home, records.Agent{Key: "lead-one", Name: "lead-one", Role: "worker", PaneID: "w1:p1", Status: records.AgentLive, AgentStatus: "idle", Generation: 1})
	seedAgent(t, home, records.Agent{Key: "worker-one", Name: "worker-one", Role: "worker", PaneID: "w2:p1", Status: records.AgentLive, AgentStatus: "working", Parent: "lead-one", Generation: 1})
}

func stopAgent(t *testing.T, home, key string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatal(err)
	}
	record.Status = records.AgentStopped
	if err := registry.WriteIn([]string{"agents"}, key, record); err != nil {
		t.Fatal(err)
	}
}

const oneQuestion = `{"questions":[{"question":"Restart now?","header":"Model","options":[
 {"label":"Yes (REC)","description":"stop it"},
 {"label":"No","description":"no change"}]}]}`

const twoQuestions = `{"questions":[
 {"question":"Which parts?","header":"Scope","multiSelect":true,"options":[
  {"label":"Mouse wheel","description":"scroll"},
  {"label":"Click","description":"focus"},
  {"label":"Modes","description":"i in"}]},
 {"question":"Which first?","header":"Order","options":[
  {"label":"Modes first","description":"new modes","preview":"one\ntwo"},
  {"label":"Card first","description":"later"}]}]}`

func askFiles(t *testing.T, home string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(home, ".hive", "state", "asks", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func readRecord(t *testing.T, home, id string) Record {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "asks", id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}

func TestCreateWritesOneOpenRecordAndReturnsAtOnce(t *testing.T) {
	home := askHome(t)
	seedPair(t, home)
	fake := newFakeHerdr(t)
	started := time.Now()
	envelope := Create(context.Background(), home, CreateOptions{Agent: "worker-one", Data: []byte(oneQuestion)})
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("Create took %s; want an immediate return", elapsed)
	}
	if envelope.Outcome != "ok" {
		t.Fatalf("Create() = %#v", envelope)
	}
	result, ok := envelope.Result.(Record)
	if !ok || result.ID != "a1" {
		t.Fatalf("result = %#v; want ask a1", envelope.Result)
	}
	if files := askFiles(t, home); len(files) != 1 || filepath.Base(files[0]) != "a1.json" {
		t.Fatalf("ask files = %v; want a1.json", files)
	}
	record := readRecord(t, home, "a1")
	if record.Asker != "worker-one" || record.AskerKey != "worker-one" || record.Parent != "lead-one" || record.ParentKey != "lead-one" {
		t.Fatalf("asker and parent = %q %q %q %q", record.Asker, record.AskerKey, record.Parent, record.ParentKey)
	}
	if record.State != StateOpen || record.CreatedAt == "" || len(record.Questions) != 1 || record.Questions[0].Header != "Model" {
		t.Fatalf("record = %#v", record)
	}
	if got := len(fake.recorded("agent.prompt")); got != 0 {
		t.Fatalf("Create sent %d prompts; want none", got)
	}
	if got := Text(envelope); got != "ask a1 open\n" {
		t.Fatalf("Text() = %q", got)
	}
}

func TestCreateRejectsEachLimitWithNoRecord(t *testing.T) {
	option := `{"label":"L","description":"D"}`
	options := func(n int) string { return strings.TrimSuffix(strings.Repeat(option+",", n), ",") }
	question := func(header, opts string) string {
		return fmt.Sprintf(`{"question":"Q?","header":%q,"options":[%s]}`, header, opts)
	}
	for _, tc := range []struct {
		name, data, want string
	}{
		{"not JSON", `{`, "ask JSON"},
		{"unknown field", `{"questions":[` + question("H", options(2)) + `],"extra":1}`, `unknown field "extra"`},
		{"no question", `{"questions":[]}`, "1 to 4 questions"},
		{"five questions", `{"questions":[` + strings.TrimSuffix(strings.Repeat(question("H", options(2))+",", 5), ",") + `]}`, "1 to 4 questions"},
		{"long header", `{"questions":[` + question("thirteen char", options(2)) + `]}`, "header of 12 characters at most"},
		{"empty header", `{"questions":[` + question("", options(2)) + `]}`, "header"},
		{"one option", `{"questions":[` + question("H", options(1)) + `]}`, "2 to 5 options"},
		{"six options", `{"questions":[` + question("H", options(6)) + `]}`, "2 to 5 options"},
		{"empty question", `{"questions":[{"question":"","header":"H","options":[` + options(2) + `]}]}`, "question text"},
		{"empty label", `{"questions":[` + question("H", `{"label":"","description":"D"},`+option) + `]}`, "label"},
		{"empty description", `{"questions":[` + question("H", `{"label":"L","description":""},`+option) + `]}`, "description"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := askHome(t)
			seedPair(t, home)
			envelope := Create(context.Background(), home, CreateOptions{Agent: "worker-one", Data: []byte(tc.data)})
			if envelope.Outcome != "failed" || envelope.Phase != "validate" || !strings.Contains(fmt.Sprint(envelope.Error), tc.want) {
				t.Fatalf("Create() = %#v; want a validate error with %q", envelope, tc.want)
			}
			if files := askFiles(t, home); len(files) != 0 {
				t.Fatalf("ask files = %v; want none", files)
			}
		})
	}
}

func TestCreateNeedsALiveAsker(t *testing.T) {
	home := askHome(t)
	envelope := Create(context.Background(), home, CreateOptions{Agent: "nobody", Data: []byte(oneQuestion)})
	if envelope.Outcome != "failed" || envelope.Phase != "identity" {
		t.Fatalf("Create() = %#v; want an identity error", envelope)
	}
	if files := askFiles(t, home); len(files) != 0 {
		t.Fatalf("ask files = %v; want none", files)
	}
}

func TestAnswerTextOfOneQuestion(t *testing.T) {
	record := Record{ID: "a7", Asker: "manager-hive", Questions: mustParse(t, oneQuestion).Questions}
	got := AnswerText(record, []Reply{{Picks: []int{1}}}, false)
	want := "⬢ user: answer to ask a7 from manager-hive · Model → Yes (REC)"
	if got != want {
		t.Fatalf("AnswerText() = %q\nwant %q", got, want)
	}
}

func TestAnswerTextOfTwoQuestionsWithANoteAndOther(t *testing.T) {
	record := Record{ID: "a8", Asker: "coordinator", Questions: mustParse(t, twoQuestions).Questions}
	replies := []Reply{
		{Picks: []int{1, 4}, Other: "Header count", Note: "keep it small"},
		{Picks: []int{2}},
	}
	got := AnswerText(record, replies, false)
	want := "⬢ user: answer to ask a8 from coordinator · Scope → Mouse wheel, Header count · note: keep it small; Order → Card first"
	if got != want {
		t.Fatalf("AnswerText() = %q\nwant %q", got, want)
	}
}

func TestAnswerTextOfAStoppedAsker(t *testing.T) {
	record := Record{ID: "a9", Asker: "worker-one", Questions: mustParse(t, oneQuestion).Questions}
	got := AnswerText(record, []Reply{{Picks: []int{3}, Other: "later"}}, true)
	want := "⬢ user: answer to ask a9 from worker-one (asker stopped) · Model → later"
	if got != want {
		t.Fatalf("AnswerText() = %q\nwant %q", got, want)
	}
}

func TestDismissText(t *testing.T) {
	record := Record{ID: "a3", Asker: "coordinator", Questions: mustParse(t, twoQuestions).Questions}
	want := "⬢ hive: the user dismissed ask a3 from coordinator · Scope · Order. No answer comes."
	if got := DismissText(record); got != want {
		t.Fatalf("DismissText() = %q\nwant %q", got, want)
	}
}

func mustParse(t *testing.T, data string) Request {
	t.Helper()
	request, err := Parse([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func createAsk(t *testing.T, home, data string) Record {
	t.Helper()
	envelope := Create(context.Background(), home, CreateOptions{Agent: "worker-one", Data: []byte(data)})
	record, ok := envelope.Result.(Record)
	if envelope.Outcome != "ok" || !ok {
		t.Fatalf("Create() = %#v", envelope)
	}
	return record
}

func TestAnswerSendsThePromptToTheAskerPane(t *testing.T) {
	home := askHome(t)
	seedPair(t, home)
	fake := newFakeHerdr(t)
	record := createAsk(t, home, oneQuestion)
	envelope := Answer(context.Background(), home, record.ID, []Reply{{Picks: []int{2}}})
	if envelope.Outcome != "ok" {
		t.Fatalf("Answer() = %#v", envelope)
	}
	prompts := fake.prompts()
	want := promptParams{Target: "w2:p1", Text: "⬢ user: answer to ask a1 from worker-one · Model → No"}
	if len(prompts) != 1 || prompts[0] != want {
		t.Fatalf("prompts = %#v; want %#v", prompts, want)
	}
	stored := readRecord(t, home, record.ID)
	if stored.State != StateAnswered || stored.DeliveredAt == "" || stored.DeliveredTo != "worker-one" || len(stored.Replies) != 1 {
		t.Fatalf("record = %#v", stored)
	}
}

func TestAnswerOfAStoppedAskerGoesToTheParent(t *testing.T) {
	home := askHome(t)
	seedPair(t, home)
	fake := newFakeHerdr(t)
	record := createAsk(t, home, oneQuestion)
	stopAgent(t, home, "worker-one")
	envelope := Answer(context.Background(), home, record.ID, []Reply{{Picks: []int{1}, Note: "now"}})
	if envelope.Outcome != "ok" {
		t.Fatalf("Answer() = %#v", envelope)
	}
	want := promptParams{Target: "w1:p1", Text: "⬢ user: answer to ask a1 from worker-one (asker stopped) · Model → Yes (REC) · note: now"}
	if prompts := fake.prompts(); len(prompts) != 1 || prompts[0] != want {
		t.Fatalf("prompts = %#v; want %#v", prompts, want)
	}
	if stored := readRecord(t, home, record.ID); stored.DeliveredTo != "lead-one" {
		t.Fatalf("delivered to %q; want lead-one", stored.DeliveredTo)
	}
}

func TestDismissSendsTheDismissPrompt(t *testing.T) {
	home := askHome(t)
	seedPair(t, home)
	fake := newFakeHerdr(t)
	record := createAsk(t, home, oneQuestion)
	envelope := Dismiss(context.Background(), home, record.ID)
	if envelope.Outcome != "ok" {
		t.Fatalf("Dismiss() = %#v", envelope)
	}
	want := promptParams{Target: "w2:p1", Text: "⬢ hive: the user dismissed ask a1 from worker-one · Model. No answer comes."}
	if prompts := fake.prompts(); len(prompts) != 1 || prompts[0] != want {
		t.Fatalf("prompts = %#v; want %#v", prompts, want)
	}
	if stored := readRecord(t, home, record.ID); stored.State != StateDismissed || stored.DeliveredAt == "" {
		t.Fatalf("record = %#v", stored)
	}
}

func TestAnswerRejectsAClosedAsk(t *testing.T) {
	home := askHome(t)
	seedPair(t, home)
	newFakeHerdr(t)
	record := createAsk(t, home, oneQuestion)
	if envelope := Dismiss(context.Background(), home, record.ID); envelope.Outcome != "ok" {
		t.Fatalf("Dismiss() = %#v", envelope)
	}
	envelope := Answer(context.Background(), home, record.ID, []Reply{{Picks: []int{1}}})
	if envelope.Outcome != "failed" || !strings.Contains(fmt.Sprint(envelope.Error), "not open") {
		t.Fatalf("Answer() = %#v; want a not-open error", envelope)
	}
}

func TestFailedDeliveryKeepsTheAnswerAndTriesAgain(t *testing.T) {
	home := askHome(t)
	seedPair(t, home)
	fake := newFakeHerdr(t)
	fake.setPromptFails(true)
	record := createAsk(t, home, oneQuestion)
	envelope := Answer(context.Background(), home, record.ID, []Reply{{Picks: []int{2}}})
	if envelope.Outcome != "partial" || envelope.Phase != "delivery" {
		t.Fatalf("Answer() = %#v; want a partial delivery result", envelope)
	}
	stored := readRecord(t, home, record.ID)
	if stored.State != StateAnswered || stored.DeliveredAt != "" || stored.Attempts != 1 || stored.LastError == "" || len(stored.Replies) != 1 {
		t.Fatalf("record after the failed delivery = %#v", stored)
	}

	// The asker is working, so the retry waits, as a nudge does.
	if err := DeliverPending(context.Background(), home); err != nil {
		t.Fatal(err)
	}
	if got := len(fake.prompts()); got != 1 {
		t.Fatalf("prompts = %d; want no retry while the asker works", got)
	}

	fake.setPromptFails(false)
	setAgentStatus(t, home, "worker-one", "idle")
	if err := DeliverPending(context.Background(), home); err != nil {
		t.Fatal(err)
	}
	prompts := fake.prompts()
	want := promptParams{Target: "w2:p1", Text: "⬢ user: answer to ask a1 from worker-one · Model → No"}
	if len(prompts) != 2 || prompts[1] != want {
		t.Fatalf("prompts = %#v; want the retry %#v", prompts, want)
	}
	stored = readRecord(t, home, record.ID)
	if stored.DeliveredAt == "" || stored.Attempts != 2 {
		t.Fatalf("record after the retry = %#v", stored)
	}

	// A delivered answer is not sent again.
	if err := DeliverPending(context.Background(), home); err != nil {
		t.Fatal(err)
	}
	if got := len(fake.prompts()); got != 2 {
		t.Fatalf("prompts = %d; want 2", got)
	}
}

func setAgentStatus(t *testing.T, home, key, status string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatal(err)
	}
	record.AgentStatus = status
	if err := registry.WriteIn([]string{"agents"}, key, record); err != nil {
		t.Fatal(err)
	}
}

func TestListGivesTheRecordsOldestFirst(t *testing.T) {
	home := askHome(t)
	seedPair(t, home)
	for range 11 {
		createAsk(t, home, oneQuestion)
	}
	listed, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 11 || listed[0].ID != "a1" || listed[1].ID != "a2" || listed[10].ID != "a11" {
		ids := make([]string, len(listed))
		for index, record := range listed {
			ids[index] = record.ID
		}
		t.Fatalf("ids = %v; want a1 to a11 in order", ids)
	}
}

func TestRouteGivesTheAskerOrTheParent(t *testing.T) {
	record := Record{Asker: "worker-one", AskerKey: "worker-one-g2", Parent: "lead", ParentKey: "lead-g1"}
	agents := []records.Agent{
		{Key: "other-g1", Name: "other", Status: records.AgentLive},
		{Key: "worker-one-g2", Name: "worker-one", Status: records.AgentLive},
		{Key: "lead-g1", Name: "lead", Status: records.AgentStopped},
		{Key: "lead-g2", Name: "lead", Status: records.AgentLive},
	}
	target, stopped, ok := Route(record, agents)
	if !ok || stopped || target.Key != "worker-one-g2" {
		t.Fatalf("Route() = %q %v %v; want the live asker", target.Key, stopped, ok)
	}
	agents[1].Status = records.AgentStopped
	target, stopped, ok = Route(record, agents)
	if !ok || !stopped || target.Key != "lead-g2" {
		t.Fatalf("Route() = %q %v %v; want the live parent lead-g2", target.Key, stopped, ok)
	}
	agents[3].Status = records.AgentStopped
	if _, stopped, ok = Route(record, agents); ok || !stopped {
		t.Fatalf("Route() = %v %v; want no target", stopped, ok)
	}
}

type promptParams struct {
	Target string `json:"target"`
	Text   string `json:"text"`
}

type herdrRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// fakeHerdr is the herdr socket. It records each request and answers
// agent.prompt.
type fakeHerdr struct {
	t         *testing.T
	listener  net.Listener
	mu        sync.Mutex
	requests  []herdrRequest
	promptErr bool
}

func newFakeHerdr(t *testing.T) *fakeHerdr {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeHerdr{t: t, listener: listener}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(func() { _ = listener.Close() })
	return fake
}

func (f *fakeHerdr) serve() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *fakeHerdr) handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return
	}
	var request herdrRequest
	if err := json.Unmarshal(line, &request); err != nil {
		f.t.Errorf("decode request: %v", err)
		return
	}
	f.mu.Lock()
	f.requests = append(f.requests, request)
	fail := f.promptErr
	f.mu.Unlock()
	body := fmt.Sprintf(`{"id":%q`, request.ID)
	switch {
	case request.Method == "ping":
		body += `,"result":{"type":"pong","version":"0.9.5","protocol":22}}`
	case request.Method == "agent.prompt" && fail:
		body += `,"error":{"code":"agent_prompt_stalled","message":"the pane waits for a prompt"}}`
	case request.Method == "agent.prompt":
		body += `,"result":{"type":"agent_prompted","agent":{"terminal_id":"t","agent_status":"working","workspace_id":"w1","tab_id":"w1:t1","pane_id":"w1:p1","focused":false,"revision":1}}}`
	default:
		f.t.Errorf("unexpected herdr method %q", request.Method)
		return
	}
	_, _ = conn.Write([]byte(body + "\n"))
}

func (f *fakeHerdr) recorded(method string) []herdrRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	var found []herdrRequest
	for _, request := range f.requests {
		if request.Method == method {
			found = append(found, request)
		}
	}
	return found
}

func (f *fakeHerdr) prompts() []promptParams {
	var out []promptParams
	for _, request := range f.recorded("agent.prompt") {
		var params promptParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			f.t.Fatal(err)
		}
		out = append(out, params)
	}
	return out
}

func (f *fakeHerdr) setPromptFails(fail bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.promptErr = fail
}
