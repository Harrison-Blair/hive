package home

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ConfigPath returns the path of the global home configuration file.
func ConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "hive", "config.toml"), nil
}

// ReadConfig reads the absolute home path from the global configuration file.
func ReadConfig() (string, error) {
	path, err := ConfigPath()
	if err != nil {
		return "", fmt.Errorf("home config path: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("home config %s: %w", path, err)
	}
	home, err := parseConfig(data)
	if err != nil {
		return "", fmt.Errorf("home config %s: %w", path, err)
	}
	return home, nil
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

// skippedTable reports a table of an old config that hive reads past with
// all its keys: [coordinator] and each [tui.*] table.
func skippedTable(table string) bool {
	return table == "coordinator" || strings.HasPrefix(table, "tui.")
}

// parseConfig returns the home path of the config. It skips the old tables
// and rejects every other table and key.
func parseConfig(data []byte) (string, error) {
	home := ""
	table := ""
	seen := map[string]bool{}
	for lineNumber, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(stripConfigComment(line))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") || strings.Count(line, "[") != 1 || strings.Count(line, "]") != 1 {
				return "", fmt.Errorf("line %d: invalid table %q", lineNumber+1, line)
			}
			table = strings.TrimSpace(line[1 : len(line)-1])
			if !skippedTable(table) {
				return "", fmt.Errorf("line %d: unknown table %q", lineNumber+1, table)
			}
			if seen[table] {
				return "", fmt.Errorf("line %d: duplicate table %q", lineNumber+1, table)
			}
			seen[table] = true
			continue
		}
		if table != "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return "", fmt.Errorf("line %d: expected key = value", lineNumber+1)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return "", fmt.Errorf("line %d: empty key", lineNumber+1)
		}
		if key != "home" {
			return "", fmt.Errorf("line %d: unknown key %q", lineNumber+1, key)
		}
		if home != "" {
			return "", fmt.Errorf("line %d: duplicate key %q", lineNumber+1, key)
		}
		parsed, err := strconv.Unquote(strings.TrimSpace(value))
		if err != nil {
			return "", fmt.Errorf("line %d: %s must be a quoted string", lineNumber+1, key)
		}
		if parsed == "" || !filepath.IsAbs(parsed) {
			return "", fmt.Errorf("home must be an absolute quoted path")
		}
		home = parsed
	}
	if home == "" {
		return "", fmt.Errorf("missing home key")
	}
	return home, nil
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
