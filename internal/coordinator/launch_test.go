package coordinator

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

type execCall struct {
	dir, binary string
	args        []string
}

// launchOptions runs Run in the root pane of a fake herdr workspace, so no
// test reaches the herdr of the user.
func launchOptions(t *testing.T, home, dir string, pick func(Choice, bool) (Choice, bool, error), calls *[]execCall) Options {
	t.Helper()
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("home", home, false)
	return Options{
		Home: home,
		Dir:  dir,
		Pane: pane.ID,
		Pick: pick,
		Exec: func(dir, binary string, args []string) error {
			*calls = append(*calls, execCall{dir, binary, args})
			return nil
		},
		Now:    func() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) },
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestRunWritesRecordAndExecsHarness(t *testing.T) {
	cases := []struct {
		choice     Choice
		wantBinary string
		wantArgs   []string
	}{
		{Choice{"pi", "openai-codex/gpt-6-sol"}, "pi", []string{"--model", "openai-codex/gpt-6-sol"}},
		{Choice{"claude", "claude-opus-5-5"}, "claude", []string{"--dangerously-skip-permissions", "--model", "claude-opus-5-5"}},
		{Choice{"codex", "gpt-6-sol"}, "codex", []string{"--no-daemon", "--dangerously-bypass-approvals-and-sandbox", "--model", "gpt-6-sol"}},
	}
	for _, tc := range cases {
		t.Run(tc.choice.Harness, func(t *testing.T) {
			home := t.TempDir()
			dir := t.TempDir()
			var calls []execCall
			opts := launchOptions(t, home, dir, func(Choice, bool) (Choice, bool, error) { return tc.choice, true, nil }, &calls)
			if err := Run(opts); err != nil {
				t.Fatal(err)
			}
			record, err := agent.Load(home, "coordinator")
			if err != nil {
				t.Fatal(err)
			}
			want := agent.Record{Name: "coordinator", Pane: opts.Pane, Harness: tc.choice.Harness, Model: tc.choice.Model, Cwd: dir, StartedAt: time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)}
			if record != want {
				t.Fatalf("record = %+v; want %+v", record, want)
			}
			if len(calls) != 1 || calls[0].dir != dir || calls[0].binary != tc.wantBinary || !slices.Equal(calls[0].args, tc.wantArgs) {
				t.Fatalf("exec calls = %+v; want one %s %v in %s", calls, tc.wantBinary, tc.wantArgs, dir)
			}
			saved, ok, err := LoadChoice(home)
			if err != nil || !ok || saved != tc.choice {
				t.Fatalf("saved choice = %+v %v %v; want %+v", saved, ok, err, tc.choice)
			}
		})
	}
}

func TestRunPassesLastChoiceToPicker(t *testing.T) {
	home := t.TempDir()
	last := Choice{"claude", "claude-sonnet-5-5"}
	if err := SaveChoice(home, last); err != nil {
		t.Fatal(err)
	}
	var got Choice
	var gotOK bool
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(c Choice, ok bool) (Choice, bool, error) {
		got, gotOK = c, ok
		return Choice{}, false, nil
	}, &calls)
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	if !gotOK || got != last {
		t.Fatalf("picker got %+v %v; want %+v true", got, gotOK, last)
	}
}

func TestRunQuitWritesNothingAndDoesNotExec(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{}, false, nil }, &calls)
	if err := Run(opts); err != nil {
		t.Fatalf("Run = %v; want nil on quit", err)
	}
	if len(calls) != 0 {
		t.Fatalf("exec calls = %+v; want none", calls)
	}
	if _, err := agent.Load(home, "coordinator"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("Load = %v; want ErrNotFound", err)
	}
	if _, err := os.Stat(filepath.Join(home, "coordinator-choice.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("choice file stat = %v; want not exist", err)
	}
}

