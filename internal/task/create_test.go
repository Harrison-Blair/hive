package task

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// newHome makes one fixture home with its marker.
func newHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

// registerProject writes one project record.
func registerProject(t *testing.T, home, id string) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Project{ID: id, Path: filepath.Join(home, "repo-"+id), Base: "main"}
	if err := registry.Write("projects", id, record); err != nil {
		t.Fatal(err)
	}
}

// createOK creates one task and checks a success outcome.
func createOK(t *testing.T, home string, options CreateOptions) records.Task {
	t.Helper()
	envelope := Create(home, options)
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Create(%+v) = %#v; want ok", options, envelope)
	}
	task, ok := envelope.Result.(records.Task)
	if !ok {
		t.Fatalf("Create(%+v) result = %#v; want a task record", options, envelope.Result)
	}
	return task
}

func counterValue(t *testing.T, home string) (uint64, bool) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, ".hive", "state", "counters", "task.json"))
	if errors.Is(err, os.ErrNotExist) {
		return 0, false
	}
	if err != nil {
		t.Fatal(err)
	}
	var counter struct {
		Value uint64 `json:"value"`
	}
	if err := json.Unmarshal(data, &counter); err != nil {
		t.Fatal(err)
	}
	return counter.Value, true
}

func assertNoTaskRecord(t *testing.T, home, id string) {
	t.Helper()
	path := filepath.Join(home, ".hive", "state", "tasks", id+".json")
	if info, err := os.Lstat(path); err == nil {
		t.Fatalf("task path %s exists with mode %s; want no task record", path, info.Mode())
	}
}

func TestCreateCounterAcrossProjects(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	registerProject(t, home, "two")

	first := createOK(t, home, CreateOptions{Project: "one", Title: "First step"})
	if first.ID != "1-first-step" {
		t.Fatalf("first ID = %q; want 1-first-step", first.ID)
	}
	second := createOK(t, home, CreateOptions{Project: "two", Title: "Second step"})
	if second.ID != "2-second-step" {
		t.Fatalf("second ID = %q; want 2-second-step from the one home counter", second.ID)
	}
	if value, ok := counterValue(t, home); !ok || value != 2 {
		t.Fatalf("counter = %d, %v; want 2", value, ok)
	}
}

func TestCreateRecordFields(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")

	envelope := Create(home, CreateOptions{Project: "one", Title: "Spawn worker"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Create() = %#v; want ok", envelope)
	}
	if len(envelope.Effects) != 3 {
		t.Fatalf("effects = %#v; want counter, brief, and task", envelope.Effects)
	}
	wantEffects := []string{"updated counter", "created brief", "created task"}
	for i, want := range wantEffects {
		got := fmt.Sprintf("%s %s", envelope.Effects[i].Action, envelope.Effects[i].Resource)
		if got != want {
			t.Errorf("effect %d = %q; want %q", i, got, want)
		}
	}
	task, ok := envelope.Result.(records.Task)
	if !ok {
		t.Fatalf("result = %#v; want a task record", envelope.Result)
	}
	if task.ID != "1-spawn-worker" || task.Project != "one" || task.Title != "Spawn worker" {
		t.Errorf("record = %#v; want the ID, project, and title", task)
	}
	if task.ActingAgent != "" {
		t.Errorf("acting_agent = %q; want empty", task.ActingAgent)
	}
	if task.Status != "created" {
		t.Errorf("status = %q; want created", task.Status)
	}
	if task.DependsOn == nil || len(task.DependsOn) != 0 {
		t.Errorf("depends_on = %#v; want an empty list", task.DependsOn)
	}
	if task.Verification != nil || task.Worktree != nil || task.Delivery != nil {
		t.Errorf("optional records = %#v; want nil", task)
	}
	if task.CreatedAt == "" || task.UpdatedAt == "" {
		t.Errorf("timestamps = %q, %q; want values", task.CreatedAt, task.UpdatedAt)
	}
	if _, err := time.Parse(time.RFC3339Nano, task.CreatedAt); err != nil {
		t.Errorf("created_at = %q; want RFC 3339", task.CreatedAt)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "tasks", task.ID+".json")); err != nil {
		t.Errorf("task record: %v", err)
	}
	brief, err := os.ReadFile(filepath.Join(home, ".hive", "state", "briefs", task.ID+".md"))
	if err != nil {
		t.Fatalf("brief file: %v", err)
	}
	if string(brief) != "Spawn worker" {
		t.Fatalf("brief = %q; want the title", brief)
	}
}

