// Package gitx runs the bounded Git queries that hive uses.
//
// gitx.go owns the Git commands, the exit-code checks, the plain output, and
// the canonical paths.
package gitx

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/run"
)

// RefExists reports whether a ref exists in the repository. It separates an
// absent ref from a repository or command failure.
func RefExists(ctx context.Context, dir, ref string) (bool, error) {
	result, err := runGit(ctx, dir, "rev-parse", "--verify", "--quiet", ref)
	if err != nil {
		return false, err
	}
	if result.ExitCode == 0 {
		return true, nil
	}
	if result.ExitCode == 1 {
		return false, nil
	}
	return false, gitFailure("rev-parse --verify", result)
}

// BranchExists reports whether a branch exists in the repository.
func BranchExists(ctx context.Context, dir, branch string) (bool, error) {
	result, err := runPlain(ctx, dir, "branch", "--list", branch)
	if err != nil {
		return false, err
	}
	if result.ExitCode != 0 {
		return false, gitFailure("branch --list", result)
	}
	for _, line := range strings.Split(result.Stdout, "\n") {
		if trimBranchLine(line) == branch {
			return true, nil
		}
	}
	return false, nil
}

// IsDirty reports whether the checkout has a change that Git does not commit.
func IsDirty(ctx context.Context, dir string) (bool, error) {
	result, err := runPlain(ctx, dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	if result.ExitCode != 0 {
		return false, gitFailure("status --porcelain", result)
	}
	return strings.TrimSpace(result.Stdout) != "", nil
}

// IsMerged reports whether branch is merged into base.
func IsMerged(ctx context.Context, dir, base, branch string) (bool, error) {
	result, err := runPlain(ctx, dir, "branch", "--merged", base)
	if err != nil {
		return false, err
	}
	if result.ExitCode != 0 {
		return false, gitFailure("branch --merged", result)
	}
	for _, line := range strings.Split(result.Stdout, "\n") {
		if trimBranchLine(line) == branch {
			return true, nil
		}
	}
	return false, nil
}

// CurrentBranch returns the current branch of the checkout.
func CurrentBranch(ctx context.Context, dir string) (string, error) {
	result, err := runPlain(ctx, dir, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", gitFailure("branch --show-current", result)
	}
	return strings.TrimSpace(result.Stdout), nil
}

// IsLinkedWorktree reports whether the checkout is a linked Git worktree.
func IsLinkedWorktree(ctx context.Context, dir string) (bool, error) {
	gitDir, err := revParse(ctx, dir, "--git-dir")
	if err != nil {
		return false, err
	}
	common, err := CommonDir(ctx, dir)
	if err != nil {
		return false, err
	}
	resolved, err := canonicalFrom(dir, gitDir)
	if err != nil {
		return false, fmt.Errorf("git common dir: %s", plainError(err))
	}
	return resolved != common, nil
}

// CommonDir returns the absolute Git common directory of the checkout.
func CommonDir(ctx context.Context, dir string) (string, error) {
	value, err := revParse(ctx, dir, "--git-common-dir")
	if err != nil {
		return "", err
	}
	resolved, err := canonicalFrom(dir, value)
	if err != nil {
		return "", fmt.Errorf("git common dir: %s", plainError(err))
	}
	return resolved, nil
}

// WorktreeRemove runs git worktree remove for one checkout.
func WorktreeRemove(ctx context.Context, dir, path string, force bool) (*run.Result, error) {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	return run.Command(ctx, dir, "git", args...)
}

func runGit(ctx context.Context, dir string, args ...string) (*run.Result, error) {
	return run.Command(ctx, dir, "git", args...)
}

func runPlain(ctx context.Context, dir string, args ...string) (*run.Result, error) {
	plain := append([]string{"-c", "color.ui=false"}, args...)
	return run.Command(ctx, dir, "git", plain...)
}

func revParse(ctx context.Context, dir, option string) (string, error) {
	result, err := runGit(ctx, dir, "rev-parse", option)
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", gitFailure("rev-parse "+option, result)
	}
	return strings.TrimSpace(result.Stdout), nil
}

func canonicalFrom(base, value string) (string, error) {
	path := strings.TrimSpace(value)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(base, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(filepath.Clean(absolute))
}

func trimBranchLine(line string) string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "* ")
	trimmed = strings.TrimPrefix(trimmed, "+ ")
	return strings.TrimSpace(trimmed)
}

func gitFailure(command string, result *run.Result) error {
	text := strings.TrimSpace(result.Stderr)
	if text == "" {
		text = strings.TrimSpace(result.Stdout)
	}
	if text == "" {
		return fmt.Errorf("git %s exited with code %d", command, result.ExitCode)
	}
	return fmt.Errorf("git %s: %s", command, firstLine(text))
}

func firstLine(text string) string {
	if index := strings.IndexByte(text, '\n'); index >= 0 {
		return strings.TrimSpace(text[:index])
	}
	return strings.TrimSpace(text)
}

func plainError(err error) string {
	if err == nil {
		return ""
	}
	return strings.Map(func(value rune) rune {
		if value < 0x20 || value == 0x7f {
			return ' '
		}
		return value
	}, err.Error())
}
