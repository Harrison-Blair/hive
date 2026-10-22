package boardexport

import (
	"os"
	"path/filepath"
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
