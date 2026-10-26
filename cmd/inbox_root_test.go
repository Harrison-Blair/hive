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
	manager := records.Agent{Key: "manager-hive-g1", Name: "manager-hive", Profile: records.ProfilePlanner, Owns: "hive", Project: "hive", PaneID: "w1:p1", Status: records.AgentLive, Generation: 1}
	if err := st.WriteIn([]string{"agents"}, manager.Key, manager); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteIn([]string{"claims"}, "owner-hive", records.Claim{Key: "owner-hive", Agent: manager.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
	if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "manager-hive", Kind: records.InboxMessage, Text: "owner message"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox", "--agent", "manager-hive", "--inbox", "manager-hive"}, strings.NewReader(""), &stdout, &stderr)
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
	coordinator := records.Agent{Key: "coordinator-g1", Name: "coordinator", Kind: records.AgentKindCoordinator, Profile: records.ProfileCoordinator, PaneID: "w0:p1", Status: records.AgentLive, Generation: 1}
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
	status := executeWithStreams([]string{"inbox", "--agent", "coordinator", "--inbox", "../x", "--all"}, strings.NewReader(""), &stdout, &stderr)
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
	if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "BODY-MUST-NOT-LOG"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stdout, stderr bytes.Buffer
	seedInboxCoordinator(t, home)
	status := executeWithStreams([]string{"inbox", "--agent", "coordinator", "--inbox", "coordinator"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want 0", status, stderr.String())
	}
	events := readRootLogEvents(t, home)
	event := findRootLogEvent(events, "inbox record consumed")
	if event == nil || event["inbox"] != "coordinator" || event["sequence"] != float64(1) || event["agent"] != "worker-x" || event["agent_id"] != "worker-x" {
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
				if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: text}); envelope.Outcome != "ok" {
					t.Fatalf("Append() = %#v", envelope)
				}
			}
			seedInboxCoordinator(t, home)
			args := []string{}
			if jsonMode {
				args = append(args, "--json")
			}
			args = append(args, "inbox", "--agent", "coordinator", "--inbox", "coordinator")
			var stderr bytes.Buffer
			status := executeWithStreams(args, strings.NewReader(""), &shortWriter{}, &stderr)
			if status != 1 {
				t.Fatalf("status = %d; want 1 after a short stdout write (stderr %q)", status, stderr.String())
			}
			for _, id := range []string{"consumed 1", "consumed 2", "hive inbox --inbox coordinator --all"} {
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
	if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox", "--agent", "coordinator", "--inbox", "coordinator", "--all"}, strings.NewReader(""), &failWriter{message: "stdout failed"}, &stderr)
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
	testHome(t)
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox"}, strings.NewReader(""), &bytes.Buffer{}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "caller unknown: pass --agent or run inside herdr") {
		t.Fatalf("stderr = %q; want the caller hint", stderr.String())
	}
}

