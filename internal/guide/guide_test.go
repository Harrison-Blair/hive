package guide

import (
	"strings"
	"testing"
)

// fixed holds a newline, a double quote, and a backslash.
const fixed = "a \"b\"\nc\\d\n"

func TestOneLineReplacesEachNewline(t *testing.T) {
	if got, want := OneLine(fixed), `a "b"\nc\d\n`; got != want {
		t.Fatalf("OneLine() = %q; want %q", got, want)
	}
}

func TestTOMLStringEscapesQuoteBackslashAndNewline(t *testing.T) {
	if got, want := TOMLString(fixed), `"a \"b\"\nc\\d\n"`; got != want {
		t.Fatalf("TOMLString() = %q; want %q", got, want)
	}
}

func TestRenderListsCommandsThenNotes(t *testing.T) {
	commands := []Command{
		{Use: "hive demo NAME", Short: "run the demo", Flags: []Flag{
			{Name: "count", Type: "int", Default: "3", Usage: "how many"},
			{Name: "fast", Shorthand: "f", Type: "bool", Default: "false", Usage: "go fast"},
			{Name: "label", Type: "string", Usage: "the label"},
		}},
		{Use: "hive empty", Short: "has no flags"},
	}
	global := []Flag{{Name: "json", Type: "bool", Default: "false", Usage: "write JSON output"}}
	want := "# hive guide\n" +
		"## Commands\n" +
		"Each command takes -h/--help. Global flags:\n" +
		"  --json bool: write JSON output\n" +
		"hive demo NAME: run the demo\n" +
		"  --count int=3: how many\n" +
		"  -f/--fast bool: go fast\n" +
		"  --label string: the label\n" +
		"hive empty: has no flags\n"
	got := Render(commands, global)
	if !strings.HasPrefix(got, want) {
		t.Fatalf("Render() commands part =\n%s\nwant\n%s", got, want)
	}
	if rest := strings.TrimPrefix(got, want); rest != Notes() {
		t.Fatalf("Render() after the commands =\n%s\nwant the usage notes", rest)
	}
}

func TestNotesCoverTheSevenItems(t *testing.T) {
	notes := Notes()
	for _, phrase := range []string{
		"## Usage notes",
		"hive agent list",
		"⬢ user:",
		"⬢ NAME:",
		"It does not widen that work.",
		"hive agent send PARENT TEXT",
		"`--project P --worktree NAME`",
		"--prompt-file",
		"blocked",
		"`send --wait`",
		"--remove-worktree",
		"refuses a dirty checkout",
		"`ok`, `partial`, or `failed`",
		"status 2",
		"`hive logs --agent NAME`",
	} {
		if !strings.Contains(notes, phrase) {
			t.Errorf("Notes() has no %q", phrase)
		}
	}
	for i := 1; i <= 7; i++ {
		if !strings.Contains(notes, "\n"+string(rune('0'+i))+". ") {
			t.Errorf("Notes() has no item %d", i)
		}
	}
}

func TestRedactReplacesTheGuideValue(t *testing.T) {
	argv := []string{"claude", "--model", "m", "--append-system-prompt", "line one\\nline two"}
	got := Redact(argv, []string{"--append-system-prompt", "line one\\nline two"})
	want := []string{"claude", "--model", "m", "--append-system-prompt", "<guide 18 bytes>"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("Redact() = %q; want %q", got, want)
	}
	if argv[4] != "line one\\nline two" {
		t.Fatalf("Redact() changed its input: %q", argv)
	}
	if got := Redact([]string{"pi"}, nil); strings.Join(got, "|") != "pi" {
		t.Fatalf("Redact() with no guide args = %q", got)
	}
}
