package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	// briefLimit is the largest brief that readBrief accepts.
	briefLimit = 1 << 20
	// briefDeadline is the default total deadline of one brief read.
	briefDeadline = 30 * time.Second
	// maxSlugLength is the largest slug.
	maxSlugLength = 40
	// briefDirName is the state directory of the briefs.
	briefDirName = "briefs"
	// tasksCollection is the task collection below the state root.
	tasksCollection = "tasks"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// CreateOptions holds the values for one task creation.
type CreateOptions struct {
	Project    string
	Title      string
	Slug       string
	SlugSet    bool
	DependsOn  string
	DependsSet bool
	BriefPath  string
	BriefSet   bool
	Input      io.Reader
	Deadline   time.Duration
	Logger     *slog.Logger
}

// createPlan holds the validated values that the locked part uses.
type createPlan struct {
	project string
	title   string
	slug    string
	deps    []string
	brief   []byte
	now     time.Time
}

// Create makes one task, its brief, and its counter effect.
func Create(home string, options CreateOptions) output.Envelope {
	log := openOperationLog(home, options.Logger)
	defer log.close()
	fail := func(phase, message string) output.Envelope {
		log.failure("task create", options.Project, "", "", phase, message)
		return failedEnvelope(phase, message)
	}
	if options.Project == "" {
		return fail("validate", "project requires a value")
	}
	slug := options.Slug
	if !options.SlugSet {
		slug = slugify(options.Title)
	}
	if err := checkSlug(slug); err != nil {
		return fail("validate", err.Error())
	}
	deps, err := parseDeps(options.DependsOn)
	if err != nil {
		return fail("validate", err.Error())
	}
	brief := []byte(options.Title)
	if options.BriefSet {
		deadline := options.Deadline
		if deadline <= 0 {
			deadline = briefDeadline
		}
		brief, err = readBrief(options.BriefPath, options.Input, deadline)
		if err != nil {
			return fail("brief", err.Error())
		}
	}

	registry, err := store.Open(home)
	if err != nil {
		return fail("store", err.Error())
	}
	plan := createPlan{
		project: options.Project,
		title:   options.Title,
		slug:    slug,
		deps:    deps,
		brief:   brief,
		now:     time.Now().UTC(),
	}
	var envelope output.Envelope
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		envelope = createLocked(locked, home, plan, log)
		return nil
	})
	if lockErr != nil {
		return fail("store", lockErr.Error())
	}
	return envelope
}

