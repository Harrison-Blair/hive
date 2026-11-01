package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func TestRootAgentListJSONEnvelope(t *testing.T) {
	testHome(t)
	agenttest.NewHerdr(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--json", "agent", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v; output %q", err, stdout.String())
	}
	if string(envelope["outcome"]) != `"ok"` || string(envelope["result"]) != "[]" {
		t.Fatalf("envelope = %s", stdout.String())
	}
}

func TestRootAgentCommandsAreRegistered(t *testing.T) {
	root := NewRootCmd()
	for _, path := range [][]string{{"agent", "spawn"}, {"agent", "stop"}, {"agent", "send"}, {"agent", "list"}, {"agent", "graph"}} {
		found, _, err := root.Find(path)
		if err != nil || found.Name() != path[1] {
			t.Fatalf("hive %s: found %v, err %v", strings.Join(path, " "), found, err)
		}
	}
}

func TestRootAgentGraphJSONIsTheListSnapshot(t *testing.T) {
	home := testHome(t)
	fake := agenttest.NewHerdr(t)
	fake.AddWorkspace("mine", "", false)
	pane := fake.AddPane("w1", "")
	fake.StartAgent(pane.ID, "hive-coordinator", "claude", "working")
	if err := agent.Save(home, agent.Record{Name: "hive-coordinator", Pane: pane.ID, Harness: "claude", Model: "opus"}); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		var stdout, stderr bytes.Buffer
		if status := executeWithStreams(args, strings.NewReader(""), &stdout, &stderr); status != 0 {
			t.Fatalf("hive %v: status = %d, stderr = %q; want 0", args, status, stderr.String())
		}
		return stdout.String()
	}
	graph, list := run("--json", "agent", "graph"), run("--json", "agent", "list")
	if graph != list {
		t.Fatalf("graph --json = %s; want the list snapshot %s", graph, list)
	}
	if !strings.Contains(graph, `"status":"working"`) {
		t.Fatalf("graph --json = %s; want the live status", graph)
	}
}
