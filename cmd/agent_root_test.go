package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestAgentRegisterUsesHerdrEnvironment(t *testing.T) {
	home := testHome(t)
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	t.Setenv("HERDR_WORKSPACE_ID", "w1")
	t.Setenv("HERDR_TAB_ID", "t1")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "agent", "register", "--role", "coordinator"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, "coordinator-g1", &record); err != nil {
		t.Fatal(err)
	}
	if record.PaneID != "w1:p1" || record.WorkspaceID != "w1" || record.TabID != "t1" {
		t.Fatalf("record = %#v; want the environment pane values", record)
	}
}

func TestAgentRegisterExplicitPaneLeavesWorkspaceAndTabEmpty(t *testing.T) {
	home := testHome(t)
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	t.Setenv("HERDR_WORKSPACE_ID", "w1")
	t.Setenv("HERDR_TAB_ID", "t1")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "agent", "register", "--role", "coordinator", "--pane", "w2:p2"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, "coordinator-g1", &record); err != nil {
		t.Fatal(err)
	}
	if record.PaneID != "w2:p2" || record.WorkspaceID != "" || record.TabID != "" {
		t.Fatalf("record = %#v; want explicit pane only", record)
	}
}

func TestAgentRegisterRejectsUnsafeHerdrPaneValues(t *testing.T) {
	for _, pane := range []string{"", "bad\npane", "bad\x01pane", strings.Repeat("p", 600)} {
		t.Run(strings.ReplaceAll(pane, "\n", "\\n"), func(t *testing.T) {
			home := testHome(t)
			t.Setenv("HERDR_PANE_ID", pane)
			var stdout, stderr bytes.Buffer
			status := executeWithStreams([]string{"--home", home, "agent", "register", "--role", "coordinator"}, strings.NewReader(""), &stdout, &stderr)
			if status != 1 {
				t.Fatalf("status = %d, stderr = %q; want 1", status, stderr.String())
			}
			if strings.Count(stderr.String(), "\n") != 1 || !strings.HasPrefix(stderr.String(), "hive: ") {
				t.Fatalf("stderr = %q; want one hive error line", stderr.String())
			}
			registry, err := store.Open(home)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := registry.ListIn([]string{"agents"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(raw) != 0 {
				t.Fatalf("unsafe pane wrote %d agent records", len(raw))
			}
		})
	}
}
