package repos

import (
	"context"
	"fmt"
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

func Test_TimeSeriesMetaDataRepo_CanCRUD(t *testing.T) {
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
		t.Fatalf("last refreshed time did not match, inserted %s, got back %s", ex.FmtLong(testMetaData.LastRefreshed), ex.FmtLong(res.LastRefreshed))
	}

	newLR := time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC)

	if err := pg.UpdateLastRefreshedDate(ctx, symbol, newLR, nil); err != nil {
		t.Fatalf("error updating last refreshed date: %s", err)
	}

	newRes, err := pg.GetMetaDataBySymbol(ctx, symbol)
	if err != nil {
		t.Fatalf("error getting updated meta data for symbol %s", symbol)
	}
	if newLR != newRes.LastRefreshed {
		t.Fatalf("error updating meta data last refreshed date, expected %s, got %s", ex.FmtLong(newLR), ex.FmtLong(newRes.LastRefreshed))
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
		SourceId:  testMetaData.Id,
		Timestamp: time.Date(2025, time.October, 30, 0, 0, 0, 0, time.UTC),
		TimeSeriesOHLCV: m.TimeSeriesOHLCV{
			Open:   100,
			High:   105,
			Low:    95,
			Close:  102,
			Volume: 1000,
		},
		AdjustedClose:  50,
		DividendAmount: 1,
	}
	testTimeSeriesData[1] = &m.TimeSeriesData{
		SourceId:  testMetaData.Id,
		Timestamp: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
		TimeSeriesOHLCV: m.TimeSeriesOHLCV{
			Open:   102,
			High:   107,
			Low:    97,
			Close:  104,
			Volume: 2000,
		},
		AdjustedClose:  51,
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

func Test_ScenarioRepo_CanCRUD(t *testing.T) {
	ctx := context.Background()
	pg := getConnection(t, ctx)

	suffix := time.Now().UnixNano()
	assetA := m.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_TEST_SCN_A_%d", suffix),
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
	}
	assetB := m.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_TEST_SCN_B_%d", suffix),
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
	}

	if err := pg.InsertNewMetaData(ctx, &assetA, nil); err != nil {
		t.Fatalf("error inserting metadata A: %s", err)
	}
	if err := pg.InsertNewMetaData(ctx, &assetB, nil); err != nil {
		t.Fatalf("error inserting metadata B: %s", err)
	}

	defer pg.deleteTestTimeSeriesData(t, ctx, assetA.Id)
	defer pg.deleteTestTimeSeriesData(t, ctx, assetB.Id)

	scenarioName := fmt.Sprintf("Test Scenario %d", suffix)
	newScenario := m.NewScenario{
		Name:          scenarioName,
		FloatedWeight: false,
		Components: []m.NewComponent{
			{AssetId: assetA.Id, Weight: 0.6},
			{AssetId: assetB.Id, Weight: 0.4},
		},
	}

	created, err := pg.InsertNewScenario(ctx, newScenario, nil)
	if err != nil {
		t.Fatalf("error inserting scenario: %s", err)
	}
	if created.Id == 0 {
		t.Fatalf("scenario id was not set")
	}

	defer pg.deleteTestScenarioData(t, ctx, created.Id)

	fetched, err := pg.GetScenarioByID(ctx, created.Id)
	if err != nil {
		t.Fatalf("error fetching scenario: %s", err)
	}
	if fetched == nil {
		t.Fatalf("expected scenario to be returned")
	}
	if fetched.Name != scenarioName {
		t.Fatalf("scenario name mismatch, expected %s, got %s", scenarioName, fetched.Name)
	}
	if len(fetched.Components) != len(newScenario.Components) {
		t.Fatalf("expected %d components, got %d", len(newScenario.Components), len(fetched.Components))
	}

	componentLookup := make(map[int32]m.ScenarioConfigurationComponent)
	for _, c := range fetched.Components {
		componentLookup[c.AssetId] = c
	}
	if componentLookup[assetA.Id].Weight != 0.6 {
		t.Fatalf("component weight mismatch for asset A, expected 0.6, got %.2f", componentLookup[assetA.Id].Weight)
	}
	if componentLookup[assetB.Id].Weight != 0.4 {
		t.Fatalf("component weight mismatch for asset B, expected 0.4, got %.2f", componentLookup[assetB.Id].Weight)
	}

	updatedScenario := m.NewScenario{
		Name:          scenarioName + "_UPDATED",
		FloatedWeight: true,
		Components: []m.NewComponent{
			{AssetId: assetA.Id, Weight: 0.7},
			{AssetId: assetB.Id, Weight: 0.3},
		},
	}

	updated, err := pg.UpdateExistingScenario(ctx, created.Id, updatedScenario, nil)
	if err != nil {
		t.Fatalf("error updating scenario: %s", err)
	}
	if updated.Name != updatedScenario.Name {
		t.Fatalf("scenario name mismatch after update, expected %s, got %s", updatedScenario.Name, updated.Name)
	}
	if len(updated.Components) != len(updatedScenario.Components) {
		t.Fatalf("expected %d components after update, got %d", len(updatedScenario.Components), len(updated.Components))
	}

	if err := pg.DeleteScenario(ctx, created.Id, nil); err != nil {
		t.Fatalf("error deleting scenario: %s", err)
	}
	afterDelete, err := pg.GetScenarioByID(ctx, created.Id)
	if err != nil {
		t.Fatalf("error fetching deleted scenario: %s", err)
	}
	if afterDelete != nil {
		t.Fatalf("expected scenario to be deleted")
	}
}

func compareTimeSeriesData(t *testing.T, expected, actual *m.TimeSeriesData) {
	t.Helper()
	if expected.Timestamp.Before(actual.Timestamp) {
		t.Fatalf("value mismatch for timestamp, expected %v, got %v", expected.Timestamp.Format(time.RFC3339), actual.Timestamp.Format(time.RFC3339))
	}
	ex.AssertAreEqual(t, "open", expected.Open, actual.Open)
	ex.AssertAreEqual(t, "high", expected.High, actual.High)
	ex.AssertAreEqual(t, "low", expected.Low, actual.Low)
	ex.AssertAreEqual(t, "close", expected.Close, actual.Close)
	ex.AssertAreEqual(t, "volume", expected.Volume, actual.Volume)
	ex.AssertAreEqual(t, "adjusted close", expected.AdjustedClose, actual.AdjustedClose)
	ex.AssertAreEqual(t, "dividend amount", expected.DividendAmount, actual.DividendAmount)
}

func getConnection(t *testing.T, ctx context.Context) Postgres {
	t.Helper()
	err := godotenv.Load("../.env")
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

func (pg *Postgres) deleteTestScenarioData(t *testing.T, ctx context.Context, id int32) {
	t.Helper()

	args := pgx.NamedArgs{"scenario_id": id}
	_, err1 := pg.db.Exec(ctx, "DELETE FROM scenario_configuration_component WHERE configuration_id = @scenario_id", args)
	if err1 != nil {
		t.Errorf("cleanup scenario_configuration_component failed: %s", err1)
	}

	_, err2 := pg.db.Exec(ctx, "DELETE FROM scenario_configuration WHERE id = @scenario_id", args)
	if err2 != nil {
		t.Errorf("cleanup scenario_configuration failed: %s", err2)
	}
}
