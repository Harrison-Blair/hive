package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func seedProject(t *testing.T, home, id string) {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	record := records.Project{ID: id, Path: "/tmp/" + id, Base: "dev", RegisteredAt: "2026-01-01T00:00:00Z"}
	if err := st.WriteIn([]string{"projects"}, id, record); err != nil {
		t.Fatal(err)
	}
}

func readAgent(t *testing.T, home, key string) records.Agent {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := st.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatalf("read agent %s: %v", key, err)
	}
	return record
}

func readClaim(t *testing.T, home, key string) records.Claim {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var claim records.Claim
	if err := st.ReadIn([]string{"claims"}, key, &claim); err != nil {
		t.Fatalf("read claim %s: %v", key, err)
	}
	return claim
}

func agentKeys(t *testing.T, home string) []string {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := st.ListIn([]string{"agents"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(raw))
	for _, data := range raw {
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, record.Key)
	}
	return keys
}

func TestRegisterRoleRules(t *testing.T) {
	cases := []struct {
		name    string
		options RegisterOptions
		want    string
	}{
		{"missing role", RegisterOptions{PaneID: "w1:p1"}, "invalid role: "},
		{"unknown role", RegisterOptions{Role: "boss", PaneID: "w1:p1"}, "invalid role: boss"},
		{"manager without project", RegisterOptions{Role: "manager", PaneID: "w1:p1"}, "manager requires --project"},
		{"worker without project", RegisterOptions{Role: "worker", PaneID: "w1:p1"}, "worker requires --project and --task"},
		{"worker without task", RegisterOptions{Role: "worker", Project: "hive", PaneID: "w1:p1"}, "worker requires --project and --task"},
		{"verifier without task", RegisterOptions{Role: "verifier", Project: "hive", PaneID: "w1:p1"}, "verifier requires --project and --task"},
		{"coordinator with project", RegisterOptions{Role: "coordinator", Project: "hive", PaneID: "w1:p1"}, "coordinator takes no --project or --task"},
		{"supervisor with task", RegisterOptions{Role: "supervisor", Task: "T1", PaneID: "w1:p1"}, "supervisor takes no --project or --task"},
		{"no pane", RegisterOptions{Role: "coordinator"}, "no pane: run inside herdr or pass --pane"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			seedProject(t, home, "hive")
			envelope := Register(home, test.options)
			if envelope.Outcome != "failed" {
				t.Fatalf("Register() = %#v; want failed", envelope)
			}
			message, _ := envelope.Error.(string)
			if !strings.Contains(message, test.want) {
				t.Fatalf("error = %q; want text %q", message, test.want)
			}
		})
	}
}

func TestRegisterManagerWithoutTaskSucceeds(t *testing.T) {
	home := t.TempDir()
	seedProject(t, home, "hive")
	envelope := Register(home, RegisterOptions{Role: "manager", Project: "hive", PaneID: "w2:p1"})
	if envelope.Outcome != "ok" || envelope.Error != nil {
		t.Fatalf("Register() = %#v; want success", envelope)
	}
	record := readAgent(t, home, "manager-hive-g1")
	if record.Role != "manager" || record.Project != "hive" || record.Task != "" || record.Status != records.AgentLive {
		t.Fatalf("manager record = %#v", record)
	}
	claim := readClaim(t, home, "manager-hive")
	if claim.Generation != 1 || claim.Agent != "manager-hive-g1" {
		t.Fatalf("claim = %#v", claim)
	}
}

func TestRegisterRejectsUnregisteredProject(t *testing.T) {
	home := t.TempDir()
	envelope := Register(home, RegisterOptions{Role: "manager", Project: "nope", PaneID: "w1:p1"})
	if envelope.Outcome != "failed" || envelope.Error != "project not found: nope" {
		t.Fatalf("Register() = %#v; want project not found", envelope)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents")); !os.IsNotExist(err) {
		t.Fatalf("unregistered project created an agent record: %v", err)
	}
}

func TestRegisterRejectsInvalidNamesBeforeFileOperation(t *testing.T) {
	names := []string{"", "A", "1abc", "a b", "a/b", "..", "a.", strings.Repeat("a", 33), "a\nb", `a"b`, "a#b", "a?b", "a%b", "a&b", "a;b", "a+b", "é"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			envelope := Register(home, RegisterOptions{Role: "coordinator", Name: name, NameSet: true, PaneID: "w1:p1"})
			if envelope.Outcome != "failed" {
				t.Fatalf("Register() = %#v; want failed", envelope)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
				t.Fatalf("invalid name created state: %v", err)
			}
		})
	}
}

