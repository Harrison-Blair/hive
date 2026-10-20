package records

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGoalIDValidation(t *testing.T) {
	for _, id := range []string{"g1-demo", "g20-multi-word-goal", "g2-" + strings.Repeat("a", 60)} {
		if !ValidGoalID(id) {
			t.Errorf("ValidGoalID(%q) = false; want true", id)
		}
	}
	for _, id := range []string{"1-demo", "g0-demo", "g1-Upper", "g1-two--words", "g1-"} {
		if ValidGoalID(id) {
			t.Errorf("ValidGoalID(%q) = true; want false", id)
		}
	}
}

func TestGoalJSONShape(t *testing.T) {
	goal := Goal{
		ID: "g1-demo", Project: "hive", Title: "Demo", Grant: "grant line", Status: "active",
		Waves:     []GoalWave{{ID: "A", Title: "First", DoneWhen: "dev", Tasks: []string{"1-task"}, Status: "open"}},
		CreatedAt: "2026-10-20T00:00:00Z", UpdatedAt: "2026-10-20T00:00:00Z",
	}
	data, err := json.Marshal(goal)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"id":"g1-demo"`, `"grant":"grant line"`, `"done_when":"dev"`, `"tasks":["1-task"]`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("goal JSON = %s; want %s", data, want)
		}
	}
	if strings.Contains(string(data), `"needs_user"`) {
		t.Errorf("goal JSON = %s; empty needs_user should be omitted", data)
	}
}
