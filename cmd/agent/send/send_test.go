package send

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/spf13/cobra"
)

func execute(t *testing.T, args ...string) (int, string) {
	t.Helper()
	root := &cobra.Command{Use: "hive", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("json", false, "")
	root.AddCommand(New())
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	var status *common.StatusError
	if errors.As(err, &status) {
		return status.Status, out.String()
	}
	if err != nil {
		return -1, err.Error()
	}
	return 0, out.String()
}

// claudeAgent starts a fake claude agent in pane with an inbox socket.
func claudeAgent(t *testing.T, fake *agenttest.Herdr, paneID string) <-chan string {
	t.Helper()
	profile, err := harness.Lookup("claude")
	if err != nil {
		t.Fatal(err)
	}
	fake.StartAgent(paneID, "kid", "claude", "idle")
	fake.SetArgs(paneID, profile.Args("m", nil))
	pane, _ := fake.Pane(paneID)
	return agenttest.ListenClaude(t, pane.PID)
}

func TestSendCLISendsTextFromCallerPane(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	t.Setenv("HIVE_HOME", home)
	_, boss := fake.AddWorkspace("mine", "", false)
	kid := fake.AddPane("w1", "")
	lines := claudeAgent(t, fake, kid.ID)
	for _, record := range []agent.Record{{Name: "boss", Pane: boss.ID}, {Name: "kid", Pane: kid.ID, Harness: "claude"}} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HERDR_PANE_ID", boss.ID)
	if status, out := execute(t, "send", "kid", "hi there"); status != 0 {
		t.Fatalf("status %d, output %q", status, out)
	}
	want := `{"type":"user","message":{"role":"user","content":"⬢ boss: hi there"}}` + "\n"
	select {
	case line := <-lines:
		if line != want {
			t.Fatalf("line = %q; want %q", line, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no message reached the socket")
	}
}

func TestSendCLINeedsNameAndText(t *testing.T) {
	if status, _ := execute(t, "send", "kid"); status == 0 {
		t.Fatalf("status 0 for a send without TEXT")
	}
}

func TestSendCLILogsCallerAndTargetOnEachEvent(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	t.Setenv("HIVE_HOME", home)
	_, boss := fake.AddWorkspace("mine", "", false)
	kid := fake.AddPane("w1", "")
	claudeAgent(t, fake, kid.ID)
	for _, record := range []agent.Record{{Name: "boss", Pane: boss.ID}, {Name: "kid", Pane: kid.ID, Harness: "claude"}} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HERDR_PANE_ID", boss.ID)
	if status, out := execute(t, "send", "kid", "hi"); status != 0 {
		t.Fatalf("status %d, output %q", status, out)
	}
	agenttest.CheckAgentFields(t, agenttest.CommandEvents(t, home, "agent send"), "boss", "kid",
		"command start", "herdr request", "herdr reply", "agent send", "command end")
}

func TestSendCLILogsUserWhenThePaneHasNoRecord(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	t.Setenv("HIVE_HOME", home)
	fake.AddWorkspace("mine", "", false)
	kid := fake.AddPane("w1", "")
	claudeAgent(t, fake, kid.ID)
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: kid.ID, Harness: "claude"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_PANE_ID", "")
	if status, out := execute(t, "send", "kid", "hi"); status != 0 {
		t.Fatalf("status %d, output %q", status, out)
	}
	agenttest.CheckAgentFields(t, agenttest.CommandEvents(t, home, "agent send"), "user", "kid", "command start", "command end")
}

func TestSendCLILogsTheTargetOnAFailedCallerLookup(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	t.Setenv("HIVE_HOME", home)
	if err := os.MkdirAll(filepath.Join(home, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "agents", "broken.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	execute(t, "send", "kid", "hi")
	agenttest.CheckAgentFields(t, agenttest.CommandEvents(t, home, "agent send"), "user", "kid", "caller lookup failed", "command start", "command end")
}

// TestSendCLILogsTheLevelOfTheOutcome checks the level of the agent send
// event in the log file: WARN for partial and ERROR for failed.
func TestSendCLILogsTheLevelOfTheOutcome(t *testing.T) {
	cases := []struct {
		name, outcome, level string
		args                 []string
	}{
		{"partial", "partial", "WARN", []string{"send", "--wait", "kid", "hi"}},
		{"failed", "failed", "ERROR", []string{"send", "ghost", "hi"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fake := agenttest.NewHerdr(t)
			home := agenttest.Home(t)
			t.Setenv("HIVE_HOME", home)
			t.Setenv("HERDR_PANE_ID", "")
			fake.AddWorkspace("mine", "", false)
			kid := fake.AddPane("w1", "")
			claudeAgent(t, fake, kid.ID)
			fake.NoTurn = true
			if err := agent.Save(home, agent.Record{Name: "kid", Pane: kid.ID, Harness: "claude"}); err != nil {
				t.Fatal(err)
			}
			execute(t, test.args...)
			for _, event := range agenttest.CommandEvents(t, home, "agent send") {
				if event["msg"] == "agent send" {
					if event["outcome"] != test.outcome || event["level"] != test.level {
						t.Fatalf("agent send event = %v; want outcome %s at level %s", event, test.outcome, test.level)
					}
					return
				}
			}
			t.Fatal("no agent send event")
		})
	}
}
