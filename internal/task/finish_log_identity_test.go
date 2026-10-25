package task

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// finishLogActor is one resolved caller and the routes that find it.
type finishLogActor struct {
	key     string
	name    string
	pane    string
	history string
}

// finishLogRoute is one way to name a caller.
type finishLogRoute struct {
	name  string
	agent string
	pane  string
}

func finishLogRoutes(actor finishLogActor) []finishLogRoute {
	return []finishLogRoute{
		{"key", actor.key, ""},
		{"name", actor.name, ""},
		{"pane", "", actor.pane},
		{"history", "", actor.history},
	}
}

var (
	finishLogWorker   = finishLogActor{key: "worker-key", name: "worker-name", pane: "worker-pane", history: "worker-history"}
	finishLogVerifier = finishLogActor{key: "verifier-key", name: "verifier-name", pane: "verifier-pane", history: "verifier-history"}
)

// seedFinishLogHome writes one task and two live callers. The task names the
// worker name as its acting agent.
func seedFinishLogHome(t *testing.T, status string) string {
	t.Helper()
	home := newHome(t)
	registerProject(t, home, "one")
	seedTask(t, home, records.Task{
		ID: "1-literal", Project: "one", Title: "Literal",
		ActingAgent: "worker-name", Status: status, DependsOn: []string{},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
	seedAgent(t, home, records.Agent{
		Key: "worker-key", Name: "worker-name", Role: "worker", Project: "one", Task: "1-literal",
		Status: records.AgentLive, PaneID: "worker-pane", PaneHistory: []string{"worker-history"},
		Generation: 1, StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
	seedAgent(t, home, records.Agent{
		Key: "verifier-key", Name: "verifier-name", Role: "verifier", Project: "one", Task: "1-literal",
		Status: records.AgentLive, PaneID: "verifier-pane", PaneHistory: []string{"verifier-history"},
		Generation: 1, StartedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	})
	return home
}

// readOnlyFinishDir makes one real state directory static and unwritable. It
// returns a restore function. It uses the actual filesystem, not a seam.
func readOnlyFinishDir(t *testing.T, path string) func() {
	t.Helper()
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o500); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Chmod(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
}

// finishLogLines returns the task transition, effect, and error events.
func finishLogLines(t *testing.T, raw string) []map[string]any {
	t.Helper()
	events := []map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line %q: %v", line, err)
		}
		switch event["msg"] {
		case "task transition", "task effect", "task error":
			events = append(events, event)
		}
	}
	return events
}

func finishLogResources(envelope output.Envelope) []string {
	resources := make([]string, 0, len(envelope.Effects))
	for _, effect := range envelope.Effects {
		resources = append(resources, effect.Resource)
	}
	return resources
}

func TestDirectFinishLogsResolvedKeyWithoutOutputPointer(t *testing.T) {
	operations := []struct {
		name    string
		command string
		status  string
		fail    bool
		reason  string
		success finishLogActor
		refusal finishLogActor
	}{
		{"complete", "complete", "assigned", false, "", finishLogWorker, finishLogVerifier},
		{"verify", "verify", "completed", false, "", finishLogVerifier, finishLogWorker},
		{"verify-fail", "verify", "completed", true, "private-reason-72e9", finishLogVerifier, finishLogWorker},
	}
	phases := []struct {
		name    string
		dir     string
		outcome string
		phase   string
		effects []string
	}{
		{"success", "", "ok", "complete", []string{"task", "counter", "inbox", "agent"}},
		{"task", filepath.Join(".hive", "state", "tasks"), "failed", "task", nil},
		{"counter", filepath.Join(".hive", "state", "counters"), "partial", "counter", []string{"task"}},
		{"record", filepath.Join(".hive", "state", "inbox", "coordinator"), "partial", "record", []string{"task", "counter"}},
		{"finished", filepath.Join(".hive", "state", "agents"), "partial", "finished", []string{"task", "counter", "inbox"}},
		{"refusal", "", "failed", "identity", nil},
	}
	for _, op := range operations {
		for _, phase := range phases {
			for _, route := range finishLogRoutes(op.success) {
				t.Run(op.name+"/"+phase.name+"/"+route.name, func(t *testing.T) {
					actor := op.success
					if phase.name == "refusal" {
						actor = op.refusal
					}
					opts := TransitionOptions{
						Task: "1-literal", Agent: route.agent, PaneID: route.pane,
						Fail: op.fail, Reason: op.reason,
					}
					if phase.name == "refusal" {
						// Refusal routes keep the same actor and use the refusal
						// actor key, name, pane, or history.
						refusalRoutes := finishLogRoutes(actor)
						for _, candidate := range refusalRoutes {
							if candidate.name == route.name {
								opts.Agent = candidate.agent
								opts.PaneID = candidate.pane
							}
						}
					}
					home := seedFinishLogHome(t, op.status)
					if phase.dir != "" {
						restore := readOnlyFinishDir(t, filepath.Join(home, phase.dir))
						defer restore()
					}
					var logs bytes.Buffer
					opts.Logger = slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
					var envelope output.Envelope
					if op.command == "complete" {
						envelope = Complete(home, opts)
					} else {
						envelope = Verify(home, opts)
					}
					if envelope.Outcome != phase.outcome || envelope.Phase != phase.phase {
						t.Fatalf("finish = %#v; want %s/%s", envelope, phase.outcome, phase.phase)
					}
					if got, want := finishLogResources(envelope), phase.effects; strings.Join(got, ",") != strings.Join(want, ",") {
						t.Fatalf("effects = %v; want %v", got, want)
					}
					if phase.name == "refusal" && envelope.Error == nil {
						t.Fatal("refusal error = nil; want a diagnostic")
					}
					events := finishLogLines(t, logs.String())
					wantEvents := len(phase.effects)
					if len(phase.effects) > 0 {
						wantEvents++
					}
					if envelope.Error != nil {
						wantEvents++
					}
					if len(events) != wantEvents {
						t.Fatalf("events = %d %#v; want %d", len(events), events, wantEvents)
					}
					resources := []string{}
					for _, event := range events {
						if event["agent_id"] != actor.key {
							t.Errorf("%s: agent_id=%v, want resolved key %s", event["msg"], event["agent_id"], actor.key)
						}
						if event["task_id"] != "1-literal" || event["project_id"] != "one" {
							t.Errorf("event lost task or project IDs: %#v", event)
						}
						switch event["msg"] {
						case "task effect":
							resource, _ := event["resource"].(string)
							resources = append(resources, resource)
							if resource == "counter" || resource == "inbox" {
								if event["role"] != "project:one" {
									t.Errorf("effect %s lost the project inbox: %#v", resource, event)
								}
							}
						case "task error":
							if event["phase"] != envelope.Phase {
								t.Errorf("error phase = %v; want %s", event["phase"], envelope.Phase)
							}
						}
					}
					if strings.Join(resources, ",") != strings.Join(phase.effects, ",") {
						t.Errorf("logged effects = %v; want %v", resources, phase.effects)
					}
					if op.reason != "" && strings.Contains(logs.String(), op.reason) {
						t.Error("log holds the reason body")
					}
				})
			}
		}
	}
}

// readDefaultFinishLog reads the actual file log of one command.
func readDefaultFinishLog(t *testing.T, home, command string) []map[string]any {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, ".hive", "logs", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	events := []map[string]any{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range finishLogLines(t, string(raw)) {
			if event["component"] == "task" && event["command"] == command {
				events = append(events, event)
			}
		}
	}
	return events
}

func TestDirectFinishDefaultLoggerKeepsResolvedKey(t *testing.T) {
	operations := []struct {
		name    string
		command string
		status  string
		fail    bool
		reason  string
		key     string
		pane    string
	}{
		{"complete", "task complete", "assigned", false, "", "worker-key", "worker-pane"},
		{"verify", "task verify", "completed", false, "", "verifier-key", "verifier-history"},
		{"verify-fail", "task verify", "completed", true, "private-reason-72e9", "verifier-key", "verifier-history"},
	}
	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			home := seedFinishLogHome(t, op.status)
			opts := TransitionOptions{Task: "1-literal", PaneID: op.pane, Fail: op.fail, Reason: op.reason}
			var envelope output.Envelope
			if op.command == "task complete" {
				envelope = Complete(home, opts)
			} else {
				envelope = Verify(home, opts)
			}
			if envelope.Outcome != "ok" || len(envelope.Effects) != 4 {
				t.Fatalf("finish = %#v; want ok with four effects", envelope)
			}
			events := readDefaultFinishLog(t, home, op.command)
			if len(events) != 5 {
				t.Fatalf("actual default log events = %d %#v; want five", len(events), events)
			}
			for _, event := range events {
				if event["agent_id"] != op.key {
					t.Errorf("actual default log %s has agent_id=%v, want %s", event["msg"], event["agent_id"], op.key)
				}
			}
			if op.reason != "" {
				for _, event := range events {
					raw, _ := json.Marshal(event)
					if strings.Contains(string(raw), op.reason) {
						t.Error("actual default log holds the reason body")
					}
				}
			}
		})
	}
}

