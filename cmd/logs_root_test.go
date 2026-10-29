package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func dailyLogPath(home string) string {
	return filepath.Join(home, ".hive", "logs", time.Now().Local().Format("2006-01-02")+".jsonl")
}

func makeLogFixture(t *testing.T) (string, string) {
	t.Helper()
	home := testHome(t)
	path := dailyLogPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	return home, path
}

func TestLogsJSONReadErrorsUsePlainDiagnostics(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, home, path string)
	}{
		{
			name: "wrong type directory",
			setup: func(t *testing.T, _, path string) {
				t.Helper()
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "symbolic link",
			setup: func(t *testing.T, home, path string) {
				t.Helper()
				target := filepath.Join(home, "target.jsonl")
				if err := os.WriteFile(target, []byte("{\"msg\":\"target\"}\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "unreadable file",
			setup: func(t *testing.T, _, path string) {
				t.Helper()
				if err := os.WriteFile(path, []byte("{\"msg\":\"unreadable\"}\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(path, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home, path := makeLogFixture(t)
			test.setup(t, home, path)

			var stdout, stderr bytes.Buffer
			status := executeWithStreams([]string{"--json", "logs"}, strings.NewReader(""), &stdout, &stderr)
			if status != 1 {
				t.Fatalf("status = %d; want 1 (stdout %q, stderr %q)", status, stdout.String(), stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q; want empty", stdout.String())
			}
			got := stderr.String()
			if strings.Count(got, "\n") != 1 {
				t.Fatalf("stderr = %q; want one plain diagnostic line", got)
			}
			if !strings.HasPrefix(got, "hive: ") {
				t.Fatalf("stderr = %q; want the hive: prefix", got)
			}
			if strings.Contains(got, "\"outcome\"") || strings.Contains(got, "panic") || strings.Contains(got, "goroutine") {
				t.Fatalf("stderr = %q; want one plain diagnostic", got)
			}
		})
	}
}

func TestLogsJSONPrintsRawEventLines(t *testing.T) {
	_, path := makeLogFixture(t)
	raw := `{"time":"2026-10-18T10:00:00-04:00","level":"INFO","component":"cli","msg":"ok"}`
	if err := os.WriteFile(path, []byte(raw+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--json", "logs"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d; want 0 (stdout %q, stderr %q)", status, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q; want empty", stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 2 || lines[0] != raw {
		t.Fatalf("stdout = %q; want the raw event followed by the logs start event", stdout.String())
	}
	var start struct {
		Command   string `json:"command"`
		Message   string `json:"msg"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &start); err != nil {
		t.Fatal(err)
	}
	if start.Command != "logs" || start.Message != "command start" || !regexp.MustCompile(`^r-[0-9a-f]{8}$`).MatchString(start.RequestID) {
		t.Fatalf("logs start event = %+v; want a request ID and command start", start)
	}
}

func TestLogsHasNoTaskFlag(t *testing.T) {
	home := testHome(t)
	t.Setenv("HIVE_HOME", home)
	var output strings.Builder
	err := ExecuteWithArgs([]string{"logs", "--task", "x"}, strings.NewReader(""), &output)
	if err == nil || !strings.Contains(err.Error()+output.String(), "unknown flag: --task") {
		t.Fatalf("logs --task: err %v, output %q; want unknown flag", err, output.String())
	}
}

func TestLogsBadFlagValuesAreUsageErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"bad since", []string{"logs", "--since", "bogus"}, `invalid --since value \"bogus\": use a duration or an RFC 3339 time`},
		{"negative limit", []string{"logs", "--limit", "-1"}, "--limit must be zero or more"},
	}
	for _, test := range cases {
		t.Run(test.name+" json", func(t *testing.T) {
			testHome(t)
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(append([]string{"--json"}, test.args...), strings.NewReader(""), &stdout, &stderr)
			want := `{"outcome":"failed","phase":"usage","effects":[],"error":"` + test.want + `","result":null}` + "\n"
			if status != 2 || stdout.String() != want || stderr.Len() != 0 {
				t.Fatalf("status = %d, stdout = %q, stderr = %q; want 2, %q and no stderr", status, stdout.String(), stderr.String(), want)
			}
		})
		t.Run(test.name+" text", func(t *testing.T) {
			testHome(t)
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(test.args, strings.NewReader(""), &stdout, &stderr)
			want := "hive: " + strings.ReplaceAll(test.want, `\"`, `"`) + "\n"
			if status != 2 || stdout.Len() != 0 || stderr.String() != want {
				t.Fatalf("status = %d, stdout = %q, stderr = %q; want 2 and %q", status, stdout.String(), stderr.String(), want)
			}
		})
	}
}
