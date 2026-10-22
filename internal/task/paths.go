package task

import (
	"fmt"
	"path/filepath"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// FeaturePaths holds the repository paths linked to a task feature.
type FeaturePaths struct {
	README   string
	Brief    string
	Evidence string
}

// TaskFeaturePaths returns the repository paths for a task with a feature.
func TaskFeaturePaths(task records.Task) *FeaturePaths {
	if task.Feature == "" {
		return nil
	}
	root := filepath.Join(".hive", "spec", task.Feature)
	return &FeaturePaths{
		README:   filepath.ToSlash(filepath.Join(root, "README.md")),
		Brief:    filepath.ToSlash(filepath.Join(root, "tasks", task.ID, "brief.md")),
		Evidence: filepath.ToSlash(filepath.Join(root, "tasks", task.ID, "evidence")) + "/",
	}
}

// checkFeature validates the feature slug.
func checkFeature(feature string) error {
	if feature == "" {
		return fmt.Errorf("feature requires a value")
	}
	if len(feature) > maxSlugLength || !slugPattern.MatchString(feature) {
		return fmt.Errorf("invalid feature: %s", quoteIfControl(feature))
	}
	return nil
}
