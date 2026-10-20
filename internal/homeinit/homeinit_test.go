package homeinit

import (
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

	result, err := Initialize(homePath)
	if err != nil {
		t.Fatalf("Initialize(): %v", err)
	}
	if result.Home != homePath {
		t.Fatalf("Initialize().Home = %q; want %q", result.Home, homePath)
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if string(got) != content {
		t.Fatalf("config after init = %q; want %q", got, content)
	}
}
