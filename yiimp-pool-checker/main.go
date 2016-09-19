package main

import (
	"github.com/Elbandi/zabbix-checker/common/lld"
	"bufio"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"os"
	"github.com/bitbandi/go-yiimp-api"
	"golang.org/x/net/proxy"
)

const userAgent = "yiimp-pool-checker/1.0"

var debugPtr *bool

// DiscoverPools is a DiscoveryItemHandlerFunc for key `mpos.discovery` which returns JSON
// encoded discovery data for pool stored in a file
func DiscoverPools(request []string) (lld.DiscoveryData, error) {
	// init discovery data
	d := make(lld.DiscoveryData, 0)
	file, err := os.Open(request[0])
	if err != nil {
		return d, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line[0] == '#' {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) != 6 {
			continue
		}
		item := make(lld.DiscoveryItem, 0)
		item["NAME"] = strings.TrimSpace(fields[0])
		item["TYPE"] = strings.TrimSpace(fields[1])
		item["HOST"] = strings.TrimSpace(fields[2])
		item["ALGO"] = strings.TrimSpace(fields[3])
		item["ADDRESS"] = strings.TrimSpace(fields[4])
		item["PROXY"] = strings.TrimSpace(fields[5])
		d = append(d, item)
	}
	if err := scanner.Err(); err != nil {
		return d, err
	}
	return d, nil
}

// PoolHashrate is a Uint64ItemHandlerFunc for key `mpos.pool_hashrate` which returns the pool hashrate
// counter.
func PoolHashrate(request []string) (uint64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(*debugPtr)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, errors.New("no such algorithm")
	}
	return algo.Hashrate, nil
}

// PoolWorker is a Uint32ItemHandlerFunc for key `mpos.pool_workers` which returns the pool workers
// counter.
func PoolWorkers(request []string) (uint16, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(*debugPtr)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, errors.New("no such algorithm")
	}
	return algo.Workers, nil
}

// PoolEstimateCurrent is a DoubleItemHandlerFunc for key `mpos.pool_hashrate` which returns the pool hashrate
// counter.
func PoolEstimateCurrent(request []string) (float64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(*debugPtr)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, errors.New("no such algorithm")
	}
	return algo.EstimateCurrent, nil
}

// PoolEstimateLast24h is a DoubleItemHandlerFunc for key `mpos.pool_hashrate` which returns the pool hashrate
// counter.
func PoolEstimateLast24h(request []string) (float64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(*debugPtr)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, errors.New("no such algorithm")
	}
	return algo.EstimateLast24h, nil
}

// PoolActualLast24h is a DoubleItemHandlerFunc for key `mpos.pool_hashrate` which returns the pool hashrate
// counter.
func PoolActualLast24h(request []string) (float64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(*debugPtr)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, errors.New("no such algorithm")
	}
	return algo.ActualLast24h, nil
}

// PoolRentalCurrent is a DoubleItemHandlerFunc for key `mpos.pool_hashrate` which returns the pool hashrate
// counter.
func PoolRentalCurrent(request []string) (float64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(*debugPtr)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, errors.New("no such algorithm")
	}
	return algo.RentalCurrent, nil
}

func main() {
	proxyPtr := flag.String("proxy", "", "socks proxy")
	debugPtr = flag.Bool("debug", false, "enable request/response dump")
	flag.Parse()
	log.SetOutput(os.Stderr)

	if *proxyPtr != "" {
		proxyURL, err := url.Parse("socks5://" + *proxyPtr)
		if err != nil {
			log.Fatalf("Failed to parse proxy URL: %v", err)
		}
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			log.Fatalf("Failed to obtain proxy dialer: %v", err)
		}
		http.DefaultTransport = &http.Transport{
			Dial: dialer.Dial,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		log.Printf("Set proxy to %s", proxyURL)
	}

	switch flag.Arg(0) {
	case "discovery":
		switch flag.NArg() {
		case 2:
			if v, err := DiscoverPools(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v.Json())
			}
		default:
			log.Fatalf("Usage: %s discovery PATH", os.Args[0])
		}
	case "pool_hashrate":
		switch flag.NArg() {
		case 3:
			if v, err := PoolHashrate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s pool_hashrate URL ALGORITHM", os.Args[0])
		}
	case "pool_workers":
		switch flag.NArg() {
		case 3:
			if v, err := PoolWorkers(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s pool_workers URL ALGORITHM", os.Args[0])
		}
	case "estimate_current":
		switch flag.NArg() {
		case 3:
			if v, err := PoolEstimateCurrent(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s estimate_current URL ALGORITHM", os.Args[0])
		}
	case "estimate_last24h":
		switch flag.NArg() {
		case 3:
			if v, err := PoolEstimateLast24h(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s estimate_last24h URL ALGORITHM", os.Args[0])
		}
	case "actual_last24h":
		switch flag.NArg() {
		case 3:
			if v, err := PoolActualLast24h(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s actual_last24h URL ALGORITHM", os.Args[0])
		}
	case "rental_current":
		switch flag.NArg() {
		case 3:
			if v, err := PoolRentalCurrent(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s rental_current URL APIKEY", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: " +
			"'discovery', " +
			"'pool_hashrate', 'pool_workers', 'estimate_current', 'estimate_last24h', 'actual_last24h', " +
			"or 'rental_current'.")
	}
}