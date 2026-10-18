package handoff

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
)

func TestHandoffWriteCreatesAndReplaces(t *testing.T) {
	home := t.TempDir()
	first := "Open decisions: none.\n"
	envelope := writeHandoff(t, home, first)
	if envelope.Outcome != "ok" || len(envelope.Effects) != 1 {
		t.Fatalf("Write() = %#v; want one effect", envelope)
	}
	want := output.Effect{Resource: "handoff", Path: "handoff.md", Action: "updated"}
	if envelope.Effects[0] != want {
		t.Fatalf("effect = %#v; want %#v", envelope.Effects[0], want)
	}
	got, err := os.ReadFile(filepath.Join(home, "handoff.md"))
	if err != nil || string(got) != first {
		t.Fatalf("handoff = %q, %v; want %q", got, err, first)
	}

	second := "The next session starts here.\n"
	envelope = writeHandoff(t, home, second)
	if envelope.Outcome != "ok" {
		t.Fatalf("second Write() = %#v", envelope)
	}
	got, err = os.ReadFile(filepath.Join(home, "handoff.md"))
	if err != nil || string(got) != second {
		t.Fatalf("handoff = %q, %v; want replacement %q", got, err, second)
	}
}

func TestHandoffEmptyInputWritesNothing(t *testing.T) {
	home := t.TempDir()
	old := "Old handoff.\n"
	if envelope := writeHandoff(t, home, old); envelope.Outcome != "ok" {
		t.Fatalf("setup Write() = %#v", envelope)
	}
	envelope := writeHandoff(t, home, "")
	if envelope.Outcome != "failed" || envelope.Phase != "input" || envelope.Error != "empty input" {
		t.Fatalf("Write() = %#v; want empty input", envelope)
	}
	got, err := os.ReadFile(filepath.Join(home, "handoff.md"))
	if err != nil || string(got) != old {
		t.Fatalf("handoff = %q, %v; want old content", got, err)
	}
}

func TestHandoffOversizedInputKeepsOld(t *testing.T) {
	home := t.TempDir()
	old := "Old handoff.\n"
	if envelope := writeHandoff(t, home, old); envelope.Outcome != "ok" {
		t.Fatalf("setup Write() = %#v", envelope)
	}
	envelope := writeHandoff(t, home, strings.Repeat("x", 1048577))
	if envelope.Outcome != "failed" || envelope.Phase != "input" || envelope.Error != "input too large" {
		t.Fatalf("Write() = %#v; want input too large", envelope)
	}
	got, err := os.ReadFile(filepath.Join(home, "handoff.md"))
	if err != nil || string(got) != old {
		t.Fatalf("handoff = %q, %v; want old content", got, err)
	}
}

func TestHandoffTimeoutKeepsOld(t *testing.T) {
	home := t.TempDir()
	old := "Old handoff.\n"
	if envelope := writeHandoff(t, home, old); envelope.Outcome != "ok" {
		t.Fatalf("setup Write() = %#v", envelope)
	}
	reader, writer := io.Pipe()
	defer func() { _ = reader.Close() }()
	envelope := Write(home, WriteOptions{
		Scope:    "home",
		Input:    reader,
		Deadline: 100 * time.Millisecond,
	})
	_ = writer.Close()
	if envelope.Outcome != "failed" || envelope.Phase != "input" || envelope.Error != "input timeout" {
		t.Fatalf("Write() = %#v; want input timeout", envelope)
	}
	got, err := os.ReadFile(filepath.Join(home, "handoff.md"))
	if err != nil || string(got) != old {
		t.Fatalf("handoff = %q, %v; want old content", got, err)
	}
}

func TestHandoffStaticDirectoryRefused(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "handoff.md")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	envelope := writeHandoff(t, home, "New handoff.\n")
	if envelope.Outcome != "failed" {
		t.Fatalf("Write() = %#v; want refusal", envelope)
	}
	if info, err := os.Lstat(path); err != nil || !info.IsDir() {
		t.Fatalf("handoff.md = %#v, %v; want the directory kept", info, err)
	}
}

func TestHandoffStaticLinkRefused(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(t.TempDir(), "target.md")
	old := []byte("old target\n")
	if err := os.WriteFile(target, old, 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "handoff.md")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	envelope := writeHandoff(t, home, "New handoff.\n")
	if envelope.Outcome != "failed" {
		t.Fatalf("Write() = %#v; want refusal", envelope)
	}
	got, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(got, old) {
		t.Fatalf("target = %q, %v; want old bytes", got, err)
	}
	if info, err := os.Lstat(path); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("handoff.md = %#v, %v; want the symbolic link kept", info, err)
	}
}

func TestHandoffConcurrentReplacement(t *testing.T) {
	home := t.TempDir()
	old := "old handoff\n"
	if envelope := writeHandoff(t, home, old); envelope.Outcome != "ok" {
		t.Fatalf("setup Write() = %#v", envelope)
	}
	newContent := "new handoff\n"
	path := filepath.Join(home, "handoff.md")
	done := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		for {
			select {
			case <-done:
				errCh <- nil
				return
			default:
				data, err := os.ReadFile(path)
				if err != nil {
					errCh <- err
					return
				}
				if string(data) != old && string(data) != newContent {
					errCh <- &unexpectedContent{got: string(data)}
					return
				}
			}
		}
	}()
	for i := 0; i < 200; i++ {
		envelope := writeHandoff(t, home, newContent)
		if envelope.Outcome != "ok" {
			close(done)
			<-errCh
			t.Fatalf("Write() at %d = %#v", i, envelope)
		}
	}
	close(done)
	if err := <-errCh; err != nil {
		t.Fatalf("reader: %v", err)
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".handoff.md.tmp-") {
			t.Fatalf("temporary file remains: %s", entry.Name())
		}
	}
}

type unexpectedContent struct {
	got string
}

func (e *unexpectedContent) Error() string {
	return "partial handoff content: " + e.got
}

func writeHandoff(t *testing.T, home, content string) output.Envelope {
	t.Helper()
	return Write(home, WriteOptions{
		Scope:    "home",
		Input:    strings.NewReader(content),
		Deadline: time.Second,
	})
}
