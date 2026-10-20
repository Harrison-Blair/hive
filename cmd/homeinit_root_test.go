package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestInitCreatesConfigForOutermostMarkedHome(t *testing.T) {
	configPath := initConfigPath(t)
	home := testHome(t)
	cwd := filepath.Join(home, "projects", "hive", "worktrees", "task")
	if err := os.MkdirAll(filepath.Join(cwd, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cwd, ".hive", "README.md"), []byte("nested home"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)
	t.Setenv("HIVE_HOME", "")

	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"init"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || stderr.Len() != 0 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want successful init", status, stdout.String(), stderr.String())
	}
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if want := "home = " + strconv.Quote(home) + "\n"; string(config) != want {
		t.Fatalf("config = %q; want %q", config, want)
	}
	if !strings.Contains(stdout.String(), home) {
		t.Fatalf("stdout = %q; want initialized home %q", stdout.String(), home)
	}
	log, err := os.ReadFile(dailyLogPath(home))
	if err != nil {
		t.Fatalf("read command log: %v", err)
	}
	if !strings.Contains(string(log), `"msg":"command start"`) || !strings.Contains(string(log), `"msg":"command end"`) || !strings.Contains(string(log), `"command":"init"`) {
		t.Fatalf("command log = %s; want init start and end events", log)
	}
}

func TestInitJSONReportsConfigWriteEffect(t *testing.T) {
	configPath := initConfigPath(t)
	home := testHome(t)
	t.Chdir(home)

	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--json", "init"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || stderr.Len() != 0 || strings.Count(stdout.String(), "\n") != 1 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want one JSON success envelope", status, stdout.String(), stderr.String())
	}
	var envelope struct {
		Outcome string `json:"outcome"`
		Phase   string `json:"phase"`
		Effects []struct {
			Resource string `json:"resource"`
			Path     string `json:"path"`
			Action   string `json:"action"`
		} `json:"effects"`
		Error  any               `json:"error"`
		Result map[string]string `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("stdout is not a JSON envelope: %v; %q", err, stdout.String())
	}
	if envelope.Outcome != "ok" || envelope.Phase != "complete" || envelope.Error != nil || len(envelope.Effects) != 1 || envelope.Effects[0] != (struct {
		Resource string `json:"resource"`
		Path     string `json:"path"`
		Action   string `json:"action"`
	}{Resource: "config", Path: configPath, Action: "created"}) || envelope.Result["home"] != home {
		t.Fatalf("envelope = %+v; want successful config-created effect", envelope)
	}
}

func TestInitSameHomeDoesNotChangeConfig(t *testing.T) {
	configPath := initConfigPath(t)
	home := testHome(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	config := []byte("home = " + strconv.Quote(home) + "\n")
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(home)

	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"init"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || stderr.Len() != 0 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want successful init", status, stdout.String(), stderr.String())
	}
	after, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(config) || !before.ModTime().Equal(after.ModTime()) {
		t.Fatalf("config changed: bytes %q, modtime %v -> %v", got, before.ModTime(), after.ModTime())
	}
}

func TestInitRejectsDifferentConfigWithoutWriting(t *testing.T) {
	configPath := initConfigPath(t)
	home := testHome(t)
	other := testHome(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	config := []byte("home = " + strconv.Quote(other) + "\n")
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(home)

	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--json", "init"}, strings.NewReader(""), &stdout, &stderr)
	if status != 1 || stderr.Len() != 0 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want JSON failure", status, stdout.String(), stderr.String())
	}
	var envelope struct {
		Outcome string          `json:"outcome"`
		Phase   string          `json:"phase"`
		Effects []any           `json:"effects"`
		Error   string          `json:"error"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("stdout is not a JSON envelope: %v; %q", err, stdout.String())
	}
	if envelope.Outcome != "failed" || envelope.Phase != "command" || len(envelope.Effects) != 0 || !strings.Contains(envelope.Error, home) || !strings.Contains(envelope.Error, other) {
		t.Fatalf("envelope = %+v; want failure naming both homes", envelope)
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(config) {
		t.Fatalf("config changed to %q; want %q", got, config)
	}
}

func TestInitWithoutMarkerIgnoresOtherHomeSources(t *testing.T) {
	configPath := initConfigPath(t)
	configured := testHome(t)
	envHome := testHome(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	config := []byte("home = " + strconv.Quote(configured) + "\n")
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", envHome)
	t.Chdir(t.TempDir())

	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"init"}, strings.NewReader(""), &stdout, &stderr)
	if status != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "no marker above") {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want no-marker failure", status, stdout.String(), stderr.String())
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(config) {
		t.Fatalf("config changed to %q; want %q", got, config)
	}
}

func TestInitRejectsArguments(t *testing.T) {
	configPath := initConfigPath(t)
	home := testHome(t)
	t.Chdir(home)

	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"init", "extra"}, strings.NewReader(""), &stdout, &stderr)
	if status != 2 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want argument rejection", status, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("config stat error = %v; want no config", err)
	}
}

func initConfigPath(t *testing.T) string {
	t.Helper()
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	return filepath.Join(configDir, "hive", "config.toml")
}
