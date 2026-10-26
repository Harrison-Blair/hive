package agent

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestShowPrefersStartingGenerationByName(t *testing.T) {
	home := testHome(t)
	writeRecord(t, home, records.Agent{Key: "worker-g1", Name: "worker", Profile: records.ProfileBuilder, Status: records.AgentLive, Generation: 1})
	writeRecord(t, home, records.Agent{Key: "worker-g2", Name: "worker", Profile: records.ProfileBuilder, Status: records.AgentStarting, Generation: 2})

	envelope := Show(home, "worker")
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want the starting record", envelope)
	}
	record, ok := envelope.Result.(records.Agent)
	if !ok || record.Key != "worker-g2" || record.Status != records.AgentStarting {
		t.Fatalf("Show() = %#v; want worker-g2 starting", envelope.Result)
	}
	if text := ShowText(envelope); !strings.Contains(text, "status: starting") {
		t.Fatalf("ShowText() = %q; want starting status", text)
	} else if strings.Contains(text, "parent:") || strings.Contains(text, "replaces:") || strings.Contains(text, "status since:") {
		t.Fatalf("ShowText() = %q; want no empty lineage fields", text)
	}
}

func TestShowAndListExposeLineageFields(t *testing.T) {
	home := testHome(t)
	record := records.Agent{Key: "worker-g2", Name: "worker", Profile: records.ProfileBuilder, Status: records.AgentLive, Generation: 2}
	writeRecord(t, home, record)
	setAgentJSONFields(t, home, record.Key, map[string]string{
		"parent": "manager-g1", "replaces": "worker-g1", "status_changed_at": "2026-01-02T03:04:05Z",
	})

	envelope := Show(home, record.Key)
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want success", envelope)
	}
	text := ShowText(envelope)
	for _, want := range []string{"parent: manager-g1", "replaces: worker-g1", "status since: 2026-01-02T03:04:05Z"} {
		if !strings.Contains(text, want) {
			t.Errorf("ShowText() = %q; missing %q", text, want)
		}
	}

	listed := List(home, ListOptions{All: true})
	if listed.Outcome != "ok" {
		t.Fatalf("List() = %#v; want success", listed)
	}
	data, err := json.Marshal(listed.Result)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"parent":"manager-g1"`, `"replaces":"worker-g1"`, `"status_changed_at":"2026-01-02T03:04:05Z"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("List JSON = %s; missing %s", data, want)
		}
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

func TestShowResolvesClaimBeforeWorkerRecordKey(t *testing.T) {
	home := t.TempDir()
	if envelope := Register(home, RegisterOptions{Kind: records.AgentKindCoordinator, Name: "leader", NameSet: true, PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	worker := records.Agent{Key: "coordinator", Name: "coordinator", Profile: records.ProfileBuilder, Status: records.AgentLive, Generation: 1, PaneID: "w1:p2"}
	if err := st.WriteIn([]string{"agents"}, "coordinator", worker); err != nil {
		t.Fatal(err)
	}
	envelope := Show(home, "coordinator")
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want the claim holder", envelope)
	}
	record, ok := envelope.Result.(records.Agent)
	if !ok || record.Key != "leader-g1" {
		t.Fatalf("Show() = %#v; want leader-g1 rather than the worker alias", envelope.Result)
	}
}

func TestShowDoesNotResolveUnclaimedSingletonOrphan(t *testing.T) {
	for _, name := range []string{"coordinator", "leader"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			seam := &seamWriter{writeCollection: "claims", writeErr: errors.New("seam claim failure")}
			options := RegisterOptions{Kind: records.AgentKindCoordinator, Name: name, NameSet: true, PaneID: "orphan-pane"}
			envelope := runRegisterLocked(t, home, seam, options)
			if envelope.Outcome != "partial" || envelope.Phase != "claim" {
				t.Fatalf("runRegisterLocked() = %#v; want partial claim", envelope)
			}
			if got := Show(home, name); got.Outcome == "ok" {
				t.Fatalf("Show(%q) = %#v; want no role holder without a claim", name, got)
			}
			key := name + "-g1"
			if got := Show(home, key); got.Outcome != "ok" {
				t.Fatalf("Show(%q) = %#v; want direct generation-record inspection", key, got)
			}
		})
	}
}

func TestShowLogsReadFailure(t *testing.T) {
	home := t.TempDir()
	if envelope := Show(home, "unknown"); envelope.Outcome != "failed" {
		t.Fatalf("Show() = %#v; want failed", envelope)
	}
	events := readAgentLogEvents(t, home)
	event := findAgentLogEvent(events, "agent show failed")
	if event == nil {
		t.Fatalf("events = %#v; want an agent show failure event", events)
	}
	message, _ := event["error"].(string)
	if event["name"] != "unknown" || event["phase"] != "read" || !strings.Contains(message, "agent not found: unknown") {
		t.Fatalf("failure event = %#v; want name unknown and the read error", event)
	}
	if strings.ContainsAny(message, "\n\r\x1b") {
		t.Fatalf("error = %q; want one physical line", message)
	}
	if got := Show(home, "a\nb"); got.Outcome != "failed" {
		t.Fatalf("Show() = %#v; want failed for a control name", got)
	}
	events = readAgentLogEvents(t, home)
	for _, candidate := range events {
		if candidate["msg"] != "agent show failed" {
			continue
		}
		value, _ := candidate["error"].(string)
		if strings.ContainsAny(value, "\n\r\x1b") {
			t.Fatalf("error event = %#v; want one physical line", candidate)
		}
	}
}

func TestListLogsReadFailure(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".hive", "state", "agents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	envelope := List(home, ListOptions{Project: "hive", Profile: records.ProfileBuilder})
	if envelope.Outcome != "failed" {
		t.Fatalf("List() = %#v; want failed", envelope)
	}
	events := readAgentLogEvents(t, home)
	event := findAgentLogEvent(events, "agent list failed")
	if event == nil {
		t.Fatalf("events = %#v; want an agent list failure event", events)
	}
	message, _ := event["error"].(string)
	if event["project_id"] != "hive" || event["profile"] != "builder" || event["phase"] != "store" || message == "" {
		t.Fatalf("failure event = %#v; want the list filters, phase, and error", event)
	}
}

func TestShowResolvesCustomSingletonNameThroughItsClaim(t *testing.T) {
	home := t.TempDir()
	if envelope := Register(home, RegisterOptions{Kind: records.AgentKindCoordinator, Name: "leader", NameSet: true, PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	envelope := Show(home, "leader")
	if envelope.Outcome != "ok" {
		t.Fatalf("Show() = %#v; want the custom singleton holder", envelope)
	}
	record, ok := envelope.Result.(records.Agent)
	if !ok || record.Key != "leader-g1" {
		t.Fatalf("Show() = %#v; want leader-g1", envelope.Result)
	}
}
