package write

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	storelib "github.com/Harrison-Blair/hive/internal/lib/store"
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
	if err := writeFileFixture(t, target, old); err != nil {
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
			if readErr := <-errCh; readErr != nil {
				t.Fatalf("Write() at %d = %#v; reader error = %v", i, envelope, readErr)
			}
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

func TestHandoffLogsProjectID(t *testing.T) {
	home := t.TempDir()
	registry, err := storelib.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Write("projects", "demo", records.Project{ID: "demo", Path: home, Base: "dev"}); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "projects", "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	var log bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	success := Write(home, WriteOptions{
		Scope:    "project:demo",
		Input:    strings.NewReader("Project handoff.\n"),
		Deadline: time.Second,
		Logger:   logger,
	})
	if success.Outcome != "ok" {
		t.Fatalf("Write() = %#v; want success", success)
	}
	failure := Write(home, WriteOptions{
		Scope:    "project:demo",
		Input:    strings.NewReader(""),
		Deadline: time.Second,
		Logger:   logger,
	})
	if failure.Outcome != "failed" || failure.Phase != "input" {
		t.Fatalf("Write() = %#v; want input failure", failure)
	}
	projectEvents := 0
	for _, event := range handoffLogEvents(t, log.String()) {
		if event["scope"] != "project:demo" {
			continue
		}
		projectEvents++
		if event["project_id"] != "demo" {
			t.Fatalf("event = %#v; want project_id demo", event)
		}
	}
	if projectEvents != 2 {
		t.Fatalf("project events = %d; want handoff effect and input error", projectEvents)
	}
}

func handoffLogEvents(t *testing.T, text string) []map[string]string {
	t.Helper()
	events := []map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		if line == "" {
			continue
		}
		var event map[string]string
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line = %q: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}

func writeHandoff(t *testing.T, home, content string) output.Envelope {
	t.Helper()
	return Write(home, WriteOptions{
		Scope:    "home",
		Input:    strings.NewReader(content),
		Deadline: time.Second,
	})
}

func writeFileFixture(t *testing.T, path string, data []byte) error {
	t.Helper()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
