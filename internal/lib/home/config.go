package home

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// CoordinatorConfig holds the launch profile of the coordinator.
type CoordinatorConfig struct {
	Harness string
	Model   string
}

// TUIConfig holds the [tui.status] and [tui.keys] tables of golden-frames
// decisions 8 to 13 and 18.
type TUIConfig struct {
	// StatusLines gives one to three status rows of segment names.
	StatusLines [][]string
	// StatusCommand is the command of the command segment, or empty.
	StatusCommand string
	// Keys maps each action (agents, work, friction, health, logs, help) to
	// its Bubble Tea key name.
	Keys map[string]string
}

// Settings holds the global hive configuration.
type Settings struct {
	Home        string
	Coordinator CoordinatorConfig
	TUI         TUIConfig
}

// ReadTUISettings reads the TUI tables of the global configuration. It gives
// the defaults when the config file does not exist.
func ReadTUISettings() (TUIConfig, error) {
	path, err := ConfigPath()
	if err != nil {
		return TUIConfig{}, fmt.Errorf("home config path: %w", err)
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return DefaultTUI(), nil
	}
	settings, err := ReadSettings()
	if err != nil {
		return TUIConfig{}, err
	}
	return settings.TUI, nil
}

// statusSegments are the segment names of golden-frames decision 9.
var statusSegments = []string{"state", "supervisor", "asks", "needs_you", "project", "screen", "chords", "hints", "clock", "command"}

// keyActions are the actions that [tui.keys] can rebind, in Help order.
var keyActions = []string{"agents", "work", "friction", "health", "logs", "help"}

// fixedKeys are the keys of golden-frames decision 16 that [tui.keys]
// cannot take.
var fixedKeys = map[string]string{
	"ctrl+o": "events", "ctrl+r": "ask focus", "ctrl+]": "box focus", "ctrl+c": "quit",
	"esc": "back", "enter": "send", "tab": "autocomplete",
}

// DefaultTUI gives the TUI settings of golden-frames decisions 9 and 16.
func DefaultTUI() TUIConfig {
	return TUIConfig{
		StatusLines: [][]string{{"state", "supervisor", "asks", "needs_you", "chords", "hints", "clock"}},
		Keys: map[string]string{
			"agents": "ctrl+g", "work": "ctrl+t", "friction": "ctrl+f",
			"health": "ctrl+y", "logs": "ctrl+l", "help": "?",
		},
	}
}

// parseStatusLines reads a one-line list of lists of quoted segment names.
func parseStatusLines(value string) ([][]string, error) {
	rest := strings.TrimSpace(value)
	if !strings.HasPrefix(rest, "[") || !strings.HasSuffix(rest, "]") {
		return nil, fmt.Errorf("tui.status.lines must be a list of lists of segment names")
	}
	rest = strings.TrimSpace(rest[1 : len(rest)-1])
	var lines [][]string
	for rest != "" {
		if !strings.HasPrefix(rest, "[") {
			return nil, fmt.Errorf("tui.status.lines must be a list of lists of segment names")
		}
		end := strings.Index(rest, "]")
		if end < 0 {
			return nil, fmt.Errorf("tui.status.lines must be a list of lists of segment names")
		}
		row := []string{}
		for item := range strings.SplitSeq(rest[1:end], ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			name, err := strconv.Unquote(item)
			if err != nil {
				return nil, fmt.Errorf("tui.status.lines: %s is not a quoted segment name", item)
			}
			if !slices.Contains(statusSegments, name) {
				return nil, fmt.Errorf("tui.status.lines: unknown segment %q", name)
			}
			row = append(row, name)
		}
		if len(row) == 0 {
			return nil, fmt.Errorf("tui.status.lines: row %d has no segment", len(lines)+1)
		}
		lines = append(lines, row)
		rest = strings.TrimSpace(rest[end+1:])
		rest = strings.TrimSpace(strings.TrimPrefix(rest, ","))
	}
	if len(lines) < 1 || len(lines) > 3 {
		return nil, fmt.Errorf("tui.status.lines must have 1 to 3 rows, not %d", len(lines))
	}
	return lines, nil
}

