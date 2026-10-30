package cmd

import (
	"path/filepath"
	"testing"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
)

// longModel is longer than each model name that hive starts today.
const longModel = "openrouter/deepseek/deepseek-v4.1-flash-thinking-preview:xhigh"

// TestGuideStartLineFitsTheBudget checks the full start line of each harness
// with the guide, a long model, and the inbox arguments of a long home against
// harness.MaxStartLine.
func TestGuideStartLineFitsTheBudget(t *testing.T) {
	guide := common.Guide(NewRootCmd())
	home := filepath.Join(t.TempDir(), "home", "a-long-user-name", ".hive")
	for _, kind := range harness.KnownKinds() {
		profile, err := harness.Lookup(kind)
		if err != nil {
			t.Fatal(err)
		}
		inbox, err := profile.InboxArgs(home)
		if err != nil {
			t.Fatal(err)
		}
		extra := append(append(profile.Trust().Args, inbox...), profile.GuideArgs(guide)...)
		args := profile.Args(longModel, extra)
		length := harness.StartLineLength(profile.Kind(), args)
		t.Logf("%s start line = %d bytes", kind, length)
		if length > harness.MaxStartLine {
			t.Errorf("%s start line = %d bytes; want at most %d", kind, length, harness.MaxStartLine)
		}
	}
}
