package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"mc.data/models"
)

func (pg *Postgres) GetMetaDataBySymbol(ctx context.Context, symbol string) (*models.TimeSeriesMetadata, error) {
	query := `SELECT symbol, last_refreshed
		FROM time_series_metadata 
		WHERE symbol = @symbol`
	
	args := pgx.NamedArgs{
		"symbol": symbol,
	}

	res, err := QuerySingle[models.TimeSeriesMetadata](ctx, pg, query, args)
	if err != nil {
		return nil, fmt.Errorf("unable to query metadata by symbol (%s): %w", symbol, err)
	}
	return res, nil
}

func (pg *Postgres) InsertNewMetaData(ctx context.Context, metadata *models.TimeSeriesMetadata) (id int64, err error) {
	query := `INSERT INTO time_series_metadata (symbol, last_refreshed) 
		VALUES (@symbol, @last_refreshed) 
		RETURNING id`

	args := pgx.NamedArgs{
		"symbol": metadata.Symbol,
		"last_refreshed": metadata.LastRefreshed,
	}

	err = pg.db.QueryRow(ctx, query, args).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error inserting new metadata: %w", err)
	}

	return
}