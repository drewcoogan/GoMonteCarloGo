package alpha_vantage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	e "mc.data/extensions"
	m "mc.data/models"
	c "mc.service/api"
)

// public
const (
	HostDefault = "www.alphavantage.co"
)

// private
const (
	// default query parameters
	defaultOutputSize = "Compact"
	defaultDataType   = "JSON"
	defaultTimeout    = time.Second * 30

	// api request elements
	query    = "query"
	symbol   = "symbol"
	function = "function"
	interval = "interval"
)

var (
	timeSeriesDateFormats = []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
	}

	ohlcvResultKeys = map[string]string{
		"Open":   ". Open",
		"High":   ". High",
		"Low":    ". Low",
		"Close":  ". Close",
		"Volume": ". Volume",
	}
)


type AlphaVantageClient struct{
	*c.Client
}

func GetClient(apiKey string) AlphaVantageClient {
	return AlphaVantageClient{
		c.ClientFactory(HostDefault, apiKey, defaultTimeout),
	}
}

// https://www.alphavantage.co/documentation/#weeklyadj
func (avc *AlphaVantageClient) GetStockWeeklyAdjustedMetrics(ticker string) (*m.TimeSeriesResult, error) {
	if avc == nil {
		panic("alpha vantage client has not been set.")
	}

	endpoint := avc.buildRequestPath(map[string]string{
		function: "TIME_SERIES_WEEKLY_ADJUSTED",
		symbol:   ticker,
	})

	response, err := avc.Client.Connection.Request(endpoint)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	
	raw, err := parseRawJson(response.Body)
	if err != nil {
		return nil, err
	}

	metaData, timeZone, err := parseMetaData(raw)
	if err != nil {
		return nil, err
	}

	timeSeriesData, err := parseTimeSeriesDataResult(raw, "Weekly Adjusted Time Series", timeZone)
	if err != nil {
		return nil, err
	}

	return &m.TimeSeriesResult{
		Metadata: metaData,
		TimeSeries: timeSeriesData,
	}, nil

}

// StockTimeSeriesIntraday queries a stock symbols statistics throughout the day.
func (avc *AlphaVantageClient) GetStockIntradayMetrics(ticker string) (*m.TimeSeriesIntradayResult, error) {
	endpoint := avc.buildRequestPath(map[string]string{
		function: "TIME_SERIES_INTRADAY",
		interval: "5min",
		symbol:   ticker,
	})

	response, err := avc.Client.Connection.Request(endpoint)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	raw, err := parseRawJson(response.Body)
	if err != nil {
		return nil, err
	}

	metaData, timeZone, err := parseMetaData(raw)
	if err != nil {
		return nil, err
	}

	timeSeriesData, err := parseTimeSeriesIntradayDataResult(raw, "Time Series (5min)", timeZone)
	if err != nil {
		return nil, err
	}

	return &m.TimeSeriesIntradayResult{
		Metadata: metaData,
		TimeSeries: timeSeriesData,
	}, nil
}

func (avc *AlphaVantageClient) buildRequestPath(params map[string]string) *url.URL {
	// build our URL
	endpoint := &url.URL{}
	endpoint.Path = query

	// base parameters
	query := endpoint.Query()
	query.Set("apikey", avc.Client.ApiKey)
	query.Set("datatype", defaultDataType)
	query.Set("outputsize", defaultOutputSize)

	// additional parameters
	for key, value := range params {
		query.Set(key, value)
	}

	endpoint.RawQuery = query.Encode()

	return endpoint
}

func parseRawJson(reader io.Reader) (raw map[string]json.RawMessage, err error) {
    body, err := io.ReadAll(reader)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %w", err)
    }

	// converting to a <string, raw message> map
    if err := json.Unmarshal(body, &raw); err != nil {
        return nil, fmt.Errorf("error unmarshaling response: %w", err)
    }

	return
}

func parseMetaData(raw map[string]json.RawMessage) (*m.TimeSeriesMetadata, *time.Location, error) {
	var metadataElements map[string]string
	if err := json.Unmarshal(raw["Meta Data"], &metadataElements); err != nil {
		return nil, nil, fmt.Errorf("error unmarshaling meta data: %w", err)
	}

	metaDataKeys := slices.Collect(maps.Keys(metadataElements))

	// parse symbol
	sf := func(s string) bool { return strings.HasSuffix(s, ". Symbol") }
	symbolKey, err := e.FilterSingle(metaDataKeys, sf)
	if err != nil {
		return nil, nil, fmt.Errorf("error extracting symbol for meta data")
	}

	// parse time zone
	tzf := func(s string) bool { return strings.HasSuffix(s, ". Time Zone") }
	timeZoneKey, err := e.FilterSingle(metaDataKeys, tzf)
	if err != nil {
		return nil, nil, fmt.Errorf("error extracting time zone for meta data")
	}

	timeZone, err := getTimeZone(metadataElements[timeZoneKey])
	if err != nil {
		return nil, nil, fmt.Errorf("error converting time zone key %s, to time.Location: %w", metadataElements[timeZoneKey], err)
	}

	// parse last refreshed
	lrf := func(s string) bool { return strings.HasSuffix(s, ". Last Refreshed") }
	lastRefreshedKey, err := e.FilterSingle(metaDataKeys, lrf)
	if err != nil {
		return nil, nil, fmt.Errorf("error extracting last refreshed date")
	}

	lastRefreshed, err := parseDate(metadataElements[lastRefreshedKey], timeZone)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing last refreshed date")
	}

	res := m.TimeSeriesMetadata{
		Symbol:        metadataElements[symbolKey],
		LastRefreshed: lastRefreshed,
	}

	return &res, timeZone, nil
}

