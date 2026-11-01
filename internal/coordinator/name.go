package coordinator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/herdr"
)

// nameSuffix ends the name of a coordinator: <dir>-coordinator.
const nameSuffix = "-" + Name

// maxTries bounds the numbers that ResolveName tries for a free name.
const maxTries = 1000

// ResolveName returns the coordinator name for the start directory dir and
// the pane: Slug of the base name of dir, then -coordinator, at most 64
// bytes. An empty slug gives Name. When the record of that name belongs to
// another pane that herdr still has, it tries <slug>-2-coordinator,
// <slug>-3-coordinator, and so on, or coordinator-2 and so on for an empty
// slug.
func ResolveName(home, dir, pane string, logger *slog.Logger) (string, error) {
	base := filepath.Base(filepath.Clean(dir))
	for number := 1; number <= maxTries; number++ {
		name := candidate(base, number)
		taken, err := takenBy(home, name, pane, logger)
		if err != nil {
			return "", err
		}
		if taken == "" {
			return name, nil
		}
		logger.Info("coordinator name taken", "name", name, "pane_id", taken)
	}
	return "", fmt.Errorf("no free coordinator name for %s after %d tries", dir, maxTries)
}

// candidate returns the coordinator name of base with number. Number 1 has
// no number in the name.
func candidate(base string, number int) string {
	tag := ""
	if number > 1 {
		tag = "-" + strconv.Itoa(number)
	}
	slug := agent.Slug(base, agent.MaxNameLen-len(tag)-len(nameSuffix))
	if slug == "" {
		return Name + tag
	}
	return slug + tag + nameSuffix
}

// takenBy returns the pane of a live agent record of name in another pane,
// or an empty string when the name is free for pane.
func takenBy(home, name, pane string, logger *slog.Logger) (string, error) {
	record, err := agent.Load(home, name)
	if errors.Is(err, agent.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if record.Pane == pane {
		return "", nil
	}
	_, err = herdr.PaneGet(context.Background(), record.Pane, herdr.WithLogger(logger))
	if agent.IsCode(err, herdr.CodePaneNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("check agent %s: %w", name, err)
	}
	return record.Pane, nil
}

// isCoordinatorName reports whether name has the form of a coordinator name:
// Name, Name-N, or ...-coordinator.
func isCoordinatorName(name string) bool {
	if number, ok := strings.CutPrefix(name, Name+"-"); ok {
		if _, err := strconv.Atoi(number); err == nil {
			return true
		}
	}
	return name == Name || strings.HasSuffix(name, nameSuffix)
}

// checkName fails when the record of name belongs to another pane that herdr
// still has. A record of pane, or of a pane that is gone, can be replaced.
func checkName(home, name, pane string, logger *slog.Logger) error {
	taken, err := takenBy(home, name, pane, logger)
	if err != nil {
		return err
	}
	if taken != "" {
		return fmt.Errorf("agent %s exists in pane %s: stop it or close that pane first", name, taken)
	}
	return nil
}

// removeOldCoordinators removes each coordinator record of pane other than
// name, so the pane has one record.
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
		if record.Pane != pane || record.Name == name || !isCoordinatorName(record.Name) {
			continue
		}
		if err := agent.Remove(home, record.Name); err != nil {
			logger.Warn("old coordinator record remove failed", "record", record.Name, "error", err.Error())
			continue
		}
		logger.Info("old coordinator record removed", "record", record.Name, "pane", pane)
	}
}