func TestCreateDependsOn(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")

	first := createOK(t, home, CreateOptions{Project: "one", Title: "First step"})
	second := createOK(t, home, CreateOptions{
		Project:    "one",
		Title:      "Second step",
		DependsOn:  first.ID,
		DependsSet: true,
	})
	if len(second.DependsOn) != 1 || second.DependsOn[0] != first.ID {
		t.Fatalf("depends_on = %#v; want %q", second.DependsOn, first.ID)
	}
}

func TestCreateBriefFromFile(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	path := filepath.Join(t.TempDir(), "brief.md")
	const content = "# Brief\n\nThe body.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	task := createOK(t, home, CreateOptions{
		Project:   "one",
		Title:     "First step",
		BriefPath: path,
		BriefSet:  true,
	})
	brief, err := os.ReadFile(filepath.Join(home, ".hive", "state", "briefs", task.ID+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(brief) != content {
		t.Fatalf("brief = %q; want the file content", brief)
	}
}

func TestCreateBriefFromStdin(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	task := createOK(t, home, CreateOptions{
		Project:   "one",
		Title:     "First step",
		BriefPath: "-",
		BriefSet:  true,
		Input:     strings.NewReader("from standard input"),
	})
	brief, err := os.ReadFile(filepath.Join(home, ".hive", "state", "briefs", task.ID+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(brief) != "from standard input" {
		t.Fatalf("brief = %q; want the standard input text", brief)
	}
}

func TestCreateBriefDefaultsToTitle(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Second step"})
	brief, err := os.ReadFile(filepath.Join(home, ".hive", "state", "briefs", task.ID+".md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(brief) != "Second step" {
		t.Fatalf("brief = %q; want the title", brief)
	}
}

func TestCreateBriefCapBoundary(t *testing.T) {
	for _, size := range []int{1048575, 1048576, 1048577, 5242880} {
		t.Run(fmt.Sprintf("%d-bytes", size), func(t *testing.T) {
			home := newHome(t)
			registerProject(t, home, "one")
			path := filepath.Join(t.TempDir(), "brief.md")
			if err := os.WriteFile(path, bytes.Repeat([]byte("a"), size), 0o644); err != nil {
				t.Fatal(err)
			}
			envelope := Create(home, CreateOptions{
				Project:   "one",
				Title:     "Big brief",
				BriefPath: path,
				BriefSet:  true,
			})
			if size <= 1048576 {
				if envelope.Outcome != "ok" || envelope.Error != nil {
					t.Fatalf("Create() = %#v; want %d bytes accepted", envelope, size)
				}
				task := envelope.Result.(records.Task)
				brief, err := os.ReadFile(filepath.Join(home, ".hive", "state", "briefs", task.ID+".md"))
				if err != nil {
					t.Fatal(err)
				}
				if len(brief) != size {
					t.Fatalf("brief length = %d; want %d", len(brief), size)
				}
				return
			}
			if envelope.Outcome != "failed" || envelope.Error == nil {
				t.Fatalf("Create() = %#v; want %d bytes rejected", envelope, size)
			}
			if !strings.Contains(fmt.Sprint(envelope.Error), "brief too large") {
				t.Fatalf("error = %v; want a brief too large error", envelope.Error)
			}
			if _, ok := counterValue(t, home); ok {
				t.Fatalf("an oversized brief took a counter value")
			}
		})
	}
}

func TestCreateBriefDeadline(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = readEnd.Close() }()

	started := time.Now()
	envelope := Create(home, CreateOptions{
		Project:   "one",
		Title:     "Stalled",
		BriefPath: "-",
		BriefSet:  true,
		Input:     readEnd,
		Deadline:  50 * time.Millisecond,
	})
	elapsed := time.Since(started)
	_ = writeEnd.Close()

	if envelope.Outcome != "failed" || envelope.Error == nil {
		t.Fatalf("Create() = %#v; want a deadline failure", envelope)
	}
	if envelope.Phase != "brief" {
		t.Fatalf("phase = %q; want brief", envelope.Phase)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("elapsed = %s; want the short deadline", elapsed)
	}
	if len(envelope.Effects) != 0 {
		t.Fatalf("effects = %#v; want none", envelope.Effects)
	}
	if _, ok := counterValue(t, home); ok {
		t.Fatalf("a stalled brief took a counter value")
	}
	assertNoTaskRecord(t, home, "1-stalled")
}

func TestCreateBriefRejectsNonRegularPath(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	target := filepath.Join(t.TempDir(), "target.md")
	if err := os.WriteFile(target, []byte("target"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link.md")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "brief-dir")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		path string
		want string
	}{
		{"missing", filepath.Join(t.TempDir(), "missing.md"), "brief not found"},
		{"symbolic link", link, "not a regular file"},
		{"directory", dir, "not a regular file"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			envelope := Create(home, CreateOptions{
				Project:   "one",
				Title:     "First step",
				BriefPath: c.path,
				BriefSet:  true,
			})
			if envelope.Outcome != "failed" || envelope.Error == nil {
				t.Fatalf("Create() = %#v; want a failure", envelope)
			}
			if !strings.Contains(fmt.Sprint(envelope.Error), c.want) {
				t.Fatalf("error = %v; want %q", envelope.Error, c.want)
			}
			if _, ok := counterValue(t, home); ok {
				t.Fatalf("an invalid brief path took a counter value")
			}
		})
	}
}

func TestCreateFortyCharacterSlug(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	slug := strings.Repeat("a", 40)
	task := createOK(t, home, CreateOptions{Project: "one", Title: "Long slug", Slug: slug, SlugSet: true})
	if task.ID != "1-"+slug {
		t.Fatalf("ID = %q; want 1-%s", task.ID, slug)
	}
	if len(task.ID) != 42 {
		t.Fatalf("ID length = %d; want 42", len(task.ID))
	}
	if !records.ValidTaskID(task.ID) {
		t.Fatalf("ValidTaskID(%q) = false; want true", task.ID)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "tasks", task.ID+".json")); err != nil {
		t.Fatalf("task record for a 40-character slug: %v", err)
	}
}

func TestSlugRules(t *testing.T) {
	cases := []struct {
		name    string
		title   string
		slug    string
		slugSet bool
		want    string
		wantErr string
	}{
		{"generated from a title", "Spawn worker", "", false, "spawn-worker", ""},
		{"generated trims separators", "  Spawn  worker!  ", "", false, "spawn-worker", ""},
		{"generated drops a non-ASCII letter", "Café", "", false, "caf", ""},
		{"explicit a-b", "Any", "a-b", true, "a-b", ""},
		{"explicit uppercase and slash", "Any", "Bad/Slug", true, "", "invalid slug: Bad/Slug"},
		{"explicit leading hyphen", "Any", "-a", true, "", "invalid slug: -a"},
		{"explicit double hyphen", "Any", "a--b", true, "", "invalid slug: a--b"},
		{"forty-one characters", "Any", strings.Repeat("a", 41), true, "", "invalid slug: " + strings.Repeat("a", 41)},
		{"empty title", "   ", "", false, "", "invalid slug: "},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := newHome(t)
			registerProject(t, home, "one")
			envelope := Create(home, CreateOptions{
				Project: "one", Title: c.title, Slug: c.slug, SlugSet: c.slugSet,
			})
			if c.wantErr != "" {
				if envelope.Outcome != "failed" || envelope.Error != c.wantErr {
					t.Fatalf("Create() = %#v; want error %q", envelope, c.wantErr)
				}
				if _, ok := counterValue(t, home); ok {
					t.Fatalf("an invalid slug took a counter value")
				}
				return
			}
			if envelope.Outcome != "ok" {
				t.Fatalf("Create() = %#v; want ok", envelope)
			}
			task := envelope.Result.(records.Task)
			if task.ID != "1-"+c.want {
				t.Fatalf("ID = %q; want 1-%s", task.ID, c.want)
			}
		})
	}
}

