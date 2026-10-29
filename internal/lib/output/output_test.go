package output

import (
	"bytes"
	"testing"
)

func TestFailedHasNoEffectsAndNoResult(t *testing.T) {
	var stdout bytes.Buffer
	if err := Render(&stdout, &bytes.Buffer{}, true, Failed("agent", "no agent"), ""); err != nil {
		t.Fatal(err)
	}
	const want = `{"outcome":"failed","phase":"agent","effects":[],"error":"no agent","result":null}` + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q; want %q", stdout.String(), want)
	}
}

func TestRenderPrintsNilEffectsAsAnEmptyArray(t *testing.T) {
	var stdout bytes.Buffer
	if err := Render(&stdout, &bytes.Buffer{}, true, Envelope{Outcome: "ok", Phase: "complete"}, ""); err != nil {
		t.Fatal(err)
	}
	const want = `{"outcome":"ok","phase":"complete","effects":[],"error":null,"result":null}` + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q; want %q", stdout.String(), want)
	}
}

func TestOutcomeConstantsAndStatus(t *testing.T) {
	for _, test := range []struct {
		outcome, want string
		status        int
	}{
		{OutcomeOK, "ok", 0},
		{OutcomePartial, "partial", 1},
		{OutcomeFailed, "failed", 1},
	} {
		if test.outcome != test.want || Status(test.outcome) != test.status {
			t.Errorf("outcome %q, status %d; want %q, status %d", test.outcome, Status(test.outcome), test.want, test.status)
		}
	}
}
