package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

const askCommandJSON = `{"questions":[{"question":"Restart now?","header":"Model","options":[{"label":"Yes","description":"stop it"},{"label":"No","description":"no change"}]}]}`

func askCommandFiles(t *testing.T, home string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(home, ".hive", "state", "asks", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestAskReadsStdinAndPrintsTheID(t *testing.T) {
	home := newRootTestHome(t)
	seedChatCommandAgent(t, home, records.Agent{Key: "worker-one", Name: "worker-one", Role: "worker", PaneID: "w2:p1", Status: records.AgentLive, Generation: 1})
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"ask", "--agent", "worker-one", "--file", "-"}, strings.NewReader(askCommandJSON), &stdout, &stderr)
	if status != 0 || stdout.String() != "ask a1 open\n" {
		t.Fatalf("status = %d, stdout = %q, stderr = %q", status, stdout.String(), stderr.String())
	}
	if files := askCommandFiles(t, home); len(files) != 1 {
		t.Fatalf("ask files = %v; want one", files)
	}
}

func TestAskReadsAFileAndPrintsTheEnvelope(t *testing.T) {
	home := newRootTestHome(t)
	seedChatCommandAgent(t, home, records.Agent{Key: "worker-one", Name: "worker-one", Role: "worker", PaneID: "w2:p1", Status: records.AgentLive, Generation: 1})
	path := filepath.Join(t.TempDir(), "ask.json")
	if err := os.WriteFile(path, []byte(askCommandJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--json", "ask", "--agent", "worker-one", "--file", path}, strings.NewReader(""), &stdout, &stderr)
	var envelope struct {
		Outcome string `json:"outcome"`
		Result  struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode %q: %v", stdout.String(), err)
	}
	if status != 0 || envelope.Outcome != "ok" || envelope.Result.ID != "a1" || envelope.Result.State != "open" {
		t.Fatalf("status = %d, envelope = %+v", status, envelope)
	}
}

func TestAskRejectsABadAskWithNoRecord(t *testing.T) {
	home := newRootTestHome(t)
	seedChatCommandAgent(t, home, records.Agent{Key: "worker-one", Name: "worker-one", Role: "worker", PaneID: "w2:p1", Status: records.AgentLive, Generation: 1})
	var stdout, stderr bytes.Buffer
	bad := strings.Replace(askCommandJSON, `"Model"`, `"A long header"`, 1)
	status := executeWithStreams([]string{"ask", "--agent", "worker-one", "--file", "-"}, strings.NewReader(bad), &stdout, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), "header of 12 characters at most") {
		t.Fatalf("status = %d, stderr = %q; want the header error", status, stderr.String())
	}
	if files := askCommandFiles(t, home); len(files) != 0 {
		t.Fatalf("ask files = %v; want none", files)
	}
}

func TestAskNeedsTheFileFlag(t *testing.T) {
	newRootTestHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"ask", "--agent", "worker-one"}, strings.NewReader(""), &stdout, &stderr)
	if status != 2 || !strings.Contains(stderr.String(), "--file") {
		t.Fatalf("status = %d, stderr = %q; want a usage error that names --file", status, stderr.String())
	}
}
