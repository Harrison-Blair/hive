package home

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestResolveEnvWinsOverAncestor(t *testing.T) {
	isolateConfig(t)
	root := t.TempDir()
	envHome := makeHome(t, root, "env")
	workHome := makeHome(t, root, "work")
	cwd := filepath.Join(workHome, "projects")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(envHome, cwd)
	if err != nil || got != envHome {
		t.Fatalf("Resolve() = %q, %v; want HIVE_HOME %q despite ancestor %q", got, err, envHome, workHome)
	}
}

func TestResolveOutermostMarkerWins(t *testing.T) {
	isolateConfig(t)
	root := makeHome(t, t.TempDir(), "outer")
	worktree := filepath.Join(root, "projects", "hive", "worktrees", "task")
	if err := os.MkdirAll(filepath.Join(worktree, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, ".hive", "README.md"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve("", worktree)
	if err != nil || got != root {
		t.Fatalf("Resolve() = %q, %v; want outer home %q", got, err, root)
	}
}

func TestResolveNamedPathWithoutMarker(t *testing.T) {
	isolateConfig(t)
	path := t.TempDir()
	got, err := Resolve(path, t.TempDir())
	if got != "" || err == nil || err.Error() != "not a hive home: "+path {
		t.Fatalf("Resolve() = %q, %v; want not a hive home: %s", got, err, path)
	}
}

func TestResolveEnvPathWithoutMarker(t *testing.T) {
	isolateConfig(t)
	path := t.TempDir()
	got, err := Resolve(path, t.TempDir())
	if got != "" || err == nil || err.Error() != "not a hive home: "+path {
		t.Fatalf("Resolve() = %q, %v; want not a hive home: %s", got, err, path)
	}
}

func TestResolveNoMarker(t *testing.T) {
	isolateConfig(t)
	cwd := t.TempDir()
	got, err := Resolve("", cwd)
	if got != "" || err == nil || err.Error() != "not a hive home: no marker above "+cwd+"; run hive init in the home" {
		t.Fatalf("Resolve() = %q, %v; want init guidance for %s", got, err, cwd)
	}
}

func TestResolveEmptyEnvIsUnset(t *testing.T) {
	isolateConfig(t)
	home := makeHome(t, t.TempDir(), "home")
	cwd := filepath.Join(home, "nested")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve("", cwd)
	if err != nil || got != home {
		t.Fatalf("Resolve() = %q, %v; want %q, nil", got, err, home)
	}
}

func TestResolveRelativeEnvHomeAgainstWorkingDirectory(t *testing.T) {
	isolateConfig(t)
	root := t.TempDir()
	cwd := filepath.Join(root, "cwd")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	envHome := makeHome(t, cwd, "env-home")
	got, err := Resolve(filepath.Join("env-home"), cwd)
	if err != nil || got != envHome {
		t.Fatalf("relative HIVE_HOME resolved to %q, %v; want %q", got, err, envHome)
	}
}

func TestResolveConflictComparesResolvedPaths(t *testing.T) {
	isolateConfig(t)
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
	got, err := Resolve(alias, cwd)
	if err != nil || got != home {
		t.Fatalf("Resolve() = %q, %v; want canonical home %q", got, err, home)
	}
}

func TestResolveDoesNotTrimEnvHomeValue(t *testing.T) {
	isolateConfig(t)
	root := t.TempDir()
	cwd := filepath.Join(root, "cwd")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	path := makeHome(t, cwd, " env home ")
	got, err := Resolve(" env home ", cwd)
	if err != nil || got != path {
		t.Fatalf("Resolve() = %q, %v; want environment path with spaces %q", got, err, path)
	}
}

func TestResolveMarkerMustBeRegularFile(t *testing.T) {
	isolateConfig(t)
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
	got, err := Resolve("", root)
	if got != "" || err == nil || err.Error() != "not a hive home: no marker above "+root+"; run hive init in the home" {
		t.Fatalf("Resolve() = %q, %v; want no marker because README.md is a symlink", got, err)
	}
}

func TestResolveReturnsAbsolutePathWithoutSymlinks(t *testing.T) {
	isolateConfig(t)
	root := makeHome(t, t.TempDir(), "home")
	alias := filepath.Join(filepath.Dir(root), "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve(alias, t.TempDir())
	if err != nil || got != root || !filepath.IsAbs(got) {
		t.Fatalf("Resolve() = %q, %v; want absolute canonical path %q", got, err, root)
	}
}

func TestResolveRejectsLinkedHiveDirectory(t *testing.T) {
	isolateConfig(t)
	root := t.TempDir()
	external := filepath.Join(root, "external")
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "README.md"), []byte("linked marker"), 0o644); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(root, "home")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(home, ".hive")); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(home, t.TempDir())
	if got != "" || err == nil || err.Error() != "not a hive home: "+home {
		t.Fatalf("Resolve() = %q, %v; want no home because .hive is a symlink", got, err)
	}
}

