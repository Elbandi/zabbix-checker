package main

import (
	"encoding/csv"
	"crypto/tls"
	"strings"
	"io/ioutil"
	"regexp"
	"bytes"
	"net/http"
	"log"
	"net/http/httputil"
	"github.com/dghubble/sling"
	"fmt"
	"flag"
	"os"
)

const defaultUserAgent = "digikey-checker/1.0"

var (
	// flags
	debug bool
	userAgent string
)

type DigikeyRequest int

func (d DigikeyRequest) Do(req *http.Request) (*http.Response, error) {
	if debug {
		d.dumpRequest(req)
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	client := http.DefaultClient
	if client.Transport != nil {
		if transport, ok := client.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig != nil {
				transport.TLSClientConfig.InsecureSkipVerify = true;
			} else {
				transport.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
			}
		}
	} else {
		if transport, ok := http.DefaultTransport.(*http.Transport); ok {
			if transport.TLSClientConfig != nil {
				transport.TLSClientConfig.InsecureSkipVerify = true;
			} else {
				transport.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
			}
		}
	}
	resp, err := client.Do(req)
	if debug {
		d.dumpResponse(resp)
	}
	if err == nil {
		if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
			resp.Header.Set("Content-Type", "application/json")
		}
	}
	body, err := ioutil.ReadAll(resp.Body);
	reg, _ := regexp.Compile(": *,")
	body = reg.ReplaceAll(body, []byte(":\"\","))
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return resp, err
}

func (d DigikeyRequest) dumpRequest(r *http.Request) {
	if r == nil {
		log.Print("dumpReq ok: <nil>")
		return
	}
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Print("dumpReq err:", err)
	} else {
		log.Print("dumpReq ok:", string(dump))
	}
}

func (d DigikeyRequest) dumpResponse(r *http.Response) {
	if r == nil {
		log.Print("dumpResponse ok: <nil>")
		return
	}
	dump, err := httputil.DumpResponse(r, true)
	if err != nil {
		log.Print("dumpResponse err:", err)
	} else {
		log.Print("dumpResponse ok:", string(dump))
	}
}

func main() {
	flag.BoolVar(&debug, "debug", false, "enable request/response dump")
	flag.StringVar(&userAgent, "user-agent", defaultUserAgent, "http client user agent")
	flag.Parse()
	log.SetOutput(os.Stderr)

	if flag.NArg() != 1 {
		log.Fatal("Need an url")
	}
	var price DigikeyRequest
	response, err := sling.New().Doer(price).Base("https://www.digikey.com/product-search/download.csv").Get("?" + flag.Arg(0)).Receive(nil, nil)
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
	defer response.Body.Close()
	reader := csv.NewReader(response.Body)
	if _, err := reader.Read(); err != nil {
		//read header
		log.Fatalf("Error: %s", err.Error())
	}
	fields, err := reader.Read()
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
	fmt.Print(fields[8])
}
