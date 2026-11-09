package common

import (
	"strings"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Guide renders the full public command tree and the usage notes.
func Guide(root *cobra.Command) string {
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()
	var commands []guide.Command
	var walk func(*cobra.Command)
	walk = func(command *cobra.Command) {
		if command.Hidden {
			return
		}
		commands = append(commands, guideCommand(command))
		for _, child := range command.Commands() {
			walk(child)
		}
	}
	walk(root)
	return guide.Render(commands, guideFlags(root.PersistentFlags()))
}

// StartupGuide renders the four most-used agent commands and usage notes.
func StartupGuide(root *cobra.Command) string {
	var commands []guide.Command
	for _, name := range []string{"list", "send", "spawn", "stop"} {
		command, _, err := root.Find([]string{"agent", name})
		if err != nil || command.Name() != name {
			panic("missing agent command: " + name)
		}
		entry := guideCommand(command)
		inherited := pflag.NewFlagSet("inherited", pflag.ContinueOnError)
		command.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
			if root.PersistentFlags().Lookup(flag.Name) != flag {
				inherited.AddFlag(flag)
			}
		})
		entry.Flags = append(entry.Flags, guideFlags(inherited)...)
		commands = append(commands, entry)
	}
	return guide.Render(commands, guideFlags(root.PersistentFlags()))
}

func guideCommand(command *cobra.Command) guide.Command {
	flags := command.LocalFlags()
	if command.Parent() == nil {
		flags = command.LocalNonPersistentFlags()
	}
	return guide.Command{
		Use:   strings.TrimSuffix(command.UseLine(), " [flags]"),
		Short: command.Short,
		Flags: guideFlags(flags),
	}
}

func guideFlags(set *pflag.FlagSet) []guide.Flag {
	var flags []guide.Flag
	set.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden || flag.Name == "help" {
			return
		}
		flags = append(flags, guide.Flag{
			Name: flag.Name, Shorthand: flag.Shorthand, Type: flag.Value.Type(),
			Default: flag.DefValue, Usage: flag.Usage,
		})
	})
	return flags
}
