package api

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/guregu/null/v6"

	u "mc.data"
)

// public
const (
	HostDefault = "www.alphavantage.co"
)

// private
const (
	schemeHttps = "https"

	apiKey     = "apikey"
	dataType   = "datatype"
	outputSize = "outputsize"
	symbol     = "symbol"
	function   = "function"
	interval   = "interval"

	defaultoutputSize = "compact"
	defaultDataType    = "json"

	query = "query"

	requestTimeout = time.Second * 30
	
	Open           = "Open"
	High           = "High"
	Low            = "Low"
	Close          = "Close"
	AdjustedClose  = "AdjustedClose"
	Volume         = "Volume"
	DividendAmount = "DividendAmount"
)

var (
	timeSeriesDateFormats = []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
	}
)

type AlphaVantageClient struct {
	*Client
}

func GetClient(apiKey string) AlphaVantageClient {
	return AlphaVantageClient{
		ClientFactory(HostDefault, apiKey, requestTimeout),
	}
}

// StockTimeSeries queries a time series at a specific interval
func (avc AlphaVantageClient) StockTimeSeries(timeSeries TimeSeries, ticker string) (*TimeSeriesResult, error) {
	endpoint := avc.Client.buildRequestPath(map[string]string{
		function: timeSeries.Function(),
		symbol:   ticker,
	})

	response, err := avc.Client.connection.Request(endpoint)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	return parseTimeSeriesRequestResult(response.Body, timeSeries.TimeSeriesKey())
}

// StockTimeSeriesIntraday queries a stock symbols statistics throughout the day.
func (avc AlphaVantageClient) StockTimeSeriesIntraday(timeInterval TimeInterval, ticker string) (*TimeSeriesIntradayResult, error) {
	endpoint := avc.Client.buildRequestPath(map[string]string{
		function: "TIME_SERIES_INTRADAY",
		interval: timeInterval.Interval(),
		symbol:   ticker,
	})

	response, err := avc.Client.connection.Request(endpoint)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	
	return parseTimeSeriesIntradayRequestResult(response.Body)
}

func (c *Client) buildRequestPath(params map[string]string) *url.URL {
	// build our URL
	endpoint := &url.URL{}
	endpoint.Path = query

	// base parameters
	query := endpoint.Query()
	query.Set(apiKey, c.apiKey)
	query.Set(dataType, defaultDataType)
	query.Set(outputSize, defaultoutputSize)

	// additional parameters
	for key, value := range params {
		query.Set(key, value)
	}

	endpoint.RawQuery = query.Encode()

	return endpoint
}

type TimeSeriesResult struct {
	MetaData TimeSeriesMetaData
	TimeSeries []*TimeSeriesData
}

type TimeSeriesIntradayResult struct {
    MetaData   TimeSeriesIntradayMetaData
    TimeSeries []*TimeSeriesData
}

type TimeSeriesMetaDataRaw struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	TimeZone      string `json:"4. Time Zone"`
}

type TimeSeriesMetaData struct {
	Information   string
	Symbol        string
	LastRefreshed time.Time
	TimeZone      string
}

type TimeSeriesIntradayMetaDataRaw struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	Interval      string `json:"4. Interval"`
	OutputSize    string `json:"5. Output Size"`
	TimeZone      string `json:"6. Time Zone"`
}

type TimeSeriesIntradayMetaData struct {
	Information   string
	Symbol        string
	LastRefreshed time.Time 
	Interval      string // can probably just aggregate to a single kind
	OutputSize    string // can probably just aggregate to a single kind
	TimeZone      string
}

type TimeSeriesData struct {
	Timestamp      time.Time
    Open           null.Float
    High           null.Float
    Low            null.Float
    Close          null.Float
	AdjustedClose  null.Float
	Volume         null.Float
	DividendAmount null.Float
}

var timeSeriesDataResultKeys = map[string]string{
    Open:           ". Open",
    High:           ". High",
    Low:            ". Low",
	Close:          ". Close",
	AdjustedClose:  ". Adjusted Close",
	Volume:         ". Volume",
	DividendAmount: ". Dividend Amount",
}

