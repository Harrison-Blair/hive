package agent

import "strings"

// workerPrompt is the first prompt of one worker. It is one line.
func workerPrompt(home, task, path, branch string) string {
	return "Read " + home + "/.agents/skills/hive-worker/SKILL.md first. Then run: " +
		"hive --home " + shellQuote(home) + " task show " + shellQuote(task) + " --brief. " +
		"Follow that brief. Your worktree is " + shellQuote(path) + " on the branch " + shellQuote(branch) + ". " +
		"The skill hive-worker is /hive-worker in Claude Code and $hive-worker in Codex."
}

// verifierPrompt is the first prompt of one verifier. It is one line.
func verifierPrompt(home, task string) string {
	return "Read " + home + "/.agents/skills/hive-verifier/SKILL.md first. Then run: " +
		"hive --home " + shellQuote(home) + " task show " + shellQuote(task) + " --brief. " +
		"You are the verifier. Run each verification item again, write VERIFIED.md, then run: " +
		"hive --home " + shellQuote(home) + " task verify --task " + shellQuote(task) + ", or add --fail --reason. " +
		"The skill hive-verifier is /hive-verifier in Claude Code and $hive-verifier in Codex."
}

// shellQuote wraps one value in single quotes for the POSIX shell. Each single
// quote in the value is replaced by the four characters '\”.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