func parseTimeSeriesDataResult(raw map[string]json.RawMessage, key string, location *time.Location) ([]*m.TimeSeriesData, error) {
    var timeSeriesElements map[string]map[string]string
    if err := json.Unmarshal(raw[key], &timeSeriesElements); err != nil {
        return nil, fmt.Errorf("error unmarshaling time series: %w", err)
    }

	// populate the lookups
	var firstValue map[string]string
	for _, v := range timeSeriesElements {
		firstValue = v
		break
	}
	
	ohlcvLookup, err := getLookupKey(ohlcvResultKeys, firstValue)
	if err != nil {
		return nil, err
	}

	// parse adjusted close key in raw json lookup
	acf := func(s string) bool { return strings.HasSuffix(s, ". adjusted close") }
	adjustedCloseKey, err := e.FilterSingle(slices.Collect(maps.Keys(firstValue)), acf)
	if err != nil {
		return nil, fmt.Errorf("error extracting adjusted close key for time series")
	}

	// get dividend amount key in raw json lookup
	daf := func(s string) bool { return strings.HasSuffix(s, ". dividend amount")}
	dividendAmountKey, err := e.FilterSingle(slices.Collect(maps.Keys(firstValue)), daf)
	if err != nil {
		return nil, fmt.Errorf("error extracting dividend amount key for time series")
	}

	timeSeries := make([]*m.TimeSeriesData, 0, len(timeSeriesElements))
	for timeSeriesKey, timeSeriesValue := range timeSeriesElements{
		// get timestamp
		timestamp, err := parseDate(timeSeriesKey, location)
        if err != nil {
            return nil, fmt.Errorf("error converting TIMESTAMP from string to time.Time: %w", err)
        }

		// get OHLCV
		ohlcv, err := parseOHLCV(timeSeriesValue, ohlcvLookup)
		if err != nil {
			return nil, fmt.Errorf("error parsing OHLCV: %w", err)
		}

		timeSeries = append(timeSeries, &m.TimeSeriesData{
			Timestamp:      timestamp,
			OHLCV:          ohlcv,
			AdjustedClose:  parseFloat(timeSeriesValue[adjustedCloseKey]),
			DividendAmount: parseFloat(timeSeriesValue[dividendAmountKey]),
		})
	}

	return timeSeries, nil
}

func parseTimeSeriesIntradayDataResult(raw map[string]json.RawMessage, key string, location *time.Location) ([]*m.TimeSeriesIntradayData, error) {
    var timeSeriesElements map[string]map[string]string
    if err := json.Unmarshal(raw[key], &timeSeriesElements); err != nil {
        return nil, fmt.Errorf("error unmarshaling time series: %w", err)
    }

	// populate the lookups
	var firstValue map[string]string
	for _, v := range timeSeriesElements {
		firstValue = v
		break
	}
	
	ohlcvLookup, err := getLookupKey(ohlcvResultKeys, firstValue)
	if err != nil {
		return nil, err
	}

	timeSeries := make([]*m.TimeSeriesIntradayData, 0, len(timeSeriesElements))
	for timeSeriesKey, timeSeriesValue := range timeSeriesElements{
		// get timestamp
		timestamp, err := parseDate(timeSeriesKey, location)
        if err != nil {
            return nil, fmt.Errorf("error converting TIMESTAMP from string to time.Time: %w", err)
        }

		// get OHLCV
		ohlcv, err := parseOHLCV(timeSeriesValue, ohlcvLookup)
		if err != nil {
			return nil, fmt.Errorf("error parsing OHLCV: %w", err)
		}

		timeSeries = append(timeSeries, &m.TimeSeriesIntradayData{
			Timestamp: timestamp,
			OHLCV:     ohlcv,
		})
	}

	return timeSeries, nil
}

func parseOHLCV(value, lookup map[string]string) (res m.TimeSeriesOHLCV, err error) {
	v := reflect.ValueOf(&res).Elem()
	for jsonKey, structAttribute := range lookup {
		field := v.FieldByName(structAttribute)
		if !field.IsValid() {
			return res, fmt.Errorf("field %s does not exist", structAttribute)
		}
		if !field.CanSet() {
			return res, fmt.Errorf("field %s cannot be set", structAttribute)
		}

		pv := parseFloat(value[jsonKey])
		field.Set(reflect.ValueOf(pv))
	}
	return
}

func getLookupKey(expectedKeys, values map[string]string) (map[string]string, error) {
    res := make(map[string]string)
    responseValueHeaders := slices.Collect(maps.Keys(values))
    
    for key, value := range expectedKeys {
        f := func(s string) bool { 
            return strings.HasSuffix(strings.ToLower(s), strings.ToLower(value))
        } 
        if jsonKey, err := e.FilterSingle(responseValueHeaders, f); err == nil {
            res[jsonKey] = key
        }
    }
    
    if len(res) == 0 {
        ex := slices.Collect(maps.Keys(values))
        return nil, fmt.Errorf("error generating key value map from av response object. Available headers: %v", ex)
    }
    
    return res, nil
}

func getTimeZone(location string) (*time.Location, error) {
	var loc string
	switch strings.ToUpper(location) {
	case "US/EASTERN":
		loc = "America/New_York"
	default:
		log.Printf("default time zone hit, %s is not recognized", location)
		return time.UTC, nil
	}
	
	res, err := time.LoadLocation(loc)

	if err != nil {
		return nil, fmt.Errorf("error parsing time zone %s in time.LoadLocation", loc)
	}

	return res, nil
}

func parseDate(dateString string, location *time.Location) (time.Time, error) {
	for _, format := range timeSeriesDateFormats {
		t, err := time.ParseInLocation(format, dateString, location)
		if err != nil {
			continue
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("error converting date %s to time.Time", dateString)
}

func parseFloat(val string) float64 {
	if val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}