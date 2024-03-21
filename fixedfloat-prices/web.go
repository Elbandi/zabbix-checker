package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/elbandi/go-fixedfloat-api"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	debug bool
)

func init() {
	debug = false
}

func dumpRequest(r *http.Request) {
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

func dumpResponse(r *http.Response) {
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

func fetchPage(url string) (*html.Node, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	//	req.Header.Set("User-Agent", "qalandar-fetcher/1.0")
	if debug {
		dumpRequest(req)
	}
	resp, err := client.Do(req)
	if debug {
		dumpResponse(resp)
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	return html.Parse(r)
}

func postPage(url string, payload url.Values) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range payload {
		err := writer.WriteField(k, v[0])
		if err != nil {
			return nil, err
		}
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept-Language", "en-US;q=0.8,en;q=0.7")
	//	req.Header.Set("User-Agent", "qalandar-fetcher/1.0")
	if debug {
		dumpRequest(req)
	}
	resp, err := client.Do(req)
	if debug {
		dumpResponse(resp)
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
	}
	return response, err
}

// GetRate getting the exchange rate of a pair of currencies in the selected direction and type of rate.
func getRate(fromCurrency string, toCurrency string, amount float64) (fixedfloat.Rate, fixedfloat.Rate, error) {
	payload := url.Values{}
	payload.Set("fromCcy", strings.ToUpper(fromCurrency))
	payload.Set("toCcy", strings.ToUpper(toCurrency))
	payload.Set("fromAmount", strconv.FormatFloat(amount, 'f', -1, 64))
	payload.Set("type", "float")
	r, err := postPage("https://ff.io/ajax/exchPrice", payload)
	if err != nil {
		return fixedfloat.Rate{}, fixedfloat.Rate{}, err
	}
	var response struct {
		Code    fixedfloat.Integer `json:"code"`
		Message string             `json:"msg"`
		Data    json.RawMessage    `json:"data"`
	}
	if err = json.Unmarshal(r, &response); err != nil {
		return fixedfloat.Rate{}, fixedfloat.Rate{}, err
	}
	if response.Code != 0 {
		return fixedfloat.Rate{}, fixedfloat.Rate{}, errors.New(response.Message)
	}
	var rates struct {
		From  fixedfloat.Rate `json:"from"`
		To    fixedfloat.Rate `json:"to"`
		Error []string        `json:"errors"`
	}
	if err = json.Unmarshal(response.Data, &rates); err != nil {
		return fixedfloat.Rate{}, fixedfloat.Rate{}, err
	}
	//if len(rates.Error) > 0 {
	//	return fixedfloat.Rate{}, fixedfloat.Rate{}, errors.New(strings.Join(rates.Error, ","))
	//}
	return rates.From, rates.To, nil
}
