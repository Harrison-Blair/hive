package home

import (
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

// WriteConfig writes the absolute home path to the global configuration file.
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
	if err := os.WriteFile(path, []byte("home = "+strconv.Quote(home)+"\n"), 0o600); err != nil {
		return fmt.Errorf("home config %s: %w", path, err)
	}
	return nil
}

func parseConfig(data []byte) (string, error) {
	var home string
	found := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "home" || found {
			return "", fmt.Errorf("expected one home key")
		}
		parsed, err := strconv.Unquote(strings.TrimSpace(value))
		if err != nil || parsed == "" || !filepath.IsAbs(parsed) {
			return "", fmt.Errorf("home must be an absolute quoted path")
		}
		home = parsed
		found = true
	}
	if !found {
		return "", fmt.Errorf("missing home key")
	}
	return home, nil
}
