package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRootGuideNamesEachCommandAndFlag(t *testing.T) {
	testHome(t)
	var stdout, stderr bytes.Buffer
	if status := executeWithStreams([]string{"guide"}, strings.NewReader(""), &stdout, &stderr); status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	text := stdout.String()
	for _, command := range []string{
		"hive", "hive agent", "hive agent list", "hive agent send", "hive agent spawn", "hive agent stop",
		"hive guide", "hive init", "hive logs", "hive project", "hive project list", "hive project register", "hive version",
	} {
		if !strings.Contains(text, "\n### "+command+"\n") {
			t.Errorf("guide has no command %q", command)
		}
	}
	for _, flag := range []string{
		"- `--json` (bool): write JSON output",
		"- `--harness` (string): harness: claude, codex, or pi (required)",
		"- `--prompt-file` (string): file with the first prompt text",
		"- `--worktree` (string): worktree and branch name: start in a new worktree of --project",
		"- `--wait` (bool): wait until the agent is idle or done again",
		"- `--remove-worktree` (bool): also remove the clean worktree checkout; the branch stays",
		"- `--limit` (int, default 200): maximum matches to print",
		"- `--agent` (string): agent name: matches agent_id or target",
	} {
		if !strings.Contains(text, flag+"\n") {
			t.Errorf("guide has no flag line %q", flag)
		}
	}
	for _, absent := range []string{"### hive completion", "### hive help"} {
		if strings.Contains(text, absent) {
			t.Errorf("guide has %q", absent)
		}
	}
	if !strings.Contains(text, "## Usage notes") {
		t.Errorf("guide has no usage notes")
	}

	stdout.Reset()
	if status := executeWithStreams([]string{"--json", "guide"}, strings.NewReader(""), &stdout, &stderr); status != 0 {
		t.Fatalf("--json status = %d, stderr = %q; want 0", status, stderr.String())
	}
	var envelope struct {
		Outcome string `json:"outcome"`
		Result  struct {
			Text string `json:"text"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v; output %q", err, stdout.String())
	}
	if envelope.Outcome != "ok" || envelope.Result.Text != text {
		t.Fatalf("envelope = %+v; want ok with the text of hive guide", envelope)
	}
}
