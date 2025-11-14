package repos

import (
	"context"
	"fmt"

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

func (pg *Postgres) InsertTimeSeriesData(ctx context.Context, data []*models.TimeSeriesData) (int64, error) {
    columns := []string{
        "source_id", "timestamp", "open", "high", "low", 
        "close", "volume", "adjusted_close", "dividend_amount",
    }
    
    entries := make([][]any, len(data))
    for i, ent := range data {
        entries[i] = []any{
            ent.SourceId, ent.Timestamp, ent.Open, ent.High, ent.Low,
            ent.Close,  ent.Volume, ent.AdjustedClose, ent.DividendAmount,
        }
    }

	return pg.db.CopyFrom(ctx, pgx.Identifier{"av_time_series_data"}, columns, pgx.CopyFromRows(entries))
}