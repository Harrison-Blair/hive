package create_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	hivecmd "github.com/Harrison-Blair/hive/cmd"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestCreateFeatureSpecWithoutREADME(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HIVE_HOME", home)
	projectPath := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, projectPath, "init", "-b", "main")
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Write("projects", "one", records.Project{ID: "one", Path: projectPath, Base: "main"}); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	err = hivecmd.ExecuteWithArgs([]string{"--json", "task", "create", "--project", "one", "--title", "Feature spec", "--feature", "f", "--spec"}, nil, &stdout)
	if err != nil {
		t.Fatalf("task create --feature f --spec: %v; output %s", err, stdout.String())
	}
	var envelope struct {
		Outcome string         `json:"outcome"`
		Result  map[string]any `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode output %q: %v", stdout.String(), err)
	}
	if envelope.Outcome != "ok" || envelope.Result["feature"] != "f" || envelope.Result["spec"] != true {
		t.Fatalf("result = %#v; want feature f and spec true", envelope)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
