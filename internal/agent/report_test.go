package agent

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	inboxlib "github.com/Harrison-Blair/hive/internal/lib/inbox"
	"github.com/Harrison-Blair/hive/internal/lib/output"
	"github.com/Harrison-Blair/hive/internal/lib/records"
	"github.com/Harrison-Blair/hive/internal/lib/store"
)

// seedReport writes a live parent boss in pane w1:p1 and a live general agent
// rev in pane w1:p2 whose parent is boss.
func seedReport(t *testing.T) string {
	t.Helper()
	home := testHome(t)
	writeRecord(t, home, records.Agent{Key: "boss", Name: "boss", Profile: records.ProfileGeneral, Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
	writeRecord(t, home, records.Agent{Key: "rev", Name: "rev", Profile: records.ProfileGeneral, Status: records.AgentLive, PaneID: "w1:p2", Parent: "boss", Generation: 1})
	return home
}

func readAgentRecord(t *testing.T, home, key string) records.Agent {
	t.Helper()
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var record records.Agent
	if err := registry.ReadIn([]string{"agents"}, key, &record); err != nil {
		t.Fatal(err)
	}
	return record
}

func inboxRecords(t *testing.T, home, key string) []records.Inbox {
	t.Helper()
	found, err := inboxlib.All(home, key)
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func TestReportStdinWritesOneParentInboxRecord(t *testing.T) {
	home := seedReport(t)
	envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", Input: strings.NewReader("the plan is sound")})
	if envelope.Outcome != "ok" {
		t.Fatalf("Report() = %#v; want ok", envelope)
	}
	found := inboxRecords(t, home, "boss")
	if len(found) != 1 || found[0].Kind != "report" || found[0].Text != "the plan is sound" || found[0].Agent != "rev" || found[0].Path != "" {
		t.Fatalf("boss inbox = %#v; want one report record from rev", found)
	}
}

func TestReportSetsReportedAtAndNotFinished(t *testing.T) {
	home := seedReport(t)
	if envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", Input: strings.NewReader("one")}); envelope.Outcome != "ok" {
		t.Fatalf("Report() = %#v; want ok", envelope)
	}
	record := readAgentRecord(t, home, "rev")
	if record.ReportedAt == "" || record.Finished || record.Status != records.AgentLive {
		t.Fatalf("record = %#v; want reported_at set, not finished, live", record)
	}
}

func TestReportTwiceWritesTwoRecords(t *testing.T) {
	home := seedReport(t)
	for _, text := range []string{"first", "second"} {
		if envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", Input: strings.NewReader(text)}); envelope.Outcome != "ok" {
			t.Fatalf("Report(%s) = %#v; want ok", text, envelope)
		}
	}
	found := inboxRecords(t, home, "boss")
	if len(found) != 2 || found[0].Text != "first" || found[1].Text != "second" {
		t.Fatalf("boss inbox = %#v; want two records", found)
	}
}

func TestReportUsesTheParentFactForEachFactShape(t *testing.T) {
	cases := []struct {
		name    string
		profile string
		kind    string
		owns    string
		claim   string
	}{
		{"builder", records.ProfileBuilder, "", "", ""},
		{"verifier", records.ProfileVerifier, "", "", ""},
		{"planner", records.ProfilePlanner, "", "hive", "owner-hive"},
		{"coordinator", records.ProfileCoordinator, records.AgentKindCoordinator, "", "coordinator"},
		{"supervisor", "", records.AgentKindSupervisor, "", "supervisor"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			home := testHome(t)
			record := records.Agent{Key: "x-1", Name: "x-1", Profile: test.profile, Kind: test.kind, Owns: test.owns, Project: "hive", Status: records.AgentLive, PaneID: "w1:p2", Parent: "boss", Generation: 1}
			writeRecord(t, home, record)
			writeRecord(t, home, records.Agent{Key: "boss", Name: "boss", Status: records.AgentLive, PaneID: "w1:p1", Generation: 1})
			if test.claim != "" {
				writeClaimRecord(t, home, records.Claim{Key: test.claim, Agent: record.Key, Generation: 1})
			}
			envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", Input: strings.NewReader("text")})
			if envelope.Outcome != "ok" {
				t.Fatalf("Report() = %#v; want ok for a caller with a parent", envelope)
			}
			if found := inboxRecords(t, home, "boss"); len(found) != 1 {
				t.Fatalf("boss inbox = %#v; want one record", found)
			}
			if readAgentRecord(t, home, "x-1").ReportedAt == "" {
				t.Fatal("reported_at is empty; want the report mark")
			}
		})
	}
}

