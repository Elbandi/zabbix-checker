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
		Name:        "fixedfloat-prices",
		Version:     "v1.0",
		Description: "Get data from FixedFloat.com",
		Commands: []*cli.Command{
			&checkCommand,
			&rateCommand,
			&limitCommand,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Debug http request/response",
				Value: false,
			},
			&cli.StringFlag{
				Name:    "api-key",
				Usage:   "Api key for fixedfloat",
				EnvVars: []string{"API_KEY"},
			},
			&cli.StringFlag{
				Name:    "api-secret",
				Usage:   "Api secret for fixedfloat",
				EnvVars: []string{"API_SECRET"},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
