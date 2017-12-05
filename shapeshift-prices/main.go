package main

import (
	"github.com/hunterlong/shapeshift"
	"fmt"
	"flag"
	"log"
	"os"
)

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
		log.Fatal("You must specify one of the following action: 'rate' or 'limit'.")
	}
}
