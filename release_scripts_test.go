package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// skipWithoutReleaseTools skips a release script test when bash or git is absent.
func skipWithoutReleaseTools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"bash", "git"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not installed", tool)
		}
	}
}

// runIn runs a command in dir and returns its combined output.
func runIn(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// tagRepo makes a clone of a bare remote. The clone has one commit on
// origin/main and one commit on the branch side that origin/main does not have.
func tagRepo(t *testing.T) (repo, mainCommit, sideCommit string) {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	repo = filepath.Join(root, "repo")
	runIn(t, root, "git", "init", "-q", "--bare", "-b", "main", remote)
	runIn(t, root, "git", "clone", "-q", remote, repo)
	for _, args := range [][]string{
		{"config", "user.name", "Test"},
		{"config", "user.email", "test@example.com"},
		{"config", "tag.gpgSign", "false"},
		{"config", "commit.gpgSign", "false"},
		{"checkout", "-q", "-b", "main"},
		{"commit", "-q", "--allow-empty", "-m", "main"},
		{"push", "-q", "origin", "main"},
	} {
		runIn(t, repo, "git", args...)
	}
	mainCommit = runIn(t, repo, "git", "rev-parse", "HEAD")
	runIn(t, repo, "git", "checkout", "-q", "-b", "side")
	runIn(t, repo, "git", "commit", "-q", "--allow-empty", "-m", "side")
	sideCommit = runIn(t, repo, "git", "rev-parse", "HEAD")
	runIn(t, repo, "git", "checkout", "-q", "main")
	return repo, mainCommit, sideCommit
}

func TestCheckTag(t *testing.T) {
	skipWithoutReleaseTools(t)
	script, err := filepath.Abs(".github/scripts/check-tag.sh")
	if err != nil {
		t.Fatal(err)
	}
	repo, mainCommit, sideCommit := tagRepo(t)

	tests := []struct {
		name      string
		tag       string
		commit    string
		annotated bool
		wantKind  string
		wantError string
	}{
		{name: "release", tag: "v1.20.3", commit: mainCommit, annotated: true, wantKind: "release"},
		{name: "zero release", tag: "v0.1.0", commit: mainCommit, annotated: true, wantKind: "release"},
		{name: "prerelease", tag: "v0.1.0-rc.12", commit: mainCommit, annotated: true, wantKind: "prerelease"},
		{name: "two parts", tag: "v1.2", commit: mainCommit, annotated: true, wantError: "does not match"},
		{name: "no v", tag: "1.2.3", commit: mainCommit, annotated: true, wantError: "does not match"},
		{name: "other suffix", tag: "v1.2.3-beta.1", commit: mainCommit, annotated: true, wantError: "does not match"},
		{name: "rc without number", tag: "v1.2.3-rc", commit: mainCommit, annotated: true, wantError: "does not match"},
		{name: "trailing text", tag: "v1.2.3x", commit: mainCommit, annotated: true, wantError: "does not match"},
		{name: "leading zero major", tag: "v01.2.3", commit: mainCommit, annotated: true, wantError: "leading zero"},
		{name: "leading zero patch", tag: "v1.2.03", commit: mainCommit, annotated: true, wantError: "leading zero"},
		{name: "leading zero rc", tag: "v1.2.3-rc.01", commit: mainCommit, annotated: true, wantError: "leading zero"},
		{name: "lightweight", tag: "v1.2.4", commit: mainCommit, annotated: false, wantError: "lightweight"},
		{name: "not on main", tag: "v1.2.5", commit: sideCommit, annotated: true, wantError: "not on origin/main"},
		{name: "missing tag", tag: "v1.2.6", wantError: "not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.commit != "" {
				if tt.annotated {
					runIn(t, repo, "git", "tag", "-a", "-m", tt.tag, tt.tag, tt.commit)
				} else {
					runIn(t, repo, "git", "tag", tt.tag, tt.commit)
				}
				runIn(t, repo, "git", "push", "-q", "origin", "refs/tags/"+tt.tag)
			}
			cmd := exec.Command("bash", script, tt.tag)
			cmd.Dir = repo
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("check-tag.sh %s: %v\n%s", tt.tag, err, stderr.String())
				}
				if got := strings.TrimSpace(stdout.String()); got != tt.wantKind {
					t.Fatalf("check-tag.sh %s printed %q, want %q", tt.tag, got, tt.wantKind)
				}
				return
			}
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("check-tag.sh %s: err %v, want exit status 1\nstdout %s\nstderr %s", tt.tag, err, stdout.String(), stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.wantError) {
				t.Fatalf("check-tag.sh %s stderr %q, want it to contain %q", tt.tag, stderr.String(), tt.wantError)
			}
		})
	}
}

