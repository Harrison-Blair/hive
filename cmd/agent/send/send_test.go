package send

import (
	"bytes"
	"errors"
	"slices"
	"testing"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
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

func TestSendCLISendsTextFromCallerPane(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	t.Setenv("HIVE_HOME", home)
	_, boss := fake.AddWorkspace("mine", "", false)
	kid := fake.AddPane("w1", "")
	fake.StartAgent(kid.ID, "kid", "pi", "idle")
	for _, record := range []agent.Record{{Name: "boss", Pane: boss.ID}, {Name: "kid", Pane: kid.ID}} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HERDR_PANE_ID", boss.ID)
	if status, out := execute(t, "send", "kid", "hi there"); status != 0 {
		t.Fatalf("status %d, output %q", status, out)
	}
	pane, _ := fake.Pane(kid.ID)
	if !slices.Equal(pane.Prompts, []string{"⬢ boss: hi there"}) {
		t.Fatalf("prompts = %q", pane.Prompts)
	}
}

func TestSendCLINeedsNameAndText(t *testing.T) {
	if status, _ := execute(t, "send", "kid"); status == 0 {
		t.Fatalf("status 0 for a send without TEXT")
	}
}
