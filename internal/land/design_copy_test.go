package land

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC2: land never copies a design file of another feature back to the home,
// even when the worktree changed it or added one.
func TestLandDoesNotCopyDesignOfOtherFeatureBack(t *testing.T) {
	_, home, worktree, task := specLandFixture(t)
	writeSpec(t, home, ".hive/spec/g/design/chat.grid", "g grid v0\n")
	writeSpec(t, worktree, ".hive/spec/g/design/chat.grid", "g grid from the task\n")
	writeSpec(t, worktree, ".hive/spec/g/design/new.grid", "new grid\n")
	writeSpec(t, worktree, ".hive/spec/f/README.md", "readme v1\n")

	envelope := Land(context.Background(), home, Options{Task: task.ID, Mode: "merge", Grant: "demo"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Land() = %+v, want ok", envelope)
	}
	if got := readSpec(t, home, ".hive/spec/g/design/chat.grid"); got != "g grid v0\n" {
		t.Errorf("home g design = %q, want it unchanged", got)
	}
	if _, err := os.Lstat(filepath.Join(home, ".hive", "spec", "g", "design", "new.grid")); !os.IsNotExist(err) {
		t.Errorf("home has the new g design file (%v); want no copy", err)
	}
	for _, effect := range envelope.Effects {
		if strings.HasPrefix(effect.Path, ".hive/spec/g/") {
			t.Errorf("effect %+v names another feature; want none", effect)
		}
	}
	if got := readSpec(t, home, ".hive/spec/f/README.md"); got != "readme v1\n" {
		t.Errorf("home f README = %q, want the task change", got)
	}
}
