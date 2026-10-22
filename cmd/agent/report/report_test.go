package report_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd"
	"github.com/Harrison-Blair/hive/cmd/common"
	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// reportHome makes a home with a live parent boss in pane w1:p1 and a live
// general agent rev in pane w1:p2 whose parent is boss.
func reportHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []records.Agent{
		{Key: "boss", Name: "boss", Role: "agent", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1},
		{Key: "rev", Name: "rev", Role: "agent", Status: records.AgentLive, PaneID: "w1:p2", Parent: "boss", Generation: 1},
	} {
		if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

func runReport(t *testing.T, home, stdin string, args ...string) (int, string, string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HIVE_HOME", home)
	root := cmd.NewRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetArgs(args)
	root.SetIn(strings.NewReader(stdin))
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	status := 0
	if err := root.Execute(); err != nil {
		var statusError *common.StatusError
		if !errors.As(err, &statusError) {
			t.Fatalf("execute error = %v", err)
		}
		status = statusError.Status
	}
	return status, stdout.String(), stderr.String()
}

func bossInbox(t *testing.T, home string) []records.Inbox {
	t.Helper()
	found, err := inbox.All(home, "boss")
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func TestReportReadsStandardInputToTheParentInbox(t *testing.T) {
	home := reportHome(t)
	t.Setenv("HERDR_PANE_ID", "w1:p2")
	status, stdout, stderr := runReport(t, home, "review done", "agent", "report")
	if status != 0 || stdout != "reported to boss as record 1\n" {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want the report line", status, stdout, stderr)
	}
	found := bossInbox(t, home)
	if len(found) != 1 || found[0].Kind != "report" || found[0].Text != "review done" || found[0].Agent != "rev" {
		t.Fatalf("boss inbox = %#v; want one report record", found)
	}
}

func TestReportJSONAndFile(t *testing.T) {
	home := reportHome(t)
	t.Setenv("HERDR_PANE_ID", "w1:p2")
	file := filepath.Join(t.TempDir(), "result.md")
	if err := os.WriteFile(file, []byte("full result"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, stdout, stderr := runReport(t, home, "", "--json", "agent", "report", "--file", file)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr)
	}
	var response struct {
		Outcome string `json:"outcome"`
		Result  struct {
			Inbox  string        `json:"inbox"`
			Record records.Inbox `json:"record"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Fatalf("decode %q: %v", stdout, err)
	}
	if response.Outcome != "ok" || response.Result.Inbox != "boss" || response.Result.Record.Path != file || response.Result.Record.Text != "full result" {
		t.Fatalf("response = %#v; want the file report", response)
	}
}

func TestReportCallerOverrides(t *testing.T) {
	for _, test := range []struct {
		name  string
		args  []string
		error string
	}{
		{name: "agent override", args: []string{"--agent", "boss"}, error: "--agent boss is not the caller rev"},
		{name: "pane override", args: []string{"--pane", "w1:p1"}, error: "--pane w1:p1 is not the caller pane w1:p2"},
		{name: "own agent", args: []string{"--agent", "rev"}},
		{name: "own pane", args: []string{"--pane", "w1:p2"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := reportHome(t)
			t.Setenv("HERDR_PANE_ID", "w1:p2")
			status, _, stderr := runReport(t, home, "r", append([]string{"agent", "report"}, test.args...)...)
			found := bossInbox(t, home)
			if test.error == "" {
				if status != 0 || len(found) != 1 || found[0].Agent != "rev" {
					t.Fatalf("status = %d, stderr = %q, inbox = %#v; want one record from rev", status, stderr, found)
				}
				return
			}
			if status == 0 || !strings.Contains(stderr, test.error) || len(found) != 0 {
				t.Fatalf("status = %d, stderr = %q, inbox = %#v; want %q", status, stderr, found, test.error)
			}
		})
	}
}

func TestReportOutsideHerdrUsesTheAgentFlag(t *testing.T) {
	home := reportHome(t)
	t.Setenv("HERDR_PANE_ID", "")
	status, _, stderr := runReport(t, home, "r", "agent", "report", "--agent", "rev")
	if found := bossInbox(t, home); status != 0 || len(found) != 1 || found[0].Agent != "rev" {
		t.Fatalf("status = %d, stderr = %q, inbox = %#v; want one record from rev", status, stderr, found)
	}
}
