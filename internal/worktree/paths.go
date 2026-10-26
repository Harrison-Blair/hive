package worktree

import "path/filepath"

// CanonicalPath returns the absolute path with each symbolic link resolved,
// or the cleaned path when it cannot resolve it.
func CanonicalPath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return filepath.Clean(absolute)
	}
	return resolved
}

// SamePath reports whether two paths have the same canonical path.
func SamePath(left, right string) bool {
	return CanonicalPath(left) == CanonicalPath(right)
}
