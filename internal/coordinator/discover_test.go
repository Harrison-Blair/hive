package coordinator

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// fakeRunner returns out for the one expected command line.
func fakeRunner(t *testing.T, want string, out []byte, err error) Runner {
	t.Helper()
	return func(name string, args ...string) ([]byte, error) {
		got := strings.Join(append([]string{name}, args...), " ")
		if got != want {
			t.Fatalf("command = %q; want %q", got, want)
		}
		return out, err
	}
}

func TestDiscoverPiListsProviderSlashModel(t *testing.T) {
	models, err := Discover("pi", fakeRunner(t, "pi --list-models", fixture(t, "pi-list-models.txt"), nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 38 {
		t.Fatalf("len = %d; want 38 (39 lines less the header): %v", len(models), models)
	}
	for _, want := range []string{"openai-codex/gpt-6.1-sol", "opencode-go/claude-haiku-5-5", "opencode-go/step-5-preview"} {
		if !slices.Contains(models, want) {
			t.Errorf("models lack %q", want)
		}
	}
	if slices.Contains(models, "provider/model") {
		t.Errorf("models hold the header row")
	}
}

func TestDiscoverCodexKeepsOpenAICodexRowsWithoutPrefix(t *testing.T) {
	models, err := Discover("codex", fakeRunner(t, "pi --list-models", fixture(t, "pi-list-models.txt"), nil))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-6-astra", "gpt-6-luna", "gpt-6-sol", "gpt-6.1-sol"}
	if !slices.Equal(models, want) {
		t.Fatalf("models = %v; want %v", models, want)
	}
}

func TestDiscoverClaudeKeepsOpencodeClaudeRowsWithoutPrefix(t *testing.T) {
	models, err := Discover("claude", fakeRunner(t, "opencode models", fixture(t, "opencode-models.txt"), nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 14 {
		t.Fatalf("len = %d; want 14: %v", len(models), models)
	}
	for _, want := range []string{"claude-opus-5-5", "claude-fable-5-1", "claude-haiku-4-5"} {
		if !slices.Contains(models, want) {
			t.Errorf("models lack %q", want)
		}
	}
	for _, model := range models {
		if !strings.HasPrefix(model, "claude-") {
			t.Errorf("model %q is not a claude model without prefix", model)
		}
	}
}

func TestDiscoverCommandFailureIsAnError(t *testing.T) {
	_, err := Discover("claude", fakeRunner(t, "opencode models", nil, errors.New("exec: \"opencode\": executable file not found")))
	if err == nil || !strings.Contains(err.Error(), "opencode") {
		t.Fatalf("err = %v; want an error that names opencode", err)
	}
}

func TestDiscoverEmptyListIsAnError(t *testing.T) {
	_, err := Discover("codex", fakeRunner(t, "pi --list-models", []byte("provider  model\nanthropic  claude-x\n"), nil))
	if err == nil || !strings.Contains(err.Error(), "no models") {
		t.Fatalf("err = %v; want a no models error", err)
	}
}

func TestDiscoverUnknownHarnessIsAnError(t *testing.T) {
	_, err := Discover("vim", func(string, ...string) ([]byte, error) {
		t.Fatal("runner called for unknown harness")
		return nil, nil
	})
	if err == nil {
		t.Fatal("err = nil; want an error")
	}
}
