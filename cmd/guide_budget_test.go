package cmd

import (
	"testing"

	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
)

// longModel is longer than each model name that hive starts today.
const longModel = "openrouter/deepseek/deepseek-v4.1-flash-thinking-preview:xhigh"

// TestGuideStartLineFitsTheBudget checks the full start line of each harness
// with the guide and a long model against harness.MaxStartLine.
func TestGuideStartLineFitsTheBudget(t *testing.T) {
	guide := common.Guide(NewRootCmd())
	for _, kind := range harness.KnownKinds() {
		profile, err := harness.Lookup(kind)
		if err != nil {
			t.Fatal(err)
		}
		args := profile.Args(longModel, append(profile.Trust().Args, profile.GuideArgs(guide)...))
		length := harness.StartLineLength(profile.Kind(), args)
		t.Logf("%s start line = %d bytes", kind, length)
		if length > harness.MaxStartLine {
			t.Errorf("%s start line = %d bytes; want at most %d", kind, length, harness.MaxStartLine)
		}
	}
}
