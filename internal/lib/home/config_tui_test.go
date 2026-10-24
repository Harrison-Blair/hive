package home

import (
	"reflect"
	"strings"
	"testing"
)

const tuiHome = "home = \"/srv/hive\"\n"

// Golden-frames decisions 8, 9, 12, 16, and 18: the defaults of the TUI
// tables.
func TestParseConfigTUIDefaults(t *testing.T) {
	settings, err := parseConfig([]byte(tuiHome))
	if err != nil {
		t.Fatal(err)
	}
	wantLines := [][]string{{"state", "supervisor", "asks", "needs_you", "chords", "hints", "clock"}}
	if !reflect.DeepEqual(settings.TUI.StatusLines, wantLines) {
		t.Errorf("status lines = %q; want %q", settings.TUI.StatusLines, wantLines)
	}
	if settings.TUI.StatusLinesExplicit {
		t.Error("status lines are explicit; want the defaults")
	}
	if settings.TUI.StatusCommand != "" {
		t.Errorf("status command = %q; want empty", settings.TUI.StatusCommand)
	}
	wantKeys := map[string]string{
		"agents": "ctrl+g", "work": "ctrl+t", "friction": "ctrl+f",
		"health": "ctrl+y", "logs": "ctrl+l", "help": "?",
	}
	if !reflect.DeepEqual(settings.TUI.Keys, wantKeys) {
		t.Errorf("keys = %v; want %v", settings.TUI.Keys, wantKeys)
	}
}

func TestParseConfigTUITables(t *testing.T) {
	settings, err := parseConfig([]byte(tuiHome + `
[tui.status]
lines = [["state", "clock"], ["project", "screen"], ["command"]]
command = "statusline render hive"

[tui.keys]
agents = "ctrl+a"
help = "alt+h"
`))
	if err != nil {
		t.Fatal(err)
	}
	wantLines := [][]string{{"state", "clock"}, {"project", "screen"}, {"command"}}
	if !reflect.DeepEqual(settings.TUI.StatusLines, wantLines) {
		t.Errorf("status lines = %q; want %q", settings.TUI.StatusLines, wantLines)
	}
	if !settings.TUI.StatusLinesExplicit {
		t.Error("status lines are not explicit; want explicit")
	}
	if settings.TUI.StatusCommand != "statusline render hive" {
		t.Errorf("status command = %q", settings.TUI.StatusCommand)
	}
	if settings.TUI.Keys["agents"] != "ctrl+a" || settings.TUI.Keys["help"] != "alt+h" || settings.TUI.Keys["work"] != "ctrl+t" {
		t.Errorf("keys = %v; want agents ctrl+a, help alt+h, work ctrl+t", settings.TUI.Keys)
	}
}

// Task 155: the parser records that the config gives tui.status.lines, also
// when the rows equal the default rows of decision 9.
func TestParseConfigTUIStatusLinesExplicit(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"absent lines", "", false},
		{"command only", "[tui.status]\ncommand = \"printf task-155\"\n", false},
		{"custom row", "[tui.status]\nlines = [[\"state\", \"clock\"]]\n", true},
		{"default row", "[tui.status]\nlines = [[\"state\", \"supervisor\", \"asks\", \"needs_you\", \"chords\", \"hints\", \"clock\"]]\n", true},
		{"multiple rows", "[tui.status]\nlines = [[\"state\"], [\"clock\"]]\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings, err := parseConfig([]byte(tuiHome + tc.body))
			if err != nil {
				t.Fatal(err)
			}
			if settings.TUI.StatusLinesExplicit != tc.want {
				t.Errorf("status lines explicit = %v; want %v", settings.TUI.StatusLinesExplicit, tc.want)
			}
		})
	}
}

