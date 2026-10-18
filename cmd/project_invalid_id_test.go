package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterInvalidIDReportsRuleAndStatus(t *testing.T) {
	const want = "invalid project id: Repo_One: an ID matches ^[a-z0-9]+(-[a-z0-9]+)*$ and has at most 18 characters"
	home := testHome(t)
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		args []string
		json bool
	}{
		{"register text", []string{"--home", home, "project", "register", repo, "--id", "Repo_One"}, false},
		{"register json", []string{"--home", home, "--json", "project", "register", repo, "--id", "Repo_One"}, true},
		{"show text", []string{"--home", home, "project", "show", "Repo_One"}, false},
		{"show json", []string{"--home", home, "--json", "project", "show", "Repo_One"}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(test.args, strings.NewReader(""), &stdout, &stderr)
			if status != 1 {
				t.Fatalf("status = %d; want 1 (stdout %q, stderr %q)", status, stdout.String(), stderr.String())
			}
			if test.json {
				if stderr.Len() != 0 {
					t.Fatalf("stderr = %q; want empty", stderr.String())
				}
				if strings.Count(stdout.String(), "\n") != 1 {
					t.Fatalf("stdout = %q; want one JSON line", stdout.String())
				}
				var envelope map[string]any
				if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
					t.Fatalf("stdout is not one JSON object: %v; output %q", err, stdout.String())
				}
				if envelope["outcome"] != "failed" || envelope["error"] != want {
					t.Fatalf("envelope = %#v; want failed outcome and %q", envelope, want)
				}
				return
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q; want empty", stdout.String())
			}
			if got := stderr.String(); got != "hive: "+want+"\n" || strings.Count(got, "\n") != 1 {
				t.Fatalf("stderr = %q; want one error line with the rule", got)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
		t.Fatalf("invalid ID created state: stat error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "projects")); !os.IsNotExist(err) {
		t.Fatalf("invalid ID created project directories: stat error = %v", err)
	}
}
