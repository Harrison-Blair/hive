package cmd

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func TestCommandsWithoutAnAgentLogTheCallerAndNoTarget(t *testing.T) {
	home := requestTestHome(t)
	t.Setenv("HIVE_HOME", home)
	t.Setenv("HERDR_PANE_ID", "")
	for _, args := range [][]string{{"project", "list"}, {"logs"}} {
		var output strings.Builder
		if err := ExecuteWithArgs(args, strings.NewReader(""), &output); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, output.String())
		}
		agenttest.CheckAgentFields(t, agenttest.CommandEvents(t, home, strings.Join(args, " ")), "user", "", "command start", "command end")
	}
}
