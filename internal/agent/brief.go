package agent

import (
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/marker"
)

// workerPrompt is the first prompt of one worker. It is one line.
func workerPrompt(home, task, path, branch, feature string, spec bool) string {
	prompt := marker.Prefix + "Read " + home + "/.agents/skills/hive-worker/SKILL.md first. Then run: " +
		"hive task show " + shellQuote(task) + " --brief. " +
		"Follow that brief. Your worktree is " + shellQuote(path) + " on the branch " + shellQuote(branch) + ". "
	if feature != "" {
		brief, evidence := featurePromptPaths(task, feature)
		prompt += "The task brief is " + brief + ". Write each V<n>.md to " + evidence + ". "
	}
	if spec {
		prompt += "Write the feature README at .hive/spec/" + feature + "/README.md in the worktree. " +
			"Do not commit a file under .hive/spec/. Hive copies it to the home. "
	}
	return prompt + "The skill hive-worker is /hive-worker in Claude Code and $hive-worker in Codex."
}

// verifierPrompt is the first prompt of one verifier. It is one line.
func verifierPrompt(home, task, feature string) string {
	prompt := marker.Prefix + "Read " + home + "/.agents/skills/hive-verifier/SKILL.md first. Then run: " +
		"hive task show " + shellQuote(task) + " --brief. "
	if feature != "" {
		brief, _ := featurePromptPaths(task, feature)
		prompt += "The task brief is " + brief + ". "
	}
	prompt += "You are the verifier. Run each verification item again, write VERIFIED.md"
	if feature != "" {
		_, evidence := featurePromptPaths(task, feature)
		prompt += " to " + evidence
	}
	prompt += ", then run: hive task verify --task " + shellQuote(task) + ", or add --fail --reason. "
	if feature != "" {
		prompt += "Never write VERIFIED.md at the checkout root. "
	}
	return prompt + "The skill hive-verifier is /hive-verifier in Claude Code and $hive-verifier in Codex."
}

func featurePromptPaths(task, feature string) (string, string) {
	root := ".hive/spec/" + feature + "/tasks/" + task
	return root + "/brief.md", root + "/evidence/"
}

// shellQuote wraps one value in single quotes for the POSIX shell. Each single
// quote in the value is replaced by the four characters '\”.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
