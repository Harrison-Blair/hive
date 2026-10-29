package logs

import "testing"

func TestFilterMatchFields(t *testing.T) {
	event := Event{Fields: map[string]any{
		"level":      "WARN",
		"component":  "cli",
		"agent_id":   "worker-one",
		"request_id": "3f9a1c2e-17",
	}}
	cases := []struct {
		name   string
		filter Filter
		want   bool
	}{
		{"empty", Filter{}, true},
		{"component match", Filter{Component: "cli"}, true},
		{"component miss", Filter{Component: "tui"}, false},
		{"agent match", Filter{AgentID: "worker-one"}, true},
		{"agent miss", Filter{AgentID: "worker-two"}, false},
		{"request match", Filter{RequestID: "3f9a1c2e-17"}, true},
		{"request miss", Filter{RequestID: "other"}, false},
		{"level match", Filter{Level: "warn"}, true},
		{"level miss", Filter{Level: "error"}, false},
		{"all fields match", Filter{Level: "info", Component: "cli", AgentID: "worker-one", RequestID: "3f9a1c2e-17"}, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := test.filter.Match(event); got != test.want {
				t.Fatalf("Match() = %v; want %v", got, test.want)
			}
		})
	}
}

func TestFilterAgentMatchesCallerOrTarget(t *testing.T) {
	event := Event{Fields: map[string]any{"agent_id": "boss", "target": "kid"}}
	for agent, want := range map[string]bool{"boss": true, "kid": true, "other": false} {
		if got := (Filter{AgentID: agent}).Match(event); got != want {
			t.Errorf("Match(agent %q) = %v; want %v", agent, got, want)
		}
	}
}

func TestFilterLevelKeepsNamedAndAbove(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	cases := []struct {
		filter Filter
		want   []bool
	}{
		{Filter{Level: "debug"}, []bool{true, true, true, true}},
		{Filter{Level: "info"}, []bool{false, true, true, true}},
		{Filter{Level: "warn"}, []bool{false, false, true, true}},
		{Filter{Level: "error"}, []bool{false, false, false, true}},
	}
	for _, test := range cases {
		for index, level := range levels {
			event := Event{Fields: map[string]any{"level": level}}
			if got := test.filter.Match(event); got != test.want[index] {
				t.Errorf("filter %q matched %s = %v; want %v", test.filter.Level, level, got, test.want[index])
			}
		}
	}
}

func TestFilterMalformedOnlyMatchesEmptyFilter(t *testing.T) {
	malformed := Event{malformed: true}
	if !(Filter{}).Match(malformed) {
		t.Fatal("empty filter did not match a malformed line")
	}
	if (Filter{Level: "error"}).Match(malformed) {
		t.Fatal("filtered Match() matched a malformed line")
	}
}
