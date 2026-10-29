package coordinator

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
)

type execCall struct {
	dir, binary string
	args        []string
}

func launchOptions(t *testing.T, home string, pick func(Choice, bool) (Choice, bool, error), calls *[]execCall) Options {
	t.Helper()
	return Options{
		Home: home,
		Pane: "w1-3",
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
			var calls []execCall
			opts := launchOptions(t, home, func(Choice, bool) (Choice, bool, error) { return tc.choice, true, nil }, &calls)
			if err := Run(opts); err != nil {
				t.Fatal(err)
			}
			record, err := agent.Load(home, "coordinator")
			if err != nil {
				t.Fatal(err)
			}
			want := agent.Record{Name: "coordinator", Pane: "w1-3", Harness: tc.choice.Harness, Model: tc.choice.Model, Cwd: home, StartedAt: time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)}
			if record != want {
				t.Fatalf("record = %+v; want %+v", record, want)
			}
			if len(calls) != 1 || calls[0].dir != home || calls[0].binary != tc.wantBinary || !slices.Equal(calls[0].args, tc.wantArgs) {
				t.Fatalf("exec calls = %+v; want one %s %v in %s", calls, tc.wantBinary, tc.wantArgs, home)
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
	opts := launchOptions(t, home, func(c Choice, ok bool) (Choice, bool, error) {
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
	opts := launchOptions(t, home, func(Choice, bool) (Choice, bool, error) { return Choice{}, false, nil }, &calls)
	if err := Run(opts); err != nil {
		t.Fatalf("Run = %v; want nil on quit", err)
	}
	if len(calls) != 0 {
		t.Fatalf("exec calls = %+v; want none", calls)
	}
	if _, err := agent.Load(home, "coordinator"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("Load = %v; want ErrNotFound", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "coordinator-choice.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("choice file stat = %v; want not exist", err)
	}
}

func TestRunReturnsPickerAndExecErrors(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, func(Choice, bool) (Choice, bool, error) { return Choice{}, false, errors.New("tty gone") }, &calls)
	if err := Run(opts); err == nil || err.Error() != "picker: tty gone" {
		t.Fatalf("Run = %v; want picker: tty gone", err)
	}
	opts = launchOptions(t, home, func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
	opts.Exec = func(string, string, []string) error { return errors.New("no such file") }
	if err := Run(opts); err == nil {
		t.Fatal("Run = nil; want the exec error")
	}
}

func TestRunRejectsUnknownHarness(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, func(Choice, bool) (Choice, bool, error) { return Choice{"vim", "x"}, true, nil }, &calls)
	if err := Run(opts); err == nil || len(calls) != 0 {
		t.Fatalf("Run = %v, calls %v; want error and no exec", err, calls)
	}
	if _, err := agent.Load(home, "coordinator"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("Load = %v; want no record", err)
	}
}
