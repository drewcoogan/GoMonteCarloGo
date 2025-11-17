package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"mc.data/models"
)

func (pg *Postgres) GetTimeSeriesData(ctx context.Context, symbol string) ([]*models.TimeSeriesData, error) {
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

	res, err := Query[models.TimeSeriesData](ctx, pg, query, args)
	if err != nil {
		return nil, fmt.Errorf("unable to query data by symbol (%s): %w", symbol, err)
	}
	return res, nil
}

func (pg *Postgres) InsertTimeSeriesData(ctx context.Context, data []*models.TimeSeriesData, id *int32, tx *pgx.Tx) (int64, error) {
    columns := []string{
        "source_id", "timestamp", "open", "high", "low", 
        "close", "volume", "adjusted_close", "dividend_amount",
    }
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

func (pg *Postgres) GetMostRecentTimestampForSymbol(ctx context.Context, symbol string) (time.Time, error) {
	query := `
		SELECT 
			MAX(atsd.timestamp)
		FROM av_time_series_data atsd 
		JOIN av_time_series_metadata atsm ON atsd.source_id = atsm.id
		WHERE atsm.symbol = @symbol`

	args := pgx.NamedArgs{
		"symbol": symbol,
	}

	var ts time.Time
	if err := pg.db.QueryRow(ctx, query, args).Scan(&ts); err != nil {
		return time.Time{}, fmt.Errorf("error getting most recent timestamp for symbol %s: %w", symbol, err)
	}

	return ts, nil
}