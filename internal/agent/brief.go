package agent

import (
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/marker"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// firstPrompt returns the first prompt of one spawned agent (D3). It gives the
// facts of the record first, then the spawner text.
func firstPrompt(record records.Agent, task *records.Task, spawner string) string {
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
	text := marker.Prefix + strings.Join(facts, " ")
	if strings.TrimSpace(spawner) != "" {
		text += "\n\n" + spawner
	}
	return text
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