func TestRegisterRejectsUnsafePaneValues(t *testing.T) {
	for _, pane := range []string{"bad\npane", "bad\x01pane", strings.Repeat("p", 600)} {
		t.Run(fmt.Sprintf("%q", pane), func(t *testing.T) {
			home := t.TempDir()
			envelope := Register(home, RegisterOptions{Role: "coordinator", PaneID: pane})
			if envelope.Outcome != "failed" {
				t.Fatalf("Register() = %#v; want failed", envelope)
			}
			if _, err := os.Stat(filepath.Join(home, ".hive", "state")); !os.IsNotExist(err) {
				t.Fatalf("invalid pane created state: %v", err)
			}
		})
	}
}

func TestRegisterTwoCoordinatorsMoveTheClaim(t *testing.T) {
	home := t.TempDir()
	first := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p1"})
	if first.Outcome != "ok" {
		t.Fatal(first.Error)
	}
	second := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p2"})
	if second.Outcome != "ok" {
		t.Fatal(second.Error)
	}
	firstRecord := readAgent(t, home, "coordinator-g1")
	secondRecord := readAgent(t, home, "coordinator-g2")
	if firstRecord.Status != records.AgentStale || secondRecord.Status != records.AgentLive {
		t.Fatalf("statuses = %q, %q; want stale, live", firstRecord.Status, secondRecord.Status)
	}
	if firstRecord.PaneID != "w1:p1" || secondRecord.PaneID != "w1:p2" {
		t.Fatalf("panes = %q, %q", firstRecord.PaneID, secondRecord.PaneID)
	}
	claim := readClaim(t, home, "coordinator")
	if claim.Generation != 2 || claim.Agent != "coordinator-g2" {
		t.Fatalf("claim = %#v", claim)
	}

	supervisor := Register(home, RegisterOptions{Role: "supervisor", PaneID: "w1:p2"})
	if supervisor.Outcome != "failed" || supervisor.Error != "pane busy: w1:p2 holds coordinator" {
		t.Fatalf("supervisor on live pane = %#v; want pane busy", supervisor)
	}

	third := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p2"})
	if third.Outcome != "ok" {
		t.Fatal(third.Error)
	}
	if got := readAgent(t, home, "coordinator-g2").Status; got != records.AgentStale {
		t.Fatalf("generation 2 status = %q; want stale", got)
	}
	claim = readClaim(t, home, "coordinator")
	if claim.Generation != 3 || claim.Agent != "coordinator-g3" {
		t.Fatalf("claim = %#v", claim)
	}
	keys := agentKeys(t, home)
	if len(keys) != 3 {
		t.Fatalf("agent keys = %v; want three generation records", keys)
	}
}

func TestRegisterExplicitSingletonWritesRecordAndClaim(t *testing.T) {
	home := t.TempDir()
	envelope := Register(home, RegisterOptions{Role: "coordinator", Name: "boss", NameSet: true, PaneID: "w1:p1"})
	if envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	record := readAgent(t, home, "boss-g1")
	if record.Name != "boss" || record.Generation != 1 || record.Role != "coordinator" {
		t.Fatalf("record = %#v", record)
	}
	claim := readClaim(t, home, "coordinator")
	if claim.Agent != "boss-g1" {
		t.Fatalf("claim = %#v", claim)
	}
}

func TestRegisterDuplicateLiveNameAndBusyPane(t *testing.T) {
	home := t.TempDir()
	seedProject(t, home, "hive")
	first := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "w", NameSet: true, PaneID: "w1:p1"})
	if first.Outcome != "ok" {
		t.Fatal(first.Error)
	}
	duplicate := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "w", NameSet: true, PaneID: "w1:p2"})
	if duplicate.Outcome != "failed" || duplicate.Error != "agent exists: w" {
		t.Fatalf("duplicate name = %#v; want agent exists", duplicate)
	}
	busy := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "other", NameSet: true, PaneID: "w1:p1"})
	if busy.Outcome != "failed" || busy.Error != "pane busy: w1:p1 holds w" {
		t.Fatalf("busy pane = %#v; want pane busy", busy)
	}
	if keys := agentKeys(t, home); len(keys) != 1 {
		t.Fatalf("keys = %v; want only the first worker", keys)
	}
}

