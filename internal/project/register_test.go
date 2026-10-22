package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestRegisterCreatesRecordAndProjectDirectories(t *testing.T) {
	home := t.TempDir()
	repo := makeRepo(t, filepath.Join(t.TempDir(), "Repo-One"), true)
	envelope := Register(home, RegisterOptions{Path: repo, WorkingDir: t.TempDir()})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Register() = %#v; want success", envelope)
	}
	if len(envelope.Effects) != 4 {
		t.Fatalf("effects = %#v; want record and three directories", envelope.Effects)
	}
	if got := envelope.Result.(records.Project).ID; got != "repo-one" {
		t.Errorf("default ID = %q; want lowercase basename repo-one", got)
	}
	if envelope.Effects[0] != (output.Effect{Resource: "record", Path: ".hive/state/projects/repo-one.json", Action: "created"}) {
		t.Errorf("record effect = %#v", envelope.Effects[0])
	}
	for _, path := range []string{
		filepath.Join(home, "projects", "repo-one"),
		filepath.Join(home, "projects", "repo-one", "worktrees"),
		filepath.Join(home, "projects", "repo-one", "overrides"),
	} {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			t.Errorf("project directory %q: stat = %v, %v", path, info, err)
		}
	}
}

func TestRegisterAcceptsGitFile(t *testing.T) {
	home := t.TempDir()
	repo := makeRepo(t, filepath.Join(t.TempDir(), "repo"), false)
	envelope := Register(home, RegisterOptions{Path: repo, WorkingDir: t.TempDir()})
	if envelope.Outcome != "ok" {
		t.Fatalf("Register() error = %v; want a .git file accepted", envelope.Error)
	}
}

func TestRegisterAcceptsThreePartID(t *testing.T) {
	home := t.TempDir()
	repo := makeRepo(t, filepath.Join(t.TempDir(), "repo"), true)
	envelope := Register(home, RegisterOptions{Path: repo, ID: "a-b-c", IDSet: true, WorkingDir: t.TempDir()})
	if envelope.Outcome != "ok" {
		t.Fatalf("Register() error = %v; want valid ID a-b-c accepted", envelope.Error)
	}
}

func TestRegisterRejectsInvalidIDs(t *testing.T) {
	for _, id := range []string{"ab_c", "-ab", "abcdefghijklmnopqrs", "", "bad\nid", "bad\tid"} {
		t.Run(id, func(t *testing.T) {
			home := t.TempDir()
			repo := makeRepo(t, filepath.Join(t.TempDir(), "repo"), true)
			envelope := Register(home, RegisterOptions{Path: repo, ID: id, IDSet: true, WorkingDir: t.TempDir()})
			wantID := id
			if strings.ContainsAny(id, "\n\r\t") {
				wantID = quote(id)
			}
			want := "invalid project id: " + wantID + ": an ID matches ^[a-z0-9]+(-[a-z0-9]+)*$ and has at most 18 characters"
			if envelope.Outcome != "failed" || envelope.Error != want {
				t.Fatalf("Register() = %#v; want %q", envelope, want)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
				t.Fatalf("invalid ID created state: stat error = %v", err)
			}
		})
	}
}

func TestRegisterRejectsInvalidBaseBranches(t *testing.T) {
	for _, base := range []string{"", "-main", "feature branch", "bad\nbranch"} {
		t.Run(base, func(t *testing.T) {
			home := t.TempDir()
			repo := makeRepo(t, filepath.Join(t.TempDir(), "repo"), true)
			envelope := Register(home, RegisterOptions{Path: repo, Base: base, BaseSet: true, WorkingDir: t.TempDir()})
			want := "invalid base branch: " + quote(base)
			if envelope.Error != want {
				t.Fatalf("error = %v; want %q", envelope.Error, want)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
				t.Fatalf("invalid base created state: stat error = %v", err)
			}
		})
	}
}

func TestRegisterRejectsPathControlCharacterBeforePathAccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing") + "\x01"
	home := t.TempDir()
	envelope := Register(home, RegisterOptions{Path: path, WorkingDir: t.TempDir()})
	if envelope.Error != "invalid path: "+quote(path) {
		t.Fatalf("error = %v; want quoted invalid path", envelope.Error)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
		t.Fatalf("control-character path caused file access: stat error = %v", err)
	}
}

