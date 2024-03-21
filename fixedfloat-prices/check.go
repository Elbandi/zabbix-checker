package main

import (
	"errors"
	"fmt"
	"github.com/elbandi/go-fixedfloat-api"
	"github.com/urfave/cli/v2"
	"strings"
	"time"
)

var checkCommand = cli.Command{
	Name:  "check",
	Usage: "check availability",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "coin",
			Usage: "Coin",
			Action: func(ctx *cli.Context, v string) error {
				if len(v) == 0 {
					return cli.Exit("Flag coin cannot be empty", 1)
				}
				return nil
			},
		},
	},
	Action: cmdCheck,
}

func cmdCheck(ctx *cli.Context) error {
	client := fixedfloat.NewWithCustomTimeout(ctx.String("api-key"), ctx.String("api-secret"), 10*time.Second)
	client.SetDebug(ctx.Bool("debug"))
	currencies, err := client.GetCurrencies()
	if err != nil {
		return err
	}
	coin := strings.ToUpper(ctx.String("coin"))
	for _, c := range currencies {
		if c.Code == coin {
			fmt.Print(c.Send.Toint())
			return nil
		}
	}
	return errors.New("invalid coin")
}
