package repos

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	m "mc.data/models"
	q "mc.data/queries"
)

func (pg *Postgres) GetTimeSeriesData(ctx context.Context, symbol string) ([]*m.TimeSeriesData, error) {
	sql := q.Get(q.QueryHelper.Select.TimeSeriesData)
	args := pgx.NamedArgs{"symbol": symbol}

	if res, err := Query[m.TimeSeriesData](ctx, pg, sql, args); err != nil {
		return nil, fmt.Errorf("unable to query data by symbol (%s): %w", symbol, err)
	} else {
		return res, nil
	}
}

func (pg *Postgres) InsertTimeSeriesData(ctx context.Context, data []*m.TimeSeriesData, id *int32, tx pgx.Tx) (int64, error) {
	columns := []string{
		"source_id", "timestamp", "open", "high", "low",
		"close", "volume", "adjusted_close", "dividend_amount",
	}

	// multiply by -1 to sort the data in descending order
	// TODO: do we want this is descending order or ascending order? I feel like its usually ascending?
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
	} else {
		return tx.CopyFrom(ctx, pgx.Identifier{"av_time_series_data"}, columns, pgx.CopyFromRows(entries))
	}
}

func (pg *Postgres) GetMostRecentTimestampForSymbol(ctx context.Context, symbol string) (*time.Time, error) {
	sql := q.Get(q.QueryHelper.Select.MostRecentTimestampBySymbol)
	args := pgx.NamedArgs{"symbol": symbol}

	ts := new(time.Time)
	if err := pg.db.QueryRow(ctx, sql, args).Scan(&ts); err != nil {
		return nil, fmt.Errorf("error getting most recent timestamp for symbol %s: %w", symbol, err)
	} else {
		return ts, nil
	}
}

func (pg *Postgres) GetTimeSeriesReturns(ctx context.Context, sourceIds []int32, maxLookback time.Duration) ([]*m.TimeSeriesReturn, error) {
	sql := q.Get(q.QueryHelper.Select.TimeSeriesReturns)
	args := pgx.NamedArgs{
		"source_ids":   sourceIds,
		"max_lookback": time.Now().Add(-maxLookback),
	}

	if res, err := Query[m.TimeSeriesReturn](ctx, pg, sql, args); err != nil {
		return nil, fmt.Errorf("unable to get time series returns by source id (%v): %w", sourceIds, err)
	} else {
		return res, nil
	}
}
