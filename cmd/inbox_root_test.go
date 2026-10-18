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
