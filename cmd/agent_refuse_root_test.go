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
	coordinator := records.Agent{Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator, PaneID: "w1:coord", Status: records.AgentLive, Generation: 1, StartedByHive: true}
	worker := records.Agent{Key: "worker-demo-1", Name: "worker-demo-1", Profile: records.ProfileBuilder, Project: "demo", Task: "1-demo", PaneID: "w1:general", Status: records.AgentLive, Generation: 1}
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

// TestAgentCLIRefusesReservedNamesAndBuiltInProfile covers D1/D3: the CLI
// refuses the built-in names and the built-in coordinator profile, from each
// caller pane, and the refusal changes no claim and no record.
func TestAgentCLIRefusesReservedNamesAndBuiltInProfile(t *testing.T) {
	attempts := []struct {
		name string
		args []string
		want string
	}{
		{"register coordinator name", []string{"agent", "register", "--name", "coordinator"}, "name is reserved: coordinator"},
		{"register supervisor name", []string{"agent", "register", "--name", "supervisor"}, "name is reserved: supervisor"},
		{"spawn coordinator name", []string{"agent", "spawn", "--name", "coordinator", "--harness", "claude", "--pane", "w1:empty"}, "name is reserved: coordinator"},
		{"spawn supervisor name", []string{"agent", "spawn", "--name", "supervisor", "--harness", "claude", "--pane", "w1:empty"}, "name is reserved: supervisor"},
		{"spawn coordinator profile", []string{"agent", "spawn", "--profile", "coordinator", "--harness", "claude", "--pane", "w1:empty"}, "profile coordinator is built in"},
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
				if message := stderr.String(); !strings.Contains(message, attempt.want) {
					t.Fatalf("stderr = %q; want %q", message, attempt.want)
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
