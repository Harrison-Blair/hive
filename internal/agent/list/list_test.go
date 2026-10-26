package list

import (
	"strings"
	"testing"

	agentlib "github.com/Harrison-Blair/hive/internal/agent"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func TestListAllShowsStartingStatus(t *testing.T) {
	home := agenttest.NewHome(t)
	agenttest.WriteRecord(t, home, records.Agent{
		Key: "agent-starting", Name: "agent-starting", Profile: records.ProfileBuilder, Project: "one", Task: "1-one",
		Status: records.AgentStarting, PaneID: "w1:p1", Generation: 1,
	})

	envelope := agentlib.List(home, agentlib.ListOptions{All: true})
	if envelope.Outcome != "ok" {
		t.Fatalf("List() = %+v, want ok", envelope)
	}
	text := Text(envelope)
	if !strings.Contains(text, "agent-starting\t1\tbuilder\tstarting\tone") {
		t.Fatalf("Text() = %q; want starting status", text)
	}
}
