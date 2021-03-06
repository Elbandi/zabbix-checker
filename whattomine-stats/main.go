package main

import (
	"golang.org/x/net/proxy"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const defaultUserAgent = "wtm-stats/1.0"

var (
	// flags
	debug     bool
	output    string
	userAgent string
)

// ExchangeRate is a DoubleItemHandlerFunc for key `wtm.exchange_rate` which returns the current exchange rate
// for coin.
func ExchangeRate(request []string) (float64, error) {
	coinId, err := strconv.ParseUint(request[0], 10, 64)
	if err != nil {
		return 0.00, errors.New("Invalid coinid format")
	}

	wtmClient := NewWhatToMineClient(nil, BASE, userAgent)
	wtmClient.SetDebug(debug)
	status, err := wtmClient.GetCoin(coinId, 1000000, 0, 0)
	if err != nil {
		return 0.00, err
	}
	return status.ExchangeRate, nil
}

// ExchangeRate24 is a DoubleItemHandlerFunc for key `wtm.exchange_rate24` which returns the day exchange rate
// for coin.
func ExchangeRate24(request []string) (float64, error) {
	coinId, err := strconv.ParseUint(request[0], 10, 64)
	if err != nil {
		return 0.00, errors.New("Invalid coinid format")
	}

	wtmClient := NewWhatToMineClient(nil, BASE, userAgent)
	wtmClient.SetDebug(debug)
	status, err := wtmClient.GetCoin(coinId, 1000000, 0, 0)
	if err != nil {
		return 0.00, err
	}
	return status.ExchangeRate24, nil
}

// EstimatedRewards is a DoubleItemHandlerFunc for key `wtm.estimated_rewards` which returns the current estimated
// rewards.
func EstimatedRewards(request []string) (float64, error) {
	coinId, err := strconv.ParseUint(request[0], 10, 64)
	if err != nil {
		return 0.00, errors.New("Invalid coinid format")
	}

	wtmClient := NewWhatToMineClient(nil, BASE, userAgent)
	wtmClient.SetDebug(debug)
	status, err := wtmClient.GetCoin(coinId, 1000000, 0, 0)
	if err != nil {
		return 0.00, err
	}
	return status.EstimatedRewards, nil
}

// BtcRevenue is a DoubleItemHandlerFunc for key `wtm.btc_revenue` which returns the current possible revenue.
func BtcRevenue(request []string) (float64, error) {
	coinId, err := strconv.ParseUint(request[0], 10, 64)
	if err != nil {
		return 0.00, errors.New("Invalid coinid format")
	}

	wtmClient := NewWhatToMineClient(nil, BASE, userAgent)
	wtmClient.SetDebug(debug)
	status, err := wtmClient.GetCoin(coinId, 1000000, 0, 0)
	if err != nil {
		return 0.00, err
	}
	return status.BtcRevenue, nil
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
	case "exchange_rate":
		switch flag.NArg() {
		case 2:
			if v, err := ExchangeRate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s exchange_rate COIN", os.Args[0])
		}
	case "exchange_rate24":
		switch flag.NArg() {
		case 2:
			if v, err := ExchangeRate24(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s exchange_rate24 COIN", os.Args[0])
		}
	case "estimated_rewards":
		switch flag.NArg() {
		case 2:
			if v, err := EstimatedRewards(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s estimated_rewards COIN", os.Args[0])
		}
	case "btc_revenue":
		switch flag.NArg() {
		case 2:
			if v, err := BtcRevenue(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				if output != "" {
					ioutil.WriteFile(output, []byte(fmt.Sprint(v)), 0644)
				} else {
					fmt.Print(v)
				}
			}
		default:
			log.Fatalf("Usage: %s btc_revenue COIN", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: " +
		//			"'discovery', " +
			"'exchange_rate', 'exchange_rate24', " +
			"'estimated_rewards' or 'btc_revenue'.")

	}

}
