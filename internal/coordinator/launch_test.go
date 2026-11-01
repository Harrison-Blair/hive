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
	args, env   []string
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
		Name: Name,
		Pane: pane.ID,
		Pick: pick,
		Exec: func(dir, binary string, args, env []string) error {
			*calls = append(*calls, execCall{dir, binary, args, env})
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
		{Choice{"pi", "openai-codex/gpt-6-sol"}, "pi", []string{"--model", "openai-codex/gpt-6-sol", "-e", "<home>/extensions/hive-inbox.ts"}},
		{Choice{"claude", "claude-opus-5-5"}, "claude", []string{"--dangerously-skip-permissions", "--settings", `{"crossSessionInbound":"accept"}`, "--model", "claude-opus-5-5"}},
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
			wantArgs := slices.Clone(tc.wantArgs)
			for i, arg := range wantArgs {
				wantArgs[i] = strings.ReplaceAll(arg, "<home>", home)
			}
			if len(calls) != 1 || calls[0].dir != dir || calls[0].binary != tc.wantBinary || !slices.Equal(calls[0].args, wantArgs) {
				t.Fatalf("exec calls = %+v; want one %s %v in %s", calls, tc.wantBinary, wantArgs, dir)
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
	opts.Exec = func(string, string, []string, []string) error { return errors.New("no such file") }
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
	opts.Exec = func(dir, binary string, args, env []string) error {
		got, _ := fake.Pane(pane.ID)
		if got.TabLabel != "coordinator" {
			t.Errorf("tab label at exec = %q, want coordinator", got.TabLabel)
		}
		calls = append(calls, execCall{dir, binary, args, env})
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

func TestRunGivesTheHarnessTheHome(t *testing.T) {
	t.Setenv("HIVE_HOME", "relative-home")
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("exec calls = %+v; want one", calls)
	}
	var homes []string
	for _, pair := range calls[0].env {
		if value, ok := strings.CutPrefix(pair, "HIVE_HOME="); ok {
			homes = append(homes, value)
		}
	}
	if !slices.Equal(homes, []string{home}) {
		t.Fatalf("HIVE_HOME values in the harness environment = %q; want only %q", homes, home)
	}
	if !slices.ContainsFunc(calls[0].env, func(pair string) bool { return strings.HasPrefix(pair, "PATH=") }) {
		t.Fatalf("harness environment = %q; want the process environment kept", calls[0].env)
	}
}

// TestRunUsesItsNameEverywhere proves that a project coordinator saves its
// record, labels its tab, logs as itself, and is the caller of its pane under
// one name.
func TestRunUsesItsNameEverywhere(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	var logs strings.Builder
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
	fake := agenttest.NewHerdr(t)
	_, pane := fake.AddWorkspace("w", home, false)
	opts.Pane = pane.ID
	opts.Name = "hive-coordinator"
	opts.Logger = slog.New(slog.NewJSONHandler(&logs, nil))
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	if record, err := agent.Load(home, "hive-coordinator"); err != nil || record.Pane != pane.ID {
		t.Fatalf("record = %+v, %v; want hive-coordinator in pane %s", record, err, pane.ID)
	}
	if _, err := agent.Load(home, "coordinator"); !errors.Is(err, agent.ErrNotFound) {
		t.Fatalf("Load(coordinator) = %v; want ErrNotFound", err)
	}
	if got, _ := fake.Pane(pane.ID); got.TabLabel != "hive-coordinator" {
		t.Fatalf("tab label = %q; want hive-coordinator", got.TabLabel)
	}
	if caller, err := agent.CallerName(home, pane.ID); err != nil || caller != "hive-coordinator" {
		t.Fatalf("CallerName = %q, %v; want hive-coordinator", caller, err)
	}
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		if !strings.Contains(line, `"agent_id":"hive-coordinator"`) {
			t.Fatalf("log event %s; want agent_id hive-coordinator", line)
		}
	}
}

// TestRunReplacesTheOldCoordinatorRecordOfItsPane proves that a coordinator
// record of an earlier run in the same pane does not keep the pane, while the
// record of a worktree agent stays for agent stop.
func TestRunReplacesTheOldCoordinatorRecordOfItsPane(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
	opts.Name = "hive-coordinator"
	for _, record := range []agent.Record{
		{Name: "coordinator", Pane: opts.Pane, Harness: "pi", Model: "old"},
		{Name: "web-coordinator", Pane: opts.Pane, Harness: "pi", Model: "old"},
		{Name: "a-coordinator", Pane: opts.Pane, Harness: "pi", Model: "old", Worktree: "/wt/a"},
	} {
		if err := agent.Save(home, record); err != nil {
			t.Fatal(err)
		}
	}
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"coordinator", "web-coordinator"} {
		if _, err := agent.Load(home, name); !errors.Is(err, agent.ErrNotFound) {
			t.Errorf("Load(%s) = %v; want the old record of the pane removed", name, err)
		}
	}
	for _, name := range []string{"a-coordinator"} {
		if _, err := agent.Load(home, name); err != nil {
			t.Errorf("Load(%s) = %v; want the record kept", name, err)
		}
	}
}

