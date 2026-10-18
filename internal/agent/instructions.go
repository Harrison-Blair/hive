package agent

import (
	_ "embed"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// projectAgentsTemplate is the manager instruction file that hive writes for a
// project. T21 refines its text.
//
//go:embed templates/project-agents.md
var projectAgentsTemplate string

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