// Decisions 8 and 18: hive refuses a bad value with an error that names the
// key.
func TestParseConfigTUIRefusesBadValues(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       []string
	}{
		{"no rows", "[tui.status]\nlines = []\n", []string{"tui.status.lines", "1 to 3"}},
		{"four rows", "[tui.status]\nlines = [[\"state\"], [\"clock\"], [\"asks\"], [\"hints\"]]\n", []string{"tui.status.lines", "1 to 3"}},
		{"empty row", "[tui.status]\nlines = [[\"state\"], []]\n", []string{"tui.status.lines", "row 2"}},
		{"unknown segment", "[tui.status]\nlines = [[\"state\", \"weather\"]]\n", []string{"tui.status.lines", `"weather"`}},
		{"not a list", "[tui.status]\nlines = \"state\"\n", []string{"tui.status.lines"}},
		{"command not a string", "[tui.status]\ncommand = [\"a\"]\n", []string{"tui.status.command"}},
		{"unknown status key", "[tui.status]\nrows = 2\n", []string{`"rows"`}},
		{"unknown action", "[tui.keys]\nquit = \"ctrl+q\"\n", []string{"tui.keys", `"quit"`}},
		{"tab", "[tui.keys]\nlogs = \"tab\"\n", []string{"tui.keys.logs", "tab"}},
		{"two actions", "[tui.keys]\nlogs = \"ctrl+g\"\n", []string{"tui.keys.logs", "ctrl+g", "agents"}},
		{"fixed key", "[tui.keys]\nwork = \"ctrl+o\"\n", []string{"tui.keys.work", "ctrl+o"}},
		{"bad key name", "[tui.keys]\nwork = \"ctrl+\"\n", []string{"tui.keys.work", `"ctrl+"`}},
		{"unknown table", "[tui.colors]\n", []string{`"tui.colors"`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseConfig([]byte(tuiHome + tc.body))
			if err == nil {
				t.Fatal("no error")
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not name %s", err, want)
				}
			}
		})
	}
}

// The TUI reads its tables when the config file exists, and uses the
// defaults when it does not.
func TestReadTUISettings(t *testing.T) {
	path := configFixture(t, "")
	tui, err := ReadTUISettings()
	if err != nil {
		t.Fatalf("no config file: %v", err)
	}
	if tui.Keys["agents"] != "ctrl+g" || len(tui.StatusLines) != 1 || tui.StatusLinesExplicit {
		t.Fatalf("no config file gave %+v; want the defaults", tui)
	}
	writeConfig(t, path, tuiHome+"[tui.keys]\nagents = \"ctrl+a\"\n")
	if tui, err = ReadTUISettings(); err != nil || tui.Keys["agents"] != "ctrl+a" || tui.StatusLinesExplicit {
		t.Fatalf("config file gave %+v, %v; want agents ctrl+a", tui, err)
	}
	writeConfig(t, path, tuiHome+"[tui.status]\nlines = [[\"state\", \"clock\"]]\n")
	if tui, err = ReadTUISettings(); err != nil || !tui.StatusLinesExplicit {
		t.Fatalf("lines config gave %+v, %v; want explicit lines", tui, err)
	}
	writeConfig(t, path, tuiHome+"[tui.keys]\nagents = \"tab\"\n")
	if _, err = ReadTUISettings(); err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("bad config gave %v; want an error that names %s", err, path)
	}
}

// Decision 18: a rebound key is a Bubble Tea key name, also a named key
// such as up or f5. Task 137 repair F2.
func TestParseConfigTUIAcceptsNamedKeys(t *testing.T) {
	for _, key := range []string{"alt+up", "ctrl+pgdown", "f5", "shift+enter", "ctrl+space", "alt+home", "meta+x", "super+left"} {
		t.Run(key, func(t *testing.T) {
			settings, err := parseConfig([]byte(tuiHome + "[tui.keys]\nhelp = \"" + key + "\"\n"))
			if err != nil {
				t.Fatalf("%s refused: %v", key, err)
			}
			if settings.TUI.Keys["help"] != key {
				t.Fatalf("help = %q; want %q", settings.TUI.Keys["help"], key)
			}
		})
	}
	for _, key := range []string{"alt+upp", "ctrl+f99", "foo", "alt+"} {
		t.Run("refuse "+key, func(t *testing.T) {
			_, err := parseConfig([]byte(tuiHome + "[tui.keys]\nhelp = \"" + key + "\"\n"))
			if err == nil || !strings.Contains(err.Error(), "tui.keys.help") {
				t.Fatalf("%s gave %v; want an error that names tui.keys.help", key, err)
			}
		})
	}
}
