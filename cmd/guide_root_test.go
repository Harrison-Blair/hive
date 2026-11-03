package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func TestRootGuideNamesEachCommandAndFlag(t *testing.T) {
	testHome(t)
	var stdout, stderr bytes.Buffer
	if status := executeWithStreams([]string{"guide"}, strings.NewReader(""), &stdout, &stderr); status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	text := stdout.String()
	for _, command := range []string{
		"hive agent list", "hive agent send NAME TEXT", "hive agent spawn", "hive agent stop NAME",
		"hive logs", "hive worktree create BRANCH", "hive worktree remove BRANCH",
	} {
		if !strings.Contains(text, "\n"+command+": ") {
			t.Errorf("guide has no command %q", command)
		}
	}
	for _, flag := range []string{
		"  --json bool: write JSON output",
		"  --harness string: harness: claude, codex, or pi (required)",
		"  --prompt-file string: file with the first prompt text",
		"  --branch string: start in the checkout of the branch; make one if none",
		"  --directory string: start in this existing directory",
		"  --base string: start commit of a new branch (default: HEAD)",
		"  --wait bool: wait until the agent is idle or done again",
		"  --agent string: agent name: matches agent_id or target",
		"  --since string: duration or RFC 3339 cutoff",
	} {
		if !strings.Contains(text, flag+"\n") {
			t.Errorf("guide has no flag line %q", flag)
		}
	}
	for _, absent := range []string{
		"\nhive completion", "\nhive help", "\nhive project", "--project", "--worktree", "--remove-worktree",
		"\nhive: ", "\nhive agent: ", "\nhive agent graph", "\nhive guide", "\nhive version", "\nhive worktree: ",
		"--component", "--level", "--request", "--follow", "--limit",
	} {
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

func TestRootHelpKeepsTheCommandsAndFlagsLeftOutOfTheGuide(t *testing.T) {
	testHome(t)
	for _, test := range []struct {
		args []string
		want []string
	}{
		{[]string{"--help"}, []string{"pick a harness and a model", "\n  agent ", "\n  guide ", "\n  version ", "\n  worktree "}},
		{[]string{"agent", "--help"}, []string{"manage agents in herdr panes", "\n  graph "}},
		{[]string{"agent", "graph", "--help"}, []string{"show a live graph of the agents"}},
		{[]string{"guide", "--help"}, []string{"print the hive guide"}},
		{[]string{"version", "--help"}, []string{"Print the hive version"}},
		{[]string{"worktree", "--help"}, []string{"manage the branch checkouts"}},
		{[]string{"logs", "--help"}, []string{"--component string", "--level string", "--request string", "--follow ", "--limit int"}},
	} {
		var stdout, stderr bytes.Buffer
		if status := executeWithStreams(test.args, strings.NewReader(""), &stdout, &stderr); status != 0 {
			t.Fatalf("%v: status = %d, stderr = %q; want 0", test.args, status, stderr.String())
		}
		for _, want := range test.want {
			if !strings.Contains(stdout.String(), want) {
				t.Errorf("%v: help has no %q in:\n%s", test.args, want, stdout.String())
			}
		}
	}
}

func TestRootGuideDoesNotCreateAMissingHome(t *testing.T) {
	userHome := t.TempDir()
	missing := filepath.Join(t.TempDir(), "missing")
	for _, test := range []struct {
		name, home, hiveHome, created string
	}{
		{name: "default home", home: userHome, created: filepath.Join(userHome, ".hive")},
		{name: "HIVE_HOME", home: userHome, hiveHome: missing, created: missing},
		{name: "no user home"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("HOME", test.home)
			t.Setenv("HIVE_HOME", test.hiveHome)
			for _, args := range [][]string{{"guide"}, {"--json", "guide"}} {
				var stdout, stderr bytes.Buffer
				if status := executeWithStreams(args, strings.NewReader(""), &stdout, &stderr); status != 0 || stderr.Len() != 0 {
					t.Fatalf("%v: status = %d, stderr = %q; want 0 and empty stderr", args, status, stderr.String())
				}
				if !strings.Contains(stdout.String(), "hive agent spawn: start an agent") {
					t.Fatalf("%v: stdout = %q; want the guide", args, stdout.String())
				}
			}
			if test.created != "" {
				if _, err := os.Stat(test.created); !os.IsNotExist(err) {
					t.Fatalf("stat %s: %v; want the home not created", test.created, err)
				}
			}
		})
	}
	var stdout bytes.Buffer
	_ = executeWithStreams([]string{"--json", "guide"}, strings.NewReader(""), &stdout, &stdout)
	var envelope struct {
		Outcome string `json:"outcome"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil || envelope.Outcome != "ok" {
		t.Fatalf("--json stdout = %q; want one ok envelope", stdout.String())
	}
}

func TestRootGuideLogsInAHome(t *testing.T) {
	home := testHome(t)
	var stdout, stderr bytes.Buffer
	if status := executeWithStreams([]string{"guide"}, strings.NewReader(""), &stdout, &stderr); status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	events := agenttest.CommandEvents(t, home, "guide")
	if len(events) != 2 || events[0]["msg"] != "command start" || events[1]["msg"] != "command end" || events[1]["outcome"] != "ok" {
		t.Fatalf("events = %v; want command start and an ok command end", events)
	}
}
