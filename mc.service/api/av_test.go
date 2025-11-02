package api

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
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
	metaInfoAct := res.MetaData.Information
	if metaInfoEx != metaInfoAct {
		t.Fatalf("error parsing meta data information, expected %s, got %s", metaInfoEx, metaInfoAct)
	}

	metaSymbolAct := res.MetaData.Symbol
	if ticker != metaSymbolAct {
		t.Fatalf("error parsing meta data symbol, expected %s, got %s", ticker, metaSymbolAct)
	}

	metaLastRefEx := time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC)
	metaLastRefAct := res.MetaData.LastRefreshed
	if metaLastRefEx.Compare(metaLastRefAct) == -1 { // time is before the actual
		t.Fatalf("error parsing meta data last refreshed date, %s", metaLastRefAct)
	}

	metaTimeZoneEx := "US/Eastern"
	metaTimeZoneAct := res.MetaData.TimeZone
	if metaTimeZoneEx != metaTimeZoneAct {
		t.Fatalf("error parsing meta data time zone, expected %s, got %s", metaTimeZoneEx, metaTimeZoneAct)
	}
}