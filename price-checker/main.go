package main

import (
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"strings"
	"time"
)

var (
	MongoDBHosts string
	AuthDatabase string
	AuthUserName string
	AuthPassword string
)

func GetPrice(request []string) (float64, error) {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{MongoDBHosts},
		Database: AuthDatabase,
		Username: AuthUserName,
		Password: AuthPassword,
		Timeout:  5 * time.Second,
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return 0.00, err
	}
	database := session.DB(request[0])
	match := bson.M{"coin": request[2], "base": request[3]}
	if len(request) > 4 {
		if len(request[4]) > 1 && strings.HasPrefix(request[4], "!") {
			match["exchange"] = bson.M{"$nin": strings.Split(request[4][1:], "+")}
		} else if request[4] != "*" {
			match["exchange"] = bson.M{"$in": strings.Split(request[4], "+")}
		}
	}
	pipeline := []bson.M{
		{"$match": match},
		{"$group": bson.M{"_id": "$exchange", "lastDate": bson.M{"$last": "$date"}, "price": bson.M{"$last": "$price"}, "volume": bson.M{"$last": "$volume"}}},
		{"$project": bson.M{"_id": 0, "exchange": "$_id", "buy": "$price.buy", "sell": "$price.sell", "volume": "$volume.coin", "basevolume": "$volume.base"}},
		{"$sort": bson.M{"volume": -1}},
		{"$limit": 1},
	}
	pipe := database.C(request[1]).Pipe(pipeline)
	var resp []bson.M
	err = pipe.All(&resp)
	if err != nil {
		return 0.00, err
	}
	if len(resp) == 0 {
		return 0.00, nil
	}
	if len(request) > 5 && request[5] == "sell" {
		return resp[0]["sell"].(float64), nil
	}
	return resp[0]["buy"].(float64), nil
}

func main() {
	flag.StringVar(&MongoDBHosts, "server", "127.0.0.1", "MongoDB server ip")
	flag.StringVar(&AuthDatabase, "database", "admin", "MongoDB auth database")
	flag.StringVar(&AuthUserName, "username", "admin", "MongoDB auth username")
	flag.StringVar(&AuthPassword, "password", "", "MongoDB auth password")
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.NArg() {
	case 4, 5, 6:
		if v, err := GetPrice(flag.Args()); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	default:
		log.Fatalf("Usage: %s database collection coin basecoin exchange sell", os.Args[0])
	}
}
