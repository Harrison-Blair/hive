package create

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/worktree"
)

func createRun(t *testing.T, home, dir, branch, base string) output.Envelope {
	t.Helper()
	return Run(context.Background(), home, Options{Dir: dir, Branch: branch, Base: base, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
}

func TestCreateMakesTheCheckoutAndThenReturnsIt(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	first := createRun(t, home, repo, "feat", "")
	checkout, ok := first.Result.(worktree.Checkout)
	if first.Outcome != output.OutcomeOK || !ok || !checkout.Created {
		t.Fatalf("first create = %+v; want ok and a created checkout", first)
	}
	if text := Text(first); text != "created branch feat\ncreated worktree "+checkout.Path+" of branch feat\n" {
		t.Fatalf("text = %q", text)
	}
	again := createRun(t, home, repo, "feat", "")
	if again.Outcome != output.OutcomeOK || again.Result.(worktree.Checkout).Path != checkout.Path || len(again.Effects) != 0 {
		t.Fatalf("second create = %+v; want ok, the same path, no effects", again)
	}
	if text := Text(again); text != "branch feat has worktree "+checkout.Path+"\n" {
		t.Fatalf("text = %q", text)
	}
	main := createRun(t, home, repo, "main", "")
	if main.Outcome != output.OutcomeOK || main.Result.(worktree.Checkout).Path != repo {
		t.Fatalf("create main = %+v; want the main checkout %s", main, repo)
	}
}

func TestCreateFailsWithTheCause(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	for _, test := range []struct{ branch, base, want string }{
		{"", "", "BRANCH is required"},
		{"main", "main", "--base"},
	} {
		envelope := createRun(t, home, repo, test.branch, test.base)
		if text, _ := envelope.Error.(string); envelope.Outcome != output.OutcomeFailed || !strings.Contains(text, test.want) {
			t.Fatalf("create %q base %q = %+v; want a failure with %q", test.branch, test.base, envelope, test.want)
		}
	}
}

// TestConcurrentCreatesOfOneBranchShareOneOwnedCheckout proves that creates
// of one branch at the same time all get the one checkout, and that a create
// that loses the race does not remove the ownership record of the winner.
func TestConcurrentCreatesOfOneBranchShareOneOwnedCheckout(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	const count = 8
	envelopes := make([]output.Envelope, count)
	var group sync.WaitGroup
	for i := range count {
		group.Add(1)
		go func() {
			defer group.Done()
			envelopes[i] = createRun(t, home, repo, "feat", "")
		}()
	}
	group.Wait()
	created := 0
	for _, envelope := range envelopes {
		if envelope.Outcome != output.OutcomeOK {
			t.Fatalf("create = %+v; want ok", envelope)
		}
		if envelope.Result.(worktree.Checkout).Created {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("creates that made the checkout = %d; want 1", created)
	}
	checkout, found, err := worktree.Lookup(context.Background(), home, repo, "feat", nil)
	if err != nil || !found || !checkout.Owned {
		t.Fatalf("Lookup = %+v, %v, %v; want the owned checkout", checkout, found, err)
	}
}

// TestCreateReportsTheCheckoutThatAFailedHookLeft proves that a create whose
// git worktree add fails after it made the checkout is partial, gives the
// checkout, and reports the branch and the checkout.
func TestCreateReportsTheCheckoutThatAFailedHookLeft(t *testing.T) {
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	if err := os.WriteFile(filepath.Join(repo, ".git", "hooks", "post-checkout"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := createRun(t, home, repo, "feat", "")
	checkout, ok := envelope.Result.(worktree.Checkout)
	if envelope.Outcome != output.OutcomePartial || envelope.Phase != "worktree" || envelope.Error == nil || !ok || !checkout.Owned {
		t.Fatalf("create = %+v; want partial in phase worktree with an error and the owned checkout", envelope)
	}
	want := []output.Effect{
		{Resource: "branch", Path: "feat", Action: "created"},
		{Resource: "worktree", Path: checkout.Path, Action: "created"},
	}
	if !slices.Equal(envelope.Effects, want) {
		t.Fatalf("effects = %+v; want %+v", envelope.Effects, want)
	}
}

// TestCreateTellsOfABranchThatAFailedAddLeft proves that a create whose git
// worktree add makes the branch and then fails, because the path of the
// checkout is taken, shows the new branch in its text and its error.
func TestCreateTellsOfABranchThatAFailedAddLeft(t *testing.T) {
	repo := agenttest.GitRepo(t)
	// The key of the checkout path depends on the repository and the
	// branch, not on the home: learn it from a create in another home.
	other := agenttest.Home(t)
	probe := createRun(t, other, repo, "feat", "")
	path := probe.Result.(worktree.Checkout).Path
	for _, args := range [][]string{{"worktree", "remove", path}, {"branch", "-D", "feat"}} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	home := agenttest.Home(t)
	taken := filepath.Join(home, worktree.Dir, filepath.Base(filepath.Dir(path)), filepath.Base(path))
	if err := os.MkdirAll(taken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taken, "x"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	envelope := createRun(t, home, repo, "feat", "")
	want := []output.Effect{{Resource: "branch", Path: "feat", Action: "created"}}
	if envelope.Outcome != output.OutcomeFailed || !slices.Equal(envelope.Effects, want) {
		t.Fatalf("create = %+v; want failed with effects %+v", envelope, want)
	}
	if text := Text(envelope); text != "created branch feat\n" {
		t.Fatalf("text = %q; want the new branch", text)
	}
	if message, _ := envelope.Error.(string); !strings.Contains(message, "branch feat was made") || !strings.Contains(message, "without --base") {
		t.Fatalf("error = %q; want it to name the new branch and the retry without --base", message)
	}
}
