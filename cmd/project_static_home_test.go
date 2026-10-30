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

func TestProjectListRejectsLinkedProjectsDirectory(t *testing.T) {
	for _, jsonOutput := range []bool{false, true} {
		t.Run(map[bool]string{false: "text", true: "json"}[jsonOutput], func(t *testing.T) {
			home := testHome(t)
			external := t.TempDir()
			if err := os.MkdirAll(filepath.Join(external, "linked"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(external, "linked", "project.json"), []byte(`{"id":"linked"}`), 0o644); err != nil {
				t.Fatal(err)
			}
			projects := filepath.Join(home, "projects")
			if err := os.Symlink(external, projects); err != nil {
				t.Fatal(err)
			}
			beforeExternal := captureTree(t, external)

			args := []string{"project", "list"}
			if jsonOutput {
				args = append([]string{"--json"}, args...)
			}
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(args, strings.NewReader(""), &stdout, &stderr)
			if status != 1 {
				t.Fatalf("status = %d; want 1 (stdout %q, stderr %q)", status, stdout.String(), stderr.String())
			}
			wantError := "store path invalid: " + projects
			if jsonOutput {
				var envelope map[string]any
				if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil || stderr.Len() != 0 {
					t.Fatalf("stdout = %q, stderr = %q; want one JSON object", stdout.String(), stderr.String())
				}
				if envelope["outcome"] != "failed" || envelope["phase"] != "store" || envelope["error"] != wantError {
					t.Fatalf("envelope = %#v; want failed store error %q", envelope, wantError)
				}
			} else if got := stderr.String(); stdout.Len() != 0 || got != "hive: "+wantError+"\n" {
				t.Fatalf("stdout = %q, stderr = %q; want one store error line %q", stdout.String(), got, wantError)
			}
			assertSameTree(t, "external", beforeExternal, captureTree(t, external))
		})
	}
}

func TestProjectListAcceptsRegularHiveHome(t *testing.T) {
	home := testHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"project", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || stdout.String() != "no projects\n" || stderr.Len() != 0 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want a normal list", status, stdout.String(), stderr.String())
	}
	if info, err := os.Stat(filepath.Join(home, "logs")); err != nil || !info.IsDir() {
		t.Fatalf("home logs: stat = %v, %v; want a directory", info, err)
	}
	if info, err := os.Stat(filepath.Join(home, "schema.json")); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("home schema.json: stat = %v, %v; want a file", info, err)
	}
	for _, name := range []string{".hive", "state"} {
		if _, err := os.Stat(filepath.Join(home, name)); !os.IsNotExist(err) {
			t.Fatalf("home %s: stat error = %v; want none", name, err)
		}
	}
}

func TestProjectListAcceptsSymlinkedHomeRoot(t *testing.T) {
	home := testHome(t)
	alias := filepath.Join(filepath.Dir(home), "alias")
	if err := os.Symlink(home, alias); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", alias)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"project", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || stdout.String() != "no projects\n" || stderr.Len() != 0 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want a normal list", status, stdout.String(), stderr.String())
	}
	info, err := os.Stat(filepath.Join(home, "logs"))
	if err != nil || !info.IsDir() {
		t.Fatalf("canonical home logs: stat = %v, %v; want a directory", info, err)
	}
	if info, err := os.Lstat(alias); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("alias: lstat = %v, %v; want a symbolic link", info, err)
	}
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
