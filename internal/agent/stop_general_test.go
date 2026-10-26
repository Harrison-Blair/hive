package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

func generalStopRecord(sleep sleepProcess, tabCreated bool) records.Agent {
	record := spawnedRecord(sleep)
	record.Key, record.Name = "x", "x"
	record.Kind, record.Profile = "", records.ProfileGeneral
	record.Parent = "boss"
	record.StartedByHive = true
	record.TabCreatedByHive = tabCreated
	return record
}

func TestStopGeneralAgentThatHiveDidNotStartClosesNothing(t *testing.T) {
	home := testHome(t)
	sleep := startSleep(t)
	record := generalStopRecord(sleep, true)
	record.StartedByHive = false
	writeStopRecord(t, home, record)
	fake := newFakeHerdr(t)
	_, envelope := Stop(context.Background(), home, record)
	if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), "hive did not start") {
		t.Fatalf("envelope = %#v; want a refusal", envelope)
	}
	if len(fake.requests) != 0 {
		t.Fatalf("herdr requests = %#v; want none", fake.requests)
	}
}

func TestStopGeneralAgentClosesItsPaneAndReportsTheTab(t *testing.T) {
	for _, test := range []struct {
		name       string
		tabCreated bool
		tabs       string
		want       *output.Effect
	}{
		{name: "hive tab closes with the pane", tabCreated: true, tabs: `{"tab_id":"w1:t2","workspace_id":"w1","number":2,"label":"y","focused":false,"pane_count":1,"agent_status":"unknown"}`, want: &output.Effect{Resource: "tab", Path: "w1:t1", Action: "closed"}},
		{name: "hive tab with another pane stays", tabCreated: true, tabs: `{"tab_id":"w1:t1","workspace_id":"w1","number":1,"label":"x","focused":false,"pane_count":1,"agent_status":"unknown"}`, want: &output.Effect{Resource: "tab", Path: "w1:t1", Action: "kept: 1 other panes"}},
		{name: "given pane", tabCreated: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			sleep := startSleep(t)
			record := generalStopRecord(sleep, test.tabCreated)
			writeStopRecord(t, home, record)
			fake := newFakeHerdr(t)
			scriptStop(t, fake, sleep, "", "")
			fake.result("workspace.get", `{"type":"workspace_info","workspace":`+workspaceJSON("w1", 1)+`}`)
			if test.tabCreated {
				fake.result("tab.list", `{"type":"tab_list","tabs":[`+test.tabs+`]}`)
			}
			_, envelope := Stop(context.Background(), home, record)
			if envelope.Outcome != "ok" {
				t.Fatalf("envelope = %#v; want ok", envelope)
			}
			closes := fake.recorded("pane.close")
			if len(closes) != 1 || closes[0].Params != `{"pane_id":"w1:p1"}` {
				t.Fatalf("pane.close = %#v; want only w1:p1", closes)
			}
			var tabEffect *output.Effect
			for index := range envelope.Effects {
				if envelope.Effects[index].Resource == "tab" {
					tabEffect = &envelope.Effects[index]
				}
			}
			if (test.want == nil) != (tabEffect == nil) || (test.want != nil && *tabEffect != *test.want) {
				t.Fatalf("tab effect = %#v; want %#v", tabEffect, test.want)
			}
			result, ok := envelope.Result.(StopResult)
			if !ok || result.Agent.Status != records.AgentStopped {
				t.Fatalf("result = %#v; want stopped", envelope.Result)
			}
		})
	}
}
