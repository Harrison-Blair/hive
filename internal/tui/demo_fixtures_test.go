package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// specBoardRow matches one row of the board table in the tui-v2 README, for
// example "| `Main.dc.html` | 1 Agents, one project | 120×40 |".
var specBoardRow = regexp.MustCompile("^\\| `([A-Za-z]+)\\.dc\\.html` \\|.*\\| (\\d+)×(\\d+) \\|$")

// TestEachBoardFixtureLoadsAtItsSpecSize loads the fixture of each board in
// the spec table with the demo loader. The size of each fixture must be the
// size in the spec table.
func TestEachBoardFixtureLoadsAtItsSpecSize(t *testing.T) {
	readme, err := os.ReadFile(filepath.Join(designDir, "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	boards := 0
	for line := range strings.SplitSeq(string(readme), "\n") {
		match := specBoardRow.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		boards++
		board := match[1]
		width, _ := strconv.Atoi(match[2])
		height, _ := strconv.Atoi(match[3])
		t.Run(board, func(t *testing.T) {
			fixture, err := loadDemoFixture(filepath.Join(designDir, board+".fixture.json"))
			if err != nil {
				t.Fatal(err)
			}
			if fixture.Board != board+".dc.html" {
				t.Errorf("board = %q; want %q", fixture.Board, board+".dc.html")
			}
			if fixture.Terminal.Width != width || fixture.Terminal.Height != height {
				t.Errorf("terminal = %dx%d; want %dx%d", fixture.Terminal.Width, fixture.Terminal.Height, width, height)
			}
		})
	}
	if boards != 39 {
		t.Fatalf("spec table has %d boards; want 39", boards)
	}
}
