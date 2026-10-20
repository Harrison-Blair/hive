package home

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestReadSettingsLoadsCoordinatorAndDefaults(t *testing.T) {
	configPath := configFixture(t, "")
	configHome := makeHome(t, t.TempDir(), "configured-home")
	writeConfig(t, configPath, "# hive config\nhome = "+strconv.Quote(configHome)+"\n\n[coordinator]\n# launch profile\nharness = \"codex\"\nmodel = \"small-model\"\n")

	settings, err := ReadSettings()
	if err != nil {
		t.Fatalf("ReadSettings(): %v", err)
	}
	if settings.Home != configHome || settings.Coordinator.Harness != "codex" || settings.Coordinator.Model != "small-model" {
		t.Fatalf("ReadSettings() = %#v; want home %q, coordinator codex/small-model", settings, configHome)
	}

	writeConfig(t, configPath, "home = "+strconv.Quote(configHome)+"\n")
	settings, err = ReadSettings()
	if err != nil {
		t.Fatalf("ReadSettings() defaults: %v", err)
	}
	if settings.Coordinator.Harness != "claude" || settings.Coordinator.Model != "claude-opus-5-5" {
		t.Fatalf("coordinator defaults = %#v; want claude/claude-opus-5-5", settings.Coordinator)
	}
}

func TestReadSettingsRejectsUnknownKeyAndTableWithPath(t *testing.T) {
	configPath := configFixture(t, "")
	configHome := makeHome(t, t.TempDir(), "configured-home")
	for _, test := range []struct {
		name    string
		content string
		unknown string
	}{
		{name: "key", content: "home = " + strconv.Quote(configHome) + "\ncoordinator_extra = \"x\"\n", unknown: "coordinator_extra"},
		{name: "table", content: "home = " + strconv.Quote(configHome) + "\n[other]\n", unknown: "other"},
	} {
		t.Run(test.name, func(t *testing.T) {
			writeConfig(t, configPath, test.content)
			_, err := ReadSettings()
			if err == nil || !strings.Contains(err.Error(), configPath) || !strings.Contains(err.Error(), test.unknown) {
				t.Fatalf("ReadSettings() error = %v; want config path %q and key/table %q", err, configPath, test.unknown)
			}
		})
	}
}

func TestWriteConfigChangesOnlyHomeLine(t *testing.T) {
	configPath := configFixture(t, "")
	oldHome := makeHome(t, t.TempDir(), "old-home")
	newHome := makeHome(t, t.TempDir(), "new-home")
	before := "# keep this comment\nhome = " + strconv.Quote(oldHome) + "\n\n[coordinator]\n# keep this too\nharness = \"codex\"\nmodel = \"small-model\"\n"
	writeConfig(t, configPath, before)

	if err := WriteConfig(newHome); err != nil {
		t.Fatalf("WriteConfig(): %v", err)
	}
	want := strings.Replace(before, strconv.Quote(oldHome), strconv.Quote(newHome), 1)
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	if string(got) != want {
		t.Fatalf("updated config = %q; want %q", got, want)
	}
	if filepath.Clean(newHome) == filepath.Clean(oldHome) {
		t.Fatal("test homes are equal")
	}
}
