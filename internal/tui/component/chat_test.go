package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// gfMessages are the messages of GfComponents.dc.html rows 37 to 46. The
// sheet draws them with no blank rows, so each test renders one message.
var gfMessages = []struct {
	name string
	msg  ChatMessage
	r    rect
}{
	{"user", ChatMessage{Time: "14:02", Name: "you", Kind: ChatUser, Text: "move the status out of the composer border."}, rect{37, 0, 62, 2}},
	{"agent", ChatMessage{Time: "14:06", Name: "coordinator", Kind: ChatAgent, Text: "## Status row, summary\n\n- Segments are cut from the right. Config under `[tui.status]`."}, rect{39, 0, 62, 3}},
	{"event", ChatMessage{Time: "14:09", Name: "manager-hive", Kind: ChatEvent, Text: "quiet"}, rect{42, 0, 120, 1}},
	{"open event", ChatMessage{Time: "14:05", Name: "manager-store", Kind: ChatEvent, Open: true, Text: "g2-live-scroll DONE: task 121-live-transcript-scroll verified by t121-verifier-2 · 2772 unit passes · 0 fails"}, rect{43, 0, 120, 2}},
	{"answered", ChatMessage{Time: "14:23", Name: "you", Kind: ChatUser, Text: "answered manager-hive · Model → Yes, restart now"}, rect{45, 0, 120, 2}},
}

// gfChatSample gives the messages of the component sheet.
func gfChatSample() []ChatMessage {
	msgs := make([]ChatMessage, len(gfMessages))
	for i, tc := range gfMessages {
		msgs[i] = tc.msg
	}
	return msgs
}

// TestGfMessagesGolden renders each message of the component sheet at 120
// cells and compares it with its rows of GfComponents.grid (chat.md rules 1
// to 4). The user and agent rows compare columns 0 to 61, left of the legend.
func TestGfMessagesGolden(t *testing.T) {
	for _, tc := range gfMessages {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := ChatLines([]ChatMessage{tc.msg}, 120)
			if err != nil {
				t.Fatal(err)
			}
			if len(lines) != tc.r.h {
				t.Fatalf("%d lines, want %d: %q", len(lines), tc.r.h, lines)
			}
			rows := make([]string, len(lines))
			for i, l := range lines {
				rows[i] = fitRow(l, tc.r.w)
			}
			checkGolden(t, strings.Join(rows, "\n"), gfSheet, tc.r)
		})
	}
}

// Rule 2: one blank row separates two messages. Rule 3: two folded events
// have no blank row between them. Two ask lines have none either.
func TestChatBlankRows(t *testing.T) {
	user := ChatMessage{Time: "14:02", Name: "you", Kind: ChatUser, Text: "hi"}
	event := ChatMessage{Time: "14:03", Name: "manager-hive", Kind: ChatEvent, Text: "quiet"}
	open := ChatMessage{Time: "14:03", Name: "manager-hive", Kind: ChatEvent, Text: "quiet", Open: true}
	asked := ChatMessage{Time: "14:04", Name: "manager-hive", Kind: ChatAsk, Text: "asked · Model"}
	for _, tc := range []struct {
		name   string
		msgs   []ChatMessage
		second int
		want   []string
	}{
		{"user, user", []ChatMessage{user, user}, 3, []string{" › you 14:02", " hi", "", " › you 14:02", " hi"}},
		{"event, event", []ChatMessage{event, event}, 1, []string{" ▸ manager-hive 14:03  quiet", " ▸ manager-hive 14:03  quiet"}},
		{"open, event", []ChatMessage{open, event}, 3, []string{" ▾ manager-hive 14:03", " quiet", "", " ▸ manager-hive 14:03  quiet"}},
		{"ask, ask", []ChatMessage{asked, asked}, 1, []string{" ? manager-hive 14:04  asked · Model", " ? manager-hive 14:04  asked · Model"}},
		{"user, ask", []ChatMessage{user, asked}, 3, []string{" › you 14:02", " hi", "", " ? manager-hive 14:04  asked · Model"}},
	} {
		lines, err := ChatLines(tc.msgs, 60)
		if err != nil {
			t.Fatal(err)
		}
		got := make([]string, len(lines))
		for i, l := range lines {
			got[i] = strings.TrimRight(ansi.Strip(l), " ")
		}
		if strings.Join(got, "\n") != strings.Join(tc.want, "\n") {
			t.Errorf("%s: lines = %q, want %q", tc.name, got, tc.want)
		}
		starts, err := ChatStarts(tc.msgs, 60)
		if err != nil {
			t.Fatal(err)
		}
		if len(starts) != 2 || starts[0] != 0 || starts[1] != tc.second {
			t.Errorf("%s: starts = %v, want [0 %d]", tc.name, starts, tc.second)
		}
	}
}

// Rule 1: the text starts at column 1 at every width and wraps at column
// width-2, with no gutter.
func TestChatWrapsAtFullWidth(t *testing.T) {
	text := strings.TrimSpace(strings.Repeat("word ", 60))
	for _, width := range []int{40, 80, 120} {
		lines, err := ChatLines([]ChatMessage{{Time: "09:00", Name: "you", Kind: ChatUser, Text: text}}, width)
		if err != nil {
			t.Fatal(err)
		}
		for i, l := range lines[1:] {
			plain := ansi.Strip(l)
			if !strings.HasPrefix(plain, " word") {
				t.Errorf("width %d line %d does not start at column 1: %q", width, i+1, plain)
			}
			if w := ansi.StringWidth(plain); w > width-1 {
				t.Errorf("width %d line %d ends at column %d, want at most %d", width, i+1, w-1, width-2)
			}
		}
		if len(lines) < 3 {
			t.Errorf("width %d: %d lines, want the text to wrap", width, len(lines))
		}
	}
}

// Rule 3: a folded event is one row, cut with …. At 120 cells the row ends
// at column 116, as GfComponents row 42 pads it.
func TestFoldedEventCutsAtColumn118(t *testing.T) {
	lines, err := ChatLines([]ChatMessage{{Time: "14:09", Name: "manager-hive", Kind: ChatEvent, Text: strings.Repeat("x", 200) + "\nsecond line"}}, 120)
	if err != nil {
		t.Fatal(err)
	}
	plain := ansi.Strip(lines[0])
	if len(lines) != 1 || ansi.StringWidth(plain) != 117 || !strings.HasSuffix(plain, "…") || strings.Contains(plain, "second") {
		t.Fatalf("folded event = %q (%d cells), want one row cut with … at column 116", plain, ansi.StringWidth(plain))
	}
}
