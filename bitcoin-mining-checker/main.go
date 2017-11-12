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
	ErrAlgoNotFound = errors.New("Algo not found")

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
		return uint64(f.(map[string]interface{})["networkhashps"].(float64)), nil
	}
	return 0, ErrAlgoNotFound
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
	default:
		log.Fatal("You must specify one of the following action: " +
			"'discovery', " +
			"'difficulty' or 'networkhashps'.")
	}
}
