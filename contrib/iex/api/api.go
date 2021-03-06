package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

const (
	base      = "https://api.iextrading.com/1.0"
	BatchSize = 100
)

var (
	NY, _ = time.LoadLocation("America/New_York")
)

type GetBarsResponse map[string]ChartResponse

type ChartResponse struct {
	Chart []Chart `json:"chart"`
}

type Chart struct {
	Date                 string  `json:"date"`
	Minute               string  `json:"minute"`
	Label                string  `json:"label"`
	High                 float32 `json:"high"`
	Low                  float32 `json:"low"`
	Average              float64 `json:"average"`
	Volume               int32   `json:"volume"`
	Notional             float64 `json:"notional"`
	NumberOfTrades       int     `json:"numberOfTrades"`
	MarketHigh           float64 `json:"marketHigh"`
	MarketLow            float64 `json:"marketLow"`
	MarketAverage        float64 `json:"marketAverage"`
	MarketVolume         int     `json:"marketVolume"`
	MarketNotional       float64 `json:"marketNotional"`
	MarketNumberOfTrades int     `json:"marketNumberOfTrades"`
	Open                 float32 `json:"open"`
	Close                float32 `json:"close"`
	MarketOpen           float64 `json:"marketOpen,omitempty"`
	MarketClose          float64 `json:"marketClose,omitempty"`
	ChangeOverTime       float64 `json:"changeOverTime"`
	MarketChangeOverTime float64 `json:"marketChangeOverTime"`
}

func (c *Chart) GetTimestamp() (ts time.Time, err error) {
	if c.Minute == "" {
		// daily bar
		ts, err = time.ParseInLocation("2006-01-02", c.Date, NY)
	} else {
		// intraday bar
		tStr := fmt.Sprintf("%v %v", c.Date, c.Minute)
		ts, err = time.ParseInLocation("20060102 15:04", tStr, NY)
	}
	return
}

func SupportedRange(r string) bool {
	switch r {
	case "5y":
	case "2y":
	case "1y":
	case "ytd":
	case "6m":
	case "3m":
	case "1m":
	case "1d":
	case "date":
	case "dynamic":
	default:
		return false
	}
	return true
}

func GetBars(symbols []string, barRange string, limit *int, retries int) (*GetBarsResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/stock/market/batch", base))
	if err != nil {
		return nil, err
	}

	if len(symbols) == 0 {
		return &GetBarsResponse{}, nil
	}

	q := u.Query()

	q.Set("symbols", strings.Join(symbols, ","))
	q.Set("types", "chart")

	if SupportedRange(barRange) {
		q.Set("range", barRange)
	} else {
		return nil, fmt.Errorf("%v is not a supported bar range", barRange)
	}

	if limit != nil && *limit > 0 {
		q.Set("chartLast", strconv.FormatInt(int64(*limit), 10))
	}

	u.RawQuery = q.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		if retries > 0 {
			<-time.After(time.Second)
			return GetBars(symbols, barRange, limit, retries-1)
		}

		return nil, fmt.Errorf("retry count exceeded")
	}

	var resp GetBarsResponse

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type ListSymbolsResponse []struct {
	Symbol    string `csv:"symbol"`
	Name      string `csv:"name"`
	Date      string `csv:"date"`
	IsEnabled bool   `csv:"isEnabled"`
	Type      string `csv:"type"`
	IexID     int64  `csv:"iexId"`
}

func ListSymbols() (*ListSymbolsResponse, error) {
	url := fmt.Sprintf("%s/ref-data/symbols?format=csv", base)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode > http.StatusMultipleChoices {
		return nil, fmt.Errorf("status code %v", res.StatusCode)
	}

	var resp ListSymbolsResponse

	if err = gocsv.Unmarshal(res.Body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
