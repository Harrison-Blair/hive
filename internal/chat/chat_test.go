package chat

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

	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func chatHome(t *testing.T) string {
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

func seedCoordinator(t *testing.T, home string) {
	t.Helper()
	coordinator := records.Agent{Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, PaneID: "w1:p1", WorkspaceID: "w1", Status: records.AgentLive, Generation: 1}
	seedAgent(t, home, coordinator)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, "coordinator", records.Claim{Key: "coordinator", Agent: coordinator.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
}

func seedWorker(t *testing.T, home string) {
	t.Helper()
	seedAgent(t, home, records.Agent{Key: "worker-one", Name: "worker-one", Profile: records.ProfileBuilder, PaneID: "w2:p1", WorkspaceID: "w2", Status: records.AgentLive, Generation: 1})
}

type promptRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type fakeHerdr struct {
	t             *testing.T
	listener      net.Listener
	path          string
	mu            sync.Mutex
	requests      []promptRequest
	promptErr     bool
	paneWorkspace string
	before        func(method string, params json.RawMessage)
}

func newFakeHerdr(t *testing.T) *fakeHerdr {
	t.Helper()
	path := filepath.Join(t.TempDir(), "herdr.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeHerdr{t: t, listener: listener, path: path, paneWorkspace: "w1"}
	t.Setenv("HERDR_SOCKET_PATH", path)
	go fake.serve()
	t.Cleanup(func() { _ = listener.Close(); _ = os.Remove(path) })
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
	requestLine, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return
	}
	var request promptRequest
	if err := json.Unmarshal(requestLine, &request); err != nil {
		f.t.Errorf("decode request: %v", err)
		return
	}
	f.mu.Lock()
	f.requests = append(f.requests, request)
	promptErr := f.promptErr
	before := f.before
	f.mu.Unlock()
	if before != nil {
		before(request.Method, request.Params)
	}
	var result, serverError string
	switch request.Method {
	case "ping":
		result = `{"type":"pong","version":"0.9.5","protocol":22}`
	case "pane.get":
		var params struct {
			PaneID string `json:"pane_id"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			f.t.Errorf("decode pane.get params: %v", err)
			return
		}
		f.mu.Lock()
		workspaceID := f.paneWorkspace
		f.mu.Unlock()
		result = fmt.Sprintf(`{"type":"pane_info","pane":{"pane_id":%q,"terminal_id":"term_tui","workspace_id":%q,"tab_id":"%s:t1","focused":false,"agent_status":"idle","revision":1}}`, params.PaneID, workspaceID, workspaceID)
	case "agent.prompt":
		if promptErr {
			serverError = `,"error":{"code":"agent_prompt_stalled","message":"stalled"}`
		} else {
			result = `{"type":"agent_prompted","agent":{"terminal_id":"term_a","agent_status":"idle","workspace_id":"w1","tab_id":"w1:t1","pane_id":"w1:p1","focused":false,"revision":1}}`
		}
	default:
		f.t.Errorf("unexpected herdr method %q", request.Method)
		return
	}
	body := fmt.Sprintf(`{"id":%q`, request.ID)
	if serverError != "" {
		body += serverError + "}\n"
	} else {
		body += `,"result":` + result + "}\n"
	}
	_, _ = conn.Write([]byte(body))
}

func (f *fakeHerdr) recorded(method string) []promptRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	var found []promptRequest
	for _, request := range f.requests {
		if request.Method == method {
			found = append(found, request)
		}
	}
	return found
}

func (f *fakeHerdr) failPrompt() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.promptErr = true
}

func (f *fakeHerdr) setPaneWorkspace(workspaceID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.paneWorkspace = workspaceID
}

func (f *fakeHerdr) setBefore(fn func(method string, params json.RawMessage)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.before = fn
}

func chatFiles(t *testing.T, home string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(home, ".hive", "state", "chat", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestSendRequiresLiveCoordinatorBeforeChatWrite(t *testing.T) {
	home := chatHome(t)
	envelope := Send(context.Background(), home, SendOptions{Text: "hello"})
	if envelope.Outcome != "failed" || !strings.Contains(fmt.Sprint(envelope.Error), "coordinator") {
		t.Fatalf("Send() = %#v; want no-live-coordinator failure", envelope)
	}
	if files := chatFiles(t, home); len(files) != 0 {
		t.Fatalf("chat files = %v; want none", files)
	}
}

func TestSendRefusesRegisteredAgentBeforeChatWrite(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedWorker(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{Agent: "worker-one", Text: "hello"})
	message := fmt.Sprint(envelope.Error)
	if envelope.Outcome != "failed" || !strings.HasPrefix(message, "chat send is for the user only") || !strings.Contains(message, "rule 1") {
		t.Fatalf("Send() = %#v; want a rule 1 refusal", envelope)
	}
	if files := chatFiles(t, home); len(files) != 0 {
		t.Fatalf("chat files = %v; want none", files)
	}
	if got := len(fake.recorded("agent.prompt")); got != 0 {
		t.Fatalf("agent.prompt calls = %d; want none", got)
	}
}

func TestSendRecordsMultilineUserPromptAndDelivery(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{Text: "line one\nline two"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Send() = %#v; want ok", envelope)
	}
	result, ok := envelope.Result.(SendResult)
	if !ok || result.ID != "c1" || result.Delivery != "sent" {
		t.Fatalf("result = %#v; want c1 sent", envelope.Result)
	}
	prompts := fake.recorded("agent.prompt")
	if len(prompts) != 1 {
		t.Fatalf("agent.prompt calls = %d; want one", len(prompts))
	}
	var params herdr.AgentPromptParams
	if err := json.Unmarshal(prompts[0].Params, &params); err != nil {
		t.Fatal(err)
	}
	if params.Target != "w1:p1" || params.Text != "⬢ user: line one\nline two" {
		t.Fatalf("prompt = %#v; want one marked multiline submission to coordinator pane", params)
	}
	files := chatFiles(t, home)
	if len(files) != 1 {
		t.Fatalf("chat files = %v; want one daily file", files)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("chat lines = %d; want pending entry and delivery update", len(lines))
	}
	var message, delivery map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &message); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &delivery); err != nil {
		t.Fatal(err)
	}
	if message["id"] != "c1" || message["from"] != "user" || message["delivery"] != "pending" || delivery["id"] != "c1" || delivery["delivery"] != "sent" {
		t.Fatalf("chat lines = %#v / %#v; want c1 user pending then c1 sent", message, delivery)
	}
	at, err := time.Parse(time.RFC3339Nano, message["at"].(string))
	if err != nil || filepath.Base(files[0]) != at.UTC().Format("2006-01-02")+".jsonl" {
		t.Fatalf("daily file = %s, at = %v; want UTC date", filepath.Base(files[0]), message["at"])
	}
	logEnvelope := Log(home, LogOptions{})
	logResult, ok := logEnvelope.Result.(LogResult)
	if logEnvelope.Outcome != "ok" || !ok || len(logResult.Entries) != 1 || logResult.Entries[0].Delivery != "sent" {
		t.Fatalf("Log() = %#v; want user entry with sent delivery", logEnvelope)
	}
}

func TestSendMarksPromptFailureUnknownWithoutRetry(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	fake := newFakeHerdr(t)
	fake.failPrompt()
	envelope := Send(context.Background(), home, SendOptions{Text: "hello"})
	result, ok := envelope.Result.(SendResult)
	if envelope.Outcome != "failed" || !ok || result.ID != "c1" || result.Delivery != "unknown" {
		t.Fatalf("Send() = %#v; want c1 unknown failure", envelope)
	}
	if got := len(fake.recorded("agent.prompt")); got != 1 {
		t.Fatalf("agent.prompt calls = %d; want one and no retry", got)
	}
}

func TestReplyRequiresCoordinatorAndSetsReplyTo(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedWorker(t, home)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Next("chat"); err != nil {
		t.Fatal(err)
	}
	appendChatLine(t, home, "2026-01-02", Entry{ID: "c1", At: "2026-01-02T03:04:05Z", From: "user", Text: "hello", Delivery: "pending"})
	appendChatLine(t, home, "2026-01-02", Entry{ID: "c1", At: "2026-01-02T03:04:06Z", Delivery: "sent"})
	denied := Reply(home, ReplyOptions{Agent: "worker-one", Text: "answer"})
	if denied.Outcome != "failed" || !strings.Contains(fmt.Sprint(denied.Error), "coordinator") {
		t.Fatalf("Reply() as worker = %#v; want coordinator-role error", denied)
	}
	allowed := Reply(home, ReplyOptions{Agent: "coordinator", To: "c1", Text: "answer"})
	if allowed.Outcome != "ok" {
		t.Fatalf("Reply() as coordinator = %#v; want ok", allowed)
	}
	result, ok := allowed.Result.(ReplyResult)
	if !ok || result.Entry.ID != "c2" || result.Entry.From != "coordinator" || result.Entry.ReplyTo != "c1" {
		t.Fatalf("reply result = %#v; want coordinator c2 reply_to c1", allowed.Result)
	}
}

func TestLogSkipsBadLinesAppliesSinceAndLimit(t *testing.T) {
	home := chatHome(t)
	appendChatLine(t, home, "2026-01-01", Entry{ID: "c1", At: "2026-01-01T01:00:00Z", From: "user", Text: "old", Delivery: "pending"})
	appendChatLine(t, home, "2026-01-01", Entry{ID: "c1", At: "2026-01-01T01:00:01Z", Delivery: "sent"})
	appendChatLine(t, home, "2026-01-02", Entry{ID: "c2", At: "2026-01-02T01:00:00Z", From: "user", Text: "first", Delivery: "pending"})
	appendChatLine(t, home, "2026-01-02", Entry{ID: "c2", At: "2026-01-02T01:01:00Z", Delivery: "sent"})
	appendChatLine(t, home, "2026-01-02", Entry{ID: "c3", At: "2026-01-02T01:02:00Z", From: "coordinator", Text: "last", ReplyTo: "c2"})
	appendChatRaw(t, home, "2026-01-02", "not-json")
	envelope := Log(home, LogOptions{Since: "2026-01-02", Limit: 1})
	result, ok := envelope.Result.(LogResult)
	if envelope.Outcome != "ok" || !ok || len(result.Entries) != 1 || result.BadLines != 1 {
		t.Fatalf("Log() = %#v; want one latest entry and one bad line", envelope)
	}
	entry := result.Entries[0]
	if entry.ID != "c3" || entry.ReplyTo != "c2" {
		t.Fatalf("entry = %#v; want coordinator reply to c2", entry)
	}
}

func appendChatLine(t *testing.T, home, date string, entry Entry) {
	t.Helper()
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	appendChatRaw(t, home, date, string(data))
}

func appendChatRaw(t *testing.T, home, date, line string) {
	t.Helper()
	dir := filepath.Join(home, ".hive", "state", "chat")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(dir, date+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(line + "\n"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSendRefusesLiveAgentPane(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedWorker(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w2:p1", Text: "hello"})
	assertSendRefused(t, home, fake, envelope, "rule 1")
}

func TestSendRefusesLiveAgentPaneWithoutCoordinator(t *testing.T) {
	home := chatHome(t)
	seedWorker(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w2:p1", Text: "hello"})
	assertSendRefused(t, home, fake, envelope, "rule 1")
}

func TestSendRefusesCallerInLiveAgentWorkspace(t *testing.T) {
	for _, role := range []string{"manager", "worker", "verifier"} {
		t.Run(role, func(t *testing.T) {
			home := chatHome(t)
			seedCoordinator(t, home)
			agent := records.Agent{
				Key: "agent-one", Name: "agent-one", Project: "demo", Task: "task-one",
				WorkspaceID: "w-agent", Status: records.AgentLive, Generation: 1,
			}
			switch role {
			case "manager":
				agent.Profile = records.ProfilePlanner
				agent.Owns = "demo"
			case "worker":
				agent.Profile = records.ProfileBuilder
			case "verifier":
				agent.Profile = records.ProfileVerifier
			}
			seedAgent(t, home, agent)
			fake := newFakeHerdr(t)
			fake.setPaneWorkspace("w-agent")
			envelope := Send(context.Background(), home, SendOptions{PaneID: "w-user:p1", Text: "hello"})
			assertSendRefused(t, home, fake, envelope, "rule 3")
		})
	}
}

func TestSendAllowsUnregisteredPaneInCoordinatorWorkspace(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w1:p2", Text: "hello"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Send() = %#v; want ok for an unregistered user pane", envelope)
	}
	if got := len(fake.recorded("pane.get")); got != 1 {
		t.Fatalf("pane.get calls = %d; want one workspace check", got)
	}
	if got := len(fake.recorded("agent.prompt")); got != 1 {
		t.Fatalf("agent.prompt calls = %d; want one", got)
	}
}

func TestSendAllowsUserPaneInCoordinatorWorkspaceWithManager(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedManager(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w1:p3", Text: "hello"})
	if envelope.Outcome != "ok" {
		t.Fatalf("Send() = %#v; want ok from the coordinator workspace", envelope)
	}
	if got := len(fake.recorded("agent.prompt")); got != 1 {
		t.Fatalf("agent.prompt calls = %d; want one", got)
	}
}

func TestSendRefusesManagerPaneInCoordinatorWorkspaceByRule1(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedManager(t, home)
	fake := newFakeHerdr(t)
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w1:p2", Text: "hello"})
	assertSendDenied(t, home, fake, envelope, "chat send is for the user only: only the user can send chat; rule 1 (caller pane belongs to the live agent manager-hive)")
}

func TestSendRefusesUserPaneInLiveWorkerWorkspaceByRule3(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	seedWorker(t, home)
	fake := newFakeHerdr(t)
	fake.setPaneWorkspace("w2")
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w2:p2", Text: "hello"})
	assertSendDenied(t, home, fake, envelope, "chat send is for the user only: only the user can send chat; rule 3 (caller pane shares the workspace of the live agent worker-one)")
}

func TestValidateCallerRefusesWorkerWorkspaceWithoutCoordinatorByRule3(t *testing.T) {
	home := chatHome(t)
	seedWorker(t, home)
	fake := newFakeHerdr(t)
	fake.setPaneWorkspace("w2")
	err := ValidateCaller(context.Background(), home, "", "w2:p2", "")
	if err == nil || err.Error() != "chat send is for the user only: only the user can send chat; rule 3 (caller pane shares the workspace of the live agent worker-one)" {
		t.Fatalf("ValidateCaller() error = %v; want rule 3 refusal", err)
	}
}

func TestSendRefusesCoordinatorWorkspaceCallerBelowProjectsByRule2(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	fake := newFakeHerdr(t)
	workingDir := filepath.Join(home, "projects", "demo")
	if err := os.MkdirAll(workingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := Send(context.Background(), home, SendOptions{PaneID: "w1:p2", WorkingDir: workingDir, Text: "hello"})
	assertSendDenied(t, home, fake, envelope, "chat send is for the user only: only the user can send chat; rule 2 (working directory is below projects/)")
}

func seedManager(t *testing.T, home string) {
	t.Helper()
	manager := records.Agent{Key: "manager-hive-g1", Name: "manager-hive", Profile: records.ProfilePlanner, Owns: "hive", Project: "hive", PaneID: "w1:p2", WorkspaceID: "w1", Status: records.AgentLive, Generation: 1}
	seedAgent(t, home, manager)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"claims"}, "owner-hive", records.Claim{Key: "owner-hive", Agent: manager.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
}

func assertSendDenied(t *testing.T, home string, fake *fakeHerdr, envelope output.Envelope, want string) {
	t.Helper()
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != want {
		t.Fatalf("Send() = %#v; want error %q", envelope, want)
	}
	if files := chatFiles(t, home); len(files) != 0 {
		t.Fatalf("chat files = %v; want none", files)
	}
	if got := len(fake.recorded("agent.prompt")); got != 0 {
		t.Fatalf("agent.prompt calls = %d; want none", got)
	}
}

func assertSendRefused(t *testing.T, home string, fake *fakeHerdr, envelope output.Envelope, rule string) {
	t.Helper()
	message := fmt.Sprint(envelope.Error)
	if envelope.Outcome != "failed" || !strings.HasPrefix(message, "chat send is for the user only") || !strings.Contains(message, rule) {
		t.Fatalf("Send() = %#v; want %s refusal", envelope, rule)
	}
	if files := chatFiles(t, home); len(files) != 0 {
		t.Fatalf("chat files = %v; want none", files)
	}
	if got := len(fake.recorded("agent.prompt")); got != 0 {
		t.Fatalf("agent.prompt calls = %d; want none", got)
	}
}

// probeStoreWrite writes one unrelated record while a remote call waits.
func probeStoreWrite(registry *store.Store, mutate func(locked *store.LockedStore) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	return registry.WithLock(ctx, mutate)
}

// TestValidateCallerKeepsStoreUnlockedDuringPaneGet proves that the caller
// pane read does not hold the global store lock.
func TestValidateCallerKeepsStoreUnlockedDuringPaneGet(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	fake := newFakeHerdr(t)
	var probeErr error
	fake.setBefore(func(method string, params json.RawMessage) {
		if method != "pane.get" {
			return
		}
		probeErr = probeStoreWrite(registry, func(locked *store.LockedStore) error {
			_, err := locked.Next("probe")
			return err
		})
	})
	if err := ValidateCaller(context.Background(), home, "", "w1:p2", ""); err != nil {
		t.Fatalf("ValidateCaller() error = %v; want the user pane allowed", err)
	}
	if probeErr != nil {
		t.Fatalf("store write during pane.get = %v; want no global lock", probeErr)
	}
}

// TestValidateCallerRefusesChangedIdentityAfterPaneGet proves that the final
// locked check reads the caller identity and claim facts again.
func TestValidateCallerRefusesChangedIdentityAfterPaneGet(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	fake := newFakeHerdr(t)
	fake.setPaneWorkspace("wUser")
	var mutationErr error
	fake.setBefore(func(method string, params json.RawMessage) {
		if method != "pane.get" {
			return
		}
		mutationErr = probeStoreWrite(registry, func(locked *store.LockedStore) error {
			record := records.Agent{
				Key: "manager-demo-g1", Name: "manager-demo", Profile: records.ProfilePlanner, Owns: "demo", Project: "demo",
				PaneID: "wUser:p1", WorkspaceID: "wUser", Status: records.AgentLive, Generation: 1,
			}
			if err := locked.WriteIn([]string{"agents"}, record.Key, record); err != nil {
				return err
			}
			return locked.WriteIn([]string{"claims"}, "owner-demo", records.Claim{Key: "owner-demo", Agent: record.Key, Generation: 1})
		})
	})
	err = ValidateCaller(context.Background(), home, "", "wUser:p1", "")
	if mutationErr != nil {
		t.Fatalf("caller identity change = %v; want the store write to complete", mutationErr)
	}
	want := "chat send is for the user only: only the user can send chat; rule 1 (caller pane belongs to the live agent manager-demo)"
	if err == nil || err.Error() != want {
		t.Fatalf("ValidateCaller() error = %v; want the fresh rule 1 refusal", err)
	}
}

// TestValidateCallerRefusesRemovedCoordinatorClaimAfterPaneGet proves that the
// final locked check reads the coordinator claim again.
func TestValidateCallerRefusesRemovedCoordinatorClaimAfterPaneGet(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	fake := newFakeHerdr(t)
	var mutationErr error
	fake.setBefore(func(method string, params json.RawMessage) {
		if method != "pane.get" {
			return
		}
		mutationErr = probeStoreWrite(registry, func(locked *store.LockedStore) error {
			if err := locked.RemoveIn([]string{"claims"}, "coordinator"); err != nil {
				return err
			}
			record := records.Agent{
				Key: "worker-late", Name: "worker-late", Profile: records.ProfileBuilder, Project: "demo",
				PaneID: "w1:p9", WorkspaceID: "w1", Status: records.AgentLive, Generation: 1,
			}
			return locked.WriteIn([]string{"agents"}, record.Key, record)
		})
	})
	err = ValidateCaller(context.Background(), home, "", "w1:p2", "")
	if mutationErr != nil {
		t.Fatalf("caller claim change = %v; want the store write to complete", mutationErr)
	}
	want := "chat send is for the user only: only the user can send chat; rule 3 (caller pane shares the workspace of the live agent worker-late)"
	if err == nil || err.Error() != want {
		t.Fatalf("ValidateCaller() error = %v; want the fresh rule 3 refusal", err)
	}
}

// TestSendRefusesChangedCallerWorkspaceBeforeAppend proves that the final
// locked check reads fresh facts before the chat append.
func TestSendRefusesChangedCallerWorkspaceBeforeAppend(t *testing.T) {
	home := chatHome(t)
	seedCoordinator(t, home)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	fake := newFakeHerdr(t)
	fake.setPaneWorkspace("wUser")
	var mutationErr error
	fake.setBefore(func(method string, params json.RawMessage) {
		if method != "pane.get" {
			return
		}
		mutationErr = probeStoreWrite(registry, func(locked *store.LockedStore) error {
			record := records.Agent{
				Key: "worker-late", Name: "worker-late", Profile: records.ProfileBuilder, Project: "demo",
				PaneID: "wUser:p9", WorkspaceID: "wUser", Status: records.AgentLive, Generation: 1,
			}
			return locked.WriteIn([]string{"agents"}, record.Key, record)
		})
	})
	envelope := Send(context.Background(), home, SendOptions{PaneID: "wUser:p1", Text: "hello"})
	if mutationErr != nil {
		t.Fatalf("caller workspace change = %v; want the store write to complete", mutationErr)
	}
	want := "chat send is for the user only: only the user can send chat; rule 3 (caller pane shares the workspace of the live agent worker-late)"
	if envelope.Outcome != "failed" || fmt.Sprint(envelope.Error) != want {
		t.Fatalf("Send() = %#v; want the fresh rule 3 refusal", envelope)
	}
	if files := chatFiles(t, home); len(files) != 0 {
		t.Fatalf("chat files = %v; want none", files)
	}
	if got := len(fake.recorded("agent.prompt")); got != 0 {
		t.Fatalf("agent.prompt calls = %d; want none", got)
	}
}
