package worktree

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// TestRemoveRefusesUnknownOccupancy proves B1: a failed agent.list reads an
// unknown occupancy. The removal must refuse before workspace.close and the
// Git removal, with the force flag or without it.
func TestRemoveRefusesUnknownOccupancy(t *testing.T) {
	for _, force := range []bool{false, true} {
		t.Run(fmt.Sprintf("force=%t", force), func(t *testing.T) {
			repo, home, path := guardFixture(t)
			fake := startFakeHerdr(t, func(request fakeRequest) fakeReply {
				switch request.Method {
				case "workspace.get":
					return fakeReply{Result: workspaceInfoReply(t, linkedWorkspace("w2", path, repo))}
				case "agent.list":
					return fakeReply{Error: `{"code":"timeout","message":"request timed out"}`}
				}
				return fakeReply{Error: `{"code":"unknown_method","message":"unexpected method"}`}
			})
			envelope := Remove(context.Background(), home, RemoveOptions{Task: "1-live-demo", Force: force})
			if envelope.Outcome != "failed" || envelope.Phase != "agent" {
				t.Fatalf("Remove() = %+v, want a failed agent guard", envelope)
			}
			want := "agent occupancy unknown: timeout: request timed out"
			if got := fmt.Sprint(envelope.Error); got != want {
				t.Fatalf("Remove() error = %q, want %q", got, want)
			}
			for _, method := range fake.methods() {
				if method == "workspace.close" {
					t.Fatalf("workspace.close ran after the occupancy read failed; methods = %v", fake.methods())
				}
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("worktree removed after the occupancy read failed: %v", err)
			}
		})
	}
}
