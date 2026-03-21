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
	if err := pg.Ping(ctx); err != nil {
		t.Fatalf("error pinging postgres database: %s", err)
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

	compareTimeSeriesData(t, testTimeSeriesData[0], ts[0])
	compareTimeSeriesData(t, testTimeSeriesData[1], ts[1])
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
	newScenario := m.Scenario{
		ScenarioConfiguration: m.ScenarioConfiguration{
			Name:          scenarioName,
			FloatedWeight: false,
		},
		Components: []m.ScenarioConfigurationComponent{
			{ConfigurationId: 0, AssetId: assetA.Id, Weight: 0.6},
			{ConfigurationId: 0, AssetId: assetB.Id, Weight: 0.4},
		},
	}

	created, err := pg.InsertNewScenario(ctx, newScenario)
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

	updatedScenario := m.Scenario{
		ScenarioConfiguration: m.ScenarioConfiguration{
			Name:          scenarioName + "_UPDATED",
			FloatedWeight: true,
		},
		Components: []m.ScenarioConfigurationComponent{
			{ConfigurationId: 0, AssetId: assetA.Id, Weight: 0.7},
			{ConfigurationId: 0, AssetId: assetB.Id, Weight: 0.3},
		},
	}

	updated, err := pg.UpdateExistingScenario(ctx, created.Id, updatedScenario)
	if err != nil {
		t.Fatalf("error updating scenario: %s", err)
	}
	if updated.Name != updatedScenario.Name {
		t.Fatalf("scenario name mismatch after update, expected %s, got %s", updatedScenario.Name, updated.Name)
	}
	if len(updated.Components) != len(updatedScenario.Components) {
		t.Fatalf("expected %d components after update, got %d", len(updatedScenario.Components), len(updated.Components))
	}

	if err := pg.DeleteScenario(ctx, created.Id); err != nil {
		t.Fatalf("error deleting scenario: %s", err)
	}

	afterDelete, err := pg.GetScenarioByID(ctx, created.Id)
	if err == nil {
		t.Fatalf("expected error fetching deleted scenario, but did not get an error")
	}
	if afterDelete != nil {
		t.Fatalf("expected scenario to be deleted")
	}
}

func Test_SimulationRunHistoryRepo_CanInsertAndGet(t *testing.T) {
	ctx := context.Background()
	pg := getConnection(t, ctx)

	suffix := time.Now().UnixNano()
	assetA := m.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_TEST_RUN_A_%d", suffix),
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
	}
	assetB := m.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_TEST_RUN_B_%d", suffix),
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

	scenarioName := fmt.Sprintf("Test Run Scenario %d", suffix)
	newScenario := m.Scenario{
		ScenarioConfiguration: m.ScenarioConfiguration{
			Name:          scenarioName,
			FloatedWeight: false,
		},
		Components: []m.ScenarioConfigurationComponent{
			{ConfigurationId: 0, AssetId: assetA.Id, Weight: 0.6},
			{ConfigurationId: 0, AssetId: assetB.Id, Weight: 0.4},
		},
	}
	createdScenario, err := pg.InsertNewScenario(ctx, newScenario)
	if err != nil {
		t.Fatalf("error inserting scenario: %s", err)
	}
	defer pg.deleteTestScenarioData(t, ctx, createdScenario.Id)

	runHistory := m.SimulationRunHistory{
		DistributionType:     "standardNormal",
		SimulationUnitOfTime: "weekly",
		SimulationDuration:   52,
		MaxLookback:          time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		Iterations:           1000,
		Seed:                 42,
		DegreesOfFreedom:     10,
	}
	runId, err := pg.InsertSimulationRunHistory(ctx, createdScenario.Id, runHistory)
	if err != nil {
		t.Fatalf("error inserting simulation run history: %s", err)
	}
	if runId == 0 {
		t.Fatalf("run id was not set")
	}
	defer pg.deleteTestSimulationRunHistory(t, ctx, runId)

	runs, err := pg.GetSimulationRunHistories(ctx, createdScenario.Id, 10)
	if err != nil {
		t.Fatalf("error getting simulation run histories: %s", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	r := runs[0]
	if r.Id != runId {
		t.Fatalf("run id mismatch: expected %d, got %d", runId, r.Id)
	}
	if r.ScenarioId != createdScenario.Id {
		t.Fatalf("scenario id mismatch: expected %d, got %d", createdScenario.Id, r.ScenarioId)
	}
	if r.Name != scenarioName {
		t.Fatalf("run name mismatch: expected %s, got %s", scenarioName, r.Name)
	}
	if r.DistributionType != runHistory.DistributionType {
		t.Fatalf("distribution_type mismatch: expected %s, got %s", runHistory.DistributionType, r.DistributionType)
	}
	if r.Iterations != runHistory.Iterations {
		t.Fatalf("iterations mismatch: expected %d, got %d", runHistory.Iterations, r.Iterations)
	}
	if len(r.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(r.Components))
	}
}

