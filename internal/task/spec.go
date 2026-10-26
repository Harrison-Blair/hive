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
			return nil, fmt.Errorf("read evidence: %s", OneLine(err.Error()))
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
			return effects, fmt.Errorf("read evidence %s: %s", relative, OneLine(err.Error()))
		}
		if err := WriteWorktreeBrief(home, relative, data); err != nil {
			return effects, fmt.Errorf("copy evidence %s: %s", relative, OneLine(err.Error()))
		}
		effects = append(effects, output.Effect{Resource: "evidence", Path: relative, Action: "copied"})
	}
	return effects, nil
}
