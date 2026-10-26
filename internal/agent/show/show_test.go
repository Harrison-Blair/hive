package show

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestTextShowsStartingGenerationByName(t *testing.T) {
	home := agenttest.NewHome(t)
	agenttest.WriteRecord(t, home, records.Agent{Key: "worker-g1", Name: "worker", Profile: records.ProfileBuilder, Status: records.AgentLive, Generation: 1})
	agenttest.WriteRecord(t, home, records.Agent{Key: "worker-g2", Name: "worker", Profile: records.ProfileBuilder, Status: records.AgentStarting, Generation: 2})

	envelope := agentlib.Show(home, "worker")
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want the starting record", envelope)
	}
	if text := Text(envelope); !strings.Contains(text, "status: starting") {
		t.Fatalf("Text() = %q; want starting status", text)
	} else if strings.Contains(text, "parent:") || strings.Contains(text, "replaces:") || strings.Contains(text, "status since:") {
		t.Fatalf("Text() = %q; want no empty lineage fields", text)
	}
}

func TestTextShowsLineageFields(t *testing.T) {
	home := agenttest.NewHome(t)
	record := records.Agent{Key: "worker-g2", Name: "worker", Profile: records.ProfileBuilder, Status: records.AgentLive, Generation: 2}
	agenttest.WriteRecord(t, home, record)
	setAgentJSONFields(t, home, record.Key, map[string]string{
		"parent": "manager-g1", "replaces": "worker-g1", "status_changed_at": "2026-01-02T03:04:05Z",
	})

	envelope := agentlib.Show(home, record.Key)
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want success", envelope)
	}
	text := Text(envelope)
	for _, want := range []string{"parent: manager-g1", "replaces: worker-g1", "status since: 2026-01-02T03:04:05Z"} {
		if !strings.Contains(text, want) {
			t.Errorf("Text() = %q; missing %q", text, want)
		}
	}
}

func TestTextShowsAgentFacts(t *testing.T) {
	home := agenttest.NewHome(t)
	agenttest.WriteProjectPath(t, home, "hive", "/tmp/hive")
	if envelope := agentlib.Register(home, agentlib.RegisterOptions{Profile: records.ProfilePlanner, Owns: "hive", Project: "hive", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	shown := agentlib.Show(home, "manager-hive-g1")
	if shown.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want ok", shown)
	}
	text := Text(shown)
	for _, want := range []string{"profile: planner", "kind: ", "owns: hive"} {
		if !strings.Contains(text, want) {
			t.Errorf("Text() = %q; missing %q", text, want)
		}
	}
	if strings.Contains(text, "role:") {
		t.Errorf("Text() = %q; want no role field", text)
	}
}

func setAgentJSONFields(t *testing.T, home, key string, fields map[string]string) {
	t.Helper()
	path := filepath.Join(home, ".hive", "state", "agents", key+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	for field, value := range fields {
		record[field] = value
	}
	data, err = json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
