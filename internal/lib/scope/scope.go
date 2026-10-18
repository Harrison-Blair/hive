package scope

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

const (
	projectPrefix = "project:"
)

// Scope is one resolved memory or handoff location.
type Scope struct {
	Value string
	Dir   string
}

// DefaultDeadline is the stdin read deadline for the memory and handoff commands.
const DefaultDeadline = 30 * time.Second

// Resolve parses one scope value and checks its directory.
func Resolve(home, value string) (Scope, error) {
	if value == "home" {
		if err := CheckDir(home, home); err != nil {
			return Scope{}, err
		}
		return Scope{Value: value, Dir: home}, nil
	}
	if !strings.HasPrefix(value, projectPrefix) {
		return Scope{}, fmt.Errorf("bad scope: %s", display(value))
	}
	id := strings.TrimPrefix(value, projectPrefix)
	if !records.ValidID(id) {
		return Scope{}, fmt.Errorf("bad scope: %s", display(value))
	}
	registry, err := store.Open(home)
	if err != nil {
		return Scope{}, err
	}
	var project records.Project
	if err := registry.Read("projects", id, &project); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Scope{}, fmt.Errorf("project not found: %s", id)
		}
		return Scope{}, err
	}
	dir := filepath.Join(home, "projects", id)
	if err := CheckDir(home, dir); err != nil {
		return Scope{}, err
	}
	return Scope{Value: value, Dir: dir}, nil
}

// CheckDir checks one directory path in the home.
func CheckDir(home, dir string) error {
	return checkPath(home, dir, true)
}

// CheckFile checks one file path in the home.
func CheckFile(home, path string) error {
	return checkPath(home, path, false)
}

func checkPath(home, target string, wantDir bool) error {
	homePath, err := filepath.Abs(home)
	if err != nil {
		return err
	}
	targetPath, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(homePath, targetPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path outside home: %s", display(target))
	}
	info, err := os.Lstat(homePath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("path invalid: %s", display(homePath))
	}
	if relative == "." {
		if wantDir {
			return nil
		}
		return fmt.Errorf("path invalid: %s", display(targetPath))
	}
	current := homePath
	parts := strings.Split(relative, string(filepath.Separator))
	for index, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		last := index == len(parts)-1
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path invalid: %s", display(current))
		}
		if last {
			if wantDir && !info.IsDir() {
				return fmt.Errorf("path invalid: %s", display(current))
			}
			if !wantDir && !info.Mode().IsRegular() {
				return fmt.Errorf("path invalid: %s", display(current))
			}
			continue
		}
		if !info.IsDir() {
			return fmt.Errorf("path invalid: %s", display(current))
		}
	}
	return nil
}

func display(value string) string {
	if !utf8.ValidString(value) {
		return fmt.Sprintf("%q", value)
	}
	for _, char := range value {
		if unicode.IsControl(char) {
			return fmt.Sprintf("%q", value)
		}
	}
	return value
}
