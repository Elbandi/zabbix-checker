package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
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
)

func GetDifficulty(request []string) (float64, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         Hostname + ":" + strconv.Itoa(Port),
		User:         Username,
		Pass:         Password,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
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
	connCfg := &rpcclient.ConnConfig{
		Host:         Hostname + ":" + strconv.Itoa(Port),
		User:         Username,
		Pass:         Password,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
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
	connCfg := &rpcclient.ConnConfig{
		Host:         Hostname + ":" + strconv.Itoa(Port),
		User:         Username,
		Pass:         Password,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
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

func main() {
	flag.StringVar(&Hostname, "hostname", "localhost", "Send commands to node running on host")
	flag.IntVar(&Port, "port", 1234, "Connect to JSON-RPC port")
	flag.StringVar(&Username, "username", "rpc", "Username for JSON-RPC connections")
	flag.StringVar(&Password, "password", "", "Password for JSON-RPC connections")
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.Arg(0) {
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
	default:
		log.Fatal("You must specify one of the following action: " +
			"'discovery', " +
			"'difficulty' or 'networkhashps'.")
	}
}
