package coordinator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChoiceRoundTripUnderHiveDir(t *testing.T) {
	home := t.TempDir()
	if _, ok, err := LoadChoice(home); err != nil || ok {
		t.Fatalf("LoadChoice on empty home = ok %v, err %v; want none", ok, err)
	}
	want := Choice{Harness: "pi", Model: "openai-codex/gpt-6-sol"}
	if err := SaveChoice(home, want); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "coordinator-choice.json")); err != nil {
		t.Fatalf("choice file: %v", err)
	}
	got, ok, err := LoadChoice(home)
	if err != nil || !ok || got != want {
		t.Fatalf("LoadChoice = %+v, %v, %v; want %+v", got, ok, err, want)
	}
}

func TestLoadChoiceBadFileIsAnError(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".hive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".hive", "coordinator-choice.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadChoice(home); err == nil {
		t.Fatal("err = nil; want a parse error")
	}
}
