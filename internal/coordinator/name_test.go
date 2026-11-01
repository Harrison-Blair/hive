package coordinator

import (
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// resolve returns the name that ResolveName gives for dir and pane, and fails
// the test on an error.
func resolve(t *testing.T, home, dir, pane string) string {
	t.Helper()
	name, err := ResolveName(home, dir, pane, discard())
	if err != nil {
		t.Fatalf("ResolveName(%s) error = %v; want none", dir, err)
	}
	return name
}

func TestResolveNameUsesTheNormalizedDirectoryName(t *testing.T) {
	agenttest.NewHerdr(t)
	home := t.TempDir()
	cases := []struct{ dir, want string }{
		{"/src/hive", "hive-coordinator"},
		{"/src/hive/", "hive-coordinator"},
		{"/src/My Repo", "my-repo-coordinator"},
		{"/src/ror.desktop", "ror-desktop-coordinator"},
		{"/src/__x__", "__x__-coordinator"},
		{"/src/--App--", "app-coordinator"},
		{"/src/日本", "coordinator"},
		{"/", "coordinator"},
	}
	for _, test := range cases {
		if got := resolve(t, home, test.dir, "w1:p1"); got != test.want {
			t.Errorf("ResolveName(%s) = %q; want %q", test.dir, got, test.want)
		}
	}
}

func TestResolveNameFitsTheAgentNameLimit(t *testing.T) {
	agenttest.NewHerdr(t)
	home := t.TempDir()
	dir := filepath.Join("/src", strings.Repeat("a", 100))
	got := resolve(t, home, dir, "w1:p1")
	if len(got) != 64 || !strings.HasSuffix(got, "-coordinator") || agent.ValidateName(got) != nil {
		t.Fatalf("ResolveName = %q (%d bytes); want a valid 64 byte name with -coordinator", got, len(got))
	}
}

// TestResolveNameAddsANumberWhenALiveAgentHasTheName proves that a name of a
// record in another live pane gets the next free number before -coordinator,
// while a record of a gone pane or of the own pane leaves the name free.
func TestResolveNameAddsANumberWhenALiveAgentHasTheName(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := t.TempDir()
	ws, mine := fake.AddWorkspace("w", home, false)
	other := fake.AddPane(ws.ID, home)
	third := fake.AddPane(ws.ID, home)
	save := func(name, pane string) {
		t.Helper()
		if err := agent.Save(home, agent.Record{Name: name, Pane: pane}); err != nil {
			t.Fatal(err)
		}
	}
	save("app-coordinator", other.ID)
	if got := resolve(t, home, "/src/app", mine.ID); got != "app-2-coordinator" {
		t.Fatalf("ResolveName = %q; want app-2-coordinator", got)
	}
	save("app-2-coordinator", third.ID)
	if got := resolve(t, home, "/src/app", mine.ID); got != "app-3-coordinator" {
		t.Fatalf("ResolveName = %q; want app-3-coordinator", got)
	}
	save("app-3-coordinator", "w9:p9")
	if got := resolve(t, home, "/src/app", mine.ID); got != "app-3-coordinator" {
		t.Fatalf("ResolveName = %q; want app-3-coordinator, the record of a gone pane", got)
	}
	save("app-coordinator", mine.ID)
	if got := resolve(t, home, "/src/app", mine.ID); got != "app-coordinator" {
		t.Fatalf("ResolveName = %q; want app-coordinator, the record of the own pane", got)
	}
	save("coordinator", other.ID)
	if got := resolve(t, home, "/", mine.ID); got != "coordinator-2" {
		t.Fatalf("ResolveName(/) = %q; want coordinator-2", got)
	}
}

func TestResolveNameKeepsTheNumberedNameInTheLimit(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := t.TempDir()
	ws, mine := fake.AddWorkspace("w", home, false)
	other := fake.AddPane(ws.ID, home)
	dir := filepath.Join("/src", strings.Repeat("a", 100))
	first := resolve(t, home, dir, mine.ID)
	if err := agent.Save(home, agent.Record{Name: first, Pane: other.ID}); err != nil {
		t.Fatal(err)
	}
	got := resolve(t, home, dir, mine.ID)
	if len(got) > 64 || !strings.HasSuffix(got, "-2-coordinator") || agent.ValidateName(got) != nil {
		t.Fatalf("ResolveName = %q (%d bytes); want a valid name of at most 64 bytes with -2-coordinator", got, len(got))
	}
}

func TestResolveNameFailsWhenHerdrFails(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := t.TempDir()
	if err := agent.Save(home, agent.Record{Name: "app-coordinator", Pane: "w1:p1"}); err != nil {
		t.Fatal(err)
	}
	fake.Fail("pane.get", "internal", "boom")
	if _, err := ResolveName(home, "/src/app", "w2:p2", discard()); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("ResolveName error = %v; want the herdr error", err)
	}
}
