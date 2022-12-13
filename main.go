package main

import (
	"os"

	"gitlab.com/rarimo/savers/sol-saver-svc/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