func TestResolveRejectsNonDirectoryHive(t *testing.T) {
	isolateConfig(t)
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive"), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(home, t.TempDir())
	if got != "" || err == nil || err.Error() != "not a hive home: "+home {
		t.Fatalf("Resolve() = %q, %v; want no home because .hive is a regular file", got, err)
	}
}

func TestResolveAncestorSkipsLinkedHiveMarker(t *testing.T) {
	isolateConfig(t)
	outer := makeHome(t, t.TempDir(), "outer")
	linked := filepath.Join(outer, "linked")
	if err := os.Mkdir(linked, 0o755); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(linked, "external")
	if err := os.Mkdir(external, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "README.md"), []byte("linked marker"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(linked, ".hive")); err != nil {
		t.Fatal(err)
	}
	cwd := filepath.Join(linked, "nested")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve("", cwd)
	if err != nil || got != outer {
		t.Fatalf("Resolve() = %q, %v; want outer home %q", got, err, outer)
	}
}

func TestResolveUsesConfigFromUnmarkedDirectory(t *testing.T) {
	configPath := configFixture(t, "")
	configHome := makeHome(t, t.TempDir(), "config-home")
	writeConfig(t, configPath, "home = "+strconv.Quote(configHome)+"\n")
	cwd := t.TempDir()

	got, err := Resolve("", cwd)
	if err != nil || got != configHome {
		t.Fatalf("Resolve() = %q, %v; want config home %q", got, err, configHome)
	}
}

func TestResolveConfigRequiresMarkedHome(t *testing.T) {
	configPath := configFixture(t, "")
	configHome := t.TempDir()
	writeConfig(t, configPath, "home = "+strconv.Quote(configHome)+"\n")

	got, err := Resolve("", t.TempDir())
	if got != "" || err == nil || !strings.Contains(err.Error(), configPath) || !strings.Contains(err.Error(), configHome) {
		t.Fatalf("Resolve() = %q, %v; want invalid config home error naming %s and %s", got, err, configPath, configHome)
	}
}

func TestResolveConfigConflictNamesBothHomes(t *testing.T) {
	configPath := configFixture(t, "")
	configHome := makeHome(t, t.TempDir(), "config-home")
	workHome := makeHome(t, t.TempDir(), "work-home")
	writeConfig(t, configPath, "home = "+strconv.Quote(configHome)+"\n")
	cwd := filepath.Join(workHome, "projects", "nested")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve("", cwd)
	if got != "" || err == nil || !strings.HasPrefix(err.Error(), "home conflict:") || !strings.Contains(err.Error(), configHome) || !strings.Contains(err.Error(), workHome) {
		t.Fatalf("Resolve() = %q, %v; want conflict naming %q and %q", got, err, configHome, workHome)
	}
}

func TestResolveEnvWinsOverConfigAndAncestor(t *testing.T) {
	configPath := configFixture(t, "")
	configHome := makeHome(t, t.TempDir(), "config-home")
	envHome := makeHome(t, t.TempDir(), "env-home")
	workHome := makeHome(t, t.TempDir(), "work-home")
	writeConfig(t, configPath, "home = "+strconv.Quote(configHome)+"\n")
	cwd := filepath.Join(workHome, "nested")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(envHome, cwd)
	if err != nil || got != envHome {
		t.Fatalf("Resolve() = %q, %v; want HIVE_HOME %q", got, err, envHome)
	}
}

func TestResolveConfigErrorsNameConfigPath(t *testing.T) {
	validHome := makeHome(t, t.TempDir(), "config-home")
	for _, test := range []struct {
		name    string
		content string
	}{
		{name: "missing key", content: "other = \"value\"\n"},
		{name: "relative home", content: "home = \"relative\"\n"},
		{name: "invalid string", content: "home = [\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			configPath := configFixture(t, "")
			writeConfig(t, configPath, test.content)
			got, err := Resolve("", t.TempDir())
			if got != "" || err == nil || !strings.Contains(err.Error(), configPath) {
				t.Fatalf("Resolve() = %q, %v; want config error naming %s", got, err, configPath)
			}
		})
	}

	configPath := configFixture(t, "")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(configPath, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve("", validHome)
	if got != "" || err == nil || !strings.Contains(err.Error(), configPath) {
		t.Fatalf("Resolve() = %q, %v; want read error naming %s", got, err, configPath)
	}
}

func configFixture(t *testing.T, content string) string {
	t.Helper()
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	path := filepath.Join(configDir, "hive", "config.toml")
	if content != "" {
		writeConfig(t, path, content)
	}
	return path
}

func writeConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func isolateConfig(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
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
