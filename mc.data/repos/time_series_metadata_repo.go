package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	m "mc.data/models"
)

func (pg *Postgres) GetMetaDataBySymbol(ctx context.Context, symbol string) (*m.TimeSeriesMetadata, error) {
	query := `
		SELECT 
			id, 
			symbol, 
			last_refreshed
		FROM av_time_series_metadata 
		WHERE symbol = @symbol`

	args := pgx.NamedArgs{
		"symbol": symbol,
	}

	res, err := Query[m.TimeSeriesMetadata](ctx, pg, query, args)
	if err != nil {
		return nil, fmt.Errorf("unable to query metadata by symbol (%s): %w", symbol, err)
	}

	if len(res) == 0 {
		return nil, nil
	}

	return res[0], nil
}

func (pg *Postgres) InsertNewMetaData(ctx context.Context, metadata *m.TimeSeriesMetadata, tx *pgx.Tx) error {
	query := `
		INSERT INTO av_time_series_metadata 
			(symbol, last_refreshed) 
		VALUES 
			(@symbol, @last_refreshed) 
		RETURNING id`

	args := pgx.NamedArgs{
		"symbol":         metadata.Symbol,
		"last_refreshed": metadata.LastRefreshed,
	}

	var err error
	if tx == nil {
		err = pg.db.QueryRow(ctx, query, args).Scan(&metadata.Id)
	} else {
		err = (*tx).QueryRow(ctx, query, args).Scan(&metadata.Id)
	}

	if err != nil {
		return fmt.Errorf("error inserting new metadata: %w", err)
	}

	return nil
}

func (pg *Postgres) UpdateLastRefreshedDate(ctx context.Context, symbol string, lastRefreshed time.Time, tx *pgx.Tx) (err error) {
	query := `
		UPDATE av_time_series_metadata
		SET last_refreshed = @last_refreshed
		WHERE symbol = @symbol`

	args := pgx.NamedArgs{
		"last_refreshed": lastRefreshed,
		"symbol":         symbol,
	}

	if tx == nil {
		_, err = pg.db.Exec(ctx, query, args)
	} else {
		_, err = (*tx).Exec(ctx, query, args)
	}

	return
}