func TestCreateMissingDependency(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	envelope := Create(home, CreateOptions{
		Project:    "one",
		Title:      "First step",
		DependsOn:  "9-missing",
		DependsSet: true,
	})
	if envelope.Outcome != "failed" || envelope.Error != "dependency not found: 9-missing" {
		t.Fatalf("Create() = %#v; want a missing dependency error", envelope)
	}
	if _, ok := counterValue(t, home); ok {
		t.Fatalf("a missing dependency took a counter value")
	}
}

func TestCreateDependencyOtherProject(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	registerProject(t, home, "two")
	other := createOK(t, home, CreateOptions{Project: "one", Title: "Other"})
	envelope := Create(home, CreateOptions{
		Project:    "two",
		Title:      "First step",
		DependsOn:  other.ID,
		DependsSet: true,
	})
	if envelope.Outcome != "failed" || envelope.Error != "dependency not in project two: "+other.ID {
		t.Fatalf("Create() = %#v; want a project mismatch error", envelope)
	}
	if value, ok := counterValue(t, home); !ok || value != 1 {
		t.Fatalf("counter = %d, %v; want 1", value, ok)
	}
}

func TestCreateInvalidDependencyStrings(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	for _, value := range []string{"../x", "a/b", "1-a,,2-b", "1-a,1-a", "bad\nid"} {
		t.Run(fmt.Sprintf("%q", value), func(t *testing.T) {
			envelope := Create(home, CreateOptions{
				Project:    "one",
				Title:      "First step",
				DependsOn:  value,
				DependsSet: true,
			})
			if envelope.Outcome != "failed" || envelope.Error == nil {
				t.Fatalf("Create(%q) = %#v; want a failure", value, envelope)
			}
			if _, ok := counterValue(t, home); ok {
				t.Fatalf("an invalid dependency list took a counter value")
			}
		})
	}
}

