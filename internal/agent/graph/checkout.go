package graph

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

// checkout reads the Git state of the directory where the agent started.
func checkout(ctx context.Context, dir string, logger *slog.Logger) string {
	git := func(args ...string) (string, string, error) {
		start := time.Now()
		cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(), "LC_ALL=C")
		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if logger != nil {
			attrs := []any{"component", "git", "cwd", dir, "args", args, "duration_ms", time.Since(start).Milliseconds()}
			if err != nil {
				attrs = append(attrs, "error", err.Error())
			}
			logger.Debug("git", attrs...)
		}
		return strings.TrimSpace(string(out)), stderr.String(), err
	}
	if dir == "" {
		return "checkout unavailable"
	}
	inside, stderr, err := git("rev-parse", "--is-inside-work-tree")
	if err != nil {
		if ctx.Err() == nil && strings.Contains(stderr, "not a git repository") {
			return "no checkout"
		}
		return "checkout unavailable"
	}
	if inside != "true" {
		return "no checkout"
	}
	branch, _, err := git("symbolic-ref", "--quiet", "--short", "HEAD")
	if err == nil && branch != "" {
		return branch
	}
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || ctx.Err() != nil {
		return "checkout unavailable"
	}
	commit, _, err := git("rev-parse", "--short", "HEAD")
	if err != nil || commit == "" {
		return "checkout unavailable"
	}
	return "detached · " + commit
}