func TestReportStdinLimits(t *testing.T) {
	for _, test := range []struct {
		name  string
		text  string
		error string
	}{
		{name: "4096 bytes", text: strings.Repeat("a", 4096)},
		{name: "4097 bytes", text: strings.Repeat("a", 4097), error: "use --file"},
		{name: "invalid UTF-8", text: "ok\xff", error: "not valid UTF-8"},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := seedReport(t)
			envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", Input: strings.NewReader(test.text)})
			found := inboxRecords(t, home, "boss")
			if test.error == "" {
				if envelope.Outcome != "ok" || len(found) != 1 || found[0].Text != test.text {
					t.Fatalf("Report() = %#v, inbox = %d records; want ok with the text", envelope, len(found))
				}
				return
			}
			if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), test.error) || len(found) != 0 {
				t.Fatalf("Report() = %#v, inbox = %d records; want %q and no record", envelope, len(found), test.error)
			}
		})
	}
}

func TestReportFileKeepsAbsolutePathAndCutsAtRuneBoundary(t *testing.T) {
	home := seedReport(t)
	directory := t.TempDir()
	// 4095 ASCII bytes and a 2-byte rune: the rune crosses the 4096 limit.
	body := strings.Repeat("a", 4095) + "é" + "tail"
	if err := os.WriteFile(filepath.Join(directory, "result.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(directory)
	envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", File: "result.md", FileSet: true})
	if envelope.Outcome != "ok" {
		t.Fatalf("Report() = %#v; want ok", envelope)
	}
	found := inboxRecords(t, home, "boss")
	if len(found) != 1 || found[0].Path != filepath.Join(directory, "result.md") || found[0].Text != strings.Repeat("a", 4095) {
		t.Fatalf("boss inbox = %#v; want the absolute path and 4095 bytes of text", found)
	}
}

func TestReportFileRefusals(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "bad.md"), []byte("ok\xff"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		file  string
		error string
	}{
		{name: "absent", file: filepath.Join(directory, "absent.md"), error: "no such file"},
		{name: "directory", file: directory, error: "not a regular file"},
		{name: "invalid UTF-8", file: filepath.Join(directory, "bad.md"), error: "not valid UTF-8"},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := seedReport(t)
			envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", File: test.file, FileSet: true})
			if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), test.error) {
				t.Fatalf("Report() = %#v; want %q", envelope, test.error)
			}
			if found := inboxRecords(t, home, "boss"); len(found) != 0 {
				t.Fatalf("boss inbox = %#v; want none", found)
			}
		})
	}
}

func TestReportRoutesWithTheParentRule(t *testing.T) {
	t.Run("stopped parent goes to its live successor", func(t *testing.T) {
		home := testHome(t)
		writeRecord(t, home, records.Agent{Key: "boss", Name: "boss", Profile: records.ProfileGeneral, Status: records.AgentStopped, Generation: 1})
		writeRecord(t, home, records.Agent{Key: "boss-2", Name: "boss-2", Profile: records.ProfileGeneral, Status: records.AgentLive, PaneID: "w1:p1", Replaces: "boss", Generation: 1})
		writeRecord(t, home, records.Agent{Key: "rev", Name: "rev", Profile: records.ProfileGeneral, Status: records.AgentLive, PaneID: "w1:p2", Parent: "boss", Generation: 1})
		if envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", Input: strings.NewReader("r")}); envelope.Outcome != "ok" {
			t.Fatalf("Report() = %#v; want ok", envelope)
		}
		if found := inboxRecords(t, home, "boss-2"); len(found) != 1 {
			t.Fatalf("boss-2 inbox = %#v; want one record", found)
		}
	})
	t.Run("a stopped parent with no successor goes to coordinator", func(t *testing.T) {
		home := testHome(t)
		writeRecord(t, home, records.Agent{Key: "boss", Name: "boss", Profile: records.ProfileGeneral, Status: records.AgentStopped, Generation: 1})
		writeRecord(t, home, records.Agent{Key: "rev", Name: "rev", Profile: records.ProfileGeneral, Status: records.AgentLive, PaneID: "w1:p2", Parent: "boss", Generation: 1})
		if envelope := Report(home, ReportOptions{HerdrPane: "w1:p2", Input: strings.NewReader("r")}); envelope.Outcome != "ok" {
			t.Fatalf("Report() = %#v; want ok", envelope)
		}
		if found := inboxRecords(t, home, "coordinator"); len(found) != 1 {
			t.Fatalf("coordinator inbox = %#v; want one record", found)
		}
	})
}

