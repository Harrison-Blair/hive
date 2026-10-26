// Package create makes one task record, its brief, and its counter effect.
package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	tasklib "github.com/Harrison-Blair/hive/internal/task"
)

// Options holds the values for one task creation.
type Options struct {
	Project        string
	Title          string
	Feature        string
	RequireFeature bool
	Spec           bool
	Slug           string
	SlugSet        bool
	DependsOn      string
	DependsSet     bool
	BriefPath      string
	BriefSet       bool
	Input          io.Reader
	Deadline       time.Duration
	Logger         *slog.Logger
}

// createPlan holds the validated values that the locked part uses.
type createPlan struct {
	project string
	title   string
	feature string
	spec    bool
	slug    string
	deps    []string
	brief   []byte
	now     time.Time
}

// Run makes one task, its brief, and its counter effect.
func Run(home string, options Options) output.Envelope {
	log := tasklib.OpenLog(home, options.Logger)
	defer log.Close()
	fail := func(phase, message string) output.Envelope {
		log.Failure("task create", options.Project, "", "", phase, message)
		return tasklib.FailedEnvelope(phase, message)
	}
	if options.Project == "" {
		return fail("validate", "project requires a value")
	}
	if options.RequireFeature || options.Feature != "" || options.Spec {
		if err := tasklib.CheckFeature(options.Feature); err != nil {
			return fail("validate", err.Error())
		}
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
			deadline = tasklib.BriefDeadline
		}
		brief, err = tasklib.ReadBrief(options.BriefPath, options.Input, deadline)
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
		feature: options.Feature,
		spec:    options.Spec,
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
func createLocked(locked *store.LockedStore, home string, plan createPlan, log tasklib.Log) output.Envelope {
	fail := func(phase, message, taskID string) output.Envelope {
		log.Failure("task create", plan.project, taskID, "", phase, message)
		return tasklib.FailedEnvelope(phase, message)
	}
	if err := tasklib.CheckBriefDir(home); err != nil {
		return fail("brief", err.Error(), "")
	}
	var project records.Project
	if err := locked.ReadIn([]string{"projects"}, plan.project, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fail("project", "project not found: "+tasklib.QuoteIfControl(plan.project), "")
		}
		return fail("project", err.Error(), "")
	}
	if plan.feature != "" && !plan.spec {
		if err := checkFeatureREADME(home, plan.feature); err != nil {
			return fail("validate", err.Error(), "")
		}
	}
	for _, dep := range plan.deps {
		var dependency records.Task
		if err := locked.ReadIn([]string{tasklib.TasksCollection}, dep, &dependency); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fail("dependency", "dependency not found: "+tasklib.QuoteIfControl(dep), "")
			}
			return fail("dependency", err.Error(), "")
		}
		if dependency.Project != project.ID {
			return fail("dependency", fmt.Sprintf("dependency not in project %s: %s", tasklib.QuoteIfControl(project.ID), tasklib.QuoteIfControl(dep)), "")
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
	log.Effect("task create", project.ID, id, "counter", counterRelPath("task"), "updated")
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
		ID:                  id,
		Project:             project.ID,
		Title:               plan.title,
		Feature:             plan.feature,
		Spec:                plan.spec,
		ActingAgent:         "",
		Status:              "created",
		DependsOn:           dependsOn,
		VerificationHistory: []records.VerificationHistory{},
		CreatedAt:           plan.now.Format(time.RFC3339Nano),
		UpdatedAt:           plan.now.Format(time.RFC3339Nano),
	}
	if err := tasklib.WriteBrief(home, id, plan.brief); err != nil {
		message := tasklib.OneLine(err.Error())
		envelope.Outcome = "partial"
		envelope.Phase = "brief"
		envelope.Error = message
		log.Failure("task create", project.ID, id, "", "brief", message)
		return envelope
	}
	log.Effect("task create", project.ID, id, "brief", tasklib.BriefRelPath(id), "created")
	envelope.Effects = append(envelope.Effects, output.Effect{
		Resource: "brief",
		Path:     tasklib.BriefRelPath(id),
		Action:   "created",
	})
	if paths := tasklib.TaskFeaturePaths(record); paths != nil {
		if err := tasklib.WriteWorktreeBrief(home, paths.Brief, plan.brief); err != nil {
			message := "write feature brief: " + tasklib.OneLine(err.Error())
			envelope.Outcome = "partial"
			envelope.Phase = "brief"
			envelope.Error = message
			log.Failure("task create", project.ID, id, "", "brief", message)
			return envelope
		}
		log.Effect("task create", project.ID, id, "brief", paths.Brief, "created")
		envelope.Effects = append(envelope.Effects, output.Effect{Resource: "brief", Path: paths.Brief, Action: "created"})
	}
	if err := locked.WriteIn([]string{tasklib.TasksCollection}, id, record); err != nil {
		message := tasklib.OneLine(err.Error())
		envelope.Outcome = "partial"
		envelope.Phase = "task"
		envelope.Error = message
		log.Failure("task create", project.ID, id, "", "task", message)
		return envelope
	}
	log.Effect("task create", project.ID, id, "task", tasklib.TaskRelPath(id), "created")
	envelope.Effects = append(envelope.Effects, output.Effect{
		Resource: "task",
		Path:     tasklib.TaskRelPath(id),
		Action:   "created",
	})
	envelope.Result = record
	log.Transition("task create", project.ID, id, "", "", record.Status)
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
			if err := locked.ReadIn([]string{tasklib.TasksCollection}, dep, &dependency); err != nil {
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

// checkSlug checks the slug rule.
func checkSlug(slug string) error {
	if !tasklib.ValidSlug(slug) {
		return fmt.Errorf("invalid slug: %s", tasklib.QuoteIfControl(slug))
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
			return nil, fmt.Errorf("invalid dependency: %s", tasklib.QuoteIfControl(id))
		}
		if seen[id] {
			return nil, fmt.Errorf("duplicate dependency: %s", tasklib.QuoteIfControl(id))
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

// counterRelPath returns the home-relative path of one counter.
func counterRelPath(name string) string {
	return filepath.ToSlash(filepath.Join(".hive", "state", "counters", name+".json"))
}

// featureREADMEPath returns the absolute path of the README of one feature in
// the home.
func featureREADMEPath(home, feature string) string {
	return filepath.Join(home, ".hive", "spec", feature, "README.md")
}

// checkFeatureREADME checks that the home has the README of the feature.
func checkFeatureREADME(home, feature string) error {
	path := featureREADMEPath(home, feature)
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("feature README missing in home: %s", path)
	}
	if err != nil {
		return fmt.Errorf("check feature README %s: %s", path, tasklib.OneLine(err.Error()))
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("feature README is not a regular file: %s", path)
	}
	return nil
}

// Text renders the effects of task creation and the new ID.
func Text(envelope output.Envelope) string {
	if record, ok := envelope.Result.(records.Task); ok && envelope.Error == nil {
		return fmt.Sprintf("id: %s\nbrief: %s\n", record.ID, tasklib.BriefRelPath(record.ID))
	}
	return tasklib.EffectsText(envelope.Effects)
}
