package records

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAgentStartupMarkRoundTripAndLegacyJSON(t *testing.T) {
	var started Agent
	if err := json.Unmarshal([]byte(`{"key":"coordinator-g1","name":"coordinator","role":"coordinator","status":"live","started_by_hive":true}`), &started); err != nil {
		t.Fatalf("decode startup record: %v", err)
	}
	encoded, err := json.Marshal(started)
	if err != nil {
		t.Fatalf("encode startup record: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("decode encoded startup record: %v", err)
	}
	if got["started_by_hive"] != true {
		t.Fatalf("started_by_hive = %#v; want true", got["started_by_hive"])
	}

	var legacy Agent
	if err := json.Unmarshal([]byte(`{"key":"old","name":"old","role":"coordinator","status":"live"}`), &legacy); err != nil {
		t.Fatalf("decode legacy record: %v", err)
	}
	legacyEncoded, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("encode legacy record: %v", err)
	}
	if strings.Contains(string(legacyEncoded), `"started_by_hive"`) {
		t.Fatalf("legacy record has a startup mark: %s", legacyEncoded)
	}
}

func TestAgentLineageFieldsRoundTripAndLegacyJSON(t *testing.T) {
	const fixture = `{"key":"worker-g2","name":"worker","role":"worker","status":"live","parent":"manager-g1","replaces":"worker-g1","status_changed_at":"2026-01-02T03:04:05Z"}`
	var agent Agent
	if err := json.Unmarshal([]byte(fixture), &agent); err != nil {
		t.Fatalf("decode lineage record: %v", err)
	}
	encoded, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("encode lineage record: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("decode encoded record: %v", err)
	}
	for key, want := range map[string]string{
		"parent":            "manager-g1",
		"replaces":          "worker-g1",
		"status_changed_at": "2026-01-02T03:04:05Z",
	} {
		if got[key] != want {
			t.Errorf("%s = %#v; want %q", key, got[key], want)
		}
	}

	var old Agent
	if err := json.Unmarshal([]byte(`{"key":"old","name":"old","role":"worker","status":"live"}`), &old); err != nil {
		t.Fatalf("decode old record: %v", err)
	}
	oldEncoded, err := json.Marshal(old)
	if err != nil {
		t.Fatalf("encode old record: %v", err)
	}
	for _, field := range []string{"parent", "replaces", "status_changed_at"} {
		if strings.Contains(string(oldEncoded), `"`+field+`"`) {
			t.Errorf("old record encoding has optional field %q: %s", field, oldEncoded)
		}
	}
}