// createLocked runs the validation reads, the counter, the brief write, and
// the task write under one lock. It reports each completed effect.
func createLocked(locked *store.LockedStore, home string, plan createPlan, log operationLog) output.Envelope {
	fail := func(phase, message, taskID string) output.Envelope {
		log.failure("task create", plan.project, taskID, "", phase, message)
		return failedEnvelope(phase, message)
	}
	if err := checkBriefDir(home); err != nil {
		return fail("brief", err.Error(), "")
	}
	var project records.Project
	if err := locked.ReadIn([]string{"projects"}, plan.project, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fail("project", "project not found: "+quoteIfControl(plan.project), "")
		}
		return fail("project", err.Error(), "")
	}
	for _, dep := range plan.deps {
		var dependency records.Task
		if err := locked.ReadIn([]string{tasksCollection}, dep, &dependency); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fail("dependency", "dependency not found: "+quoteIfControl(dep), "")
			}
			return fail("dependency", err.Error(), "")
		}
		if dependency.Project != project.ID {
			return fail("dependency", fmt.Sprintf("dependency not in project %s: %s", quoteIfControl(project.ID), quoteIfControl(dep)), "")
		}
	}
	nextValue, err := peekCounter(locked, "task")
	if err != nil {
		return fail("counter", err.Error(), "")
	}
	prospectiveID := fmt.Sprintf("%d-%s", nextValue, plan.slug)
	if err := checkCycle(locked, prospectiveID, plan.deps); err != nil {
		return fail("dependency", err.Error(), "")
	}

	number, err := locked.Next("task")
	if err != nil {
		return fail("counter", err.Error(), "")
	}
	id := fmt.Sprintf("%d-%s", number, plan.slug)
	log.effect("task create", project.ID, id, "counter", counterRelPath("task"), "updated")
	envelope := output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: []output.Effect{
			{Resource: "counter", Path: counterRelPath("task"), Action: "updated"},
		},
		Error: nil,
	}
	dependsOn := plan.deps
	if dependsOn == nil {
		dependsOn = []string{}
	}
	record := records.Task{
		ID:              id,
		Project:         project.ID,
		Title:           plan.title,
		ControllingRole: "manager-" + project.ID,
		ActingAgent:     "",
		Status:          "created",
		DependsOn:       dependsOn,
		CreatedAt:       plan.now.Format(time.RFC3339Nano),
		UpdatedAt:       plan.now.Format(time.RFC3339Nano),
	}
	if err := writeBrief(home, id, plan.brief); err != nil {
		message := oneLine(err.Error())
		envelope.Outcome = "partial"
		envelope.Phase = "brief"
		envelope.Error = message
		log.failure("task create", project.ID, id, "", "brief", message)
		return envelope
	}
	log.effect("task create", project.ID, id, "brief", briefRelPath(id), "created")
	envelope.Effects = append(envelope.Effects, output.Effect{
		Resource: "brief",
		Path:     briefRelPath(id),
		Action:   "created",
	})
	if err := locked.WriteIn([]string{tasksCollection}, id, record); err != nil {
		message := oneLine(err.Error())
		envelope.Outcome = "partial"
		envelope.Phase = "task"
		envelope.Error = message
		log.failure("task create", project.ID, id, "", "task", message)
		return envelope
	}
	log.effect("task create", project.ID, id, "task", taskRelPath(id), "created")
	envelope.Effects = append(envelope.Effects, output.Effect{
		Resource: "task",
		Path:     taskRelPath(id),
		Action:   "created",
	})
	envelope.Result = record
	log.transition("task create", project.ID, id, "", "", record.Status)
	return envelope
}

// peekCounter reads the current counter value under the lock and returns the
// value that the next increment gives.
func peekCounter(locked *store.LockedStore, name string) (uint64, error) {
	var counter struct {
		Value uint64 `json:"value"`
	}
	err := locked.ReadIn([]string{"counters"}, name, &counter)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}
	return counter.Value + 1, nil
}

// checkCycle fails when the dependency graph that starts at start has a cycle.
func checkCycle(locked *store.LockedStore, start string, deps []string) error {
	onPath := map[string]bool{start: true}
	path := []string{start}
	var walk func(edges []string) error
	walk = func(edges []string) error {
		for _, dep := range edges {
			if onPath[dep] {
				index := slices.Index(path, dep)
				cycle := append(append([]string{}, path[index:]...), dep)
				return fmt.Errorf("dependency cycle: %s", strings.Join(cycle, " -> "))
			}
			var dependency records.Task
			if err := locked.ReadIn([]string{tasksCollection}, dep, &dependency); err != nil {
				continue
			}
			path = append(path, dep)
			onPath[dep] = true
			if err := walk(dependency.DependsOn); err != nil {
				return err
			}
			path = path[:len(path)-1]
			onPath[dep] = false
		}
		return nil
	}
	return walk(deps)
}

// readBrief reads one brief with a byte cap and a total deadline. A path that
// is not a regular file fails before the open.
func readBrief(path string, in io.Reader, deadline time.Duration) ([]byte, error) {
	if deadline <= 0 {
		deadline = briefDeadline
	}
	if path == "-" {
		if in == nil {
			return nil, errors.New("read brief: no input")
		}
		return readBounded(in, deadline)
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("brief not found: %s", quoteIfControl(path))
	}
	if err != nil {
		return nil, fmt.Errorf("read brief: %s", oneLine(err.Error()))
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("brief is not a regular file: %s", quoteIfControl(path))
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read brief: %s", oneLine(err.Error()))
	}
	defer func() { _ = file.Close() }()
	return readBounded(file, deadline)
}

