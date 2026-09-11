// Command tellme is the entrypoint for the tellme CLI.
package main

import (
	"os"

	"github.com/gosharplite/tellme/internal/cli"
)

// version is the single version symbol for the binary, injected at build time
// via `-ldflags "-X main.version=<value>"`. Do not introduce a second version
// source (research.md Decision 1 / Decision 6).
var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], version))
}
