package homeinit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestInitializeKeepsCoordinatorConfigAndComments(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	homePath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(homePath, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(homePath, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "hive", "config.toml")
	content := "# keep this comment\nhome = " + strconv.Quote(homePath) + "\n\n[coordinator]\n# launch profile\nharness = \"codex\"\nmodel = \"small-model\"\n"
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	envelope := Initialize(homePath)
	if envelope.Outcome != "ok" || len(envelope.Effects) != 0 || envelope.Result != (Result{Home: homePath}) {
		t.Fatalf("Initialize() = %+v; want ok with no effects and home %q", envelope, homePath)
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if string(got) != content {
		t.Fatalf("config after init = %q; want %q", got, content)
	}
}

func markedHome(t *testing.T) string {
	t.Helper()
	homePath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(homePath, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(homePath, ".hive", "README.md"), []byte("home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return homePath
}

func TestInitializeWritesConfigAndReportsTheEffect(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	homePath := markedHome(t)

	envelope := Initialize(homePath)
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "hive", "config.toml")
	want := `{"outcome":"ok","phase":"complete","effects":[{"resource":"config","path":` + strconv.Quote(configPath) + `,"action":"created"}],"error":null,"result":{"home":` + strconv.Quote(homePath) + `}}`
	if string(data) != want {
		t.Fatalf("Initialize() = %s; want %s", data, want)
	}
	if got := Text(envelope); got != "Initialized hive home at "+homePath+"\n" {
		t.Fatalf("Text() = %q", got)
	}
}

func TestTextOfAnExistingConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	homePath := markedHome(t)
	_ = Initialize(homePath)

	if got := Text(Initialize(homePath)); got != "Hive home already initialized at "+homePath+"\n" {
		t.Fatalf("Text() = %q", got)
	}
}

func TestInitializeConflictFails(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	other := markedHome(t)
	_ = Initialize(other)
	homePath := markedHome(t)

	envelope := Initialize(homePath)
	want := "home conflict: config is " + other + ", working directory is in " + homePath
	if envelope.Outcome != "failed" || envelope.Phase != "command" || envelope.Error != want || envelope.Result != nil || Text(envelope) != "" {
		t.Fatalf("Initialize() = %+v, Text %q; want failed command envelope %q and no text", envelope, Text(envelope), want)
	}
}

func TestFindHomeReturnsTheOutermostMarkedHome(t *testing.T) {
	homePath := markedHome(t)
	nested := filepath.Join(homePath, "projects", "p", "worktrees", "w")
	if err := os.MkdirAll(filepath.Join(nested, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, ".hive", "README.md"), []byte("nested\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := FindHome(nested); err != nil || got != homePath {
		t.Fatalf("FindHome() = %q, %v; want %q", got, err, homePath)
	}
	outside := t.TempDir()
	if _, err := FindHome(outside); err == nil || err.Error() != "not a hive home: no marker above "+outside {
		t.Fatalf("FindHome(outside) error = %v; want the no-marker error", err)
	}
}
