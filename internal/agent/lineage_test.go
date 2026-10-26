package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestSpawnedRegistrationSelectsNewestMatchingReplacement(t *testing.T) {
	cases := []struct {
		name       string
		profile    string
		kind       string
		owns       string
		project    string
		task       string
		candidates []records.Agent
		want       string
	}{
		{
			name:    "worker same task",
			profile: records.ProfileBuilder, project: "hive", task: testTaskID,
			candidates: []records.Agent{
				{Key: "worker-old", Name: "worker-old", Profile: records.ProfileBuilder, Project: "hive", Task: testTaskID, Status: records.AgentStopped, StartedAt: "2026-01-01T00:00:00Z"},
				{Key: "worker-newer", Name: "worker-newer", Profile: records.ProfileBuilder, Project: "hive", Task: testTaskID, Status: records.AgentExited, StartedAt: "2026-01-02T00:00:00Z"},
				{Key: "worker-wrong-task", Name: "worker-wrong-task", Profile: records.ProfileBuilder, Project: "hive", Task: "2-task", Status: records.AgentStopped, StartedAt: "2026-01-03T00:00:00Z"},
				{Key: "worker-live", Name: "worker-live", Profile: records.ProfileBuilder, Project: "hive", Task: testTaskID, Status: records.AgentLive, StartedAt: "2026-01-04T00:00:00Z"},
			},
			want: "worker-newer",
		},
		{
			name:    "verifier same task",
			profile: records.ProfileVerifier, project: "hive", task: testTaskID,
			candidates: []records.Agent{
				{Key: "verifier-old", Name: "verifier-old", Profile: records.ProfileVerifier, Project: "hive", Task: testTaskID, Status: records.AgentStopped, StartedAt: "2026-01-01T00:00:00Z"},
				{Key: "verifier-other-task", Name: "verifier-other-task", Profile: records.ProfileVerifier, Project: "hive", Task: "2-task", Status: records.AgentStopped, StartedAt: "2026-01-02T00:00:00Z"},
			},
			want: "verifier-old",
		},
		{
			name:    "manager same project",
			profile: records.ProfilePlanner, owns: "hive", project: "hive",
			candidates: []records.Agent{
				{Key: "manager-old", Name: "manager-old", Profile: records.ProfilePlanner, Owns: "hive", Project: "hive", Status: records.AgentStale, StartedAt: "2026-01-01T00:00:00Z"},
				{Key: "manager-other-project", Name: "manager-other-project", Profile: records.ProfilePlanner, Owns: "other", Project: "other", Status: records.AgentStopped, StartedAt: "2026-01-02T00:00:00Z"},
			},
			want: "manager-old",
		},
		{
			name: "coordinator same role",
			kind: records.AgentKindCoordinator, project: "",
			candidates: []records.Agent{
				{Key: "coordinator-old", Name: "old-coordinator", Kind: records.AgentKindCoordinator, Status: records.AgentStopped, StartedAt: "2026-01-01T00:00:00Z"},
				{Key: "coordinator-newer", Name: "new-coordinator", Kind: records.AgentKindCoordinator, Project: "other", Status: records.AgentExited, StartedAt: "2026-01-02T00:00:00Z"},
			},
			want: "coordinator-newer",
		},
		{
			name:       "no matching record",
			profile:    records.ProfileBuilder,
			project:    "hive",
			task:       testTaskID,
			candidates: []records.Agent{{Key: "other-task", Name: "other-task", Profile: records.ProfileBuilder, Project: "hive", Task: "2-task", Status: records.AgentStopped}},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			seedProject(t, home, "hive")
			for _, candidate := range test.candidates {
				writeRecord(t, home, candidate)
			}
			options := RegisterOptions{
				Profile: test.profile, Kind: test.kind, Owns: test.owns, Project: test.project, Task: test.task,
				Name: "new-agent", NameSet: true, PaneID: "w2:p1", Spawned: true,
			}
			envelope := Register(home, options)
			if envelope.Outcome != "ok" {
				t.Fatalf("Register() = %#v; want success", envelope)
			}
			result, ok := envelope.Result.(RegisterResult)
			if !ok {
				t.Fatalf("result = %T; want RegisterResult", envelope.Result)
			}
			got := readAgentField(t, home, result.Agent.Key, "replaces")
			if got != test.want {
				t.Fatalf("replaces = %q; want %q", got, test.want)
			}
		})
	}
}

func readAgentField(t *testing.T, home, key, field string) string {
	t.Helper()
	path := filepath.Join(home, ".hive", "state", "agents", key+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read agent JSON: %v", err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("decode agent JSON: %v", err)
	}
	value, _ := record[field].(string)
	return value
}
