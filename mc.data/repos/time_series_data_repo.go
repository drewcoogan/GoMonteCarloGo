package repos

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	m "mc.data/models"
)

func (pg *Postgres) GetTimeSeriesData(ctx context.Context, symbol string) ([]*m.TimeSeriesData, error) {
	query := `
		SELECT 
			atsd.source_id,
			atsd."timestamp", 
			atsd."open", 
			atsd.high, 
			atsd.low, 
			atsd."close", 
			atsd.volume, 
			atsd.adjusted_close, 
			atsd.dividend_amount
		FROM av_time_series_data atsd 
		JOIN av_time_series_metadata atsm ON atsd.source_id = atsm.id
		WHERE atsm.symbol = @symbol
		ORDER BY atsd."timestamp" DESC`

	args := pgx.NamedArgs{
		"symbol": symbol,
	}

	res, err := Query[m.TimeSeriesData](ctx, pg, query, args)
	if err != nil {
		return nil, fmt.Errorf("unable to query data by symbol (%s): %w", symbol, err)
	}
	return res, nil
}

func (pg *Postgres) InsertTimeSeriesData(ctx context.Context, data []*m.TimeSeriesData, id *int32, tx *pgx.Tx) (int64, error) {
	columns := []string{
		"source_id", "timestamp", "open", "high", "low",
		"close", "volume", "adjusted_close", "dividend_amount",
	}

	// multiply by -1 to sort the data in descending order
	slices.SortFunc(data, func(i, j *m.TimeSeriesData) int {
		return -1 * i.Timestamp.Compare(j.Timestamp)
	})

	// all time stamps here are date only, so no need to worry about UTC
	entries := make([][]any, len(data))
	for i, ent := range data {
		sourceId := ent.SourceId
		if id != nil {
			sourceId = int32(*id)
		}
		entries[i] = []any{
			sourceId, ent.Timestamp, ent.Open, ent.High, ent.Low,
			ent.Close, ent.Volume, ent.AdjustedClose, ent.DividendAmount,
		}
	}

	if tx == nil {
		return pg.db.CopyFrom(ctx, pgx.Identifier{"av_time_series_data"}, columns, pgx.CopyFromRows(entries))
	}

	return (*tx).CopyFrom(ctx, pgx.Identifier{"av_time_series_data"}, columns, pgx.CopyFromRows(entries))
}

func (pg *Postgres) GetMostRecentTimestampForSymbol(ctx context.Context, symbol string) (*time.Time, error) {
	query := `
		SELECT 
			MAX(atsd.timestamp)
		FROM av_time_series_data atsd 
		JOIN av_time_series_metadata atsm ON atsd.source_id = atsm.id
		WHERE atsm.symbol = @symbol`

	args := pgx.NamedArgs{
		"symbol": symbol,
	}

	ts := new(time.Time)
	if err := pg.db.QueryRow(ctx, query, args).Scan(&ts); err != nil {
		return nil, fmt.Errorf("error getting most recent timestamp for symbol %s: %w", symbol, err)
	}

	return ts, nil
}

func (pg *Postgres) GetTimeSeriesReturns(ctx context.Context, sourceIds []int32, maxLookback time.Duration) ([]*m.TimeSeriesReturn, error) {
	query :=
		`
		WITH price_data AS (
		SELECT 
			t.source_id,
			t.timestamp,
			t.adjusted_close,
			LAG(t.adjusted_close) OVER (PARTITION BY t.source_id ORDER BY t.timestamp) AS prev_close
		FROM av_time_series_data t
		INNER JOIN valid_timestamps v ON t.timestamp = v.timestamp
		WHERE t.source_id = ANY(@source_ids)
			AND t.timestamp >= @max_lookback
		)
		SELECT 
			source_id,
			timestamp,
			LN(adjusted_close / prev_close) AS log_return
		FROM price_data
		WHERE prev_close IS NOT NULL
		ORDER BY source_id, timestamp DESC
	`

	args := pgx.NamedArgs{
		"source_ids":   sourceIds,
		"max_lookback": time.Now().Add(-maxLookback),
	}

	res, err := Query[m.TimeSeriesReturn](ctx, pg, query, args)
	if err != nil {
		return nil, fmt.Errorf("unable to query data by source id (%v): %w", sourceIds, err)
	}

	return res, nil
}