func TestPackage(t *testing.T) {
	skipWithoutReleaseTools(t)
	for _, tool := range []string{"go", "tar", "gzip", "sha256sum"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not installed", tool)
		}
	}
	const tag = "v0.1.0-rc.1"
	commitTime, err := strconv.ParseInt(runIn(t, ".", "git", "show", "-s", "--format=%ct", "HEAD"), 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	license, err := os.ReadFile("LICENSE")
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	root := t.TempDir()
	first, second := filepath.Join(root, "first"), filepath.Join(root, "second")
	for _, dist := range []string{first, second} {
		runIn(t, ".", "bash", ".github/scripts/package.sh", tag, dist)
	}
	t.Logf("two package runs took %s", time.Since(start).Round(time.Second))

	archives := []string{"hive_" + tag + "_linux_amd64.tar.gz", "hive_" + tag + "_linux_arm64.tar.gz"}
	files := append(slices.Clone(archives), "checksums.txt")
	for _, dist := range []string{first, second} {
		entries, err := os.ReadDir(dist)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		if !slices.Equal(names, []string{"checksums.txt", archives[0], archives[1]}) {
			t.Fatalf("%s holds %v, want %v", dist, names, files)
		}
	}

	for _, name := range files {
		a, err := os.ReadFile(filepath.Join(first, name))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(second, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, b) {
			t.Errorf("%s differs between two package runs", name)
		}
	}

	checksums, err := os.ReadFile(filepath.Join(first, "checksums.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var wantChecksums strings.Builder
	for _, name := range archives {
		data, err := os.ReadFile(filepath.Join(first, name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		fmt.Fprintf(&wantChecksums, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	if string(checksums) != wantChecksums.String() {
		t.Errorf("checksums.txt is\n%s\nwant\n%s", checksums, wantChecksums.String())
	}

	for _, name := range archives {
		members := readArchive(t, filepath.Join(first, name))
		var memberNames []string
		for _, member := range members {
			memberNames = append(memberNames, member.header.Name)
			if member.header.Uid != 0 || member.header.Gid != 0 {
				t.Errorf("%s: %s has owner %d:%d, want 0:0", name, member.header.Name, member.header.Uid, member.header.Gid)
			}
			if member.header.ModTime.Unix() != commitTime {
				t.Errorf("%s: %s has time %d, want commit time %d", name, member.header.Name, member.header.ModTime.Unix(), commitTime)
			}
		}
		if !slices.Equal(memberNames, []string{"LICENSE", "hive"}) {
			t.Fatalf("%s holds %v, want [LICENSE hive]", name, memberNames)
		}
		if !bytes.Equal(members[0].data, license) {
			t.Errorf("%s: LICENSE differs from the LICENSE of the checkout", name)
		}
		if members[1].header.Mode&0o111 == 0 {
			t.Errorf("%s: hive mode %o is not executable", name, members[1].header.Mode)
		}
		if runtime.GOOS == "linux" && strings.HasSuffix(name, "_linux_"+runtime.GOARCH+".tar.gz") {
			binary := filepath.Join(t.TempDir(), "hive")
			if err := os.WriteFile(binary, members[1].data, 0o755); err != nil {
				t.Fatal(err)
			}
			out := runIn(t, ".", binary, "version")
			if !strings.HasPrefix(out, "hive "+tag+" (") {
				t.Errorf("%s: hive version printed %q, want the prefix %q", name, out, "hive "+tag+" (")
			}
		}
	}
}

type archiveMember struct {
	header *tar.Header
	data   []byte
}

// readArchive returns the members of a .tar.gz file in archive order.
func readArchive(t *testing.T, path string) []archiveMember {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	zr, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	// A gzip time from the clock would make two runs differ.
	if !zr.ModTime.IsZero() {
		t.Errorf("%s: gzip header time %s, want none", path, zr.ModTime)
	}
	tr := tar.NewReader(zr)
	var members []archiveMember
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return members
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		members = append(members, archiveMember{header: header, data: data})
	}
}
