package tui

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

// requireSpecFile skips the test when a file or a folder of the local spec is
// missing. Git does not track .hive/spec/, so a clean clone has no grid and
// no fixture. The test runs as before when the file exists.
func requireSpecFile(t testing.TB, path string) {
	t.Helper()
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		t.Skipf("local spec file missing: %s: Git does not track .hive/spec/", path)
	}
}
