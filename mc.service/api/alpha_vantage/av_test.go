package alpha_vantage

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"

	e "mc.data/extensions"
	m "mc.data/models"
)

var errMissingAPIKey = errors.New("ALPHAVANTAGE_API_KEY not set (load .env)")

// only use when needing to test alpha vantage api queries, will burn daily API quota quickly
var long = flag.Bool("long", false, "run long-running tests")

const (
	avKeyName = "ALPHAVANTAGE_API_KEY"
	ticker    = "AAPL"
)

// shared API data: fetched once by getSharedAlphaVantageData and reused by all tests
var (
	sharedOnce        sync.Once
	sharedWeekly      *m.TimeSeriesResult
	sharedWeeklyErr   error
	sharedIntraday    *m.TimeSeriesIntradayResult
	sharedIntradayErr error
)

// getSharedAlphaVantageData fetches weekly and intraday data once per test run and returns the same data to every caller.
// All tests that need Alpha Vantage API data should use this so we make at most 2 API calls total.
func getSharedAlphaVantageData(t *testing.T) (weekly *m.TimeSeriesResult, intraday *m.TimeSeriesIntradayResult) {
	t.Helper()
	sharedOnce.Do(func() {
		apiKey := getApiKey(t)
		if apiKey == "" {
			sharedWeeklyErr = errMissingAPIKey
			return
		}
		c := GetClient(apiKey)
		sharedWeekly, sharedWeeklyErr = c.GetStockWeeklyAdjustedMetrics(ticker)
		time.Sleep(2 * time.Second)
		sharedIntraday, sharedIntradayErr = c.GetStockIntradayMetrics(ticker)
	})
	if sharedWeeklyErr != nil {
		t.Fatalf("error getting weekly data: %s", sharedWeeklyErr)
	}
	if sharedIntradayErr != nil {
		t.Fatalf("error getting intraday data: %s", sharedIntradayErr)
	}
	return sharedWeekly, sharedIntraday
}

func Test_AlphaVantage_GetApiKey(t *testing.T) {
	err := godotenv.Load("../testenv")
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
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Errorf("error loading environment: %s", err)
	}

	return os.Getenv(avKeyName)
}

func Test_AlphaVantage_StockIntradayTimeSeries(t *testing.T) {
	if !*long {
		t.Skip("skipping test that utilizes alpha vantage api queries, set flag -long to run")
	}

	_, res := getSharedAlphaVantageData(t)

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("error parsing time zone: %s", err)
	}

	// meta data
	if res.Metadata.Symbol != ticker {
		t.Fatalf("metadata symbol did not match expected value %s != %s", res.Metadata.Symbol, ticker)
	}

	targetDate := time.Date(2025, time.October, 31, 19, 0, 0, 0, location)
	if targetDate.After(res.Metadata.LastRefreshed) {
		t.Fatalf("error parsing meta data last refreshed date, %s", res.Metadata.LastRefreshed)
	}

	if len(res.TimeSeries) == 0 {
		t.Fatalf("no time series data seems to have been returned")
	}

	// time series element tieout
	// because the window of what is returned form av slides, we can just check if non default values are provided
	s := res.TimeSeries[0]
	jsonData, _ := json.MarshalIndent(s, "", "  ")
	t.Logf("JSON: %s", jsonData)

	if s.Open == 0 {
		t.Fatalf("open price mismatch, expected non zero, got %v", s.Open)
	}
	if s.High == 0 {
		t.Fatalf("high price mismatch, expected non zero, got %v", s.High)
	}
	if s.Low == 0 {
		t.Fatalf("low price mismatch, expected non zero, got %v", s.Low)
	}
	if s.Close == 0 {
		t.Fatalf("close price mismatch, expected non zero, got %v", s.Close)
	}
	if s.Volume == 0 {
		t.Fatalf("volume mismatch, expected non zero, got %v", s.Volume)
	}
}

func Test_AlphaVantage_StockTimeSeries(t *testing.T) {
	if !*long {
		t.Skip("skipping test that utilizes alpha vantage api queries, set flag -long to run")
	}

	res, _ := getSharedAlphaVantageData(t)

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("error parsing time zone: %s", err)
	}

	// meta data
	if res.Metadata.Symbol != ticker {
		t.Fatalf("metadata symbol did not match expected value %s != %s", res.Metadata.Symbol, ticker)
	}

	targetDate := time.Date(2025, time.October, 31, 0, 0, 0, 0, location)
	if targetDate.After(res.Metadata.LastRefreshed) {
		t.Fatalf("error parsing meta data last refreshed date, %s", res.Metadata.LastRefreshed)
	}

	// time series element tieout
	f := func(tsd *m.TimeSeriesData) bool { return targetDate.Equal(tsd.Timestamp) }
	s, err := e.FilterSingle(res.TimeSeries, f)
	if err != nil {
		t.Fatalf("error filtering single time series element: %v", err)
	}
	if s == nil {
		t.Fatalf("error filtering single time series element, resulted in nil")
	}

	jsonData, _ := json.MarshalIndent(s, "", "  ")
	t.Logf("JSON: %s", jsonData)

	expected := m.TimeSeriesData{
		TimeSeriesOHLCV: m.TimeSeriesOHLCV{
			Open:   264.88,
			High:   277.320,
			Low:    264.6501,
			Close:  270.37,
			Volume: 293563310,
		},
		AdjustedClose:  270.1093,
		DividendAmount: 0,
	}

	if s.Open != expected.Open {
		t.Fatalf("open price mismatch, expected %v, got %v", expected.Open, s.Open)
	}

	if s.High != expected.High {
		t.Fatalf("high price mismatch, expected %v, got %v", expected.High, s.High)
	}

	if s.Low != expected.Low {
		t.Fatalf("low price mismatch, expected %v, got %v", expected.Low, s.Low)
	}

	if s.Close != expected.Close {
		t.Fatalf("close price mismatch, expected %v, got %v", expected.Close, s.Close)
	}

	if s.Volume != expected.Volume {
		t.Fatalf("volume mismatch, expected %v, got %v", expected.Volume, s.Volume)
	}

	if s.AdjustedClose != expected.AdjustedClose {
		t.Fatalf("adjusted close price mismatch, expected %v, got %v", expected.AdjustedClose, s.AdjustedClose)
	}

	if s.DividendAmount != expected.DividendAmount {
		t.Fatalf("dividend amount mismatch, expected %v, got %v", expected.DividendAmount, s.DividendAmount)
	}
}
