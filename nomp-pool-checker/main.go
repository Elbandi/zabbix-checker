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
	"github.com/bitbandi/go-nomp-api"
	"golang.org/x/net/proxy"
)

const userAgent = "nomp-pool-checker/1.0"

var (
	// Errors
	ErrPoolNotFound = errors.New("Pool not found")
	ErrWorkerNotFound = errors.New("Worker not found")

	// flags
	debugPtr *bool
	userAgentPtr *string
)

// DiscoverPools is a DiscoveryItemHandlerFunc for key `nomp.discovery` which returns JSON
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
		item["POOL"] = strings.TrimSpace(fields[3])
		item["WORKER"] = strings.TrimSpace(fields[4])
		item["PROXY"] = strings.TrimSpace(fields[5])
		d = append(d, item)
	}
	if err := scanner.Err(); err != nil {
		return d, err
	}
	return d, nil
}

// PoolHashrate is a DoubleItemHandlerFunc for key `nomp.pool_hashrate` which returns the pool hashrate
// counter.
func PoolHashrate(request []string) (float64, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0.00, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0.00, ErrPoolNotFound
	}
	return pool.Hashrate, nil
}

// PoolWorker is a Uint32ItemHandlerFunc for key `nomp.pool_workers` which returns the pool workers
// counter.
func PoolWorkers(request []string) (uint32, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0, ErrPoolNotFound
	}
	return uint32(pool.WorkerCount), nil
}

// PoolSharesValid is a Uint32ItemHandlerFunc for key `nomp.pool_shares_valid` which returns the pool valid
// shares.
func PoolSharesValid(request []string) (uint32, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0, ErrPoolNotFound
	}
	return pool.Stat.ValidShares, nil
}

// PoolSharesInvalid is a Uint32ItemHandlerFunc for key `nomp.pool_shares_invalid` which returns the pool invalid
// shares.
func PoolSharesInvalid(request []string) (uint32, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0, ErrPoolNotFound
	}
	return pool.Stat.InvalidShares, nil
}

// PoolPendingBlock is a Uint32ItemHandlerFunc for key `nomp.pool_blocks_pending` which returns the pool pending
// block count.
func PoolPendingBlock(request []string) (uint32, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0, ErrPoolNotFound
	}
	return uint32(pool.Blocks.Pending), nil
}

// PoolConfirmedBlock is a Uint32ItemHandlerFunc for key `nomp.pool_blocks_confirmed` which returns the pool confirmed
// block count.
func PoolConfirmedBlock(request []string) (uint32, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0, ErrPoolNotFound
	}
	return pool.Blocks.Confirmed, nil
}

// UserHashrate is a DoubleItemHandlerFunc for key `nomp.user_hashrate` which returns the user hashrate
// counter.
func UserHashrate(request []string) (float64, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0.00, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0.00, ErrPoolNotFound
	}
	worker, ok := pool.Workers[request[2]]
	if !ok {
		return 0.00, ErrWorkerNotFound
	}
	return worker.Hashrate, nil
}

// UserSharesValid is a DoubleItemHandlerFunc for key `nomp.user_shares_valid` which returns the user valid
// shares.
func UserSharesValid(request []string) (float64, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0.00, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0.00, ErrPoolNotFound
	}
	worker, ok := pool.Workers[request[2]]
	if !ok {
		return 0.00, ErrWorkerNotFound
	}
	return worker.Shares, nil
}

// UserSharesInvalid is a DoubleItemHandlerFunc for key `nomp.user_shares_invalid` which returns the user invalid
// shares.
func UserSharesInvalid(request []string) (float64, error) {
	nompClient := nomp.NewNompClient(nil, request[0], *userAgentPtr)
	nompClient.SetDebug(*debugPtr)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return 0.00, err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return 0.00, ErrPoolNotFound
	}
	worker, ok := pool.Workers[request[2]]
	if !ok {
		return 0.00, ErrWorkerNotFound
	}
	return worker.InvalidShares, nil
}

func main() {
	proxyPtr := flag.String("proxy", "", "socks proxy")
	debugPtr = flag.Bool("debug", false, "enable request/response dump")
	userAgentPtr = flag.String("user-agent", userAgent, "http client user agent")
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
			log.Fatalf("Usage: %s pool_hashrate URL POOL", os.Args[0])
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
			log.Fatalf("Usage: %s pool_workers URL POOL", os.Args[0])
		}
	case "pool_blocks_pending":
		switch flag.NArg() {
		case 3:
			if v, err := PoolPendingBlock(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s pool_blocks_confirmed URL POOL", os.Args[0])
		}
	case "pool_blocks_confirmed":
		switch flag.NArg() {
		case 3:
			if v, err := PoolConfirmedBlock(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s pool_blocks_confirmed URL POOL", os.Args[0])
		}
	case "pool_shares_valid":
		switch flag.NArg() {
		case 3:
			if v, err := PoolSharesValid(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s pool_shares_valid URL POOL", os.Args[0])
		}
	case "pool_shares_invalid":
		switch flag.NArg() {
		case 3:
			if v, err := PoolSharesInvalid(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s pool_shares_invalid URL POOL", os.Args[0])
		}
	case "user_hashrate":
		switch flag.NArg() {
		case 4:
			if v, err := UserHashrate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s user_hashrate URL POOL WORKER", os.Args[0])
		}
	case "user_shares_valid":
		switch flag.NArg() {
		case 4:
			if v, err := UserSharesValid(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s user_shares_valid URL POOL WORKER", os.Args[0])
		}
	case "user_shares_invalid":
		switch flag.NArg() {
		case 4:
			if v, err := UserSharesInvalid(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s user_shares_invalid URL POOL WORKER", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: " +
			"'discovery', " +
			"'pool_hashrate', 'pool_workers', 'pool_blocks_pending', 'pool_blocks_confirmed', " +
			"'pool_shares_valid', 'pool_shares_invalid', " +
			"'user_hashrate', 'user_shares_valid', 'user_shares_invalid'.")

	}
}
