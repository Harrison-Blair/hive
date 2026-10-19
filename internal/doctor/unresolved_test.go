package doctor

import (
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

func seedDoctorAgent(t *testing.T, home string, record records.Agent) {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	if err := registry.WriteIn([]string{"agents"}, record.Key, record); err != nil {
		t.Fatalf("write agent: %v", err)
	}
}

func TestDoctorUnresolved(t *testing.T) {
	home := newHome(t)
	check := checkUnresolvedAgents(home)
	if check.Name != "unresolved agents" || check.Status != StatusOK || check.Detail != "0" {
		t.Fatalf("check = %+v, want ok with 0", check)
	}
	if line := Render(Result{Checks: []Check{check}}); line != "ok unresolved agents: 0\n" {
		t.Fatalf("Render() = %q, want the ok line", line)
	}

	for index, status := range []string{records.AgentUnresolved, records.AgentUnresolved, records.AgentLive, records.AgentExited} {
		key := "agent-" + string(rune('a'+index))
		seedDoctorAgent(t, home, records.Agent{Key: key, Status: status})
	}
	check = checkUnresolvedAgents(home)
	if check.Status != StatusWarn || check.Detail != "2" {
		t.Fatalf("check = %+v, want warn with 2", check)
	}
	if line := Render(Result{Checks: []Check{check}}); line != "warn unresolved agents: 2\n" {
		t.Fatalf("Render() = %q, want the warn line", line)
	}
}
