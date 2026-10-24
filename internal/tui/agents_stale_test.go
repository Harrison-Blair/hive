package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Harrison-Blair/hive/internal/lib/process"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/charmbracelet/x/ansi"
)

// staleFixture reads the agent records of testdata/agents-stale. They have
// the shape of the records of a real home: the old generations have no
// parent and no replaces, and the newest generation has replaces.
func staleFixture(t *testing.T, keys ...string) []records.Agent {
	t.Helper()
	var out []records.Agent
	for _, key := range keys {
		data, err := os.ReadFile(filepath.Join("testdata", "agents-stale", key+".json"))
		if err != nil {
			t.Fatal(err)
		}
		var record records.Agent
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		out = append(out, record)
	}
	return out
}

// generations gives the keys name-g1 to name-gN.
func generations(name string, last int) []string {
	var keys []string
	for index := 1; index <= last; index++ {
		keys = append(keys, name+"-g"+string(rune('0'+index)))
	}
	return keys
}

// staleHome gives a home with the projects qmeter and stenographer, a live
// supervisor claim, and the agent records.
func staleHome(t *testing.T, agents []records.Agent) (string, *testClock) {
	t.Helper()
	home := newTestHome(t)
	clock := fixedClock(t, time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC))
	pidStart, err := process.Start(int64(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, home, []string{"claims"}, "supervisor", records.Claim{Key: "supervisor", Agent: "supervisor", PID: int64(os.Getpid()), PIDStart: pidStart})
	for _, id := range []string{"qmeter", "stenographer"} {
		writeRecord(t, home, []string{"projects"}, id, records.Project{ID: id, Path: "/repo/" + id, Base: "dev"})
	}
	for _, record := range agents {
		writeAgentRecord(t, home, record)
	}
	return home, clock
}

func allStaleFixtures(t *testing.T) []records.Agent {
	t.Helper()
	keys := generations("coordinator", 8)
	keys = append(keys, generations("manager-qmeter", 4)...)
	keys = append(keys, generations("manager-stenographer", 8)...)
	return staleFixture(t, keys...)
}

// Each agent name has one current record: the live record, or else the
// newest record.
func TestCurrentAgentKeysGiveOneRecordForEachName(t *testing.T) {
	current := currentAgentKeys(allStaleFixtures(t))
	var got []string
	for key := range current {
		got = append(got, key)
	}
	slices.Sort(got)
	want := []string{"coordinator-g8", "manager-qmeter-g4", "manager-stenographer-g8"}
	if !slices.Equal(got, want) {
		t.Fatalf("current records = %v; want %v", got, want)
	}
}

// The live record is current when a stale record of its lineage is newer.
func TestCurrentAgentKeysPreferTheLiveRecord(t *testing.T) {
	agents := staleFixture(t, "manager-qmeter-g4", "manager-qmeter-g3")
	// A stale manager of qmeter with another name, a parent, and a change
	// after the live g4 changed.
	other := &agents[1]
	other.Key, other.Name, other.Parent = "manager-qmeter-other", "manager-qmeter-other", "coordinator-g8"
	other.StatusChangedAt = "2026-10-23T02:30:00Z"
	current := currentAgentKeys(agents)
	if !current["manager-qmeter-g4"] || len(current) != 1 {
		t.Fatalf("current records = %v; want only manager-qmeter-g4", current)
	}
}

// The overview draws one card for each name, from the live or the newest
// record.
func TestAgentsOverviewDrawsOneCardForEachName(t *testing.T) {
	home, clock := staleHome(t, allStaleFixtures(t))
	m := openAgents(t, home, clock)
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m,
		"◐ coordinator", "claude · idle · 9m",
		"● manager-qmeter", "pi · working · 1h",
		"○ manager-stenograp…", "claude · stopped · …",
	)
	rejectView(t, m, "unknown")
	view := ansi.Strip(m.View().Content)
	for _, name := range []string{"manager-qmeter", "manager-stenograp"} {
		if count := strings.Count(view, name); count != 1 {
			t.Fatalf("the overview shows %d cards of %s; want 1:\n%s", count, name, view)
		}
	}
}

// The one-pod view of qmeter shows the live manager record.
func TestAgentsPodViewDrawsTheLiveManager(t *testing.T) {
	home, clock := staleHome(t, allStaleFixtures(t))
	m := openAgents(t, home, clock)
	m.graph.pod, m.graph.podView = "qmeter", true
	m.invalidateView()
	requireView(t, m, "● manager-qmeter", "pi · working · 1h")
	rejectView(t, m, "claude · unknown")
}

// A stale worker with no live successor keeps its stale state on the graph.
// h hides it with the stopped agents and counts it.
func TestAgentsHHidesAStaleWorkerWithNoSuccessor(t *testing.T) {
	agents := staleFixture(t, "coordinator-g8", "manager-qmeter-g4")
	// A worker record with the shape of manager-qmeter-g1: stale, with no
	// parent and no replaces.
	worker := staleFixture(t, "manager-qmeter-g1")[0]
	worker.Key, worker.Name, worker.Role, worker.Task = "worker-qmeter-g1", "worker-qmeter", "worker", "07-old"
	agents = append(agents, worker)
	home, clock := staleHome(t, agents)
	m := openAgents(t, home, clock)
	requireView(t, m, "● manager-qmeter", "h shows 1 stopped")
	rejectView(t, m, "worker-qmeter")
	m, _ = raise(t, m, key('h', "h"))
	m.graph.pod, m.graph.podView = "qmeter", true
	m.invalidateView()
	requireView(t, m, "○ worker-qmeter", "claude · stale", "h hides 1 stopped")
	m, _ = raise(t, m, key('h', "h"))
	requireView(t, m, "h shows 1 stopped")
	rejectView(t, m, "worker-qmeter")
}

// The details popup of the coordinator card gives the live record.
func TestAgentsCoordinatorIsTheLiveRecord(t *testing.T) {
	home, clock := staleHome(t, allStaleFixtures(t))
	m := openAgents(t, home, clock)
	node, ok := m.coordinatorNode()
	if !ok || node.record.Key != "coordinator-g8" || node.status != "idle" {
		t.Fatalf("coordinator card = %s %s (%t); want coordinator-g8 idle", node.record.Key, node.status, ok)
	}
}

// Health › Agents lists one row for each name, from the live or the newest
// record.
func TestHealthAgentsListOneRowForEachName(t *testing.T) {
	source := agentsSource{agents: allStaleFixtures(t), supervisorLive: true}
	state := healthState{agents: buildAgents(source, time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC))}
	var got []string
	for _, node := range state.agentList() {
		got = append(got, node.record.Key+" "+node.status)
	}
	want := []string{"coordinator-g8 idle", "manager-qmeter-g4 working", "manager-stenographer-g8 stopped"}
	if !slices.Equal(got, want) {
		t.Fatalf("Health agents = %v; want %v", got, want)
	}
}
