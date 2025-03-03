package main

import (
	"github.com/EvgenyRomanov/sql-migrator/internal/cli"
	"os"
)

func main() {
	args := os.Args
	if len(args) > 1 && args[1] == "version" {
		printVersion()
		return
	}

	cli.Main()
}
