package main

import (
	"github.com/bitbandi/go-miningrigrentals-api"
	"github.com/Elbandi/zabbix-checker/common/lld"
	"github.com/Elbandi/zabbix-checker/common/filemutex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// DiscoverRentals is a DiscoveryItemHandlerFunc for key `mrr.discovery` which returns JSON
// encoded discovery data for all rentals
func DiscoverRentals(request []string) (lld.DiscoveryData, error) {
	// init discovery data
	d := make(lld.DiscoveryData, 0)
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "mrr-" + request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := miningrigrentals.New(request[0], request[1])
	rentals, err := client.ListMyRentals()
	if err != nil {
		return nil, err
	}
	for _, rent := range rentals {
		item := make(lld.DiscoveryItem, 0)
		item["ID"] = strconv.FormatInt(int64(rent.Id), 10)
		item["TYPE"] = rent.Type
		item["NAME"] = rent.Name
		d = append(d, item)
	}
	return d, nil
}


// QueryRigStatus is a StringItemHandlerFunc for key `mrr.rigstatus` which returns the status
// of rig for a rentals.
func QueryRigStatus(request []string) (string, error) {
	// parse first param as int64
	rentalid, err := strconv.ParseInt(request[2], 10, 64)
	if err != nil {
		return "na", errors.New("Invalid rentalid format")
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "mrr-" + request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := miningrigrentals.New(request[0], request[1])
	rentals, err := client.GetRentalDetails(rentalid)
	if err != nil {
		return "na", err
	}
	rig, err := client.GetRigDetails(int64(rentals.RigId))
	if err != nil {
		return "na", err
	}
	return rig.Status, nil
}

// QueryStatus is a StringItemHandlerFunc for key `mrr.status` which returns the status
// of a rentals.
func QueryStatus(request []string) (string, error) {
	// parse first param as int64
	rentalid, err := strconv.ParseInt(request[2], 10, 64)
	if err != nil {
		return "na", errors.New("Invalid rentalid format")
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "mrr-" + request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := miningrigrentals.New(request[0], request[1])
	rentals, err := client.GetRentalDetails(rentalid)
	if err != nil {
		return "na", err
	}
	return rentals.Status, nil
}

// QueryLeft is a Uint64ItemHandlerFunc for key `mrr.left` which returns the left time
// of a rentals.
func QueryLeft(request []string) (uint64, error) {
	// parse first param as int64
	rentalid, err := strconv.ParseInt(request[2], 10, 64)
	if err != nil {
		return 0, errors.New("Invalid rentalid format")
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "mrr-" + request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := miningrigrentals.New(request[0], request[1])
	rentals, err := client.GetRentalDetails(rentalid)
	if err != nil {
		return 0, err
	}
	return uint64(rentals.Left * 60), nil
}

// QuerySpeed is a DoubleItemHandlerFunc for key `mrr.speedpercent` which returns the speed percentage
// for a rentals.
func QuerySpeed(request []string) (float64, error) {
	// parse first param as int64
	rentalid, err := strconv.ParseInt(request[2], 10, 64)
	if err != nil {
		return 0.00, errors.New("Invalid rentalid format")
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "mrr-" + request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := miningrigrentals.New(request[0], request[1])
	rentals, err := client.GetRentalDetails(rentalid)
	if err != nil {
		return 0.00, err
	}
	var speedpercent float64
	if speedpercent = 0.00; rentals.HashRate.HashRate5m > 0 {
		speedpercent = 100.0 * float64(rentals.HashRate.Advertised) / rentals.HashRate.HashRate5m
	}
	//	fmt.Printf("%T %+v\n", rentals, rentals)
	return speedpercent, nil
}

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.Arg(0) {
	case "discovery":
		switch flag.NArg() {
		case 3:
			if v, err := DiscoverRentals(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v.Json())
			}
		default:
			log.Fatalf("Usage: %s discovery KEY SECRET", os.Args[0])
		}
	case "rigstatus":
		switch flag.NArg() {
		case 4:
			if v, err := QueryRigStatus(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s rigstatus KEY SECRET RENTALID", os.Args[0])
		}
	case "status":
		switch flag.NArg() {
		case 4:
			if v, err := QueryStatus(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s status KEY SECRET RENTALID", os.Args[0])
		}
	case "left":
		switch flag.NArg() {
		case 4:
			if v, err := QueryLeft(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s left KEY SECRET RENTALID", os.Args[0])
		}
	case "speedpercent":
		switch flag.NArg() {
		case 4:
			if v, err := QuerySpeed(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s speedpercent KEY SECRET RENTALID", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: 'discovery', 'status' , 'rigstatus' or 'speedpercent'.")
	}
}