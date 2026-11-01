package stop

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/agent/spawn"
	"github.com/Harrison-Blair/hive/internal/lib/output"
)

// TestStopKeepsTheRecordOfASpawnOfTheSameNameThatFollows proves that a spawn
// of the name that starts while stop closes the pane gets its record, and
// stop does not remove that record.
func TestStopKeepsTheRecordOfASpawnOfTheSameNameThatFollows(t *testing.T) {
	fake := agenttest.NewHerdr(t)
	home := agenttest.Home(t)
	ws, old := fake.AddWorkspace("w", t.TempDir(), false)
	if err := agent.Save(home, agent.Record{Name: "kid", Pane: old.ID}); err != nil {
		t.Fatal(err)
	}
	spawned := make(chan output.Envelope, 1)
	var once sync.Once
	fake.After("pane.close", func() {
		once.Do(func() {
			go func() {
				spawned <- spawn.Run(context.Background(), home, spawn.Options{
					Name: "kid", Harness: "pi", Model: "prov/m1", Directory: t.TempDir(),
				})
			}()
			time.Sleep(300 * time.Millisecond)
		})
	})
	if envelope := stopRun(t, home, Options{Name: "kid"}); envelope.Outcome != output.OutcomeOK {
		t.Fatalf("stop = %+v", envelope)
	}
	envelope := <-spawned
	if envelope.Outcome != output.OutcomeOK {
		t.Fatalf("spawn = %+v", envelope)
	}
	record, err := agent.Load(home, "kid")
	if err != nil || record.Pane == old.ID || record.Pane != envelope.Result.(agent.Record).Pane {
		t.Fatalf("record = %+v, %v; want the record of the new pane, not %s of workspace %s", record, err, old.ID, ws.ID)
	}
}
