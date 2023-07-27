package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
)

var fetchCommand = cli.Command{
	Name:   "fetch",
	Usage:  "fetch last data",
	Action: cmdFetch,
}

func cmdFetch(_ *cli.Context) error {
	response, err := FetchData("https://johoe.jochen-hoenicke.de/queue/2h.js")
	if err != nil {
		return err
	}
	var data []Data
	err = json.Unmarshal([]byte(response), &data)
	if err != nil {
		return err
	}
	latestElement := data[len(data)-1]
	res, err := ProcessDatat(latestElement)
	if err != nil {
		return err
	}
	d, err := json.Marshal(res)
	if err != nil {
		return err
	}
	fmt.Println(string(d))
	return nil
}
