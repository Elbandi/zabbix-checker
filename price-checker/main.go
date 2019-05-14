package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func GetPrice(request []string) (float64, error) {
	session, err := mgo.Dial(request[0])
	if err != nil {
		return 0.00, err
	}
	database := session.DB(request[1])
	match := bson.M{"coin": request[3], "base": request[4]}
	if len(request) > 5 && request[5] != "*" {
		match["exchange"] = bson.M{"$nin": strings.Split(request[5], "+")}
	}
	pipeline := []bson.M{
		{"$match": match},
		{"$group": bson.M{"_id": "$exchange", "lastDate": bson.M{"$last": "$date"}, "price": bson.M{"$last": "$price"}, "volume": bson.M{"$last": "$volume"}}},
		{"$project": bson.M{"_id": 0, "exchange": "$_id", "buy": "$price.buy", "sell": "$price.sell", "volume": "$volume.coin", "basevolume": "$volume.base"}},
		{"$sort": bson.M{"volume": -1}},
		{"$limit": 1},
	}
	pipe := database.C(request[2]).Pipe(pipeline);
	var resp []bson.M
	err = pipe.All(&resp)
	if err != nil {
		return 0.00, err
	}
	if len(resp) == 0 {
		return 0.00, nil
	}
	if len(request) > 6 && request[6] == "sell" {
		return resp[0]["sell"].(float64), nil
	}
	return resp[0]["buy"].(float64), nil
}

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.NArg() {
	case 5, 6, 7:
		if v, err := GetPrice(flag.Args()); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	default:
		log.Fatalf("Usage: %s mongoserver database collection coin basecoin sell", os.Args[0])
	}
}

