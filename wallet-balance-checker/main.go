package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"strconv"
	"strings"
)

type BitcoinConfig struct {
	Hostname string `ini:"rpcconnect"`
	Port     int    `ini:"rpcport"`
	Username string `ini:"rpcuser"`
	Password string `ini:"rpcpassword"`
}

func FatalErr(err error, str string) {
	if err != nil {
		log.Fatalf("%s: %s", str, err.Error())
	}
}

type balanceData map[string]interface{}

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	if flag.NArg() == 0 || len(flag.Arg(0)) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	configPath := flag.Arg(0)
	fi, err := os.Stat(configPath)
	FatalErr(err, "Failed to check config file")

	if fi.Size() == 0 {
		log.Fatal("empty file")
	}
	config := &BitcoinConfig{Hostname: "127.0.0.1", Port: 8232}
	err = ini.MapTo(config, configPath)
	FatalErr(err, "Failed to load config file")

	connCfg := &rpcclient.ConnConfig{
		Host:         config.Hostname + ":" + strconv.Itoa(config.Port),
		User:         config.Username,
		Pass:         config.Password,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	client, err := rpcclient.New(connCfg, nil)
	FatalErr(err, "Failed to connect to wallet")
	defer client.Shutdown()

	addresses, err := client.ListAddressGroupings()
	FatalErr(err, "Failed to get addresses")

	balances := make(map[string]float64)
	for _, a := range addresses {
		if len(a.Account) == 0 || a.Amount == 0 {
			continue
		}
		account := a.Account
		if idx := strings.Index(account, "-"); idx != -1 {
			// truncate account name at "-"
			account = account[:idx]
		}
		balances[account] += a.Amount
	}
	result := make([]balanceData, 0)
	for name, balance := range balances {
		result = append(result, balanceData{"name": name, "balance": balance})
	}
	d, err := json.Marshal(result)
	FatalErr(err, "Failed to marshal balances")
	fmt.Print(string(d))
}
