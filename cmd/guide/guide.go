// Package guide wires the hive guide command.
package guide

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"

	"github.com/Harrison-Blair/hive/cmd/common"
	guidepkg "github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// referencePath is the tracked reference file, relative to the working
// directory of the checkout that runs the write.
const referencePath = "internal/guide/reference.md"

type guideResult struct {
	Flows     string `json:"flows"`
	Reference string `json:"reference"`
}

// New returns the guide command.
func New() *cobra.Command {
	command := &cobra.Command{
		Use:   "guide",
		Short: "print the hive guide",
		Args:  cobra.NoArgs,
	}
	command.Flags().Bool("write", false, "write the reference of the live command tree")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		write, _ := cmd.Flags().GetBool("write")
		return common.Run(cmd, "guide", func(string) (output.Envelope, string, error) {
			effects := []output.Effect{}
			reference := guidepkg.Reference()
			if write {
				reference = referenceText(cmd.Root())
				if err := writeReferenceText(referencePath, reference); err != nil {
					return output.Envelope{
						Outcome: "failed",
						Phase:   "write",
						Effects: effects,
						Error:   err.Error(),
						Result:  nil,
					}, "", nil
				}
				effects = append(effects, output.Effect{Resource: "guide reference", Path: referencePath, Action: "updated"})
			}
			text := guidepkg.Flows() + "\n" + reference
			return output.Envelope{
				Outcome: "ok",
				Phase:   "complete",
				Effects: effects,
				Error:   nil,
				Result:  guideResult{Flows: guidepkg.Flows(), Reference: reference},
			}, text, nil
		})
	}
	return command
}

// referenceText builds the reference part from the live command tree.
func referenceText(root *cobra.Command) string {
	return guidepkg.RenderReference(Plain(root))
}

// writeReferenceText writes text to path, unless path holds the same text.
func writeReferenceText(path, text string) error {
	data := []byte(text)
	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, data) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Plain turns the Cobra tree into the plain types of internal/guide. It omits
// each hidden command and help. It adds the help flag and the default
// completion and help commands first, so the tree does not depend on which
// command ran.
func Plain(root *cobra.Command) []guidepkg.Command {
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()
	var out []guidepkg.Command
	var visit func(*cobra.Command)
	visit = func(command *cobra.Command) {
		command.InitDefaultHelpFlag()
		if command.Hidden || command.Name() == "help" {
			return
		}
		out = append(out, guidepkg.Command{
			Path:  command.CommandPath(),
			Short: command.Short,
			Flags: flagsOf(command),
		})
		for _, child := range command.Commands() {
			visit(child)
		}
	}
	visit(root)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// flagsOf gives the flags that apply to a command, sorted by name.
func flagsOf(command *cobra.Command) []guidepkg.Flag {
	flags := []guidepkg.Flag{}
	command.Flags().VisitAll(func(flag *pflag.Flag) {
		flags = append(flags, guidepkg.Flag{
			Name:      flag.Name,
			Shorthand: flag.Shorthand,
			Type:      flag.Value.Type(),
			Usage:     flag.Usage,
		})
	})
	sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })
	return flags
}
