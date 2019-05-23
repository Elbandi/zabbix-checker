package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"golang.org/x/net/http2"
)

const defaultUserAgent = "coinzark-checker/1.0"

var (
	// flags
	debug     bool
	userAgent string
)

type myTransport struct {
	rt *http.Transport
}

func (d myTransport) dumpRequest(r *http.Request) {
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

func (d myTransport) dumpResponse(r *http.Response) {
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

func (t *myTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", userAgent)
	r.Header.Set("Cache-Control", "max-age=0")
	r.Header.Set("Accept-Language", "en-us")
	if t.rt.TLSClientConfig != nil {
		t.rt.TLSClientConfig.InsecureSkipVerify = true
		t.rt.TLSClientConfig.ServerName = r.Host
	}
	if debug {
		t.dumpRequest(r)
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
	if debug {
		t.dumpResponse(response)
	}
	return response, err
}

func init() {
	flag.BoolVar(&debug, "debug", false, "Print debug infos")
	flag.StringVar(&userAgent, "user-agent", defaultUserAgent, "http client user agent")
}

type Rate struct {
	DepositAmount             float64 `json:"depositAmount,string"`
	FinalAmount               float64 `json:"finalAmount,string"`
	MinimumDeposit            float64 `json:"minimumDeposit,string"`
	MaximumDeposit            float64 `json:"maximumDeposit,string"`
	ReceiveNetworkFee         float64 `json:"receive_network_fee,string"`
	ReceiveNetworkFeeIncluded bool    `json:"receive_network_fee_included"`
}

func (r *Rate) UnmarshalJSON(data []byte) error {
	var err error
	type Alias Rate
	aux := &struct {
		FinalAmount interface{} `json:"finalAmount"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	switch v := aux.FinalAmount.(type) {
	case int64:
		r.FinalAmount = float64(v)
	case float64:
		r.FinalAmount = v
	case string:
		r.FinalAmount, err = strconv.ParseFloat(v, 64)
	}
	return err
}

type RateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  Rate   `json:"result"`
}

func main() {
	proxyPtr := flag.String("proxy", "", "socks proxy")
	flag.Parse()
	log.SetOutput(os.Stderr)

	if flag.NArg() < 2 {
		log.Fatalf("Usage: %s [-proxy ip:port] FROM TO", os.Args[0])
	}
	client := http.Client{}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		log.Fatalf("Failed to get the default http transport")
	}
	if *proxyPtr != "" {
		proxyURL, err := url.Parse("socks5://" + *proxyPtr)
		if err != nil {
			log.Fatalf("Failed to parse proxy URL: %v", err)
		}
		log.Printf("Set proxy to %s", proxyURL)
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			return proxyURL, nil
		}
	}
	http2.ConfigureTransport(transport)
	http.DefaultTransport = &myTransport{
		rt: transport,
	}
	client.Transport = http.DefaultTransport

	req, err := http.NewRequest("GET", "https://www.coinzark.com/api/v2/swap/rate?amount=1", nil)
	q := req.URL.Query()
	q.Add("from", flag.Args()[0])
	q.Add("to", flag.Args()[1])
	req.URL.RawQuery = q.Encode()
	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get rate: %v", err)
	}
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	var rate RateResponse
	err = decoder.Decode(&rate)
	if err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}
	if !rate.Success {
		log.Fatalf("Bad response: %s", rate.Message)
	}
	fmt.Printf("%.8f\n", rate.Result.FinalAmount)
}
