package main

import (
	"github.com/bitbandi/go-nicehash-api"
	"github.com/Elbandi/zabbix-checker/common/lld"
	"github.com/Elbandi/zabbix-checker/common/filemutex"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"golang.org/x/net/proxy"
)

const defaultUserAgent = "nicehash-checker/1.0"

var (
	// Errors
	ErrInvalidOrderId  = errors.New("Invalid orderid format")
	ErrInvalidAlgo     = errors.New("Invalid algo format")
	ErrInvalidLocation = errors.New("Invalid location format")
	ErrOrderNotFound   = errors.New("Order not found")

	// flags
	debug     bool
	userAgent string
)

func FindOrder(id uint64, orders []nicehash.MyOrders) *nicehash.MyOrders {
	for _, order := range orders {
		if order.Id == id {
			return &order
		}
	}
	return nil
}

// DiscoverOrders is a DiscoveryItemHandlerFunc for key `nicehash.discovery` which returns JSON
// encoded discovery data for all orders
func DiscoverOrders(request []string) (lld.DiscoveryData, error) {
	// init discovery data
	d := make(lld.DiscoveryData, 0)
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "nicehash-"+request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := nicehash.NewNicehashClient(nil, "", request[0], request[1], userAgent)
	client.SetDebug(debug)
	orders, err := client.GetMyOrders(0, 0)
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		item := make(lld.DiscoveryItem, 0)
		item["ID"] = strconv.FormatUint(order.Id, 10)
		item["TYPE"] = order.Type.ToString()
		item["ALGO"] = strconv.FormatUint(uint64(order.Algo), 0)
		//		item["LOCATION"] = strconv.FormatUint(order.Location, 0)
		item["NAME"] = fmt.Sprintf("%s #%d", order.Type.ToString()[0], order.Id)
		d = append(d, item)
	}
	return d, nil
}

