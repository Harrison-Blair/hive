package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
)

type failWriter struct {
	message string
}

func (w *failWriter) Write([]byte) (int, error) {
	return 0, errors.New(w.message)
}

type shortWriter struct{}

func (w *shortWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	return len(data) - 1, nil
}

func TestInboxShortStdoutReturnsOneAndReportsActualIDs(t *testing.T) {
	for _, jsonMode := range []bool{false, true} {
		t.Run(map[bool]string{false: "text", true: "json"}[jsonMode], func(t *testing.T) {
			home := testHome(t)
			for _, text := range []string{"one", "two"} {
				if envelope := inbox.Append(home, inbox.AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: text}); envelope.Outcome != "ok" {
					t.Fatalf("Append() = %#v", envelope)
				}
			}
			args := []string{"--home", home}
			if jsonMode {
				args = append(args, "--json")
			}
			args = append(args, "inbox", "--role", "coordinator")
			var stderr bytes.Buffer
			status := executeWithStreams(args, strings.NewReader(""), &shortWriter{}, &stderr)
			if status != 1 {
				t.Fatalf("status = %d; want 1 after a short stdout write (stderr %q)", status, stderr.String())
			}
			for _, id := range []string{"consumed 1", "consumed 2", "hive inbox --role coordinator --all"} {
				if !strings.Contains(stderr.String(), id) {
					t.Fatalf("stderr = %q; want %q", stderr.String(), id)
				}
			}
			marked, err := inbox.All(home, "coordinator")
			if err != nil || len(marked) != 2 {
				t.Fatalf("All() = %#v, %v; want two marked records", marked, err)
			}
			for _, record := range marked {
				if record.ReadAt == "" {
					t.Fatalf("record %d has no read_at; want the completed mark", record.Seq)
				}
			}
		})
	}
}

func TestInboxAllFailedStdoutClaimsNoConsumption(t *testing.T) {
	home := testHome(t)
	if envelope := inbox.Append(home, inbox.AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox", "--role", "coordinator", "--all"}, strings.NewReader(""), &failWriter{message: "stdout failed"}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1", status)
	}
	if strings.Contains(stderr.String(), "consumed") {
		t.Fatalf("stderr = %q; want no consumption claim after --all", stderr.String())
	}
	records, err := inbox.All(home, "coordinator")
	if err != nil || len(records) != 1 || records[0].ReadAt != "" {
		t.Fatalf("All() = %#v, %v; want one unread record", records, err)
	}
}

func TestInboxMissingRoleAsksForRole(t *testing.T) {
	home := testHome(t)
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox"}, strings.NewReader(""), &bytes.Buffer{}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "pass --role") {
		t.Fatalf("stderr = %q; want pass --role", stderr.String())
	}
	if strings.Contains(stderr.String(), "invalid inbox key") {
		t.Fatalf("stderr = %q; want the absent-role message", stderr.String())
	}
}

func TestInboxExplicitEmptyRoleIsInvalidKey(t *testing.T) {
	home := testHome(t)
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox", "--role", ""}, strings.NewReader(""), &bytes.Buffer{}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "invalid inbox key:") {
		t.Fatalf("stderr = %q; want invalid inbox key", stderr.String())
	}
}

func TestInboxFailedStdoutKeepsMarksAndReportsAll(t *testing.T) {
	home := testHome(t)
	if envelope := inbox.Append(home, inbox.AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox", "--role", "coordinator"}, strings.NewReader(""), &failWriter{message: "stdout failed"}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "1") || !strings.Contains(stderr.String(), "hive inbox --role coordinator --all") {
		t.Fatalf("stderr = %q; want consumed sequence 1 and the --all command", stderr.String())
	}

	var stdout, nextStderr bytes.Buffer
	status = executeWithStreams([]string{"--home", home, "--json", "inbox", "--role", "coordinator", "--all"}, strings.NewReader(""), &stdout, &nextStderr)
	if status != 0 || nextStderr.Len() != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, nextStderr.String())
	}
	var envelope struct {
		Outcome string `json:"outcome"`
		Result  struct {
			Records []records.Inbox `json:"records"`
		} `json:"result"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v; output %q", err, stdout.String())
	}
	if envelope.Outcome != "ok" || len(envelope.Result.Records) != 1 || envelope.Result.Records[0].ReadAt == "" {
		t.Fatalf("envelope = %#v; want one marked record", envelope)
	}
	if envelope.Result.Records[0].Seq != 1 {
		t.Fatalf("seq = %d; want 1", envelope.Result.Records[0].Seq)
	}
}
