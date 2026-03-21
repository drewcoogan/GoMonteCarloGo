package core

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"os"

	"github.com/joho/godotenv"

	dm "mc.data/models"
	r "mc.data/repos"
	av "mc.service/api/alpha_vantage"
	sm "mc.service/models"
)

func Test_validateScenario(t *testing.T) {
	t.Run("valid weights sum to 1", func(t *testing.T) {
		scenario := &dm.Scenario{
			ScenarioConfiguration: dm.ScenarioConfiguration{Name: "x", FloatedWeight: false},
			Components: []dm.ScenarioConfigurationComponent{
				{AssetId: 1, Weight: 0.6},
				{AssetId: 2, Weight: 0.4},
			},
		}
		if err := validateScenario(scenario); err != nil {
			t.Errorf("expected no error: %v", err)
		}
	})

	t.Run("weights do not sum to 1", func(t *testing.T) {
		scenario := &dm.Scenario{
			ScenarioConfiguration: dm.ScenarioConfiguration{Name: "x", FloatedWeight: false},
			Components: []dm.ScenarioConfigurationComponent{
				{AssetId: 1, Weight: 0.5},
				{AssetId: 2, Weight: 0.4},
			},
		}
		if err := validateScenario(scenario); err == nil {
			t.Error("expected error when weights do not sum to 1")
		}
	})

	t.Run("duplicate asset id", func(t *testing.T) {
		scenario := &dm.Scenario{
			ScenarioConfiguration: dm.ScenarioConfiguration{Name: "x", FloatedWeight: false},
			Components: []dm.ScenarioConfigurationComponent{
				{AssetId: 1, Weight: 0.5},
				{AssetId: 1, Weight: 0.5},
			},
		}
		if err := validateScenario(scenario); err == nil {
			t.Error("expected error for duplicate asset id")
		}
	})
}

// insertScenarioWithTimeSeriesForSimulation creates two metadata rows, time series data (same dates for both),
// and a scenario. Caller must defer: DeleteSimulationRunByID (for any run), DeleteScenario, DeleteMetadataByID x2.
func insertScenarioWithTimeSeriesForSimulation(t *testing.T, sc ServiceContext) (scenarioID int32, assetAID, assetBID int32) {
	t.Helper()
	ctx := sc.Context
	pg := sc.PostgresConnection

	suffix := time.Now().UnixNano()
	assetA := dm.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_SIM_A_%d", suffix),
		LastRefreshed: time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC),
	}
	assetB := dm.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_SIM_B_%d", suffix),
		LastRefreshed: time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC),
	}
	if err := pg.InsertNewMetaData(ctx, &assetA, nil); err != nil {
		t.Fatalf("insert metadata A: %v", err)
	}
	if err := pg.InsertNewMetaData(ctx, &assetB, nil); err != nil {
		t.Fatalf("insert metadata B: %v", err)
	}
	assetAID, assetBID = assetA.Id, assetB.Id

	// Same dates for both assets so GetTimeSeriesReturns returns aligned data (verifySeriesReturnIntegrity).
	// Use enough points (e.g. 20 weeks) so the covariance matrix is positive definite.
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var tsData []*dm.TimeSeriesData
	for i := range 21 {
		ts := baseDate.AddDate(0, 0, 7*i)
		adjA := 100.0 * (1.0 + 0.001*float64(i) + 0.01*float64(i%3))
		adjB := 50.0 * (1.0 + 0.002*float64(i) - 0.005*float64(i%2))
		tsData = append(tsData,
			&dm.TimeSeriesData{SourceId: assetA.Id, Timestamp: ts, TimeSeriesOHLCV: dm.TimeSeriesOHLCV{Open: adjA - 0.5, High: adjA + 0.5, Low: adjA - 0.5, Close: adjA, Volume: 1000}, AdjustedClose: adjA, DividendAmount: 0},
			&dm.TimeSeriesData{SourceId: assetB.Id, Timestamp: ts, TimeSeriesOHLCV: dm.TimeSeriesOHLCV{Open: adjB - 0.25, High: adjB + 0.25, Low: adjB - 0.25, Close: adjB, Volume: 1000}, AdjustedClose: adjB, DividendAmount: 0},
		)
	}
	if _, err := pg.InsertTimeSeriesData(ctx, tsData, nil, nil); err != nil {
		t.Fatalf("insert time series data: %v", err)
	}

	scenario := dm.Scenario{
		ScenarioConfiguration: dm.ScenarioConfiguration{
			Name:          fmt.Sprintf("Sim Test %d", suffix),
			FloatedWeight: false,
		},
		Components: []dm.ScenarioConfigurationComponent{
			{ConfigurationId: 0, AssetId: assetA.Id, Weight: 0.6},
			{ConfigurationId: 0, AssetId: assetB.Id, Weight: 0.4},
		},
	}
	created, err := pg.InsertNewScenario(ctx, scenario)
	if err != nil {
		t.Fatalf("insert scenario: %v", err)
	}
	return created.Id, assetAID, assetBID
}

