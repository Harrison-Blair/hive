package common

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/spf13/cobra"
)

func TestGuideIncludesPublicTreeAndExcludesHiddenItems(t *testing.T) {
	root := &cobra.Command{Use: "hive", Short: "the root", Run: func(*cobra.Command, []string) {}}
	root.PersistentFlags().Bool("json", false, "write JSON output")
	group := &cobra.Command{Use: "group", Short: "the group"}
	child := &cobra.Command{Use: "child NAME", Short: "run child", Run: func(*cobra.Command, []string) {}}
	child.Flags().IntP("count", "n", 2, "the count")
	child.Flags().Bool("secret", false, "private")
	if err := child.Flags().MarkHidden("secret"); err != nil {
		t.Fatal(err)
	}
	group.AddCommand(child, &cobra.Command{Use: "hidden", Hidden: true, Run: func(*cobra.Command, []string) {}})
	root.AddCommand(group)
	got := Guide(root)
	for _, want := range []string{"\nhive: the root\n", "\nhive group: the group\n", "\nhive group child NAME: run child\n  -n/--count int=2: the count\n", "\nhive help [command]:", "\nhive completion bash:", "--no-descriptions", "--json bool: write JSON output"} {
		if !strings.Contains(got, want) {
			t.Errorf("Guide() missing %q", want)
		}
	}
	for _, absent := range []string{"--secret", "hive group hidden"} {
		if strings.Contains(got, absent) {
			t.Errorf("Guide() contains %q", absent)
		}
	}
	if !strings.HasSuffix(got, guide.Notes()) {
		t.Error("Guide() missing notes")
	}
}

func TestStartupGuideSelectsCoreCommandsAndTracksMetadata(t *testing.T) {
	root := &cobra.Command{Use: "hive", Short: "root"}
	root.PersistentFlags().String("global-canary", "v", "global canary")
	agent := &cobra.Command{Use: "agent", Short: "group"}
	agent.PersistentFlags().Bool("group-canary", false, "inherited by agent commands")
	for _, name := range []string{"list", "send", "spawn", "stop", "graph"} {
		c := &cobra.Command{Use: name + " ARG", Short: "description of " + name, Run: func(*cobra.Command, []string) {}}
		c.Flags().IntP("canary", "c", 7, "canary for "+name)
		agent.AddCommand(c)
	}
	root.AddCommand(agent, &cobra.Command{Use: "other", Short: "other", Run: func(*cobra.Command, []string) {}})
	got := StartupGuide(root)
	if !strings.HasSuffix(got, guide.Notes()) {
		t.Error("StartupGuide() missing the complete usage notes")
	}
	for _, name := range []string{"list", "send", "spawn", "stop"} {
		want := "\nhive agent " + name + " ARG: description of " + name + "\n  -c/--canary int=7: canary for " + name + "\n"
		if !strings.Contains(got, want) {
			t.Errorf("StartupGuide() missing %q", want)
		}
	}
	for _, want := range []string{"--global-canary string=v: global canary", "--group-canary bool: inherited by agent commands", "-h/--help", "hive guide", "## Usage notes"} {
		if !strings.Contains(got, want) {
			t.Errorf("StartupGuide() missing %q", want)
		}
	}
	for _, absent := range []string{"hive agent graph:", "hive agent graph ARG:", "hive other:"} {
		if strings.Contains(got, absent) {
			t.Errorf("StartupGuide() contains %q", absent)
		}
	}
	spawn, _, err := root.Find([]string{"agent", "spawn"})
	if err != nil {
		t.Fatal(err)
	}
	spawn.Short = "changed description"
	spawn.Flags().String("new-flag", "default", "new description")
	got = StartupGuide(root)
	for _, want := range []string{"hive agent spawn ARG: changed description", "--new-flag string=default: new description"} {
		if !strings.Contains(got, want) {
			t.Errorf("StartupGuide() after metadata change missing %q", want)
		}
	}
}