func Test_SimulationRunHistoryRepo_UpdateAsFailureAndSuccess(t *testing.T) {
	ctx := context.Background()
	pg := getConnection(t, ctx)

	suffix := time.Now().UnixNano()
	assetA := m.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_TEST_FAIL_A_%d", suffix),
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
	}
	if err := pg.InsertNewMetaData(ctx, &assetA, nil); err != nil {
		t.Fatalf("error inserting metadata: %s", err)
	}
	defer pg.deleteTestTimeSeriesData(t, ctx, assetA.Id)

	newScenario := m.Scenario{
		ScenarioConfiguration: m.ScenarioConfiguration{
			Name:          fmt.Sprintf("Test Fail Scenario %d", suffix),
			FloatedWeight: false,
		},
		Components: []m.ScenarioConfigurationComponent{
			{ConfigurationId: 0, AssetId: assetA.Id, Weight: 1.0},
		},
	}
	createdScenario, err := pg.InsertNewScenario(ctx, newScenario)
	if err != nil {
		t.Fatalf("error inserting scenario: %s", err)
	}
	defer pg.deleteTestScenarioData(t, ctx, createdScenario.Id)

	runHistory := m.SimulationRunHistory{
		DistributionType:     "standardNormal",
		SimulationUnitOfTime: "weekly",
		SimulationDuration:   52,
		MaxLookback:          time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		Iterations:           100,
		Seed:                 1,
		DegreesOfFreedom:     10,
	}
	runId, err := pg.InsertSimulationRunHistory(ctx, createdScenario.Id, runHistory)
	if err != nil {
		t.Fatalf("error inserting simulation run history: %s", err)
	}
	defer pg.deleteTestSimulationRunHistory(t, ctx, runId)

	if err := pg.UpdateSimulationRunAsFailure(ctx, runId, " test error message "); err != nil {
		t.Fatalf("error updating run as failure: %s", err)
	}
	runs, err := pg.GetSimulationRunHistories(ctx, createdScenario.Id, 10)
	if err != nil {
		t.Fatalf("error getting run histories: %s", err)
	}
	if len(runs) != 1 || runs[0].ErrorMessage == nil || *runs[0].ErrorMessage != "test error message" {
		msg := ""
		if len(runs) > 0 && runs[0].ErrorMessage != nil {
			msg = *runs[0].ErrorMessage
		}
		t.Fatalf("expected error_message to be set after failure; got %q", msg)
	}

	if err := pg.UpdateSimulationRunAsSuccess(ctx, runId); err != nil {
		t.Fatalf("error updating run as success: %s", err)
	}
	runs, err = pg.GetSimulationRunHistories(ctx, createdScenario.Id, 10)
	if err != nil {
		t.Fatalf("error getting run histories after success: %s", err)
	}
	if len(runs) != 1 || (runs[0].ErrorMessage != nil && *runs[0].ErrorMessage != "") {
		msg := ""
		if len(runs) > 0 && runs[0].ErrorMessage != nil {
			msg = *runs[0].ErrorMessage
		}
		t.Fatalf("expected error_message to be cleared after success; got %q", msg)
	}
}

