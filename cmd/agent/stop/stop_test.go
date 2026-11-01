package stop

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/spf13/cobra"
)

func TestStopCLIRejectsRemoveWorktree(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	t.Setenv("HIVE_HOME", home)
	fake.AddWorkspace("mine", "", false)
	pane := fake.AddPane("w1", "")
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: pane.ID}); err != nil {
		t.Fatal(err)
	}
	root := &cobra.Command{Use: "hive", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("json", false, "")
	root.AddCommand(New())
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"stop", "kid", "--remove-worktree"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "unknown flag: --remove-worktree") {
		t.Fatalf("err %v, want unknown flag --remove-worktree", err)
	}
	if _, ok := fake.Pane(pane.ID); !ok {
		t.Fatalf("pane %s closed", pane.ID)
	}
}

func TestStopCLILogsCallerAndTargetOnEachEvent(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	t.Setenv("HIVE_HOME", home)
	_, boss := fake.AddWorkspace("mine", "", false)
	kid := fake.AddPane("w1", "")
	for _, record := range []agent.Record{{Name: "boss", Pane: boss.ID}, {Name: "kid", Pane: kid.ID}} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HERDR_PANE_ID", boss.ID)
	root := &cobra.Command{Use: "hive", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("json", false, "")
	root.AddCommand(New())
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"stop", "kid"})
	var status *common.StatusError
	if err := root.Execute(); !errors.As(err, &status) || status.Status != 0 {
		t.Fatalf("err %v, output %q; want status 0", err, out.String())
	}
	agenttest.CheckAgentFields(t, agenttest.CommandEvents(t, home, "agent stop"), "boss", "kid",
		"command start", "herdr request", "herdr reply", "agent stop", "command end")
}