// readBounded reads at most briefLimit bytes before deadline.
func readBounded(reader io.Reader, deadline time.Duration) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		data, err := io.ReadAll(io.LimitReader(reader, briefLimit+1))
		done <- result{data: data, err: err}
	}()
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	select {
	case res := <-done:
		if res.err != nil {
			return nil, fmt.Errorf("read brief: %s", oneLine(res.err.Error()))
		}
		if len(res.data) > briefLimit {
			return nil, fmt.Errorf("brief too large: %d bytes", len(res.data))
		}
		return res.data, nil
	case <-timer.C:
		return nil, fmt.Errorf("read brief: deadline exceeded after %s", deadline)
	}
}

// writeBrief writes one brief atomically. It is the only writer of a brief path.
func writeBrief(home, id string, data []byte) error {
	if err := checkBriefDir(home); err != nil {
		return err
	}
	path := briefPath(home, id)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("write brief: %s", oneLine(err.Error()))
	}
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("write brief: %s", oneLine(err.Error()))
	}
	if err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular()) {
		return fmt.Errorf("brief path invalid: %s", quoteIfControl(path))
	}
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return fmt.Errorf("write brief: %s", oneLine(err.Error()))
	}
	temp := file.Name()
	defer func() { _ = os.Remove(temp) }()
	if err := file.Chmod(0o644); err != nil {
		_ = file.Close()
		return fmt.Errorf("write brief: %s", oneLine(err.Error()))
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write brief: %s", oneLine(err.Error()))
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("write brief: %s", oneLine(err.Error()))
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("write brief: %s", oneLine(err.Error()))
	}
	if err := os.Rename(temp, path); err != nil {
		return fmt.Errorf("write brief: %s", oneLine(err.Error()))
	}
	return nil
}

// checkBriefDir rejects a static symbolic link or a wrong type in the
// computed brief ancestors. A missing directory is valid because writeBrief
// creates it.
func checkBriefDir(home string) error {
	current := home
	for _, name := range []string{".hive", "state", briefDirName} {
		current = filepath.Join(current, name)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("brief path invalid: %s", quoteIfControl(current))
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("brief path invalid: %s", quoteIfControl(current))
		}
	}
	return nil
}

// checkSlug checks the slug rule.
func checkSlug(slug string) error {
	if len(slug) > maxSlugLength || !slugPattern.MatchString(slug) {
		return fmt.Errorf("invalid slug: %s", quoteIfControl(slug))
	}
	return nil
}

// parseDeps checks and splits the dependency list.
func parseDeps(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return []string{}, nil
	}
	seen := map[string]bool{}
	deps := []string{}
	for _, part := range strings.Split(value, ",") {
		id := strings.TrimSpace(part)
		if !records.ValidTaskID(id) {
			return nil, fmt.Errorf("invalid dependency: %s", quoteIfControl(id))
		}
		if seen[id] {
			return nil, fmt.Errorf("duplicate dependency: %s", quoteIfControl(id))
		}
		seen[id] = true
		deps = append(deps, id)
	}
	return deps, nil
}

// slugify turns a title into a candidate slug.
func slugify(title string) string {
	var builder strings.Builder
	separator := false
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if separator && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			separator = false
			builder.WriteRune(r)
			continue
		}
		separator = true
	}
	return builder.String()
}

// quoteIfControl quotes a value that holds a control character.
func quoteIfControl(value string) string {
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Sprintf("%q", value)
		}
	}
	return value
}

// oneLine replaces each control character with a space for one-line output.
func oneLine(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
}

// briefPath returns the absolute path of one brief.
func briefPath(home, id string) string {
	return filepath.Join(home, ".hive", "state", briefDirName, id+".md")
}

// briefRelPath returns the home-relative path of one brief.
func briefRelPath(id string) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", briefDirName, id+".md"))
}

// taskRelPath returns the home-relative path of one task record.
func taskRelPath(id string) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", tasksCollection, id+".json"))
}

// counterRelPath returns the home-relative path of one counter.
func counterRelPath(name string) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", "counters", name+".json"))
}
