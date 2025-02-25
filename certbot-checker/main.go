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
		Name:        "certbot-checker",
		Version:     "v1.0",
		Description: "Certificate checker for letsencrypt certbot",
		Authors: []*cli.Author{
			{
				Name:  "Elbandi",
				Email: "elso.andras@gmail.com",
			},
		},
		Commands: []*cli.Command{
			&certInfoCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
