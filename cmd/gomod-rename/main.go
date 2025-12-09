// gomod-rename replaces Go module import paths across a codebase.
package main

import (
	"os"

	"github.com/lancekrogers/gomod-rename/internal/cli"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	c := cli.New(os.Stdout, os.Stderr, os.Stdin, version)
	os.Exit(c.Run(os.Args[1:]))
}