// TestRunKeepsAnOldCoordinatorOfAnotherPane proves that the coordinator
// record of another pane stays, so agent stop can still close that pane.
func TestRunKeepsAnOldCoordinatorOfAnotherPane(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "a/b"}, true, nil }, &calls)
	opts.Name = "hive-coordinator"
	old := agent.Record{Name: "coordinator", Pane: "w9:p9", Harness: "pi", Model: "old"}
	if err := agent.Save(home, old); err != nil {
		t.Fatal(err)
	}
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	if got, err := agent.Load(home, "coordinator"); err != nil || got.Pane != old.Pane {
		t.Fatalf("Load(coordinator) = %+v, %v; want the record of pane w9:p9 kept", got, err)
	}
}

// TestRunRefusesTheNameOfALiveAgent proves that bare hive does not take the
// record of an agent whose pane herdr still has, and replaces the record of a
// pane that is gone.
func TestRunRefusesTheNameOfALiveAgent(t *testing.T) {
	for _, name := range []string{"coordinator", "hive-coordinator"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			var calls []execCall
			picked := false
			opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { picked = true; return Choice{"pi", "a/b"}, true, nil }, &calls)
			fake := agenttest.NewHerdr(t)
			ws, mine := fake.AddWorkspace("w", home, false)
			other := fake.AddPane(ws.ID, home)
			opts.Pane = mine.ID
			opts.Name = name
			live := agent.Record{Name: name, Pane: other.ID, Harness: "claude", Model: "m", Worktree: "/wt/x"}
			if err := agent.Save(home, live); err != nil {
				t.Fatal(err)
			}
			err := Run(opts)
			if err == nil || !strings.Contains(err.Error(), other.ID) {
				t.Fatalf("Run = %v; want an error that names pane %s", err, other.ID)
			}
			if picked || len(calls) != 0 {
				t.Fatalf("picked %v, exec calls %+v; want no picker and no exec", picked, calls)
			}
			if got, _ := agent.Load(home, name); got != live {
				t.Fatalf("record = %+v; want %+v unchanged", got, live)
			}
			if got, _ := fake.Pane(mine.ID); got.TabLabel != "" {
				t.Fatalf("tab label = %q; want none", got.TabLabel)
			}

			fake.ClosePane(other.ID)
			if err := Run(opts); err != nil {
				t.Fatalf("Run after the pane closed = %v; want nil", err)
			}
			if got, err := agent.Load(home, name); err != nil || got.Pane != mine.ID {
				t.Fatalf("record = %+v, %v; want pane %s", got, err, mine.ID)
			}
		})
	}
}

// TestRunReplacesItsOwnRecordOnARerun proves that bare hive starts again in
// the pane that holds the record of its name.
func TestRunReplacesItsOwnRecordOnARerun(t *testing.T) {
	home := t.TempDir()
	var calls []execCall
	opts := launchOptions(t, home, t.TempDir(), func(Choice, bool) (Choice, bool, error) { return Choice{"pi", "new/model"}, true, nil }, &calls)
	opts.Name = "hive-coordinator"
	if err := agent.Save(home, agent.Record{Name: opts.Name, Pane: opts.Pane, Harness: "pi", Model: "old/model"}); err != nil {
		t.Fatal(err)
	}
	if err := Run(opts); err != nil {
		t.Fatalf("Run = %v; want nil", err)
	}
	if got, err := agent.Load(home, opts.Name); err != nil || got.Model != "new/model" || len(calls) != 1 {
		t.Fatalf("record = %+v, %v, exec calls %d; want model new/model and one exec", got, err, len(calls))
	}
}
