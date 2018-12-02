package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"github.com/Elbandi/btcd/rpcclient"
)

var (
	// Errors
	ErrAlgoNotFound    = errors.New("Algo not found")
	ErrNetHashNotFound = errors.New("Network hashps not found")

	// flags
	Hostname string
	Port int
	Username string
	Password string
	CertFile string
)

func newRpcClient() (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         Hostname + ":" + strconv.Itoa(Port),
		User:         Username,
		Pass:         Password,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	if len(CertFile) > 0 {
		content, err := ioutil.ReadFile(CertFile)
		if err != nil {
			return nil, err
		}
		connCfg.Certificates = content
		connCfg.DisableTLS = false
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	return rpcclient.New(connCfg, nil)
}

func GetInfo(request []string) (string, error) {
	client, err := newRpcClient()
	if err != nil {
		return "{}", err
	}
	defer client.Shutdown()

	res, err := client.GetInfoAsync().ReceiveFuture()
	if err != nil {
		return "{}", err
	}
	return string(res), nil
}

func GetBalance(request []string) (float64, error) {
	client, err := newRpcClient()
	if err != nil {
		return 0, err
	}
	defer client.Shutdown()

	res, err := client.GetBalance("")
	if err != nil {
		return 0, err
	}
	return res.ToBTC(), nil
}

func GetBlockCount(request []string) (uint64, error) {
	client, err := newRpcClient()
	if err != nil {
		return 0, err
	}
	defer client.Shutdown()

	res, err := client.GetBlockCount()
	if err != nil {
		return 0, err
	}
	return uint64(res), nil
}

func GetConnectionCount(request []string) (uint64, error) {
	client, err := newRpcClient()
	if err != nil {
		return 0, err
	}
	defer client.Shutdown()

	res, err := client.GetConnectionCount()
	if err != nil {
		return 0, err
	}
	return uint64(res), nil
}

func GetDifficulty(request []string) (float64, error) {
	client, err := newRpcClient()
	if err != nil {
		return 0, err
	}
	defer client.Shutdown()

	res, err := client.GetMiningInfoAsync().ReceiveFuture()
	if err != nil {
		return 0, err
	}
	var f interface{}
	err = json.Unmarshal(res, &f)
	if err != nil {
		return 0, err
	}

	if len(request) > 0 {
		for k, v := range f.(map[string]interface{}) {
			if k == "difficulty_" + request[0] {
				return v.(float64), nil
			}
		}
	} else {
		diff := f.(map[string]interface{})["difficulty"]
		switch v := diff.(type) {
		case float64:
			return v, nil
		case map[string]interface{}:
			return v["proof-of-work"].(float64), nil
		default:
			return 0, errors.New("unknown difficulty type")
		}
	}
	return 0, ErrAlgoNotFound
}

func GetNetworkHashPS(request []string) (uint64, error) {
	client, err := newRpcClient()
	if err != nil {
		return 0, err
	}
	defer client.Shutdown()

	res, err := client.GetMiningInfoAsync().ReceiveFuture()
	if err != nil {
		return 0, err
	}
	var f interface{}
	err = json.Unmarshal(res, &f)
	if err != nil {
		return 0, err
	}

	if len(request) > 0 {
		for k, v := range f.(map[string]interface{}) {
			if k == "networkhashps_" + request[0] {
				return uint64(v.(float64)), nil
			}
		}
	} else {
		if val, ok := f.(map[string]interface{})["netmhashps"]; ok {
			return uint64(val.(float64) * 1024 * 1024), nil
		}
		if val, ok := f.(map[string]interface{})["networkhashps"]; ok {
			return uint64(val.(float64)), nil
		}
		return 0, ErrNetHashNotFound
	}
	return 0, ErrAlgoNotFound
}

func GetLastRecipient(request []string) (string, error) {
	client, err := newRpcClient()
	if err != nil {
		return "", err
	}
	defer client.Shutdown()

	var blockCount int64
	if len(request) > 0 {
		blockCount, err = strconv.ParseInt(request[0], 10, 0)
		if err != nil {
			return "", err
		}
	} else {
		blockCount, err = client.GetBlockCount()
		if err != nil {
			return "", err
		}
	}
	for ; blockCount > 0; blockCount-- {
		blockHash, err := client.GetBlockHash(blockCount)
		if err != nil {
			return "", err
		}
		block, err := client.GetBlockVerbose(blockHash)
		if err != nil {
			return "", err
		}
		if block.Nonce == 0 {
			// pos block has 0 nonce value, skip
			continue
		}
		for _, tx := range block.RawTx {
			if tx.Vin[0].Coinbase == "" {
				continue
			}
			return tx.Vout[0].ScriptPubKey.Addresses[0], nil
		}
	}

	return "", errors.New("Coinbase not found")
}

func GetLastMinedHeight(request []string) (int64, error) {
	var err error
	txCount := 5
	if len(request[0]) == 0 {
		return 0, errors.New("Empty account name")
	}
	if len(request) > 1 && len(request[1]) > 0 {
		txCount, err = strconv.Atoi(request[1])
		if err != nil {
			return 0, errors.New("Invalid transactions count format")
		}
	}
	client, err := newRpcClient()
	if err != nil {
		return 0, err
	}
	defer client.Shutdown()

	txs, err := client.ListTransactionsCount(request[0], txCount)
	if err != nil {
		return 0, err
	}
	maxHeight := int64(math.MaxInt64)
	for _, tx := range txs {
		if tx.Category != "immature" && tx.Category != "generate" {
			continue
		}
		if maxHeight > tx.Confirmations {
			maxHeight = tx.Confirmations
		}
	}
	return maxHeight, nil
}

func main() {
	flag.StringVar(&Hostname, "hostname", "localhost", "Send commands to node running on host")
	flag.IntVar(&Port, "port", 1234, "Connect to JSON-RPC port")
	flag.StringVar(&Username, "username", "rpc", "Username for JSON-RPC connections")
	flag.StringVar(&Password, "password", "", "Password for JSON-RPC connections")
	flag.StringVar(&CertFile, "cert", "", "Load certificate from this file")
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.Arg(0) {
	case "info":
		if v, err := GetInfo(flag.Args()[1:]); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	case "balance":
		if v, err := GetBalance(flag.Args()[1:]); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	case "blocks":
		if v, err := GetBlockCount(flag.Args()[1:]); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	case "connections":
		if v, err := GetConnectionCount(flag.Args()[1:]); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	case "difficulty":
		if v, err := GetDifficulty(flag.Args()[1:]); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	case "networkhashps":
		if v, err := GetNetworkHashPS(flag.Args()[1:]); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	case "lastrecipient":
		if v, err := GetLastRecipient(flag.Args()[1:]); err != nil {
			log.Fatalf("Error: %s", err.Error())
		} else {
			fmt.Print(v)
		}
	case "lastminedheight":
		switch flag.NArg() {
		case 2, 3:
			if v, err := GetLastMinedHeight(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s lasttransactionheight ACCOUNT [COUNT]", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: " +
			"'info', 'balance', 'blocks', 'connections', " +
			"'difficulty', 'networkhashps', 'lastrecipient' or 'lastminedheight'.")
	}
}
