package repos

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	ex "mc.data/extensions"
	m "mc.data/models"
)

func Test_Base_CanGetConnectionAndPing(t *testing.T) {
	ctx := context.Background()
	pg := getConnection(t, ctx)
	err := pg.Ping(ctx)

	if err != nil {
		t.Errorf("error pinging postgres database: %s", err)
	}
}

func Test_TimeSeriesMetaDataRepo_CanInsertAndGet(t *testing.T) {
	symbol := "_TEST"
	
	testMetaData := m.TimeSeriesMetadata{
		Symbol:        symbol,
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
	}

	ctx := context.Background()
	pg := getConnection(t, ctx)

	exists, err := pg.GetMetaDataBySymbol(ctx, symbol)
	if err != nil {
		t.Fatalf("error determining if meta symbol exists for %s (should be false): %s", symbol, err)
	}
	if exists != nil {
		t.Fatalf("symbol %s has not been inserted yet, so exists should be false", symbol)
	}
	
	if err := pg.InsertNewMetaData(ctx, &testMetaData, nil); err != nil {
		t.Fatalf("error inserting new meta data: %s", err)
	}
	if testMetaData.Id == 0 {
		t.Fatalf("id for test meta data failted to set properly")
	}

	defer pg.deleteTestTimeSeriesData(t, ctx, testMetaData.Id)

	res, err := pg.GetMetaDataBySymbol(ctx, symbol)

	if err != nil {
		t.Fatalf("error getting meta data by symbol, %s", err)
	}
	if testMetaData.Id != res.Id {
		t.Fatalf("ids did not match, inserted %d, got back %d", testMetaData.Id, res.Id)
	}
	if testMetaData.Symbol != res.Symbol {
		t.Fatalf("symbols did not match, inserted %s, got back %s", testMetaData.Symbol, res.Symbol)
	}
	if testMetaData.LastRefreshed != res.LastRefreshed {
		t.Fatalf("last refreshed time did not match, inserted %s, got back %s", testMetaData.LastRefreshed.Format(time.RFC3339), res.LastRefreshed.Format(time.RFC3339))
	}	
}

func Test_TimeSeriesDataRepo_CanInsertAndGet(t *testing.T) {
	symbol := "_TEST2"
	
	testMetaData := m.TimeSeriesMetadata{
		Symbol:        symbol,
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
	}

	ctx := context.Background()
	pg := getConnection(t, ctx)
	
	if err := pg.InsertNewMetaData(ctx, &testMetaData, nil); err != nil {
		t.Fatalf("error inserting new meta data: %s", err)
	}

	defer pg.deleteTestTimeSeriesData(t, ctx, testMetaData.Id)

	testTimeSeriesData := make([]*m.TimeSeriesData, 2)
	testTimeSeriesData[0] = &m.TimeSeriesData{
		SourceId: testMetaData.Id,
		Timestamp: time.Date(2025, time.October, 30, 0, 0, 0, 0, time.UTC),
		TimeSeriesOHLCV: m.TimeSeriesOHLCV{
			Open: 100,
			High: 105,
			Low: 95,
			Close: 102,
			Volume: 1000,
		},
		AdjustedClose: 50,
		DividendAmount: 1,
	}
	testTimeSeriesData[1] = &m.TimeSeriesData{
		SourceId: testMetaData.Id,
		Timestamp: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
		TimeSeriesOHLCV: m.TimeSeriesOHLCV{
			Open: 102,
			High: 107,
			Low: 97,
			Close: 104,
			Volume: 2000,
		},
		AdjustedClose: 51,
		DividendAmount: 2,
	}

	ct, err := pg.InsertTimeSeriesData(ctx, testTimeSeriesData, nil, nil)
	if err != nil {
		t.Fatalf("error inserting time series data: %s", err)
	}
	if ct != int64(len(testTimeSeriesData)) {
		t.Fatalf("expected to insert %d time series data rows, but inserted %d", len(testTimeSeriesData), ct)
	}

	ts, err := pg.GetTimeSeriesData(ctx, symbol)
	if err != nil {
		t.Fatalf("error getting time series data by symbol: %s", err)
	}

	compareTimeSeriesData(t, testTimeSeriesData[1], ts[0])
	compareTimeSeriesData(t, testTimeSeriesData[0], ts[1])
}

func compareTimeSeriesData(t *testing.T, expected, actual *m.TimeSeriesData) {
	t.Helper()
	if expected.Timestamp.Compare(actual.Timestamp) == 1 { // time is before the actual
        t.Fatalf("value mismatch for timestamp, expected %v, got %v", expected.Timestamp.Format(time.RFC3339), actual.Timestamp.Format(time.RFC3339))
	}
	ex.AssertAreEqual(t, "open", expected.Open, actual.Open)
	ex.AssertAreEqual(t, "high", expected.High, actual.High)
	ex.AssertAreEqual(t, "low", expected.Low, actual.Low)
	ex.AssertAreEqual(t, "close", expected.Close, actual.Close)
	ex.AssertAreEqual(t, "volume", expected.Volume, actual.Volume)
	ex.AssertAreEqual(t, "adjusted close", expected.AdjustedClose, actual.AdjustedClose)
	ex.AssertAreEqual(t, "dividend amount", expected.DividendAmount, expected.DividendAmount)
}

func getConnection(t *testing.T, ctx context.Context) *Postgres {
	t.Helper()
	err := godotenv.Load("../.env");
	if err != nil {
		t.Fatalf("error loading environment: %s", err)
	}

	connectionString := os.Getenv("DATABASE_URL")
	res, err := GetPostgresConnection(ctx, connectionString)

	if err != nil {
		t.Fatalf("error getting postgres connection: %s", err)
	}

	t.Cleanup(func() {
		res.Close()
	})

	return res
}

func (pg *Postgres) deleteTestTimeSeriesData(t *testing.T, ctx context.Context, id int32) {
	t.Helper()
	
	args := pgx.NamedArgs{"source_id": id}
	_, err1 := pg.db.Exec(ctx, "DELETE FROM av_time_series_data WHERE source_id = @source_id", args)
	if err1 != nil {
		t.Errorf("cleanup av_time_series_data failed: %s", err1)
	}

	_, err2 := pg.db.Exec(ctx, "DELETE FROM av_time_series_metadata WHERE id = @source_id", args)
	if err2 != nil {
		t.Errorf("cleanup av_time_series_metadata failed: %s", err2)
	}
}