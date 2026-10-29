package records

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestAgentRecordJSONHasNoRole proves AC1/D13: the agent record keeps facts,
// and it has no role compatibility field.
func TestAgentRecordJSONHasNoRole(t *testing.T) {
	record := Agent{
		Key: "owner-hive-g1", Name: "owner-hive", Profile: ProfilePlanner,
		Owns: "hive", Project: "hive", Status: AgentLive, Generation: 1,
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"role"`) {
		t.Fatalf("agent JSON = %s; want no role field", data)
	}
	for _, want := range []string{`"profile":"planner"`, `"owns":"hive"`, `"generation":1`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("agent JSON = %s; missing %s", data, want)
		}
	}
}

// TestInboxRecordJSONUsesInbox proves D2: the inbox record field role becomes
// inbox.
func TestInboxRecordJSONUsesInbox(t *testing.T) {
	record := Inbox{Seq: 1, Inbox: "coordinator", Kind: InboxMessage, Text: "hello"}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"role"`) {
		t.Fatalf("inbox JSON = %s; want no role field", data)
	}
	if !strings.Contains(string(data), `"inbox":"coordinator"`) {
		t.Fatalf("inbox JSON = %s; want the inbox field", data)
	}
}
