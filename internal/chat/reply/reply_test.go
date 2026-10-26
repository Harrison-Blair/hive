package reply

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	chatlib "github.com/Harrison-Blair/hive/internal/chat"
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

func appendChatLine(t *testing.T, home, date string, entry chatlib.Entry) {
	t.Helper()
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".hive", "state", "chat")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(dir, date+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(string(data) + "\n"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
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
	appendChatLine(t, home, "2026-01-02", chatlib.Entry{ID: "c1", At: "2026-01-02T03:04:05Z", From: "user", Text: "hello", Delivery: "pending"})
	appendChatLine(t, home, "2026-01-02", chatlib.Entry{ID: "c1", At: "2026-01-02T03:04:06Z", Delivery: "sent"})
	denied := Reply(home, Options{Agent: "worker-one", Text: "answer"})
	if denied.Outcome != "failed" || !strings.Contains(fmt.Sprint(denied.Error), "coordinator") {
		t.Fatalf("Reply() as worker = %#v; want coordinator-role error", denied)
	}
	allowed := Reply(home, Options{Agent: "coordinator", To: "c1", Text: "answer"})
	if allowed.Outcome != "ok" {
		t.Fatalf("Reply() as coordinator = %#v; want ok", allowed)
	}
	result, ok := allowed.Result.(Result)
	if !ok || result.Entry.ID != "c2" || result.Entry.From != "coordinator" || result.Entry.ReplyTo != "c1" {
		t.Fatalf("reply result = %#v; want coordinator c2 reply_to c1", allowed.Result)
	}
}
