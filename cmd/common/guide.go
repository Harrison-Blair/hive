package common

import (
	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Guide renders the hive guide from the command tree of root: each command
// except help, completion, and the hidden commands, and then the usage notes.
func Guide(root *cobra.Command) string {
	var commands []guide.Command
	var walk func(*cobra.Command)
	walk = func(command *cobra.Command) {
		if command.Hidden || command.Name() == "help" || command.Name() == "completion" {
			return
		}
		commands = append(commands, guide.Command{
			Path:  command.CommandPath(),
			Use:   command.UseLine(),
			Short: command.Short,
			Flags: guideFlags(command.LocalNonPersistentFlags()),
		})
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
		if flag.Hidden || flag.Name == "help" {
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
