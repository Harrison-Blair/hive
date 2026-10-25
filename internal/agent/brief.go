package agent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/guide"
	"github.com/Harrison-Blair/hive/internal/lib/harness"
	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// firstPrompt returns the first prompt of one spawned agent (D3). It gives the
// facts of the record first, then the spawner text.
func firstPrompt(record records.Agent, task *records.Task, spawner string) string {
	return firstPromptWithInstructions(record, task, spawner, "", false)
}

// firstPromptWithInstructions builds the first prompt. When the harness takes
// the profile and the guide in the first prompt, that text goes before the
// facts.
func firstPromptWithInstructions(record records.Agent, task *records.Task, spawner, instructions string, instructionsInPrompt bool) string {
	facts := []string{"You are " + record.Name + "."}
	if record.Parent != "" {
		facts = append(facts, "Your parent is "+record.Parent+".")
	}
	if record.Project != "" {
		facts = append(facts, "Your project is "+record.Project+".")
	}
	if record.Task != "" {
		facts = append(facts, "Your task is "+record.Task+".")
	}
	if record.Worktree != "" {
		facts = append(facts, "Your worktree is "+record.Worktree+".")
	}
	if task != nil {
		brief, evidence := taskPromptPaths(*task)
		if brief != "" {
			facts = append(facts, "Your brief is "+brief+".")
		}
		if evidence != "" {
			facts = append(facts, "Your evidence is "+evidence+".")
		}
	}
	body := strings.Join(facts, " ")
	if instructionsInPrompt && instructions != "" {
		body = instructions + "\n\n" + body
	}
	text := marker.Prefix + body
	if strings.TrimSpace(spawner) != "" {
		text += "\n\n" + spawner
	}
	return text
}

// instructionsText returns the text that an agent reads before its first turn:
// the profile, then the guide, then the profile path and the guide command
// (D4). It returns the empty string when the home has no profile file.
func instructionsText(home, profile string) (string, string) {
	if profile == "" {
		return "", ""
	}
	data, err := os.ReadFile(filepath.Join(home, "profiles", profile+".md"))
	if errors.Is(err, os.ErrNotExist) {
		return "", ""
	}
	if err != nil {
		return "", "profile: " + err.Error()
	}
	parts := []string{strings.TrimSpace(string(data)), strings.TrimSpace(guide.Text())}
	parts = append(parts, fmt.Sprintf("Read `profiles/%s.md` and run `hive guide` after a context reset.", profile))
	return strings.Join(parts, "\n\n"), ""
}

// systemPromptArgs returns the start arguments that carry the instructions.
// It returns nil when the harness takes the instructions in the first prompt.
func systemPromptArgs(profile harness.Profile, instructions string) []string {
	flag := profile.SystemPromptArg()
	if flag == "" || instructions == "" {
		return nil
	}
	return []string{flag, instructions}
}

// taskPromptPaths gives the brief and evidence paths of one task. The evidence
// path is empty for a task with no feature, because its brief names it.
func taskPromptPaths(task records.Task) (string, string) {
	if task.Feature == "" {
		return ".hive/state/briefs/" + task.ID + ".md", ""
	}
	root := ".hive/spec/" + task.Feature + "/tasks/" + task.ID
	return root + "/brief.md", root + "/evidence/"
}

// shellQuote wraps one value in single quotes for the POSIX shell. Each single
// quote in the value is replaced by the four characters '\”.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
