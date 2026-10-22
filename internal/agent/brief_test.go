package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBriefPromptsStartWithHiveMarker(t *testing.T) {
	for name, prompt := range map[string]string{
		"worker":   workerPrompt("/tmp/home", "task", "/tmp/worktree", "task/branch", "", false),
		"verifier": verifierPrompt("/tmp/home", "task", ""),
	} {
		t.Run(name, func(t *testing.T) {
			if !strings.HasPrefix(prompt, "⬢ ") {
				t.Fatalf("prompt = %q; want the hive marker prefix", prompt)
			}
		})
	}
}

func TestBriefPrompt(t *testing.T) {
	home := "/tmp/a b'c"
	task := "1-worker-demo"
	path := "/tmp/wt with space"
	branch := "task/1-worker-demo"
	workerWant := `⬢ Read /tmp/a b'c/.agents/skills/hive-worker/SKILL.md first. Then run: hive task show '1-worker-demo' --brief. Follow that brief. Your worktree is '/tmp/wt with space' on the branch 'task/1-worker-demo'. The skill hive-worker is /hive-worker in Claude Code and $hive-worker in Codex.`
	verifierWant := `⬢ Read /tmp/a b'c/.agents/skills/hive-verifier/SKILL.md first. Then run: hive task show '1-worker-demo' --brief. You are the verifier. Run each verification item again, write VERIFIED.md, then run: hive task verify --task '1-worker-demo', or add --fail --reason. The skill hive-verifier is /hive-verifier in Claude Code and $hive-verifier in Codex.`
	workerFeatureWant := `⬢ Read /tmp/a b'c/.agents/skills/hive-worker/SKILL.md first. Then run: hive task show '1-worker-demo' --brief. Follow that brief. Your worktree is '/tmp/wt with space' on the branch 'task/1-worker-demo'. The task brief is .hive/spec/feature-a/tasks/1-worker-demo/brief.md. Write each V<n>.md to .hive/spec/feature-a/tasks/1-worker-demo/evidence/. The skill hive-worker is /hive-worker in Claude Code and $hive-worker in Codex.`
	workerSpecWant := `⬢ Read /tmp/a b'c/.agents/skills/hive-worker/SKILL.md first. Then run: hive task show '1-worker-demo' --brief. Follow that brief. Your worktree is '/tmp/wt with space' on the branch 'task/1-worker-demo'. The task brief is .hive/spec/feature-a/tasks/1-worker-demo/brief.md. Write each V<n>.md to .hive/spec/feature-a/tasks/1-worker-demo/evidence/. When hive worktree create shows allowlist_needed, add !.hive/spec/ and !.hive/spec/** to the project .gitignore after the rules that ignore the path. Commit this change with the feature README. The skill hive-worker is /hive-worker in Claude Code and $hive-worker in Codex.`
	verifierFeatureWant := `⬢ Read /tmp/a b'c/.agents/skills/hive-verifier/SKILL.md first. Then run: hive task show '1-worker-demo' --brief. The task brief is .hive/spec/feature-a/tasks/1-worker-demo/brief.md. You are the verifier. Run each verification item again, write VERIFIED.md to .hive/spec/feature-a/tasks/1-worker-demo/evidence/, then run: hive task verify --task '1-worker-demo', or add --fail --reason. Never write VERIFIED.md at the checkout root. The skill hive-verifier is /hive-verifier in Claude Code and $hive-verifier in Codex.`

	if got := workerPrompt(home, task, path, branch, "", false); got != workerWant {
		t.Fatalf("worker prompt = %q; want %q", got, workerWant)
	}
	if got := verifierPrompt(home, task, ""); got != verifierWant {
		t.Fatalf("verifier prompt = %q; want %q", got, verifierWant)
	}
	if got := workerPrompt(home, task, path, branch, "feature-a", false); got != workerFeatureWant {
		t.Fatalf("feature worker prompt = %q; want %q", got, workerFeatureWant)
	}
	if got := workerPrompt(home, task, path, branch, "feature-a", true); got != workerSpecWant {
		t.Fatalf("spec worker prompt = %q; want %q", got, workerSpecWant)
	}
	if got := verifierPrompt(home, task, "feature-a"); got != verifierFeatureWant {
		t.Fatalf("feature verifier prompt = %q; want %q", got, verifierFeatureWant)
	}
}

func TestBriefPromptShellSafe(t *testing.T) {
	home := "/tmp/home path"
	task := "1-worker-demo"
	path := "/tmp/wt with space"
	branch := "task/1-worker-demo"
	workerWant := `⬢ Read /tmp/home path/.agents/skills/hive-worker/SKILL.md first. Then run: hive task show '1-worker-demo' --brief. Follow that brief. Your worktree is '/tmp/wt with space' on the branch 'task/1-worker-demo'. The skill hive-worker is /hive-worker in Claude Code and $hive-worker in Codex.`
	verifierWant := `⬢ Read /tmp/home path/.agents/skills/hive-verifier/SKILL.md first. Then run: hive task show '1-worker-demo' --brief. You are the verifier. Run each verification item again, write VERIFIED.md, then run: hive task verify --task '1-worker-demo', or add --fail --reason. The skill hive-verifier is /hive-verifier in Claude Code and $hive-verifier in Codex.`

	dir := t.TempDir()
	script := filepath.Join(dir, "hive")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nfor arg in \"$@\"; do printf '%s\\n' \"$arg\"; done\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		prompt string
		want   string
	}{
		{workerPrompt(home, task, path, branch, "", false), workerWant},
		{verifierPrompt(home, task, ""), verifierWant},
	} {
		prompt := testCase.prompt
		if prompt != testCase.want {
			t.Fatalf("prompt = %q; want %q", prompt, testCase.want)
		}
		start := strings.Index(prompt, "hive task show ")
		end := strings.Index(prompt, "--brief")
		if start < 0 || end < 0 {
			t.Fatalf("prompt = %q; want the brief retrieval command", prompt)
		}
		command := prompt[start : end+len("--brief")]
		cmd := exec.Command("sh", "-c", command)
		cmd.Env = append(os.Environ(), "PATH="+dir)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("sh -c %q error = %v", command, err)
		}
		want := strings.Join([]string{"task", "show", task, "--brief"}, "\n") + "\n"
		if string(out) != want {
			t.Fatalf("script args = %q; want %q", out, want)
		}
	}
}
