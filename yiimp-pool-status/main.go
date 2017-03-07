package main

import (
	"github.com/Elbandi/zabbix-checker/common/lld"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"github.com/bitbandi/go-yiimp-api"
)

const userAgent = "yiimp-status-checker/1.0"

func algo_mBTC_factor(algo string) (uint16) {
	switch algo {
	case "sha256",
		"blake",
		"blakecoin",
		"blake2s",
		"decred",
		"vanilla":
		return 1000;
	default:
		return 1;
	}
}
func main() {
	var hostname, url string
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable request/response dump")
	flag.StringVar(&hostname, "hostname", "", "zabbix hostname")
	flag.StringVar(&url, "url", "", "pool url")
	flag.Parse()
	log.SetOutput(os.Stderr)

	if hostname == "" || url == "" {
		flag.Usage()
		os.Exit(1)
	}

	discovery := make(lld.DiscoveryData, 0)
	yiimpClient := yiimp.NewYiimpClient(nil, url, "", userAgent)
	yiimpClient.SetDebug(debug)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
	for key, pool := range status {
		item := make(lld.DiscoveryItem, 0)
		item["NAME"] = strings.TrimSpace(pool.Name)
		item["ALGO"] = strings.TrimSpace(key)
		discovery = append(discovery, item)
	}
	fmt.Printf("\"%s\" \"yiimpstatus.discovery\" %s\n", hostname, strconv.Quote(discovery.JsonLine()))
	for _, element := range discovery {
		key := element["ALGO"]
		fmt.Printf("\"%s\" \"yiimpstatus.hashrate[%s]\" \"%f\"\n", hostname, element["ALGO"], status[key].Hashrate)
		fmt.Printf("\"%s\" \"yiimpstatus.hashrate24h[%s]\" \"%f\"\n", hostname, element["ALGO"], status[key].Hashrate24h)
		fmt.Printf("\"%s\" \"yiimpstatus.workers[%s]\" \"%d\"\n", hostname, element["ALGO"], status[key].Workers)
		btcmhday := status[key].ActualLast24h / 1e3 // float64(algo_mBTC_factor(key))
		fmt.Printf("\"%s\" \"yiimpstatus.btcmhday[%s]\" \"%f\"\n", hostname, element["ALGO"], btcmhday)
		btctotal := status[key].Hashrate24h * status[key].ActualLast24h / float64(algo_mBTC_factor(key)) / 1e9
		fmt.Printf("\"%s\" \"yiimpstatus.btctotal[%s]\" \"%f\"\n", hostname, element["ALGO"], btctotal)

	}
}