func Test_RunSimulation_fullFlow(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("loading .env: %v", err)
	}
	ctx := context.Background()
	pg, err := r.GetPostgresConnection(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("postgres connection: %v", err)
	}
	t.Cleanup(func() { pg.Close() })

	sc := ServiceContext{
		Context:            ctx,
		PostgresConnection: pg,
		AlphaVantageClient: av.AlphaVantageClient{},
	}

	scenarioID, assetAID, assetBID := insertScenarioWithTimeSeriesForSimulation(t, sc)
	defer func() {
		_ = pg.DeleteScenarioPermanent(ctx, scenarioID)
		_ = pg.DeleteMetadataByID(ctx, assetAID)
		_ = pg.DeleteMetadataByID(ctx, assetBID)
	}()

	// MaxLookback so that maxLookbackDate is before our time series (2024-06-01).
	settings := sm.SimulationRequestSettings{
		DistributionType:     sm.StandardNormal,
		SimulationUnitOfTime: sm.Weekly,
		SimulationDuration:   52,
		MaxLookback:          2 * 365 * 24 * time.Hour,
		Iterations:           100,
		Seed:                 42,
		DegreesOfFreedom:     10,
	}

	response, err := sc.RunSimulation(scenarioID, settings)
	if err != nil {
		t.Fatalf("RunSimulation: %v", err)
	}
	if response == nil {
		t.Fatal("expected non-nil response")
	}

	// Run creates a simulation run; we need to clean it up. Get run history to get run id.
	runs, err := pg.GetSimulationRunHistories(ctx, scenarioID, 10)
	if err != nil {
		t.Fatalf("get run histories: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	defer func() { _ = pg.DeleteSimulationRunByID(ctx, runs[0].Id) }()

	if math.IsNaN(response.RiskMetrics.MeanFinalValue) || math.IsInf(response.RiskMetrics.MeanFinalValue, 0) {
		t.Errorf("MeanFinalValue not finite: %f", response.RiskMetrics.MeanFinalValue)
	}
	if len(response.SamplePaths) == 0 {
		t.Error("expected non-empty SamplePaths")
	}
	sampleRoleCount := 0
	for _, p := range response.SamplePaths {
		if p.Role == dm.PathRoleSample {
			sampleRoleCount++
		}
	}
	if sampleRoleCount == 0 {
		t.Error("expected at least one path with role sample")
	}
	if sampleRoleCount > randomScenarioPathsCount {
		t.Errorf("sample-role paths %d > cap %d", sampleRoleCount, randomScenarioPathsCount)
	}
	if len(response.Summary.Mean) == 0 {
		t.Error("expected non-empty Summary.Mean")
	}

	// Run should be marked success (no error message).
	if runs[0].ErrorMessage != nil && *runs[0].ErrorMessage != "" {
		t.Errorf("run should be success, error_message=%q", *runs[0].ErrorMessage)
	}

	// Result should be persisted.
	gotResult, err := pg.GetSimulationResult(ctx, runs[0].Id)
	if err != nil {
		t.Fatalf("get simulation result: %v", err)
	}
	if gotResult == nil {
		t.Fatal("expected persisted result")
	}
	if gotResult.RiskMetrics.MeanFinalValue != response.RiskMetrics.MeanFinalValue {
		t.Errorf("persisted MeanFinalValue %f != response %f", gotResult.RiskMetrics.MeanFinalValue, response.RiskMetrics.MeanFinalValue)
	}
}