func TestRegisterGeneratedNamesAreUniqueUnderParallelLoad(t *testing.T) {
	home := t.TempDir()
	seedProject(t, home, "hive")
	const workers = 20
	var wg sync.WaitGroup
	errs := make(chan string, workers)
	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			envelope := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", PaneID: fmt.Sprintf("w%d:p1", i)})
			if envelope.Outcome != "ok" {
				errs <- fmt.Sprintf("%v", envelope.Error)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for message := range errs {
		t.Fatal(message)
	}
	keys := agentKeys(t, home)
	if len(keys) != workers {
		t.Fatalf("agent count = %d; want %d", len(keys), workers)
	}
	seen := map[string]bool{}
	for _, key := range keys {
		if seen[key] {
			t.Fatalf("duplicate key %q", key)
		}
		seen[key] = true
	}
}

type seamWriter struct {
	*store.LockedStore
	writeCollection  string
	updateCollection string
	writeErr         error
	updateErr        error
	nextErr          error
	writes           int
	updates          int
	nexts            int
}

func (w *seamWriter) WriteIn(collection []string, id string, value any) error {
	w.writes++
	if w.writeErr != nil && strings.Join(collection, "/") == w.writeCollection {
		return w.writeErr
	}
	return w.LockedStore.WriteIn(collection, id, value)
}

func (w *seamWriter) UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	w.updates++
	if w.updateErr != nil && strings.Join(collection, "/") == w.updateCollection {
		return w.updateErr
	}
	return w.LockedStore.UpdateIn(collection, id, fn)
}

func (w *seamWriter) Next(name string) (uint64, error) {
	w.nexts++
	if w.nextErr != nil {
		return 0, w.nextErr
	}
	return w.LockedStore.Next(name)
}

