package alpha_vantage

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/guregu/null/v6"
	"github.com/joho/godotenv"

	u "mc.data"
	ex "mc.data/extensions"
	m "mc.data/models"
)

const (
	avKeyName = "ALPHAVANTAGE_API_KEY"
)

func Test_DoesNull64WorkHowIThink(t *testing.T) {
	var nullInt null.Int64

	if nullInt.Valid {
		t.Fatalf("expected .valid to be false")
	}

	validInt := null.NewInt(64, true)

	if !validInt.Valid {
		t.Fatalf("value is set, expected .value to be true now")
	}

	ex.AssertAreEqual(t, "value", 64, validInt.Ptr())
}

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
	ex.AssertAreEqual(t, "information", metaInfoEx, res.Metadata.Information.Ptr())
	ex.AssertAreEqual(t, "symbol", ticker, &res.Metadata.Symbol)

	targetDate := time.Date(2025, time.October, 31, 19, 0, 0, 0, location)
	if targetDate.Compare(res.Metadata.LastRefreshed) == 1 { // time is before the actual
		t.Fatalf("error parsing meta data last refreshed date, %s", res.Metadata.LastRefreshed)
	}

	ex.AssertAreEqual(t, "interval", "60min", res.Metadata.Interval.Ptr())
	ex.AssertAreEqual(t, "output size", defaultOutputSize, res.Metadata.OutputSize.Ptr())
	ex.AssertAreEqual(t, "time zone", "US/Eastern", &res.Metadata.TimeZone)

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

	ex.AssertAreEqual(t, "open", 270.10, s.Open.Ptr())
	ex.AssertAreEqual(t, "high", 270.14, s.High.Ptr())
	ex.AssertAreEqual(t, "low", 269.90, s.Low.Ptr())
	ex.AssertAreEqual(t, "close", 269.9995, s.Close.Ptr())
	ex.AssertNillability(t, "adjusted close", true, s.AdjustedClose.Ptr())
	ex.AssertAreEqual(t, "volume", float64(36838), s.Volume.Ptr())
	ex.AssertNillability(t, "dividend amount", true, s.DividendAmount.Ptr())

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
	ex.AssertAreEqual(t, "information", metaInfoEx, res.Metadata.Information.Ptr())
	ex.AssertAreEqual(t, "symbol", ticker, &res.Metadata.Symbol)

	targetDate := time.Date(2025, time.October, 31, 0, 0, 0, 0, location)
	if targetDate.Compare(res.Metadata.LastRefreshed) == 1 { // time is before the actual
		t.Fatalf("error parsing meta data last refreshed date, %s", res.Metadata.LastRefreshed)
	}

	ex.AssertNillability(t, "interval", true, res.Metadata.Interval.Ptr())
	ex.AssertNillability(t, "output size", true, res.Metadata.OutputSize.Ptr())
	ex.AssertAreEqual(t, "time zone", "US/Eastern", &res.Metadata.TimeZone)

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

	ex.AssertAreEqual(t, "open", 264.88, s.Open.Ptr())
	ex.AssertAreEqual(t, "high", 277.320, s.High.Ptr())
	ex.AssertAreEqual(t, "low", 264.6501, s.Low.Ptr())
	ex.AssertAreEqual(t, "close", 270.37, s.Close.Ptr())
	ex.AssertAreEqual(t, "adjusted close", 270.37, s.AdjustedClose.Ptr())
	ex.AssertAreEqual(t, "volume", float64(293563310), s.Volume.Ptr())
	ex.AssertAreEqual(t, "dividend amount", 0, s.DividendAmount.Ptr())
}