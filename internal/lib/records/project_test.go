package records

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestProjectDecodesOldWorkspaceFields proves that a project record written
// with the removed workspace fields still decodes, and that a rewrite drops
// them.
func TestProjectDecodesOldWorkspaceFields(t *testing.T) {
	const old = `{"id":"hive","path":"/src/hive","base":"main","repo_workspace_id":"w1",` +
		`"workspace":{"id":"w2","root_pane":"w2:p1","created_by_hive":true,"tab_id":"w2:t1"},` +
		`"registered_at":"2026-01-02T03:04:05Z"}`
	var project Project
	if err := json.Unmarshal([]byte(old), &project); err != nil {
		t.Fatalf("decode old project: %v", err)
	}
	want := Project{ID: "hive", Path: "/src/hive", Base: "main", RegisteredAt: "2026-01-02T03:04:05Z"}
	if project != want {
		t.Fatalf("project = %+v; want %+v", project, want)
	}
	data, err := json.Marshal(project)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"repo_workspace_id", "workspace"} {
		if strings.Contains(string(data), `"`+field+`"`) {
			t.Errorf("project JSON = %s; want no %s field", data, field)
		}
	}
}
