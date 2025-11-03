package api

import (
	"fmt"
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

func Test_AlphaVantage_StockTimeSeries(t *testing.T) {
	err := godotenv.Load("../.env");
	if err != nil {
		t.Errorf("error loading environment: %s", err)
	}

	ticker := "AAPL"
	apiKey := os.Getenv(avKeyName)

	c := GetClient(apiKey)
	res, err := c.StockTimeSeries(TimeSeriesWeeklyAdjusted, ticker)

	if err != nil {
		t.Fatalf("error getting stock time series: %s", err)
	}

	metaInfoEx := "Weekly Adjusted Prices and Volumes" 
	if metaInfoEx != res.MetaData.Information {
		t.Fatalf("error parsing meta data information, expected %s, got %s", metaInfoEx, res.MetaData.Information)
	}

	metaSymbolAct := res.MetaData.Symbol
	if ticker != metaSymbolAct {
		t.Fatalf("error parsing meta data symbol, expected %s, got %s", ticker, metaSymbolAct)
	}

	metaLastRefEx := time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC)
	if metaLastRefEx.Compare(res.MetaData.LastRefreshed) == -1 { // time is before the actual
		t.Fatalf("error parsing meta data last refreshed date, %s", res.MetaData.LastRefreshed)
	}

	metaTimeZoneEx := "US/Eastern"
	if metaTimeZoneEx != res.MetaData.TimeZone {
		t.Fatalf("error parsing meta data time zone, expected %s, got %s", metaTimeZoneEx, res.MetaData.TimeZone)
	}

	targetDate := time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC)
	f := func(e *TimeSeriesData) bool { return targetDate.Compare(e.Timestamp) == 0 }
	s, err := u.FilterSingle(res.TimeSeries, f)
	if err != nil || s == nil {
		t.Fatalf("error filtering single time series element: %v", err)
	}

	assertExpectation(264.88, s.Open.Ptr(), "open")
	assertExpectation(277.320, s.High.Ptr(), "high")
	assertExpectation(264.6501, s.Low.Ptr(), "low")
	assertExpectation(270.37, s.Close.Ptr(), "close")
	assertExpectation(float64(293563310), s.Volume.Ptr(), "volume")
}

func assertExpectation(expected float64, actual *float64, name string) error {
	if actual == nil {
		return fmt.Errorf("error parsing %s, attributed value was nil", name)
	}
	
	if expected != *actual {
		return fmt.Errorf("value mismatch for %s, expected %f, got %f", name, expected, *actual)
	}

	return nil
}