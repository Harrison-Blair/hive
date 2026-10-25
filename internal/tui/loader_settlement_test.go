package tui

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// TestRunProgramSettlesAdmittedHeaderLoad covers AC1 on the real header
// loader. The scratch home needs the store migration, so the real loader and
// store write state files and log events. The barrier only schedules the hold;
// the real loader runs after the test releases it.
func TestRunProgramSettlesAdmittedHeaderLoad(t *testing.T) {
	home := newTestHome(t)
	stateDir := filepath.Join(home, ".hive", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "schema.json"), []byte("{\n  \"version\": 1\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	model := New(home, nil)
	defer model.Stop()
	realLoader := model.headerLoader
	started := make(chan struct{})
	release := make(chan struct{})
	loaded := make(chan struct{})
	model.headerLoader = func(ctx context.Context, h string) (headerState, error) {
		close(started)
		<-release
		state, err := realLoader(ctx, h)
		close(loaded)
		return state, err
	}

	input, keys := io.Pipe()
	defer func() { _ = keys.Close() }()
	done := make(chan error, 1)
	go func() {
		done <- runProgram(model, input, io.Discard, defaultWidth, defaultHeight)
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("the admitted header load did not start")
	}
	if _, err := fmt.Fprint(keys, "\x03\x03"); err != nil {
		t.Fatalf("send the quit keys: %v", err)
	}

	returned := false
	select {
	case err := <-done:
		returned = true
		if err != nil {
			t.Fatalf("runProgram: %v", err)
		}
	case <-time.After(2 * time.Second):
	}
	if returned {
		close(release)
		select {
		case <-loaded:
		case <-time.After(5 * time.Second):
			t.Fatal("the held header load did not finish")
		}
		t.Fatal("runProgram returned while the admitted header load was outstanding")
	}
	close(release)
	select {
	case <-loaded:
	case <-time.After(5 * time.Second):
		t.Fatal("the held header load did not finish")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runProgram: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runProgram did not return after the loader settled")
	}
	select {
	case <-loaded:
	default:
		t.Fatal("runProgram returned before the header loader completed")
	}

	version, err := store.SchemaVersion(home)
	if err != nil {
		t.Fatal(err)
	}
	if version != store.CurrentSchemaVersion {
		t.Fatalf("schema version after return = %d; want %d", version, store.CurrentSchemaVersion)
	}
	logText := readLogFile(t, onlyLogFile(t, home))
	for _, want := range []string{`"component":"store"`, `"msg":"state migrate start"`, `"msg":"state migrate done"`} {
		if !strings.Contains(logText, want) {
			t.Fatalf("log after return omits %s:\n%s", want, logText)
		}
	}
	record := slog.NewRecord(time.Now(), slog.LevelDebug, "late tui event", 0)
	if err := model.logger.Handler().Handle(context.Background(), record); err == nil {
		t.Fatal("the TUI writer accepted an event after runProgram returned")
	}
}