func TestDirectFinishDefaultLoggerKeepsResolvedKeyOnCounterFailure(t *testing.T) {
	home := seedFinishLogHome(t, "assigned")
	restore := readOnlyFinishDir(t, filepath.Join(home, ".hive", "state", "counters"))
	defer restore()
	envelope := Complete(home, TransitionOptions{Task: "1-literal", PaneID: "worker-pane"})
	if envelope.Outcome != "partial" || envelope.Phase != "counter" {
		t.Fatalf("finish = %#v; want partial/counter", envelope)
	}
	events := readDefaultFinishLog(t, home, "task complete")
	if len(events) != 3 {
		t.Fatalf("actual default log events = %d %#v; want transition, task effect, and error", len(events), events)
	}
	for _, event := range events {
		if event["agent_id"] != "worker-key" {
			t.Errorf("actual default log %s has agent_id=%v, want worker-key", event["msg"], event["agent_id"])
		}
	}
}

func TestDirectFinishOutputPointerAndInvocationIsolation(t *testing.T) {
	home := seedFinishLogHome(t, "assigned")
	var actor records.Agent
	var first bytes.Buffer
	firstEnvelope := Complete(home, TransitionOptions{
		Task: "1-literal", PaneID: "worker-pane", ResolvedActor: &actor,
		Logger: slog.New(slog.NewJSONHandler(&first, &slog.HandlerOptions{Level: slog.LevelDebug})),
	})
	if firstEnvelope.Outcome != "ok" || actor.Key != "worker-key" {
		t.Fatalf("first finish = %#v, actor %#v; want ok with worker-key", firstEnvelope, actor)
	}
	if count := strings.Count(first.String(), `"agent_id":"worker-key"`); count != 5 {
		t.Fatalf("first log has %d resolved key events; want five: %s", count, first.String())
	}
	var second bytes.Buffer
	secondEnvelope := Verify(home, TransitionOptions{
		Task: "1-literal", PaneID: "verifier-pane", ResolvedActor: &actor,
		Logger: slog.New(slog.NewJSONHandler(&second, &slog.HandlerOptions{Level: slog.LevelDebug})),
	})
	if secondEnvelope.Outcome != "ok" || actor.Key != "verifier-key" {
		t.Fatalf("second finish = %#v, actor %#v; want ok with verifier-key", secondEnvelope, actor)
	}
	if count := strings.Count(second.String(), `"agent_id":"verifier-key"`); count != 5 {
		t.Fatalf("second log has %d current key events; want five: %s", count, second.String())
	}
	if strings.Contains(second.String(), `"agent_id":"worker-key"`) {
		t.Fatalf("second log reused the prior actor: %s", second.String())
	}
	var third bytes.Buffer
	thirdHome := seedFinishLogHome(t, "assigned")
	thirdEnvelope := Complete(thirdHome, TransitionOptions{
		Task: "1-literal", PaneID: "worker-history",
		Logger: slog.New(slog.NewJSONHandler(&third, &slog.HandlerOptions{Level: slog.LevelDebug})),
	})
	if thirdEnvelope.Outcome != "ok" || actor.Key != "verifier-key" {
		t.Fatalf("third finish = %#v, old actor %#v; want ok and an untouched prior output", thirdEnvelope, actor)
	}
	events := finishLogLines(t, third.String())
	if len(events) != 5 {
		t.Fatalf("third log = %#v; want five events", events)
	}
	for _, event := range events {
		if event["agent_id"] != "worker-key" {
			t.Fatalf("third log reused the prior actor: %#v", event)
		}
	}
}
