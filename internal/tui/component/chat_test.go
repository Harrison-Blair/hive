package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// chatReply is the coordinator reply of Chat.dc.html lines 41-62.
const chatReply = "## Interrogate, round 2\n\n" +
	"**1. What happens to the old tab on restart?**\n" +
	"Bare hive makes the `coordinator` tab, but its record has `spawned=false`. `agent stop` refuses it.\n\n" +
	"| Choice | Answer | Pros | Cons |\n| --- | --- | --- | --- |\n" +
	"| A (REC) | Record that bare hive started it | No tab pile-up | Field change |\n" +
	"| B | Leave the old tab, open a new one | Smallest change | Dead tabs pile up |\n" +
	"| C | Reuse the same pane | One tab always | Must exit harness |\n\n" +
	"- **Recommended:** A, because hive may close only panes it started.\n" +
	"- Reply like `1A, 2A, 3A`, or `simplify N`.\n\n" +
	"```\nhive agent stop manager-hive\nhive agent spawn --role manager\n```\n"

// chatSample is the feed of Chat.dc.html lines 39-63.
var chatSample = []ChatMessage{
	{Time: "14:02", Name: "you", Kind: ChatUser, Text: "I want to give the user an ability to restart the coordinator or supervisor within the hive UI."},
	{Time: "14:19", Name: "manager-hive", Kind: ChatEvent, Text: "goal g1-tui-restart created · wave A started · worker t2-worker-1 live"},
	{Time: "14:21", Name: "coordinator", Kind: ChatAgent, Text: chatReply},
	{Time: "14:24", Name: "you", Kind: ChatUser, Text: "1A, 2A, 3B"},
}

// TestChatGolden renders the feed of Chat.dc.html for the feed panel of 120
// cells and compares the text cells, cols 2-117, with Chat.grid. Rows 15-19
// hold the table. They join the check when the board shows the Glamour
// table (decision 32).
func TestChatGolden(t *testing.T) {
	lines, err := ChatLines(chatSample, 120)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 23 {
		t.Fatalf("%d lines, want 23 (board rows 5-27)", len(lines))
	}
	part := func(from, to int) string {
		var rows []string
		for _, l := range lines[from:to] {
			rows = append(rows, fitRow(l, 116))
		}
		return strings.Join(rows, "\n")
	}
	t.Run("before the table", func(t *testing.T) { checkGolden(t, part(0, 10), "Chat", rect{5, 2, 116, 10}) })
	t.Run("after the table", func(t *testing.T) { checkGolden(t, part(15, 23), "Chat", rect{20, 2, 116, 8}) })
}

// TestChatNarrow checks the feed below 120 columns (decision 30): the time
// and the name use their own line, and the text starts at the padding and
// wraps at the panel width less 4. An event folds to one dim line.
func TestChatNarrow(t *testing.T) {
	msgs := []ChatMessage{
		{Time: "14:02", Name: "you", Kind: ChatUser, Text: "restart the coordinator please"},
		{Time: "14:19", Name: "manager-hive", Kind: ChatEvent, Text: "goal created · wave A started"},
		{Time: "14:21", Name: "coordinator", Kind: ChatAgent, Text: "Done **now**."},
	}
	lines, err := ChatLines(msgs, 24)
	if err != nil {
		t.Fatal(err)
	}
	want := "hive-grid 1 20 8\n" +
		"t|14:02 you           \n" + "k|dddddfuuufffffffffff\n" + "b|00000011100000000000\n" +
		"t|restart the         \n" + "k|uuuuuuuuuuufffffffff\n" + "b|00000000000000000000\n" +
		"t|coordinator please  \n" + "k|uuuuuuuuuuuuuuuuuuff\n" + "b|00000000000000000000\n" +
		"t|                    \n" + "k|ffffffffffffffffffff\n" + "b|00000000000000000000\n" +
		"t|14:19 · manager-hiv…\n" + "k|dddddddddddddddddddd\n" + "b|00000000000000000000\n" +
		"t|                    \n" + "k|ffffffffffffffffffff\n" + "b|00000000000000000000\n" +
		"t|14:21 coordinator   \n" + "k|dddddfiiiiiiiiiiifff\n" + "b|00000011111111111000\n" +
		"t|Done now.           \n" + "k|ffffffffffffffffffff\n" + "b|00000111000000000000\n"
	checkHandGrid(t, strings.Join(lines, "\n"), want)
}

// TestChatWrapWidth checks the gutter from 120 columns: the text starts at
// col 21 of the line, panel col 23, and wraps at the panel width less 26.
func TestChatWrapWidth(t *testing.T) {
	for _, width := range []int{120, 160} {
		text := strings.TrimSpace(strings.Repeat("word ", 60))
		lines, err := ChatLines([]ChatMessage{{Time: "09:00", Name: "you", Kind: ChatUser, Text: text}}, width)
		if err != nil {
			t.Fatal(err)
		}
		for i, l := range lines {
			plain := ansi.Strip(l)
			if plain == "" {
				continue
			}
			if w := ansi.StringWidth(plain); w > 21+width-26 {
				t.Errorf("width %d line %d is %d cells, want at most %d", width, i, w, 21+width-26)
			}
			if i > 0 && (!strings.HasPrefix(plain, strings.Repeat(" ", 21)) || plain[21] == ' ') {
				t.Errorf("width %d line %d does not start at col 21: %q", width, i, plain)
			}
		}
	}
}
