package main

import (
	"github.com/Elbandi/zabbix-checker/common/lld"
	"github.com/btcsuite/btcrpcclient"
	"gopkg.in/ini.v1"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"strconv"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type BitcoinConfig struct {
	Hostname string `ini:"rpcconnect"`
	Port     int    `ini:"rpcport"`
	Username string `ini:"rpcuser"`
	Password string `ini:"rpcpassword"`
}

func main() {
	var excludesearch arrayFlags
	hostnamePtr := flag.String("hostname", "", "zabbix hostname")
	basepathPtr := flag.String("basepath", "/srv", "base path")
	searchSuffixPtr := flag.String("searchSuffix", "-data", "search suffix")
	flag.Var(&excludesearch, "exclude", "exclude from search list")
	flag.Parse()
	log.SetOutput(os.Stderr)

	if *hostnamePtr == "" {
		flag.Usage()
		os.Exit(1)
	}

	discovery := make(lld.DiscoveryData, 0)
	err := filepath.Walk(*basepathPtr, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, *searchSuffixPtr) {
			name := strings.TrimSuffix(filepath.Base(path), *searchSuffixPtr)
			if stringInSlice(name, excludesearch) {
				return nil
			}
			item := make(lld.DiscoveryItem, 0)
			item["NAME"] = name
			item["PATH"] = path
			discovery = append(discovery, item)
		}
		return nil
	})
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	fmt.Printf("\"%s\" \"wallet.discovery\" %s\n", *hostnamePtr, strconv.Quote(discovery.JsonLine()))
	for _, element := range discovery {
		logPath := filepath.Join(element["PATH"], "debug.log")
		fi, err := os.Stat(logPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("\"%s\" \"vfs.file.size[%s]\" \"0\"\n", *hostnamePtr, logPath)
			} else {
				log.Print(err)
			}
		} else {
			fmt.Printf("\"%s\" \"vfs.file.size[%s]\" \"%d\"\n", *hostnamePtr, logPath, fi.Size())
		}

		config := &BitcoinConfig{Hostname:"127.0.0.1", Port:8332}
		err = ini.MapTo(config, filepath.Join(element["PATH"], element["NAME"] + ".conf"))
		if err != nil {
			log.Print(err)
			continue
		}
		connCfg := &btcrpcclient.ConnConfig{
			Host:         config.Hostname + ":" + strconv.Itoa(config.Port),
			User:         config.Username,
			Pass:         config.Password,
			HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   true, // Bitcoin core does not provide TLS by default
		}
		// Notice the notification parameter is nil since notifications are
		// not supported in HTTP POST mode.
		client, err := btcrpcclient.New(connCfg, nil)
		if err != nil {
			log.Print(err)
			continue
		}
		defer client.Shutdown()

		// Get the current block count.
		blockCount, err := client.GetBlockCount()
		if err != nil {
			log.Print(err)
			continue
		}
		fmt.Printf("\"%s\" \"wallet.blocks[%s]\" \"%d\"\n", *hostnamePtr, element["NAME"], blockCount)
		balance, err := client.GetBalance("*")
		if err != nil {
			log.Print(err)
			continue
		}
		fmt.Printf("\"%s\" \"wallet.balance[%s]\" \"%f\"\n", *hostnamePtr, element["NAME"], balance.ToBTC())
	}
}
