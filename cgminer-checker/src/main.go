package main

import (
	"github.com/Elbandi/go-cgminer-api"
	"lld"
	"log"
	"flag"
	"fmt"
	"os"
	"strconv"
	"errors"
)

// var omitNewline = flag.Bool("n", false, "don't print final newline")

func QueryDevice(request []string) (*cgminer.Devs, error) {
	// parse first param as float64
	port, err := strconv.ParseInt(request[1], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid port format")
	}
	devid, err := strconv.ParseInt(request[2], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid deviceid format")
	}
	miner := cgminer.New(request[0], port)
	devices, err := miner.Devs()
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to CGMiner: %s", err.Error())
	}
	if int64(len(*devices)) <= devid {
		return nil, errors.New("Invalid device id")
	}
	/*
			for _, dev := range *devices {
					fmt.Printf("Dev %d temp: %f\n", dev.ID, dev.Temperature)
			}
			res2B, _ := json.Marshal((*devices)[0])
			fmt.Println(string(res2B))
	*/
	dev := (*devices)[devid];
	return &dev, nil
}

// DiscoverDevs is a DiscoveryItemHandlerFunc for key `cgminer.dev.discovery` which returns JSON
// encoded discovery data for all devices from cgminer
func DiscoverDevs(request []string) (lld.DiscoveryData, error) {
	// parse first param as float64
	port, err := strconv.ParseInt(request[1], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid port format")
	}
	// init discovery data
	d := make(lld.DiscoveryData, 0)

	miner := cgminer.New(request[0], port)
	devices, err := miner.Devs()
	if err != nil {
		return nil, err
	}

	item := make(lld.DiscoveryItem, 0)
	for _, dev := range *devices {
		item["ID"] = strconv.FormatInt(dev.ID, 10)
		item["NAME"] = fmt.Sprintf("%s %d", dev.Name, dev.ID)
		d = append(d, item)
	}

	return d, nil
}


// AcceptedShares is a DoubleItemHandlerFunc for key `cgminer.accept_shares` which returns the accepted shares
// counter.
func AcceptedShares(request []string) (float64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	return dev.DifficultyAccepted, nil
}

// HardwareErrors is a Uint64ItemHandlerFunc for key `cgminer.hwerrors` which returns the hardware errors
// counter.
func HardwareErrors(request []string) (uint64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0, err
	}
	return uint64(dev.HardwareErrors), nil
}

// Rate is a DoubleItemHandlerFunc for key `cgminer.hashrate` which returns
// the device 5 sec hashrate.
func Rate(request []string) (float64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	if (dev.KHS5s > 0) {
		return dev.KHS5s * 1000, nil
	}
	return dev.MHS5s * 1000 * 1000, nil
}

// RateAv is a DoubleItemHandlerFunc for key `cgminer.hashrate_av` which returns
// the device average hashrate.
func RateAv(request []string) (float64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	if (dev.KHSav > 0) {
		return dev.KHSav * 1000, nil
	}
	return dev.MHSav * 1000 * 1000, nil
}

// AcceptedShares is a DoubleItemHandlerFunc for key `cgminer.rejected` which returns the rejected shares
// counter.
func Rejected(request []string) (float64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	return dev.DeviceRejected, nil
}

// AcceptedShares is a DoubleItemHandlerFunc for key `cgminer.temperature` which returns the rejected shares
// counter.
func Temperature(request []string) (float64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	return dev.Temperature, nil
}


func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.Arg(0) {
	case "discovery":
		switch flag.NArg() {
		case 3:
			if v, err := DiscoverDevs(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v.Json())
			}
		default:
			log.Fatalf("Usage: %s discovery IP PORT", os.Args[0])
		}
	case "accept_shares":
		switch flag.NArg() {
		case 4:
			if v, err := AcceptedShares(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s accept_shares IP PORT DEVICEID", os.Args[0])
		}
	case "hwerrors":
		switch flag.NArg() {
		case 4:
			if v, err := HardwareErrors(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s hwerrors IP PORT DEVICEID", os.Args[0])
		}
	case "hashrate":
		switch flag.NArg() {
		case 4:
			if v, err := Rate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s hashrate IP PORT DEVICEID", os.Args[0])
		}
	case "hashrate_av":
		switch flag.NArg() {
		case 4:
			if v, err := RateAv(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s hashrate_av IP PORT DEVICEID", os.Args[0])
		}
	case "rejected":
		switch flag.NArg() {
		case 4:
			if v, err := Rejected(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s rejected IP PORT DEVICEID", os.Args[0])
		}
	case "temperature":
		switch flag.NArg() {
		case 4:
			if v, err := Temperature(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s temperature IP PORT DEVICEID", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: 'discovery', 'accept_shares', 'hwerrors', 'hashrate', 'hashrate_av', 'rejected' or 'temperature'.")
	}
}
