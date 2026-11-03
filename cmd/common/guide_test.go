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
		"  --json bool: write JSON output\n",
		"\nhive: the root\n",
		"\nhive canary NAME: chirp once\n  -n/--count int=2: chirp count\n  --loud bool: be loud\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Guide() has no %q in:\n%s", want, got)
		}
	}
	for _, absent := range []string{"completion", "hive help", "secret", "  --help"} {
		if strings.Contains(got, absent) {
			t.Errorf("Guide() has %q; want it left out", absent)
		}
	}
	if !strings.HasSuffix(got, guide.Notes()) {
		t.Errorf("Guide() does not end with the usage notes")
	}
}

func TestGuideLeavesOutAnnotatedCommandsAndFlags(t *testing.T) {
	root := &cobra.Command{Use: "hive", Short: "the root", Annotations: map[string]string{GuideOmit: "true"}, Run: func(*cobra.Command, []string) {}}
	group := &cobra.Command{Use: "group", Short: "the group", Annotations: map[string]string{GuideOmit: "true"}}
	child := &cobra.Command{Use: "child", Short: "the child", Run: func(*cobra.Command, []string) {}}
	child.Flags().Bool("kept", false, "stays in the guide")
	child.Flags().Bool("dropped", false, "leaves the guide")
	if err := child.Flags().SetAnnotation("dropped", GuideOmit, []string{"true"}); err != nil {
		t.Fatal(err)
	}
	group.AddCommand(child)
	root.AddCommand(group)

	got := Guide(root)
	if !strings.Contains(got, "\nhive group child: the child\n  --kept bool: stays in the guide\n") {
		t.Errorf("Guide() has no child of the omitted group in:\n%s", got)
	}
	for _, absent := range []string{"\nhive: the root", "\nhive group: the group", "--dropped"} {
		if strings.Contains(got, absent) {
			t.Errorf("Guide() has %q; want it left out", absent)
		}
	}
}
