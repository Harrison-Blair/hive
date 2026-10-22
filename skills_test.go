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
	".agents/skills/hive-agent/SKILL.md",
	"internal/agent/templates/project-agents.md",
	"internal/agent/templates/presets.md",
}

// TestSkillCommands checks the fixture, each real file, and the JSON flag skip.
func TestSkillCommands(t *testing.T) {
	root := cmd.NewRootCmd()

	t.Run("fixture reports each unknown command", func(t *testing.T) {
		const fixture = "hive agent fly\nhive --json agent fly\n"
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

	t.Run("JSON flag is skipped", func(t *testing.T) {
		problems := unknownSkillCommands(root, "x.md", "hive --json task complete --task 1-step\n")
		if len(problems) != 0 {
			t.Fatalf("got %v; want none", problems)
		}
	})

	t.Run("an absent root command is unknown", func(t *testing.T) {
		problems := unknownSkillCommands(root, "x.md", "hive absent-command\n")
		if len(problems) != 1 {
			t.Fatalf("got %d problems %v; want 1", len(problems), problems)
		}
		if !strings.Contains(problems[0], "absent-command") {
			t.Errorf("problem %q does not name the absent command", problems[0])
		}
	})

	t.Run("a valid positional argument is accepted", func(t *testing.T) {
		problems := unknownSkillCommands(root, "x.md", "hive project show repo-one\n")
		if len(problems) != 0 {
			t.Fatalf("got %v; want none", problems)
		}
	})
}

// TestCoordinatorPreference checks that the coordinator writes a preference
// record to overrides/ and not to memory/.
func TestCoordinatorPreference(t *testing.T) {
	const file = ".agents/skills/hive-coordinator/SKILL.md"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	section := markdownSection(string(data), "## Preference record")
	if section == "" {
		t.Fatalf("%s has no ## Preference record section", file)
	}
	if !strings.Contains(section, "overrides/") {
		t.Errorf("preference procedure does not name overrides/:\n%s", section)
	}
	if strings.Contains(section, "memory write") {
		t.Errorf("preference procedure writes a memory:\n%s", section)
	}
}

func TestCoordinatorStartupUsesBareHiveTabs(t *testing.T) {
	const skillPath = ".agents/skills/hive-coordinator/SKILL.md"
	skillData, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read %s: %v", skillPath, err)
	}
	startup := markdownSection(string(skillData), "## Startup")
	for _, forbidden := range []string{"herdr tab create --workspace \"$workspace_id\" --cwd \"$PWD\" --label logs", "herdr pane run \"$right_pane_id\" hive logs --follow", "herdr pane split \"$left_pane_id\""} {
		if strings.Contains(startup, forbidden) {
			t.Errorf("coordinator startup still uses the logs tab command %q", forbidden)
		}
	}
	for _, required := range []string{"claim is stale and the `supervisor` tab is gone", "foreground_processes", "[\"hive\", \"logs\", \"--follow\"]", "[\"hive\", \"supervise\"]", "herdr tab close \"$logs_tab_id\"", "Tell the user when you close the old tab."} {
		if !strings.Contains(startup, required) {
			t.Errorf("coordinator startup omits %q", required)
		}
	}

	agents, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	coordinator := markdownSection(string(agents), "## Coordinator")
	if !strings.Contains(coordinator, "Bare `hive` starts the coordinator and the supervisor in their own tabs.") {
		t.Errorf("AGENTS.md coordinator section does not describe bare hive startup:\n%s", coordinator)
	}
}

// markdownSection returns the body of one level-two Markdown section.
func markdownSection(text, heading string) string {
	start := strings.Index(text, heading)
	if start < 0 {
		return ""
	}
	rest := text[start+len(heading):]
	if end := strings.Index(rest, "\n## "); end >= 0 {
		return rest[:end]
	}
	return rest
}

// unknownSkillCommands returns one message for each command line that starts
// with hive and names a command path outside root. A prose sentence is not a
// command line, so the generated-file line of each template is skipped.
func unknownSkillCommands(root *cobra.Command, source, text string) []string {
	var problems []string
	for index, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		fields := splitCommandLine(line)
		if len(fields) == 0 || fields[0] != "hive" {
			continue
		}
		if sentencePunctuation(line) {
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
// the persistent JSON flag before the command path. It returns the known path,
// and false when a token names no child of a command
// that has children. A leaf command accepts the token as a positional
// argument.
func resolveCommandPath(root *cobra.Command, args []string) (string, bool) {
	current := root
	path := []string{}
	index := 0
	for index < len(args) {
		token := args[index]
		if token == "--json" {
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
			if len(current.Commands()) > 0 {
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

// sentencePunctuation reports whether line ends a sentence. The generated-file
// line of each template starts with the word hive and is prose, not a command.
func sentencePunctuation(line string) bool {
	if line == "" {
		return false
	}
	switch line[len(line)-1] {
	case '.', '!', '?':
		return true
	}
	return false
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

// TestWorkerAndVerifierRequireNoStartup checks that each skill tells the agent
// to run bare hive only with the no-startup switch or variable.
func TestWorkerAndVerifierRequireNoStartup(t *testing.T) {
	for _, file := range []string{".agents/skills/hive-worker/SKILL.md", ".agents/skills/hive-verifier/SKILL.md"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, want := range []string{"hive --no-startup", "HIVE_NO_STARTUP=1 hive"} {
			if !strings.Contains(string(data), want) {
				t.Errorf("%s does not contain %q", file, want)
			}
		}
	}
}

// TestGeneralAgentSkill checks that the hive-agent skill states the work, the
// report, and that the agent has no task (decision 23).
func TestGeneralAgentSkill(t *testing.T) {
	data, err := os.ReadFile(".agents/skills/hive-agent/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Work", "## Report", "You have no task.", "hive agent report --file PATH", "hive --no-startup"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("hive-agent skill does not contain %q", want)
		}
	}
}
