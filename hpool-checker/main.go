package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bitbandi/go-hpool"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultUserAgent = "hpool-pool-checker/1.0"

var (
	// flags
	debug     bool
	tokenFile string
	output    string
	userAgent string
)

type Pool struct {
	ApiKey                          string  `json:"api_key"`
	BlockReward                     float64 `json:"block_reward"`
	BlockTime                       int     `json:"block_time"`
	Capacity                        int     `json:"capacity"`
	Coin                            string  `json:"coin"`
	DepositMortgageBalance          float64 `json:"deposit_mortgage_balance"`
	DepositMortgageEffectiveBalance float64 `json:"deposit_mortgage_effective_balance"`
	DepositMortgageFreeBalance      float64 `json:"deposit_mortgage_free_balance"`
	DepositRate                     float64 `json:"deposit_rate"`
	Fee                             float64 `json:"fee"`
	LoanMortgageBalance             float64 `json:"loan_mortgage_balance"`
	Mortgage                        float64 `json:"mortgage"`
	Name                            string  `json:"name"`
	Offline                         int     `json:"offline"`
	Online                          int     `json:"online"`
	PaymentTime                     string  `json:"payment_time"`
	PointDepositBalance             float64 `json:"point_deposit_balance"`
	PoolAddress                     string  `json:"pool_address"`
	PoolIncome                      float64 `json:"pool_income"`
	PoolType                        string  `json:"pool_type"`
	PreviousIncomePb                float64 `json:"previous_income_pb"`
	TheoryMortgageBalance           float64 `json:"theory_mortgage_balance"`
	Type                            string  `json:"type"`
	UndistributedIncome             float64 `json:"undistributed_income"`
}

// PoolDetail is a StringItemHandlerFunc for key `hpool.pool` which returns the pool details.
func PoolDetail(request []string) (string, error) {
	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return "{}", err
	}
	hpoolClient := hpool.New(strings.TrimSpace(string(token)))
	hpoolClient.SetDebug(debug)
	details, err := hpoolClient.PoolDetail(request[0])
	if err != nil {
		return "{}", err
	}
	bytes, err := json.Marshal(Pool(details))
	return string(bytes), err
}

// Miners is a CustomItemHandlerFunc for key `hpool.miners` which returns the pool miners.
func Miners(request []string) (string, error) {
	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return "{}", err
	}
	hpoolClient := hpool.New(strings.TrimSpace(string(token)))
	hpoolClient.SetDebug(debug)
	miners, err := hpoolClient.Miners(request[0])
	if err != nil {
		return "{}", err
	}
	bytes, err := json.Marshal(miners)
	return string(bytes), err
}

func main() {
	proxyPtr := flag.String("proxy", "", "socks proxy")
	flag.BoolVar(&debug, "debug", false, "enable request/response dump")
	flag.StringVar(&tokenFile, "token", "", "token file")
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
	case "pool":
		switch flag.NArg() {
		case 2:
			if v, err := PoolDetail(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					if err = ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644); err != nil {
						log.Fatalf("Error: %s", err.Error())
					}
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s pool POOL", os.Args[0])
		}
	case "workers":
		switch flag.NArg() {
		case 2:
			if v, err := Miners(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					if err = ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644); err != nil {
						log.Fatalf("Error: %s", err.Error())
					}
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s workers POOL", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: " +
			"'pool', 'workers'.")
	}
}