func runRegisterLocked(t *testing.T, home string, writer *seamWriter, options RegisterOptions) output.Envelope {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var envelope output.Envelope
	err = st.WithLock(context.Background(), func(locked *store.LockedStore) error {
		writer.LockedStore = locked
		envelope = registerLocked(locked, writer, options)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return envelope
}

func TestRegisterStaleMarkFailureChangesNothing(t *testing.T) {
	home := t.TempDir()
	if envelope := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	seam := &seamWriter{updateCollection: "agents", updateErr: errors.New("seam stale failure")}
	envelope := runRegisterLocked(t, home, seam, RegisterOptions{Role: "coordinator", PaneID: "w1:p2"})
	if envelope.Outcome != "failed" || envelope.Phase != "agent" || len(envelope.Effects) != 0 {
		t.Fatalf("Register() = %#v; want failed with no effect", envelope)
	}
	if claim := readClaim(t, home, "coordinator"); claim.Generation != 1 || claim.Agent != "coordinator-g1" {
		t.Fatalf("claim changed: %#v", claim)
	}
	if record := readAgent(t, home, "coordinator-g1"); record.Status != records.AgentLive {
		t.Fatalf("holder status = %q; want live", record.Status)
	}
	if keys := agentKeys(t, home); len(keys) != 1 {
		t.Fatalf("keys = %v; want the old record only", keys)
	}
}

func TestRegisterNewAgentFailureLeavesVacantRoleAndRetries(t *testing.T) {
	home := t.TempDir()
	if envelope := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	seam := &seamWriter{writeCollection: "agents", writeErr: errors.New("seam agent failure")}
	envelope := runRegisterLocked(t, home, seam, RegisterOptions{Role: "coordinator", PaneID: "w1:p2"})
	if envelope.Outcome != "partial" || envelope.Phase != "agent" {
		t.Fatalf("Register() = %#v; want partial agent", envelope)
	}
	if len(envelope.Effects) != 1 || envelope.Effects[0] != (output.Effect{Resource: "agent", Path: ".hive/state/agents/coordinator-g1.json", Action: "stale"}) {
		t.Fatalf("effects = %#v; want one stale effect", envelope.Effects)
	}
	if claim := readClaim(t, home, "coordinator"); claim.Generation != 1 || claim.Agent != "coordinator-g1" {
		t.Fatalf("claim = %#v; want unchanged", claim)
	}
	if record := readAgent(t, home, "coordinator-g1"); record.Status != records.AgentStale {
		t.Fatalf("old holder status = %q; want stale", record.Status)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "agents", "coordinator-g2.json")); !os.IsNotExist(err) {
		t.Fatalf("failed new-agent write created a record: %v", err)
	}
	retry := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p2"})
	if retry.Outcome != "ok" {
		t.Fatal(retry.Error)
	}
	if claim := readClaim(t, home, "coordinator"); claim.Generation != 2 || claim.Agent != "coordinator-g2" {
		t.Fatalf("retry claim = %#v", claim)
	}
	if record := readAgent(t, home, "coordinator-g2"); record.Status != records.AgentLive {
		t.Fatalf("retry record = %#v", record)
	}
}

func TestRegisterClaimFailureLeavesOrphanAndRetriesWithUpdate(t *testing.T) {
	home := t.TempDir()
	if envelope := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	seam := &seamWriter{writeCollection: "claims", writeErr: errors.New("seam claim failure")}
	envelope := runRegisterLocked(t, home, seam, RegisterOptions{Role: "coordinator", PaneID: "w1:p2"})
	if envelope.Outcome != "partial" || envelope.Phase != "claim" {
		t.Fatalf("Register() = %#v; want partial claim", envelope)
	}
	if len(envelope.Effects) != 2 || envelope.Effects[0].Action != "stale" || envelope.Effects[1].Action != "created" {
		t.Fatalf("effects = %#v; want stale and created", envelope.Effects)
	}
	if claim := readClaim(t, home, "coordinator"); claim.Generation != 1 || claim.Agent != "coordinator-g1" {
		t.Fatalf("claim = %#v; want unchanged", claim)
	}
	if record := readAgent(t, home, "coordinator-g1"); record.Status != records.AgentStale {
		t.Fatalf("old holder = %#v; want stale", record)
	}
	if record := readAgent(t, home, "coordinator-g2"); record.Status != records.AgentLive {
		t.Fatalf("orphan = %#v; want live", record)
	}
	retry := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p2"})
	if retry.Outcome != "ok" {
		t.Fatal(retry.Error)
	}
	if len(retry.Effects) != 2 || retry.Effects[0].Action != "updated" || retry.Effects[1].Action != "updated" {
		t.Fatalf("retry effects = %#v; want agent updated and claim updated", retry.Effects)
	}
	if claim := readClaim(t, home, "coordinator"); claim.Generation != 2 || claim.Agent != "coordinator-g2" {
		t.Fatalf("retry claim = %#v", claim)
	}
	if keys := agentKeys(t, home); len(keys) != 2 {
		t.Fatalf("keys = %v; want both generation records", keys)
	}
}

func TestRegisterGeneratedNameCounterFailureKeepsGap(t *testing.T) {
	home := t.TempDir()
	seedProject(t, home, "hive")
	seam := &seamWriter{writeCollection: "agents", writeErr: errors.New("seam agent failure")}
	options := RegisterOptions{Role: "worker", Project: "hive", Task: "T1", PaneID: "w1:p1"}
	envelope := runRegisterLocked(t, home, seam, options)
	if envelope.Outcome != "partial" || envelope.Phase != "agent" {
		t.Fatalf("Register() = %#v; want partial agent", envelope)
	}
	if len(envelope.Effects) != 1 || envelope.Effects[0].Resource != "counter" || envelope.Effects[0].Action != "updated" {
		t.Fatalf("effects = %#v; want one counter effect", envelope.Effects)
	}
	if keys := agentKeys(t, home); len(keys) != 0 {
		t.Fatalf("keys = %v; want no record", keys)
	}
	retry := Register(home, options)
	if retry.Outcome != "ok" {
		t.Fatal(retry.Error)
	}
	keys := agentKeys(t, home)
	if len(keys) != 1 || keys[0] != "worker-hive-2" {
		t.Fatalf("retry keys = %v; want worker-hive-2", keys)
	}
}

func TestRegisterExplicitWorkerUsesNoCounter(t *testing.T) {
	home := t.TempDir()
	seedProject(t, home, "hive")
	envelope := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "fixed", NameSet: true, PaneID: "w1:p1"})
	if envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "counters", "agent.json")); !os.IsNotExist(err) {
		t.Fatalf("explicit worker wrote the agent counter: %v", err)
	}
}

