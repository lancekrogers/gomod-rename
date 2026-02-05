// gomod-rename replaces Go module import paths across a codebase.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lancekrogers/gomod-rename/internal/cli"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := cli.New(os.Stdout, os.Stderr, os.Stdin, version)
	os.Exit(c.Run(ctx, os.Args[1:]))
}
