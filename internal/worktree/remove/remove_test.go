package remove

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/worktree"
)

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func git(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func removeRun(t *testing.T, home, dir, branch string) output.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return Run(ctx, home, Options{Dir: dir, Branch: branch, Logger: discard()})
}

func errorText(envelope output.Envelope) string {
	text, _ := envelope.Error.(string)
	return text
}

// made returns a fake herdr, a home, a repository, and the checkout of
// branch feat that hive made.
func made(t *testing.T) (*agenttest.Herdr, string, string, worktree.Checkout) {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	checkout, err := worktree.Ensure(context.Background(), home, repo, "feat", "", discard())
	if err != nil {
		t.Fatal(err)
	}
	return fake, home, repo, checkout
}

func checkoutKept(t *testing.T, checkout worktree.Checkout) {
	t.Helper()
	if _, err := os.Stat(checkout.Path); err != nil {
		t.Fatalf("checkout %s stat = %v; want it kept", checkout.Path, err)
	}
}

func TestRemoveDeletesAnIdleCleanCheckoutAndKeepsTheBranch(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, _ := fake.AddWorkspace("", checkout.Path, false)
	envelope := removeRun(t, home, repo, "feat")
	if envelope.Outcome != output.OutcomeOK {
		t.Fatalf("envelope = %+v", envelope)
	}
	if _, err := os.Stat(checkout.Path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checkout stat = %v; want it removed", err)
	}
	if strings.Contains(git(t, "-C", repo, "worktree", "list"), checkout.Path) {
		t.Fatal("git still lists the checkout")
	}
	if git(t, "-C", repo, "branch", "--list", "feat") == "" {
		t.Fatal("branch feat removed")
	}
	if got := fake.Requests("workspace.close"); len(got) != 1 || got[0].Params["workspace_id"] != ws.ID {
		t.Fatalf("workspace.close = %+v; want %s", got, ws.ID)
	}
	if _, err := os.Stat(filepath.Dir(checkout.Path)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("record directory stat = %v; want it removed", err)
	}
	for _, want := range []output.Effect{
		{Resource: "worktree", Path: checkout.Path, Action: "removed"},
		{Resource: "workspace", Path: ws.ID, Action: "closed"},
	} {
		found := false
		for _, effect := range envelope.Effects {
			found = found || effect == want
		}
		if !found {
			t.Errorf("effects = %+v; want %+v", envelope.Effects, want)
		}
	}
}

func TestRemoveWithoutAnOpenWorkspaceClosesNothing(t *testing.T) {
	fake, home, repo, checkout := made(t)
	if envelope := removeRun(t, home, repo, "feat"); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := len(fake.Requests("workspace.close")); got != 0 {
		t.Fatalf("workspace.close calls = %d; want 0", got)
	}
	if _, err := os.Stat(checkout.Path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checkout stat = %v; want it removed", err)
	}
}

func TestRemoveRefusesACheckoutThatHiveDidNotMake(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	path := filepath.Join(home, worktree.Dir, "by-hand", "feat")
	git(t, "-C", repo, "worktree", "add", "-q", "-b", "feat", path)
	for _, branch := range []string{"feat", "main"} {
		envelope := removeRun(t, home, repo, branch)
		if envelope.Outcome != output.OutcomeFailed || !strings.Contains(errorText(envelope), "not made by hive") {
			t.Fatalf("remove %s = %+v; want a failure: not made by hive", branch, envelope)
		}
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("checkout stat = %v; want it kept", err)
	}
}

func TestRemoveRefusesABranchWithoutACheckout(t *testing.T) {
	agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	repo := agenttest.GitRepo(t)
	git(t, "-C", repo, "branch", "kept")
	envelope := removeRun(t, home, repo, "kept")
	if envelope.Outcome != output.OutcomeFailed || !strings.Contains(errorText(envelope), "no checkout") {
		t.Fatalf("envelope = %+v; want a failure: no checkout", envelope)
	}
}

