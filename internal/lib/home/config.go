package home

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CoordinatorConfig holds the launch profile of the coordinator.
type CoordinatorConfig struct {
	Harness string
	Model   string
}

// Settings holds the global hive configuration.
type Settings struct {
	Home        string
	Coordinator CoordinatorConfig
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
	settings := Settings{Coordinator: CoordinatorConfig{Harness: "claude", Model: "claude-opus-5-5"}}
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
			if table != "coordinator" {
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
		parsed, err := strconv.Unquote(strings.TrimSpace(value))
		if err != nil {
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
		default:
			return Settings{}, fmt.Errorf("line %d: unknown key %q", lineNumber+1, key)
		}
	}
	if settings.Home == "" {
		return Settings{}, fmt.Errorf("missing home key")
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
