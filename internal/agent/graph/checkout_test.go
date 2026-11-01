package graph

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent/list"
)

func TestCheckoutReportsCurrentGitState(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-b", "first")
	git("-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "initial")
	ctx := context.Background()
	if got := checkout(ctx, dir, nil); got != "first" {
		t.Fatalf("branch = %q; want first", got)
	}
	e := entry("worker", "", "idle", "pi", "model", 0)
	e.Cwd = dir
	m := New(Options{Fetch: fetchOf([]list.Entry{e}), Now: func() time.Time { return clock }, Checkout: func(ctx context.Context, path string) string {
		return checkout(ctx, path, nil)
	}})
	m = update(t, m, m.fetch()())
	if !strings.Contains(screen(m), "first") {
		t.Fatalf("first refresh does not show branch:\n%s", screen(m))
	}
	git("checkout", "-b", "second")
	if got := checkout(ctx, dir, nil); got != "second" {
		t.Fatalf("changed branch = %q; want second", got)
	}
	m = update(t, m, m.fetch()())
	if !strings.Contains(screen(m), "second") || strings.Contains(screen(m), "first") {
		t.Fatalf("second refresh does not show changed branch:\n%s", screen(m))
	}
	short := git("rev-parse", "--short", "HEAD")
	git("checkout", "--detach")
	if got := checkout(ctx, dir, nil); got != "detached · "+short {
		t.Fatalf("detached = %q; want detached · %s", got, short)
	}
	if got := checkout(ctx, t.TempDir(), nil); got != "no checkout" {
		t.Fatalf("outside Git = %q; want no checkout", got)
	}
	if got := checkout(ctx, filepath.Join(dir, "missing"), nil); got != "checkout unavailable" {
		t.Fatalf("missing directory = %q; want checkout unavailable", got)
	}
}
