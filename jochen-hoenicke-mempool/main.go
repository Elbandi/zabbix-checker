package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "print-version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}
	app := &cli.App{
		Name:        "jochen-hoenicke-mempool",
		Version:     "v1.0",
		Description: "Get data from jochen-hoenicke.de",
		Commands: []*cli.Command{
			&fetchCommand,
			&trapperCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
