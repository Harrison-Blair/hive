package doctor

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", dir}, args...)...)
	command.Env = append(command.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

// homeRepo makes a Git repository with one commit and the given tags at HEAD.
func homeRepo(t *testing.T, tags ...string) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")
	gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", "first")
	for _, tag := range tags {
		gitRun(t, dir, "tag", tag)
	}
	return dir
}

func TestHomeVersionDevBuildPasses(t *testing.T) {
	check := checkHomeVersion(context.Background(), t.TempDir(), "dev", "git")
	if check.Name != "home version" || check.Status != StatusOK || check.Required {
		t.Fatalf("check = %+v, want an optional ok home version", check)
	}
	if !strings.Contains(check.Detail, "development build") {
		t.Fatalf("detail = %q, want the development build text", check.Detail)
	}
}

func TestHomeVersionSameTagPasses(t *testing.T) {
	check := checkHomeVersion(context.Background(), homeRepo(t, "v0.1.0"), "v0.1.0", "git")
	if check.Status != StatusOK || check.Required {
		t.Fatalf("check = %+v, want an optional ok", check)
	}
	if !strings.Contains(check.Detail, "v0.1.0") {
		t.Fatalf("detail = %q, want the version", check.Detail)
	}
}

func TestHomeVersionDifferentTagWarns(t *testing.T) {
	check := checkHomeVersion(context.Background(), homeRepo(t, "v0.2.0"), "v0.1.0", "git")
	if check.Status != StatusWarn || check.Required {
		t.Fatalf("check = %+v, want an optional warn", check)
	}
	for _, want := range []string{"binary v0.1.0", "home tags v0.2.0", "git -C HOME_DIR fetch --tags", "git -C HOME_DIR checkout v0.1.0", "go install github.com/Harrison-Blair/hive@v0.1.0", "restart"} {
		if !strings.Contains(check.Detail, want) {
			t.Fatalf("detail = %q, want %q", check.Detail, want)
		}
	}
}

func TestHomeVersionNoTagWarnsWithCommit(t *testing.T) {
	home := homeRepo(t)
	commit := gitRun(t, home, "rev-parse", "--short", "HEAD")
	check := checkHomeVersion(context.Background(), home, "v0.1.0", "git")
	if check.Status != StatusWarn {
		t.Fatalf("check = %+v, want warn", check)
	}
	for _, want := range []string{"binary v0.1.0", "no tag at commit " + commit, "git -C HOME_DIR checkout v0.1.0"} {
		if !strings.Contains(check.Detail, want) {
			t.Fatalf("detail = %q, want %q", check.Detail, want)
		}
	}
}

func TestHomeVersionTwoTagsOneMatchPasses(t *testing.T) {
	check := checkHomeVersion(context.Background(), homeRepo(t, "v0.1.0-rc.1", "v0.1.0"), "v0.1.0", "git")
	if check.Status != StatusOK {
		t.Fatalf("check = %+v, want ok", check)
	}
}

func TestHomeVersionTwoTagsNoMatchWarns(t *testing.T) {
	check := checkHomeVersion(context.Background(), homeRepo(t, "v0.1.0-rc.1", "v0.1.0"), "v0.1.0-rc", "git")
	if check.Status != StatusWarn || !strings.Contains(check.Detail, "home tags v0.1.0, v0.1.0-rc.1") {
		t.Fatalf("check = %+v, want warn that lists both tags", check)
	}
}

func TestHomeVersionGitFailureWarns(t *testing.T) {
	notRepo := t.TempDir()
	check := checkHomeVersion(context.Background(), notRepo, "v0.1.0", "git")
	if check.Status != StatusWarn || check.Required || !strings.Contains(check.Detail, "not a git repository") {
		t.Fatalf("check = %+v, want an optional warn with the git error", check)
	}
	missing := checkHomeVersion(context.Background(), homeRepo(t, "v0.1.0"), "v0.1.0", "/nonexistent/git")
	if missing.Status != StatusWarn || missing.Detail == "" {
		t.Fatalf("check = %+v, want warn with the error", missing)
	}
}

func TestHomeVersionInDoctorRun(t *testing.T) {
	t.Setenv("HERDR_SOCKET_PATH", "")
	result := Run(context.Background(), newHome(t), nil)
	check := findCheck(t, result, "home version")
	if check.Required {
		t.Fatalf("check = %+v, want optional", check)
	}
}
