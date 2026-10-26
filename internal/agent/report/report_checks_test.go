package report

import (
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

// TestReportRefusesCallerWithoutParent proves the exact D8 report refusal.
func TestReportRefusesCallerWithoutParent(t *testing.T) {
	home := agenttest.NewHome(t)
	agenttest.WriteRecord(t, home, records.Agent{
		Key: "rev", Name: "rev", Profile: records.ProfileGeneral,
		Status: records.AgentLive, PaneID: "w1:p2", Generation: 1,
	})
	before := readAgentRecord(t, home, "rev")
	envelope := Run(home, Options{HerdrPane: "w1:p2", Input: strings.NewReader("text")})
	if envelope.Outcome != "failed" || envelope.Error != "agent report needs a parent: rev has none" {
		t.Fatalf("Run() = %#v; want the exact D8 refusal", envelope)
	}
	for _, key := range []string{"coordinator", "rev"} {
		if found := inboxRecords(t, home, key); len(found) != 0 {
			t.Fatalf("%s inbox = %#v; want none", key, found)
		}
	}
	if after := readAgentRecord(t, home, "rev"); after.ReportedAt != before.ReportedAt {
		t.Fatalf("reported_at = %q; want %q", after.ReportedAt, before.ReportedAt)
	}
}

// TestReportAllowsCallerWithParentWithoutRole proves that the parent fact is
// the rule, and the role mirror is not.
func TestReportAllowsCallerWithParentWithoutRole(t *testing.T) {
	home := agenttest.NewHome(t)
	agenttest.WriteRecord(t, home, records.Agent{Key: "boss", Name: "boss", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
	agenttest.WriteRecord(t, home, records.Agent{
		Key: "rev", Name: "rev", Profile: records.ProfileGeneral, Parent: "boss",
		Status: records.AgentLive, PaneID: "w1:p2", Generation: 1,
	})
	envelope := Run(home, Options{HerdrPane: "w1:p2", Input: strings.NewReader("text")})
	if envelope.Outcome != "ok" {
		t.Fatalf("Run() = %#v; want ok for a caller with a parent", envelope)
	}
	found := inboxRecords(t, home, "boss")
	if len(found) != 1 || found[0].Agent != "rev" || found[0].Text != "text" {
		t.Fatalf("boss inbox = %#v; want one report from rev", found)
	}
}
