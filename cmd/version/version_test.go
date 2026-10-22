package version_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/cmd"
)

func TestVersionJSONEnvelope(t *testing.T) {
	var out bytes.Buffer
	if err := cmd.ExecuteWithArgs([]string{"--json", "version"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("version --json: %v", err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("output %q: %v", out.String(), err)
	}
	if envelope["outcome"] != "ok" {
		t.Fatalf("outcome = %v, want ok", envelope["outcome"])
	}
	result, ok := envelope["result"].(map[string]any)
	if !ok {
		t.Fatalf("result = %v, want an object", envelope["result"])
	}
	for _, key := range []string{"version", "commit", "modified", "go"} {
		if _, ok := result[key]; !ok {
			t.Fatalf("result has no %q: %v", key, result)
		}
	}
	if result["version"] != "dev" {
		t.Fatalf("version = %v, want dev", result["version"])
	}
}
