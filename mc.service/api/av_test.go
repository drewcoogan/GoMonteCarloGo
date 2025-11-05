package api

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	u "mc.data"
)

const (
	avKeyName = "ALPHAVANTAGE_API_KEY"
)

func Test_AlphaVantage_GetApiKey(t *testing.T) {
	err := godotenv.Load("testenv");
	if err != nil {
		t.Fatalf("error loading environment: %s", err)
	}

	actual := os.Getenv(avKeyName)
	if actual == "" {
		t.Fatalf("error finding key %s in .env", avKeyName)
	}

	expected := "av-test-api-key"
	if actual != expected {
		t.Fatalf("error validating key. expected %s, got %s", expected, actual)
	}
}

func getApiKey(t *testing.T) string {
	t.Helper()
	err := godotenv.Load("../.env");
	if err != nil {
		t.Errorf("error loading environment: %s", err)
	}

	return os.Getenv(avKeyName)
}


func Test_AplhaVantage_StockIntradayTimeSeries(t *testing.T) {
	ticker := "AAPL"
	apiKey := getApiKey(t)
	c := GetClient(apiKey)
	res, err := c.StockTimeSeriesIntraday(TimeIntervalSixtyMinute, ticker)

	if err != nil {
		t.Fatalf("error getting stock time series: %s", err)
	}

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("error parsing time zone: %s", err)
	}

	metaInfoEx := "Intraday (60min) open, high, low, close prices and volume"
	if metaInfoEx != res.MetaData.Information.String {
		t.Fatalf("error parsing meta data information, expected %s, got %s", metaInfoEx, res.MetaData.Information.String)
	}

	metaSymbolAct := res.MetaData.Symbol
	if ticker != metaSymbolAct.String {
		t.Fatalf("error parsing meta data symbol, expected %s, got %s", ticker, metaSymbolAct.String)
	}

	targetDate := time.Date(2025, time.October, 31, 19, 0, 0, 0, location)
	if targetDate.Compare(res.MetaData.LastRefreshed) == 1 { // time is before the actual
		t.Fatalf("error parsing meta data last refreshed date, %s", res.MetaData.LastRefreshed)
	}

	metaTimeZoneEx := "US/Eastern"
	if metaTimeZoneEx != res.MetaData.TimeZone.String {
		t.Fatalf("error parsing meta data time zone, expected %s, got %s", metaTimeZoneEx, res.MetaData.TimeZone.String)
	}

	f := func(e *TimeSeriesData) bool { return targetDate.Compare(e.Timestamp) == 0 }
	s, err := u.FilterSingle(res.TimeSeries, f)
	if err != nil {
		t.Fatalf("error filtering single time series element: %v", err)
	}
	if s == nil {
		t.Fatalf("error filtering single time series element, resulted in nil")
	}

	// Print the struct with all fields visible, only if test fails
	//t.Logf("TimeSeriesData: %+v", s)
	jsonData, _ := json.MarshalIndent(s, "", "  ")
	t.Logf("JSON: %s", jsonData)

	assertExpectation(t, 270.10, s.Open.Ptr(), "open")
	assertExpectation(t, 270.14, s.High.Ptr(), "high")
	assertExpectation(t, 269.90, s.Low.Ptr(), "low")
	assertExpectation(t, 269.9995, s.Close.Ptr(), "close")
	assertExpectation(t, float64(36838), s.Volume.Ptr(), "volume")

	if s.DividendAmount.Ptr() != nil {
		t.Fatalf("error expecting nil dividend amount, got %f instead", s.DividendAmount.Float64)
	}

}

func Test_AlphaVantage_StockTimeSeries(t *testing.T) {
	ticker := "AAPL"
	apiKey := getApiKey(t)
	c := GetClient(apiKey)
	res, err := c.StockTimeSeries(TimeSeriesWeeklyAdjusted, ticker)

	if err != nil {
		t.Fatalf("error getting stock time series: %s", err)
	}

	metaInfoEx := "Weekly Adjusted Prices and Volumes" 
	if metaInfoEx != res.MetaData.Information.String {
		t.Fatalf("error parsing meta data information, expected %s, got %s", metaInfoEx, res.MetaData.Information.String)
	}

	metaSymbolAct := res.MetaData.Symbol
	if ticker != metaSymbolAct.String {
		t.Fatalf("error parsing meta data symbol, expected %s, got %s", ticker, metaSymbolAct.String)
	}

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("error parsing time zone: %s", err)
	}

	targetDate := time.Date(2025, time.October, 31, 0, 0, 0, 0, location)
	if targetDate.Compare(res.MetaData.LastRefreshed) == 1 { // time is before the actual
		t.Fatalf("error parsing meta data last refreshed date, %s", res.MetaData.LastRefreshed)
	}

	metaTimeZoneEx := "US/Eastern"
	if metaTimeZoneEx != res.MetaData.TimeZone.String {
		t.Fatalf("error parsing meta data time zone, expected %s, got %s", metaTimeZoneEx, res.MetaData.TimeZone.String)
	}

	f := func(e *TimeSeriesData) bool { return targetDate.Compare(e.Timestamp) == 0 }
	s, err := u.FilterSingle(res.TimeSeries, f)
	if err != nil {
		t.Fatalf("error filtering single time series element: %v", err)
	}
	if s == nil {
		t.Fatalf("error filtering single time series element, resulted in nil")
	}

	// Print the struct with all fields visible, only if test fails
	//t.Logf("TimeSeriesData: %+v", s)
	jsonData, _ := json.MarshalIndent(s, "", "  ")
	t.Logf("JSON: %s", jsonData)

	assertExpectation(t, 264.88, s.Open.Ptr(), "open")
	assertExpectation(t, 277.320, s.High.Ptr(), "high")
	assertExpectation(t, 264.6501, s.Low.Ptr(), "low")
	assertExpectation(t, 270.37, s.Close.Ptr(), "close")
	assertExpectation(t, float64(293563310), s.Volume.Ptr(), "volume")
	assertExpectation(t, 0, s.DividendAmount.Ptr(), "dividend amount")
}

func assertExpectation(t *testing.T, expected float64, actual *float64, name string) {
	t.Helper()
	
	if actual == nil {
		t.Fatalf("error parsing %s, attributed value was nil", name)
	}
	
	if expected != *actual {
		t.Fatalf("value mismatch for %s, expected %f, got %f", name, expected, *actual)
	}
}