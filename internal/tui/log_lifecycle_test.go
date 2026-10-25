package tui

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// blockingLogHandler holds one log event until the test releases it. It proves
// that a command which writes a log event cannot keep the writer after
// runProgram returns.
type blockingLogHandler struct {
	next    slog.Handler
	started chan struct{}
	release chan struct{}
	written chan struct{}

	startOnce  sync.Once
	releaseOne sync.Once
	writeOnce  sync.Once
}

func (h *blockingLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *blockingLogHandler) Handle(ctx context.Context, record slog.Record) error {
	h.startOnce.Do(func() { close(h.started) })
	h.releaseOne.Do(func() { <-h.release })
	err := h.next.Handle(ctx, record)
	h.writeOnce.Do(func() { close(h.written) })
	return err
}

func (h *blockingLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &blockingLogHandler{next: h.next.WithAttrs(attrs), started: h.started, release: h.release, written: h.written}
}

func (h *blockingLogHandler) WithGroup(name string) slog.Handler {
	return &blockingLogHandler{next: h.next.WithGroup(name), started: h.started, release: h.release, written: h.written}
}

// TestRunProgramStopsLogWritesBeforeReturn checks that a log command after the
// return writes no event, and that the event of the program is kept (AC1, AC2).
func TestRunProgramStopsLogWritesBeforeReturn(t *testing.T) {
	home := newTestHome(t)
	model := New(home, nil)
	defer model.Stop()

	blocked := &blockingLogHandler{
		next:    model.logger.Handler(),
		started: make(chan struct{}),
		release: make(chan struct{}),
		written: make(chan struct{}),
	}
	model.logger = slog.New(blocked)

	input, keys := io.Pipe()
	done := make(chan error, 1)
	go func() {
		done <- runProgram(model, input, io.Discard, defaultWidth, defaultHeight)
	}()

	if _, err := fmt.Fprint(keys, "\x03"); err != nil {
		t.Fatalf("send the first quit key: %v", err)
	}
	select {
	case <-blocked.started:
	case <-time.After(5 * time.Second):
		t.Fatal("the quit log command did not start")
	}
	if _, err := fmt.Fprint(keys, "\x03"); err != nil {
		t.Fatalf("send the second quit key: %v", err)
	}
	if err := keys.Close(); err != nil {
		t.Fatalf("close the input: %v", err)
	}
	select {
	case err := <-done:
		close(blocked.release)
		<-blocked.written
		t.Fatalf("runProgram returned before the log write ended: %v", err)
	case <-time.After(2 * time.Second):
		close(blocked.release)
	}
	select {
	case <-blocked.written:
	case <-time.After(5 * time.Second):
		t.Fatal("the log event did not reach the file")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runProgram: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runProgram did not return")
	}

	path := onlyLogFile(t, home)
	before := readLogFile(t, path)
	if !strings.Contains(before, `"action":"quit request"`) {
		t.Fatalf("log file after the return omits the quit event:\n%s", before)
	}

	command := model.logAction("after return")
	if command != nil {
		if msg := command(); msg != nil {
			t.Fatalf("a log command after the return gave %T; want no write", msg)
		}
	}
	if after := readLogFile(t, path); after != before {
		t.Fatalf("a log command after the return changed the log file:\nbefore:\n%s\nafter:\n%s", before, after)
	}

	model.Stop()
	record := slog.NewRecord(time.Now(), slog.LevelDebug, "after close", 0)
	if err := model.logger.Handler().Handle(context.Background(), record); err == nil {
		t.Fatal("the closed writer accepted an event after Stop")
	}
}

// TestRunHomeKeepsLogEventsBeforeReturn checks that the quit event of the
// program is in the log file after the return, and that the writer is closed.
func TestRunHomeKeepsLogEventsBeforeReturn(t *testing.T) {
	home := newTestHome(t)
	var out strings.Builder
	if err := RunHome(home, strings.NewReader("\x03\x03"), &out); err != nil {
		t.Fatalf("RunHome: %v", err)
	}
	path := onlyLogFile(t, home)
	after := readLogFile(t, path)
	if !strings.Contains(after, `"action":"quit request"`) {
		t.Fatalf("log file after the return omits the quit event:\n%s", after)
	}
}

// TestInputEndedTickQuitsWithoutQuitAction isolates hypothesis A: the input
// end tick can quit the program before a ctrl+c key Update runs. The tick path
// returns tea.Quit and creates no quit action.
func TestInputEndedTickQuitsWithoutQuitAction(t *testing.T) {
	home := newTestHome(t)
	model := New(home, nil)
	defer model.Stop()
	ended := make(chan struct{})
	close(ended)
	model.inputEnded = ended

	_, cmd := model.Update(tickMsg{})
	if cmd == nil {
		t.Fatal("the input-end tick did not return a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("the input-end tick returned %T; want tea.QuitMsg", cmd())
	}
	if model.quitNotice {
		t.Fatal("the input-end tick set the quit notice")
	}
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "logs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(home, ".hive", "logs", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), `"action":"quit request"`) {
			t.Fatalf("the input-end tick wrote a quit action:\n%s", data)
		}
	}
}

