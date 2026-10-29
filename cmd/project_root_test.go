package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootProjectListJSONEnvelope(t *testing.T) {
	testHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--json", "project", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q; want empty", stderr.String())
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v; output %q", err, stdout.String())
	}
	for _, key := range []string{"outcome", "phase", "effects", "error", "result"} {
		if _, ok := envelope[key]; !ok {
			t.Errorf("envelope has no %q field: %s", key, stdout.String())
		}
	}
	if string(envelope["result"]) != "[]" {
		t.Errorf("result = %s; want []", envelope["result"])
	}
}

func TestRootMissingHomeUsesOneErrorLine(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HIVE_HOME", "")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"project", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stdout %q, stderr %q)", status, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q; want empty", stdout.String())
	}
	if got := stderr.String(); !strings.HasPrefix(got, "hive: not a hive home: no marker above ") || strings.Count(got, "\n") != 1 {
		t.Fatalf("stderr = %q; want one hive error line", got)
	}
}

func TestRootMissingHomeJSONEnvelope(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HIVE_HOME", "")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--json", "project", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 1 || stderr.Len() != 0 {
		t.Fatalf("status = %d, stderr = %q; want status 1 and empty stderr", status, stderr.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("stdout is not one JSON object: %v; output %q", err, stdout.String())
	}
	if envelope["outcome"] != "failed" || envelope["phase"] != "home" || envelope["error"] == nil {
		t.Fatalf("envelope = %#v; want failed home error", envelope)
	}
}

func TestRegisterEmptyPathIsUsageError(t *testing.T) {
	testHome(t)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"project", "register", ""}, strings.NewReader(""), &stdout, &stderr)
	if status != 2 || stdout.Len() != 0 || stderr.String() != "hive: PATH requires a value\n" {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want one PATH usage line", status, stdout.String(), stderr.String())
	}
}

func TestHomeFlagUnknownPrecedesHomeResolution(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HIVE_HOME", "")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--" + "home", "", "project", "list"}, strings.NewReader(""), &stdout, &stderr)
	if status != 2 || stdout.Len() != 0 || stderr.String() != "hive: unknown flag: --"+"home\n" {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want unknown flag status 2", status, stdout.String(), stderr.String())
	}
}

func testHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	marker := filepath.Join(home, ".hive", "README.md")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("home"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", home)
	return home
}