func TestRunReturnsPickerAndExecErrors(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{}, false, errors.New("tty gone") }, &calls)
	if err := Run(opts); err == nil || err.Error() != "picker: tty gone" {
		t.Fatalf("Run = %v; want picker: tty gone", err)
	}
	opts = launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
	opts.Exec = func(string, string, []string) error { return errors.New("no such file") }
	if err := Run(opts); err == nil {
		t.Fatal("Run = nil; want the exec error")
	}
}

func TestRunRejectsUnknownHarness(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"vim", "x"}, true, nil }, &calls)
	if err := Run(opts); err == nil || len(calls) != 0 {
		t.Fatalf("Run = %v, calls %v; want error and no exec", err, calls)
	}
	if _, err := agent.Load(home, "coordinator"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("Load = %v; want no record", err)
	}
}

func TestRunLabelsItsTabCoordinatorBeforeExec(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
	// This fake replaces the fake of launchOptions as HERDR_SOCKET_PATH.
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("my workspace", home, false)
	opts.Pane = pane.ID
	opts.Exec = func(dir, binary string, args []string) error {
		got, _ := fake.Pane(pane.ID)
		if got.TabLabel != "coordinator" {
			t.Errorf("tab label at exec = %q, want coordinator", got.TabLabel)
		}
		calls = append(calls, execCall{dir, binary, args})
		return nil
	}
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("exec calls = %+v; want one", calls)
	}
	if ws := fake.Workspaces(); len(ws) != 1 || ws[0].Label != "my workspace" {
		t.Fatalf("workspaces = %+v; want the label unchanged", ws)
	}
}

func TestRunExecsWhenTheTabLabelFails(t *testing.T) {
	for _, method := range []string{"pane.get", "tab.rename"} {
		t.Run(method, func(t *testing.T) {
			home := t.TempDir()
			var calls []execCall
			var logs strings.Builder
			opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
			fake := agenttest.NewHerdr(t)
			_, pane := fake.AddWorkspace("w", home, false)
			fake.Fail(method, "internal_error", "boom")
			opts.Pane = pane.ID
			opts.Logger = slog.New(slog.NewJSONHandler(&logs, nil))
			if err := Run(opts); err != nil {
				t.Fatal(err)
			}
			if len(calls) != 1 {
				t.Fatalf("exec calls = %+v; want one", calls)
			}
			if got, _ := fake.Pane(pane.ID); got.TabLabel != "" {
				t.Fatalf("tab label = %q; want none", got.TabLabel)
			}
			if !strings.Contains(logs.String(), `"level":"WARN"`) || !strings.Contains(logs.String(), "boom") {
				t.Fatalf("logs = %s; want a WARN event with the error", logs.String())
			}
		})
	}
}

func TestRunWithoutPaneDoesNotLabelATab(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
	fake := agenttest.NewHerdr(t)
	opts.Pane = ""
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("exec calls = %+v; want one", calls)
	}
	if got := len(fake.Requests("pane.get")) + len(fake.Requests("tab.rename")); got != 0 {
		t.Fatalf("herdr calls = %d; want none", got)
	}
}

func TestRunExecsTheHarnessWithTheGuide(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	var logs strings.Builder
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"codex", "gpt-x"}, true, nil }, &calls)
	opts.Guide = "say \"hi\"\nnow"
	opts.Logger = slog.New(slog.NewJSONHandler(&logs, nil))
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	want := []string{"--no-daemon", "--dangerously-bypass-approvals-and-sandbox", "--model", "gpt-x", "-c", `developer_instructions="say \"hi\"\nnow"`}
	if len(calls) != 1 || !slices.Equal(calls[0].args, want) {
		t.Fatalf("exec calls = %+v; want args %q", calls, want)
	}
	if !strings.Contains(logs.String(), `"<guide 40 bytes>"`) || strings.Contains(logs.String(), "say ") {
		t.Fatalf("logs = %s; want the guide placeholder and no guide text", logs.String())
	}
}
