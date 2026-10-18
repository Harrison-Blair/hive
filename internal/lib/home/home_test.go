package home

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveFlagWins(t *testing.T) {
	root := t.TempDir()
	flagHome := makeHome(t, root, "flag")
	envHome := makeHome(t, root, "env")
	workHome := makeHome(t, root, "work")
	cwd := filepath.Join(workHome, "projects", "nested")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(flagHome, true, envHome, cwd)
	if err != nil || got != flagHome {
		t.Fatalf("Resolve() = %q, %v; want %q, nil", got, err, flagHome)
	}
}

func TestResolveEnvWinsOverAncestor(t *testing.T) {
	root := t.TempDir()
	envHome := makeHome(t, root, "env")
	workHome := makeHome(t, root, "work")
	cwd := filepath.Join(workHome, "projects")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve("", false, envHome, cwd)
	want := "home conflict: HIVE_HOME is " + envHome + ", working directory is in " + workHome
	if err == nil || got != "" || err.Error() != want {
		t.Fatalf("Resolve() = %q, %v; want %q", got, err, want)
	}
}

func TestResolveOutermostMarkerWins(t *testing.T) {
	root := makeHome(t, t.TempDir(), "outer")
	worktree := filepath.Join(root, "projects", "hive", "worktrees", "task")
	if err := os.MkdirAll(filepath.Join(worktree, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, ".hive", "README.md"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve("", false, "", worktree)
	if err != nil || got != root {
		t.Fatalf("Resolve() = %q, %v; want outer home %q", got, err, root)
	}
}

func TestResolveFlagSkipsConflictCheck(t *testing.T) {
	root := t.TempDir()
	flagHome := makeHome(t, root, "flag")
	envHome := makeHome(t, root, "env")
	cwd := filepath.Join(envHome, "projects")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(flagHome, true, envHome, cwd)
	if err != nil || got != flagHome {
		t.Fatalf("Resolve() = %q, %v; want %q, nil", got, err, flagHome)
	}
}

func TestResolveNamedPathWithoutMarker(t *testing.T) {
	path := t.TempDir()
	got, err := Resolve(path, true, "", t.TempDir())
	if got != "" || err == nil || err.Error() != "not a hive home: "+path {
		t.Fatalf("Resolve() = %q, %v; want not a hive home: %s", got, err, path)
	}
}

func TestResolveEnvPathWithoutMarker(t *testing.T) {
	path := t.TempDir()
	got, err := Resolve("", false, path, t.TempDir())
	if got != "" || err == nil || err.Error() != "not a hive home: "+path {
		t.Fatalf("Resolve() = %q, %v; want not a hive home: %s", got, err, path)
	}
}

func TestResolveNoMarker(t *testing.T) {
	cwd := t.TempDir()
	got, err := Resolve("", false, "", cwd)
	if got != "" || err == nil || err.Error() != "not a hive home: no marker above "+cwd {
		t.Fatalf("Resolve() = %q, %v; want no marker error for %s", got, err, cwd)
	}
}

func TestResolveEmptyHomeFlagIsUsageError(t *testing.T) {
	_, err := Resolve("", true, "", t.TempDir())
	var usage *UsageError
	if !errors.As(err, &usage) {
		t.Fatalf("Resolve() error = %v; want UsageError", err)
	}
}

func TestResolveEmptyEnvIsUnset(t *testing.T) {
	home := makeHome(t, t.TempDir(), "home")
	cwd := filepath.Join(home, "nested")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve("", false, "", cwd)
	if err != nil || got != home {
		t.Fatalf("Resolve() = %q, %v; want %q, nil", got, err, home)
	}
}

func TestResolveRelativeHomePathsAgainstWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "cwd")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	flagHome := makeHome(t, cwd, "flag-home")
	got, err := Resolve(filepath.Join("flag-home"), true, "", cwd)
	if err != nil || got != flagHome {
		t.Fatalf("relative --home resolved to %q, %v; want %q", got, err, flagHome)
	}

	envHome := makeHome(t, cwd, "env-home")
	got, err = Resolve("", false, filepath.Join("env-home"), cwd)
	if err != nil || got != envHome {
		t.Fatalf("relative HIVE_HOME resolved to %q, %v; want %q", got, err, envHome)
	}
}

func TestResolveConflictComparesResolvedPaths(t *testing.T) {
	root := t.TempDir()
	home := makeHome(t, root, "home")
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(home, alias); err != nil {
		t.Fatal(err)
	}
	cwd := filepath.Join(home, "nested")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve("", false, alias, cwd)
	if err != nil || got != home {
		t.Fatalf("Resolve() = %q, %v; want canonical home %q", got, err, home)
	}
}

func TestResolveDoesNotTrimHomeValues(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "cwd")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	path := makeHome(t, cwd, " home ")
	got, err := Resolve(" home ", true, "", cwd)
	if err != nil || got != path {
		t.Fatalf("Resolve() = %q, %v; want path with spaces %q", got, err, path)
	}
}

func TestResolveDoesNotTrimEnvHomeValue(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "cwd")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	path := makeHome(t, cwd, " env home ")
	got, err := Resolve("", false, " env home ", cwd)
	if err != nil || got != path {
		t.Fatalf("Resolve() = %q, %v; want environment path with spaces %q", got, err, path)
	}
}

func TestResolveMarkerMustBeRegularFile(t *testing.T) {
	root := t.TempDir()
	markerDir := filepath.Join(root, ".hive")
	if err := os.Mkdir(markerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("home"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(markerDir, "README.md")); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve("", false, "", root)
	if got != "" || err == nil || err.Error() != "not a hive home: no marker above "+root {
		t.Fatalf("Resolve() = %q, %v; want no marker because README.md is a symlink", got, err)
	}
}

func TestResolveReturnsAbsolutePathWithoutSymlinks(t *testing.T) {
	root := makeHome(t, t.TempDir(), "home")
	alias := filepath.Join(filepath.Dir(root), "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve(alias, true, "", t.TempDir())
	if err != nil || got != root || !filepath.IsAbs(got) {
		t.Fatalf("Resolve() = %q, %v; want absolute canonical path %q", got, err, root)
	}
}

func makeHome(t *testing.T, parent, name string) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.MkdirAll(filepath.Join(path, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, ".hive", "README.md"), []byte("home"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
