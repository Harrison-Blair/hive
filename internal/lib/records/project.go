package records

import "regexp"

var projectIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Project is a registered Git repository.
type Project struct {
	ID              string `json:"id"`
	Path            string `json:"path"`
	Base            string `json:"base"`
	RepoWorkspaceID string `json:"repo_workspace_id,omitempty"`
	RegisteredAt    string `json:"registered_at"`
}

// ValidID reports whether id is valid for a project record and directory.
func ValidID(id string) bool {
	return len(id) <= 18 && projectIDPattern.MatchString(id)
}
