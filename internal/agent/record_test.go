package agent

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sample(name, pane string) Record {
	return Record{
		Name:      name,
		Pane:      pane,
		Parent:    "manager-hive",
		Harness:   "claude",
		Model:     "opus",
		Cwd:       "/src/hive",
		Worktree:  "/src/hive/projects/hive/worktrees/207",
		StartedAt: time.Date(2026, 10, 28, 9, 30, 15, 0, time.UTC),
	}
}

func agentsDir(home string) string {
	return filepath.Join(home, "agents")
}

func TestSaveLoadRoundTripsAllFields(t *testing.T) {
	home := t.TempDir()
	want := sample("t207-builder-1", "w1:p2")
	if err := Save(home, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(home, "t207-builder-1")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Load = %+v, want %+v", got, want)
	}
}

func TestSaveWritesSnakeCaseJSONAtNamePath(t *testing.T) {
	home := t.TempDir()
	if err := Save(home, sample("a", "p1")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(agentsDir(home), "a.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"name":       "a",
		"pane":       "p1",
		"parent":     "manager-hive",
		"harness":    "claude",
		"model":      "opus",
		"cwd":        "/src/hive",
		"worktree":   "/src/hive/projects/hive/worktrees/207",
		"started_at": "2026-10-28T09:30:15Z",
	}
	if len(fields) != len(want) {
		t.Fatalf("fields = %v, want exactly %v", fields, want)
	}
	for key, value := range want {
		if fields[key] != value {
			t.Errorf("field %s = %v, want %v", key, fields[key], value)
		}
	}
}

func TestSaveStoresStartedAtInUTC(t *testing.T) {
	home := t.TempDir()
	record := sample("a", "p1")
	record.StartedAt = time.Date(2026, 10, 28, 11, 0, 0, 0, time.FixedZone("CET", 3600))
	if err := Save(home, record); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(agentsDir(home), "a.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"started_at":"2026-10-28T10:00:00Z"`) {
		t.Fatalf("record = %s, want started_at in UTC", data)
	}
}

func TestSaveLeavesNoTempFile(t *testing.T) {
	home := t.TempDir()
	if err := Save(home, sample("a", "p1")); err != nil {
		t.Fatal(err)
	}
	if err := Save(home, sample("a", "p2")); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(agentsDir(home))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "a.json" {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("directory = %v, want only a.json", names)
	}
}

func TestLoadMissingIsNotFound(t *testing.T) {
	home := t.TempDir()
	_, err := Load(home, "ghost")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load error = %v, want ErrNotFound", err)
	}
}

func TestListSortsByNameAndSkipsOtherFiles(t *testing.T) {
	home := t.TempDir()
	for _, name := range []string{"zeta", "alpha", "mid-1"} {
		if err := Save(home, sample(name, "p-"+name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(agentsDir(home), "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir(home), ".a.json.tmp-1"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	records, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, record := range records {
		names = append(names, record.Name)
	}
	if strings.Join(names, ",") != "alpha,mid-1,zeta" {
		t.Fatalf("List names = %v, want [alpha mid-1 zeta]", names)
	}
}

func TestListEmptyHome(t *testing.T) {
	records, err := List(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("List = %v, want none", records)
	}
}

func TestFindByPane(t *testing.T) {
	home := t.TempDir()
	if err := Save(home, sample("a", "w1:p1")); err != nil {
		t.Fatal(err)
	}
	if err := Save(home, sample("b", "w1:p2")); err != nil {
		t.Fatal(err)
	}
	got, err := FindByPane(home, "w1:p2")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "b" {
		t.Fatalf("FindByPane name = %q, want b", got.Name)
	}
	if _, err := FindByPane(home, "w9:p9"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("FindByPane missing error = %v, want ErrNotFound", err)
	}
}

func TestRemove(t *testing.T) {
	home := t.TempDir()
	if err := Save(home, sample("a", "p1")); err != nil {
		t.Fatal(err)
	}
	if err := Remove(home, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(home, "a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load after Remove error = %v, want ErrNotFound", err)
	}
	if err := Remove(home, "a"); err != nil {
		t.Fatalf("Remove absent error = %v, want nil", err)
	}
	if err := Remove(t.TempDir(), "never"); err != nil {
		t.Fatalf("Remove in empty home error = %v, want nil", err)
	}
}

func TestSaveRejectsBadNamesAndWritesNothing(t *testing.T) {
	for _, name := range []string{
		"",
		"Upper",
		"has space",
		"dot.name",
		"../escape",
		"slash/name",
		strings.Repeat("a", 65),
	} {
		home := t.TempDir()
		if err := Save(home, sample(name, "p1")); err == nil {
			t.Errorf("Save(%q) error = nil, want error", name)
		}
		if _, err := os.Stat(filepath.Join(home, "agents")); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Save(%q) wrote to the home: stat error = %v", name, err)
		}
	}
}

func TestSaveAcceptsLongestValidName(t *testing.T) {
	home := t.TempDir()
	name := strings.Repeat("a", 60) + "_-09"
	if err := Save(home, sample(name, "p1")); err != nil {
		t.Fatalf("Save(64-char name) error = %v", err)
	}
}
