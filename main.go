package main

import (
	"fmt"
	"os"

	"github.com/Harrison-Blair/hive/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "hive:", err)
		os.Exit(1)
	}
}
