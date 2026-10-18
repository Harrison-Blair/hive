package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// RegisterOptions holds the values for one project registration.
type RegisterOptions struct {
	Path       string
	ID         string
	IDSet      bool
	Base       string
	BaseSet    bool
	WorkingDir string
}

// Register validates and records one Git repository.
func Register(home string, options RegisterOptions) output.Envelope {
	if hasControl(options.Path) {
		return failed("validate", fmt.Sprintf("invalid path: %q", options.Path))
	}
	if options.Path == "" {
		return failed("validate", "PATH requires a value")
	}
	base := options.Base
	if !options.BaseSet {
		base = "main"
	}
	if invalidBase(base) {
		return failed("validate", fmt.Sprintf("invalid base branch: %q", base))
	}
	path, err := resolveProjectPath(options.Path, options.WorkingDir)
	if err != nil || !isGitRepository(path) {
		return failed("validate", fmt.Sprintf("not a git repository: %s", path))
	}
	id := options.ID
	if !options.IDSet {
		id = strings.ToLower(filepath.Base(path))
	}
	if !records.ValidID(id) {
		if hasControl(id) {
			id = fmt.Sprintf("%q", id)
		}
		return failed("validate", fmt.Sprintf("invalid project id: %s", id))
	}

	registry, err := store.Open(home)
	if err != nil {
		return failed("store", err.Error())
	}
	found, err := registry.List("projects")
	if err != nil {
		return failed("store", err.Error())
	}
	for _, raw := range found {
		var existing records.Project
		if err := json.Unmarshal(raw, &existing); err != nil {
			return failed("store", err.Error())
		}
		if existing.ID == id {
			return failed("register", "project exists: "+id)
		}
		if existing.Path == path {
			return failed("register", "path registered as "+existing.ID)
		}
	}

	record := records.Project{
		ID:           id,
		Path:         path,
		Base:         base,
		RegisteredAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := registry.Write("projects", id, record); err != nil {
		return failed("store", err.Error())
	}
	envelope := output.Envelope{
		Outcome: "ok",
		Phase:   "complete",
		Effects: []output.Effect{{Resource: "record", Path: filepath.ToSlash(filepath.Join(".hive", "state", "projects", id+".json")), Action: "created"}},
		Error:   nil,
		Result:  record,
	}
	for _, dir := range []string{
		filepath.Join("projects", id),
		filepath.Join("projects", id, "worktrees"),
		filepath.Join("projects", id, "overrides"),
	} {
		fullPath := filepath.Join(home, dir)
		action := "created"
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			action = "exists"
		}
		if err := os.MkdirAll(fullPath, 0o755); err != nil {
			envelope.Outcome = "partial"
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

func hasControl(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func failed(phase, message string) output.Envelope {
	return output.Envelope{Outcome: "failed", Phase: phase, Effects: []output.Effect{}, Error: message, Result: nil}
}
