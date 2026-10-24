package boardexport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// designDir holds the boards and the golden grids, relative to this package.
const designDir = "../../.hive/spec/tui-v2/design"

// TestGoldenGridFiles reads each committed grid file. It checks the size
// against the spec table and checks each symbol. It runs no Node.js.
func TestGoldenGridFiles(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(designDir, "*.grid"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(Sizes) {
		t.Errorf("found %d grid files, want %d: %v", len(files), len(Sizes), files)
	}
	for board, size := range Sizes {
		t.Run(board, func(t *testing.T) {
			f, err := os.Open(filepath.Join(designDir, board+".grid"))
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

// TestGfGoldenGridFiles reads each committed Gf grid file. Task 137 copies
// the 15 boards of Chat and the component sheet. Each grid has the size of
// golden-frames decision 23 and the symbols and tokens of decision 7.
func TestGfGoldenGridFiles(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(gfDesignDir, "*.grid"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 15 {
		t.Errorf("found %d Gf grid files, want 15: %v", len(files), files)
	}
	for _, file := range files {
		board := strings.TrimSuffix(filepath.Base(file), ".grid")
		t.Run(board, func(t *testing.T) {
			size, ok := GfSizes[board]
			if !ok {
				t.Fatalf("%s is not in the board table of decision 23", board)
			}
			f, err := os.Open(file)
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
