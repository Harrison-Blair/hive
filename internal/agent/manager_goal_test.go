package agent

import (
	"os"
	"strings"
	"testing"
)

func TestManagerInstructionsListGoalCommands(t *testing.T) {
	for _, path := range []string{"../../profiles/planner.md", "templates/project-agents.md"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(data)
		for _, want := range []string{"hive goal create", "hive goal edit", "hive goal show", "hive goal list"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s does not contain %q", path, want)
			}
		}
	}
	data, err := os.ReadFile("../../profiles/planner.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Add each new task to its wave.") {
		t.Fatal("manager skill does not tell managers to add each new task to its wave")
	}
}