// validKeyName accepts a Bubble Tea key name: one character or a named key
// such as up or f5, with optional ctrl+, alt+, shift+, meta+, hyper+, or
// super+ prefixes.
func validKeyName(name string) bool {
	base := name
	for {
		trimmed := base
		for _, mod := range []string{"ctrl+", "alt+", "shift+", "meta+", "hyper+", "super+"} {
			trimmed = strings.TrimPrefix(trimmed, mod)
		}
		if trimmed == base {
			break
		}
		base = trimmed
	}
	if utf8.RuneCountInString(base) == 1 {
		return strings.TrimSpace(base) != ""
	}
	return slices.Contains(namedKeys(), base)
}

// namedKeys gives each Bubble Tea name of a key with no one-character name,
// for example space, enter, up, and f5.
func namedKeys() []string {
	var names []string
	add := func(code rune) {
		if name := (tea.Key{Code: code}).String(); utf8.RuneCountInString(name) > 1 {
			names = append(names, name)
		}
	}
	for code := rune(0); code <= ' '; code++ {
		add(code)
	}
	add(tea.KeyBackspace)
	for code := tea.KeyExtended + 1; code <= tea.KeyExtended+512; code++ {
		add(code)
	}
	return names
}

// checkKeys refuses tab, a fixed key, and one key for two actions.
func checkKeys(keys map[string]string) error {
	for _, action := range keyActions {
		key := keys[action]
		if key == "tab" {
			return fmt.Errorf("tui.keys.%s: tab is kept for autocomplete", action)
		}
		if fixed, ok := fixedKeys[key]; ok {
			return fmt.Errorf("tui.keys.%s: %s is the %s key", action, key, fixed)
		}
		if !validKeyName(key) {
			return fmt.Errorf("tui.keys.%s: %q is not a key name", action, key)
		}
		for _, other := range keyActions[:slices.Index(keyActions, action)] {
			if keys[other] == key {
				return fmt.Errorf("tui.keys.%s: %s is also the key of %s", action, key, other)
			}
		}
	}
	return nil
}

// ConfigPath returns the path of the global home configuration file.
func ConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "hive", "config.toml"), nil
}

