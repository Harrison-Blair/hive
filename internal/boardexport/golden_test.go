package boardexport

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// designDir holds the boards and the golden grids, relative to this package.
const designDir = "../../.hive/spec/tui-v2/design"

// TestGoldenGridFiles reads each committed grid file. It checks the size
// against the spec table and checks each symbol. It runs no Node.js.
func TestGoldenGridFiles(t *testing.T) {
	requireSpecFile(t, designDir)
	files, err := filepath.Glob(filepath.Join(designDir, "*.grid"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if _, ok := Sizes[strings.TrimSuffix(filepath.Base(file), ".grid")]; !ok {
			t.Errorf("%s is not in the board table", file)
		}
	}
	for board, size := range Sizes {
		t.Run(board, func(t *testing.T) {
			path := filepath.Join(designDir, board+".grid")
			requireSpecFile(t, path)
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = f.Close() }()
			g, err := cellgrid.ReadGrid(f)
			if err != nil {
				t.Fatal(err)
			}
			if g.Width != size[0] || g.Height != size[1] {
				t.Errorf("size %d×%d, want %d×%d", g.Width, g.Height, size[0], size[1])
			}
			if err := CheckGrid(board, g); err != nil {
				t.Error(err)
			}
		})
	}
}

// gfDesignDir holds the Gf boards and their golden grids of task 137.
const gfDesignDir = "../../.hive/spec/golden-frames/design"

// gfGrids gives the Gf grids in design/: the 15 boards of Chat and the
// component sheet of task 137, the Help board of task 148, the 4 Agents
// boards of task 147, and the Friction boards of task 144.
var gfGrids = []string{
	"GfChat", "GfChatAsk", "GfChatAskInput", "GfChatAskMany", "GfChatAskMulti",
	"GfChatAskPreview", "GfChatAskReview", "GfChatAskStopped", "GfChatEmpty", "GfChatEvents",
	"GfChatMarkdown", "GfChatNoCoordinator", "GfChatNotice", "GfChatTall", "GfComponents",
	"GfHelp",
	"GfAgents", "GfAgentsPod", "GfAgentsDetails", "GfAgentsNone",
	"GfFriction", "GfFrictionEmpty", "GfFrictionFilter", "GfFrictionResolved",
}

// TestGfGoldenGridFiles reads each committed Gf grid file of gfGrids. Each
// grid has the size of golden-frames decision 23 and the symbols and tokens
// of decision 7.
func TestGfGoldenGridFiles(t *testing.T) {
	requireSpecFile(t, gfDesignDir)
	files, err := filepath.Glob(filepath.Join(gfDesignDir, "*.grid"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if !slices.Contains(gfGrids, strings.TrimSuffix(filepath.Base(file), ".grid")) {
			t.Errorf("%s is not one of the Gf grids of gfGrids", file)
		}
	}
	for _, board := range gfGrids {
		t.Run(board, func(t *testing.T) {
			size, ok := GfSizes[board]
			if !ok {
				t.Fatalf("%s is not in the board table of decision 23", board)
			}
			path := filepath.Join(gfDesignDir, board+".grid")
			requireSpecFile(t, path)
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = f.Close() }()
			g, err := cellgrid.ReadGrid(f)
			if err != nil {
				t.Fatal(err)
			}
			if g.Width != size[0] || g.Height != size[1] {
				t.Errorf("size %d×%d, want %d×%d", g.Width, g.Height, size[0], size[1])
			}
			if err := CheckGrid(board, g); err != nil {
				t.Error(err)
			}
		})
	}
}

// requireSpecFile skips the test when a file or a folder of the local spec is
// missing. Git does not track .hive/spec/, so a clean clone has no grid and
// no fixture. The test runs as before when the file exists.
func requireSpecFile(t testing.TB, path string) {
	t.Helper()
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		t.Skipf("local spec file missing: %s: Git does not track .hive/spec/", path)
	}
}
