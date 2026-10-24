package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/home"
)

// statusLineHome is the home line of a config file for these tests. The TUI
// reads the [tui.status] table; the home line only makes the file valid.
const statusLineHome = "home = \"/srv/hive\"\n"

// statusLineScreens gives each screen other than Chat and its screen segment
// name.
var statusLineScreens = []struct {
	tab  int
	name string
}{
	{tabAgents, "agents"},
	{tabWork, "work"},
	{tabFriction, "friction"},
	{tabHealth, "health"},
	{tabLogs, "logs"},
}

// statusLineConfig writes one config file to an isolated config home and
// gives the [tui.status] settings that the TUI reads from it. An empty body
// writes no file.
func statusLineConfig(t *testing.T, body string) home.TUIConfig {
	t.Helper()
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	if body != "" {
		path := filepath.Join(configDir, "hive", "config.toml")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(statusLineHome+body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := readTUISettings()
	if err != nil {
		t.Fatalf("read TUI settings: %v", err)
	}
	return cfg
}

func requireStatusParts(t *testing.T, row string, parts []string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(row, part) {
			t.Errorf("status row %q does not contain %q", row, part)
		}
	}
}

func rejectStatusParts(t *testing.T, row string, parts []string) {
	t.Helper()
	for _, part := range parts {
		if strings.Contains(row, part) {
			t.Errorf("status row %q contains %q", row, part)
		}
	}
}

// Golden-frames decision 34 and task 155: a config with [tui.status] lines
// keeps its rows on each screen other than Chat, also when a row equals the
// default row of decision 9. Without explicit lines, each screen keeps the
// a27 default form.
func TestFrameStatusLinesFromConfigKeepTheirRows(t *testing.T) {
	const defaultRow = `["state", "supervisor", "asks", "needs_you", "chords", "hints", "clock"]`
	implicitAbsent := []string{"supervisor", "esc chat", "? help", "? 0 asks", "needs you 0"}
	cases := []struct {
		name     string
		body     string
		implicit bool
		first    []string
		second   []string
	}{
		{
			name:     "absent config",
			implicit: true,
		},
		{
			name:     "absent lines",
			implicit: true,
		},
		{
			name:     "command only",
			body:     "[tui.status]\ncommand = \"printf task-155\"\n",
			implicit: true,
		},
		{
			name:  "explicit default row",
			body:  "[tui.status]\nlines = [" + defaultRow + "]\n",
			first: []string{"supervisor", "? 0 asks", "needs you 0", "esc chat", "? help"},
		},
		{
			name:   "explicit default first of two",
			body:   "[tui.status]\nlines = [" + defaultRow + `, ["state", "asks", "clock"]]` + "\n",
			first:  []string{"supervisor", "? 0 asks", "needs you 0", "esc chat", "? help"},
			second: []string{"? 0 asks"},
		},
		{
			name:  "explicit custom row",
			body:  "[tui.status]\nlines = [[\"state\", \"chords\", \"clock\"]]\n",
			first: []string{"esc chat"},
		},
		{
			name:   "explicit custom multiple rows",
			body:   "[tui.status]\nlines = [[\"state\", \"chords\", \"clock\"], [\"state\", \"asks\", \"needs_you\", \"clock\"]]\n",
			first:  []string{"esc chat"},
			second: []string{"? 0 asks", "needs you 0"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := statusLineConfig(t, tc.body)
			for _, screen := range statusLineScreens {
				t.Run(screen.name, func(t *testing.T) {
					m, _ := newFrameModel(t, 160, 40)
					m.setTUIConfig(cfg)
					m.selectScreen(screen.tab, 0)
					rows := viewRows(m)
					if tc.implicit {
						row := rows[39]
						requireStatusParts(t, row, []string{" · " + screen.name + " · "})
						rejectStatusParts(t, row, implicitAbsent)
						return
					}
					screenPart := " · " + screen.name + " · "
					if len(tc.second) == 0 {
						row := rows[39]
						requireStatusParts(t, row, tc.first)
						rejectStatusParts(t, row, []string{screenPart})
						return
					}
					first, second := rows[38], rows[39]
					requireStatusParts(t, first, tc.first)
					requireStatusParts(t, second, tc.second)
					rejectStatusParts(t, first, []string{screenPart})
					rejectStatusParts(t, second, []string{screenPart, "supervisor", "esc chat", "? help"})
				})
			}
		})
	}
}