func TestRegisterResolvesSymbolicPath(t *testing.T) {
	home := t.TempDir()
	repo := makeRepo(t, filepath.Join(t.TempDir(), "repo"), true)
	alias := filepath.Join(filepath.Dir(repo), "repo-alias")
	if err := os.Symlink(repo, alias); err != nil {
		t.Fatal(err)
	}
	envelope := Register(home, RegisterOptions{Path: alias, WorkingDir: t.TempDir()})
	if envelope.Outcome != "ok" || envelope.Result.(records.Project).Path != repo {
		t.Fatalf("Register() = %#v; want canonical repository path %q", envelope, repo)
	}
}

func TestRegisterResolvesRelativePathWithoutTrimming(t *testing.T) {
	workingDir := t.TempDir()
	repo := makeRepo(t, filepath.Join(workingDir, " repo "), true)
	envelope := Register(t.TempDir(), RegisterOptions{Path: " repo ", ID: "spaced", IDSet: true, WorkingDir: workingDir})
	if envelope.Outcome != "ok" {
		t.Fatalf("Register() error = %v; want path with spaces accepted", envelope.Error)
	}
	if got := envelope.Result.(records.Project).Path; got != repo {
		t.Fatalf("record path = %q; want %q", got, repo)
	}
}

func TestRegisterRejectsNonGitRepository(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(t.TempDir(), "not-git")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := Register(home, RegisterOptions{Path: path, WorkingDir: t.TempDir()})
	if envelope.Error != "not a git repository: "+path {
		t.Fatalf("error = %v; want not a git repository", envelope.Error)
	}
}

func TestRegisterRejectsSymbolicGitMarker(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(t.TempDir(), "repo")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	gitDir := filepath.Join(t.TempDir(), "gitdir")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(gitDir, filepath.Join(path, ".git")); err != nil {
		t.Fatal(err)
	}
	envelope := Register(home, RegisterOptions{Path: path, WorkingDir: t.TempDir()})
	if envelope.Error != "not a git repository: "+path {
		t.Fatalf("error = %v; want symbolic .git marker refused", envelope.Error)
	}
}

func TestRegisterRejectsDuplicateIDAndPath(t *testing.T) {
	home := t.TempDir()
	repo := makeRepo(t, filepath.Join(t.TempDir(), "repo"), true)
	first := Register(home, RegisterOptions{Path: repo, ID: "first", IDSet: true, WorkingDir: t.TempDir()})
	if first.Outcome != "ok" {
		t.Fatal(first.Error)
	}
	duplicateID := Register(home, RegisterOptions{Path: repo, ID: "first", IDSet: true, WorkingDir: t.TempDir()})
	if duplicateID.Error != "project exists: first" {
		t.Fatalf("duplicate ID error = %v", duplicateID.Error)
	}
	duplicatePath := Register(home, RegisterOptions{Path: repo, ID: "second", IDSet: true, WorkingDir: t.TempDir()})
	if duplicatePath.Error != "path registered as first" {
		t.Fatalf("duplicate path error = %v", duplicatePath.Error)
	}
}

func TestRegisterReportsPartialDirectoryFailure(t *testing.T) {
	home := t.TempDir()
	repo := makeRepo(t, filepath.Join(t.TempDir(), "repo-two"), true)
	if err := os.Mkdir(filepath.Join(home, "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "projects", "repo-two"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	envelope := Register(home, RegisterOptions{Path: repo, WorkingDir: t.TempDir()})
	if envelope.Outcome != "partial" || envelope.Phase != "dirs" || len(envelope.Effects) != 1 || !strings.Contains(envelope.Error.(string), "projects/repo-two") {
		t.Fatalf("Register() = %#v; want partial dirs outcome", envelope)
	}
}

func TestRegisterExistingDirectoriesRemainAndReportExists(t *testing.T) {
	home := t.TempDir()
	repo := makeRepo(t, filepath.Join(t.TempDir(), "repo"), true)
	if err := os.MkdirAll(filepath.Join(home, "projects", "repo", "worktrees"), 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := Register(home, RegisterOptions{Path: repo, WorkingDir: t.TempDir()})
	if envelope.Outcome != "ok" || envelope.Effects[2].Action != "exists" {
		t.Fatalf("Register() = %#v; want existing worktrees directory", envelope)
	}
}

func makeRepo(t *testing.T, path string, gitDirectory bool) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	gitPath := filepath.Join(path, ".git")
	if gitDirectory {
		if err := os.Mkdir(gitPath, 0o755); err != nil {
			t.Fatal(err)
		}
	} else if err := os.WriteFile(gitPath, []byte("gitdir: elsewhere"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func quote(value string) string {
	return fmt.Sprintf("%q", value)
}
