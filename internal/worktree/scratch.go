package worktree

import (
	"context"
	"errors"
	"fmt"
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", false, err
	}
	err = os.Mkdir(path, 0o755)
	created := err == nil
	if errors.Is(err, fs.ErrExist) {
		var info fs.FileInfo
		info, err = os.Stat(path)
		if err == nil && !info.IsDir() {
			err = fmt.Errorf("%s is not a directory", path)
		}
	}
	if err != nil {
		return "", false, err
	}
	if created {
		logger.Info("scratch place created", "component", "scratch", "path", path)
	}
	return path, created, nil
}