// TestQueuedQuitActionSurvivesClose isolates hypothesis B: the first ctrl+c
// is accepted and creates the quit log command, but that command has not
// started when close runs. The regression holds the real command at its start
// hook, closes the real gate, then releases it. A correct drain retains the
// quit event; the current gate drops it.
func TestQueuedQuitActionSurvivesClose(t *testing.T) {
	home := newTestHome(t)
	model := New(home, nil)
	defer model.Stop()
	model.logger.Debug("seed")

	started := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})
	closeWaiting := make(chan struct{})
	closeSealed := make(chan struct{})
	var startOnce, finishOnce, waitingOnce, sealedOnce sync.Once
	model.logs.testHook = func(event string) {
		switch event {
		case "command-start":
			startOnce.Do(func() { close(started) })
			<-release
		case "command-end", "command-dropped":
			finishOnce.Do(func() { close(finished) })
		case "close-waiting":
			waitingOnce.Do(func() { close(closeWaiting) })
		case "close-sealed":
			sealedOnce.Do(func() { close(closeSealed) })
		}
	}

	model, cmd := raise(t, model, ctrlC())
	if !model.quitNotice {
		t.Fatal("the first ctrl+c was not accepted")
	}
	if cmd == nil {
		t.Fatal("the first ctrl+c returned no command")
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("the first ctrl+c command gave %T; want tea.BatchMsg", msg)
	}
	for _, sub := range batch {
		go sub()
	}
	<-started

	closed := make(chan struct{})
	go func() {
		model.logs.close()
		close(closed)
	}()
	select {
	case <-closeWaiting:
	case <-closeSealed:
	}
	close(release)
	<-finished
	<-closed
	model.Stop()

	path := onlyLogFile(t, home)
	after := readLogFile(t, path)
	if !strings.Contains(after, `"action":"quit request"`) {
		t.Fatalf("the accepted pre-close quit action was dropped:\n%s", after)
	}
}

// TestNeverDispatchedQuitActionKeepsOneEvent covers the case where Bubble Tea
// quits before it dispatches the command: the accepted first ctrl+c must write
// the quit event exactly once at creation, close must return, and invoking the
// held command after close must not append a second event.
func TestNeverDispatchedQuitActionKeepsOneEvent(t *testing.T) {
	home := newTestHome(t)
	model := New(home, nil)
	defer model.Stop()
	model.logger.Debug("seed")

	model, cmd := raise(t, model, ctrlC())
	if !model.quitNotice {
		t.Fatal("the first ctrl+c was not accepted")
	}
	if cmd == nil {
		t.Fatal("the first ctrl+c returned no command")
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("the first ctrl+c command gave %T; want tea.BatchMsg", msg)
	}
	if len(batch) == 0 {
		t.Fatal("the first ctrl+c command batch is empty")
	}
	held := batch[0]
	// Do not invoke held. Close must return within the existing test bound.
	closed := make(chan struct{})
	go func() {
		model.Stop()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("close did not return within 5s")
	}

	path := onlyLogFile(t, home)
	before := readLogFile(t, path)
	if got := strings.Count(before, `"action":"quit request"`); got != 1 {
		t.Fatalf("quit events after close = %d; want exactly 1:\n%s", got, before)
	}
	held()
	after := readLogFile(t, path)
	if after != before {
		t.Fatalf("late command invocation changed the log bytes:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if got := strings.Count(after, `"action":"quit request"`); got != 1 {
		t.Fatalf("quit events after late command invocation = %d; want exactly 1:\n%s", got, after)
	}
}

// TestHeaderLoadAfterCloseWritesNoLog checks that a header load which starts
// after the close cannot make the store migration writer create a log file.
// The header load runs store.Open, which logs a schema migration.
func TestHeaderLoadAfterCloseWritesNoLog(t *testing.T) {
	home := newTestHome(t)
	model := New(home, nil)

	logsDir := filepath.Join(home, ".hive", "logs")
	if err := os.RemoveAll(logsDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(home, ".hive", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "schema.json"), []byte("{\n  \"version\": 1\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := model.startHeaderLoad()
	if cmd == nil {
		t.Fatal("no header load")
	}
	model.Stop()
	cmd()

	entries, err := os.ReadDir(logsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".jsonl") {
			t.Fatalf("a header load after close wrote %s", entry.Name())
		}
	}
}

// onlyLogFile gives the one daily log file of home.
func onlyLogFile(t *testing.T, home string) string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".hive", "logs"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("log files = %d; want one", len(entries))
	}
	return filepath.Join(home, ".hive", "logs", entries[0].Name())
}

func readLogFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
