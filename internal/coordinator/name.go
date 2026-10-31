package coordinator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
	projectlib "github.com/Harrison-Blair/hive/internal/project"
)

// nameSuffix ends the name of a project coordinator: <project-id>-coordinator.
const nameSuffix = "-" + Name

// ResolveName returns <id>-coordinator when dir is in the checkout of the
// registered project id, and Name otherwise. It compares the paths after it
// resolves symlinks. When the checkouts of several projects hold dir, the
// innermost checkout wins. When two projects have that same checkout, or the
// registry is unreadable, it returns Name and an error with the cause.
func ResolveName(home, dir string) (string, error) {
	start, err := canonical(dir)
	if err != nil {
		return Name, fmt.Errorf("resolve %s: %w", dir, err)
	}
	registry, err := store.Open(home)
	if err != nil {
		return Name, err
	}
	projects, err := projectlib.List(registry)
	if err != nil {
		return Name, err
	}
	var matches []string
	longest := -1
	for _, project := range projects {
		root, err := canonical(project.Path)
		if err != nil || !within(start, root) {
			continue
		}
		switch {
		case len(root) > longest:
			matches, longest = []string{project.ID}, len(root)
		case len(root) == longest:
			matches = append(matches, project.ID)
		}
	}
	switch {
	case len(matches) == 0:
		return Name, nil
	case len(matches) > 1:
		return Name, fmt.Errorf("projects %s have the checkout of %s", strings.Join(matches, ", "), start)
	case !records.ValidID(matches[0]):
		return Name, errors.New(projectlib.InvalidIDMessage(matches[0]))
	}
	return matches[0] + nameSuffix, nil
}

// canonical returns the absolute path of path with each symlink resolved.
func canonical(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

// within reports whether the clean path is root or below root.
func within(path, root string) bool {
	return path == root || strings.HasPrefix(path, strings.TrimSuffix(root, string(filepath.Separator))+string(filepath.Separator))
}

// isCoordinatorName reports whether name is Name or <project-id>-coordinator.
func isCoordinatorName(name string) bool {
	id, ok := strings.CutSuffix(name, nameSuffix)
	return name == Name || ok && records.ValidID(id)
}

// checkName fails when the record of name belongs to another pane that herdr
// still has. A record of pane, or of a pane that is gone, can be replaced.
func checkName(home, name, pane string, logger *slog.Logger) error {
	record, err := agent.Load(home, name)
	if errors.Is(err, agent.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if record.Pane == pane {
		return nil
	}
	_, err = herdr.PaneGet(context.Background(), record.Pane, herdr.WithLogger(logger))
	if agent.IsCode(err, herdr.CodePaneNotFound) {
		logger.Info("coordinator record replaced", "pane_id", record.Pane)
		return nil
	}
	if err != nil {
		return fmt.Errorf("check agent %s: %w", name, err)
	}
	return fmt.Errorf("agent %s exists in pane %s: stop it or close that pane first", name, record.Pane)
}

// removeOldCoordinators removes each coordinator record of pane other than
// name, so the pane has one record. It keeps a record with a worktree, because
// agent stop needs it to remove the worktree.
func removeOldCoordinators(home, name, pane string, logger *slog.Logger) {
	if pane == "" {
		return
	}
	all, err := agent.List(home)
	if err != nil {
		logger.Warn("old coordinator records unreadable", "error", err.Error())
		return
	}
	for _, record := range all {
		if record.Pane != pane || record.Name == name || record.Worktree != "" || !isCoordinatorName(record.Name) {
			continue
		}
		if err := agent.Remove(home, record.Name); err != nil {
			logger.Warn("old coordinator record remove failed", "record", record.Name, "error", err.Error())
			continue
		}
		logger.Info("old coordinator record removed", "record", record.Name, "pane", pane)
	}
}