func parseTimeSeriesRequestResult(reader io.Reader, timeSeriesKey string) (*TimeSeriesResult, error) {
    body, err := io.ReadAll(reader)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %w", err)
    }

	// converting to a <string, raw message> map
    var raw map[string]json.RawMessage
    if err := json.Unmarshal(body, &raw); err != nil {
        return nil, fmt.Errorf("error unmarshaling response: %w", err)
    }
	
	// meta data
	var rawMetaData TimeSeriesMetaDataRaw
    if err := json.Unmarshal(raw["Meta Data"], &rawMetaData); err != nil {
        return nil, fmt.Errorf("error unmarshaling metadata: %w", err)
    }

	lastRefreshed, err := parseDate(rawMetaData.LastRefreshed)
	if err != nil {
		return nil, err
	}

	tsmd := TimeSeriesMetaData{
		Information:   rawMetaData.Information,
		Symbol:        rawMetaData.Symbol,
		LastRefreshed: lastRefreshed,
		TimeZone:      rawMetaData.TimeZone,
	}

	ts, err:= parseTimeSeries(raw, timeSeriesKey)
	if err != nil {
		return nil, err
	}

	return &TimeSeriesResult{
		MetaData: tsmd,
		TimeSeries: ts,
		}, nil
}

func parseTimeSeriesIntradayRequestResult(reader io.Reader) (*TimeSeriesIntradayResult, error) {
    body, err := io.ReadAll(reader)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %w", err)
    }
    
	// converting to a <string, raw message> map
    var raw map[string]json.RawMessage
    if err := json.Unmarshal(body, &raw); err != nil {
        return nil, fmt.Errorf("error unmarshaling response: %w", err)
    }
    
	var result TimeSeriesIntradayResult

	// meta data
    var rawMetaData TimeSeriesIntradayMetaDataRaw
    if err := json.Unmarshal(raw["Meta Data"], &rawMetaData); err != nil {
        return nil, fmt.Errorf("error unmarshaling metadata: %w", err)
    }

	lastRefreshed, err := parseDate(rawMetaData.LastRefreshed)
	if err != nil {
		return nil, err
	}

	result.MetaData = TimeSeriesIntradayMetaData{
		Information:   rawMetaData.Information,
		Symbol:        rawMetaData.Symbol,
		LastRefreshed: lastRefreshed,
		Interval:      rawMetaData.Interval,
		OutputSize:    rawMetaData.OutputSize,
		TimeZone:      rawMetaData.TimeZone,
	}

	// time series data
    timeSeriesKey := fmt.Sprintf("Time Series (%s)", result.MetaData.Interval)
	ts, err := parseTimeSeries(raw, timeSeriesKey)

	if err != nil {
		return nil, err
	}

	result.TimeSeries = ts
    
    return &result, nil
}

func parseTimeSeries(raw map[string]json.RawMessage, key string) ([]*TimeSeriesData, error) {
    var timeSeriesElements map[string]map[string]string
    if err := json.Unmarshal(raw[key], &timeSeriesElements); err != nil {
        return nil, fmt.Errorf("error unmarshaling time series: %w", err)
    }

    timeSeries := make([]*TimeSeriesData, 0, len(timeSeriesElements))
	lookup := make(map[string]string) 	// <json result key string, time series data attribtue name>
    for timeSeriesKey, timeSeriesValue := range timeSeriesElements {
		if len(lookup) == 0 {
			avResponseValueHeaders := slices.Collect(maps.Keys(timeSeriesValue))
			for key, value := range timeSeriesDataResultKeys {
				f := func(s string) bool { 
					return strings.HasSuffix(strings.ToLower(s), strings.ToLower(value))
				} 
				if jsonKey, err := u.FilterSingle(avResponseValueHeaders, f); err == nil {
					lookup[jsonKey] = key // <json key, result attribute>
				}
			}
			if len(lookup) == 0 {
				return nil, fmt.Errorf("error generating key value map from av response object. Available headers: %v", avResponseValueHeaders)
			}
		}

		tsd := TimeSeriesData{}

        timestamp, err := parseDate(timeSeriesKey)
        if err != nil {
            return nil, fmt.Errorf("error converting TIMESTAMP from string to time.Time: %w", err)
        }

		tsd.Timestamp = timestamp

		v := reflect.ValueOf(&tsd).Elem()

		for jsonKey, structAttribute := range lookup{
			pv := parseFloat(timeSeriesValue[jsonKey]);
			field := v.FieldByName(structAttribute)

			if !field.IsValid() {
				return nil, fmt.Errorf("field %s does not exist", structAttribute)
			}
		
			if !field.CanSet() {
				return nil, fmt.Errorf("field %s cannot be set", structAttribute)
			}

			field.Set(reflect.ValueOf(pv))
		}

        timeSeries = append(timeSeries, &tsd)
    }

    return timeSeries, nil
}

func parseDate(dateString string) (time.Time, error) {
	for _, format := range timeSeriesDateFormats {
		t, err := time.Parse(format, dateString)
		if err != nil {
			continue
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("error converting date %s to time.Time", dateString)
}

func parseFloat(val string) null.Float {
	if val != "" {
		if conv, err := strconv.ParseFloat(val, 64); err == nil {
			return null.NewFloat(conv, true)
		}
	}
	return null.NewFloat(0, false)
}