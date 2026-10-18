package task

import (
	"io"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

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
	_ = home
	_ = options
	return output.Envelope{}
}

// createLocked runs the validation reads, the counter, the brief write, and
// the task write under one lock.
func createLocked(locked *store.LockedStore, home string, plan createPlan) output.Envelope {
	_, _, _, _ = locked, home, plan, 0
	return output.Envelope{}
}

// readBrief reads one brief with a byte cap and a total deadline.
func readBrief(path string, in io.Reader, deadline time.Duration) ([]byte, error) {
	_, _, _ = path, in, deadline
	return nil, nil
}

// writeBrief writes one brief atomically.
func writeBrief(home, id string, data []byte) error {
	_, _, _ = home, id, data
	return nil
}

// checkSlug checks the slug rule.
func checkSlug(slug string) error {
	_ = slug
	return nil
}

// parseDeps checks and splits the dependency list.
func parseDeps(value string) ([]string, error) {
	_ = value
	return nil, nil
}

// checkCycle checks the dependency graph for a cycle.
func checkCycle(locked *store.LockedStore, start string, deps []string) error {
	_, _, _ = locked, start, deps
	return nil
}

// slugify turns a title into a candidate slug.
func slugify(title string) string {
	_ = title
	return ""
}
