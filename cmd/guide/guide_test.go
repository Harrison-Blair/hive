package guide

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	guidepkg "github.com/Harrison-Blair/hive/internal/guide"
	"github.com/spf13/cobra"
)

// TestPlainCoversEachNodeAndOmitHidden proves AC2 for the tree conversion:
// the root, a parent, a leaf, and the built-in completion nodes are present.
// A hidden command and its subtree and help are absent.
func TestPlainCoversEachNodeAndOmitHidden(t *testing.T) {
	root := &cobra.Command{Use: "hive", Short: "hive"}
	parent := &cobra.Command{Use: "parent", Short: "parent command"}
	leaf := &cobra.Command{Use: "leaf", Short: "leaf command"}
	leaf.Flags().Int("count", 0, "how many")
	hidden := &cobra.Command{Use: "secret", Short: "secret command", Hidden: true}
	hidden.AddCommand(&cobra.Command{Use: "nested", Short: "nested command"})
	parent.AddCommand(leaf)
	root.AddCommand(parent, hidden)

	got := Plain(root)
	var paths []string
	for _, command := range got {
		paths = append(paths, command.Path)
	}
	want := []string{
		"hive", "hive completion", "hive completion bash", "hive completion fish",
		"hive completion powershell", "hive completion zsh", "hive parent", "hive parent leaf",
	}
	if !slices.Equal(paths, want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}

	command := commandByPath(t, got, "hive parent leaf")
	var count *guidepkg.Flag
	foundHelp := false
	for i := range command.Flags {
		switch command.Flags[i].Name {
		case "count":
			count = &command.Flags[i]
		case "help":
			foundHelp = true
			if command.Flags[i].Shorthand != "h" || command.Flags[i].Type != "bool" {
				t.Fatalf("help flag = %+v, want shorthand h and type bool", command.Flags[i])
			}
		}
	}
	if count == nil {
		t.Fatalf("hive parent leaf has no count flag: %v", command.Flags)
	}
	if count.Type != "int" || count.Usage != "how many" || count.Shorthand != "" {
		t.Fatalf("count = %+v, want type int, usage %q, no shorthand", *count, "how many")
	}
	if !foundHelp {
		t.Fatalf("hive parent leaf has no help flag: %v", command.Flags)
	}
}

func commandByPath(t *testing.T, commands []guidepkg.Command, path string) guidepkg.Command {
	t.Helper()
	for _, command := range commands {
		if command.Path == path {
			return command
		}
	}
	t.Fatalf("no command %q in %v", path, commands)
	return guidepkg.Command{}
}

// TestTemporaryFlagFailsDriftThenRegenerationPasses proves AC3 end to end on a
// test tree. A visible flag and a changed command entry each make the check
// fail. Regeneration makes the flag case pass.
func TestTemporaryFlagFailsDriftThenRegenerationPasses(t *testing.T) {
	root := &cobra.Command{Use: "hive", Short: "hive"}
	version := &cobra.Command{Use: "version", Short: "Print the version"}
	root.AddCommand(version)
	tracked := referenceText(root)
	if got := referenceText(root); got != tracked {
		t.Fatal("clean tree differs from its own reference")
	}
	version.Flags().String("temporary-drift", "", "temporary drift probe")
	if got := referenceText(root); got == tracked {
		t.Fatal("the reference did not change after a visible flag was added")
	}

	path := filepath.Join(t.TempDir(), "reference.md")
	if err := writeReferenceText(path, referenceText(root)); err != nil {
		t.Fatalf("writeReferenceText: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written reference: %v", err)
	}
	if got, want := string(data), referenceText(root); got != want {
		t.Fatalf("written reference = %d bytes, want %d", len(got), len(want))
	}
	if got := referenceText(root); got != string(data) {
		t.Fatal("the regenerated reference differs from the live tree")
	}

	version.Short = "changed short help"
	if got := referenceText(root); got == string(data) {
		t.Fatal("the reference did not change after a command entry changed")
	}
}

// TestWriteReferenceSkipsIdenticalContent proves that a second write leaves
// the file alone. The file is read-only, so a write would fail.
func TestWriteReferenceSkipsIdenticalContent(t *testing.T) {
	root := &cobra.Command{Use: "hive", Short: "hive"}
	path := filepath.Join(t.TempDir(), "reference.md")
	if err := writeReferenceText(path, referenceText(root)); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := writeReferenceText(path, referenceText(root)); err != nil {
		t.Fatalf("second write: %v", err)
	}
}
