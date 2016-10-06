package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"github.com/Elbandi/zabbix-checker/common/lld"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
)

func Contains(list []string, elem string) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

func RunMongo(request []string, f func(*mgo.Collection) error) (error) {
	session, err := mgo.Dial(request[0])
	if err != nil {
		return err
	}
	defer session.Close()

	databases, err := session.DatabaseNames()
	if err != nil {
		return err
	}
	if !Contains(databases, request[1]) {
		return errors.New("No such database")
	}
	database := session.DB(request[1])

	collections, err := database.CollectionNames()
	if err != nil {
		return err
	}
	if !Contains(collections, request[2]) {
		return errors.New("No such collection")
	}
	collection := database.C(request[2])
	return f(collection)
}

// DiscoverExchanges is a DiscoveryItemHandlerFunc for key `coinmarket.discovery` which returns JSON
// encoded discovery data for all exchanges
func DiscoverExchanges(request []string) (lld.DiscoveryData, error) {
	// init discovery data
	d := make(lld.DiscoveryData, 0)

	var result []string
	err := RunMongo(request, func(collection *mgo.Collection) error {
		return collection.Find(nil).Distinct("exchange", &result)
	})
	if err != nil {
		return d, err
	}
	for _, exchange := range result {
		item := make(lld.DiscoveryItem, 0)
		item["NAME"] = exchange
		d = append(d, item)
	}
	return d, nil
}

// QueryLastDate is a Int64ItemHandlerFunc for key `coinmarket.lastdate` which returns the last item
// date.
func QueryLastDate(request []string) (int64, error) {
	var resp []bson.M
	err := RunMongo(request, func(collection *mgo.Collection) error {
		return collection.Find(bson.M{"exchange": request[3]}).Sort("-date").Limit(1).All(&resp)
	})
	if err != nil {
		return 0, err
	}
	if len(resp) == 0 {
		return 0, nil
	}
	return resp[0]["date"].(time.Time).Unix(), nil
}

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.Arg(0) {
	case "discovery":
		switch flag.NArg() {
		case 4:
			if v, err := DiscoverExchanges(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v.Json())
			}
		default:
			log.Fatalf("Usage: %s discovery mongoserver database collection", os.Args[0])
		}
	case "lastdate":
		switch flag.NArg() {
		case 5:
			if v, err := QueryLastDate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s lastdate mongoserver database collection EXCHANGE", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: 'discovery' or 'lastdate'.")
	}
}
