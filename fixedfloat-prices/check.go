package main

import (
	"errors"
	"fmt"
	"github.com/Elbandi/zabbix-checker/common/urfavecli"
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
		&cli.GenericFlag{
			Name:  "direction",
			Usage: "Exchange direction",
			Value: &urfavecli.EnumValue{
				Enum:    []string{"from", "to"},
				Default: "from",
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
	xpath := fmt.Sprintf("//select[@id='select_currency_%s']/option[@data-tag]", ctx.String("direction"))
	currencies, err := htmlquery.QueryAll(doc, xpath)
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
