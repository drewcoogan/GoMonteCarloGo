package api

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

	"github.com/guregu/null/v6"

	u "mc.data"
)

// public
const (
	HostDefault = "www.alphavantage.co"
)

// private
const (
	apiKey     = "apikey"
	dataType   = "datatype"
	outputSize = "outputsize"
	symbol     = "symbol"
	function   = "function"
	interval   = "interval"

	defaultoutputSize = "compact"
	defaultDataType   = "json"
	query             = "query"

	requestTimeout = time.Second * 30
	
	Information   = "Information"
	Symbol        = "Symbol"
	LastRefreshed = "Last Refreshed"
	Interval      = "Interval"
	OutputSize    = "OutputSize"
	TimeZone      = "TimeZone"

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
func (avc AlphaVantageClient) StockTimeSeriesIntraday(timeInterval TimeInterval, ticker string) (*TimeSeriesResult, error) {
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
	
	return parseTimeSeriesRequestResult(response.Body, "")
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
	MetaData *TimeSeriesMetaData
	TimeSeries []*TimeSeriesData
}

type TimeSeriesMetaData struct {
	Information   null.String
	Symbol        null.String
	LastRefreshed time.Time 
	Interval      null.String
	OutputSize    null.String
	TimeZone      null.String
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

var timeSeriesMetaDataKeys = map[string]string{
	Information:   ". Information",
	Symbol:        ". Symbol",
	Interval:      ". Interval",
	OutputSize:    ". Output Size",
	TimeZone:      ". Time Zone",
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
	m, err := parseMetaData(raw)
	if err != nil {
		return nil, err
	}

	// time zone
	timeZone, err := getTimeZone(m.TimeZone.String)
	if err != nil {
		return nil, err
	}
	
	// time series values
	if timeSeriesKey == "" {
		timeSeriesKey = fmt.Sprintf("Time Series (%s)", m.Interval.String)
	}

	ts, err := parseTimeSeries(raw, timeSeriesKey, timeZone)
	if err != nil {
		return nil, err
	}

	return &TimeSeriesResult{
		MetaData:   m,
		TimeSeries: ts,
		}, nil
}

func parseMetaData(raw map[string]json.RawMessage) (*TimeSeriesMetaData, error) {
	var metaDataElements map[string]string
	if err := json.Unmarshal(raw["Meta Data"], &metaDataElements); err != nil {
		return nil, fmt.Errorf("error unmarshaling meta data: %w", err)
	}

	res := TimeSeriesMetaData{}
	lookup := getLookupKey(timeSeriesMetaDataKeys, metaDataElements)

	// populate simple string fields via reflection using the lookup
	v := reflect.ValueOf(&res).Elem()
	for metaDataKey, metaDataValue := range lookup {
		field := v.FieldByName(metaDataValue)
		if !field.IsValid() {
			return nil, fmt.Errorf("field %s does not exist", metaDataValue)
		}
		if !field.CanSet() {
			return nil, fmt.Errorf("field %s cannot be set", metaDataValue)
		}
		
		v := null.NewString(metaDataElements[metaDataKey], true)
		field.Set(reflect.ValueOf(v))
	}

	// parse time.Time type with a little more care
	f := func(s string) bool { return strings.HasSuffix(s, ". Last Refreshed") }
	lastRefreshedKey, err := u.FilterSingle(slices.Collect(maps.Keys(metaDataElements)), f)
	if err != nil {
		return nil, fmt.Errorf("error extracting last refreshed date")
	}

	lastRefreshed, err := parseDateAsUtc(metaDataElements[lastRefreshedKey])

	if err != nil {
		return nil, fmt.Errorf("error parsing last refreshed date")
	}

	res.LastRefreshed = lastRefreshed

	return &res, nil
}

func parseTimeSeries(raw map[string]json.RawMessage, key string, location *time.Location) ([]*TimeSeriesData, error) {
    var timeSeriesElements map[string]map[string]string
    if err := json.Unmarshal(raw[key], &timeSeriesElements); err != nil {
        return nil, fmt.Errorf("error unmarshaling time series: %w", err)
    }

    timeSeries := make([]*TimeSeriesData, 0, len(timeSeriesElements))
	lookup := make(map[string]string) 	// <json result key string, time series data attribtue name>
    for timeSeriesKey, timeSeriesValue := range timeSeriesElements {
		if len(lookup) == 0 {
			lookup = getLookupKey(timeSeriesDataResultKeys, timeSeriesValue)
			if len(lookup) == 0 {
				ex := slices.Collect(maps.Keys(timeSeriesValue))
				return nil, fmt.Errorf("error generating key value map from av response object. Available headers: %v", ex)
			}
		}

		tsd := TimeSeriesData{}

        timestamp, err := parseDate(timeSeriesKey, location)
        if err != nil {
            return nil, fmt.Errorf("error converting TIMESTAMP from string to time.Time: %w", err)
        }

		tsd.Timestamp = timestamp

		v := reflect.ValueOf(&tsd).Elem()

		for jsonKey, structAttribute := range lookup{
			field := v.FieldByName(structAttribute)
			if !field.IsValid() {
				return nil, fmt.Errorf("field %s does not exist", structAttribute)
			}
			if !field.CanSet() {
				return nil, fmt.Errorf("field %s cannot be set", structAttribute)
			}

			pv := parseFloat(timeSeriesValue[jsonKey]);
			field.Set(reflect.ValueOf(pv))
		}

        timeSeries = append(timeSeries, &tsd)
    }

    return timeSeries, nil
}

func getLookupKey(expectedKeys map[string]string, values map[string]string) map[string]string {
	res := make(map[string]string)
	responseValueHeaders := slices.Collect(maps.Keys(values))
	for key, value := range expectedKeys {
		f := func(s string) bool { 
			return strings.HasSuffix(strings.ToLower(s), strings.ToLower(value))
		} 
		if jsonKey, err := u.FilterSingle(responseValueHeaders, f); err == nil {
			res[jsonKey] = key // <json key, result attribute>
		}
	}
	return res
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

func parseDateAsUtc(dateString string) (time.Time, error) {
	return parseDate(dateString, time.UTC)
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

func parseFloat(val string) null.Float {
	if val != "" {
		if conv, err := strconv.ParseFloat(val, 64); err == nil {
			return null.NewFloat(conv, true)
		}
	}
	return null.NewFloat(0, false)
}