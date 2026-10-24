package task

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// featureREADMEPath returns the absolute path of the README of one feature in
// the home.
func featureREADMEPath(home, feature string) string {
	return filepath.Join(home, ".hive", "spec", feature, "README.md")
}

// checkFeatureREADME checks that the home has the README of the feature.
func checkFeatureREADME(home, feature string) error {
	path := featureREADMEPath(home, feature)
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("feature README missing in home: %s", path)
	}
	if err != nil {
		return fmt.Errorf("check feature README %s: %s", path, oneLine(err.Error()))
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("feature README is not a regular file: %s", path)
	}
	return nil
}

// copyEvidenceOut copies the evidence of one task from its worktree to the
// home. Complete copies the evidence folder. Verify copies VERIFIED.md. A
// task without a feature or a worktree, or a missing source, copies nothing.
func copyEvidenceOut(home, command string, task records.Task) ([]output.Effect, error) {
	paths := TaskFeaturePaths(task)
	if paths == nil || task.Worktree == nil || task.Worktree.Path == "" {
		return nil, nil
	}
	source := filepath.Join(task.Worktree.Path, filepath.FromSlash(paths.Evidence))
	var files []string
	if command == "verify" {
		files = []string{"VERIFIED.md"}
	} else {
		err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			relative, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(relative))
			return nil
		})
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read evidence: %s", oneLine(err.Error()))
		}
	}
	effects := []output.Effect{}
	for _, name := range files {
		relative := paths.Evidence + name
		data, err := os.ReadFile(filepath.Join(source, filepath.FromSlash(name)))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return effects, fmt.Errorf("read evidence %s: %s", relative, oneLine(err.Error()))
		}
		if err := writeWorktreeBrief(home, relative, data); err != nil {
			return effects, fmt.Errorf("copy evidence %s: %s", relative, oneLine(err.Error()))
		}
		effects = append(effects, output.Effect{Resource: "evidence", Path: relative, Action: "copied"})
	}
	return effects, nil
}
