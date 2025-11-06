package api

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	u "mc.data"
	m "mc.data/models"
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


func Test_AlphaVantage_StockIntradayTimeSeries(t *testing.T) {
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

	// meta data
	metaInfoEx := "Intraday (60min) open, high, low, close prices and volume"
	assertAreEqual(t, "information", metaInfoEx, res.MetaData.Information.Ptr())
	assertAreEqual(t, "symbol", ticker, res.MetaData.Symbol.Ptr())

	targetDate := time.Date(2025, time.October, 31, 19, 0, 0, 0, location)
	if targetDate.Compare(res.MetaData.LastRefreshed) == 1 { // time is before the actual
		t.Fatalf("error parsing meta data last refreshed date, %s", res.MetaData.LastRefreshed)
	}

	assertAreEqual(t, "interval", "60min", res.MetaData.Interval.Ptr())
	assertAreEqual(t, "output size", defaultOutputSize, res.MetaData.OutputSize.Ptr())
	assertAreEqual(t, "time zone", "US/Eastern", res.MetaData.TimeZone.Ptr())

	// time series element tieout
	f := func(e *m.TimeSeriesData) bool { return targetDate.Compare(e.Timestamp) == 0 }
	s, err := u.FilterSingle(res.TimeSeries, f)
	if err != nil {
		t.Fatalf("error filtering single time series element: %v", err)
	}
	if s == nil {
		t.Fatalf("error filtering single time series element, resulted in nil")
	}

	jsonData, _ := json.MarshalIndent(s, "", "  ")
	t.Logf("JSON: %s", jsonData)

	assertAreEqual(t, "open", 270.10, s.Open.Ptr())
	assertAreEqual(t, "high", 270.14, s.High.Ptr())
	assertAreEqual(t, "low", 269.90, s.Low.Ptr())
	assertAreEqual(t, "close", 269.9995, s.Close.Ptr())
	assertNillability(t, "adjusted close", true, s.AdjustedClose.Ptr())
	assertAreEqual(t, "volume", float64(36838), s.Volume.Ptr())
	assertNillability(t, "dividend amount", true, s.DividendAmount.Ptr())

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
	
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("error parsing time zone: %s", err)
	}

	// meta data
	metaInfoEx := "Weekly Adjusted Prices and Volumes" 
	assertAreEqual(t, "information", metaInfoEx, res.MetaData.Information.Ptr())
	assertAreEqual(t, "symbol", ticker, res.MetaData.Symbol.Ptr())

	targetDate := time.Date(2025, time.October, 31, 0, 0, 0, 0, location)
	if targetDate.Compare(res.MetaData.LastRefreshed) == 1 { // time is before the actual
		t.Fatalf("error parsing meta data last refreshed date, %s", res.MetaData.LastRefreshed)
	}

	assertNillability(t, "interval", true, res.MetaData.Interval.Ptr())
	assertNillability(t, "output size", true, res.MetaData.OutputSize.Ptr())
	assertAreEqual(t, "time zone", "US/Eastern", res.MetaData.TimeZone.Ptr())

	// time series element tieout
	f := func(e *m.TimeSeriesData) bool { return targetDate.Compare(e.Timestamp) == 0 }
	s, err := u.FilterSingle(res.TimeSeries, f)
	if err != nil {
		t.Fatalf("error filtering single time series element: %v", err)
	}
	if s == nil {
		t.Fatalf("error filtering single time series element, resulted in nil")
	}

	jsonData, _ := json.MarshalIndent(s, "", "  ")
	t.Logf("JSON: %s", jsonData)

	assertAreEqual(t, "open", 264.88, s.Open.Ptr())
	assertAreEqual(t, "high", 277.320, s.High.Ptr())
	assertAreEqual(t, "low", 264.6501, s.Low.Ptr())
	assertAreEqual(t, "close", 270.37, s.Close.Ptr())
	assertAreEqual(t, "adjusted close", 270.37, s.AdjustedClose.Ptr())
	assertAreEqual(t, "volume", float64(293563310), s.Volume.Ptr())
	assertAreEqual(t, "dividend amount", 0, s.DividendAmount.Ptr())
}

func assertAreEqual[T comparable](t *testing.T, name string, expected T, actual *T) {
    t.Helper()
    if actual == nil {
        t.Fatalf("error parsing %s, attributed value was nil", name)
    }
    if expected != *actual {
        t.Fatalf("value mismatch for %s, expected %v, got %v", name, expected, *actual)
    }
}

func assertNillability[T comparable](t *testing.T, name string, expected bool, actual *T) {
	t.Helper()
	if (actual == nil) != expected {
        t.Fatalf("value mismatch for %s, expected %v, got %v", name, expected, (actual == nil))
	}
}