package tui

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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
	checkTuiV2Grids(t, designDir)
}

// checkTuiV2Grids runs the check of TestEachTuiV2GridIsRetiredOrKept on the
// grids in dir.
func checkTuiV2Grids(t testing.TB, dir string) {
	t.Helper()
	requireSpecFile(t, dir)
	pattern := filepath.Join(dir, "*.grid")
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("tui-v2 grids: %v", err)
	}
	if len(files) == 0 {
		t.Skipf("local spec file missing: %s: Git does not track .hive/spec/", pattern)
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
		path := filepath.Join(dir, board+".grid")
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

// outcomeTB records the failure and the skip of a check. Fatalf and Skipf
// stop the goroutine of the check, as they do for a test.
type outcomeTB struct {
	testing.TB
	failed, skipped string
}

func (r *outcomeTB) Helper()                      {}
func (r *outcomeTB) Logf(string, ...any)          {}
func (r *outcomeTB) Errorf(f string, args ...any) { r.failed += fmt.Sprintf(f, args...) + "\n" }
func (r *outcomeTB) Fatalf(f string, args ...any) {
	r.failed += fmt.Sprintf(f, args...) + "\n"
	runtime.Goexit()
}
func (r *outcomeTB) Skipf(f string, args ...any) {
	r.skipped = fmt.Sprintf(f, args...)
	runtime.Goexit()
}

// runCheck runs check with an outcomeTB and returns the outcome.
func runCheck(t *testing.T, check func(testing.TB)) *outcomeTB {
	r := &outcomeTB{TB: t}
	done := make(chan struct{})
	go func() {
		defer close(done)
		check(r)
	}()
	<-done
	return r
}

// AC8: the grid check skips and names the missing path for each partial
// state of the tui-v2 design folder. It still fails on a grid that is
// neither retired nor kept.
func TestTuiV2GridCheckSkipsEachPartialDesignFolder(t *testing.T) {
	for _, tc := range []struct {
		name     string
		files    []string
		skipped  string
		failures bool
	}{
		{name: "no folder", skipped: "no-such-folder"},
		{name: "empty folder", skipped: "*.grid"},
		{name: "fixture without grid", files: []string{"Main.fixture.json"}, skipped: "*.grid"},
		{name: "some grids", files: []string{"Main.grid"}, skipped: "Help.grid"},
		{name: "unknown grid", files: []string{"Unknown.grid"}, skipped: "Main.grid", failures: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.name == "no folder" {
				dir = filepath.Join(dir, "no-such-folder")
			}
			for _, name := range tc.files {
				if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got := runCheck(t, func(r testing.TB) { checkTuiV2Grids(r, dir) })
			if (got.failed != "") != tc.failures {
				t.Errorf("failures = %q, want failures %v", got.failed, tc.failures)
			}
			if !strings.Contains(got.skipped, "local spec file") || !strings.Contains(got.skipped, tc.skipped) {
				t.Errorf("skip = %q, want a skip that names %q", got.skipped, tc.skipped)
			}
		})
	}
}
