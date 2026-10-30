package logs

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	return home
}

func runLogs(t *testing.T, home string, opts Options) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), home, opts, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

func messages(t *testing.T, home string, now time.Time, names ...string) {
	t.Helper()
	lines := make([]string, 0, len(names))
	for index, name := range names {
		lines = append(lines, eventLine(t, now.Add(time.Duration(index-len(names))*time.Second), map[string]any{"level": "INFO", "component": "cli", "msg": name}))
	}
	writeDaily(t, home, now, lines...)
}

func TestRunLimitKeepsTheLastTextMatches(t *testing.T) {
	home := runHome(t)
	messages(t, home, time.Now(), "one", "two", "three")

	stdout, stderr, err := runLogs(t, home, Options{Limit: 2})
	if err != nil || stderr != "" {
		t.Fatalf("Run: err %v, stderr %q", err, stderr)
	}
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], " two") || !strings.Contains(lines[1], " three") {
		t.Fatalf("stdout = %q; want the text lines of two and three", stdout)
	}
}

func TestRunLimitZeroPrintsNothing(t *testing.T) {
	home := runHome(t)
	messages(t, home, time.Now(), "one")

	stdout, stderr, err := runLogs(t, home, Options{Limit: 0})
	if err != nil || stdout != "" || stderr != "" {
		t.Fatalf("Run: err %v, stdout %q, stderr %q; want no output", err, stdout, stderr)
	}
}

func TestRunLimitKeepsTheLastJSONEventsAndCountsOnlyValidOnes(t *testing.T) {
	home := runHome(t)
	now := time.Now()
	first := eventLine(t, now.Add(-3*time.Second), map[string]any{"msg": "one"})
	second := eventLine(t, now.Add(-2*time.Second), map[string]any{"msg": "two"})
	third := eventLine(t, now.Add(-time.Second), map[string]any{"msg": "three"})
	writeDaily(t, home, now, first, second, "not json", third)

	stdout, stderr, err := runLogs(t, home, Options{Limit: 2, JSON: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if want := second + "\n" + third + "\n"; stdout != want {
		t.Fatalf("stdout = %q; want %q", stdout, want)
	}
	if want := "hive: malformed line " + dailyPath(home, now) + ":3\n"; stderr != want {
		t.Fatalf("stderr = %q; want %q", stderr, want)
	}
}

func TestRunTextPrintsAMalformedLineAsAMatch(t *testing.T) {
	home := runHome(t)
	now := time.Now()
	writeDaily(t, home, now, eventLine(t, now.Add(-time.Second), map[string]any{"msg": "one"}), "not json")

	stdout, stderr, err := runLogs(t, home, Options{Limit: 1})
	if err != nil || stderr != "" {
		t.Fatalf("Run: err %v, stderr %q", err, stderr)
	}
	if stdout != "? not json\n" {
		t.Fatalf("stdout = %q; want the malformed line only", stdout)
	}
}

func TestRunSinceDurationKeepsRecentEvents(t *testing.T) {
	home := runHome(t)
	now := time.Now()
	old := eventLine(t, now.Add(-15*time.Minute), map[string]any{"msg": "old"})
	recent := eventLine(t, now.Add(-5*time.Minute), map[string]any{"msg": "recent"})
	if day := now.Add(-15 * time.Minute); day.Day() != now.Day() {
		writeDaily(t, home, day, old)
		writeDaily(t, home, now, recent)
	} else {
		writeDaily(t, home, now, old, recent)
	}

	stdout, _, err := runLogs(t, home, Options{Since: "10m", Limit: 10, JSON: true})
	if err != nil || stdout != recent+"\n" {
		t.Fatalf("Run: err %v, stdout %q; want only %q", err, stdout, recent)
	}
}

func TestRunSinceRFC3339KeepsEventsAtAndAfter(t *testing.T) {
	home := runHome(t)
	now := time.Now().Truncate(time.Second)
	cutoff := now.Add(-2 * time.Second)
	before := eventLine(t, cutoff.Add(-time.Second), map[string]any{"msg": "before"})
	at := eventLine(t, cutoff, map[string]any{"msg": "at"})
	if cutoff.Add(-time.Second).Day() != now.Day() {
		t.Skip("the events cross midnight")
	}
	writeDaily(t, home, now, before, at)

	stdout, _, err := runLogs(t, home, Options{Since: cutoff.Format(time.RFC3339), Limit: 10, JSON: true})
	if err != nil || stdout != at+"\n" {
		t.Fatalf("Run: err %v, stdout %q; want only %q", err, stdout, at)
	}
}

func TestRunRejectsBadOptions(t *testing.T) {
	cases := []struct {
		name string
		opts Options
		want string
	}{
		{"bad since", Options{Since: "bogus", Limit: 1}, `invalid --since value "bogus": use a duration or an RFC 3339 time`},
		{"negative limit", Options{Limit: -1}, "--limit must be zero or more"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, err := runLogs(t, runHome(t), test.opts)
			var option *OptionError
			if !errors.As(err, &option) || option.Error() != test.want || stdout != "" || stderr != "" {
				t.Fatalf("Run: err %v, stdout %q, stderr %q; want OptionError %q and no output", err, stdout, stderr, test.want)
			}
		})
	}
}

func TestRunReadErrorWritesOneHiveLine(t *testing.T) {
	home := runHome(t)
	if err := os.Mkdir(dailyPath(home, time.Now()), 0o755); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := runLogs(t, home, Options{Limit: 1, JSON: true})
	if err == nil || stdout != "" || stderr != "hive: "+err.Error()+"\n" {
		t.Fatalf("Run: err %v, stdout %q, stderr %q; want one hive: line with the error", err, stdout, stderr)
	}
}
