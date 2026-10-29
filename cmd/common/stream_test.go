package common

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/Harrison-Blair/hive/internal/agent/agenttest"
	"github.com/spf13/cobra"
)

func TestStreamLogsTheOutcomeOfTheBody(t *testing.T) {
	cases := []struct {
		name       string
		bodyErr    error
		wantStatus any
		outcome    string
		errText    string
	}{
		{"ok", nil, nil, "ok", ""},
		{"usage", &UsageError{Message: "bad value"}, &UsageError{Message: "bad value"}, "failed", "bad value"},
		{"failure", errors.New("read failed"), &StatusError{Status: 1}, "failed", "read failed"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := agenttest.Home(t)
			t.Setenv("HERDR_PANE_ID", "")
			ran := ""
			err := Stream(&cobra.Command{}, "stream test", func() (string, error) { return home, nil },
				func(string, error) error { t.Fatal("fail called"); return nil },
				func(got string, _ *slog.Logger) error { ran = got; return test.bodyErr })
			if ran != home {
				t.Fatalf("body home = %q; want %q", ran, home)
			}
			switch want := test.wantStatus.(type) {
			case nil:
				if err != nil {
					t.Fatalf("Stream() = %v; want nil", err)
				}
			case *UsageError:
				var usage *UsageError
				if !errors.As(err, &usage) || usage.Message != want.Message {
					t.Fatalf("Stream() = %v; want usage %q", err, want.Message)
				}
			case *StatusError:
				var status *StatusError
				if !errors.As(err, &status) || status.Status != want.Status {
					t.Fatalf("Stream() = %v; want status %d", err, want.Status)
				}
			}
			events := agenttest.CommandEvents(t, home, "stream test")
			if len(events) != 2 || events[0]["msg"] != "command start" || events[1]["msg"] != "command end" {
				t.Fatalf("events = %v; want command start and command end", events)
			}
			end := events[1]
			text, hasError := end["error"].(string)
			if end["outcome"] != test.outcome || end["agent_id"] != "user" || end["request_id"] != events[0]["request_id"] || end["request_id"] == nil || text != test.errText || hasError != (test.errText != "") {
				t.Fatalf("command end = %v; want outcome %q, error %q, agent_id user, and the start request_id", end, test.outcome, test.errText)
			}
		})
	}
}

func TestStreamRendersASetupFailureWithoutRunningTheBody(t *testing.T) {
	var gotPhase string
	err := Stream(&cobra.Command{}, "stream test", func() (string, error) { return "", errors.New("no home") },
		func(phase string, err error) error {
			gotPhase = phase + ": " + err.Error()
			return &StatusError{Status: 1}
		},
		func(string, *slog.Logger) error { t.Fatal("body ran"); return nil })
	var status *StatusError
	if gotPhase != "home: no home" || !errors.As(err, &status) || status.Status != 1 {
		t.Fatalf("fail got %q, Stream() = %v; want home: no home and status 1", gotPhase, err)
	}
}