// ReadSettings reads the global hive configuration.
func ReadSettings() (Settings, error) {
	path, err := ConfigPath()
	if err != nil {
		return Settings{}, fmt.Errorf("home config path: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, fmt.Errorf("home config %s: %w", path, err)
	}
	settings, err := parseConfig(data)
	if err != nil {
		return Settings{}, fmt.Errorf("home config %s: %w", path, err)
	}
	return settings, nil
}

// ReadConfig reads the absolute home path from the global configuration file.
func ReadConfig() (string, error) {
	settings, err := ReadSettings()
	if err != nil {
		return "", err
	}
	return settings.Home, nil
}

// WriteConfig writes the absolute home path and keeps every other config line.
func WriteConfig(home string) error {
	path, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("home config path: %w", err)
	}
	if !filepath.IsAbs(home) {
		return fmt.Errorf("home config %s: home path must be absolute", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("home config %s: %w", path, err)
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("home config %s: %w", path, err)
	}
	updated := replaceHomeLine(data, home)
	if err := os.WriteFile(path, updated, 0o600); err != nil {
		return fmt.Errorf("home config %s: %w", path, err)
	}
	return nil
}

func parseConfig(data []byte) (Settings, error) {
	settings := Settings{Coordinator: CoordinatorConfig{Harness: "claude", Model: "claude-opus-5-5"}, TUI: DefaultTUI()}
	table := ""
	seen := map[string]bool{}
	for lineNumber, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(stripConfigComment(line))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") || strings.Count(line, "[") != 1 || strings.Count(line, "]") != 1 {
				return Settings{}, fmt.Errorf("line %d: invalid table %q", lineNumber+1, line)
			}
			table = strings.TrimSpace(line[1 : len(line)-1])
			if table != "coordinator" && table != "tui.status" && table != "tui.keys" {
				return Settings{}, fmt.Errorf("line %d: unknown table %q", lineNumber+1, table)
			}
			if seen["table:"+table] {
				return Settings{}, fmt.Errorf("line %d: duplicate table %q", lineNumber+1, table)
			}
			seen["table:"+table] = true
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Settings{}, fmt.Errorf("line %d: expected key = value", lineNumber+1)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return Settings{}, fmt.Errorf("line %d: empty key", lineNumber+1)
		}
		seenKey := table + ":" + key
		if seen[seenKey] {
			return Settings{}, fmt.Errorf("line %d: duplicate key %q", lineNumber+1, key)
		}
		seen[seenKey] = true
		if table == "tui.status" && key != "lines" && key != "command" {
			return Settings{}, fmt.Errorf("line %d: unknown key %q", lineNumber+1, key)
		}
		if table == "tui.status" && key == "lines" {
			lines, err := parseStatusLines(value)
			if err != nil {
				return Settings{}, fmt.Errorf("line %d: %w", lineNumber+1, err)
			}
			settings.TUI.StatusLines = lines
			continue
		}
		parsed, err := strconv.Unquote(strings.TrimSpace(value))
		if err != nil {
			if strings.HasPrefix(table, "tui.") {
				key = table + "." + key
			}
			return Settings{}, fmt.Errorf("line %d: %s must be a quoted string", lineNumber+1, key)
		}
		switch {
		case table == "" && key == "home":
			if parsed == "" || !filepath.IsAbs(parsed) {
				return Settings{}, fmt.Errorf("home must be an absolute quoted path")
			}
			settings.Home = parsed
		case table == "coordinator" && key == "harness":
			if parsed == "" {
				return Settings{}, fmt.Errorf("coordinator.harness must not be empty")
			}
			settings.Coordinator.Harness = parsed
		case table == "coordinator" && key == "model":
			if parsed == "" {
				return Settings{}, fmt.Errorf("coordinator.model must not be empty")
			}
			settings.Coordinator.Model = parsed
		case table == "tui.status" && key == "command":
			settings.TUI.StatusCommand = parsed
		case table == "tui.keys" && slices.Contains(keyActions, key):
			settings.TUI.Keys[key] = parsed
		case table == "tui.keys":
			return Settings{}, fmt.Errorf("line %d: tui.keys: unknown action %q", lineNumber+1, key)
		default:
			return Settings{}, fmt.Errorf("line %d: unknown key %q", lineNumber+1, key)
		}
	}
	if settings.Home == "" {
		return Settings{}, fmt.Errorf("missing home key")
	}
	if err := checkKeys(settings.TUI.Keys); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

func stripConfigComment(line string) string {
	quoted := false
	escaped := false
	for index, value := range line {
		if escaped {
			escaped = false
			continue
		}
		if value == '\\' && quoted {
			escaped = true
			continue
		}
		if value == '"' {
			quoted = !quoted
			continue
		}
		if value == '#' && !quoted {
			return line[:index]
		}
	}
	return line
}

func replaceHomeLine(data []byte, home string) []byte {
	lines := strings.SplitAfter(string(data), "\n")
	table := ""
	found := false
	for index, line := range lines {
		content := strings.TrimSuffix(line, "\n")
		content = strings.TrimSuffix(content, "\r")
		trimmed := strings.TrimSpace(stripConfigComment(content))
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			table = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			continue
		}
		if table != "" {
			continue
		}
		key, _, ok := strings.Cut(trimmed, "=")
		if !ok || strings.TrimSpace(key) != "home" {
			continue
		}
		ending := "\n"
		if strings.HasSuffix(line, "\r\n") {
			ending = "\r\n"
		} else if !strings.HasSuffix(line, "\n") {
			ending = ""
		}
		lines[index] = "home = " + strconv.Quote(home) + ending
		found = true
		break
	}
	if !found {
		joined := strings.Join(lines, "")
		if joined != "" && !strings.HasSuffix(joined, "\n") {
			joined += "\n"
		}
		return []byte(joined + "home = " + strconv.Quote(home) + "\n")
	}
	return []byte(strings.Join(lines, ""))
}
