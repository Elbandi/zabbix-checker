package main

import (
	"github.com/hunterlong/shapeshift"
	"errors"
	"fmt"
	"flag"
	"log"
	"os"
	"reflect"
	"strings"
)

// Check is a Uint64ItemHandlerFunc for key `shapeshift.check` which returns the availability status
// by Shapeshift.
func Check(request []string) (uint64, error) {
	s, err := shapeshift.Coins()
	if err != nil {
		return 0, err
	}
	value := reflect.ValueOf(s)
	field := reflect.Indirect(value).FieldByName(strings.ToUpper(request[0]))
	if !field.IsValid() {
		return 0, errors.New("invalid coin")
	}
	if field.Interface().(struct{ shapeshift.Coin }).Coin.Status == "available" {
		return 1, nil
	}
	return 0, nil
}

// GetRate is a DoubleItemHandlerFunc for key `shapeshift.rate` which returns the current rate offered
// by Shapeshift.
func GetRate(request []string) (float64, error) {
	pair := shapeshift.Pair{request[0] + "_" + request[1]}
	rate, err := pair.GetRates()
	if err != nil {
		return 0, err
	}
	return rate, nil
}

// GetLimit is a DoubleItemHandlerFunc for key `shapeshift.rate` which returns the current rate offered
// by Shapeshift.
func GetLimit(request []string) (float64, error) {
	pair := shapeshift.Pair{request[0] + "_" + request[1]}
	limit, err := pair.GetLimits()
	if err != nil {
		return 0, err
	}
	return limit, nil
}

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.Arg(0) {
	case "check":
		switch flag.NArg() {
		case 2:
			if v, err := Check(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s check COIN", os.Args[0])
		}
	case "rate":
		switch flag.NArg() {
		case 3:
			if v, err := GetRate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s rate FROM TO", os.Args[0])
		}
	case "limit":
		switch flag.NArg() {
		case 3:
			if v, err := GetLimit(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s limit FROM TO", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: 'check', 'rate' or 'limit'.")
	}
}