// QueryProfitability is a DoubleItemHandlerFunc for key `nicehash.profitability` which returns the paying price
// for algo.
func QueryProfitability(request []string) (float64, error) {
	// parse third param as uint64
	algo, err := strconv.ParseUint(request[0], 10, 64)
	if err != nil {
		return 0.00, ErrInvalidAlgo
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "nicehash-"+request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := nicehash.NewNicehashClient(nil, "", "", "", userAgent)
	client.SetDebug(debug)
	stats, err := client.GetStatsGlobalCurrent()
	if err != nil {
		return 0.00, err
	}
	for _, stat := range stats {
		if stat.Algo == nicehash.AlgoType(algo) {
			return stat.Price, nil
		}
	}
	return 0.00, nil
}

// QueryLowPrice is a DoubleItemHandlerFunc for key `nicehash.lowprice` which returns the lowest price
// for public orders.
func QueryLowPrice(request []string) (float64, error) {
	// parse third param as uint64
	algo, err := strconv.ParseUint(request[0], 10, 64)
	if err != nil {
		return 0.00, ErrInvalidAlgo
	}
	// parse third param as uint64
	location, err := strconv.ParseUint(request[1], 10, 64)
	if err != nil {
		return 0.00, ErrInvalidLocation
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "nicehash-"+request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := nicehash.NewNicehashClient(nil, "", "", "", userAgent)
	client.SetDebug(debug)
	orders, err := client.GetOrders(nicehash.AlgoType(algo), nicehash.Location(location))
	if err != nil {
		return 0.00, err
	}
	MinPrice := math.MaxFloat64
	for _, order := range orders {
		if order.Alive && order.Workers > 0 && order.Price < MinPrice {
			MinPrice = order.Price
		}
	}
	return MinPrice, nil
}

// QueryPrice is a DoubleItemHandlerFunc for key `nicehash.price` which returns the current price
// for a orders.
func QueryPrice(request []string) (float64, error) {
	// parse third param as uint64
	orderid, err := strconv.ParseUint(request[2], 10, 64)
	if err != nil {
		return 0.00, ErrInvalidOrderId
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "nicehash-"+request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := nicehash.NewNicehashClient(nil, "", request[0], request[1], userAgent)
	client.SetDebug(debug)
	orders, err := client.GetMyOrders(0, 0)
	if err != nil {
		return 0.00, err
	}
	order := FindOrder(orderid, orders)
	if order == nil {
		return 0.00, ErrOrderNotFound
	}
	return order.Price, nil
}

// QuerySpeed is a DoubleItemHandlerFunc for key `nicehash.btcavail` which returns the speed percentage
// for a orders.
func QueryBtcAvail(request []string) (float64, error) {
	// parse third param as uint64
	orderid, err := strconv.ParseUint(request[2], 10, 64)
	if err != nil {
		return 0.00, ErrInvalidOrderId
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "nicehash-"+request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := nicehash.NewNicehashClient(nil, "", request[0], request[1], userAgent)
	client.SetDebug(debug)
	orders, err := client.GetMyOrders(0, 0)
	if err != nil {
		return 0.00, err
	}
	order := FindOrder(orderid, orders)
	if order == nil {
		return 0.00, ErrOrderNotFound
	}
	return order.BtcAvail, nil
}

// QueryStatus is a StringItemHandlerFunc for key `nicehash.status` which returns the status
// of a orders.
func QueryStatus(request []string) (string, error) {
	// parse third param as uint64
	orderid, err := strconv.ParseUint(request[2], 10, 64)
	if err != nil {
		return "na", ErrInvalidOrderId
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "nicehash-"+request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := nicehash.NewNicehashClient(nil, "", request[0], request[1], userAgent)
	client.SetDebug(debug)
	orders, err := client.GetMyOrders(0, 0)
	if err != nil {
		return "na", err
	}
	order := FindOrder(orderid, orders)
	if order == nil {
		return "na", ErrOrderNotFound
	}
	if order.Alive {
		return "Alive", nil
	}
	return "Dead", nil
}

// QuerySpeed is a DoubleItemHandlerFunc for key `nicehash.speedpercent` which returns the speed percentage
// for a orders.
func QuerySpeed(request []string) (float64, error) {
	// parse third param as uint64
	orderid, err := strconv.ParseUint(request[2], 10, 64)
	if err != nil {
		return 0.00, ErrInvalidOrderId
	}
	lock := filemutex.MakeFileMutex(filepath.Join(os.TempDir(), "nicehash-"+request[0]))
	lock.Lock()
	defer lock.Unlock()
	client := nicehash.NewNicehashClient(nil, "", request[0], request[1], userAgent)
	client.SetDebug(debug)
	orders, err := client.GetMyOrders(0, 0)
	if err != nil {
		return 0.00, err
	}
	order := FindOrder(orderid, orders)
	if order == nil {
		return 0.00, ErrOrderNotFound
	}
	var speedpercent float64
	if speedpercent = 0.00; order.LimitSpeed > 0 {
		speedpercent = 100.0 * float64(order.AcceptedSpeed) / order.LimitSpeed
	}
	return speedpercent, nil
}

func main() {
	proxyPtr := flag.String("proxy", "", "socks proxy")
	flag.BoolVar(&debug, "debug", false, "enable request/response dump")
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
		case 3:
			if v, err := DiscoverOrders(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v.Json())
			}
		default:
			log.Fatalf("Usage: %s discovery APIID APIKEY", os.Args[0])
		}
	case "profitability":
		switch flag.NArg() {
		case 2:
			if v, err := QueryProfitability(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s price ALGO", os.Args[0])
		}
	case "lowprice":
		switch flag.NArg() {
		case 3:
			if v, err := QueryLowPrice(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s price ALGO LOCATION", os.Args[0])
		}
	case "price":
		switch flag.NArg() {
		case 4:
			if v, err := QueryPrice(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s price APIID APIKEY ORDERID", os.Args[0])
		}
	case "btcavail":
		switch flag.NArg() {
		case 4:
			if v, err := QueryBtcAvail(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s btcavail APIID APIKEY ORDERID", os.Args[0])
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
			log.Fatalf("Usage: %s status APIID APIKEY ORDERID", os.Args[0])
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
			log.Fatalf("Usage: %s speedpercent APIID APIKEY ORDERID", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: 'discovery', 'profitability', 'lowprice', 'price', 'status', 'btcavail' or 'speedpercent'.")
	}
}