// The corrected T04 contract refuses the generated name before the counter
// write. A retry therefore sees the same collision and still writes nothing.
func TestRegisterGeneratedNameCollisionWritesNothing(t *testing.T) {
	home := t.TempDir()
	seedProject(t, home, "hive")
	explicit := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "worker-hive-1", NameSet: true, PaneID: "w1:p1"})
	if explicit.Outcome != "ok" {
		t.Fatal(explicit.Error)
	}
	options := RegisterOptions{Role: "worker", Project: "hive", Task: "T1", PaneID: "w1:p2"}
	envelope := Register(home, options)
	if envelope.Outcome != "failed" || len(envelope.Effects) != 0 {
		t.Fatalf("Register() = %#v; want failed with no effect before the counter", envelope)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "counters", "agent.json")); !os.IsNotExist(err) {
		t.Fatalf("refused registration wrote the agent counter: %v", err)
	}
	if keys := agentKeys(t, home); len(keys) != 1 {
		t.Fatalf("keys = %v; want the explicit worker only", keys)
	}
	retry := Register(home, options)
	if retry.Outcome != "failed" || len(retry.Effects) != 0 {
		t.Fatalf("retry = %#v; want the same collision with no effect", retry)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "counters", "agent.json")); !os.IsNotExist(err) {
		t.Fatalf("retry wrote the agent counter: %v", err)
	}
	if keys := agentKeys(t, home); len(keys) != 1 {
		t.Fatalf("retry keys = %v; want the explicit worker only", keys)
	}
}

func TestRegisterBusyPaneRefusalWritesNothing(t *testing.T) {
	home := t.TempDir()
	seedProject(t, home, "hive")
	first := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "busy", NameSet: true, PaneID: "w1:p1"})
	if first.Outcome != "ok" {
		t.Fatal(first.Error)
	}
	before := agentKeys(t, home)
	envelope := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", PaneID: "w1:p1"})
	if envelope.Outcome != "failed" || len(envelope.Effects) != 0 {
		t.Fatalf("Register() = %#v; want failed with no effect on a busy pane", envelope)
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "counters", "agent.json")); !os.IsNotExist(err) {
		t.Fatalf("busy-pane refusal wrote the agent counter: %v", err)
	}
	after := agentKeys(t, home)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Fatalf("agent keys = %v; want unchanged %v", after, before)
	}
}

func TestRegisterRejectsLiveNameAcrossRoles(t *testing.T) {
	t.Run("worker then coordinator", func(t *testing.T) {
		home := t.TempDir()
		seedProject(t, home, "hive")
		worker := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "shared", NameSet: true, PaneID: "w1:p1"})
		if worker.Outcome != "ok" {
			t.Fatal(worker.Error)
		}
		coordinator := Register(home, RegisterOptions{Role: "coordinator", Name: "shared", NameSet: true, PaneID: "w1:p2"})
		if coordinator.Outcome != "failed" || len(coordinator.Effects) != 0 {
			t.Fatalf("Register() = %#v; want refused live-name collision", coordinator)
		}
		if _, err := os.Stat(filepath.Join(home, ".hive", "state", "claims", "coordinator.json")); !os.IsNotExist(err) {
			t.Fatalf("collision wrote a coordinator claim: %v", err)
		}
	})
	t.Run("coordinator then worker", func(t *testing.T) {
		home := t.TempDir()
		seedProject(t, home, "hive")
		coordinator := Register(home, RegisterOptions{Role: "coordinator", Name: "shared", NameSet: true, PaneID: "w1:p1"})
		if coordinator.Outcome != "ok" {
			t.Fatal(coordinator.Error)
		}
		worker := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "shared", NameSet: true, PaneID: "w1:p2"})
		if worker.Outcome != "failed" || len(worker.Effects) != 0 {
			t.Fatalf("Register() = %#v; want refused live-name collision", worker)
		}
	})
	t.Run("singleton name matches a record key", func(t *testing.T) {
		home := t.TempDir()
		if first := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p1"}); first.Outcome != "ok" {
			t.Fatal(first.Error)
		}
		supervisor := Register(home, RegisterOptions{Role: "supervisor", Name: "coordinator-g1", NameSet: true, PaneID: "w1:p2"})
		if supervisor.Outcome != "failed" || len(supervisor.Effects) != 0 {
			t.Fatalf("Register() = %#v; want refused record-key name collision", supervisor)
		}
	})
	t.Run("coordinator generation then supervisor", func(t *testing.T) {
		home := t.TempDir()
		if first := Register(home, RegisterOptions{Role: "coordinator", Name: "shared", NameSet: true, PaneID: "w1:p1"}); first.Outcome != "ok" {
			t.Fatal(first.Error)
		}
		if second := Register(home, RegisterOptions{Role: "coordinator", Name: "shared", NameSet: true, PaneID: "w1:p2"}); second.Outcome != "ok" {
			t.Fatal(second.Error)
		}
		supervisor := Register(home, RegisterOptions{Role: "supervisor", Name: "shared", NameSet: true, PaneID: "w1:p3"})
		if supervisor.Outcome != "failed" || len(supervisor.Effects) != 0 {
			t.Fatalf("Register() = %#v; want refused live-name collision", supervisor)
		}
		if _, err := os.Stat(filepath.Join(home, ".hive", "state", "claims", "supervisor.json")); !os.IsNotExist(err) {
			t.Fatalf("collision wrote a supervisor claim: %v", err)
		}
	})
}

