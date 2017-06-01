package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/Elbandi/go-ccminer-api"
	"github.com/Elbandi/zabbix-checker/common/lld"
	"github.com/stefantalpalaru/pool"
)

const (
	localAddr = "127.0.0.1"
	mCastPort = 4068
	mCastReport = 4067
	maxDatagramSize = 8192
)

var (
	mCastAddr = &net.UDPAddr{
		//	IP:   net.IPv4(127, 0, 0, 1),
		IP:   net.IPv4(224, 0, 0, 75),
		Port: mCastPort,
	}
	listenAddr = &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: mCastReport,
	}
)

// var omitNewline = flag.Bool("n", false, "don't print final newline")

func QueryDevice(request []string) (*ccminer.Device, error) {
	// parse first param as int64
	port, err := strconv.ParseInt(request[0], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid port format")
	}
	devid, err := strconv.ParseInt(request[1], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid deviceid format")
	}
	miner := ccminer.New(localAddr, port)
	devices, err := miner.Devs()
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to CGMiner: %s", err.Error())
	}
	if int64(len(devices)) <= devid {
		return nil, errors.New("Invalid device id")
	}
/*
	for _, dev := range *devices {
			fmt.Printf("Dev %d temp: %f\n", dev.ID, dev.Temperature)
	}
	res2B, _ := json.Marshal((*devices)[0])
	fmt.Println(string(res2B))
*/
	dev := devices[devid];
	return &dev, nil
}

func sendDiscoveryMsg(port int) {
	time.Sleep(100 * time.Millisecond)
	c, err := net.DialUDP("udp", nil, mCastAddr)
	defer c.Close()
	if err != nil {
		return
	}
	msg := fmt.Sprintf("ccminer-FTW-%d", port)
	c.Write([]byte(msg))
}

func isMyAddress(ip net.IP) bool {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.Equal(ip) {
				return true
			}
		}
	}
	return false
}

// DiscoverMiner is a DiscoveryItemHandlerFunc for key `ccminer.discovery` which returns JSON
// encoded discovery data for all running ccminer
func DiscoverMiner(request []string) (lld.DiscoveryData, error) {
	// init discovery data
	d := make(lld.DiscoveryData, 0)

	discoverypool := pool.New(4)
	discoverypool.Run()

	go sendDiscoveryMsg(mCastReport)
	l, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("Unable to listen on %s: %s", err.Error())
	}
	l.SetReadBuffer(maxDatagramSize)
	l.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		b := make([]byte, maxDatagramSize)
		n, addr, err := l.ReadFromUDP(b)
		if err != nil {
			break
		}
		if isMyAddress(addr.IP) {
			continue
		}
		msg := strings.Split(string(b[:n]), "-")
		if len(msg) < 3 {
			continue
		}
		port, err := strconv.ParseInt(msg[2], 10, 64)
		if err == nil {
			discoverypool.Add(DiscoverDevs, port)
		}
	}

	//  status := mypool.Status()
	//  log.Println(status.Submitted, "submitted jobs,", status.Running, "running,", status.Completed, "completed.")
	discoverypool.Wait()
	completed_jobs := discoverypool.Results()
	for _, job := range completed_jobs {
		if job.Result == nil {
			// TODO: handle this
			log.Println("got error:", job.Err)
		} else {
			item := job.Result.(lld.DiscoveryData)
			if item != nil {
				d = append(d, item...)
			}
		}
	}

	return d, nil
}

func DiscoverDevs(args ...interface{}) interface{} {
	port := args[0].(int64)

	// init discovery data
	d := make(lld.DiscoveryData, 0)

	miner := ccminer.New(localAddr, port)
	devices, err := miner.Devs()
	if err != nil {
		panic(err)
	}

	for _, dev := range devices {
		item := make(lld.DiscoveryItem, 0)
		item["PORT"] = strconv.FormatInt(port, 10)
		item["DEVID"] = strconv.FormatUint(uint64(dev.Id), 10)
		item["NAME"] = dev.Card
		//  item["NAME"] = fmt.Sprintf("%s %d", dev.Card, dev.Id)
		d = append(d, item)
	}

	return d
}

// AcceptedShares is a Uint64ItemHandlerFunc for key `ccminer.accept_shares` which returns the accepted shares
// counter.
func AcceptedShares(request []string) (uint64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	return dev.Accepted, nil
}

// HardwareErrors is a Uint64ItemHandlerFunc for key `ccminer.hwerrors` which returns the hardware errors
// counter.
func HardwareErrors(request []string) (uint64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0, err
	}
	return uint64(dev.HardwareErrors), nil
}

// Frequency is a Uint64ItemHandlerFunc for key `ccminer.frequency` which returns the device frequency.
func Frequency(request []string) (uint64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	return uint64(dev.GpuFreq), nil
}

// Rate is a DoubleItemHandlerFunc for key `ccminer.hashrate` which returns
// the device average hashrate.
func Rate(request []string) (float64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	return float64(dev.Khs * 1000), nil
}


// RejectedShares is a Uint64ItemHandlerFunc for key `ccminer.rejected_shares` which returns the rejected shares
// counter.
func RejectedShares(request []string) (uint64, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	return dev.Rejected, nil
}

// Temperature is a DoubleItemHandlerFunc for key `ccminer.temperature` which returns the rejected shares
// counter.
func Temperature(request []string) (float32, error) {
	dev, err := QueryDevice(request)
	if err != nil {
		return 0.00, err
	}
	return dev.Temp, nil
}

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	switch flag.Arg(0) {
	case "discovery":
		switch flag.NArg() {
		case 1:
			if v, err := DiscoverMiner(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v.Json())
			}
		default:
			log.Fatalf("Usage: %s discovery", os.Args[0])
		}
	case "accepted_shares":
		switch flag.NArg() {
		case 3:
			if v, err := AcceptedShares(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s accept_shares PORT DEVICEID", os.Args[0])
		}
	case "frequency":
		switch flag.NArg() {
		case 3:
			if v, err := Frequency(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s frequency PORT DEVICEID", os.Args[0])
		}
	case "hwerrors":
		switch flag.NArg() {
		case 3:
			if v, err := HardwareErrors(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s hwerrors PORT DEVICEID", os.Args[0])
		}
	case "hashrate":
		switch flag.NArg() {
		case 3:
			if v, err := Rate(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s hashrate PORT DEVICEID", os.Args[0])
		}
	case "rejected_shares":
		switch flag.NArg() {
		case 3:
			if v, err := RejectedShares(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s rejected PORT DEVICEID", os.Args[0])
		}
	case "temperature":
		switch flag.NArg() {
		case 3:
			if v, err := Temperature(flag.Args()[1:]); err != nil {
				log.Fatalf("Error: %s", err.Error())
			} else {
				fmt.Print(v)
			}
		default:
			log.Fatalf("Usage: %s temperature PORT DEVICEID", os.Args[0])
		}
	default:
		log.Fatal("You must specify one of the following action: 'discovery', 'accepted_shares', 'frequency', 'hwerrors', 'hashrate', 'rejected_shares' or 'temperature'.")
	}
}
