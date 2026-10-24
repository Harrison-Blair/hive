// Package list_test checks the agent list command.
package list_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func listHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []map[string]any{
		{
			"key": "builder-1", "name": "builder-1", "role": "worker", "profile": "builder",
			"project": "demo", "task": "1-demo", "worktree": "/tmp/1-demo",
			"status": "live", "generation": 1, "pane_id": "w1:p1",
		},
		{
			"key": "planner-1", "name": "planner-1", "role": "manager", "profile": "planner",
			"project": "demo", "task": "", "owns": "demo",
			"status": "live", "generation": 1, "pane_id": "w1:p2",
		},
	} {
		if err := registry.WriteIn([]string{"agents"}, record["key"].(string), record); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

// TestAgentListFiltersByProfile covers AC5.
func TestAgentListFiltersByProfile(t *testing.T) {
	home := listHome(t)
	t.Setenv("HIVE_HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out bytes.Buffer
	if err := cmd.ExecuteWithArgs([]string{"agent", "list", "--profile", "builder"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("agent list --profile builder = %v", err)
	}
	if !strings.Contains(out.String(), "builder-1") {
		t.Fatalf("agent list = %q; want builder-1", out.String())
	}
	if strings.Contains(out.String(), "planner-1") {
		t.Fatalf("agent list = %q; want no planner-1", out.String())
	}
}
