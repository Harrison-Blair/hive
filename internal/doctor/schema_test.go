package doctor

import (
	"context"
	"testing"
)

// TestDoctorShowsSchemaVersion covers AC6.
func TestDoctorShowsSchemaVersion(t *testing.T) {
	home := newHome(t)
	t.Setenv("HERDR_SOCKET_PATH", startHerdr(t, 22))
	t.Setenv("PATH", pathWithGitOnly(t))
	result := Run(context.Background(), home, nil)
	check := findCheck(t, result, "schema version")
	if check.Status != StatusOK || check.Detail != "3" {
		t.Fatalf("schema version = %+v; want ok with 3", check)
	}
}
