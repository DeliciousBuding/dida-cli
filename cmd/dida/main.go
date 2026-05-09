package main

import (
	"os"

	"github.com/DeliciousBuding/dida-cli/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], version, os.Stdout, os.Stderr))
}
