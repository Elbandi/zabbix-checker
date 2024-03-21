package main

import (
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
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
	Action: func(ctx *cli.Context) error {
		if ctx.IsSet("api-key") {
			return cmdCheckApi(ctx)
		}
		return cmdCheckWeb(ctx)
	},
}

func cmdCheckApi(ctx *cli.Context) error {
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

func cmdCheckWeb(ctx *cli.Context) error {
	debug = ctx.Bool("debug")
	doc, err := fetchPage("https://ff.io/")
	if err != nil {
		return err
	}
	currencies, err := htmlquery.QueryAll(doc, "//select[@id='select_currency_from']/option[@data-tag]")
	if err != nil {
		return err
	}
	coin := strings.ToUpper(ctx.String("coin"))
	for _, c := range currencies {
		if htmlquery.SelectAttr(c, "value") == coin {
			if htmlquery.SelectAttr(c, "data-inactive") == "0" {
				fmt.Print(1)
			} else {
				fmt.Print(0)
			}
			return nil
		}
	}
	return errors.New("invalid coin")
}
