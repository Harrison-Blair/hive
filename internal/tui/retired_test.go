package tui

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Reasons for retiring a tui-v2 grid in golden-frames task 137.
const (
	retiredFrame   = "golden-frames decision 2: no header and no tab bar; the screen keeps its body under the title row and the rule"
	retiredChat    = "golden-frames decisions 1 to 3 and chat.md: the GfChat boards replace the Chat screen"
	retiredCompact = "golden-frames decisions 4 and 27: the 80×24 form of each screen is deferred"
	retiredMain    = "golden-frames decision 1: Chat is home, so the Agents start screen of Main is gone"
	retiredHelp    = "golden-frames decisions 5 and 16: Help dims the full screen and lists the chords"
)

// retiredGrids names each tui-v2 grid that the golden frames retire, with
// the reason. Each other tui-v2 grid keeps its check.
var retiredGrids = map[string]string{
	"AgentDetails": retiredFrame, "AgentDetailsGeneral": retiredFrame,
	"AgentsCoordinatorChild": retiredFrame, "AgentsHiddenParent": retiredFrame,
	"AgentsHome": retiredFrame, "AgentsMany": retiredFrame, "AgentsNested": retiredFrame,
	"Chat": retiredChat, "ChatAsk": retiredChat, "ChatAskMulti": retiredChat,
	"ChatAskPreview": retiredChat, "ChatAskReview": retiredChat, "ChatAskReviewMulti": retiredChat,
	"ChatAskStopped": retiredChat, "ChatEvents": retiredChat, "ChatInput": retiredChat,
	"ChatNormal": retiredChat, "ChatStatus": retiredChat,
	"Compact": retiredCompact, "CompactAgents": retiredCompact, "CompactAgentsHome": retiredCompact,
	"CompactChat": retiredCompact, "CompactChatAsk": retiredCompact, "CompactChatAskReview": retiredCompact,
	"CompactChatNormal": retiredCompact, "CompactChatStatus": retiredCompact, "CompactFriction": retiredCompact,
	"CompactHealth": retiredCompact, "CompactLiveOverview": retiredCompact, "CompactLivePod": retiredCompact,
	"CompactLogs": retiredCompact, "CompactTreeNested": retiredCompact, "CompactTreeWide": retiredCompact,
	"Friction": retiredFrame, "FrictionResolved": retiredFrame,
	"Health": retiredFrame, "HealthAgents": retiredFrame, "HealthAgentsGeneral": retiredFrame,
	"Help":         retiredHelp,
	"LiveOverview": retiredFrame, "LivePod": retiredFrame, "Logs": retiredFrame,
	"Main":       retiredMain,
	"TreeNested": retiredFrame, "TreeWide": retiredFrame,
	"Work": retiredFrame, "WorkGoals": retiredFrame, "WorkQueue": retiredFrame,
}

// keptGrids names each tui-v2 grid that keeps its check, and the test that
// checks it.
var keptGrids = map[string]string{
	"TooSmall":   "TestDemoTooSmallMatchesItsBoardGrid",
	"Components": "the GOLDEN tests of internal/tui/component",
}

// AC9: each tui-v2 grid is retired with a reason or keeps its check. The
// kept checks run in the tests that keptGrids names.
func TestEachTuiV2GridIsRetiredOrKept(t *testing.T) {
	requireSpecFile(t, designDir)
	files, err := filepath.Glob(filepath.Join(designDir, "*.grid"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no tui-v2 grids: %v", err)
	}
	var boards []string
	for _, file := range files {
		board := strings.TrimSuffix(filepath.Base(file), ".grid")
		boards = append(boards, board)
		reason, retired := retiredGrids[board]
		_, kept := keptGrids[board]
		switch {
		case retired && kept:
			t.Errorf("%s is retired and kept", board)
		case retired && reason == "":
			t.Errorf("%s is retired with no reason", board)
		case !retired && !kept:
			t.Errorf("%s is neither retired nor kept", board)
		}
	}
	var missing []string
	for board := range retiredGrids {
		path := filepath.Join(designDir, board+".grid")
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, path)
		} else if err != nil {
			t.Errorf("retired grid %s: %v", board, err)
		}
	}
	sort.Strings(boards)
	t.Logf("%d tui-v2 grids: %d retired, %d kept", len(boards), len(retiredGrids), len(keptGrids))
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Skipf("local spec files missing: %s: Git does not track .hive/spec/", strings.Join(missing, ", "))
	}
}

// requireRetired fails when board is not a retired grid. A golden test of a
// retired tui-v2 grid calls it in place of the grid check.
func requireRetired(t *testing.T, board string) {
	t.Helper()
	if reason := retiredGrids[board]; reason == "" {
		t.Fatalf("%s is not a retired grid", board)
	}
}

// skipRetired reports that board is a retired grid. A golden check of a
// tui-v2 grid returns at once for a retired board: the grid has no frame to
// match, and TestEachTuiV2GridIsRetiredOrKept gives the reason.
func skipRetired(t *testing.T, board string) bool {
	t.Helper()
	if _, retired := retiredGrids[board]; !retired {
		return false
	}
	requireRetired(t, board)
	t.Logf("%s: retired: %s", board, retiredGrids[board])
	return true
}
