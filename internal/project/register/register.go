// Package register validates and records one project repository.
package register

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	projectlib "github.com/Harrison-Blair/hive/internal/project"
)

// Options holds the values for one project registration.
type Options struct {
	Path       string
	ID         string
	IDSet      bool
	Base       string
	BaseSet    bool
	WorkingDir string
}

// Run validates and records one Git repository.
func Run(home string, options Options) output.Envelope {
	if projectlib.HasControl(options.Path) {
		return output.Failed("validate", fmt.Sprintf("invalid path: %q", options.Path))
	}
	if options.Path == "" {
		return output.Failed("validate", "PATH requires a value")
	}
	base := options.Base
	if !options.BaseSet {
		base = "main"
	}
	if invalidBase(base) {
		return output.Failed("validate", fmt.Sprintf("invalid base branch: %q", base))
	}
	path, err := resolveProjectPath(options.Path, options.WorkingDir)
	if err != nil || !isGitRepository(path) {
		return output.Failed("validate", fmt.Sprintf("not a git repository: %s", path))
	}
	id := options.ID
	if !options.IDSet {
		id = strings.ToLower(filepath.Base(path))
	}
	if !records.ValidID(id) {
		return output.Failed("validate", projectlib.InvalidIDMessage(id))
	}

	registry, err := store.Open(home)
	if err != nil {
		return output.Failed("store", err.Error())
	}
	found, err := projectlib.List(registry)
	if err != nil {
		return output.Failed("store", err.Error())
	}
	for _, existing := range found {
		if existing.ID == id {
			return output.Failed("register", "project exists: "+id)
		}
		if existing.Path == path {
			return output.Failed("register", "path registered as "+existing.ID)
		}
	}

	record := records.Project{
		ID:           id,
		Path:         path,
		Base:         base,
		RegisteredAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := projectlib.Write(registry, record); err != nil {
		return output.Failed("store", err.Error())
	}
	envelope := output.Envelope{
		Outcome: output.OutcomeOK,
		Phase:   "complete",
		Effects: []output.Effect{{Resource: "record", Path: filepath.ToSlash(projectlib.RecordPath(id)), Action: "created"}},
		Error:   nil,
		Result:  record,
	}
	for _, dir := range []string{
		filepath.Join("projects", id, "worktrees"),
	} {
		fullPath := filepath.Join(home, dir)
		action := "created"
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			action = "exists"
		}
		if err := os.MkdirAll(fullPath, 0o755); err != nil {
			envelope.Outcome = output.OutcomePartial
			envelope.Phase = "dirs"
			envelope.Error = fmt.Sprintf("create directory %s: %v", filepath.ToSlash(dir), err)
			return envelope
		}
		envelope.Effects = append(envelope.Effects, output.Effect{
			Resource: "directory",
			Path:     filepath.ToSlash(dir),
			Action:   action,
		})
	}
	return envelope
}

func resolveProjectPath(path, workingDir string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(workingDir, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path), err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return filepath.Clean(absolute), err
	}
	return resolved, nil
}

func isGitRepository(path string) bool {
	info, err := os.Lstat(filepath.Join(path, ".git"))
	return err == nil && (info.IsDir() || info.Mode().IsRegular())
}

func invalidBase(base string) bool {
	if base == "" || strings.HasPrefix(base, "-") {
		return true
	}
	for _, r := range base {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// Text renders the effects of project registration.
func Text(envelope output.Envelope) string {
	var lines []string
	for _, effect := range envelope.Effects {
		lines = append(lines, fmt.Sprintf("%s %s %s", effect.Action, effect.Resource, effect.Path))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
