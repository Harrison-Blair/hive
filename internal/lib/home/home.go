// Package home resolves the hive home.
package home

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve returns the hive home: envHome when it is not empty, otherwise
// ~/.hive. A relative envHome is relative to cwd. Resolve does not create the
// home.
func Resolve(envHome, cwd string) (string, error) {
	if envHome != "" {
		return resolvePath(envHome, cwd), nil
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("hive home: %w", err)
	}
	return resolvePath(filepath.Join(userHome, ".hive"), ""), nil
}

func resolvePath(path, cwd string) string {
	if !filepath.IsAbs(path) && cwd != "" {
		path = filepath.Join(cwd, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(absolute); err == nil {
		return resolved
	}
	return filepath.Clean(absolute)
}
