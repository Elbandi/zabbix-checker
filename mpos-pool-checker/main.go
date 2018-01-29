package main

import (
	"github.com/Elbandi/zabbix-checker/common/lld"
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"os"
	"github.com/bitbandi/go-mpos-api"
	"golang.org/x/net/proxy"
)

const defaultUserAgent = "mpos-pool-checker/1.0"

var (
	// flags
	debug bool
	output string
	userAgent string
)

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
		if len(fields) < 5 {
			continue
		}
		if strings.TrimSpace(fields[1]) != "MPOS" {
			continue
		}
		item := make(lld.DiscoveryItem, 0)
		item["NAME"] = strings.TrimSpace(fields[0])
		item["TYPE"] = strings.TrimSpace(fields[1])
		item["HOST"] = strings.TrimSpace(fields[2])
		item["APIKEY"] = strings.TrimSpace(fields[3])
		item["PROXY"] = strings.TrimSpace(fields[4])
		if len(fields) > 5 {
			item["LOW_POOL_LIMIT"] = strings.TrimSpace(fields[5])
		}
		if len(fields) > 6 {
			item["HIGH_POOL_LIMIT"] = strings.TrimSpace(fields[6])
		}
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
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	return uint64(status.Hashrate * 1000), nil
}

// PoolWorker is a Uint32ItemHandlerFunc for key `mpos.pool_workers` which returns the pool workers
// counter.
func PoolWorkers(request []string) (uint32, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	return status.Workers, nil
}

// PoolEfficiency is a DoubleItemHandlerFunc for key `mpos.pool_efficiency` which returns the pool efficiency
// ratio.
func PoolEfficiency(request []string) (float64, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetPoolStatus()
	if err != nil {
		return 0.00, err
	}
	return status.Efficiency, nil
}

// PoolLastBlock is a Uint32ItemHandlerFunc for key `mpos.pool_lastblock` which returns the pool last block
// height.
func PoolLastBlock(request []string) (uint32, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	return status.LastBlock, nil
}

// PoolLastBlock is a Uint32ItemHandlerFunc for key `mpos.pool_nextblock` which returns the pool next block
// height.
func PoolNextBlock(request []string) (uint32, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetPoolStatus()
	if err != nil {
		return 0, err
	}
	return status.NextNetworkBlock, nil
}


// UserHashrate is a Uint64ItemHandlerFunc for key `mpos.user_hashrate` which returns the user hashrate
// counter.
func UserHashrate(request []string) (uint64, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetUserStatus()
	if err != nil {
		return 0, err
	}
	return uint64(status.Hashrate * 1000), nil
}

// UserSharerate is a DoubleItemHandlerFunc for key `mpos.user_sharerate` which returns the user sharerate
// counter.
func UserSharerate(request []string) (float64, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetUserStatus()
	if err != nil {
		return 0.00, err
	}
	return status.Sharerate, nil
}

// UserSharesValid is a DoubleItemHandlerFunc for key `mpos.user_shares_valid` which returns the user valid
// shares.
func UserSharesValid(request []string) (float64, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetUserStatus()
	if err != nil {
		return 0.00, err
	}
	return status.Shares.Valid, nil
}

// UserSharesInvalid is a DoubleItemHandlerFunc for key `mpos.user_shares_invalid` which returns the user invalid
// shares.
func UserSharesInvalid(request []string) (float64, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetUserStatus()
	if err != nil {
		return 0.00, err
	}
	return status.Shares.Invalid, nil
}

// UserBalanceConfirmed is a DoubleItemHandlerFunc for key `mpos.user_balance_confirmed` which returns the user
// confirmed balance.
func UserBalanceConfirmed(request []string) (float64, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetUserBalance()
	if err != nil {
		return 0.00, err
	}
	return status.Confirmed, nil
}

// UserBalanceConfirmed is a DoubleItemHandlerFunc for key `mpos.user_balance_unconfirmed` which returns the user
// unconfirmed balance.
func UserBalanceUnconfirmed(request []string) (float64, error) {
	mposClient := mpos.NewMposClient(nil, request[0], request[1], userAgent)
	mposClient.SetDebug(debug)
	status, err := mposClient.GetUserBalance()
	if err != nil {
		return 0.00, err
	}
	return status.Unconfirmed, nil
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
			log.Fatalf("Usage: %s pool_hashrate URL APIKEY", os.Args[0])
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
			log.Fatalf("Usage: %s pool_workers URL APIKEY", os.Args[0])
		}
	case "pool_efficiency":
		switch flag.NArg() {
		case 3:
			if v, err := PoolEfficiency(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s pool_efficiency URL APIKEY", os.Args[0])
		}
	case "pool_lastblock":
		switch flag.NArg() {
		case 3:
			if v, err := PoolLastBlock(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s pool_lastblock URL APIKEY", os.Args[0])
		}
	case "pool_nextblock":
		switch flag.NArg() {
		case 3:
			if v, err := PoolNextBlock(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s pool_nextblock URL APIKEY", os.Args[0])
		}
	case "user_hashrate":
		switch flag.NArg() {
		case 3:
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
			log.Fatalf("Usage: %s user_hashrate URL APIKEY", os.Args[0])
		}
	case "user_sharerate":
		switch flag.NArg() {
		case 3:
			if v, err := UserSharerate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s user_sharerate URL APIKEY", os.Args[0])
		}
	case "user_shares_valid":
		switch flag.NArg() {
		case 3:
			if v, err := UserSharesValid(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s user_shares_valid URL APIKEY", os.Args[0])
		}
	case "user_shares_invalid":
		switch flag.NArg() {
		case 3:
			if v, err := UserSharesInvalid(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s user_shares_invalid URL APIKEY", os.Args[0])
		}
	case "user_balance_confirmed":
		switch flag.NArg() {
		case 3:
			if v, err := UserBalanceConfirmed(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s user_balance_confirmed URL APIKEY", os.Args[0])
		}
	case "user_balance_unconfirmed":
		switch flag.NArg() {
		case 3:
			if v, err := UserBalanceUnconfirmed(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s user_balance_unconfirmed URL APIKEY", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: " +
			"'discovery', " +
			"'pool_hashrate', 'pool_workers', 'pool_efficiency', 'pool_lastblock', 'pool_nextblock', " +
			"'user_hashrate', 'user_sharerate', 'user_shares_valid', 'user_shares_invalid', " +
			"'user_balance_confirmed' or 'user_balance_unconfirmed'.")

	}
}
