package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"github.com/bitbandi/go-mpos-api"
	"github.com/bitbandi/go-nomp-api"
	"github.com/bitbandi/go-yiimp-api"
	"github.com/Elbandi/zabbix-checker/common/lld"
	"golang.org/x/net/http2"
	"io"
	"os/exec"
)

var (
	// Errors
	ErrAlgoNotFound = errors.New("no such algorithm")
	ErrPoolNotFound = errors.New("pool not found")

	// flags
	debug          bool
	userAgent      string
	zabbixHostName string
	zabbixServer   string
)

type myTransport struct {
	proxyURL *url.URL
	rt       *http.Transport
}

func (t *myTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if len(userAgent) > 0 {
		r.Header.Set("User-Agent", userAgent)
	}
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
		if err == nil {
			response.Request.URL.Host = tempHost
		}
	}
	return response, err
}

// DiscoverPools is a DiscoveryItemHandlerFunc for key `pool.discovery` which returns JSON
// encoded discovery data for pool stored in a file
func DiscoverPools(request []string) (error) {
	// init discovery data
	yiimpDiscovery := make(lld.DiscoveryData, 0)
	mposDiscovery := make(lld.DiscoveryData, 0)
	nompDiscovery := make(lld.DiscoveryData, 0)
	file, err := os.Open(request[0])
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line[0] == '#' {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) < 2 {
			continue
		}

		switch strings.TrimSpace(fields[1]) {
		case "YIIMP":
			if len(fields) > 5 {
				item := make(lld.DiscoveryItem, 0)
				item["NAME"] = strings.TrimSpace(fields[0])
				item["TYPE"] = strings.TrimSpace(fields[1])
				item["HOST"] = strings.TrimSpace(fields[2])
				item["ALGO"] = strings.TrimSpace(fields[3])
				item["ADDRESS"] = strings.TrimSpace(fields[4])
				item["PROXY"] = strings.TrimSpace(fields[5])
				if len(fields) > 6 {
					item["LOW_POOL_LIMIT"] = strings.TrimSpace(fields[6])
				}
				if len(fields) > 7 {
					item["HIGH_POOL_LIMIT"] = strings.TrimSpace(fields[7])
				}
				yiimpDiscovery = append(yiimpDiscovery, item)
			}
		case "MPOS":
			if len(fields) > 4 {
				item := make(lld.DiscoveryItem, 0)
				item["NAME"] = strings.TrimSpace(fields[0])
				item["TYPE"] = strings.TrimSpace(fields[1])
				item["HOST"] = strings.TrimSpace(fields[2])
				item["APIKEY"] = strings.TrimSpace(fields[3])
				item["PROXY"] = strings.TrimSpace(fields[4])
				if len(fields) > 5 {
					item["LOW_POOL_LIMIT"] = strings.TrimSpace(fields[5])
				}
				if len(fields) > 6 {
					item["HIGH_POOL_LIMIT"] = strings.TrimSpace(fields[6])
				}
				mposDiscovery = append(mposDiscovery, item)

			}
		case "NOMP":
			if len(fields) > 5 {
				item := make(lld.DiscoveryItem, 0)
				item["NAME"] = strings.TrimSpace(fields[0])
				item["TYPE"] = strings.TrimSpace(fields[1])
				item["HOST"] = strings.TrimSpace(fields[2])
				item["POOL"] = strings.TrimSpace(fields[3])
				item["WORKER"] = strings.TrimSpace(fields[4])
				item["PROXY"] = strings.TrimSpace(fields[5])
				if len(fields) > 6 {
					item["LOW_POOL_LIMIT"] = strings.TrimSpace(fields[6])
				}
				if len(fields) > 7 {
					item["HIGH_POOL_LIMIT"] = strings.TrimSpace(fields[7])
				}
				nompDiscovery = append(nompDiscovery, item)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	fmt.Printf("\"%s\" \"yiimp.discovery\" %s\n", zabbixHostName, strconv.Quote(yiimpDiscovery.JsonLine()))
	fmt.Printf("\"%s\" \"mpos.discovery\" %s\n", zabbixHostName, strconv.Quote(mposDiscovery.JsonLine()))
	fmt.Printf("\"%s\" \"nomp.discovery\" %s\n", zabbixHostName, strconv.Quote(nompDiscovery.JsonLine()))
	return nil
}

type CmdAction func(args []string)

func RunQueryAndSend(args []string) {
	log.Println(strings.Join(args, " "))
	cmdSender := exec.Command("zabbix_sender", "-vv", "-z", zabbixServer, "-s", zabbixHostName, "-i", "-")
	//cmdSender := exec.Command("cat")
	cmdSender.SysProcAttr = GetSysProcAttr()

	cmdQuery := exec.Command(os.Args[0], "slave")
	cmdQuery.SysProcAttr = GetSysProcAttr()
	//cmdQuery.Stdout = os.Stdout
	cmdSender.Stdin, _ = cmdQuery.StdoutPipe()
	cmdSender.Stdout = os.Stdout
	cmdSender.Stderr = os.Stderr
	cmdQuery.Stderr = os.Stderr
	stdin, err := cmdQuery.StdinPipe()
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := cmdSender.Start(); err != nil {
		fmt.Println(err)
		return
	}
	if err := cmdQuery.Start(); err != nil {
		cmdSender.Process.Kill()
		fmt.Println(err)
		return
	}
	io.WriteString(stdin, strings.Join(args, "\n"))
	stdin.Close()
}

func PrintCommand(args []string) {
	fmt.Println(strings.Join(args, " "))
}

func RunQuery(request []string, action CmdAction) (error) {
	file, err := os.Open(request[0])
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line[0] == '#' {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) < 2 {
			continue
		}

		switch strings.TrimSpace(fields[1]) {
		case "YIIMP":
			if len(fields) > 5 {
				args := []string{"-hostname", zabbixHostName}
				if proxy := strings.TrimSpace(fields[5]); len(proxy) > 0 {
					args = append(args, "-proxy", proxy)
				}
				if debug {
					args = append(args, "-debug")
				}
				if len(userAgent) > 0 {
					args = append(args, "-user-agent", userAgent)
				}
				args = append(args, "yiimp", strings.TrimSpace(fields[2]), strings.TrimSpace(fields[3]), strings.TrimSpace(fields[4]))
				action(args)
			}
		case "MPOS":
			if len(fields) > 4 {
				args := []string{"-hostname", zabbixHostName}
				if proxy := strings.TrimSpace(fields[4]); len(proxy) > 0 {
					args = append(args, "-proxy", proxy)
				}
				if debug {
					args = append(args, "-debug")
				}
				if len(userAgent) > 0 {
					args = append(args, "-user-agent", userAgent)
				}
				args = append(args, "mpos", strings.TrimSpace(fields[2]), strings.TrimSpace(fields[3]))
				action(args)
			}
		case "NOMP":
			if len(fields) > 5 {
				args := []string{"-hostname", zabbixHostName}
				if proxy := strings.TrimSpace(fields[5]); len(proxy) > 0 {
					args = append(args, "-proxy", proxy)
				}
				if debug {
					args = append(args, "-debug")
				}
				if len(userAgent) > 0 {
					args = append(args, "-user-agent", userAgent)
				}
				args = append(args, "nomp", strings.TrimSpace(fields[2]), strings.TrimSpace(fields[3]), strings.TrimSpace(fields[4]))
				action(args)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func YiimpStatus(request []string) (error) {
	yiimpClient := yiimp.NewYiimpClient(nil, request[0], "", userAgent)
	yiimpClient.SetDebug(debug)
	status, err := yiimpClient.GetStatus()
	if err != nil {
		return err
	}
	algo, ok := status[request[1]]
	if !ok {
		return ErrAlgoNotFound
	}
	fmt.Printf("\"%s\" \"yiimp.pool_hashrate[%s,%s]\" \"%.0f\"\n", zabbixHostName, request[0], request[1], algo.Hashrate)
	fmt.Printf("\"%s\" \"yiimp.pool_workers[%s,%s]\" \"%.0d\"\n", zabbixHostName, request[0], request[1], algo.Workers)
	fmt.Printf("\"%s\" \"yiimp.pool_estimate_current[%s,%s]\" \"%.8f\"\n", zabbixHostName, request[0], request[1], algo.EstimateCurrent)
	fmt.Printf("\"%s\" \"yiimp.pool_estimate_last24h[%s,%s]\" \"%.8f\"\n", zabbixHostName, request[0], request[1], algo.EstimateLast24h)
	fmt.Printf("\"%s\" \"yiimp.pool_actual_last24h[%s,%s]\" \"%.8f\"\n", zabbixHostName, request[0], request[1], algo.ActualLast24h)
	fmt.Printf("\"%s\" \"yiimp.pool_rental[%s,%s]\" \"%.8f\"\n", zabbixHostName, request[0], request[1], algo.RentalCurrent)

	walletstatus, err := yiimpClient.GetWalletEx(request[2])
	if err != nil && err != io.EOF {
		return err
	}
	if err == nil {
		var userHashrate uint64 = 0
		for _, miner := range walletstatus.Miners {
			if miner.Algo == request[1] {
				userHashrate += uint64(miner.Accepted)
			}
		}
		fmt.Printf("\"%s\" \"yiimp.user_hashrate[%s,%s,%s]\" \"%.0d\"\n", zabbixHostName, request[0], request[1], request[2], userHashrate)
	}
	return nil
}

func splitApiKey(key string) (string, uint64, error) {
	if !strings.Contains(key, "_") {
		return key, 0, nil
	}
	keyArray := strings.SplitN(key, "_", 2)
	userId, err := strconv.ParseUint(keyArray[1], 10, 64)
	return keyArray[0], userId, err
}

func MposStatus(request []string) (error) {
	apikey, userid, err := splitApiKey(request[1])
	if err != nil {
		return err
	}
	mposClient := mpos.NewMposClient(nil, request[0], apikey, userid, userAgent)
	mposClient.SetDebug(debug)
	poolStatus, err := mposClient.GetPoolStatus()
	if err != nil {
		return err
	}
	userStatus, err := mposClient.GetUserStatus()
	if err != nil {
		return err
	}
	balance, err := mposClient.GetUserBalance()
	if err != nil {
		return err
	}

	fmt.Printf("\"%s\" \"mpos.pool_hashrate[%s,%s]\" \"%.0f\"\n", zabbixHostName, request[0], request[1], poolStatus.Hashrate*1000)
	fmt.Printf("\"%s\" \"mpos.pool_workers[%s,%s]\" \"%d\"\n", zabbixHostName, request[0], request[1], poolStatus.Workers)
	fmt.Printf("\"%s\" \"mpos.pool_efficiency[%s,%s]\" \"%f\"\n", zabbixHostName, request[0], request[1], poolStatus.Efficiency)
	fmt.Printf("\"%s\" \"mpos.pool_lastblock[%s,%s]\" \"%d\"\n", zabbixHostName, request[0], request[1], poolStatus.LastBlock)
	fmt.Printf("\"%s\" \"mpos.pool_nextblock[%s,%s]\" \"%d\"\n", zabbixHostName, request[0], request[1], poolStatus.NextNetworkBlock)
	fmt.Printf("\"%s\" \"mpos.user_hashrate[%s,%s]\" \"%.0f\"\n", zabbixHostName, request[0], request[1], userStatus.Hashrate*1000)
	fmt.Printf("\"%s\" \"mpos.user_sharerate[%s,%s]\" \"%f\"\n", zabbixHostName, request[0], request[1], userStatus.Sharerate)
	fmt.Printf("\"%s\" \"mpos.user_shares_valid[%s,%s]\" \"%f\"\n", zabbixHostName, request[0], request[1], userStatus.Shares.Valid)
	fmt.Printf("\"%s\" \"mpos.user_shares_invalid[%s,%s]\" \"%f\"\n", zabbixHostName, request[0], request[1], userStatus.Shares.Invalid)
	fmt.Printf("\"%s\" \"mpos.user_balance_confirmed[%s,%s]\" \"%f\"\n", zabbixHostName, request[0], request[1], balance.Confirmed)
	fmt.Printf("\"%s\" \"mpos.user_balance_unconfirmed[%s,%s]\" \"%f\"\n", zabbixHostName, request[0], request[1], balance.Unconfirmed)
	return nil
}

func NompStatus(request []string) (error) {
	nompClient := nomp.NewNompClient(nil, request[0], userAgent)
	nompClient.SetDebug(debug)
	status, err := nompClient.GetPoolStatus()
	if err != nil {
		return err
	}
	pool, ok := status.Pools[request[1]]
	if !ok {
		return ErrPoolNotFound
	}
	var hashrate float64 = 0
	var shares float64 = 0
	var invalidshares float64 = 0
	for idx, worker := range pool.Workers {
		if strings.HasPrefix(idx, request[2]) {
			hashrate += worker.Hashrate
			shares += worker.Shares
			invalidshares += worker.InvalidShares
		}
	}
	fmt.Printf("\"%s\" \"nomp.pool_hashrate[%s,%s]\" \"%.0f\"\n", zabbixHostName, request[0], request[1], pool.Hashrate)
	fmt.Printf("\"%s\" \"nomp.pool_workers[%s,%s]\" \"%d\"\n", zabbixHostName, request[0], request[1], pool.WorkerCount)
	fmt.Printf("\"%s\" \"nomp.pool_shares_valid[%s,%s]\" \"%d\"\n", zabbixHostName, request[0], request[1], pool.Stat.ValidShares)
	fmt.Printf("\"%s\" \"nomp.pool_shares_invalid[%s,%s]\" \"%d\"\n", zabbixHostName, request[0], request[1], pool.Stat.InvalidShares)
	fmt.Printf("\"%s\" \"nomp.pool_blocks_pending[%s,%s]\" \"%d\"\n", zabbixHostName, request[0], request[1], pool.Blocks.Pending)
	fmt.Printf("\"%s\" \"nomp.pool_blocks_confirmed[%s,%s]\" \"%d\"\n", zabbixHostName, request[0], request[1], pool.Blocks.Confirmed)
	fmt.Printf("\"%s\" \"nomp.user_hashrate[%s,%s,%s]\" \"%.0f\"\n", zabbixHostName, request[0], request[1], request[2], hashrate)
	fmt.Printf("\"%s\" \"nomp.user_shares_valid[%s,%s,%s]\" \"%f\"\n", zabbixHostName, request[0], request[1], request[2], shares)
	fmt.Printf("\"%s\" \"nomp.user_shares_invalid[%s,%s,%s]\" \"%f\"\n", zabbixHostName, request[0], request[1], request[2], invalidshares)
	return nil
}

func main() {
	proxyPtr := flag.String("proxy", "", "socks proxy")
	flag.BoolVar(&debug, "debug", false, "enable request/response dump")
	flag.StringVar(&userAgent, "user-agent", "", "http client user agent")
	flag.StringVar(&zabbixHostName, "hostname", "", "zabbix hostname")
	flag.StringVar(&zabbixServer, "zabbix-server", "", "zabbix server")
	log.SetOutput(os.Stderr)
	if len(os.Args) > 1 && os.Args[1] == "slave" {
		var lines []string
		reader := bufio.NewScanner(os.Stdin)
		//file, _ := os.Open("aa")
		//reader := bufio.NewScanner(file)
		for reader.Scan() {
			lines = append(lines, reader.Text())
		}
		flag.CommandLine.Parse(lines)
	} else {
		flag.Parse()
	}

	if zabbixHostName == "" {
		flag.Usage()
		os.Exit(1)
	}

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

	switch flag.Arg(0) {
	case "discovery":
		switch flag.NArg() {
		case 2:
			if err := DiscoverPools(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			}
		default:
			log.Fatalf("Usage: %s discovery PATH", os.Args[0])
		}
	case "getcmd":
		switch flag.NArg() {
		case 2:
			if err := RunQuery(flag.Args()[1:], PrintCommand); err != nil {
				log.Fatalf("Error: %s", err.Error())
			}
		default:
			log.Fatalf("Usage: %s getcmd PATH", os.Args[0])
		}
	case "runquery":
		if zabbixServer == "" {
			flag.Usage()
			os.Exit(1)
		}
		switch flag.NArg() {
		case 2:
			if err := RunQuery(flag.Args()[1:], RunQueryAndSend); err != nil {
				log.Fatalf("Error: %s", err.Error())
			}
		default:
			log.Fatalf("Usage: %s runquery PATH", os.Args[0])
		}
	case "yiimp":
		switch flag.NArg() {
		case 4:
			if err := YiimpStatus(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			}
		default:
			log.Fatalf("Usage: %s yiimp URL ALGORITHM ADDRESS", os.Args[0])
		}
	case "mpos":
		switch flag.NArg() {
		case 3:
			if err := MposStatus(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			}
		default:
			log.Fatalf("Usage: %s mpos URL APIKEY", os.Args[0])
		}
	case "nomp":
		switch flag.NArg() {
		case 4:
			if err := NompStatus(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			}
		default:
			log.Fatalf("Usage: %s nomp URL POOL WORKER", os.Args[0])
		}

	default:
		log.Fatal("You must specify one of the following action: " +
			"'discovery', 'getcmd', 'runquery', 'yiimp', 'mpos' or 'nomp'.")

	}
}
