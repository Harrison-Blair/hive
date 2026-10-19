package agent

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// projectAgentsTemplate is the manager instruction file that hive writes for a
// project. T21 refines its text.
//
//go:embed templates/project-agents.md
var projectAgentsTemplate string

// presetsTemplate is the default presets document that hive writes at the
// first manager spawn of a home.
//
//go:embed templates/presets.md
var presetsTemplate string

// writeProjectInstructions writes projects/<id>/AGENTS.md from the template.
// It returns an error message, or the empty string on success.
func writeProjectInstructions(home string, project records.Project, logger *slog.Logger) string {
	directory := filepath.Join(home, "projects", project.ID)
	if message := checkStaticDirectory(home, directory); message != "" {
		return message
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "create project directory: " + err.Error()
	}
	dest := filepath.Join(directory, "AGENTS.md")
	if message := checkRegularDestination(dest); message != "" {
		return message
	}
	text := strings.NewReplacer(
		"{{PROJECT_ID}}", project.ID,
		"{{HOME}}", home,
		"{{REPOSITORY}}", project.Path,
	).Replace(projectAgentsTemplate)
	if err := os.WriteFile(dest, []byte(text), 0o644); err != nil {
		return "write project instructions: " + err.Error()
	}
	if logger != nil {
		logger.Debug("manager instructions written", "command", "agent spawn", "project_id", project.ID, "path", filepath.ToSlash(filepath.Join("projects", project.ID, "AGENTS.md")))
	}
	return ""
}

// writePresets writes overrides/presets.md when the file is absent. It returns
// an error message, or the empty string on success. The file check and the
// write run under the store lock, so two first spawns leave one file.
func writePresets(home string, logger *slog.Logger) string {
	directory := filepath.Join(home, "overrides")
	dest := filepath.Join(directory, "presets.md")
	if message := checkStaticOverrides(home, directory, dest); message != "" {
		return message
	}
	registry, err := store.Open(home)
	if err != nil {
		return "open store: " + err.Error()
	}
	writeMessage := ""
	lockErr := registry.WithLock(context.Background(), func(locked *store.LockedStore) error {
		writeMessage = createPresetsFile(directory, dest)
		return nil
	})
	if lockErr != nil {
		return "presets lock: " + lockErr.Error()
	}
	if writeMessage != "" {
		return writeMessage
	}
	if logger != nil {
		logger.Debug("presets written", "command", "agent spawn", "path", filepath.ToSlash(filepath.Join("overrides", "presets.md")))
	}
	return ""
}

// createPresetsFile writes the presets document when the destination is
// absent. It keeps each byte of an existing file. A link to the destination
// makes the write atomic and never replaces the first writer.
func createPresetsFile(directory, dest string) string {
	info, err := os.Lstat(dest)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return "presets file is a symbolic link: " + dest
		}
		if !info.Mode().IsRegular() {
			return "presets file is not a regular file: " + dest
		}
		return ""
	case !os.IsNotExist(err):
		return "presets file: " + err.Error()
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "create overrides directory: " + err.Error()
	}
	temp, err := os.CreateTemp(directory, ".presets.tmp-")
	if err != nil {
		return "presets temporary file: " + err.Error()
	}
	tempName := temp.Name()
	defer func() { _ = os.Remove(tempName) }()
	if _, err := temp.WriteString(presetsTemplate); err != nil {
		_ = temp.Close()
		return "write presets: " + err.Error()
	}
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return "write presets: " + err.Error()
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return "write presets: " + err.Error()
	}
	if err := temp.Close(); err != nil {
		return "write presets: " + err.Error()
	}
	if err := os.Link(tempName, dest); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ""
		}
		return "write presets: " + err.Error()
	}
	return ""
}

// checkStaticOverrides rejects a symbolic link or a wrong type at the home,
// the overrides directory, each ancestor, and the presets destination.
func checkStaticOverrides(home, directory, dest string) string {
	if message := checkStaticComponent(home); message != "" {
		return "presets path: " + message
	}
	relative, err := filepath.Rel(home, directory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "presets path is outside the home"
	}
	current := home
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return ""
			}
			return "presets path: " + err.Error()
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "presets directory is a symbolic link: " + current
		}
		if !info.IsDir() {
			return "presets directory is not a directory: " + current
		}
	}
	info, err := os.Lstat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		return "presets file: " + err.Error()
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "presets file is a symbolic link: " + dest
	}
	if !info.Mode().IsRegular() {
		return "presets file is not a regular file: " + dest
	}
	return ""
}

// checkStaticDirectory rejects a symbolic link or a non-directory at the home,
// the project directory, and each ancestor between them.
func checkStaticDirectory(home, directory string) string {
	relative, err := filepath.Rel(home, directory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "instruction path is outside the home"
	}
	if message := checkStaticComponent(home); message != "" {
		return message
	}
	current := home
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return ""
			}
			return "instruction path: " + err.Error()
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "instruction path is a symbolic link: " + current
		}
		if !info.IsDir() {
			return "instruction path is not a directory: " + current
		}
	}
	return ""
}

func checkStaticComponent(path string) string {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		return "instruction path: " + err.Error()
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "instruction path is a symbolic link: " + path
	}
	if !info.IsDir() {
		return "instruction path is not a directory: " + path
	}
	return ""
}

// checkRegularDestination rejects a symbolic link or a non-regular file at the
// instruction destination.
func checkRegularDestination(path string) string {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		return "instruction file: " + err.Error()
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "instruction file is a symbolic link: " + path
	}
	if !info.Mode().IsRegular() {
		return "instruction file is not a regular file: " + path
	}
	return ""
}
