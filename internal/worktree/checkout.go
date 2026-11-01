package worktree

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// Dir is the directory below the home that holds the checkouts that hive
// makes, each with its ownership record.
const Dir = "worktrees"

// recordID is the name of the ownership record in the directory of a key.
const recordID = "worktree"

// Checkout is the checkout of one branch.
type Checkout struct {
	// Repo is the common Git directory of the repository, with each symlink
	// resolved.
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Path   string `json:"path"`
	// Created is true when this call made the checkout.
	Created bool `json:"created"`
	// BranchCreated is true when this call made the branch.
	BranchCreated bool `json:"branch_created"`
	// Owned is true when hive made the checkout: its ownership record names
	// this repository, branch, and path.
	Owned bool `json:"owned"`
}

// Record is the ownership record of a checkout that hive made, at
// <home>/worktrees/<key>/worktree.json.
type Record struct {
	Repo      string    `json:"repo"`
	Branch    string    `json:"branch"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// Ensure returns the checkout of branch in the Git repository of dir. A
// branch that a checkout has gives that checkout. Otherwise Ensure makes a
// checkout below <home>/worktrees and writes its ownership record. A new
// branch starts at base, or at the current commit of dir when base is empty.
// A base for a branch that exists is an error. On an error, the checkout
// gives what Ensure made before the failure, as Effects shows. The caller
// holds agent.Hold, so no other Ensure, spawn, or remove acts on the
// checkouts of home.
func Ensure(ctx context.Context, home, dir, branch, base string, logger *slog.Logger) (Checkout, error) {
	logger = agent.Logger(logger)
	if _, err := gitRun(ctx, logger, dir, "check-ref-format", "--branch", branch); err != nil {
		return Checkout{}, fmt.Errorf("invalid branch name %q: %w", branch, err)
	}
	repo, err := openRepository(ctx, logger, dir)
	if err != nil {
		return Checkout{}, err
	}
	exists := false
	if _, err := gitRun(ctx, logger, dir, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
		exists = true
	}
	if exists && base != "" {
		return Checkout{}, fmt.Errorf("branch %s exists: --base applies only to a new branch", branch)
	}
	if path := repo.checkoutOf(branch); path != "" {
		checkout := Checkout{Repo: repo.identity, Branch: branch, Path: path}
		checkout.Owned, err = owned(home, repo, branch, path)
		return checkout, err
	}

	key := keyOf(repo, branch)
	path := filepath.Join(home, Dir, key, leafOf(branch))
	args := []string{"worktree", "add", path, branch}
	if !exists {
		ref := base
		if ref == "" {
			ref = "HEAD"
		}
		// The commit hash keeps a base that starts with "-" out of the
		// options of git worktree add.
		start, err := gitRun(ctx, logger, dir, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
		if err != nil {
			return Checkout{}, fmt.Errorf("start commit %s of branch %s: %w", ref, branch, err)
		}
		args = []string{"worktree", "add", "-b", branch, path, start}
	}
	records, err := store.Open(home)
	if err != nil {
		return Checkout{}, err
	}
	record := Record{Repo: repo.identity, Branch: branch, Path: path, CreatedAt: time.Now().UTC()}
	// The record comes first, so a checkout that hive made always has one. A
	// record without its checkout owns nothing, because ownership also needs
	// the checkout of the branch at the path of the record.
	if err := records.WriteIn([]string{Dir, key}, recordID, record); err != nil {
		return Checkout{}, err
	}
	if _, err := gitRun(ctx, logger, dir, args...); err != nil {
		return failedAdd(ctx, logger, records, dir, key, Checkout{Repo: repo.identity, Branch: branch}, path, !exists, err)
	}
	logger.Debug("worktree created", "path", path, "branch", branch)
	return Checkout{Repo: repo.identity, Branch: branch, Path: path, Created: true, BranchCreated: !exists, Owned: true}, nil
}

// failedAdd reads what a failed git worktree add left and returns it with
// err. git can fail after it made the branch, for example when the path
// exists, and after it made the checkout, for example when a post-checkout
// hook fails. A checkout at path keeps its ownership record. Otherwise the
// record is removed. When Git cannot be read, the record stays, because a
// record without its checkout owns nothing.
func failedAdd(ctx context.Context, logger *slog.Logger, records *store.Store, dir, key string, checkout Checkout, path string, newBranch bool, err error) (Checkout, error) {
	repo, readErr := openRepository(ctx, logger, dir)
	if readErr != nil {
		logger.Warn("worktree state unreadable after a failed add", "path", path, "error", readErr.Error())
		return checkout, err
	}
	if newBranch {
		if _, refErr := gitRun(ctx, logger, dir, "rev-parse", "--verify", "--quiet", "refs/heads/"+checkout.Branch); refErr == nil {
			checkout.BranchCreated = true
		}
	}
	if SamePath(repo.checkoutOf(checkout.Branch), path) {
		checkout.Path, checkout.Created, checkout.Owned = path, true, true
		err = fmt.Errorf("checkout %s of branch %s was made, but git failed after that, for example in a post-checkout hook: %w", path, checkout.Branch, err)
		logger.Warn("worktree created with an error", "path", path, "branch", checkout.Branch, "error", err.Error())
		return checkout, err
	}
	if removeErr := records.RemoveIn([]string{Dir, key}, recordID); removeErr != nil {
		logger.Warn("worktree record remove failed", "path", path, "error", removeErr.Error())
	}
	if checkout.BranchCreated {
		err = fmt.Errorf("branch %s was made and stays, but git made no checkout at %s: fix the cause, then run the command again without --base: %w", checkout.Branch, path, err)
	}
	return checkout, err
}

// Effects returns the effects of the Ensure that gave checkout: the branch
// and the checkout that it made.
func Effects(checkout Checkout) []output.Effect {
	effects := []output.Effect{}
	if checkout.BranchCreated {
		effects = append(effects, output.Effect{Resource: "branch", Path: checkout.Branch, Action: "created"})
	}
	if checkout.Created {
		effects = append(effects, output.Effect{Resource: "worktree", Path: checkout.Path, Action: "created"})
	}
	return effects
}

// Lookup returns the checkout of branch in the Git repository of dir. found
// is false when no checkout has branch.
func Lookup(ctx context.Context, home, dir, branch string, logger *slog.Logger) (checkout Checkout, found bool, err error) {
	logger = agent.Logger(logger)
	repo, err := openRepository(ctx, logger, dir)
	if err != nil {
		return Checkout{}, false, err
	}
	path := repo.checkoutOf(branch)
	if path == "" {
		return Checkout{Repo: repo.identity, Branch: branch}, false, nil
	}
	checkout = Checkout{Repo: repo.identity, Branch: branch, Path: path}
	checkout.Owned, err = owned(home, repo, branch, path)
	return checkout, true, err
}

// Stale returns the checkout that the ownership record of branch names when
// no checkout has branch and nothing is at the path of the record: a remove
// that removed the checkout and then failed left the record. found is false
// otherwise.
func Stale(ctx context.Context, home, dir, branch string, logger *slog.Logger) (checkout Checkout, found bool, err error) {
	logger = agent.Logger(logger)
	repo, err := openRepository(ctx, logger, dir)
	if err != nil {
		return Checkout{}, false, err
	}
	if repo.checkoutOf(branch) != "" {
		return Checkout{}, false, nil
	}
	key := keyOf(repo, branch)
	record, err := readRecord(home, key)
	if errors.Is(err, os.ErrNotExist) {
		return Checkout{}, false, nil
	}
	if err != nil {
		return Checkout{}, false, err
	}
	// Forget finds the record from the directory of the path.
	if record.Repo != repo.identity || record.Branch != branch || filepath.Base(filepath.Dir(record.Path)) != key {
		return Checkout{}, false, nil
	}
	switch _, err := os.Lstat(record.Path); {
	case err == nil:
		// Something is at the path that Git does not list: leave it.
		return Checkout{}, false, nil
	case !errors.Is(err, os.ErrNotExist):
		return Checkout{}, false, err
	}
	return Checkout{Repo: repo.identity, Branch: branch, Path: record.Path, Owned: true}, true, nil
}

// Forget removes the ownership record of checkout and its directory when
// that is empty.
func Forget(home string, checkout Checkout) error {
	records, err := store.Open(home)
	if err != nil {
		return err
	}
	return records.RemoveIn([]string{Dir, filepath.Base(filepath.Dir(checkout.Path))}, recordID)
}

// owned reports whether the ownership record of branch in repo names path.
func owned(home string, repo repository, branch, path string) (bool, error) {
	record, err := readRecord(home, keyOf(repo, branch))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return record.Repo == repo.identity && record.Branch == branch && SamePath(record.Path, path), nil
}

func readRecord(home, key string) (Record, error) {
	records, err := store.Open(home)
	if err != nil {
		return Record{}, err
	}
	var record Record
	err = records.ReadIn([]string{Dir, key}, recordID, &record)
	return record, err
}

// keyOf returns the stable directory name of the checkout of branch in repo:
// the slug of the repository name and a hash of the repository identity and
// the full branch name.
func keyOf(repo repository, branch string) string {
	sum := sha256.Sum256([]byte(repo.identity + "\x00" + branch))
	// A store name starts with a letter or a digit.
	name := strings.TrimLeft(agent.Slug(repo.name, 40), "_-")
	if name == "" {
		name = "repo"
	}
	return name + "-" + hex.EncodeToString(sum[:8])
}

// leafOf returns the readable base name of the checkout of branch. herdr
// gives a new workspace the base name of its directory as its label.
func leafOf(branch string) string {
	if leaf := agent.Slug(branch, 64); leaf != "" {
		return leaf
	}
	return "checkout"
}

// SamePath reports whether a and b name the same directory, with each symlink
// resolved when the path exists.
func SamePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if resolved, err := filepath.EvalSymlinks(a); err == nil {
		a = resolved
	}
	if resolved, err := filepath.EvalSymlinks(b); err == nil {
		b = resolved
	}
	return filepath.Clean(a) == filepath.Clean(b)
}
