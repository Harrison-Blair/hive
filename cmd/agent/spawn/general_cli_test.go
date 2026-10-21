package spawn_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// seedParent writes one live registered agent in pane w1:p9.
func seedParent(t *testing.T, home string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	parent := records.Agent{Key: "boss", Name: "boss", Role: "worker", Project: "demo", Task: "1-demo", Status: records.AgentLive, PaneID: "w1:p9", Generation: 1}
	if err := registry.WriteIn([]string{"agents"}, parent.Key, parent); err != nil {
		t.Fatal(err)
	}
}

func TestSpawnGeneralAgentRecordsRoleAndCallerParent(t *testing.T) {
	home := cliHome(t)
	seedParent(t, home)
	dir := t.TempDir()
	fake := newCLIFake(t, "agent.start", `{"error":{"code":"invalid_agent_name","message":"bad"}}`)
	fake.replies["pane.get"] = []string{
		`{"type":"pane_info","pane":{"pane_id":"w1:p9","terminal_id":"term_p","workspace_id":"w1","tab_id":"w1:t1","focused":false,"agent_status":"idle","revision":1,"cwd":"/tmp"}}`,
		`{"type":"pane_info","pane":{"pane_id":"w1:p5","terminal_id":"term_x","workspace_id":"w1","tab_id":"w1:t3","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + dir + `"}}`,
		`{"type":"pane_info","pane":{"pane_id":"w1:p5","terminal_id":"term_x","workspace_id":"w1","tab_id":"w1:t3","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + dir + `"}}`,
	}
	fake.replies["pane.process_info"] = []string{`{"type":"pane_process_info","process_info":{"pane_id":"w1:p5","shell_pid":1,"foreground_process_group_id":1,"foreground_processes":[{"pid":1,"name":"bash","argv":["/usr/bin/bash"],"cmdline":"/usr/bin/bash","cwd":"/tmp"}]}}`}
	fake.replies["tab.create"] = []string{`{"type":"tab_created","tab":{"tab_id":"w1:t3","workspace_id":"w1","number":3,"label":"x","focused":false,"pane_count":1,"agent_status":"unknown"},"root_pane":{"pane_id":"w1:p5","terminal_id":"term_x","workspace_id":"w1","tab_id":"w1:t3","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + dir + `"}}`}
	fake.replies["pane.list"] = []string{`{"type":"pane_list","panes":[{"pane_id":"w1:p5","terminal_id":"term_x","workspace_id":"w1","tab_id":"w1:t3","focused":false,"agent_status":"unknown","revision":1,"cwd":"` + dir + `"}]}`}
	fake.replies["pane.close"] = []string{`{"type":"ok"}`}
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	status, stdout, stderr := runCLI(t, home, "--json", "agent", "spawn", "--role", "agent", "--name", "x", "--prompt", "T", "--harness", "pi", "--cwd", dir)
	if status != 1 {
		t.Fatalf("status = %d; want 1 for the failed start (stderr %q)", status, stderr)
	}
	var response struct {
		Result struct {
			Agent records.Agent `json:"agent"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &response); err != nil {
		t.Fatalf("decode spawn response: %v: %q", err, stdout)
	}
	agent := response.Result.Agent
	if agent.Key != "x" || agent.Role != "agent" || agent.Parent != "boss" {
		t.Fatalf("record key/role/parent = %q/%q/%q; want x/agent/boss (stdout %s)", agent.Key, agent.Role, agent.Parent, stdout)
	}
}

func TestSpawnGeneralAgentFlags(t *testing.T) {
	home := cliHome(t)
	seedParent(t, home)
	dir := t.TempDir()
	t.Setenv("HERDR_PANE_ID", "w1:p9")
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{name: "both prompts", args: []string{"--harness", "pi", "--prompt", "T", "--prompt-file", "/none"}, want: "one of --prompt or --prompt-file"},
		{name: "prompt file", args: []string{"--harness", "pi", "--prompt-file", dir + "/none"}, want: "prompt file:"},
		{name: "cwd", args: []string{"--harness", "pi", "--prompt", "T", "--cwd", dir + "/none"}, want: "is not a directory"},
		{name: "trust refusal", args: []string{"--harness", "codex", "--prompt", "T", "--cwd", dir}, want: "add --allow-trust-dialog"},
		{name: "trust allowed", args: []string{"--harness", "codex", "--prompt", "T", "--cwd", dir, "--allow-trust-dialog"}, want: "tab create:"},
	} {
		t.Run(test.name, func(t *testing.T) {
			newCLIFake(t, "tab.create")
			args := append([]string{"agent", "spawn", "--role", "agent", "--name", "x"}, test.args...)
			status, _, stderr := runCLI(t, home, args...)
			if status != 1 || !strings.Contains(stderr, test.want) {
				t.Fatalf("status %d stderr %q; want 1 and %q", status, stderr, test.want)
			}
		})
	}
}

func TestSpawnPaneHelpStatesStopResult(t *testing.T) {
	home := cliHome(t)
	_, stdout, _ := runCLI(t, home, "agent", "spawn", "--help")
	if !strings.Contains(stdout, "agent stop closes it; herdr then closes its tab or workspace when it holds no other pane") {
		t.Fatalf("help = %q; want the --pane stop result", stdout)
	}
}
