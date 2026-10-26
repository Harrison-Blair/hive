package write

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMemoryWriteChildProcess(t *testing.T) {
	if os.Getenv("HIVE_T07_MEMORY_CHILD") != "1" {
		t.Skip("child helper for the two-process test")
	}
	home := os.Getenv("HIVE_T07_MEMORY_HOME")
	name := os.Getenv("HIVE_T07_MEMORY_NAME")
	input := memoryContent(name, "Child "+name+".", "Body "+name+".")
	envelope := Write(home, WriteOptions{
		Scope:    "home",
		Name:     name,
		Input:    strings.NewReader(input),
		Deadline: 5 * time.Second,
	})
	if envelope.Outcome != "ok" {
		t.Fatalf("child Write() = %#v", envelope)
	}
}

func TestTwoProcessesWriteMemories(t *testing.T) {
	home := t.TempDir()
	first := childCommand(t, home, "first")
	second := childCommand(t, home, "second")
	var firstOut, secondOut bytes.Buffer
	first.Stdout = &firstOut
	first.Stderr = &firstOut
	second.Stdout = &secondOut
	second.Stderr = &secondOut
	if err := first.Start(); err != nil {
		t.Fatalf("start first child: %v", err)
	}
	if err := second.Start(); err != nil {
		_ = first.Process.Kill()
		_, _ = first.Process.Wait()
		t.Fatalf("start second child: %v", err)
	}
	if err := first.Wait(); err != nil {
		t.Fatalf("first child: %v\n%s", err, firstOut.String())
	}
	if err := second.Wait(); err != nil {
		t.Fatalf("second child: %v\n%s", err, secondOut.String())
	}
	got, err := os.ReadFile(filepath.Join(home, "memory", "MEMORY.md"))
	want := "- [first](first.md): Child first.\n- [second](second.md): Child second.\n"
	if err != nil || string(got) != want {
		t.Fatalf("index = %q, %v; want %q", got, err, want)
	}
}

func childCommand(t *testing.T, home, name string) *exec.Cmd {
	t.Helper()
	command := exec.Command(os.Args[0], "-test.run=^TestMemoryWriteChildProcess$", "-test.v")
	command.Env = append(os.Environ(),
		"HIVE_T07_MEMORY_CHILD=1",
		"HIVE_T07_MEMORY_HOME="+home,
		"HIVE_T07_MEMORY_NAME="+name,
	)
	command.Dir = home
	return command
}
