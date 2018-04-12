package main

import (
	"github.com/Elbandi/zabbix-checker/common/lld"
	"bufio"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"os"
	"github.com/bitbandi/go-yiimp-api"
	"golang.org/x/net/proxy"
)

const defaultUserAgent = "yiimp-pool-checker/1.0"

var (
	// Errors
	ErrAlgoNotFound   = errors.New("no such algorithm")
	ErrWorkerNotFound = errors.New("Worker not found")

	// flags
	debug     bool
	output    string
	userAgent string
)

// DiscoverPools is a DiscoveryItemHandlerFunc for key `yiimp.discovery` which returns JSON
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
		if len(fields) < 6 {
			continue
		}
		if strings.TrimSpace(fields[1]) != "YIIMP" {
			continue
		}
		item := make(lld.DiscoveryItem, 0)
		item["NAME"] = strings.TrimSpace(fields[0])
		item["TYPE"] = strings.TrimSpace(fields[1])
		item["HOST"] = strings.TrimSpace(fields[2])
		item["ALGO"] = strings.TrimSpace(fields[3])
		item["ADDRESS"] = strings.TrimSpace(fields[4])
		item["PROXY"] = strings.TrimSpace(fields[5])
		if len(fields) > 6 {
			item["LOW_POOL_LIMIT"] = strings.TrimSpace(fields[6])
		}
		if len(fields) > 7 {
			item["HIGH_POOL_LIMIT"] = strings.TrimSpace(fields[7])
		}
		d = append(d, item)
	}
	if err := scanner.Err(); err != nil {
		return d, err
	}
	return d, nil
}

// PoolHashrate is a Uint64ItemHandlerFunc for key `yiimp.pool_hashrate` which returns the pool hashrate
// counter.
func PoolHashrate(request []string) (uint64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(debug)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, ErrAlgoNotFound
	}
	return algo.Hashrate, nil
}

// PoolWorker is a Uint32ItemHandlerFunc for key `yiimp.pool_workers` which returns the pool workers
// counter.
func PoolWorkers(request []string) (uint16, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(debug)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, ErrAlgoNotFound
	}
	return algo.Workers, nil
}

// PoolEstimateCurrent is a DoubleItemHandlerFunc for key `yiimp.pool_estimate_current` which returns the pool estimate current
// price value.
func PoolEstimateCurrent(request []string) (float64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(debug)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, ErrAlgoNotFound
	}
	return algo.EstimateCurrent, nil
}

// PoolEstimateLast24h is a DoubleItemHandlerFunc for key `yiimp.pool_estimate_last24h` which returns the pool estimate last 24h
// price value.
func PoolEstimateLast24h(request []string) (float64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(debug)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, ErrAlgoNotFound
	}
	return algo.EstimateLast24h, nil
}

// PoolActualLast24h is a DoubleItemHandlerFunc for key `yiimp.pool_actual_last24h` which returns the pool actual last 24h
// price value.
func PoolActualLast24h(request []string) (float64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(debug)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, ErrAlgoNotFound
	}
	return algo.ActualLast24h, nil
}

// PoolRentalCurrent is a DoubleItemHandlerFunc for key `yiimp.pool_rental` which returns the pool current rental
// price value.
func PoolRentalCurrent(request []string) (float64, error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(debug)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return 0.00, err
	}
	algo, ok := status[request[1]]
	if !ok {
		return 0.00, ErrAlgoNotFound
	}
	return algo.RentalCurrent, nil
}

// UserHashrate is a DoubleItemHandlerFunc for key `yiimp.user_hashrate` which returns the user hashrate
// counter.
func UserHashrate(request []string) (float64, error) {
	mposClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetWalletEx(request[1])
	if err != nil {
		return 0.00, err
	}
	var hashrate float64 = 0.00
	for _, miner := range status.Miners {
		if miner.Algo == request[2] {
			hashrate += miner.Accepted
		}
	}
	return hashrate, nil
}

func main() {
	proxyPtr := flag.String("proxy", "", "socks proxy")
	flag.BoolVar(&debug, "debug", false, "enable request/response dump")
	flag.StringVar(&output, "output", "", "output the result to file")
	flag.StringVar(&userAgent, "user-agent", defaultUserAgent, "http client user agent")
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
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v.Json())), 0644)
				} else {
					fmt.Print(v.Json())
				}
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
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
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
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
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
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
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
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
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
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
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
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s rental_current URL APIKEY", os.Args[0])
		}
	case "user_hashrate":
		switch flag.NArg() {
		case 4:
			if v, err := UserHashrate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s user_hashrate URL ADDRESS ALGORITHM", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: " +
			"'discovery', " +
			"'pool_hashrate', 'pool_workers', 'estimate_current', 'estimate_last24h', 'actual_last24h', " +
			"'user_hashrate' or 'rental_current'.")
	}
}