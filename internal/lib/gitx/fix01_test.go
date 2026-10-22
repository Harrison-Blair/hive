package gitx

import (
	"context"
	"testing"
)

func TestFixRefExistsDistinguishesAbsenceFromFailure(t *testing.T) {
	ctx := context.Background()
	notGit := t.TempDir()
	exists, err := RefExists(ctx, notGit, "main")
	if err == nil || exists {
		t.Fatalf("RefExists(non-repository) = %v, %v, want an error", exists, err)
	}

	repo := newRepo(t)
	exists, err = RefExists(ctx, repo, "main")
	if err != nil || !exists {
		t.Fatalf("RefExists(main) = %v, %v, want true, nil", exists, err)
	}
	exists, err = RefExists(ctx, repo, "missing")
	if err != nil || exists {
		t.Fatalf("RefExists(missing) = %v, %v, want false, nil", exists, err)
	}
}
