package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"
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
	queryFunction, err := timeSeries.Function()

	if err != nil {
		return nil, fmt.Errorf("error getting intraday data for %s. Ex: %s", ticker, err.Error())
	}

	endpoint := avc.Client.buildRequestPath(map[string]string{
		function: queryFunction,
		symbol: ticker,
	})

	response, err := avc.Client.connection.Request(endpoint)
	if err != nil {
		return nil, err
	}

	timeSeriesKey, err := timeSeries.TimeSeriesKey()
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	return parseTimeSeriesRequestResult(response.Body, timeSeriesKey)
}

// StockTimeSeriesIntraday queries a stock symbols statistics throughout the day.
func (avc AlphaVantageClient) StockTimeSeriesIntraday(timeInterval TimeInterval, ticker string) (*TimeSeriesIntradayResult, error) {
	queryInterval, err := timeInterval.Interval()

	if err != nil {
		return nil, fmt.Errorf("error getting intraday data for %s. Ex: %s", ticker, err.Error())
	}

	endpoint := avc.Client.buildRequestPath(map[string]string{
		function: "TIME_SERIES_INTRADAY",
		interval: queryInterval,
		symbol: ticker,
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
	Interval      string
	OutputSize    string
	TimeZone      string
}

type TimeSeriesDataRaw struct {
    Open      string `json:"1. open"`
    High      string `json:"2. high"`
    Low       string `json:"3. low"`
    Close     string `json:"4. close"`
    Volume    string `json:"5. volume"`
}

type TimeSeriesData struct {
	Timestamp time.Time
    Open      float64
    High      float64
    Low       float64
    Close     float64
    Volume    float64
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
	var rawMetaData TimeSeriesIntradayMetaDataRaw
    if err := json.Unmarshal(raw["Meta Data"], &rawMetaData); err != nil {
        return nil, fmt.Errorf("error unmarshaling metadata: %w", err)
    }

	lastRefreshed, err := parseDate(rawMetaData.LastRefreshed)
	if err != nil {
		return nil, err
	}

	tsmd := TimeSeriesMetaData{
		Information:   rawMetaData.Information,
		Symbol:        rawMetaData.Information,
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
		Symbol:        rawMetaData.Information,
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
    var timeSeriesElements map[string]TimeSeriesDataRaw
    if err := json.Unmarshal(raw[key], &timeSeriesElements); err != nil { // error here
        return nil, fmt.Errorf("error unmarshaling time series: %w", err)
    }

    timeSeries := make([]*TimeSeriesData, 0, len(timeSeriesElements))
    for timestampStr, data := range timeSeriesElements {
        
		timestamp, err := parseDate(timestampStr)
        if err != nil {
            return nil, fmt.Errorf("error converting TIMESTAMP from string to time.Time: %w", err)
        }
        
		open, err := parseFloat(data.Open)
		if err != nil {
			return nil, fmt.Errorf("error converting OPEN from string to float: %w", err)
		}

		high, err := parseFloat(data.High)
		if err != nil {
			return nil, fmt.Errorf("error converting HIGH from string to float: %w", err)
		}

		low, err := parseFloat(data.Low)
		if err != nil {
			return nil, fmt.Errorf("error converting LOW from string to float: %w", err)
		}

		close, err := parseFloat(data.Close)
		if err != nil {
			return nil, fmt.Errorf("error converting CLOSE from string to float: %w", err)
		}

		volume, err := parseFloat(data.Volume)
		if err != nil {
			return nil, fmt.Errorf("error converting VOLUME from string to float: %w", err)
		}


        timeSeries = append(timeSeries, &TimeSeriesData{
            Timestamp: timestamp,
            Open:      open,
            High:      high,
            Low:       low,
            Close:     close,
            Volume:    volume,
        })
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

func parseFloat(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}