package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
)

func TestCobraUsageErrorsLogCommandStartAndEnd(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		command string
		want    string
	}{
		{"missing argument", []string{"agent", "send", "kid"}, "agent send", "accepts 2 arg(s), received 1"},
		{"extra argument", []string{"agent", "stop", "kid", "more"}, "agent stop", "accepts 1 arg(s), received 2"},
		{"unknown flag", []string{"agent", "spawn", "--bogus"}, "agent spawn", "unknown flag: --bogus"},
		{"unknown flag of logs", []string{"logs", "--task", "x"}, "logs", "unknown flag: --task"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := requestTestHome(t)
			t.Setenv("HIVE_HOME", home)
			t.Setenv("HERDR_PANE_ID", "")
			var output strings.Builder
			if err := ExecuteWithArgs(test.args, strings.NewReader(""), &output); err == nil || !strings.Contains(output.String(), test.want) {
				t.Fatalf("%v: err %v, output %q; want a usage error with %q", test.args, err, output.String(), test.want)
			}
			events := agenttest.CommandEvents(t, home, test.command)
			agenttest.CheckAgentFields(t, events, "user", "", "command start", "command end")
			end := events[len(events)-1]
			text, _ := end["error"].(string)
			if end["msg"] != "command end" || end["outcome"] != "failed" || end["level"] != "ERROR" || !strings.Contains(text, test.want) {
				t.Fatalf("last event = %v; want a failed command end with error %q", end, test.want)
			}
		})
	}
}

func TestVersionUsageErrorWritesNoLog(t *testing.T) {
	home := requestTestHome(t)
	t.Setenv("HIVE_HOME", home)
	var output strings.Builder
	if err := ExecuteWithArgs([]string{"version", "extra"}, strings.NewReader(""), &output); err == nil {
		t.Fatalf("version extra succeeded; want a usage error")
	}
	if _, err := os.Stat(filepath.Join(home, ".hive", "logs")); !os.IsNotExist(err) {
		t.Fatalf("logs directory exists after hive version (stat err %v); want no log", err)
	}
}
