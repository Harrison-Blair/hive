package common

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/spf13/cobra"
)

func TestGuideRendersEachCommandOfTheTree(t *testing.T) {
	root := &cobra.Command{Use: "hive", Short: "the root", Run: func(*cobra.Command, []string) {}}
	root.PersistentFlags().Bool("json", false, "write JSON output")
	canary := &cobra.Command{Use: "canary NAME", Short: "chirp once", Run: func(*cobra.Command, []string) {}}
	canary.Flags().Bool("loud", false, "be loud")
	canary.Flags().IntP("count", "n", 2, "chirp count")
	root.AddCommand(canary, &cobra.Command{Use: "secret", Hidden: true, Run: func(*cobra.Command, []string) {}})
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()

	got := Guide(root)
	for _, want := range []string{
		"- `--json` (bool): write JSON output\n",
		"### hive\n\n`hive`\n\nthe root\n",
		"### hive canary\n\n`hive canary NAME [flags]`\n\nchirp once\n\n- `-n, --count` (int, default 2): chirp count\n- `--loud` (bool): be loud\n\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Guide() has no %q in:\n%s", want, got)
		}
	}
	for _, absent := range []string{"completion", "### hive help", "secret", "--help` ("} {
		if strings.Contains(got, absent) {
			t.Errorf("Guide() has %q; want it left out", absent)
		}
	}
	if !strings.HasSuffix(got, guide.Notes()) {
		t.Errorf("Guide() does not end with the usage notes")
	}
}
