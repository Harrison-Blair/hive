package guide

import (
	"os"
	"strings"
	"testing"
)

// readPart reads one embedded part from disk. The test uses the tracked files
// as an independent expectation for the embedded copies.
func readPart(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

// TestTextIsFlowsThenReference proves AC1: the guide prints flows.md, then the
// reference.
func TestTextIsFlowsThenReference(t *testing.T) {
	flows := readPart(t, "flows.md")
	reference := readPart(t, "reference.md")
	if flows == "" || reference == "" {
		t.Fatal("a guide part is empty")
	}
	want := flows + "\n" + reference
	if got := Text(); got != want {
		t.Fatalf("Text() = %d bytes, want flows then reference (%d bytes)", len(got), len(want))
	}
	if Flows() != flows {
		t.Fatalf("Flows() does not equal flows.md (%d bytes, want %d)", len(Flows()), len(flows))
	}
	if Reference() != reference {
		t.Fatalf("Reference() does not equal reference.md (%d bytes, want %d)", len(Reference()), len(reference))
	}
}

// TestFlowsGiveEachFlow proves the content of AC1: part 1 gives the task,
// verify, land, inbox, ask, and chat flows, in that order.
func TestFlowsGiveEachFlow(t *testing.T) {
	flows := readPart(t, "flows.md")
	last := -1
	for _, name := range []string{"Task", "Verify", "Land", "Inbox", "Ask", "Chat"} {
		heading := "## " + name + "\n"
		index := strings.Index(flows, heading)
		if index < 0 {
			t.Fatalf("flows.md has no heading %q", strings.TrimSpace(heading))
		}
		if index <= last {
			t.Fatalf("heading %q is out of order", strings.TrimSpace(heading))
		}
		last = index
	}
}

// TestRenderReferenceFormat proves the reference line format and the
// deterministic order of AC2. The input is unsorted on purpose.
func TestRenderReferenceFormat(t *testing.T) {
	got := RenderReference([]Command{
		{Path: "hive beta", Short: "second command"},
		{Path: "hive alpha", Short: "first command", Flags: []Flag{
			{Name: "verbose", Type: "bool", Usage: "print more"},
			{Name: "help", Shorthand: "h", Type: "bool", Usage: "help for alpha"},
			{Name: "count", Type: "int", Usage: "how many"},
		}},
	})
	want := "# hive reference\n\n" +
		"The command tree of hive. Each command line gives the path and the short\n" +
		"help. Each flag line below it gives the name, the type, and the usage.\n\n" +
		"hive alpha: first command\n" +
		"  --count int: how many\n" +
		"  -h, --help bool: help for alpha\n" +
		"  --verbose bool: print more\n\n" +
		"hive beta: second command\n\n"
	if got != want {
		t.Fatalf("RenderReference() =\n%q\nwant\n%q", got, want)
	}
}

// TestGuideFitsLimit proves AC4 against the literal 24 KiB bound.
func TestGuideFitsLimit(t *testing.T) {
	const want = 24 * 1024
	if MaxBytes != want {
		t.Fatalf("MaxBytes = %d, want %d", MaxBytes, want)
	}
	if size := len(Text()); size >= want {
		t.Fatalf("guide is %d bytes, want below %d", size, want)
	}
}
