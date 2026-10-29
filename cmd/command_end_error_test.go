package cmd

import (
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func TestCommandEndCarriesTheErrorOfAFailure(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		command string
		want    string
	}{
		{"failed envelope", []string{"agent", "spawn", "--name", "a", "--model", "m"}, "agent spawn", "--harness"},
		{"usage error", []string{"project", "register", ""}, "project register", "PATH requires a value"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := requestTestHome(t)
			t.Setenv("HIVE_HOME", home)
			var output strings.Builder
			if err := ExecuteWithArgs(test.args, strings.NewReader(""), &output); err == nil {
				t.Fatalf("%v succeeded; want a failure", test.args)
			}
			var end map[string]any
			for _, event := range agenttest.CommandEvents(t, home, test.command) {
				if event["msg"] == "command end" {
					end = event
				}
			}
			text, _ := end["error"].(string)
			if end["outcome"] != "failed" || end["level"] != "ERROR" || !strings.Contains(text, test.want) {
				t.Fatalf("command end = %v; want outcome failed, level ERROR, and error with %q", end, test.want)
			}
		})
	}
}

func TestCommandEndHasNoErrorOnSuccess(t *testing.T) {
	home := requestTestHome(t)
	t.Setenv("HIVE_HOME", home)
	var output strings.Builder
	if err := ExecuteWithArgs([]string{"project", "list"}, strings.NewReader(""), &output); err != nil {
		t.Fatalf("project list: %v\n%s", err, output.String())
	}
	for _, event := range agenttest.CommandEvents(t, home, "project list") {
		if _, ok := event["error"]; ok && event["msg"] == "command end" {
			t.Fatalf("command end = %v; want no error", event)
		}
	}
}
