package agent

import (
	"context"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/records"
)

const fix02FirstRef = `{"source":"herdr:claude","agent":"claude","kind":"id","value":"s"}`

func fix02AgentJSON(status, session string) string {
	if session == "" {
		return agentJSON("w1:p1", "term_a", "w1", "w1:t1", status, "t10-coord-g1")
	}
	return agentJSONSession("w1:p1", "term_a", "w1", "w1:t1", status, "t10-coord-g1", "herdr:claude", "id", session)
}

func fix02PaneJSON(session string) string {
	body := paneInfoJSON("w1:p1", "term_a", "w1", "w1:t1", "/tmp/fixture")
	if session == "" {
		return body
	}
	return body[:len(body)-1] + `,"agent_session":{"source":"herdr:claude","agent":"claude","kind":"id","value":"` + session + `"}}`
}

// TestFix02Session checks the ordered session observations of Spawn.
// Each expected status and reference is a literal value.
func TestFix02Session(t *testing.T) {
	cases := []struct {
		name     string
		start    string
		pane     string
		wait     string
		get      string
		paneFail bool
		getFail  bool
		process  bool
		want     string
		wantRef  string
	}{
		{name: "no session", want: records.AgentLive},
		{name: "first at start", start: "s", pane: "s", wait: "s", get: "s", want: records.AgentLive, wantRef: fix02FirstRef},
		{name: "first at pane", pane: "s", wait: "s", get: "s", want: records.AgentLive, wantRef: fix02FirstRef},
		{name: "first at wait", wait: "s", get: "s", want: records.AgentLive, wantRef: fix02FirstRef},
		{name: "first at get", get: "s", want: records.AgentLive, wantRef: fix02FirstRef},
		{name: "pane missing after start", start: "s", wait: "s", get: "s", want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "wait missing after start and pane", start: "s", pane: "s", get: "s", want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "get missing after start pane wait", start: "s", pane: "s", wait: "s", want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "get missing after first at wait", wait: "s", want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "pane differs from start", start: "s", pane: "other", wait: "other", get: "other", want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "wait differs from pane", pane: "s", wait: "other", get: "other", want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "get differs from wait", wait: "s", get: "other", want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "pane read fails after start", start: "s", paneFail: true, want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "process fails after first at wait", wait: "s", get: "s", process: true, want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "process fails after first at get", get: "s", process: true, want: records.AgentUnresolved, wantRef: fix02FirstRef},
		{name: "get error keeps first", start: "s", pane: "s", wait: "s", getFail: true, want: records.AgentLive, wantRef: fix02FirstRef},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			sleep := startSleep(t)
			fake := newFakeHerdr(t)
			fake.result("agent.start", `{"type":"agent_started","agent":`+fix02AgentJSON("unknown", test.start)+`,"argv":["claude"]}`)
			if test.paneFail {
				fake.failure("pane.get", "pane_not_found", "pane w1:p1 not found")
			} else {
				fake.result("pane.get", `{"type":"pane_info","pane":`+fix02PaneJSON(test.pane)+`}`)
			}
			fake.result("agent.wait", `{"type":"agent_info","agent":`+fix02AgentJSON("idle", test.wait)+`}`)
			if test.getFail {
				fake.failure("agent.get", "agent_not_found", "agent target w1:p1 not found")
			} else {
				fake.result("agent.get", `{"type":"agent_info","agent":`+fix02AgentJSON("idle", test.get)+`}`)
			}
			if test.process {
				fake.failure("pane.process_info", "pane_not_found", "pane w1:p1 not found")
			} else {
				fake.result("pane.process_info", `{"type":"pane_process_info","process_info":`+processInfoJSON("w1:p1", processEntry(sleep.pid, "claude", "claude"))+`}`)
			}

			envelope := Spawn(context.Background(), home, baseSpawnOptions("t10-coord"))
			record := readAgent(t, home, "t10-coord-g1")
			if record.Status != test.want {
				t.Errorf("record status = %q; want %q (envelope %#v)", record.Status, test.want, envelope)
			}
			if record.AgentSessionRef != test.wantRef {
				t.Errorf("record session = %q; want %q", record.AgentSessionRef, test.wantRef)
			}
			result, ok := envelope.Result.(SpawnResult)
			if !ok {
				t.Fatalf("result type = %T; want SpawnResult", envelope.Result)
			}
			if result.Agent.AgentSessionRef != test.wantRef {
				t.Errorf("result session = %q; want %q", result.Agent.AgentSessionRef, test.wantRef)
			}
			if test.want == records.AgentLive && envelope.Outcome != "ok" {
				t.Errorf("outcome = %q; want ok", envelope.Outcome)
			}
			if test.want != records.AgentLive && envelope.Outcome == "ok" {
				t.Errorf("outcome = %q; want a partial outcome", envelope.Outcome)
			}
			if !hasEffect(envelope.Effects, "agent", "started") {
				t.Errorf("effects = %#v; want the started effect", envelope.Effects)
			}
			if len(fake.recorded("pane.close")) != 0 {
				t.Errorf("pane.close ran %d times; want none", len(fake.recorded("pane.close")))
			}
		})
	}
}
