package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNoLetterCaseClash checks that no two tracked or new paths differ only in
// letter case. Go refuses a module zip with such paths, so go install fails.
func TestNoLetterCaseClash(t *testing.T) {
	skipWithoutGit(t)
	out, err := exec.Command("git", "ls-files", "-co", "--exclude-standard").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	seen := map[string]string{}
	for _, path := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if path == "" || strings.HasPrefix(path, "projects/") {
			continue
		}
		key := strings.ToLower(path)
		if other, ok := seen[key]; ok {
			t.Errorf("paths differ only in letter case: %s and %s", other, path)
			continue
		}
		seen[key] = path
	}
}
