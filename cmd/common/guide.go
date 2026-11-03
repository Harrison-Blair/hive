package common

import (
	"strings"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GuideOmit is the annotation key that leaves a command line or a flag line
// out of the hive guide. Set it in cobra.Command.Annotations or with
// pflag.FlagSet.SetAnnotation. The item stays in --help.
const GuideOmit = "hive_guide_omit"

// Guide renders the hive guide from the command tree of root: each command
// except help, completion, and the hidden commands, and then the usage notes.
// It leaves out the line of a command annotated with GuideOmit, but keeps its
// children, and each flag annotated with GuideOmit.
func Guide(root *cobra.Command) string {
	var commands []guide.Command
	var walk func(*cobra.Command)
	walk = func(command *cobra.Command) {
		if command.Hidden || command.Name() == "help" || command.Name() == "completion" {
			return
		}
		if _, omit := command.Annotations[GuideOmit]; !omit {
			commands = append(commands, guide.Command{
				Use:   strings.TrimSuffix(command.UseLine(), " [flags]"),
				Short: command.Short,
				Flags: guideFlags(command.LocalNonPersistentFlags()),
			})
		}
		for _, child := range command.Commands() {
			walk(child)
		}
	}
	walk(root)
	return guide.Render(commands, guideFlags(root.PersistentFlags()))
}

func guideFlags(set *pflag.FlagSet) []guide.Flag {
	var flags []guide.Flag
	set.VisitAll(func(flag *pflag.Flag) {
		if _, omit := flag.Annotations[GuideOmit]; omit || flag.Hidden || flag.Name == "help" {
			return
		}
		flags = append(flags, guide.Flag{
			Name:      flag.Name,
			Shorthand: flag.Shorthand,
			Type:      flag.Value.Type(),
			Default:   flag.DefValue,
			Usage:     flag.Usage,
		})
	})
	return flags
}