func TestRegisterRejectsClaimAliasBothOrders(t *testing.T) {
	t.Run("worker then coordinator", func(t *testing.T) {
		home := t.TempDir()
		seedProject(t, home, "hive")
		worker := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "coordinator", NameSet: true, PaneID: "w1:p1"})
		if worker.Outcome != "ok" {
			t.Fatal(worker.Error)
		}
		coordinator := Register(home, RegisterOptions{Role: "coordinator", Name: "leader", NameSet: true, PaneID: "w1:p2"})
		if coordinator.Outcome != "failed" || len(coordinator.Effects) != 0 {
			t.Fatalf("Register() = %#v; want refused claim-key alias", coordinator)
		}
		if _, err := os.Stat(filepath.Join(home, ".hive", "state", "claims", "coordinator.json")); !os.IsNotExist(err) {
			t.Fatalf("collision wrote a coordinator claim: %v", err)
		}
	})
	t.Run("coordinator then worker", func(t *testing.T) {
		home := t.TempDir()
		seedProject(t, home, "hive")
		coordinator := Register(home, RegisterOptions{Role: "coordinator", Name: "leader", NameSet: true, PaneID: "w1:p1"})
		if coordinator.Outcome != "ok" {
			t.Fatal(coordinator.Error)
		}
		worker := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "coordinator", NameSet: true, PaneID: "w1:p2"})
		if worker.Outcome != "failed" || len(worker.Effects) != 0 {
			t.Fatalf("Register() = %#v; want refused claim-key alias", worker)
		}
	})
}

func TestRegisterOrphanRetryRequiresSamePane(t *testing.T) {
	home := t.TempDir()
	options := RegisterOptions{Role: "coordinator", PaneID: "intended-pane"}
	seam := &seamWriter{writeCollection: "claims", writeErr: errors.New("seam claim failure")}
	first := runRegisterLocked(t, home, seam, options)
	if first.Outcome != "partial" || first.Phase != "claim" {
		t.Fatalf("first registration = %#v; want partial claim", first)
	}
	before, err := os.ReadFile(filepath.Join(home, ".hive", "state", "agents", "coordinator-g1.json"))
	if err != nil {
		t.Fatal(err)
	}
	options.PaneID = "different-pane"
	retry := Register(home, options)
	if retry.Outcome != "failed" || len(retry.Effects) != 0 {
		t.Fatalf("retry = %#v; want refused different pane with no effect", retry)
	}
	after, err := os.ReadFile(filepath.Join(home, ".hive", "state", "agents", "coordinator-g1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("refused retry changed the orphan record")
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "state", "claims", "coordinator.json")); !os.IsNotExist(err) {
		t.Fatalf("refused retry wrote a claim: %v", err)
	}
}

