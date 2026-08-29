package main

import (
	"os"

	"github.com/unmango/kubepkgs/tools/update/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