func TestInboxExplicitEmptyInboxIsInvalidKey(t *testing.T) {
	home := testHome(t)
	seedInboxCoordinator(t, home)
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox", "--agent", "coordinator", "--inbox", ""}, strings.NewReader(""), &bytes.Buffer{}, &stderr)
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
	if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "coordinator", Kind: records.InboxMessage, Agent: "worker-x", Text: "hello"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox", "--agent", "coordinator", "--inbox", "coordinator"}, strings.NewReader(""), &failWriter{message: "stdout failed"}, &stderr)
	if status != 1 {
		t.Fatalf("status = %d; want 1 (stderr %q)", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "1") || !strings.Contains(stderr.String(), "hive inbox --inbox coordinator --all") {
		t.Fatalf("stderr = %q; want consumed sequence 1 and the --all command", stderr.String())
	}

	var stdout, nextStderr bytes.Buffer
	status = executeWithStreams([]string{"--json", "inbox", "--agent", "coordinator", "--inbox", "coordinator", "--all"}, strings.NewReader(""), &stdout, &nextStderr)
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

func seedInboxPaneAgents(t *testing.T, home string) {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, agent := range []records.Agent{
		{Key: "agent-a", Name: "agent-a", Profile: records.ProfileBuilder, Project: "hive", Task: "1-a", PaneID: "w1:p1", Status: records.AgentLive, Generation: 1},
		{Key: "agent-b", Name: "agent-b", Profile: records.ProfileBuilder, Project: "hive", Task: "2-b", PaneID: "w2:p1", Status: records.AgentLive, Generation: 1},
	} {
		if err := st.WriteIn([]string{"agents"}, agent.Key, agent); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInboxSendAgentOverrideInPaneFailsWithoutRecord(t *testing.T) {
	home := testHome(t)
	seedInboxPaneAgents(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox", "send", "--to", "manager-hive", "--agent", "agent-b", "forged"}, strings.NewReader(""), &stdout, &stderr)
	if status == 0 {
		t.Fatalf("status = 0, stdout = %q; want a refusal", stdout.String())
	}
	if !strings.Contains(stderr.String(), "--agent agent-b is not the caller agent-a") {
		t.Fatalf("stderr = %q; want the --agent refusal", stderr.String())
	}
	found, err := inbox.All(home, "manager-hive")
	if err != nil || len(found) != 0 {
		t.Fatalf("All() = %#v, %v; want no record", found, err)
	}
}

func TestInboxOverridesInPaneFailAndChangeNothing(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"read agent", []string{"inbox", "--agent", "agent-b"}, "--agent agent-b is not the caller agent-a"},
		{"read pane", []string{"inbox", "--pane", "w2:p1"}, "--pane w2:p1 is not the caller pane w1:p1"},
		{"send pane", []string{"inbox", "send", "--to", "agent-b", "--pane", "w2:p1", "forged"}, "--pane w2:p1 is not the caller pane w1:p1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := testHome(t)
			seedInboxPaneAgents(t, home)
			if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "agent-b", Kind: records.InboxMessage, Text: "for b"}); envelope.Outcome != "ok" {
				t.Fatalf("Append() = %#v", envelope)
			}
			t.Setenv("HERDR_PANE_ID", "w1:p1")
			var stdout, stderr bytes.Buffer
			status := executeWithStreams(tc.args, strings.NewReader(""), &stdout, &stderr)
			if status == 0 || !strings.Contains(stderr.String(), tc.want) {
				t.Fatalf("status = %d, stdout = %q, stderr = %q; want a refusal with %q", status, stdout.String(), stderr.String(), tc.want)
			}
			found, err := inbox.All(home, "agent-b")
			if err != nil || len(found) != 1 || found[0].ReadAt != "" {
				t.Fatalf("All(agent-b) = %#v, %v; want one unread record", found, err)
			}
		})
	}
}

func TestInboxAgentInUnregisteredPaneFails(t *testing.T) {
	home := testHome(t)
	seedInboxPaneAgents(t, home)
	t.Setenv("HERDR_PANE_ID", "w9:p1")
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox", "send", "--to", "manager-hive", "--agent", "agent-a", "forged"}, strings.NewReader(""), &stdout, &stderr)
	if status == 0 || !strings.Contains(stderr.String(), "pane w9:p1 has no live agent") {
		t.Fatalf("status = %d, stderr = %q; want the no-agent refusal", status, stderr.String())
	}
	if found, err := inbox.All(home, "manager-hive"); err != nil || len(found) != 0 {
		t.Fatalf("All() = %#v, %v; want no record", found, err)
	}
}

func TestInboxSendBetweenAgentInboxesInPanes(t *testing.T) {
	home := testHome(t)
	seedInboxPaneAgents(t, home)
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	var stdout, stderr bytes.Buffer
	if status := executeWithStreams([]string{"inbox", "send", "--to", "agent-b", "--agent", "agent-a", "hello b"}, strings.NewReader(""), &stdout, &stderr); status != 0 {
		t.Fatalf("send status = %d, stderr = %q; want 0", status, stderr.String())
	}
	t.Setenv("HERDR_PANE_ID", "w2:p1")
	stdout.Reset()
	stderr.Reset()
	if status := executeWithStreams([]string{"inbox"}, strings.NewReader(""), &stdout, &stderr); status != 0 || !strings.Contains(stdout.String(), "hello b") {
		t.Fatalf("read status = %d, stdout = %q, stderr = %q; want the record", status, stdout.String(), stderr.String())
	}
	found, err := inbox.All(home, "agent-b")
	if err != nil || len(found) != 1 || found[0].Agent != "agent-a" || found[0].ReadAt == "" {
		t.Fatalf("All(agent-b) = %#v, %v; want one consumed record from agent-a", found, err)
	}
}

