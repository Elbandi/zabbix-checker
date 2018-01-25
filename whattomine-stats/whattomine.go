package main

import (
	"github.com/dghubble/sling"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"strconv"
	"time"
)

const (
	BASE string = "https://whattomine.com/"
)

type WhatToMineClient struct {
	sling      *sling.Sling
	httpClient *wtmHttpClient
}

// wtm send the api response with text/html content type
// we fix this: change content type to json
type wtmHttpClient struct {
	client    *http.Client
	debug     bool
	useragent string
}

func (d wtmHttpClient) Do(req *http.Request) (*http.Response, error) {
	if d.debug {
		d.dumpRequest(req)
	}
	if d.useragent != "" {
		req.Header.Set("User-Agent", d.useragent)
	}
	client := func() (*http.Client) {
		if d.client != nil {
			return d.client
		} else {
			return http.DefaultClient
		}
	}()
	if client.Transport != nil {
		if transport, ok := client.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig != nil {
				transport.TLSClientConfig.InsecureSkipVerify = true
			} else {
				transport.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
			}
		}
	} else {
		if transport, ok := http.DefaultTransport.(*http.Transport); ok {
			if transport.TLSClientConfig != nil {
				transport.TLSClientConfig.InsecureSkipVerify = true
			} else {
				transport.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
			}
		}
	}
	resp, err := client.Do(req)
	if d.debug {
		d.dumpResponse(resp)
	}
	if err == nil {
		if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
			resp.Header.Set("Content-Type", "application/json")
		}
	}
	return resp, err
}

func (d wtmHttpClient) dumpRequest(r *http.Request) {
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

func (d wtmHttpClient) dumpResponse(r *http.Response) {
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

func NewWhatToMineClient(client *http.Client, BaseURL string, UserAgent string) *WhatToMineClient {
	httpClient := &wtmHttpClient{client: client, useragent: UserAgent}
	return &WhatToMineClient{
		httpClient: httpClient,
		sling:      sling.New().Doer(httpClient).Base(BaseURL),
	}
}

func (client WhatToMineClient) SetDebug(debug bool) {
	client.httpClient.debug = debug
}

// Struct type that represents one coin from www.whattomine.com
type Coin struct {
	Id                 uint64    `json:"id"`
	Tag                string    `json:"tag"`
	Algorithm          string    `json:"algorithm"`
	BlockTime          float64   `json:"block_time"`
	BlockReward        float64   `json:"block_reward"`
	BlockReward24      float64   `json:"block_reward24"`
	LastBlock          uint64    `json:"last_block"`
	Difficulty         float64   `json:"difficulty"`
	Difficulty24       float64   `json:"difficulty24"`
	NetHash            float64   `json:"nethash"`
	ExchangeRate       float64   `json:"exchange_rate"`
	ExchangeRate24     float64   `json:"exchange_rate24"`
	ExchangeRageVol    float64   `json:"exchange_rage_vol"`
	ExchangeRageCurr   string    `json:"exchange_rage_curr"`
	MarketCap          string    `json:"market_cap"`
	EstimatedRewards   float64   `json:"estimated_rewards,string"`
	EstimatedRewards24 float64   `json:"estimated_rewards24,string"`
	BtcRevenue         float64   `json:"btc_revenue,string"`
	BtcRevenue24       float64   `json:"btc_revenue24,string"`
	Profitability      uint64    `json:"profitability"`
	Profitability24    uint64    `json:"profitability24"`
	Lagging            bool      `json:"lagging"`
	Timestamp          time.Time `json:"timestamp"`
}

func (t *Coin) UnmarshalJSON(data []byte) error {
	var err error
	type Alias Coin
	aux := &struct {
		BlockTime        json.Number `json:"block_time"`
		EstimatedRewards string      `json:"estimated_rewards"`
		Timestamp        int64       `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err = json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.Timestamp = time.Unix(aux.Timestamp, 0)
	t.EstimatedRewards, err = strconv.ParseFloat(strings.Replace(aux.EstimatedRewards, ",", "", -1), 64)
	if err != nil {
		return err
	}
	t.BlockTime, err = aux.BlockTime.Float64()
	return err
}

type coinsRequest struct {
	HashRate     float64 `url:"hr,omitempty"`
	Power        float64 `url:"p,omitempty"`
	PoolFee      float64 `url:"fee,omitempty"`
	PowerCost    float64 `url:"cost,omitempty"`
	HardwareCost float64 `url:"hcost,omitempty"`
}
type Coins map[string]Coin

type coinsResponse struct {
	Coins Coins `json:"coins"`
}

func (client *WhatToMineClient) GetCoins(hashRate, power, poolFee float64) (Coins, error) {
	response := coinsResponse{}
	req := &coinsRequest{HashRate: hashRate, Power: power, PoolFee: poolFee}
	_, err := client.sling.New().Get("coins.json").QueryStruct(req).ReceiveSuccess(&response)
	if err != nil {
		return nil, err
	}
	return response.Coins, nil
}

func (client *WhatToMineClient) GetCoin(id uint64, hashRate, power, poolFee float64) (coin Coin, err error) {
	req := &coinsRequest{HashRate: hashRate, Power: power, PoolFee: poolFee}
	_, err = client.sling.New().Get("coins/" + strconv.FormatUint(id, 10) + ".json").QueryStruct(req).ReceiveSuccess(&coin)
	return
}
