package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// stateSnapshot returns the bytes of each agent record and each claim.
func stateSnapshot(t *testing.T, home string) map[string]string {
	t.Helper()
	snapshot := map[string]string{}
	for _, collection := range []string{"agents", "claims"} {
		dir := filepath.Join(home, ".hive", "state", collection)
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			snapshot[collection+"/"+entry.Name()] = string(data)
		}
	}
	return snapshot
}

// seedWorkerTask writes the project and the task of a worker registration.
func seedWorkerTask(t *testing.T, home string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"projects"}, "demo", records.Project{ID: "demo", Path: home}); err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"tasks"}, "1-demo", records.Task{ID: "1-demo", Project: "demo", Title: "demo", Status: "assigned"}); err != nil {
		t.Fatal(err)
	}
}

func seedRefusalState(t *testing.T, home string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.WriteIn([]string{"projects"}, "demo", records.Project{ID: "demo", Path: home}); err != nil {
		t.Fatal(err)
	}
	coordinator := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w1:coord", Status: records.AgentLive, Generation: 1, StartedByHive: true}
	worker := records.Agent{Key: "worker-demo-1", Name: "worker-demo-1", Role: "worker", Project: "demo", Task: "1-demo", PaneID: "w1:general", Status: records.AgentLive, Generation: 1}
	for _, record := range []records.Agent{coordinator, worker} {
		if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
			t.Fatal(err)
		}
	}
	claim := records.Claim{Key: "coordinator", Agent: "coordinator-g1", Generation: 1, ClaimedAt: "2026-10-21T00:00:00Z"}
	if err := registry.WriteIn([]string{"claims"}, "coordinator", claim); err != nil {
		t.Fatal(err)
	}
}

// TestAgentCLIRefusesHiveOwnedRoles covers decisions 30 and 37: the CLI
// refuses each role that hive code registers, from each caller pane, and the
// refusal changes no claim and no record.
func TestAgentCLIRefusesHiveOwnedRoles(t *testing.T) {
	attempts := []struct {
		name string
		args []string
	}{
		{"register coordinator", []string{"agent", "register", "--role", "coordinator"}},
		{"register supervisor", []string{"agent", "register", "--role", "supervisor"}},
		{"register manager", []string{"agent", "register", "--role", "manager", "--project", "demo"}},
		{"spawn coordinator", []string{"agent", "spawn", "--role", "coordinator", "--harness", "claude", "--pane", "w1:empty"}},
		{"spawn supervisor", []string{"agent", "spawn", "--role", "supervisor", "--harness", "claude", "--pane", "w1:empty"}},
	}
	callers := []struct {
		name  string
		pane  string
		extra []string
	}{
		{"empty pane", "w1:empty", nil},
		{"registered agent pane", "w1:general", nil},
		{"coordinator pane", "w1:coord", nil},
		{"pane override", "w1:general", []string{"--pane", "w1:other"}},
	}
	for _, attempt := range attempts {
		for _, caller := range callers {
			t.Run(attempt.name+"/"+caller.name, func(t *testing.T) {
				home := testHome(t)
				seedRefusalState(t, home)
				t.Setenv("HERDR_ENV", "1")
				t.Setenv("HERDR_PANE_ID", caller.pane)
				t.Setenv("HERDR_SOCKET_PATH", filepath.Join(t.TempDir(), "absent.sock"))
				before := stateSnapshot(t, home)
				args := append([]string{}, attempt.args...)
				args = append(args, caller.extra...)
				var stdout, stderr bytes.Buffer
				status := executeWithStreams(args, strings.NewReader(""), &stdout, &stderr)
				if status == 0 {
					t.Fatalf("status = 0, stdout = %q; want a refusal", stdout.String())
				}
				message := stderr.String()
				if !strings.Contains(message, "hive code registers") {
					t.Fatalf("stderr = %q; want the hive-owned role refusal", message)
				}
				if !strings.Contains(attempt.name, "manager") && (!strings.Contains(message, "bare hive") || !strings.Contains(message, "Health › Agents")) {
					t.Fatalf("stderr = %q; want the advice to run bare hive or the restart in Health › Agents", message)
				}
				after := stateSnapshot(t, home)
				if len(after) != len(before) {
					t.Fatalf("state files = %d after, %d before; want no change", len(after), len(before))
				}
				for path, data := range before {
					if after[path] != data {
						t.Fatalf("%s changed:\nbefore %s\nafter  %s", path, data, after[path])
					}
				}
			})
		}
	}
}
