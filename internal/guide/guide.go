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
	// Use is the usage line with no flags part, for example
	// "hive agent send NAME TEXT".
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

// Render returns the guide as dense Markdown: the global flags, one line for
// each command with one indented line for each of its flags, and then the
// usage notes. The guide goes on one start command line, which herdr types
// into a shell, so each byte counts.
func Render(commands []Command, global []Flag) string {
	var b strings.Builder
	b.WriteString("# hive guide\n## Commands\nEach command takes -h/--help. Global flags:\n")
	writeFlags(&b, global)
	for _, command := range commands {
		fmt.Fprintf(&b, "%s: %s\n", command.Use, command.Short)
		writeFlags(&b, command.Flags)
	}
	b.WriteString(notes)
	return b.String()
}

func writeFlags(b *strings.Builder, flags []Flag) {
	for _, flag := range flags {
		name := "--" + flag.Name
		if flag.Shorthand != "" {
			name = "-" + flag.Shorthand + "/" + name
		}
		kind := flag.Type
		if flag.Default != "" && flag.Default != "false" {
			kind += "=" + flag.Default
		}
		fmt.Fprintf(b, "  %s %s: %s\n", name, kind, flag.Usage)
	}
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

// Placeholder returns the text "<guide N bytes>" that a log event shows in
// place of the guide argument value.
func Placeholder(value string) string {
	return fmt.Sprintf("<guide %d bytes>", len(value))
}

// Redact returns a copy of argv with the last value of guideArgs replaced by
// its Placeholder, so that a log line does not carry the guide.
func Redact(argv, guideArgs []string) []string {
	redacted := append([]string(nil), argv...)
	if len(guideArgs) == 0 {
		return redacted
	}
	value := guideArgs[len(guideArgs)-1]
	for i, arg := range redacted {
		if arg == value {
			redacted[i] = Placeholder(value)
		}
	}
	return redacted
}
