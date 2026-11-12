package worktree

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// gitRun runs git with args in dir and returns its trimmed standard output.
// It logs the process. An error carries the standard error of git.
func gitRun(ctx context.Context, logger *slog.Logger, dir string, args ...string) (string, error) {
	started := time.Now()
	command := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	// The C locale gives untranslated messages, so hive can read them.
	command.Env = append(os.Environ(), "LC_ALL=C")
	var stderr strings.Builder
	command.Stderr = &stderr
	out, err := command.Output()
	attrs := []any{"component", "git", "cwd", dir, "args", args, "duration_ms", time.Since(started).Milliseconds()}
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		logger.Debug("git", append(attrs, "error", message)...)
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	logger.Debug("git", attrs...)
	return strings.TrimSpace(string(out)), nil
}

// repository is the Git repository of a directory.
type repository struct {
	// identity is the common Git directory with each symlink resolved. Each
	// checkout of the repository has the same identity.
	identity string
	// name is the base name of the main checkout.
	name string
	// checkouts are the checkouts that git worktree list gives.
	checkouts []listed
}

// listed is one checkout of git worktree list.
type listed struct {
	path   string
	branch string
}

// openRepository reads the repository of dir.
func openRepository(ctx context.Context, logger *slog.Logger, dir string) (repository, error) {
	common, err := gitRun(ctx, logger, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return repository{}, fmt.Errorf("%s is not in a Git repository: %w", dir, err)
	}
	identity, err := filepath.EvalSymlinks(common)
	if err != nil {
		return repository{}, err
	}
	out, err := gitRun(ctx, logger, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return repository{}, err
	}
	repo := repository{identity: identity, checkouts: parseList(out)}
	if len(repo.checkouts) > 0 {
		repo.name = filepath.Base(repo.checkouts[0].path)
	}
	return repo, nil
}

// parseList reads the porcelain output of git worktree list.
func parseList(out string) []listed {
	var checkouts []listed
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			checkouts = append(checkouts, listed{path: path})
			continue
		}
		if ref, ok := strings.CutPrefix(line, "branch refs/heads/"); ok && len(checkouts) > 0 {
			checkouts[len(checkouts)-1].branch = ref
		}
	}
	return checkouts
}

// checkoutOf returns the path of the checkout that has branch, or an empty
// string.
func (r repository) checkoutOf(branch string) string {
	for _, checkout := range r.checkouts {
		if checkout.branch == branch {
			return checkout.path
		}
	}
	return ""
}
