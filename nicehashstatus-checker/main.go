package main

import (
	"github.com/bitbandi/go-nicehash-api"
	"github.com/Elbandi/zabbix-checker/common/lld"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
	"golang.org/x/net/http2"
)

const defaultUserAgent = "nicehash-checker/1.0"

var (
	ApiId string
	ApiKey string
	UpdateInterval uint
	hostname string
	baseurl string
	debug bool
	userAgent string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Print debug infos")
	flag.StringVar(&baseurl, "base", "https://www.nicehash.com", "nicehash base domain")
	flag.StringVar(&ApiId, "apiid", "", "Nicehash api id")
	flag.StringVar(&ApiKey, "apikey", "", "Nicehash api key")
	flag.UintVar(&UpdateInterval, "updateinterval", 300, "Update interval")
	flag.StringVar(&hostname, "hostname", "", "zabbix hostname")
	flag.StringVar(&userAgent, "user-agent", defaultUserAgent, "http client user agent")
}

type myTransport struct {
	proxyURL *url.URL
	rt       *http.Transport
}

func (t *myTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", userAgent)
	r.Header.Set("Cache-Control", "max-age=0")
	r.Header.Set("Accept-Language", "en-us")
	t.rt.TLSClientConfig.InsecureSkipVerify = true
	t.rt.TLSClientConfig.ServerName = r.Host
	t.rt.Proxy = func(req *http.Request) (*url.URL, error) {
		return t.proxyURL, nil
	}
	response, err := t.rt.RoundTrip(r)
	if err != nil {
		return response, err
	}
	if response.StatusCode == 403 && response.Header.Get("Cf-Chl-Bypass") != "" { // cloudflare captcha
		tempHost := r.URL.Host
		r.URL.Host = "cflaresuje2rb7w2u3w43pn4luxdi6o7oatv6r2zrfb5xvsugj35d2qd.onion"
		response, err = t.rt.RoundTrip(r)
		r.URL.Host = tempHost
		response.Request.URL.Host = tempHost
	}
	return response, err
}


func main() {
	var allorders []nicehash.MyOrders

	proxyPtr := flag.String("proxy", "", "socks proxy")
	flag.Parse()
	log.SetOutput(os.Stderr)

	if *proxyPtr != "" {
		proxyURL, err := url.Parse("socks5://" + *proxyPtr)
		if err != nil {
			log.Fatalf("Failed to parse proxy URL: %v", err)
		}
		log.Printf("Set proxy to %s", proxyURL)
		transport, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			log.Fatalf("Failed to get the default http transport")
		}
		http2.ConfigureTransport(transport)
		http.DefaultTransport = &myTransport{
			proxyURL: proxyURL,
			rt:       transport,
		}
	}

	if len(ApiKey) == 0 || len(ApiKey) == 0 {
		log.Fatalf("No api id/key specified")
	}

	var pairs []struct{ nicehash.AlgoType; nicehash.Location }

	client := nicehash.NewNicehashClient(nil, baseurl, ApiId, ApiKey, userAgent)
	client.SetDebug(debug)

	discovery := make(lld.DiscoveryData, 0)
	for loc := nicehash.LocationNiceHash; loc < nicehash.LocationMAX; loc ++ {
		for algo := nicehash.AlgoTypeScrypt; algo < nicehash.AlgoTypeMAX; algo++ {
			orders, err := client.GetMyOrders(algo, loc)
			if err != nil {
				continue
			}
			if len(orders) > 0 {
				pairs = append(pairs, struct{ nicehash.AlgoType; nicehash.Location }{algo, loc})
			}
			allorders = append(allorders, orders...)
			for _, order := range orders {
				item := make(lld.DiscoveryItem, 0)
				item["ID"] = strconv.FormatUint(order.Id, 10)
				item["TYPE"] = order.Type.ToString()
				item["ALGO"] = order.Algo.ToString()
				item["LOCATION"] = loc.ToString()
				item["HOST"] = order.PoolHost
				item["PORT"] = strconv.FormatUint(uint64(order.PoolPort), 10)
				item["USER"] = order.PoolUser
				item["SPEED"] = strconv.FormatFloat(order.LimitSpeed, 'f', -1, 64)
				item["NAME"] = fmt.Sprintf("%c #%d", order.Type.ToString()[0], order.Id)
				discovery = append(discovery, item)
			}
			time.Sleep(2 * time.Second)
		}
	}
	fmt.Printf("\"%s\" \"nicehash.discovery\" %s\n", hostname, strconv.Quote(discovery.JsonLine()))

	for _, order := range allorders {
		fmt.Printf("\"%s\" \"nicehash.price[%d]\" \"%f\"\n", hostname, order.Id, order.Price)
		fmt.Printf("\"%s\" \"nicehash.btcavail[%d]\" \"%f\"\n", hostname, order.Id, order.BtcAvail)
		if order.Alive {
			fmt.Printf("\"%s\" \"nicehash.status[%d]\" \"Alive\"\n", hostname, order.Id)
		} else {
			fmt.Printf("\"%s\" \"nicehash.status[%d]\" \"Dead\"\n", hostname, order.Id)
		}
		var speedpercent float64
		if speedpercent = 0.00; order.LimitSpeed > 0 {
			speedpercent = 100.0 * float64(order.AcceptedSpeed) / order.LimitSpeed
		}
		fmt.Printf("\"%s\" \"nicehash.speedpercent[%d]\" \"%f\"\n", hostname, order.Id, speedpercent)
	}

	for _, pair := range pairs {
		orders, err := client.GetOrders(pair.AlgoType, pair.Location)
		if err != nil {
			continue
		}
		minprice := math.MaxFloat64
		for _, order := range orders {
			if order.Alive && order.Workers > 0 && order.Price < minprice {
				minprice = order.Price
			}
		}
		if minprice < math.MaxFloat64 {
			fmt.Printf("\"%s\" \"nicehash.lowprice[%s,%s]\" \"%f\"\n", hostname, pair.Location.ToString(), pair.AlgoType.ToString(), minprice)
		}
	}
}
