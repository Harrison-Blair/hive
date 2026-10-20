package records

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTaskNeedsUserJSON(t *testing.T) {
	var task Task
	if err := json.Unmarshal([]byte(`{"id":"1-task","needs_user":"choose an option"}`), &task); err != nil {
		t.Fatal(err)
	}
	if task.NeedsUser != "choose an option" {
		t.Fatalf("NeedsUser = %q; want choose an option", task.NeedsUser)
	}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"needs_user":"choose an option"`) {
		t.Errorf("task JSON = %s; want needs_user", data)
	}

	var old Task
	if err := json.Unmarshal([]byte(`{"id":"2-old-task"}`), &old); err != nil {
		t.Fatalf("decode old task without needs_user: %v", err)
	}
	if old.NeedsUser != "" {
		t.Fatalf("old task NeedsUser = %q; want empty", old.NeedsUser)
	}
	data, err = json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"needs_user"`) {
		t.Errorf("old task JSON = %s; must omit empty needs_user", data)
	}
}

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
