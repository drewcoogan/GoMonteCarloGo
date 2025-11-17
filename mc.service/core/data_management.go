package core

import (
	"fmt"
	"log"
	"time"

	ex "mc.data/extensions"
	m "mc.data/models"
)

func (sc *ServiceContext) SyncData(symbol string) error {
	md, err := sc.PostgresConnection.GetMetaDataBySymbol(sc.Context, symbol) 

	if err != nil {
		return fmt.Errorf("error determining if meta data exists in sync data: %w", err)
	}

	if md == nil {
		log.Printf("adding new symbol to db: %s", symbol)
		md = &m.TimeSeriesMetadata{
			Symbol: symbol,
			LastRefreshed: time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		if err := sc.PostgresConnection.InsertNewMetaData(sc.Context, md, nil); err != nil {
			return fmt.Errorf("error adding %s to db: %w", symbol, err)
		}
	}

	cutoffDate := time.Now().AddDate(0, 0, -7)
	if md.LastRefreshed.After(cutoffDate) {
		return fmt.Errorf("data was refreshed less than a week ago (%s), will not sync symbol %s", ex.Fmt(md.LastRefreshed), symbol)
	}

	mrd, err := sc.PostgresConnection.GetMostRecentTimestampForSymbol(sc.Context, symbol)
	if err != nil {
		return fmt.Errorf("error getting most recent time series date for symbol %s: %w", symbol, err)
	}

	tsr, err := sc.AlphaVantageClient.GetStockWeeklyAdjustedMetrics(symbol)
	if err != nil {
		return err
	}

	f := func(t *m.TimeSeriesData) bool { return mrd.After(t.Timestamp) }
	toInsert := ex.FilterMultiplePtr(tsr.TimeSeries, f)

    tx, err := sc.PostgresConnection.GetTransaction(sc.Context)
    if err != nil {
        return fmt.Errorf("error beginning transaction: %w", err)
    }
    defer tx.Rollback(sc.Context) // this will kick off if we return before committing

	var ra int64
	if len(toInsert) > 0 {
		ra, err = sc.PostgresConnection.InsertTimeSeriesData(sc.Context, toInsert, &md.Id, &tx)
		if err != nil {
			return fmt.Errorf("error inserting time series data: %w", err)
		}
	}

	if err := sc.PostgresConnection.UpdateLastRefreshedDate(sc.Context, symbol, tsr.Metadata.LastRefreshed, &tx); err != nil {
		return err
	}

	if err := tx.Commit(sc.Context); err != nil {
		return fmt.Errorf("error committing transaction to add new symbol %s: %w", symbol, err)
	}

	log.Printf("symbol %s got %v time series elements from av, inserted %v values", symbol, len(tsr.TimeSeries), ra)
	return nil
}