func TestReportCallerOverrides(t *testing.T) {
	for _, test := range []struct {
		name    string
		options ReportOptions
		error   string
	}{
		{name: "agent override in a pane", options: ReportOptions{HerdrPane: "w1:p2", Agent: "boss"}, error: "--agent boss is not the caller rev"},
		{name: "pane override in a pane", options: ReportOptions{HerdrPane: "w1:p2", Pane: "w1:p1"}, error: "--pane w1:p1 is not the caller pane w1:p2"},
		{name: "own agent in a pane", options: ReportOptions{HerdrPane: "w1:p2", Agent: "rev"}},
		{name: "own pane in a pane", options: ReportOptions{HerdrPane: "w1:p2", Pane: "w1:p2"}},
		{name: "agent outside herdr", options: ReportOptions{Agent: "rev"}},
		{name: "pane outside herdr", options: ReportOptions{Pane: "w1:p2"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := seedReport(t)
			options := test.options
			options.Input = strings.NewReader("r")
			envelope := Report(home, options)
			found := inboxRecords(t, home, "boss")
			if test.error == "" {
				if envelope.Outcome != "ok" || len(found) != 1 || found[0].Agent != "rev" {
					t.Fatalf("Report() = %#v, inbox = %#v; want one record from rev", envelope, found)
				}
				return
			}
			if envelope.Outcome != "failed" || !strings.Contains(envelope.Error.(string), test.error) || len(found) != 0 {
				t.Fatalf("Report() = %#v, inbox = %#v; want %q and no record", envelope, found, test.error)
			}
			if readAgentRecord(t, home, "rev").ReportedAt != "" || readAgentRecord(t, home, "boss").ReportedAt != "" {
				t.Fatal("reported_at is set; want empty")
			}
		})
	}
}

// failAgentWrite fails each write of an agent record.
type failAgentWrite struct {
	*store.LockedStore
}

func (seam failAgentWrite) WriteIn(collection []string, id string, value any) error {
	if len(collection) == 1 && collection[0] == "agents" {
		return errors.New("disk full")
	}
	return seam.LockedStore.WriteIn(collection, id, value)
}

func (seam failAgentWrite) UpdateIn(collection []string, id string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	if len(collection) == 1 && collection[0] == "agents" {
		return errors.New("disk full")
	}
	return seam.LockedStore.UpdateIn(collection, id, fn)
}

func TestReportAgentWriteFailureListsTheInboxRecord(t *testing.T) {
	home := seedReport(t)
	registry, err := store.Open(home)
	if err != nil {
		t.Fatal(err)
	}
	var envelope output.Envelope
	err = registry.WithLock(t.Context(), func(locked *store.LockedStore) error {
		envelope = reportHeld(failAgentWrite{locked}, ReportOptions{HerdrPane: "w1:p2"}, "result", "")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Outcome == "ok" || !strings.Contains(envelope.Error.(string), "disk full") {
		t.Fatalf("reportHeld() = %#v; want a failure", envelope)
	}
	listed := false
	for _, effect := range envelope.Effects {
		if effect.Resource == "inbox" && effect.Action == "created" && strings.Contains(effect.Path, "inbox/boss/") {
			listed = true
		}
	}
	if !listed {
		t.Fatalf("effects = %#v; want the inbox record", envelope.Effects)
	}
	if found := inboxRecords(t, home, "boss"); len(found) != 1 {
		t.Fatalf("boss inbox = %#v; want the first write", found)
	}
	if readAgentRecord(t, home, "rev").ReportedAt != "" {
		t.Fatal("reported_at is set; want empty")
	}
}
