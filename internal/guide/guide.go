// Package guide holds the hive guide text: the hand-written flows and the
// generated command reference.
package guide

import (
	_ "embed"
	"fmt"
	"slices"
	"sort"
	"strings"
)

//go:embed flows.md
var flows string

//go:embed reference.md
var reference string

// MaxBytes is the upper bound of the full guide text. Hive keeps the guide
// below it so that a first prompt stays below 64 KiB.
const MaxBytes = 24 * 1024

// Command is one visible command of the Cobra tree.
type Command struct {
	Path  string
	Short string
	Flags []Flag
}

// Flag is one flag of a command.
type Flag struct {
	Name      string
	Shorthand string
	Type      string
	Usage     string
}

// Flows returns the hand-written part of the guide.
func Flows() string { return flows }

// Reference returns the generated reference part.
func Reference() string { return reference }

// Text returns the two parts, flows first.
func Text() string { return flows + "\n" + reference }

// RenderReference writes the reference text of commands. It sorts each command
// by path and each flag by name, so the text is deterministic.
func RenderReference(commands []Command) string {
	sorted := slices.Clone(commands)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
	var b strings.Builder
	b.WriteString("# hive reference\n\n")
	b.WriteString("The command tree of hive. Each command line gives the path and the short\n")
	b.WriteString("help. Each flag line below it gives the name, the type, and the usage.\n\n")
	for _, command := range sorted {
		fmt.Fprintf(&b, "%s: %s\n", command.Path, command.Short)
		flags := slices.Clone(command.Flags)
		sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })
		for _, flag := range flags {
			name := "--" + flag.Name
			if flag.Shorthand != "" {
				name = "-" + flag.Shorthand + ", " + name
			}
			fmt.Fprintf(&b, "  %s %s: %s\n", name, flag.Type, flag.Usage)
		}
		b.WriteString("\n")
	}
	return b.String()
}
