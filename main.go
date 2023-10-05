package main

import (
	"os"

	"github.com/rarimo/sol-saver-svc/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
