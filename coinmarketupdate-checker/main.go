package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Elbandi/zabbix-checker/common/lld"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
	"time"
)

var (
	MongoDBHosts string
	AuthDatabase string
	AuthUserName string
	AuthPassword string
)

func Contains(list []string, elem string) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

func RunMongo(databaseName string, collectionName string, f func(*mgo.Collection) error) error {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{MongoDBHosts},
		Database: AuthDatabase,
		Username: AuthUserName,
		Password: AuthPassword,
		Timeout:  5 * time.Second,
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return err
	}
	defer session.Close()

	databases, err := session.DatabaseNames()
	if err != nil {
		return err
	}
	if !Contains(databases, databaseName) {
		return errors.New("No such database")
	}
	database := session.DB(databaseName)

	collections, err := database.CollectionNames()
	if err != nil {
		return err
	}
	if !Contains(collections, collectionName) {
		return errors.New("No such collection")
	}
	collection := database.C(collectionName)
	return f(collection)
}

// DiscoverExchanges is a DiscoveryItemHandlerFunc for key `coinmarket.discovery` which returns JSON
// encoded discovery data for all exchanges
func DiscoverExchanges(request []string) (lld.DiscoveryData, error) {
	// init discovery data
	d := make(lld.DiscoveryData, 0)

	var result []string
	err := RunMongo(request[0], request[1], func(collection *mgo.Collection) error {
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
	err := RunMongo(request[0], request[1], func(collection *mgo.Collection) error {
		return collection.Find(bson.M{"exchange": request[2]}).Sort("-date").Limit(1).All(&resp)
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
	flag.StringVar(&MongoDBHosts, "server", "127.0.0.1", "MongoDB server ip")
	flag.StringVar(&AuthDatabase, "database", "admin", "MongoDB auth database")
	flag.StringVar(&AuthUserName, "username", "admin", "MongoDB auth username")
	flag.StringVar(&AuthPassword, "password", "", "MongoDB auth password")
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.Arg(0) {
	case "discovery":
		switch flag.NArg() {
		case 3:
			if v, err := DiscoverExchanges(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v.Json())
			}
		default:
			log.Fatalf("Usage: %s discovery [--server ip] [--username user] [--password pass] database collection", os.Args[0])
		}
	case "lastdate":
		switch flag.NArg() {
		case 4:
			if v, err := QueryLastDate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s lastdate [--server ip] [--username user] [--password pass] database collection EXCHANGE", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: 'discovery' or 'lastdate'.")
	}
}
