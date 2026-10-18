package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type staticTreeEntry struct {
	mode os.FileMode
	data []byte
}

func TestProjectListRejectsLinkedHiveHome(t *testing.T) {
	for _, test := range []struct {
		name string
		json bool
		env  bool
		cwd  bool
	}{
		{name: "flag text"},
		{name: "flag json", json: true},
		{name: "environment text", env: true},
		{name: "cwd discovery text", cwd: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			home, external := linkedHomeFixture(t)
			t.Setenv("HIVE_HOME", "")
			beforeExternal := captureTree(t, external)
			beforeHome := captureTree(t, home)

			args := []string{}
			if test.json {
				args = append(args, "--json")
			}
			t.Chdir(filepath.Dir(home))
			if test.env {
				t.Setenv("HIVE_HOME", home)
			} else if test.cwd {
				t.Chdir(home)
			} else {
				args = append(args, "--home", home)
			}
			args = append(args, "project", "list")

			var stdout, stderr bytes.Buffer
			status := executeWithStreams(args, strings.NewReader(""), &stdout, &stderr)
			if status != 1 {
				t.Fatalf("status = %d; want 1 (stdout %q, stderr %q)", status, stdout.String(), stderr.String())
			}

			wantError := "not a hive home: " + home
			if test.cwd {
				wantError = "not a hive home: no marker above " + home
			}
			if test.json {
				if stderr.Len() != 0 {
					t.Fatalf("stderr = %q; want empty", stderr.String())
				}
				if strings.Count(stdout.String(), "\n") != 1 {
					t.Fatalf("stdout = %q; want one JSON line", stdout.String())
				}
				var envelope map[string]any
				if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
					t.Fatalf("stdout is not one JSON object: %v; output %q", err, stdout.String())
				}
				if envelope["outcome"] != "failed" || envelope["phase"] != "home" || envelope["error"] != wantError {
					t.Fatalf("envelope = %#v; want failed home error %q", envelope, wantError)
				}
			} else {
				if stdout.Len() != 0 {
					t.Fatalf("stdout = %q; want empty", stdout.String())
				}
				if got := stderr.String(); got != "hive: "+wantError+"\n" {
					t.Fatalf("stderr = %q; want one home error line %q", got, wantError)
				}
			}

			assertSameTree(t, "external", beforeExternal, captureTree(t, external))
			assertSameTree(t, "home", beforeHome, captureTree(t, home))
		})
	}
}

func TestProjectListAcceptsRegularHiveHome(t *testing.T) {
	home := testHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "project", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || stdout.String() != "no projects\n" || stderr.Len() != 0 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want a normal list", status, stdout.String(), stderr.String())
	}
	for _, name := range []string{"logs", "state"} {
		info, err := os.Stat(filepath.Join(home, ".hive", name))
		if err != nil || !info.IsDir() {
			t.Fatalf("regular home .hive/%s: stat = %v, %v; want a directory", name, info, err)
		}
	}
}

func TestProjectListAcceptsSymlinkedHomeRoot(t *testing.T) {
	home := testHome(t)
	alias := filepath.Join(filepath.Dir(home), "alias")
	if err := os.Symlink(home, alias); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", alias, "project", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || stdout.String() != "no projects\n" || stderr.Len() != 0 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want a normal list", status, stdout.String(), stderr.String())
	}
	info, err := os.Stat(filepath.Join(home, ".hive", "logs"))
	if err != nil || !info.IsDir() {
		t.Fatalf("canonical home logs: stat = %v, %v; want a directory", info, err)
	}
	if info, err := os.Lstat(alias); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("alias: lstat = %v, %v; want a symbolic link", info, err)
	}
}

func linkedHomeFixture(t *testing.T) (home, external string) {
	t.Helper()
	root := t.TempDir()
	external = filepath.Join(root, "external")
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "README.md"), []byte("linked marker "+t.Name()), 0o644); err != nil {
		t.Fatal(err)
	}
	home = filepath.Join(root, "home")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(home, ".hive")); err != nil {
		t.Fatal(err)
	}
	return home, external
}

func captureTree(t *testing.T, root string) map[string]staticTreeEntry {
	t.Helper()
	entries := map[string]staticTreeEntry{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		captured := staticTreeEntry{mode: info.Mode()}
		if info.Mode().IsRegular() {
			captured.data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		}
		entries[relative] = captured
		return nil
	})
	if err != nil {
		t.Fatalf("capture %s: %v", root, err)
	}
	return entries
}

func assertSameTree(t *testing.T, label string, before, after map[string]staticTreeEntry) {
	t.Helper()
	for path, want := range before {
		got, ok := after[path]
		if !ok {
			t.Errorf("%s tree lost %s", label, path)
			continue
		}
		if got.mode != want.mode || !bytes.Equal(got.data, want.data) {
			t.Errorf("%s tree changed %s: mode %v -> %v, data %q -> %q", label, path, want.mode, got.mode, want.data, got.data)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			t.Errorf("%s tree gained %s", label, path)
		}
	}
}
