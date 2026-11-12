package worktree

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/agent"
)

// projectSlugMax is the maximum length in bytes of the project part of a
// scratch place.
const projectSlugMax = 64

// Scratch makes the scratch place <home>/tmp/<project>/ of the project of dir
// when it is missing. The project is the slug of the base name of the
// directory that holds the common Git directory of dir, or of dir itself
// outside Git. Scratch returns the absolute path and whether it made the
// place.
func Scratch(ctx context.Context, home, dir string, logger *slog.Logger) (string, bool, error) {
	logger = agent.Logger(logger)
	project := filepath.Base(dir)
	common, err := gitRun(ctx, logger, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	switch {
	case err == nil:
		project = filepath.Base(filepath.Dir(common))
	case !strings.Contains(err.Error(), "not a git repository"):
		return "", false, err
	}
	path := filepath.Join(home, "tmp", agent.Slug(project, projectSlugMax))
	_, err = os.Stat(path)
	created := errors.Is(err, fs.ErrNotExist)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", false, err
	}
	if created {
		logger.Info("scratch place created", "component", "scratch", "path", path)
	}
	return path, created, nil
}
