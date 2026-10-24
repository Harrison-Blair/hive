package cmd

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mainFixture starts on Agents with project agents to peek, attach, and
// restart. The Gf boards of task 137 are Chat boards with no project agent.
const mainFixture = "../.hive/spec/tui-v2/design/Main.fixture.json"

// gfChatFixture is the fixture of the GfChat board of task 137.
const gfChatFixture = "../.hive/spec/golden-frames/design/GfChat.fixture.json"

// TestDemoActionsChangeNoFileAndStartNoProcess runs the CLI entry with --demo
// and presses the chat send, restart, attach, and peek keys. The keys follow
// the golden-frames chords: p peeks and a attaches on Agents, ^y opens
// Health and ^] focuses its agents table where r restarts, ^y again returns
// to Chat, where the input line has the focus, and enter sends.
func TestDemoActionsChangeNoFileAndStartNoProcess(t *testing.T) {
	home := newRootTestHome(t)
	methods := newRootFakeHerdr(t)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PANE_ID", "w1:p1")

	fixtureDir := t.TempDir()
	fixture := filepath.Join(fixtureDir, "Main.fixture.json")
	requireSpecFile(t, mainFixture)
	data, err := os.ReadFile(mainFixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture, data, 0o644); err != nil {
		t.Fatal(err)
	}

	binDir := t.TempDir()
	marker := filepath.Join(t.TempDir(), "processes")
	for _, name := range []string{"herdr", "hive", "git", "claude", "codex", "pi", "gh", "sh", "bash"} {
		script := "#!/bin/sh\necho \"$0 $*\" >> " + marker + "\n"
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir)

	before := snapshotFiles(t, home, fixtureDir)
	var out bytes.Buffer
	keys := "pa\x19\x1dr\x19hi\r"
	status := executeWithStreams([]string{"--demo", fixture}, strings.NewReader(keys), &out, &out)
	if status != 0 {
		t.Fatalf("status = %d; want 0; output %q", status, out.String())
	}
	after := snapshotFiles(t, home, fixtureDir)

	logFiles := []string{}
	for path, sum := range after {
		if before[path] == sum {
			continue
		}
		rel, _ := filepath.Rel(home, path)
		if filepath.Dir(rel) != filepath.Join(".hive", "logs") || filepath.Ext(rel) != ".jsonl" {
			t.Fatalf("demo mode changed %s; want only the log file", path)
		}
		logFiles = append(logFiles, path)
	}
	for path := range before {
		if _, ok := after[path]; !ok {
			t.Fatalf("demo mode removed %s", path)
		}
	}
	if len(logFiles) != 1 {
		t.Fatalf("changed log files = %v; want one", logFiles)
	}

	blocked := map[string]bool{}
	file, err := os.Open(logFiles[0])
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
		var event map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("log line %q: %v", scanner.Text(), err)
		}
		if event["demo"] != true {
			t.Fatalf("log event %s has no demo=true", scanner.Text())
		}
		if event["action"] == "demo action blocked" {
			blocked[event["blocked"].(string)] = true
		}
	}
	if lines == 0 {
		t.Fatal("demo mode wrote no log event")
	}
	for _, action := range []string{"peek", "attach", "restart", "chat send"} {
		if !blocked[action] {
			t.Fatalf("blocked actions = %v; want %q", blocked, action)
		}
	}

	if !strings.Contains(out.String(), "demo mode: no action") {
		t.Fatalf("output %q lacks the demo notice", out.String())
	}

	if got, err := os.ReadFile(marker); err == nil {
		t.Fatalf("demo mode started processes:\n%s", got)
	}
	if got := methods(); len(got) != 0 {
		t.Fatalf("herdr methods = %v; want none", got)
	}
}

func TestDemoFlagsAreHiddenFromHelp(t *testing.T) {
	newRootTestHome(t)
	var out bytes.Buffer
	if status := executeWithStreams([]string{"--help"}, strings.NewReader(""), &out, &out); status != 0 {
		t.Fatalf("status = %d; output %q", status, out.String())
	}
	for _, flag := range []string{"--demo", "--size"} {
		if strings.Contains(out.String(), flag) {
			t.Fatalf("help lists %s:\n%s", flag, out.String())
		}
	}
}

// --demo with the GfChat fixture draws the golden-frames Chat at the board
// size: the composer with the draft and the status row. Two ctrl+c quit.
func TestDemoRunsTheGfChatBoard(t *testing.T) {
	requireSpecFile(t, gfChatFixture)
	newRootTestHome(t)
	var out bytes.Buffer
	status := executeWithStreams([]string{"--demo", gfChatFixture}, strings.NewReader("\x03\x03"), &out, &out)
	if status != 0 {
		t.Fatalf("status = %d; output %q", status, out.String())
	}
	for _, want := range []string{"> Next, the release notes.", "● working", "^g agents"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output lacks %q:\n%s", want, out.String())
		}
	}
}

func TestDemoBadSizeNamesTheValue(t *testing.T) {
	for _, value := range []string{"12x", "x40", "120x40x2", "0x40", "abcxdef", "120*40"} {
		t.Run(value, func(t *testing.T) {
			newRootTestHome(t)
			var stdout, stderr bytes.Buffer
			status := executeWithStreams([]string{"--demo", gfChatFixture, "--size", value}, strings.NewReader(""), &stdout, &stderr)
			if status == 0 {
				t.Fatalf("status = 0; want a failure for --size %q", value)
			}
			if strings.Contains(stdout.String(), "\x1b[?25l") {
				t.Fatalf("bad size started the TUI: %q", stdout.String())
			}
			if !strings.Contains(stderr.String(), `"`+value+`"`) {
				t.Fatalf("stderr = %q; want the value %q", stderr.String(), value)
			}
		})
	}
}

func TestSizeWithoutDemoFails(t *testing.T) {
	newRootTestHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--size", "120x40"}, strings.NewReader(""), &stdout, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), "--size needs --demo") {
		t.Fatalf("status = %d, stderr %q; want --size needs --demo", status, stderr.String())
	}
}

func TestDemoUnknownFixtureFieldNamesTheField(t *testing.T) {
	newRootTestHome(t)
	fixture := filepath.Join(t.TempDir(), "bad.fixture.json")
	if err := os.WriteFile(fixture, []byte(`{"board":"Main.dc.html","colour":"red"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--demo", fixture}, strings.NewReader(""), &stdout, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), `"colour"`) {
		t.Fatalf("status = %d, stderr %q; want an error that names \"colour\"", status, stderr.String())
	}
}

func snapshotFiles(t *testing.T, roots ...string) map[string][32]byte {
	t.Helper()
	files := map[string][32]byte{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[path] = sha256.Sum256(data)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

// requireSpecFile skips the test when a file or a folder of the local spec is
// missing. Git does not track .hive/spec/, so a clean clone has no grid and
// no fixture. The test runs as before when the file exists.
func requireSpecFile(t testing.TB, path string) {
	t.Helper()
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		t.Skipf("local spec file missing: %s: Git does not track .hive/spec/", path)
	}
}