func Test_SimulationResultRepo_CanInsertAndGet(t *testing.T) {
	ctx := context.Background()
	pg := getConnection(t, ctx)

	suffix := time.Now().UnixNano()
	assetA := m.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_TEST_RES_A_%d", suffix),
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
	}
	if err := pg.InsertNewMetaData(ctx, &assetA, nil); err != nil {
		t.Fatalf("error inserting metadata: %s", err)
	}
	defer pg.deleteTestTimeSeriesData(t, ctx, assetA.Id)

	newScenario := m.Scenario{
		ScenarioConfiguration: m.ScenarioConfiguration{
			Name:          fmt.Sprintf("Test Result Scenario %d", suffix),
			FloatedWeight: false,
		},
		Components: []m.ScenarioConfigurationComponent{
			{ConfigurationId: 0, AssetId: assetA.Id, Weight: 1.0},
		},
	}
	createdScenario, err := pg.InsertNewScenario(ctx, newScenario)
	if err != nil {
		t.Fatalf("error inserting scenario: %s", err)
	}
	defer pg.deleteTestScenarioData(t, ctx, createdScenario.Id)

	runHistory := m.SimulationRunHistory{
		DistributionType:     "standardNormal",
		SimulationUnitOfTime: "weekly",
		SimulationDuration:   52,
		MaxLookback:          time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		Iterations:           100,
		Seed:                 2,
		DegreesOfFreedom:     10,
	}
	runId, err := pg.InsertSimulationRunHistory(ctx, createdScenario.Id, runHistory)
	if err != nil {
		t.Fatalf("error inserting simulation run history: %s", err)
	}
	defer pg.deleteTestSimulationRunHistory(t, ctx, runId)

	simResponse := &m.SimulationResponse{
		RiskMetrics: m.SimulationRiskMetrics{
			VaR95:             0.01,
			VaR99:             0.02,
			CVaR95:            0.015,
			CVaR99:            0.025,
			ProbabilityOfLoss: 0.1,
			MaxDrawdownP95:    0.05,
			MeanFinalValue:    1.1,
			MedianFinalValue:  1.05,
		},
		SamplePaths: []m.SamplePath{
			{Role: m.PathRolePercentile, Percentile: 0.5, Values: []float64{1.0, 1.05, 1.1}, Label: "median"},
			{Role: m.PathRoleSample, Percentile: -1, Label: "Sample 1", Values: []float64{1.0, 1.02}},
		},
		Summary: m.SimulationStats{
			Mean:   []float64{1.0, 1.05},
			StdDev: []float64{0.01, 0.02},
			P5:     []float64{0.98, 1.0},
			P95:    []float64{1.02, 1.1},
		},
	}
	if err := pg.InsertSimulationResult(ctx, runId, simResponse); err != nil {
		t.Fatalf("error inserting simulation result: %s", err)
	}

	got, err := pg.GetSimulationResult(ctx, runId)
	if err != nil {
		t.Fatalf("error getting simulation result: %s", err)
	}
	if got == nil {
		t.Fatal("expected non-nil simulation result")
	}
	if got.RiskMetrics.VaR95 != simResponse.RiskMetrics.VaR95 {
		t.Fatalf("VaR95 mismatch: expected %f, got %f", simResponse.RiskMetrics.VaR95, got.RiskMetrics.VaR95)
	}
	if len(got.SamplePaths) != len(simResponse.SamplePaths) {
		t.Fatalf("samplePaths length mismatch: expected %d, got %d", len(simResponse.SamplePaths), len(got.SamplePaths))
	}
	if len(got.Summary.Mean) != len(simResponse.Summary.Mean) {
		t.Fatalf("summary mean length mismatch: expected %d, got %d", len(simResponse.Summary.Mean), len(got.Summary.Mean))
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

func getConnection(t *testing.T, ctx context.Context) *Postgres {
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
	// postgres cascade will delete the data in av_time_series_metadata if the key in metadata is deleted
	_, err := pg.db.Exec(ctx, "DELETE FROM av_time_series_metadata WHERE id = @id", pgx.NamedArgs{"id": id})
	if err != nil {
		t.Errorf("cleanup av_time_series_metadata failed: %s", err)
	}
}

func (pg *Postgres) deleteTestScenarioData(t *testing.T, ctx context.Context, id int32) {
	t.Helper()
	// postgres cascade will delete the data in scenario_configuration_component if the key in scenario is deleted
	_, err := pg.db.Exec(ctx, "DELETE FROM scenario_configuration WHERE id = @id", pgx.NamedArgs{"id": id})
	if err != nil {
		t.Errorf("cleanup scenario_configuration failed: %s", err)
	}
}

// deleteTestSimulationRunHistory removes a run by id. Cascade deletes simulation_run_history_component and simulation_result.
func (pg *Postgres) deleteTestSimulationRunHistory(t *testing.T, ctx context.Context, runId int32) {
	t.Helper()
	_, err := pg.db.Exec(ctx, "DELETE FROM simulation_run_history WHERE id = @id", pgx.NamedArgs{"id": runId})
	if err != nil {
		t.Errorf("cleanup simulation_run_history failed: %s", err)
	}
}
