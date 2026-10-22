package records

import "regexp"

var projectIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Project is a registered Git repository.
type Project struct {
	ID              string            `json:"id"`
	Path            string            `json:"path"`
	Base            string            `json:"base"`
	RepoWorkspaceID string            `json:"repo_workspace_id,omitempty"`
	Workspace       *ProjectWorkspace `json:"workspace,omitempty"`
	RegisteredAt    string            `json:"registered_at"`
}

// ProjectWorkspace is the manager workspace of one project. The ID identifies
// the workspace that hive made or adopted. CreatedByHive is true only for the
// workspace that this ID names.
type ProjectWorkspace struct {
	ID               string `json:"id"`
	RootPane         string `json:"root_pane"`
	CreatedByHive    bool   `json:"created_by_hive"`
	CreatedAt        string `json:"created_at,omitempty"`
	TabID            string `json:"tab_id,omitempty"`
	TabCreatedByHive bool   `json:"tab_created_by_hive,omitempty"`
}

// ValidID reports whether id is valid for a project record and directory.
func ValidID(id string) bool {
	return len(id) <= 18 && projectIDPattern.MatchString(id)
}
