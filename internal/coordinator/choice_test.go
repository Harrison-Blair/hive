package coordinator

import (
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestChoiceRoundTripUnderHiveDir(t *testing.T) {
	home := t.TempDir()
	if _, ok, err := LoadChoice(home); err != nil || ok {
		t.Fatalf("LoadChoice on empty home = ok %v, err %v; want none", ok, err)
	}
	want := Choice{Harness: "pi", Model: "openai-codex/gpt-6-sol", Models: map[string]string{"pi": "openai-codex/gpt-6-sol", "claude": "claude-opus-5-5"}}
	if err := SaveChoice(home, want); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, "coordinator-choice.json")); err != nil {
		t.Fatalf("choice file: %v", err)
	}
	got, ok, err := LoadChoice(home)
	if err != nil || !ok || got.Harness != want.Harness || got.Model != want.Model || !maps.Equal(got.Models, want.Models) {
		t.Fatalf("LoadChoice = %+v, %v, %v; want %+v", got, ok, err, want)
	}
}

func TestLoadChoiceBadFileIsAnError(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "coordinator-choice.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadChoice(home); err == nil {
		t.Fatal("err = nil; want a parse error")
	}
}

func TestLoadChoiceWithoutModelsSeedsTheMapFromTheLastChoice(t *testing.T) {
	home := t.TempDir()
	old := `{"harness":"codex","model":"gpt-6.1-sol"}`
	if err := os.WriteFile(filepath.Join(home, "coordinator-choice.json"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok, err := LoadChoice(home)
	if err != nil || !ok {
		t.Fatalf("LoadChoice = ok %v, err %v", ok, err)
	}
	if got.Harness != "codex" || got.Model != "gpt-6.1-sol" || !maps.Equal(got.Models, map[string]string{"codex": "gpt-6.1-sol"}) {
		t.Fatalf("LoadChoice = %+v; want codex gpt-6.1-sol with models {codex: gpt-6.1-sol}", got)
	}
}
