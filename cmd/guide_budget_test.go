package cmd

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
)

// startLineBudget bounds the command line that herdr agent.start types into a
// new pane shell. A tty in canonical mode cuts a line at 4095 bytes; see
// reference/dogfood/friction/entries/agent-start-line-cut-at-4095-bytes.md.
const startLineBudget = 4000

// longModel is longer than each model name that hive starts today.
const longModel = "openrouter/deepseek/deepseek-v4.1-flash-thinking-preview:xhigh"

// TestGuideStartLineFitsTheBudget builds the full start line of each harness
// with the guide, each argument in single quotes as herdr types it.
func TestGuideStartLineFitsTheBudget(t *testing.T) {
	guide := common.Guide(NewRootCmd())
	for _, kind := range harness.KnownKinds() {
		profile, err := harness.Lookup(kind)
		if err != nil {
			t.Fatal(err)
		}
		args := profile.Args(longModel, append(profile.Trust().Args, profile.GuideArgs(guide)...))
		line := profile.ProcessNames()[0]
		for _, arg := range args {
			line += " '" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
		}
		t.Logf("%s start line = %d bytes", kind, len(line))
		if len(line) >= startLineBudget {
			t.Errorf("%s start line = %d bytes; want below %d", kind, len(line), startLineBudget)
		}
	}
}
