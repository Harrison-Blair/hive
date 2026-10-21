package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// sampleMarkdown has each element of decision 29: a heading, strong text,
// inline code, a link, a list, and a code block with an indented line.
const sampleMarkdown = "# Title\n\nSome **strong** and `code` with [a link](https://example.com/x).\n\n- one\n- two\n\n```go\nx := 1\n  y\n```\n"

// TestMarkdownGrid renders sampleMarkdown at 40 cells and compares it with a
// hand-written grid. The heading is accent and bold with no "#". The link
// shows its text in info and no URL. The bullet "• " is accent. The code
// block is code, indented by 2, with its own leading spaces kept. No line
// has trailing spaces.
func TestMarkdownGrid(t *testing.T) {
	want := "hive-grid 1 40 10\n" +
		"t|Title                                   \n" +
		"k|aaaaafffffffffffffffffffffffffffffffffff\n" +
		"b|1111100000000000000000000000000000000000\n" +
		"t|                                        \n" +
		"k|ffffffffffffffffffffffffffffffffffffffff\n" +
		"b|0000000000000000000000000000000000000000\n" +
		"t|Some strong and code with a link.       \n" +
		"k|ffffffffffffffffccccffffffiiiiiiffffffff\n" +
		"b|0000011111100000000000000000000000000000\n" +
		"t|                                        \n" +
		"k|ffffffffffffffffffffffffffffffffffffffff\n" +
		"b|0000000000000000000000000000000000000000\n" +
		"t|• one                                   \n" +
		"k|aaffffffffffffffffffffffffffffffffffffff\n" +
		"b|0000000000000000000000000000000000000000\n" +
		"t|• two                                   \n" +
		"k|aaffffffffffffffffffffffffffffffffffffff\n" +
		"b|0000000000000000000000000000000000000000\n" +
		"t|                                        \n" +
		"k|ffffffffffffffffffffffffffffffffffffffff\n" +
		"b|0000000000000000000000000000000000000000\n" +
		"t|  x := 1                                \n" +
		"k|ffccccccffffffffffffffffffffffffffffffff\n" +
		"b|0000000000000000000000000000000000000000\n" +
		"t|    y                                   \n" +
		"k|ffcccfffffffffffffffffffffffffffffffffff\n" +
		"b|0000000000000000000000000000000000000000\n" +
		"t|                                        \n" +
		"k|ffffffffffffffffffffffffffffffffffffffff\n" +
		"b|0000000000000000000000000000000000000000\n"
	lines, err := RenderMarkdown(sampleMarkdown, 40)
	if err != nil {
		t.Fatal(err)
	}
	checkHandGrid(t, strings.Join(lines, "\n"), want)
	for i, l := range lines {
		if strings.HasSuffix(ansi.Strip(l), " ") {
			t.Errorf("line %d has trailing spaces: %q", i, l)
		}
	}
}

// TestMarkdownAutoLinks checks that an autolink shows its text in info with
// no OSC 8 hyperlink, so the output has only SGR sequences.
func TestMarkdownAutoLinks(t *testing.T) {
	want := "hive-grid 1 29 1\n" +
		"t|see https://a.b/c and me@x.io\n" +
		"k|ffffiiiiiiiiiiiiifffffiiiiiii\n" +
		"b|00000000000000000000000000000\n"
	lines, err := RenderMarkdown("see <https://a.b/c> and <me@x.io>\n", 30)
	if err != nil {
		t.Fatal(err)
	}
	checkHandGrid(t, strings.Join(lines, "\n"), want)
}

// TestMarkdownRemovesBlankFirstLine checks that a reply that starts with a
// list or a code block has no blank first line. Glamour adds one.
func TestMarkdownRemovesBlankFirstLine(t *testing.T) {
	for _, src := range []string{"- a\n", "```\n  c\n```\n"} {
		lines, err := RenderMarkdown(src, 20)
		if err != nil {
			t.Fatal(err)
		}
		if first := ansi.Strip(lines[0]); strings.TrimSpace(first) == "" {
			t.Errorf("RenderMarkdown(%q): blank first line %q", src, lines[0])
		}
	}
}
