package land

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/worktree"
)

// missingHash stands for a file that does not exist in a conflict report.
const missingHash = "none"

// copySpecOut copies each spec file of the worktree outside tasks/ that
// changed since copy-in, or that is new, to the home. It does not write a
// home file that differs from its copy-in hash. It returns one conflict
// message for each such file. A file that the task deleted stays in the home.
func copySpecOut(home string, task records.Task, logger *slog.Logger) ([]output.Effect, []string, error) {
	if task.Feature == "" || task.Worktree == nil || task.Worktree.Path == "" {
		return nil, nil, nil
	}
	relDir := ".hive/spec/" + task.Feature + "/"
	source := filepath.Join(task.Worktree.Path, filepath.FromSlash(relDir))
	if _, err := os.Stat(source); errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil
	}
	record, _, err := worktree.ReadSpecCopy(home, task.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("read spec copy-in record: %s", oneLine(err.Error()))
	}
	atCopyIn := map[string]string{}
	for _, file := range record.Files {
		atCopyIn[file.Path] = file.SHA256
	}
	effects := []output.Effect{}
	conflicts := []string{}
	seen := map[string]bool{}
	err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		inner, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		inner = filepath.ToSlash(inner)
		if entry.IsDir() {
			if inner == "tasks" {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative := relDir + inner
		seen[relative] = true
		taskHash, err := worktree.HashFile(path)
		if err != nil {
			return err
		}
		baseline, recorded := atCopyIn[relative]
		if recorded && taskHash == baseline {
			return nil
		}
		if !recorded {
			baseline = missingHash
		}
		target := filepath.Join(home, filepath.FromSlash(relative))
		homeHash, err := worktree.HashFile(target)
		if errors.Is(err, os.ErrNotExist) {
			homeHash = missingHash
		} else if err != nil {
			return err
		}
		if homeHash == taskHash {
			return nil
		}
		if homeHash != baseline {
			conflict := fmt.Sprintf("%s (copy-in sha256 %s, home sha256 %s)", relative, baseline, homeHash)
			conflicts = append(conflicts, conflict)
			effects = append(effects, output.Effect{Resource: "spec", Path: relative, Action: "conflict"})
			logSpec(logger, "land spec conflict", task, "path", relative, "copy_in_sha256", baseline, "home_sha256", homeHash)
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
		effects = append(effects, output.Effect{Resource: "spec", Path: relative, Action: "copied"})
		logSpec(logger, "land spec copy", task, "path", relative, "sha256", taskHash)
		return nil
	})
	if err != nil {
		return effects, conflicts, fmt.Errorf("copy spec out: %s", oneLine(err.Error()))
	}
	deleted := []string{}
	for path := range atCopyIn {
		if !seen[path] && !strings.HasPrefix(path, relDir+"tasks/") {
			deleted = append(deleted, path)
		}
	}
	sort.Strings(deleted)
	for _, path := range deleted {
		effects = append(effects, output.Effect{Resource: "spec", Path: path, Action: "kept"})
		logSpec(logger, "land spec kept", task, "path", path, "reason", "deleted in the worktree")
	}
	return effects, conflicts, nil
}

// landSpec runs the spec copy-out after a merge and adds its result to the
// envelope. A conflict or an error makes an ok landing partial. The code
// merge stays.
func landSpec(home string, logger *slog.Logger, task records.Task, envelope output.Envelope) output.Envelope {
	result, ok := envelope.Result.(Result)
	if !ok || result.Commit == "" {
		return envelope
	}
	effects, conflicts, err := copySpecOut(home, task, logger)
	envelope.Effects = append(envelope.Effects, effects...)
	problems := []string{}
	if len(conflicts) > 0 {
		problems = append(problems, "spec conflict: "+strings.Join(conflicts, "; ")+"; hive did not write these files: merge each file by hand from the worktree to the home")
	}
	if err != nil {
		problems = append(problems, err.Error())
	}
	if len(problems) == 0 {
		return envelope
	}
	message := strings.Join(problems, "; ")
	logFailure(logger, task.Project, task.ID, "spec", message)
	if envelope.Outcome == "ok" {
		return partialEnvelope("spec", message, envelope.Effects, envelope.Result)
	}
	envelope.Error = fmt.Sprint(envelope.Error) + "; " + oneLine(message)
	return envelope
}

func logSpec(logger *slog.Logger, message string, task records.Task, attrs ...any) {
	logEvent(logger, message, task.Project, task.ID, attrs...)
}
