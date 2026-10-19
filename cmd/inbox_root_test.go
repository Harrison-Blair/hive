package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
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

func TestInboxAgentReadConsumesOwnRole(t *testing.T) {
	home := testHome(t)
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	manager := records.Agent{Key: "manager-hive-g1", Name: "manager-hive", Role: "manager", Project: "hive", PaneID: "w1:p1", Status: records.AgentLive, Generation: 1}
	if err := st.WriteIn([]string{"agents"}, manager.Key, manager); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteIn([]string{"claims"}, "manager-hive", records.Claim{Key: "manager-hive", Agent: manager.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
	if envelope := inbox.Append(home, inbox.AppendOptions{Role: "manager-hive", Kind: records.InboxMessage, Text: "owner message"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox", "--agent", "manager-hive", "--role", "manager-hive"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || !strings.Contains(stdout.String(), "owner message") {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want the owner message", status, stdout.String(), stderr.String())
	}
	found, err := inbox.All(home, "manager-hive")
	if err != nil || len(found) != 1 || found[0].ReadAt == "" {
		t.Fatalf("All() = %#v, %v; want one consumed record", found, err)
	}
}

func seedInboxCoordinator(t *testing.T, home string) {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	coordinator := records.Agent{Key: "coordinator-g1", Name: "coordinator", Role: "coordinator", PaneID: "w0:p1", Status: records.AgentLive, Generation: 1}
	if err := st.WriteIn([]string{"agents"}, coordinator.Key, coordinator); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteIn([]string{"claims"}, "coordinator", records.Claim{Key: "coordinator", Agent: coordinator.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestInboxAllInvalidKeyLogsReadFailure(t *testing.T) {
	home := testHome(t)
	seedInboxCoordinator(t, home)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox", "--agent", "coordinator", "--role", "../x", "--all"}, strings.NewReader(""), &stdout, &stderr)
	if status != 1 {
		t.Fatalf("status = %d, stderr = %q; want 1", status, stderr.String())
	}
	events := readRootLogEvents(t, home)
	found := false
	for _, event := range events {
		message, _ := event["error"].(string)
		if event["msg"] == "inbox operation failed" && strings.Contains(message, "invalid inbox key: ../x") {
			found = true
		}
		if event["msg"] == "inbox record consumed" {
			t.Fatalf("event = %#v; an invalid --all must not claim a consume", event)
		}
	}
	if !found {
		t.Fatalf("events = %#v; want the invalid --all read failure", events)
	}
}

func TestInboxConsumeLogKeepsStoredSender(t *testing.T) {
	home := testHome(t)
	if envelope := inbox.Append(home, inbox.AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "BODY-MUST-NOT-LOG"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stdout, stderr bytes.Buffer
	seedInboxCoordinator(t, home)
	status := executeWithStreams([]string{"--home", home, "inbox", "--agent", "coordinator", "--role", "coordinator"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	events := readRootLogEvents(t, home)
	event := findRootLogEvent(events, "inbox record consumed")
	if event == nil || event["role"] != "coordinator" || event["sequence"] != float64(1) || event["agent"] != "worker-x" || event["agent_id"] != "worker-x" {
		t.Fatalf("consume event = %#v; want role, sequence, and the stored sender", event)
	}
	for _, candidate := range events {
		data, _ := json.Marshal(candidate)
		if strings.Contains(string(data), "BODY-MUST-NOT-LOG") {
			t.Fatalf("log event contains message text: %s", data)
		}
	}
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
			seedInboxCoordinator(t, home)
			args := []string{"--home", home}
			if jsonMode {
				args = append(args, "--json")
			}
			args = append(args, "inbox", "--agent", "coordinator", "--role", "coordinator")
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
	seedInboxCoordinator(t, home)
	if envelope := inbox.Append(home, inbox.AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox", "--agent", "coordinator", "--role", "coordinator", "--all"}, strings.NewReader(""), &failWriter{message: "stdout failed"}, &stderr)
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

func TestInboxMissingRoleRequiresCaller(t *testing.T) {
	home := testHome(t)
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox"}, strings.NewReader(""), &bytes.Buffer{}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "caller unknown: pass --agent or run inside herdr") {
		t.Fatalf("stderr = %q; want the caller hint", stderr.String())
	}
}

func TestInboxExplicitEmptyRoleIsInvalidKey(t *testing.T) {
	home := testHome(t)
	seedInboxCoordinator(t, home)
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox", "--agent", "coordinator", "--role", ""}, strings.NewReader(""), &bytes.Buffer{}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "invalid inbox key:") {
		t.Fatalf("stderr = %q; want invalid inbox key", stderr.String())
	}
}

func TestInboxFailedStdoutKeepsMarksAndReportsAll(t *testing.T) {
	home := testHome(t)
	seedInboxCoordinator(t, home)
	if envelope := inbox.Append(home, inbox.AppendOptions{Role: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"--home", home, "inbox", "--agent", "coordinator", "--role", "coordinator"}, strings.NewReader(""), &failWriter{message: "stdout failed"}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "1") || !strings.Contains(stderr.String(), "hive inbox --role coordinator --all") {
		t.Fatalf("stderr = %q; want consumed sequence 1 and the --all command", stderr.String())
	}

	var stdout, nextStderr bytes.Buffer
	status = executeWithStreams([]string{"--home", home, "--json", "inbox", "--agent", "coordinator", "--role", "coordinator", "--all"}, strings.NewReader(""), &stdout, &nextStderr)
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
