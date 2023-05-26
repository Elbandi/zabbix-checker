package main

import (
	"flag"
	"fmt"
	"github.com/Elbandi/zabbix-checker/common/lld"
	"github.com/btcsuite/btcd/rpcclient"
	"gopkg.in/ini.v1"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	var hostnameFlag string
	const (
		hostnameDefault     = ""
		hostnameDescription = "zabbix hostname"
	)
	flag.StringVar(&hostnameFlag, "hostname", hostnameDefault, hostnameDescription)
	flag.StringVar(&hostnameFlag, "h", hostnameDefault, hostnameDescription)

	var basepathFlag string
	const (
		basepathDefault     = "/srv"
		basepathDescription = "base path"
	)
	flag.StringVar(&basepathFlag, "basepath", basepathDefault, basepathDescription)
	flag.StringVar(&basepathFlag, "b", basepathDefault, basepathDescription)

	var searchSuffixFlag string
	const (
		searchSuffixDefault     = "-data"
		searchSuffixDescription = "search suffix"
	)
	flag.StringVar(&searchSuffixFlag, "searchSuffix", searchSuffixDefault, searchSuffixDescription)
	flag.StringVar(&searchSuffixFlag, "s", searchSuffixDefault, searchSuffixDescription)

	var excludeSearchFlag arrayFlags
	const (
		excludeSearchDescription = "exclude from search list"
	)
	flag.Var(&excludeSearchFlag, "exclude", excludeSearchDescription)
	flag.Var(&excludeSearchFlag, "e", excludeSearchDescription)

	flag.Parse()
	log.SetOutput(os.Stderr)

	if len(hostnameFlag) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	discovery := make(lld.DiscoveryData, 0)
	err := filepath.Walk(basepathFlag, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if _, err := os.Stat(path + "/.nosync"); !os.IsNotExist(err) {
			return nil
		}
		if strings.HasSuffix(path, searchSuffixFlag) {
			name := strings.TrimSuffix(filepath.Base(path), searchSuffixFlag)
			if stringInSlice(name, excludeSearchFlag) {
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
	fmt.Printf("\"%s\" \"wallet.discovery\" %s\n", hostnameFlag, strconv.Quote(discovery.JsonLine()))
	for _, element := range discovery {
		logPath := filepath.Join(element["PATH"], "debug.log")
		fi, err := os.Stat(logPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("\"%s\" \"vfs.file.size[%s]\" \"0\"\n", hostnameFlag, logPath)
			} else {
				log.Print(err)
			}
		} else {
			fmt.Printf("\"%s\" \"vfs.file.size[%s]\" \"%d\"\n", hostnameFlag, logPath, fi.Size())
		}

		config := &BitcoinConfig{Hostname: "127.0.0.1", Port: 8332}
		err = ini.MapTo(config, filepath.Join(element["PATH"], element["NAME"]+".conf"))
		if err != nil {
			log.Print(err)
			continue
		}
		connCfg := &rpcclient.ConnConfig{
			Host:         config.Hostname + ":" + strconv.Itoa(config.Port),
			User:         config.Username,
			Pass:         config.Password,
			HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   true, // Bitcoin core does not provide TLS by default
		}
		// Notice the notification parameter is nil since notifications are
		// not supported in HTTP POST mode.
		client, err := rpcclient.New(connCfg, nil)
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
		fmt.Printf("\"%s\" \"wallet.blocks[%s]\" \"%d\"\n", hostnameFlag, element["NAME"], blockCount)
		blockhash, err := client.GetBlockHash(blockCount)
		if err != nil {
			log.Print(err)
			continue
		}
		block, err := client.GetBlockVerbose(blockhash)
		if err != nil {
			log.Print(err)
			continue
		}
		fmt.Printf("\"%s\" \"wallet.blocktime[%s]\" \"%d\"\n", hostnameFlag, element["NAME"], block.Time)
		balance, err := client.GetBalance()
		if err != nil {
			log.Print(err)
			continue
		}
		fmt.Printf("\"%s\" \"wallet.balance[%s]\" \"%f\"\n", hostnameFlag, element["NAME"], balance.ToBTC())
	}
}
