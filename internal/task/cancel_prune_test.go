package task

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func TestCancelPrunesStoppedActingAgent(t *testing.T) {
	home, task := newSeed(t, "assigned")
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.UpdateIn([]string{"agents"}, "worker-1", func(raw json.RawMessage) (json.RawMessage, error) {
		var record records.Agent
		if err := json.Unmarshal(raw, &record); err != nil {
			return nil, err
		}
		record.Status = records.AgentStopped
		return json.Marshal(record)
	}); err != nil {
		t.Fatal(err)
	}

	if envelope := Cancel(home, TransitionOptions{Task: task.ID, Reason: "stop work"}); envelope.Outcome != "ok" {
		t.Fatalf("Cancel() = %#v; want ok", envelope)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, "worker-1", &record); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stopped acting agent read = %v; want not exist", err)
	}
}
