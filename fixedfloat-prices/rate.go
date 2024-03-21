package main

import (
	"fmt"
	"github.com/elbandi/go-fixedfloat-api"
	"github.com/urfave/cli/v2"
	"time"
)

var rateCommand = cli.Command{
	Name:  "rate",
	Usage: "get rate",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "from",
			Usage: "From currency",
			Action: func(ctx *cli.Context, v string) error {
				if len(v) == 0 {
					return cli.Exit("Flag 'from' cannot be empty", 1)
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:  "to",
			Usage: "to currency",
			Action: func(ctx *cli.Context, v string) error {
				if len(v) == 0 {
					return cli.Exit("Flag 'to' cannot be empty", 1)
				}
				return nil
			},
		},
		&cli.Float64Flag{
			Name:  "amount",
			Usage: "Exchange amount",
			Action: func(ctx *cli.Context, v float64) error {
				if v <= 0 {
					return cli.Exit("Flag 'amount' sould be positive", 1)
				}
				return nil
			},
		},
	},
	Action: cmdRate,
}

func cmdRate(ctx *cli.Context) error {
	client := fixedfloat.NewWithCustomTimeout(ctx.String("api-key"), ctx.String("api-secret"), 10*time.Second)
	client.SetDebug(ctx.Bool("debug"))
	_, to, err := client.GetRate(ctx.String("from"), ctx.String("to"), ctx.Float64("amount"))
	if err != nil {
		return err
	}
	fmt.Print(to.Rate)
	return nil
}
