package scope

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestResolveHomeScope(t *testing.T) {
	home := t.TempDir()
	got, err := Resolve(home, "home")
	if err != nil {
		t.Fatalf("Resolve() error = %v; want home accepted", err)
	}
	if got.Dir != home {
		t.Fatalf("Resolve().Dir = %q; want home root %q", got.Dir, home)
	}
}

func TestResolveProjectScope(t *testing.T) {
	home := t.TempDir()
	registerProject(t, home, "hive")
	got, err := Resolve(home, "project:hive")
	if err != nil {
		t.Fatalf("Resolve() error = %v; want project accepted", err)
	}
	want := filepath.Join(home, "projects", "hive")
	if got.Dir != want {
		t.Fatalf("Resolve().Dir = %q; want %q", got.Dir, want)
	}
}

func TestResolveRejectsBadScopes(t *testing.T) {
	values := []string{
		"",
		"project:",
		"project:HIVE",
		"project:a/b",
		"project:..",
		"project:%2e%2e%2f",
		"other",
		" home",
		"HOME",
		"project: hive",
		"project:hive\n",
	}
	for _, value := range values {
		t.Run(value, func(t *testing.T) {
			_, err := Resolve(t.TempDir(), value)
			if err == nil || !strings.HasPrefix(err.Error(), "bad scope: ") {
				t.Fatalf("Resolve(%q) error = %v; want bad scope", value, err)
			}
			if strings.Contains(err.Error(), "\n") {
				t.Fatalf("Resolve(%q) error = %q; want one physical line", value, err)
			}
		})
	}
}

func TestResolveReportsMissingProject(t *testing.T) {
	_, err := Resolve(t.TempDir(), "project:nope")
	if err == nil || err.Error() != "project not found: nope" {
		t.Fatalf("Resolve() error = %v; want project not found", err)
	}
}

func TestResolveRejectsLinkedProjectDirectory(t *testing.T) {
	home := t.TempDir()
	registerProject(t, home, "hive")
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(home, "projects", "hive")); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(home, "project:hive"); err == nil {
		t.Fatal("Resolve() accepted a symbolic project directory")
	}
}

func TestCheckFileRejectsStaticDirectoryAndLink(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, "memory", "MEMORY.md")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := CheckFile(home, target); err == nil {
		t.Fatal("CheckFile() accepted a static directory")
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "index.md")
	if err := os.WriteFile(outside, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, target); err != nil {
		t.Fatal(err)
	}
	if err := CheckFile(home, target); err == nil {
		t.Fatal("CheckFile() accepted a static symbolic link")
	}
}

func TestCheckFileRejectsLinkedAncestor(t *testing.T) {
	home := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, "memory")); err != nil {
		t.Fatal(err)
	}
	if err := CheckFile(home, filepath.Join(home, "memory", "fact.md")); err == nil {
		t.Fatal("CheckFile() accepted a linked ancestor")
	}
	if err := CheckDir(home, filepath.Join(home, "memory")); err == nil {
		t.Fatal("CheckDir() accepted a linked directory")
	}
}

func TestCheckFileRejectsOutsideHome(t *testing.T) {
	if err := CheckFile(t.TempDir(), filepath.Join(t.TempDir(), "fact.md")); err == nil {
		t.Fatal("CheckFile() accepted a path outside the home")
	}
}

func TestCheckFileAllowsMissingPath(t *testing.T) {
	home := t.TempDir()
	if err := CheckFile(home, filepath.Join(home, "memory", "fact.md")); err != nil {
		t.Fatalf("CheckFile() error = %v; want missing path accepted", err)
	}
	if err := CheckDir(home, filepath.Join(home, "memory")); err != nil {
		t.Fatalf("CheckDir() error = %v; want missing directory accepted", err)
	}
}

func TestReadInputAtLimit(t *testing.T) {
	for _, size := range []int{1048575, 1048576} {
		data := bytes.Repeat([]byte("a"), size)
		got, err := ReadInput(bytes.NewReader(data), 1<<20, time.Second)
		if err != nil {
			t.Fatalf("ReadInput(%d bytes) error = %v; want success", size, err)
		}
		if len(got) != size {
			t.Fatalf("ReadInput(%d bytes) length = %d; want %d", size, len(got), size)
		}
	}
}

func TestReadInputAboveLimit(t *testing.T) {
	data := bytes.Repeat([]byte("a"), 1048577)
	if _, err := ReadInput(bytes.NewReader(data), 1<<20, time.Second); !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("ReadInput(1048577 bytes) error = %v; want ErrInputTooLarge", err)
	}
}

func TestReadInputTimeout(t *testing.T) {
	reader, writer := io.Pipe()
	defer func() { _ = reader.Close() }()
	started := time.Now()
	_, err := ReadInput(reader, 1<<20, 100*time.Millisecond)
	_ = writer.Close()
	if !errors.Is(err, ErrInputTimeout) {
		t.Fatalf("ReadInput() error = %v; want ErrInputTimeout", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("ReadInput() took %s; want a bounded timeout", elapsed)
	}
}

func registerProject(t *testing.T, home, id string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Write("projects", id, records.Project{ID: id, Path: home, Base: "dev"}); err != nil {
		t.Fatal(err)
	}
}
