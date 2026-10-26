package send

import (
	"context"
	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"strings"
	"testing"
)

func TestPromptLogBodyStaysBounded(t *testing.T) {
	home := agenttest.NewHome(t)
	agenttest.WriteRecord(t, home, agenttest.LiveRecord("worker-hive-1", "worker-hive-1"))
	fake := agenttest.NewFakeHerdr(t)
	fake.Result("agent.prompt", `{"type":"agent_prompted","agent":`+agenttest.AgentJSON("w1:p1", "term_a", "w1", "w1:t1", "idle", "worker-hive-1")+`}`)
	long := "a" + strings.Repeat("é", 3000) + "END"
	if envelope := Run(context.Background(), home, Options{Agent: "worker-hive-1", Text: long}); envelope.Outcome != "ok" {
		t.Fatalf("envelope = %#v", envelope)
	}
	event := agenttest.FindLogEvent(t, agenttest.LogEvents(t, home), "agent prompt")
	agenttest.CheckBody(t, event, long)
	if event["command"] != "agent send" {
		t.Fatalf("command = %#v", event["command"])
	}
}
