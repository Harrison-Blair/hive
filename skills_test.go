// skills_test.go checks that each hive command named in a skill or a template
// exists in the command tree.
package main

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd"
	"github.com/spf13/cobra"
)

// skillFiles lists the skill and template files that name hive commands.
var skillFiles = []string{
	".agents/skills/hive-coordinator/SKILL.md",
	".agents/skills/hive-manager/SKILL.md",
	".agents/skills/hive-worker/SKILL.md",
	".agents/skills/hive-verifier/SKILL.md",
	"internal/agent/templates/project-agents.md",
	"internal/agent/templates/presets.md",
}

// TestSkillCommands checks the fixture, each real file, and the persistent
// flag skip of the command checker.
func TestSkillCommands(t *testing.T) {
	root := cmd.NewRootCmd()

	t.Run("fixture reports each unknown command", func(t *testing.T) {
		const fixture = "hive --home /tmp/home agent fly\nhive --json agent fly\n"
		problems := unknownSkillCommands(root, "fixture.md", fixture)
		if len(problems) != 2 {
			t.Fatalf("got %d problems %v; want 2", len(problems), problems)
		}
		for _, want := range []string{"agent fly", "fixture.md"} {
			if !strings.Contains(problems[0]+problems[1], want) {
				t.Errorf("problems %v do not name %q", problems, want)
			}
		}
	})

	t.Run("each real file names a command", func(t *testing.T) {
		for _, file := range skillFiles {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			for _, problem := range unknownSkillCommands(root, file, string(data)) {
				t.Errorf("%s", problem)
			}
		}
	})

	t.Run("persistent flags are skipped", func(t *testing.T) {
		problems := unknownSkillCommands(root, "x.md", "hive --home /tmp/home task complete --task 1-step\n")
		if len(problems) != 0 {
			t.Fatalf("got %v; want none", problems)
		}
	})
}

// unknownSkillCommands returns one message for each command line that starts
// with hive and names a command path outside root.
func unknownSkillCommands(root *cobra.Command, source, text string) []string {
	var problems []string
	for index, raw := range strings.Split(text, "\n") {
		fields := splitCommandLine(strings.TrimSpace(raw))
		if len(fields) == 0 || fields[0] != "hive" {
			continue
		}
		path, known := resolveCommandPath(root, fields[1:])
		if known {
			continue
		}
		if path == "" {
			path = "hive <command>"
		}
		problems = append(problems, source+":"+strconv.Itoa(index+1)+": unknown command: "+path)
	}
	return problems
}

// resolveCommandPath walks the command children of root over args. It skips
// the persistent flags --home PATH and --json before the command path. It
// returns the known path, and false when a token names no child of a command
// that needs a child.
func resolveCommandPath(root *cobra.Command, args []string) (string, bool) {
	current := root
	path := []string{}
	index := 0
	for index < len(args) {
		token := args[index]
		if token == "--home" {
			index += 2
			continue
		}
		if strings.HasPrefix(token, "--home=") || token == "--json" {
			index++
			continue
		}
		break
	}
	for index < len(args) {
		token := args[index]
		if strings.HasPrefix(token, "-") {
			break
		}
		child := findChild(current, token)
		if child == nil {
			if len(current.Commands()) > 0 && !runnable(current) {
				return strings.Join(append(path, token), " "), false
			}
			break
		}
		path = append(path, child.Name())
		current = child
		index++
	}
	return strings.Join(path, " "), true
}

// findChild returns the direct child of current with the given name or alias.
func findChild(current *cobra.Command, name string) *cobra.Command {
	for _, child := range current.Commands() {
		if child.Name() == name || child.HasAlias(name) {
			return child
		}
	}
	return nil
}

// runnable reports whether the command runs a function of its own.
func runnable(command *cobra.Command) bool {
	return command.Run != nil || command.RunE != nil
}

// splitCommandLine splits one command line into fields. It keeps a single
// quoted value together and removes the quote characters.
func splitCommandLine(line string) []string {
	var fields []string
	var current strings.Builder
	quote := rune(0)
	for _, r := range line {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t':
			if current.Len() > 0 {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}
	return fields
}