func TestInboxCoordinatorReadsAgentInboxWithoutConsume(t *testing.T) {
	home := testHome(t)
	seedInboxCoordinator(t, home)
	seedInboxPaneAgents(t, home)
	if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "agent-b", Kind: records.InboxMessage, Text: "for b"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	t.Setenv("HERDR_PANE_ID", "w0:p1")
	var stdout, stderr bytes.Buffer
	if status := executeWithStreams([]string{"inbox", "--inbox", "agent-b"}, strings.NewReader(""), &stdout, &stderr); status != 0 || !strings.Contains(stdout.String(), "for b") {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want the record", status, stdout.String(), stderr.String())
	}
	found, err := inbox.All(home, "agent-b")
	if err != nil || len(found) != 1 || found[0].ReadAt != "" {
		t.Fatalf("All(agent-b) = %#v, %v; want one unread record", found, err)
	}
}

func TestInboxAgentCannotReadOtherAgentInbox(t *testing.T) {
	home := testHome(t)
	seedInboxPaneAgents(t, home)
	if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "agent-b", Kind: records.InboxMessage, Text: "for b"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	t.Setenv("HERDR_PANE_ID", "w1:p1")
	for _, args := range [][]string{{"inbox", "--inbox", "agent-b"}, {"inbox", "--inbox", "agent-b", "--all"}} {
		var stdout, stderr bytes.Buffer
		if status := executeWithStreams(args, strings.NewReader(""), &stdout, &stderr); status == 0 || !strings.Contains(stderr.String(), "inbox agent-b belongs to the agent agent-b") {
			t.Fatalf("%v: status = %d, stdout = %q, stderr = %q; want the ownership refusal", args, status, stdout.String(), stderr.String())
		}
	}
	found, err := inbox.All(home, "agent-b")
	if err != nil || len(found) != 1 || found[0].ReadAt != "" {
		t.Fatalf("All(agent-b) = %#v, %v; want one unread record", found, err)
	}
}

func seedProjectOwner(t *testing.T, home string) {
	t.Helper()
	st, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	owner := records.Agent{
		Key: "owner-hive-g1", Name: "manager-hive", Profile: records.ProfilePlanner, Owns: "hive", Project: "hive", PaneID: "w1:p1", Status: records.AgentLive, Generation: 1,
	}
	if err := st.WriteIn([]string{"agents"}, owner.Key, owner); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteIn([]string{"claims"}, "owner-hive", records.Claim{Key: "owner-hive", Agent: owner.Key, Generation: 1}); err != nil {
		t.Fatal(err)
	}
}

// TestInboxProjectAddressSendAndRead proves AC1 through the CLI: send to
// project:hive writes to the owner, and --inbox project:hive reads at read
// time.
func TestInboxProjectAddressSendAndRead(t *testing.T) {
	home := testHome(t)
	seedProjectOwner(t, home)
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox", "send", "--to", "project:hive", "--agent", "worker-x", "hello"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q; want send to the owner", status, stderr.String())
	}
	found, err := inbox.All(home, "manager-hive")
	if err != nil || len(found) != 1 || found[0].Text != "hello" || found[0].Agent != "worker-x" {
		t.Fatalf("All(manager-hive) = %#v, %v; want the sent record", found, err)
	}

	stdout.Reset()
	stderr.Reset()
	status = executeWithStreams([]string{"inbox", "--agent", "manager-hive", "--inbox", "project:hive"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || !strings.Contains(stdout.String(), "hello") {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want the owner read", status, stdout.String(), stderr.String())
	}
	found, err = inbox.All(home, "manager-hive")
	if err != nil || len(found) != 1 || found[0].ReadAt == "" {
		t.Fatalf("All(manager-hive) = %#v, %v; want one consumed record", found, err)
	}
}

// TestInboxProjectAddressFallsBackToCoordinator proves the read of a project
// with no live owner uses the coordinator inbox.
func TestInboxProjectAddressFallsBackToCoordinator(t *testing.T) {
	home := testHome(t)
	seedInboxCoordinator(t, home)
	if envelope := inbox.Append(home, inbox.AppendOptions{Inbox: "coordinator", Kind: records.InboxMessage, Text: "for coordinator"}); envelope.Outcome != "ok" {
		t.Fatalf("Append() = %#v", envelope)
	}
	var stdout, stderr bytes.Buffer
	status := executeWithStreams([]string{"inbox", "--agent", "coordinator", "--inbox", "project:missing"}, strings.NewReader(""), &stdout, &stderr)
	if status != 0 || !strings.Contains(stdout.String(), "for coordinator") {
		t.Fatalf("status = %d, stdout = %q, stderr = %q; want the coordinator read", status, stdout.String(), stderr.String())
	}
}
