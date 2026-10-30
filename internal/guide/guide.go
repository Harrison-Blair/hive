// Package guide renders the hive command guide: the command reference and the
// usage notes. Hive gives the guide to each agent in its system prompt, and
// hive guide prints it.
package guide

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed usage.md
var notes string

// Command is one command of the guide.
type Command struct {
	// Path is the full command name, for example "hive agent spawn".
	Path string
	// Use is the usage line.
	Use   string
	Short string
	Flags []Flag
}

// Flag is one flag of a command.
type Flag struct {
	Name      string
	Shorthand string
	Type      string
	Default   string
	Usage     string
}

// Result is the result of hive guide.
type Result struct {
	Text string `json:"text"`
}

// Notes returns the usage notes.
func Notes() string { return notes }

// Render returns the guide as Markdown: the global flags, each command, and
// then the usage notes.
func Render(commands []Command, global []Flag) string {
	var b strings.Builder
	b.WriteString("# hive guide\n\n## Commands\n\nEach command also takes `-h, --help`.\n\nGlobal flags:\n\n")
	writeFlags(&b, global)
	for _, command := range commands {
		fmt.Fprintf(&b, "### %s\n\n`%s`\n\n%s\n\n", command.Path, command.Use, command.Short)
		writeFlags(&b, command.Flags)
	}
	b.WriteString(notes)
	return b.String()
}

func writeFlags(b *strings.Builder, flags []Flag) {
	if len(flags) == 0 {
		return
	}
	for _, flag := range flags {
		name := "--" + flag.Name
		if flag.Shorthand != "" {
			name = "-" + flag.Shorthand + ", " + name
		}
		kind := flag.Type
		if flag.Default != "" && flag.Default != "false" {
			kind += ", default " + flag.Default
		}
		fmt.Fprintf(b, "- `%s` (%s): %s\n", name, kind, flag.Usage)
	}
	b.WriteString("\n")
}

// OneLine replaces each newline with the two characters \n. Herdr rejects a
// start argument with a newline.
func OneLine(text string) string {
	return strings.ReplaceAll(text, "\n", `\n`)
}

// TOMLString returns text as one TOML basic string on one line.
func TOMLString(text string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(text) + `"`
}

// Redact returns a copy of argv with the last value of guideArgs replaced by
// "<guide N bytes>", so that a log line does not carry the guide.
func Redact(argv, guideArgs []string) []string {
	redacted := append([]string(nil), argv...)
	if len(guideArgs) == 0 {
		return redacted
	}
	value := guideArgs[len(guideArgs)-1]
	for i, arg := range redacted {
		if arg == value {
			redacted[i] = fmt.Sprintf("<guide %d bytes>", len(value))
		}
	}
	return redacted
}