func TestRegisterErrorsEscapeControlInput(t *testing.T) {
	cases := []struct {
		label string
		want  string
		get   func(string) output.Envelope
	}{
		{"role", "invalid role", func(home string) output.Envelope {
			return Register(home, RegisterOptions{Role: "a\nb", PaneID: "w1:p1"})
		}},
		{"name", "invalid agent name", func(home string) output.Envelope {
			return Register(home, RegisterOptions{Role: "coordinator", Name: "a\nb", NameSet: true, PaneID: "w1:p1"})
		}},
		{"project", "project not found", func(home string) output.Envelope {
			return Register(home, RegisterOptions{Role: "manager", Project: "a\nb", PaneID: "w1:p1"})
		}},
		{"show", "agent not found", func(home string) output.Envelope {
			return Show(home, "a\nb")
		}},
	}
	for _, test := range cases {
		t.Run(test.label, func(t *testing.T) {
			envelope := test.get(t.TempDir())
			if envelope.Outcome != "failed" {
				t.Fatalf("envelope = %#v; want failed", envelope)
			}
			message, _ := envelope.Error.(string)
			if !strings.Contains(message, test.want) {
				t.Fatalf("error = %q; want text %q", message, test.want)
			}
			if strings.ContainsAny(message, "\n\r\x1b") {
				t.Fatalf("error = %q; want one line without terminal controls", message)
			}
		})
	}
}

func readAgentLogEvents(t *testing.T, home string) []map[string]any {
	t.Helper()
	dir := filepath.Join(home, ".hive", "logs")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Fatalf("log line is not JSON: %v; line %q", err, line)
			}
			events = append(events, event)
		}
	}
	return events
}

func findAgentLogEvent(events []map[string]any, message string) map[string]any {
	for _, event := range events {
		if event["msg"] == message {
			return event
		}
	}
	return nil
}

func TestRegisterLogsTransitionsAndErrors(t *testing.T) {
	home := t.TempDir()
	if envelope := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
		t.Fatal(envelope.Error)
	}
	events := readAgentLogEvents(t, home)
	recordEvent := findAgentLogEvent(events, "agent record written")
	if recordEvent == nil || recordEvent["agent_id"] != "coordinator-g1" || recordEvent["role"] != "coordinator" {
		t.Fatalf("record event = %#v; want agent_id coordinator-g1 and role coordinator", recordEvent)
	}
	claimEvent := findAgentLogEvent(events, "agent claim written")
	if claimEvent == nil || claimEvent["agent_id"] != "coordinator-g1" {
		t.Fatalf("claim event = %#v; want agent_id coordinator-g1", claimEvent)
	}

	errorHome := t.TempDir()
	if envelope := Register(errorHome, RegisterOptions{Role: "boss", PaneID: "w1:p1"}); envelope.Outcome != "failed" {
		t.Fatalf("invalid role = %#v; want failed", envelope)
	}
	errorEvents := readAgentLogEvents(t, errorHome)
	failureEvent := findAgentLogEvent(errorEvents, "agent registration failed")
	if failureEvent == nil {
		t.Fatalf("events = %#v; want an agent registration failure event", errorEvents)
	}
	message, _ := failureEvent["error"].(string)
	if !strings.Contains(message, "invalid role: boss") || failureEvent["role"] != "boss" {
		t.Fatalf("failure event = %#v; want role boss and the error text", failureEvent)
	}
}

func TestRegisterRejectsCrossClaimAndGenerationKeyCollisions(t *testing.T) {
	t.Run("cross claim", func(t *testing.T) {
		home := t.TempDir()
		if envelope := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
			t.Fatal(envelope.Error)
		}
		envelope := Register(home, RegisterOptions{Role: "supervisor", Name: "coordinator", NameSet: true, PaneID: "w1:p2"})
		if envelope.Outcome != "failed" {
			t.Fatalf("Register() = %#v; want collision failure", envelope)
		}
		if _, err := os.Stat(filepath.Join(home, ".hive", "state", "claims", "supervisor.json")); !os.IsNotExist(err) {
			t.Fatalf("collision wrote a supervisor claim: %v", err)
		}
		if keys := agentKeys(t, home); len(keys) != 1 {
			t.Fatalf("keys = %v; want old coordinator only", keys)
		}
	})

	t.Run("worker generation key", func(t *testing.T) {
		home := t.TempDir()
		seedProject(t, home, "hive")
		if envelope := Register(home, RegisterOptions{Role: "coordinator", PaneID: "w1:p1"}); envelope.Outcome != "ok" {
			t.Fatal(envelope.Error)
		}
		envelope := Register(home, RegisterOptions{Role: "worker", Project: "hive", Task: "T1", Name: "coordinator-g1", NameSet: true, PaneID: "w1:p2"})
		if envelope.Outcome != "failed" {
			t.Fatalf("Register() = %#v; want generation key collision", envelope)
		}
		if keys := agentKeys(t, home); len(keys) != 1 {
			t.Fatalf("keys = %v; want old coordinator only", keys)
		}
	})
}
