package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	ranges = [46]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 17, 20, 25, 30, 40, 50, 60, 70, 80, 100, 120, 140, 170, 200, 250, 300, 400, 500, 600, 700, 800, 1000, 1200, 1400, 1700, 2000, 2500, 3000, 4000, 5000, 6000, 7000, 8000, 10000}
)

type zabbix struct {
	Range  int `json:"txfee"`
	Count  int `json:"count"`
	Weight int `json:"weight"`
	Fees   int `json:"fees"`
}

type Data struct {
	Date   time.Time
	Count  []int
	Weight []int
	Fees   []int
}

func ConvertArray(array []interface{}) []int {
	res := make([]int, 0)
	for _, d := range array {
		res = append(res, int(d.(float64)))
	}
	return res
}

func (d *Data) UnmarshalJSON(data []byte) error {
	var parsed []interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&parsed)
	if err != nil {
		return err
	}
	if len(parsed) != 4 {
		return errors.New("Invalid data format")
	}
	d.Date = time.Unix(int64(parsed[0].(float64)), 0)
	d.Count = ConvertArray(parsed[1].([]interface{}))
	d.Weight = ConvertArray(parsed[2].([]interface{}))
	d.Fees = ConvertArray(parsed[3].([]interface{}))
	return nil
}

func FetchData(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}
		data := strings.Replace(string(body), "call(", "", 1)
		data = strings.Replace(data, ",\n])", "]", 1)
		return data, nil
	} else {
		return "", fmt.Errorf("unexpected response status code: %d", response.StatusCode)
	}
}

func ProcessDatat(data Data) ([]zabbix, error) {
	if len(data.Weight) != len(ranges) {
		return nil, errors.New("Invald data length")
	}
	res := make([]zabbix, 0)
	for idx, r := range ranges {
		res = append(res, zabbix{
			Range:  r,
			Count:  data.Count[idx],
			Weight: data.Weight[idx],
			Fees:   data.Fees[idx],
		})
	}
	return res, nil
}
