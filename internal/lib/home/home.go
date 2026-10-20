// Package home resolves and validates a hive home.
package home

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// UsageError reports an invalid command-line value.
type UsageError struct {
	Message string
}

func (e *UsageError) Error() string { return e.Message }

// Resolve finds the hive home from the command-line value, environment value, or working directory.
func Resolve(homeFlag string, homeFlagSet bool, envHome, cwd string) (string, error) {
	workingDir := resolvePath(cwd, "")
	if homeFlagSet {
		if homeFlag == "" {
			return "", &UsageError{Message: "--home requires a path"}
		}
		return namedHome(homeFlag, workingDir)
	}
	if envHome != "" {
		return namedHome(envHome, workingDir)
	}
	configHome, err := ReadConfig()
	if err == nil {
		resolved, err := namedHome(configHome, workingDir)
		if err != nil {
			configPath, _ := ConfigPath()
			return "", fmt.Errorf("home config %s: %w", configPath, err)
		}
		if workingHome, ok := markerAbove(workingDir); ok && workingHome != resolved {
			return "", fmt.Errorf("home conflict: config is %s, working directory is in %s", resolved, workingHome)
		}
		return resolved, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if marker, ok := markerAbove(workingDir); ok {
		return marker, nil
	}
	return "", fmt.Errorf("not a hive home: no marker above %s; run hive init in the home", workingDir)
}

func namedHome(path, cwd string) (string, error) {
	resolved := resolvePath(path, cwd)
	if !hasMarker(resolved) {
		return "", fmt.Errorf("not a hive home: %s", resolved)
	}
	return resolved, nil
}

// FindMarkerAbove returns the outermost marked home above path.
func FindMarkerAbove(path string) (string, bool) {
	return markerAbove(resolvePath(path, ""))
}

// Validate checks that path names a marked home and returns its canonical path.
func Validate(path string) (string, error) {
	return namedHome(path, "")
}

func markerAbove(path string) (string, bool) {
	var outermost string
	for dir := path; ; dir = filepath.Dir(dir) {
		if hasMarker(dir) {
			outermost = dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return outermost, outermost != ""
}

func hasMarker(dir string) bool {
	hiveDir, err := os.Lstat(filepath.Join(dir, ".hive"))
	if err != nil || !hiveDir.IsDir() {
		return false
	}
	info, err := os.Lstat(filepath.Join(dir, ".hive", "README.md"))
	return err == nil && info.Mode().IsRegular()
}

func resolvePath(path, cwd string) string {
	if !filepath.IsAbs(path) {
		if cwd != "" {
			path = filepath.Join(cwd, path)
		}
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
