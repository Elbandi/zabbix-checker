package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
)

var trapperCommand = cli.Command{
	Name:  "trapper",
	Usage: "trapper data",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "key",
			Usage: "Item key",
			Action: func(ctx *cli.Context, v string) error {
				if len(v) == 0 {
					return cli.Exit("Flag key cannot be empty", 1)
				}
				return nil
			},
		},
		&cli.GenericFlag{
			Name:  "period",
			Usage: "Period data",
			Value: &EnumValue{
				Enum:    []string{"2h", "8h", "24h", "2d", "4d", "1w", "2w", "30d", "3m", "6m", "1y", "all"},
				Default: "2h",
			},
		},
	},
	Action: cmdTrapper,
}

func cmdTrapper(ctx *cli.Context) error {
	url := fmt.Sprintf("https://johoe.jochen-hoenicke.de/queue/%s.js", ctx.String("period"))
	response, err := FetchData(url)
	if err != nil {
		return err
	}
	var data []Data
	err = json.Unmarshal([]byte(response), &data)
	if err != nil {
		return err
	}
	for _, item := range data {
		res, err := ProcessDatat(item)
		if err != nil {
			return err
		}
		d, err := json.Marshal(res)
		if err != nil {
			return err
		}
		fmt.Printf("- %s %d %q\n", ctx.String("key"), item.Date.Unix(), string(d))
	}
	return nil
}
