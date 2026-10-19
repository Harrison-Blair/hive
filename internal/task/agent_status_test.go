package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func seedBrief(t *testing.T, home, id string) {
	t.Helper()
	path := filepath.Join(home, ".hive", "state", "briefs", id+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("Seed"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestShowAgentStatus(t *testing.T) {
	for _, status := range []string{records.AgentLive, records.AgentUnresolved, records.AgentExited, records.AgentStopped} {
		t.Run(status, func(t *testing.T) {
			home, task := newSeed(t, "assigned")
			seedAgent(t, home, records.Agent{
				Key: "worker-1", Name: "worker-1", Role: "worker", Project: "one", Task: task.ID,
				Status: status, PaneID: "w1:p1", Generation: 1,
				StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
			})
			seedBrief(t, home, task.ID)
			envelope := Show(home, task.ID, ShowOptions{})
			if envelope.Outcome != "ok" {
				t.Fatalf("Show() = %+v, want ok", envelope)
			}
			if text := ShowText(envelope); !strings.Contains(text, "acting_agent_status: "+status) {
				t.Fatalf("ShowText() = %q, want the %s status", text, status)
			}
		})
	}

	t.Run("missing record", func(t *testing.T) {
		home, task := newSeed(t, "assigned")
		task.ActingAgent = "missing-agent"
		seedTask(t, home, task)
		seedBrief(t, home, task.ID)
		envelope := Show(home, task.ID, ShowOptions{})
		if envelope.Outcome != "ok" {
			t.Fatalf("Show() = %+v, want ok", envelope)
		}
		if text := ShowText(envelope); !strings.Contains(text, "acting_agent_status: no record") {
			t.Fatalf("ShowText() = %q, want no record", text)
		}
	})
}
