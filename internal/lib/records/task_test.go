package records

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTaskFeatureAndSpecJSON(t *testing.T) {
	var task Task
	if err := json.Unmarshal([]byte(`{"id":"1-task","feature":"f","spec":true}`), &task); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"feature":"f"`, `"spec":true`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("task JSON = %s; want %s", data, want)
		}
	}

	var old Task
	if err := json.Unmarshal([]byte(`{"id":"2-old-task"}`), &old); err != nil {
		t.Fatalf("decode old task without feature: %v", err)
	}
}
