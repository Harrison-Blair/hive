// Command boardexport writes the golden grid of each tui-v2 design board.
// It is a development tool for "go run" and is not part of the hive binary.
//
//	go run ./internal/boardexport/cmd/boardexport [-out OUT] DIR
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Harrison-Blair/hive/internal/boardexport"
)

func main() {
	out := flag.String("out", "", "folder for the grid files (default: DIR)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: boardexport [-out OUT] DIR")
		os.Exit(2)
	}
	dir := flag.Arg(0)
	if *out == "" {
		*out = dir
	}
	if err := boardexport.ExportDir(dir, *out); err != nil {
		fmt.Fprintln(os.Stderr, "boardexport:", err)
		os.Exit(1)
	}
}
