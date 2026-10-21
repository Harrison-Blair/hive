package boardexport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Harrison-Blair/hive/internal/lib/cellgrid"
)

// boardScript matches the one board script of a board page.
var boardScript = regexp.MustCompile(`(?s)<script type="text/x-dc" data-dc-script[^>]*>(.*?)</script>`)

// harness runs the board script from stdin once in a fresh context. It wraps
// T to keep each grid that T makes, calls renderVals, and prints the grids as
// JSON. The rows() output is not used.
const harness = `
const vm = require('vm');
const src = require('fs').readFileSync(0, 'utf8');
const ctx = { DCLogic: class {} };
vm.createContext(ctx);
const grids = vm.runInContext(src + '\n;var __T = T, __grids = []; T = function (w, h) { var t = __T(w, h); __grids.push(t); return t; }; new Component().renderVals(); __grids;', ctx);
process.stdout.write(JSON.stringify(grids));
`

// Script returns the board script of a board page.
func Script(page string) (string, error) {
	m := boardScript.FindAllStringSubmatch(page, -1)
	if len(m) != 1 {
		return "", fmt.Errorf("found %d board scripts, want 1", len(m))
	}
	return m[0][1], nil
}

// run runs a board script once with Node.js and returns the grid of T.
func run(node, script string) (Raw, error) {
	cmd := exec.Command(node, "-e", harness)
	cmd.Stdin = strings.NewReader(script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return Raw{}, fmt.Errorf("node: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var grids []Raw
	if err := json.Unmarshal(stdout.Bytes(), &grids); err != nil {
		return Raw{}, fmt.Errorf("read the grids of T: %w", err)
	}
	if len(grids) != 1 {
		return Raw{}, fmt.Errorf("the script made %d grids with T, want 1", len(grids))
	}
	return grids[0], nil
}

// ExportDir runs the script of each .dc.html board in dir, checks its grid,
// and writes <Board>.grid to out. It writes no grid when a board fails.
func ExportDir(dir, out string) error {
	node, err := exec.LookPath("node")
	if err != nil {
		return fmt.Errorf("the exporter needs Node.js to run the board scripts: %w", err)
	}
	pages, err := filepath.Glob(filepath.Join(dir, "*.dc.html"))
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return fmt.Errorf("no .dc.html board in %s", dir)
	}
	grids := map[string]cellgrid.Grid{}
	var errs []error
	for _, page := range pages {
		name := strings.TrimSuffix(filepath.Base(page), ".dc.html")
		g, err := exportBoard(node, name, page)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		grids[name] = g
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	for name, g := range grids {
		var b bytes.Buffer
		if err := cellgrid.WriteGrid(&b, g); err != nil {
			return fmt.Errorf("board %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(out, name+".grid"), b.Bytes(), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func exportBoard(node, name, page string) (cellgrid.Grid, error) {
	data, err := os.ReadFile(page)
	if err != nil {
		return cellgrid.Grid{}, err
	}
	script, err := Script(string(data))
	if err != nil {
		return cellgrid.Grid{}, fmt.Errorf("board %s: %w", name, err)
	}
	r, err := run(node, script)
	if err != nil {
		return cellgrid.Grid{}, fmt.Errorf("board %s: %w", name, err)
	}
	if err := Check(name, r); err != nil {
		return cellgrid.Grid{}, err
	}
	return ToGrid(r)
}
