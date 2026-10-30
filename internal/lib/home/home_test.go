package home

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUsesEnvHome(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	envHome := filepath.Join(t.TempDir(), "env")

	got, err := Resolve(envHome, t.TempDir())
	if err != nil || got != envHome {
		t.Fatalf("Resolve() = %q, %v; want HIVE_HOME %q", got, err, envHome)
	}
}

func TestResolveEmptyEnvUsesDotHiveInUserHome(t *testing.T) {
	userHome := t.TempDir()
	t.Setenv("HOME", userHome)

	got, err := Resolve("", t.TempDir())
	if want := filepath.Join(userHome, ".hive"); err != nil || got != want {
		t.Fatalf("Resolve() = %q, %v; want %q", got, err, want)
	}
	if _, err := os.Stat(got); !os.IsNotExist(err) {
		t.Fatalf("Resolve() made %s: stat error = %v; want the home not created", got, err)
	}
}

func TestResolveIgnoresMarkerAboveWorkingDirectory(t *testing.T) {
	userHome := t.TempDir()
	t.Setenv("HOME", userHome)
	marked := t.TempDir()
	if err := os.MkdirAll(filepath.Join(marked, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(marked, ".hive", "README.md"), []byte("home"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve("", marked)
	if want := filepath.Join(userHome, ".hive"); err != nil || got != want {
		t.Fatalf("Resolve() = %q, %v; want %q, not the marked directory", got, err, want)
	}
}

func TestResolveRelativeEnvHomeAgainstWorkingDirectory(t *testing.T) {
	cwd := t.TempDir()
	got, err := Resolve("env-home", cwd)
	if want := filepath.Join(cwd, "env-home"); err != nil || got != want {
		t.Fatalf("Resolve() = %q, %v; want %q", got, err, want)
	}
}

func TestResolveDoesNotTrimEnvHomeValue(t *testing.T) {
	cwd := t.TempDir()
	got, err := Resolve(" env home ", cwd)
	if want := filepath.Join(cwd, " env home "); err != nil || got != want {
		t.Fatalf("Resolve() = %q, %v; want %q", got, err, want)
	}
}

func TestResolveReturnsPathWithoutSymlinks(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	real := filepath.Join(root, "real")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(alias, t.TempDir())
	if err != nil || got != real {
		t.Fatalf("Resolve() = %q, %v; want canonical home %q", got, err, real)
	}
}

func TestResolveFailsWithoutUserHome(t *testing.T) {
	t.Setenv("HOME", "")

	got, err := Resolve("", t.TempDir())
	if got != "" || err == nil {
		t.Fatalf("Resolve() = %q, %v; want an error when the user home is unknown", got, err)
	}
}
