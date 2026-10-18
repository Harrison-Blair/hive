package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBriefPrompt(t *testing.T) {
	home := "/tmp/a b'c"
	task := "1-worker-demo"
	path := "/tmp/wt with space"
	branch := "task/1-worker-demo"

	cases := []struct {
		name     string
		prompt   string
		skill    string
		contains []string
	}{
		{
			name:   "worker",
			prompt: workerPrompt(home, task, path, branch),
			skill:  "hive-worker",
			contains: []string{
				home, task, path, branch,
				"/hive-worker in Claude Code", "$hive-worker in Codex",
			},
		},
		{
			name:   "verifier",
			prompt: verifierPrompt(home, task),
			skill:  "hive-verifier",
			contains: []string{
				home, task,
				"/hive-verifier in Claude Code", "$hive-verifier in Codex",
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if strings.Contains(testCase.prompt, "\n") {
				t.Fatalf("prompt is not one line: %q", testCase.prompt)
			}
			prefix := "Read " + home + "/.agents/skills/" + testCase.skill + "/SKILL.md first. Then run: "
			if !strings.HasPrefix(testCase.prompt, prefix) {
				t.Fatalf("prompt = %q; want the prefix %q", testCase.prompt, prefix)
			}
			for _, want := range testCase.contains {
				if !strings.Contains(testCase.prompt, want) {
					t.Fatalf("prompt = %q; want %q", testCase.prompt, want)
				}
			}
			if !strings.Contains(testCase.prompt, "hive --home "+shellQuote(home)+" task show "+shellQuote(task)+" --brief") {
				t.Fatalf("prompt = %q; want the brief retrieval command", testCase.prompt)
			}
		})
	}
}

func TestBriefPromptShellSafe(t *testing.T) {
	home := filepath.Join(t.TempDir(), "a b'c")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	task := "1-worker-demo"
	path := "/tmp/wt with space"
	branch := "task/1-worker-demo"

	dir := t.TempDir()
	script := filepath.Join(dir, "hive")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nfor arg in \"$@\"; do printf '%s\\n' \"$arg\"; done\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	for _, prompt := range []string{
		workerPrompt(home, task, path, branch),
		verifierPrompt(home, task),
	} {
		start := strings.Index(prompt, "hive --home ")
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
		want := strings.Join([]string{"--home", home, "task", "show", task, "--brief"}, "\n") + "\n"
		if string(out) != want {
			t.Fatalf("script args = %q; want %q", out, want)
		}
	}
}
