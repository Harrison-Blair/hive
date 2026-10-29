package spawn

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/spf13/cobra"
)

func execute(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	root := &cobra.Command{Use: "hive", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("json", false, "")
	root.AddCommand(New())
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	err := root.Execute()
	var status *common.StatusError
	if errors.As(err, &status) {
		return status.Status, stdout.String(), stderr.String()
	}
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return 0, stdout.String(), stderr.String()
}

func TestSpawnCLIRequiresHarnessAsJSONFailure(t *testing.T) {
	agenttest.NewHerdr(t)
	t.Setenv("HIVE_HOME", agenttest.Home(t))
	status, stdout, _ := execute(t, "--json", "spawn", "--name", "a", "--model", "m")
	var envelope struct {
		Outcome string `json:"outcome"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout %q: %v", stdout, err)
	}
	if status != 1 || envelope.Outcome != "failed" || !strings.Contains(envelope.Error, "--harness") {
		t.Fatalf("status %d, envelope %+v", status, envelope)
	}
}

func TestSpawnCLIUsesHerdrPaneAsCaller(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	t.Setenv("HIVE_HOME", home)
	ws, pane := fake.AddWorkspace("mine", "", false)
	if err := agent.Save(home, agent.Record{Name: "boss", Pane: pane.ID}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_PANE_ID", pane.ID)
	status, stdout, stderr := execute(t, "spawn", "--name", "kid", "--harness", "pi", "--model", "m", "--prompt", "go")
	if status != 0 {
		t.Fatalf("status %d, stderr %q", status, stderr)
	}
	record, err := agent.Load(home, "kid")
	if err != nil || record.Parent != "boss" {
		t.Fatalf("record %+v, err %v", record, err)
	}
	if got := fake.Requests("tab.create"); len(got) != 1 || got[0].Params["workspace_id"] != ws.ID {
		t.Fatalf("tab.create = %+v", got)
	}
	if !strings.Contains(stdout, "agent kid runs in pane "+record.Pane) {
		t.Fatalf("stdout = %q", stdout)
	}
}
