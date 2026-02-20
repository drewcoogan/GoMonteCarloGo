package core

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	ex "mc.data/extensions"
	dm "mc.data/models"
	sm "mc.service/models"
)

func (sc *ServiceContext) SyncSymbolTimeSeriesData(symbol string) (time.Time, error) {
	timeSeriesMetaData, err := sc.PostgresConnection.GetMetaDataBySymbol(sc.Context, symbol)

	if err != nil {
		return time.Time{}, fmt.Errorf("error determining if meta data exists in sync data: %w", err)
	}

	if timeSeriesMetaData == nil {
		log.Printf("adding new symbol to db: %s", symbol)
		timeSeriesMetaData = &dm.TimeSeriesMetadata{
			Symbol:        symbol,
			LastRefreshed: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		if err := sc.PostgresConnection.InsertNewMetaData(sc.Context, timeSeriesMetaData, nil); err != nil {
			return time.Time{}, fmt.Errorf("error adding %s to db: %w", symbol, err)
		}
	}

	cutoffDate := time.Now().AddDate(0, 0, -7)
	if timeSeriesMetaData.LastRefreshed.After(cutoffDate) {
		return timeSeriesMetaData.LastRefreshed, fmt.Errorf("data was refreshed less than a week ago (%s), will not sync symbol %s", ex.FmtShort(timeSeriesMetaData.LastRefreshed), symbol)
	}

	// TODO: is this able to return a null postgres value to a pointer?
	mrd, err := sc.PostgresConnection.GetMostRecentTimestampForSymbol(sc.Context, symbol)
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting most recent time series date for symbol %s: %w", symbol, err)
	}

	tsr, err := sc.AlphaVantageClient.GetStockWeeklyAdjustedMetrics(symbol)
	if err != nil {
		return time.Time{}, err
	}

	f := func(t *dm.TimeSeriesData) bool { return mrd == nil || mrd.After(t.Timestamp) }
	toInsert := ex.FilterMultiplePtr(tsr.TimeSeries, f)

	tx, err := sc.PostgresConnection.GetTransaction(sc.Context)
	if err != nil {
		return time.Time{}, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback(sc.Context) // this will kick off if we return before committing

	var ra int64
	if len(toInsert) > 0 {
		ra, err = sc.PostgresConnection.InsertTimeSeriesData(sc.Context, toInsert, &timeSeriesMetaData.Id, tx)
		if err != nil {
			return time.Time{}, fmt.Errorf("error inserting time series data: %w", err)
		}
	}

	if err := sc.PostgresConnection.UpdateLastRefreshedDate(sc.Context, symbol, tsr.Metadata.LastRefreshed, tx); err != nil {
		return time.Time{}, err
	}

	if err := tx.Commit(sc.Context); err != nil {
		return time.Time{}, fmt.Errorf("error committing transaction to add new symbol %s: %w", symbol, err)
	}

	log.Printf("symbol %s got %v time series elements from av, inserted %v values", symbol, len(tsr.TimeSeries), ra)
	return tsr.Metadata.LastRefreshed, nil
}

func (sc *ServiceContext) InsertNewScenario(request sm.ScenarioRequest) (*dm.Scenario, int, error) {
	if err := validateScenarioRequest(request); err != nil {
		return nil, http.StatusBadRequest, err
	}

	mappedScenario := sm.MapScenarioRequestToDataModel(request)
	created, err := sc.PostgresConnection.InsertNewScenario(sc.Context, mappedScenario)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error creating scenario: %v", err)
	}

	return created, http.StatusCreated, nil
}

func (sc *ServiceContext) UpdateScenario(scenarioId int32, request sm.ScenarioRequest) (*dm.Scenario, int, error) {
	if err := validateScenarioRequest(request); err != nil {
		return nil, http.StatusBadRequest, err
	}

	mappedScenario := sm.MapScenarioRequestToDataModel(request)
	updated, err := sc.PostgresConnection.UpdateExistingScenario(sc.Context, scenarioId, mappedScenario)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error updating scenario: %v", err)
	}

	return updated, http.StatusOK, nil
}

func validateScenarioRequest(req sm.ScenarioRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if len(req.Components) == 0 {
		return fmt.Errorf("at least one component is required")
	}

	seen := make(map[int32]bool, len(req.Components))
	weightSum := 0.0
	for _, component := range req.Components {
		if component.AssetId == 0 {
			return fmt.Errorf("assetId must be provided")
		}
		if component.Weight <= 0 {
			return fmt.Errorf("component weights must be positive")
		}
		if seen[component.AssetId] {
			return fmt.Errorf("duplicate assetId %d", component.AssetId)
		}
		seen[component.AssetId] = true
		weightSum += component.Weight
	}

	if math.Abs(weightSum-1.0) > 0.001 {
		return fmt.Errorf("component weights must sum to 1.0, got %.4f", weightSum)
	}

	return nil
}