func TestRemoveRefusesADirtyCheckout(t *testing.T) {
	cases := map[string]func(t *testing.T, path string){
		"untracked file": func(t *testing.T, path string) {
			if err := os.WriteFile(filepath.Join(path, "new.txt"), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
		"ignored file": func(t *testing.T, path string) {
			for name, data := range map[string]string{".git/info/exclude": "scratch/\n", "scratch/notes.txt": "x"} {
				target := filepath.Join(path, name)
				if strings.HasPrefix(name, ".git/") {
					common := git(t, "-C", path, "rev-parse", "--path-format=absolute", "--git-common-dir")
					target = filepath.Join(common, strings.TrimPrefix(name, ".git/"))
				}
				if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, []byte(data), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		},
		"unstaged change": func(t *testing.T, path string) {
			if err := os.WriteFile(filepath.Join(path, "a.txt"), []byte("a"), 0o644); err != nil {
				t.Fatal(err)
			}
			git(t, "-C", path, "add", "a.txt")
			git(t, "-C", path, "commit", "-q", "-m", "a")
			if err := os.WriteFile(filepath.Join(path, "a.txt"), []byte("b"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, dirty := range cases {
		t.Run(name, func(t *testing.T) {
			fake, home, repo, checkout := made(t)
			fake.AddWorkspace("", checkout.Path, false)
			dirty(t, checkout.Path)
			envelope := removeRun(t, home, repo, "feat")
			if envelope.Outcome != output.OutcomeFailed || !strings.Contains(errorText(envelope), "dirty") {
				t.Fatalf("envelope = %+v; want a failure: dirty", envelope)
			}
			checkoutKept(t, checkout)
			if got := len(fake.Requests("workspace.close")); got != 0 {
				t.Fatalf("workspace.close calls = %d; want 0", got)
			}
		})
	}
}

func TestRemoveRefusesACheckoutWithALiveAgent(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, _ := fake.AddWorkspace("other", t.TempDir(), false)
	pane := fake.AddPane(ws.ID, checkout.Path)
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: pane.ID, Cwd: filepath.Join(checkout.Path, "sub")}); err != nil {
		t.Fatal(err)
	}
	envelope := removeRun(t, home, repo, "feat")
	if envelope.Outcome != output.OutcomeFailed || !strings.Contains(errorText(envelope), "kid") {
		t.Fatalf("envelope = %+v; want a failure that names agent kid", envelope)
	}
	checkoutKept(t, checkout)

	fake.ClosePane(pane.ID)
	if envelope := removeRun(t, home, repo, "feat"); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("envelope after the pane closed = %+v; want ok", envelope)
	}
}

func TestRemoveRefusesAWorkspaceWithAnAgentPane(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, root := fake.AddWorkspace("", checkout.Path, false)
	fake.StartAgent(root.ID, "", "claude", "idle")
	envelope := removeRun(t, home, repo, "feat")
	if envelope.Outcome != output.OutcomeFailed || !strings.Contains(errorText(envelope), root.ID) {
		t.Fatalf("envelope = %+v; want a failure that names pane %s", envelope, root.ID)
	}
	checkoutKept(t, checkout)
	if got := len(fake.Requests("workspace.close")); got != 0 {
		t.Fatalf("workspace.close calls = %d; want 0 for workspace %s", got, ws.ID)
	}
}

// TestRemoveKeepsTheCheckoutWhenTheWorkspaceCloseFails proves that remove
// closes the workspace before it removes the checkout, so a failed close
// leaves a checkout that a second remove can still find and remove.
func TestRemoveKeepsTheCheckoutWhenTheWorkspaceCloseFails(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, _ := fake.AddWorkspace("", checkout.Path, false)
	fake.Fail("workspace.close", "internal", "boom")
	envelope := removeRun(t, home, repo, "feat")
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "close" || len(envelope.Effects) != 0 {
		t.Fatalf("envelope = %+v; want failed in phase close with no effects", envelope)
	}
	checkoutKept(t, checkout)
	again := removeRun(t, home, repo, "feat")
	if again.Outcome != output.OutcomeOK {
		t.Fatalf("second remove = %+v; want ok", again)
	}
	want := []output.Effect{
		{Resource: "workspace", Path: ws.ID, Action: "closed"},
		{Resource: "worktree", Path: checkout.Path, Action: "removed"},
		{Resource: "record", Path: filepath.Join(worktree.Dir, filepath.Base(filepath.Dir(checkout.Path)), "worktree.json"), Action: "removed"},
	}
	if !slices.Equal(again.Effects, want) {
		t.Fatalf("effects = %+v; want %+v", again.Effects, want)
	}
}

// TestRemoveReportsTheClosedWorkspaceWhenGitFails proves that a git
// worktree remove that fails after the workspace close reports the closed
// workspace, keeps the checkout, and that a second remove finishes.
func TestRemoveReportsTheClosedWorkspaceWhenGitFails(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, _ := fake.AddWorkspace("", checkout.Path, false)
	// git worktree remove refuses a locked checkout.
	git(t, "-C", repo, "worktree", "lock", checkout.Path)
	envelope := removeRun(t, home, repo, "feat")
	want := []output.Effect{{Resource: "workspace", Path: ws.ID, Action: "closed"}}
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "remove" || !slices.Equal(envelope.Effects, want) {
		t.Fatalf("envelope = %+v; want failed in phase remove with effects %+v", envelope, want)
	}
	if message := errorText(envelope); !strings.Contains(message, "the checkout was not removed") || !strings.Contains(message, "closed workspaces "+ws.ID) {
		t.Fatalf("error = %q; want it to name the closed workspace %s", message, ws.ID)
	}
	checkoutKept(t, checkout)
	git(t, "-C", repo, "worktree", "unlock", checkout.Path)
	if again := removeRun(t, home, repo, "feat"); again.Outcome != output.OutcomeOK {
		t.Fatalf("second remove = %+v; want ok", again)
	}
}

// TestRemoveFinishesARemoveThatLeftTheRecord proves that a remove whose
// record removal fails after git removed the checkout is partial, and that
// a second remove of the branch, which now has no checkout, removes the
// record.
func TestRemoveFinishesARemoveThatLeftTheRecord(t *testing.T) {
	_, home, repo, checkout := made(t)
	record := filepath.Join(worktree.Dir, filepath.Base(filepath.Dir(checkout.Path)), "worktree.json")
	// The record removal needs the store lock: hold it for the first remove.
	lock, err := os.OpenFile(filepath.Join(home, "lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	envelope := removeRun(t, home, repo, "feat")
	_ = lock.Close()
	want := []output.Effect{{Resource: "worktree", Path: checkout.Path, Action: "removed"}}
	if envelope.Outcome != output.OutcomePartial || envelope.Phase != "record" || !slices.Equal(envelope.Effects, want) {
		t.Fatalf("envelope = %+v; want partial in phase record with effects %+v", envelope, want)
	}
	if _, err := os.Stat(filepath.Join(home, record)); err != nil {
		t.Fatalf("record stat = %v; want it left by the failed remove", err)
	}
	again := removeRun(t, home, repo, "feat")
	want = []output.Effect{{Resource: "record", Path: record, Action: "removed"}}
	if again.Outcome != output.OutcomeOK || !slices.Equal(again.Effects, want) {
		t.Fatalf("second remove = %+v; want ok with effects %+v", again, want)
	}
	if _, err := os.Stat(filepath.Dir(filepath.Join(home, record))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("record directory stat = %v; want it removed", err)
	}
	if third := removeRun(t, home, repo, "feat"); third.Outcome != output.OutcomeFailed || !strings.Contains(errorText(third), "no checkout") {
		t.Fatalf("third remove = %+v; want a failure: no checkout", third)
	}
}

// TestRemoveSeesAnAgentThroughASymlink proves that an agent whose directory
// reaches the checkout through a symlink keeps the checkout.
func TestRemoveSeesAnAgentThroughASymlink(t *testing.T) {
	fake, home, repo, checkout := made(t)
	if err := os.Mkdir(filepath.Join(checkout.Path, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(checkout.Path, alias); err != nil {
		t.Fatal(err)
	}
	ws, _ := fake.AddWorkspace("other", t.TempDir(), false)
	pane := fake.AddPane(ws.ID, t.TempDir())
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: pane.ID, Cwd: filepath.Join(alias, "sub")}); err != nil {
		t.Fatal(err)
	}
	envelope := removeRun(t, home, repo, "feat")
	if envelope.Outcome != output.OutcomeFailed || !strings.Contains(errorText(envelope), "kid") {
		t.Fatalf("envelope = %+v; want a failure that names agent kid", envelope)
	}
	checkoutKept(t, checkout)
}

// TestRemoveFindsAWorkspaceOnASubdirectory proves that remove treats a
// workspace whose panes are all in the checkout, for example one that spawn
// --directory made on a subdirectory, as the workspace of the checkout:
// remove closes it when it runs only shells, and refuses when it runs an
// agent without a record. The panes reach the checkout through a symlink.
func TestRemoveFindsAWorkspaceOnASubdirectory(t *testing.T) {
	for _, withAgent := range []bool{false, true} {
		t.Run(fmt.Sprintf("agent=%v", withAgent), func(t *testing.T) {
			fake, home, repo, checkout := made(t)
			sub := filepath.Join(checkout.Path, "sub")
			if err := os.Mkdir(sub, 0o755); err != nil {
				t.Fatal(err)
			}
			alias := filepath.Join(t.TempDir(), "alias")
			if err := os.Symlink(checkout.Path, alias); err != nil {
				t.Fatal(err)
			}
			ws, root := fake.AddWorkspace("sub", sub, false)
			fake.AddPane(ws.ID, filepath.Join(alias, "sub"))
			if withAgent {
				fake.StartAgent(root.ID, "", "claude", "idle")
			}
			envelope := removeRun(t, home, repo, "feat")
			if withAgent {
				if envelope.Outcome != output.OutcomeFailed || !strings.Contains(errorText(envelope), root.ID) {
					t.Fatalf("envelope = %+v; want a failure that names pane %s", envelope, root.ID)
				}
				checkoutKept(t, checkout)
				if got := len(fake.Requests("workspace.close")); got != 0 {
					t.Fatalf("workspace.close calls = %d; want 0", got)
				}
				return
			}
			if envelope.Outcome != output.OutcomeOK {
				t.Fatalf("envelope = %+v; want ok", envelope)
			}
			if got := fake.Requests("workspace.close"); len(got) != 1 || got[0].Params["workspace_id"] != ws.ID {
				t.Fatalf("workspace.close = %+v; want %s", got, ws.ID)
			}
		})
	}
}

// TestRemoveRefusesAShellInTheCheckoutInAnotherWorkspace proves that remove
// does not close a workspace that also has panes outside the checkout, and
// keeps the checkout while a pane of it is in the checkout.
func TestRemoveRefusesAShellInTheCheckoutInAnotherWorkspace(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, _ := fake.AddWorkspace("mine", t.TempDir(), false)
	inside := fake.AddPane(ws.ID, checkout.Path)
	envelope := removeRun(t, home, repo, "feat")
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "busy" || !strings.Contains(errorText(envelope), inside.ID) {
		t.Fatalf("envelope = %+v; want a busy failure that names pane %s", envelope, inside.ID)
	}
	checkoutKept(t, checkout)
	if got := len(fake.Requests("workspace.close")); got != 0 {
		t.Fatalf("workspace.close calls = %d; want 0", got)
	}
}

// TestRemoveChecksTheCheckoutAgainAfterTheWorkspaceClose proves that an
// ignored file that appears after the first check, here while the workspace
// closes, keeps the checkout.
func TestRemoveChecksTheCheckoutAgainAfterTheWorkspaceClose(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, _ := fake.AddWorkspace("", checkout.Path, false)
	common := git(t, "-C", checkout.Path, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err := os.WriteFile(filepath.Join(common, "info", "exclude"), []byte("scratch.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ignored := filepath.Join(checkout.Path, "scratch.txt")
	fake.Before("workspace.close", func() {
		if err := os.WriteFile(ignored, []byte("keep"), 0o644); err != nil {
			t.Error(err)
		}
	})
	envelope := removeRun(t, home, repo, "feat")
	want := []output.Effect{{Resource: "workspace", Path: ws.ID, Action: "closed"}}
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "dirty" || !slices.Equal(envelope.Effects, want) {
		t.Fatalf("envelope = %+v; want failed in phase dirty with effects %+v", envelope, want)
	}
	if message := errorText(envelope); strings.Contains(message, "nothing was removed") || !strings.Contains(message, "closed workspaces "+ws.ID) || !strings.Contains(message, "the checkout was not removed") {
		t.Fatalf("error = %q; want it to say that the checkout stays and workspace %s closed", message, ws.ID)
	}
	if data, err := os.ReadFile(ignored); err != nil || string(data) != "keep" {
		t.Fatalf("ignored file = %q, %v; want it kept", data, err)
	}
}

// TestRemoveRefusesToCloseTheWorkspaceOfTheCaller proves that remove run from
// a shell in a workspace of the checkout closes nothing, so the close does not
// end the remove itself, and keeps the checkout.
func TestRemoveRefusesToCloseTheWorkspaceOfTheCaller(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, root := fake.AddWorkspace("", checkout.Path, false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	envelope := Run(ctx, home, Options{Dir: checkout.Path, Branch: "feat", CallerPane: root.ID, Logger: discard()})
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "caller" || !strings.Contains(errorText(envelope), ws.ID) {
		t.Fatalf("envelope = %+v; want failed in phase caller that names workspace %s", envelope, ws.ID)
	}
	checkoutKept(t, checkout)
	if got := len(fake.Requests("workspace.close")); got != 0 {
		t.Fatalf("workspace.close calls = %d; want 0", got)
	}
	other, _ := fake.AddWorkspace("other", t.TempDir(), false)
	if envelope := Run(ctx, home, Options{Dir: repo, Branch: "feat", CallerPane: fake.AddPane(other.ID, repo).ID, Logger: discard()}); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("remove from another workspace = %+v; want ok", envelope)
	}
}

// TestRemoveRefusesAWorkspaceOnTheCheckoutWithAPaneElsewhere proves that
// remove does not close the workspace open on the checkout when one of its
// panes is outside the checkout, so no unrelated process stops.
func TestRemoveRefusesAWorkspaceOnTheCheckoutWithAPaneElsewhere(t *testing.T) {
	fake, home, repo, checkout := made(t)
	ws, _ := fake.AddWorkspace("", checkout.Path, false)
	outside := fake.AddPane(ws.ID, t.TempDir())
	envelope := removeRun(t, home, repo, "feat")
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "busy" || !strings.Contains(errorText(envelope), outside.ID) {
		t.Fatalf("envelope = %+v; want a busy failure that names pane %s", envelope, outside.ID)
	}
	checkoutKept(t, checkout)
	if got := len(fake.Requests("workspace.close")); got != 0 {
		t.Fatalf("workspace.close calls = %d; want 0", got)
	}
	fake.ClosePane(outside.ID)
	if envelope := removeRun(t, home, repo, "feat"); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("remove after the pane closed = %+v; want ok", envelope)
	}
}

// TestRemoveSeesAForegroundProcessInTheCheckout proves that a pane whose
// shell is outside the checkout, but whose foreground process runs in it,
// keeps the checkout: as an agent, or as a plain process that the close of
// its mixed workspace would stop.
func TestRemoveSeesAForegroundProcessInTheCheckout(t *testing.T) {
	for _, withAgent := range []bool{false, true} {
		t.Run(fmt.Sprintf("agent=%v", withAgent), func(t *testing.T) {
			fake, home, repo, checkout := made(t)
			_, pane := fake.AddWorkspace("main", repo, false)
			fake.SetForegroundCwd(pane.ID, checkout.Path)
			if withAgent {
				fake.StartAgent(pane.ID, "", "claude", "idle")
			}
			envelope := removeRun(t, home, repo, "feat")
			if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "busy" || !strings.Contains(errorText(envelope), pane.ID) {
				t.Fatalf("envelope = %+v; want a busy failure that names pane %s", envelope, pane.ID)
			}
			checkoutKept(t, checkout)
			if got := len(fake.Requests("workspace.close")); got != 0 {
				t.Fatalf("workspace.close calls = %d; want 0", got)
			}
		})
	}
}

// TestRemoveRefusesAForegroundProcessOutsideTheCheckout proves that remove
// does not close the workspace on the checkout while the foreground process
// of its shell runs outside the checkout.
func TestRemoveRefusesAForegroundProcessOutsideTheCheckout(t *testing.T) {
	fake, home, repo, checkout := made(t)
	_, root := fake.AddWorkspace("", checkout.Path, false)
	fake.SetForegroundCwd(root.ID, t.TempDir())
	envelope := removeRun(t, home, repo, "feat")
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "busy" || !strings.Contains(errorText(envelope), root.ID) {
		t.Fatalf("envelope = %+v; want a busy failure that names pane %s", envelope, root.ID)
	}
	checkoutKept(t, checkout)
	if got := len(fake.Requests("workspace.close")); got != 0 {
		t.Fatalf("workspace.close calls = %d; want 0", got)
	}
}

// TestRemoveNamesTheClosedWorkspaceWhenALaterCloseFails proves that a close
// that fails after an earlier close names the closed workspace in its error
// and its effects.
func TestRemoveNamesTheClosedWorkspaceWhenALaterCloseFails(t *testing.T) {
	fake, home, repo, checkout := made(t)
	first, _ := fake.AddWorkspace("", checkout.Path, false)
	sub := filepath.Join(checkout.Path, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	fake.AddWorkspace("sub", sub, false)
	calls := 0
	fake.Before("workspace.close", func() {
		calls++
		if calls == 2 {
			fake.Fail("workspace.close", "internal", "boom")
		}
	})
	envelope := removeRun(t, home, repo, "feat")
	want := []output.Effect{{Resource: "workspace", Path: first.ID, Action: "closed"}}
	if envelope.Outcome != output.OutcomeFailed || envelope.Phase != "close" || !slices.Equal(envelope.Effects, want) {
		t.Fatalf("envelope = %+v; want failed in phase close with effects %+v", envelope, want)
	}
	if message := errorText(envelope); !strings.Contains(message, "the checkout was not removed") || !strings.Contains(message, "closed workspaces "+first.ID) {
		t.Fatalf("error = %q; want it to name the closed workspace %s", message, first.ID)
	}
	checkoutKept(t, checkout)
}