func TestCreatePartialBriefDirectory(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	briefPath := filepath.Join(home, ".hive", "state", "briefs", "1-first-step.md")
	if err := os.MkdirAll(briefPath, 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := Create(home, CreateOptions{Project: "one", Title: "First step"})
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial", envelope.Outcome)
	}
	if envelope.Phase != "brief" {
		t.Fatalf("phase = %q; want brief", envelope.Phase)
	}
	if envelope.Error == nil {
		t.Fatalf("error = nil; want one plain error line")
	}
	if len(envelope.Effects) != 1 || envelope.Effects[0].Resource != "counter" || envelope.Effects[0].Action != "updated" {
		t.Fatalf("effects = %#v; want only the counter effect", envelope.Effects)
	}
	if value, ok := counterValue(t, home); !ok || value != 1 {
		t.Fatalf("counter = %d, %v; want the gap at 1", value, ok)
	}
	assertNoTaskRecord(t, home, "1-first-step")
}

func TestCreatePartialTaskDirectory(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	taskPath := filepath.Join(home, ".hive", "state", "tasks", "1-first-step.json")
	if err := os.MkdirAll(taskPath, 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := Create(home, CreateOptions{Project: "one", Title: "First step"})
	if envelope.Outcome != "partial" {
		t.Fatalf("outcome = %q; want partial", envelope.Outcome)
	}
	if envelope.Phase != "task" {
		t.Fatalf("phase = %q; want task", envelope.Phase)
	}
	if envelope.Error == nil {
		t.Fatalf("error = nil; want one plain error line")
	}
	if len(envelope.Effects) != 2 {
		t.Fatalf("effects = %#v; want the counter and the brief", envelope.Effects)
	}
	if value, ok := counterValue(t, home); !ok || value != 1 {
		t.Fatalf("counter = %d, %v; want 1", value, ok)
	}
	brief, err := os.ReadFile(filepath.Join(home, ".hive", "state", "briefs", "1-first-step.md"))
	if err != nil {
		t.Fatalf("orphan brief: %v", err)
	}
	if string(brief) != "First step" {
		t.Fatalf("orphan brief = %q; want the title", brief)
	}
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Task
	if err := registry.ReadIn([]string{"tasks"}, "1-first-step", &record); err == nil {
		t.Fatalf("ReadIn() = nil; want a failure for the directory at the task path")
	}
}

func TestCreateConcurrent(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	start := make(chan struct{})
	envelopes := make([]output.Envelope, 2)
	var wait sync.WaitGroup
	for index := range envelopes {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			envelopes[index] = Create(home, CreateOptions{
				Project: "one",
				Title:   fmt.Sprintf("Task %d", index+1),
			})
		}(index)
	}
	close(start)
	wait.Wait()

	ids := map[string]bool{}
	for index, envelope := range envelopes {
		if envelope.Outcome != "ok" || envelope.Error != nil {
			t.Fatalf("concurrent Create %d = %#v; want ok", index, envelope)
		}
		task := envelope.Result.(records.Task)
		if ids[task.ID] {
			t.Fatalf("duplicate ID %q from two creates", task.ID)
		}
		ids[task.ID] = true
	}
	if value, ok := counterValue(t, home); !ok || value != 2 {
		t.Fatalf("counter = %d, %v; want 2", value, ok)
	}
}

func TestCycle(t *testing.T) {
	home := newHome(t)
	registerProject(t, home, "one")
	first := createOK(t, home, CreateOptions{Project: "one", Title: "A"})
	second := createOK(t, home, CreateOptions{
		Project:    "one",
		Title:      "B",
		DependsOn:  first.ID,
		DependsSet: true,
	})
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	first.DependsOn = []string{second.ID}
	if err := registry.WriteIn([]string{"tasks"}, first.ID, first); err != nil {
		t.Fatal(err)
	}
	envelope := Create(home, CreateOptions{
		Project:    "one",
		Title:      "C",
		DependsOn:  first.ID,
		DependsSet: true,
	})
	want := "dependency cycle: 1-a -> 2-b -> 1-a"
	if envelope.Outcome != "failed" || envelope.Error != want {
		t.Fatalf("Create() = %#v; want %q", envelope, want)
	}
	if value, ok := counterValue(t, home); !ok || value != 2 {
		t.Fatalf("counter = %d, %v; want 2", value, ok)
	}
	assertNoTaskRecord(t, home, "3-c")
}
