package coordinator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// registerProject writes the record of project id with path, as
// hive project register writes it.
func registerProject(t *testing.T, home, id, path string) {
	t.Helper()
	dir := filepath.Join(home, "projects", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	record := `{"id":"` + id + `","path":"` + path + `","base":"main","registered_at":"2026-10-30T00:00:00Z"}`
	if err := os.WriteFile(filepath.Join(dir, "project.json"), []byte(record), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, parts ...string) string {
	t.Helper()
	dir := filepath.Join(parts...)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// nameOf returns the name that ResolveName gives and fails the test on an
// error.
func nameOf(t *testing.T, home, dir string) string {
	t.Helper()
	name, err := ResolveName(home, dir)
	if err != nil {
		t.Fatalf("ResolveName(%s) error = %v; want none", dir, err)
	}
	return name
}

func TestResolveNameUsesTheProjectOfTheCheckout(t *testing.T) {
	home := t.TempDir()
	src := t.TempDir()
	repo := mkdir(t, src, "hive")
	registerProject(t, home, "hive", repo)
	registerProject(t, home, "other", mkdir(t, src, "other"))
	for _, dir := range []string{repo, mkdir(t, repo, "internal", "coordinator"), repo + "/./internal/../"} {
		if got := nameOf(t, home, dir); got != "hive-coordinator" {
			t.Errorf("ResolveName(%s) = %q; want hive-coordinator", dir, got)
		}
	}
}

func TestResolveNameFollowsSymlinks(t *testing.T) {
	home := t.TempDir()
	src := t.TempDir()
	repo := mkdir(t, src, "real", "hive")
	mkdir(t, repo, "sub")
	link := filepath.Join(src, "link")
	if err := os.Symlink(filepath.Join(src, "real"), link); err != nil {
		t.Fatal(err)
	}
	// The record holds the link path, and the start is below the real path.
	registerProject(t, home, "hive", filepath.Join(link, "hive"))
	if got := nameOf(t, home, filepath.Join(repo, "sub")); got != "hive-coordinator" {
		t.Errorf("ResolveName(real path) = %q; want hive-coordinator", got)
	}
	// The start is below the link, and the record holds the real path.
	home2 := t.TempDir()
	registerProject(t, home2, "hive", repo)
	if got := nameOf(t, home2, filepath.Join(link, "hive", "sub")); got != "hive-coordinator" {
		t.Errorf("ResolveName(link path) = %q; want hive-coordinator", got)
	}
}

func TestResolveNameFallsBackOutsideAProject(t *testing.T) {
	home := t.TempDir()
	src := t.TempDir()
	registerProject(t, home, "hive", mkdir(t, src, "hive"))
	for name, dir := range map[string]string{
		"unrelated":      t.TempDir(),
		"sibling prefix": mkdir(t, src, "hive-old", "sub"),
		"parent":         src,
	} {
		if got := nameOf(t, home, dir); got != Name {
			t.Errorf("%s: ResolveName(%s) = %q; want %s", name, dir, got, Name)
		}
	}
	if got := nameOf(t, t.TempDir(), src); got != Name {
		t.Errorf("no projects: ResolveName = %q; want %s", got, Name)
	}
}

func TestResolveNamePicksTheInnermostProject(t *testing.T) {
	home := t.TempDir()
	outer := mkdir(t, t.TempDir(), "outer")
	inner := mkdir(t, outer, "vendor", "inner")
	registerProject(t, home, "outer", outer)
	registerProject(t, home, "inner", inner)
	if got := nameOf(t, home, mkdir(t, inner, "x")); got != "inner-coordinator" {
		t.Errorf("ResolveName(inner) = %q; want inner-coordinator", got)
	}
	if got := nameOf(t, home, mkdir(t, outer, "vendor")); got != "outer-coordinator" {
		t.Errorf("ResolveName(outer) = %q; want outer-coordinator", got)
	}
}

func TestResolveNameFallsBackWhenAmbiguous(t *testing.T) {
	home := t.TempDir()
	src := t.TempDir()
	repo := mkdir(t, src, "repo")
	link := filepath.Join(src, "alias")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatal(err)
	}
	// Two records name the same checkout through different paths.
	registerProject(t, home, "alpha", repo)
	registerProject(t, home, "beta", link)
	got, err := ResolveName(home, repo)
	if got != Name || err == nil || !strings.Contains(err.Error(), "alpha") || !strings.Contains(err.Error(), "beta") {
		t.Fatalf("ResolveName = %q, %v; want %s and an error that names both projects", got, err, Name)
	}
}

func TestResolveNameFallsBackWhenTheRegistryIsUnreadable(t *testing.T) {
	home := t.TempDir()
	dir := mkdir(t, home, "projects", "broken")
	if err := os.WriteFile(filepath.Join(dir, "project.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := ResolveName(home, t.TempDir()); got != Name || err == nil {
		t.Fatalf("ResolveName = %q, %v; want %s and an error", got, err, Name)
	}
}
