package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	guidecmd "github.com/Harrison-Blair/hive/cmd/guide"
	guidepkg "github.com/Harrison-Blair/hive/internal/guide"
)

type guideEnvelope struct {
	Outcome string          `json:"outcome"`
	Phase   string          `json:"phase"`
	Effects []guideEffect   `json:"effects"`
	Error   any             `json:"error"`
	Result  guideResultBody `json:"result"`
}

type guideEffect struct {
	Resource string `json:"resource"`
	Path     string `json:"path"`
	Action   string `json:"action"`
}

type guideResultBody struct {
	Flows     string `json:"flows"`
	Reference string `json:"reference"`
}

// TestGuideReferenceMatchesLiveTree proves AC3: the live tree and the tracked
// reference agree.
func TestGuideReferenceMatchesLiveTree(t *testing.T) {
	got := guidepkg.RenderReference(guidecmd.Plain(NewRootCmd()))
	if got != guidepkg.Reference() {
		t.Fatalf("the live tree differs from the tracked reference: run hive guide --write")
	}
}

// TestGuideTemporaryFlagFailsDriftThenRegenerationPasses proves AC3. A visible
// flag and a changed command entry each make the reference differ.
// Regeneration makes the flag case agree.
func TestGuideTemporaryFlagFailsDriftThenRegenerationPasses(t *testing.T) {
	root := NewRootCmd()
	tracked := guidepkg.Reference()
	reference := func() string { return guidepkg.RenderReference(guidecmd.Plain(root)) }
	if got := reference(); got != tracked {
		t.Fatalf("clean tree differs from the tracked reference")
	}
	version, _, err := root.Find([]string{"version"})
	if err != nil || version.Name() != "version" {
		t.Fatalf("find version command: %v, %v", version, err)
	}
	version.Flags().String("temporary-drift", "", "temporary drift probe")
	if got := reference(); got == tracked {
		t.Fatal("the reference did not change after a visible flag was added")
	}
	path := filepath.Join(t.TempDir(), "reference.md")
	if err := os.WriteFile(path, []byte(reference()), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), reference(); got != want {
		t.Fatalf("regenerated reference = %d bytes, want %d", len(got), len(want))
	}
	version.Short = "changed short help"
	if got := reference(); got == string(data) {
		t.Fatal("the reference did not change after a command entry changed")
	}
}

// TestGuideTextAndJSON proves AC1 and AC4 and the command output contract.
func TestGuideTextAndJSON(t *testing.T) {
	newRootTestHome(t)
	var out bytes.Buffer
	if err := ExecuteWithArgs([]string{"guide"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("guide: %v\n%s", err, out.String())
	}
	if got, want := out.String(), guidepkg.Text(); got != want {
		t.Fatalf("guide output = %d bytes, want flows then reference (%d bytes)", len(got), len(want))
	}
	if !strings.HasPrefix(out.String(), "# hive guide\n") || !strings.Contains(out.String(), "# hive reference\n") {
		t.Fatalf("guide output lacks a part heading")
	}

	out.Reset()
	if err := ExecuteWithArgs([]string{"--json", "guide"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("guide --json: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), `"effects":[]`) {
		t.Fatalf("guide --json lacks an empty effects array: %s", out.String())
	}
	var envelope guideEnvelope
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("output %q: %v", out.String(), err)
	}
	if envelope.Outcome != "ok" || envelope.Phase != "complete" || envelope.Error != nil {
		t.Fatalf("envelope = %+v, want outcome ok, phase complete, no error", envelope)
	}
	if len(envelope.Effects) != 0 {
		t.Fatalf("effects = %v, want none", envelope.Effects)
	}
	if envelope.Result.Flows != guidepkg.Flows() {
		t.Fatalf("result flows = %d bytes, want %d", len(envelope.Result.Flows), len(guidepkg.Flows()))
	}
	if envelope.Result.Reference != guidepkg.Reference() {
		t.Fatalf("result reference = %d bytes, want %d", len(envelope.Result.Reference), len(guidepkg.Reference()))
	}
}

// TestGuideWrite proves AC3 at the command line. The second run keeps the same
// file content.
func TestGuideWrite(t *testing.T) {
	newRootTestHome(t)
	dir := t.TempDir()
	referencePath := filepath.Join(dir, "internal", "guide", "reference.md")
	if err := os.MkdirAll(filepath.Dir(referencePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(referencePath, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	var out bytes.Buffer
	if err := ExecuteWithArgs([]string{"--json", "guide", "--write"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("guide --write: %v\n%s", err, out.String())
	}
	data, err := os.ReadFile(referencePath)
	if err != nil {
		t.Fatalf("read regenerated reference: %v", err)
	}
	if string(data) != guidepkg.Reference() {
		t.Fatalf("regenerated reference = %d bytes, want the tracked reference (%d bytes)", len(data), len(guidepkg.Reference()))
	}
	var envelope guideEnvelope
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("output %q: %v", out.String(), err)
	}
	if envelope.Outcome != "ok" || envelope.Phase != "complete" {
		t.Fatalf("envelope = %+v, want outcome ok and phase complete", envelope)
	}
	wantEffect := guideEffect{Resource: "guide reference", Path: "internal/guide/reference.md", Action: "updated"}
	if len(envelope.Effects) != 1 || envelope.Effects[0] != wantEffect {
		t.Fatalf("effects = %v, want [%+v]", envelope.Effects, wantEffect)
	}

	out.Reset()
	if err := ExecuteWithArgs([]string{"--json", "guide", "--write"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("second guide --write: %v\n%s", err, out.String())
	}
	second, err := os.ReadFile(referencePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(second, data) {
		t.Fatal("second guide --write changed the reference file")
	}
}
